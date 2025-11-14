#!/usr/bin/env python3
"""Minimal ctypes harness for libtdf.so (TDFNewWriter/TDFWriteSegment/TDFFinalize)."""

from __future__ import annotations

import argparse
import ctypes
import pathlib


STATUS_OK = 0
STATUS_ERROR = -1
STATUS_NO_WRITER = -2
STATUS_BUFFER_TOO_SMALL = -3

UInt8Ptr = ctypes.POINTER(ctypes.c_uint8)
CharPtr = ctypes.POINTER(ctypes.c_char)


def _repo_root() -> pathlib.Path:
    current = pathlib.Path(__file__).resolve()
    for candidate in [current, *current.parents]:
        if (candidate / "go.work").exists():
            return candidate
    return current.parents[-1]


ROOT = _repo_root()
LIB_DEFAULT = ROOT / "examples/wasm/tdf_poc/cshared/libtdf.so"
TESTDATA = ROOT / "examples/wasm/tdf_poc/cshared/testdata"


def _load_bytes(path: pathlib.Path) -> bytes:
    try:
        return path.read_bytes()
    except OSError as exc:
        raise SystemExit(f"failed to read {path}: {exc}") from exc


def _make_uint8_buffer(data: bytes) -> tuple[UInt8Ptr, ctypes.Array[ctypes.c_uint8]]:
    if not data:
        return ctypes.cast(ctypes.c_void_p(), UInt8Ptr), (ctypes.c_uint8 * 0)()
    arr = (ctypes.c_uint8 * len(data)).from_buffer_copy(data)
    return ctypes.cast(arr, UInt8Ptr), arr


def _get_last_error(lib) -> str:
    length = lib.TDFGetLastError(ctypes.cast(ctypes.c_void_p(), CharPtr), ctypes.c_size_t(0))
    if length <= 0:
        return ""
    buf = ctypes.create_string_buffer(length + 1)
    status = lib.TDFGetLastError(buf, ctypes.c_size_t(len(buf)))
    if status < 0:
        return ""
    return buf.value.decode("utf-8", errors="replace")


def _call_with_buffer(lib, label, func, initial_cap: int = 64 * 1024) -> bytes:
    cap = initial_cap
    while True:
        out_arr = (ctypes.c_uint8 * cap)()
        status = func(ctypes.cast(out_arr, UInt8Ptr), ctypes.c_size_t(cap))
        if status >= 0:
            return bytes(out_arr[: status])
        if status == STATUS_BUFFER_TOO_SMALL:
            cap *= 2
            continue
        raise SystemExit(f"{label} failed (status={status}): {_get_last_error(lib)}")


def main() -> None:
    parser = argparse.ArgumentParser(description="Use libtdf.so to write a TDF zip via the staged API")
    parser.add_argument("--lib", type=pathlib.Path, default=LIB_DEFAULT, help="Path to libtdf.so (default: %(default)s)")
    parser.add_argument("--kas", type=pathlib.Path, default=TESTDATA / "kas.json", help="KAS JSON input")
    parser.add_argument("--attrs", type=pathlib.Path, default=TESTDATA / "attributes.json", help="Attributes JSON input")
    parser.add_argument("--payload", type=pathlib.Path, default=TESTDATA / "payload.txt", help="Payload bytes")
    parser.add_argument("--out", type=pathlib.Path, default=pathlib.Path("demo-python.tdf"), help="Destination TDF path")
    args = parser.parse_args()

    lib = ctypes.CDLL(str(args.lib))
    lib.TDFNewWriter.argtypes = [UInt8Ptr, ctypes.c_size_t]
    lib.TDFNewWriter.restype = ctypes.c_int32
    lib.TDFWriteSegment.argtypes = [
        ctypes.c_int32,
        UInt8Ptr,
        ctypes.c_size_t,
        UInt8Ptr,
        ctypes.c_size_t,
    ]
    lib.TDFWriteSegment.restype = ctypes.c_int32
    lib.TDFFinalize.argtypes = [UInt8Ptr, ctypes.c_size_t, UInt8Ptr, ctypes.c_size_t]
    lib.TDFFinalize.restype = ctypes.c_int32
    lib.TDFGetLastError.argtypes = [CharPtr, ctypes.c_size_t]
    lib.TDFGetLastError.restype = ctypes.c_int32
    lib.TDFReset.argtypes = []
    lib.TDFReset.restype = ctypes.c_int32

    kas_bytes = _load_bytes(args.kas)
    attr_bytes = _load_bytes(args.attrs)
    payload_bytes = _load_bytes(args.payload)

    kas_ptr, kas_arr = _make_uint8_buffer(kas_bytes)
    attrs_ptr, attrs_arr = _make_uint8_buffer(attr_bytes)
    payload_ptr, payload_arr = _make_uint8_buffer(payload_bytes)

    status = lib.TDFNewWriter(kas_ptr, ctypes.c_size_t(len(kas_bytes)))
    if status != STATUS_OK:
        raise SystemExit(f"TDFNewWriter failed (status={status}): {_get_last_error(lib)}")

    try:
        segment_chunk = _call_with_buffer(
            lib,
            "TDFWriteSegment",
            lambda out_buf, out_cap: lib.TDFWriteSegment(
                ctypes.c_int32(0),
                payload_ptr,
                ctypes.c_size_t(len(payload_bytes)),
                out_buf,
                out_cap,
            ),
        )

        finalize_chunk = _call_with_buffer(
            lib,
            "TDFFinalize",
            lambda out_buf, out_cap: lib.TDFFinalize(
                attrs_ptr,
                ctypes.c_size_t(len(attr_bytes)),
                out_buf,
                out_cap,
            ),
        )
    finally:
        lib.TDFReset()

    archive = segment_chunk + finalize_chunk
    args.out.write_bytes(archive)
    print(f"Wrote {len(archive)} bytes to {args.out}")


if __name__ == "__main__":
    main()
