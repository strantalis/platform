#!/usr/bin/env node
// Minimal Node harness for libtdf.so (requires ffi-napi + ref-napi)

const fs = require("fs");
const path = require("path");
const ffi = require("ffi-napi");
const ref = require("ref-napi");

const STATUS_OK = 0;
const STATUS_BUFFER_TOO_SMALL = -3;

function findRepoRoot(startDir) {
  let dir = startDir;
  while (true) {
    if (fs.existsSync(path.join(dir, "go.work"))) {
      return dir;
    }
    const parent = path.dirname(dir);
    if (parent === dir) {
      break;
    }
    dir = parent;
  }
  return startDir;
}

const root = findRepoRoot(__dirname);
const defaultLib = path.join(root, "examples/wasm/tdf_poc/cshared/libtdf.so");
const testdataDir = path.join(root, "examples/wasm/tdf_poc/cshared/testdata");

const uint8Ptr = ref.refType(ref.types.uint8);
const charPtr = ref.refType(ref.types.char);

function parseArgs() {
  const args = process.argv.slice(2);
  const opts = {
    lib: defaultLib,
    kas: path.join(testdataDir, "kas.json"),
    attrs: path.join(testdataDir, "attributes.json"),
    payload: path.join(testdataDir, "payload.txt"),
    out: path.resolve("demo-node.tdf"),
  };

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (!arg.startsWith("--")) {
      usage();
    }
    const key = arg.slice(2);
    if (!(key in opts)) {
      usage();
    }
    i++;
    if (i >= args.length) {
      usage();
    }
    opts[key] = path.resolve(args[i]);
  }
  return opts;
}

function usage() {
  console.error("Usage: node run.cjs [--lib path] [--kas path] [--attrs path] [--payload path] [--out path]");
  process.exit(1);
}

function loadBytes(p) {
  try {
    return fs.readFileSync(p);
  } catch (err) {
    console.error(`failed to read ${p}: ${err.message}`);
    process.exit(1);
  }
}

function getLastError(lib) {
  const needed = lib.TDFGetLastError(ref.NULL, 0);
  if (needed <= 0) {
    return "";
  }
  const buf = Buffer.alloc(needed + 1);
  const status = lib.TDFGetLastError(buf, buf.length);
  if (status < 0) {
    return "";
  }
  return buf.toString("utf8").replace(/\0.*$/, "").trim();
}

function callWithBuffer(label, lib, fn) {
  let cap = 64 * 1024;
  while (true) {
    const outBuf = Buffer.alloc(cap);
    const status = fn(outBuf, cap);
    if (status >= 0) {
      return outBuf.subarray(0, status);
    }
    if (status === STATUS_BUFFER_TOO_SMALL) {
      cap *= 2;
      continue;
    }
    console.error(`${label} failed (status=${status}): ${getLastError(lib) || "unknown error"}`);
    process.exit(1);
  }
}

function main() {
  const opts = parseArgs();
  const lib = ffi.Library(opts.lib, {
    TDFNewWriter: ["int32", [uint8Ptr, ref.types.size_t]],
    TDFWriteSegment: ["int32", ["int32", uint8Ptr, ref.types.size_t, uint8Ptr, ref.types.size_t]],
    TDFFinalize: ["int32", [uint8Ptr, ref.types.size_t, uint8Ptr, ref.types.size_t]],
    TDFGetLastError: ["int32", [charPtr, ref.types.size_t]],
    TDFReset: ["int32", []],
  });

  const kasBuf = loadBytes(opts.kas);
  const attrBuf = loadBytes(opts.attrs);
  const payloadBuf = loadBytes(opts.payload);

  const newWriterStatus = lib.TDFNewWriter(kasBuf, kasBuf.length);
  if (newWriterStatus !== STATUS_OK) {
    console.error(`TDFNewWriter failed (status=${newWriterStatus}): ${getLastError(lib) || "unknown error"}`);
    process.exit(1);
  }

  let segmentChunk;
  let finalizeChunk;
  try {
    segmentChunk = callWithBuffer("TDFWriteSegment", lib, (outBuf, cap) =>
      lib.TDFWriteSegment(0, payloadBuf, payloadBuf.length, outBuf, cap),
    );

    finalizeChunk = callWithBuffer("TDFFinalize", lib, (outBuf, cap) =>
      lib.TDFFinalize(attrBuf, attrBuf.length, outBuf, cap),
    );
  } finally {
    lib.TDFReset();
  }

  const archive = Buffer.concat([segmentChunk, finalizeChunk]);
  fs.writeFileSync(opts.out, archive);
  console.log(`Wrote ${archive.length} bytes to ${opts.out}`);
}

main();
