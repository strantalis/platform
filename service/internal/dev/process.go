package dev

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	defaultShutdownTimeout = 5 * time.Second
	maxDirSearchDepth      = 6
	tcpDialTimeout         = 1 * time.Second
	waitPollInterval       = 500 * time.Millisecond
	shutdownPollInterval   = 200 * time.Millisecond
	shutdownGraceSleep     = 500 * time.Millisecond
	httpClientTimeout      = 2 * time.Second
)

func FindPlatformDir(startDir, override string) (string, error) {
	if override != "" {
		if isPlatformDir(override) {
			return override, nil
		}
		return "", fmt.Errorf("platform dir not found at %s", override)
	}

	dir := startDir
	for i := 0; i < maxDirSearchDepth; i++ {
		if isPlatformDir(dir) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", errors.New("platform dir not found; use --platform-dir")
}

func isPlatformDir(path string) bool {
	if path == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, "docker-compose.yaml")); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, "service")); err != nil {
		return false
	}
	return true
}

func StartProcess(cmd *exec.Cmd, pidPath string) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	return writePID(pidPath, cmd.Process.Pid)
}

func StopProcess(pidPath string) error {
	pid, err := readPID(pidPath)
	if err != nil {
		return err
	}
	if pid == 0 {
		return nil
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	_ = proc.Signal(os.Interrupt)

	deadline := time.Now().Add(defaultShutdownTimeout)
	for time.Now().Before(deadline) {
		if !isProcessRunning(proc) {
			_ = os.Remove(pidPath)
			return nil
		}
		time.Sleep(shutdownPollInterval)
	}

	_ = proc.Signal(syscall.SIGTERM)
	time.Sleep(shutdownGraceSleep)

	if isProcessRunning(proc) {
		_ = proc.Signal(syscall.SIGKILL)
	}

	_ = os.Remove(pidPath)
	return nil
}

func isProcessRunning(proc *os.Process) bool {
	if proc == nil {
		return false
	}
	err := proc.Signal(syscall.Signal(0))
	return err == nil
}

func writePID(path string, pid int) error {
	if err := os.MkdirAll(filepath.Dir(path), defaultDirPerm); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), defaultFilePerm)
}

func readPID(path string) (int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func WaitForTCP(ctx context.Context, addr string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		d := net.Dialer{Timeout: tcpDialTimeout}
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		if ctx.Err() != nil {
			return fmt.Errorf("timeout waiting for %s", addr)
		}
		time.Sleep(waitPollInterval)
	}
}

func WaitForHTTP(ctx context.Context, url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client := http.Client{Timeout: httpClientTimeout}
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			_ = resp.Body.Close()
			return nil
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		if ctx.Err() != nil {
			return fmt.Errorf("timeout waiting for %s", url)
		}
		time.Sleep(waitPollInterval)
	}
}

type Compose struct {
	command string
	args    []string
}

func (c Compose) Command() string {
	return c.command
}

func (c Compose) Args() []string {
	return append([]string{}, c.args...)
}

func FindCompose() (Compose, error) {
	if dockerPath, err := exec.LookPath("docker"); err == nil {
		cmd := exec.CommandContext(context.Background(), dockerPath, "compose", "version")
		if err := cmd.Run(); err == nil {
			return Compose{command: dockerPath, args: []string{"compose"}}, nil
		}
	}
	if dcPath, err := exec.LookPath("docker-compose"); err == nil {
		return Compose{command: dcPath}, nil
	}
	return Compose{}, errors.New("docker compose not found")
}

func (c Compose) Up(ctx context.Context, platformDir string, service string) error {
	composeFile := filepath.Join(platformDir, "docker-compose.yaml")
	args := append([]string{}, c.args...)
	args = append(args, "-f", composeFile, "up", "-d", service)
	//nolint:gosec // dev command runs a local binary from a controlled path
	cmd := exec.CommandContext(ctx, c.command, args...)
	cmd.Dir = platformDir
	cmd.Env = composeEnv()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c Compose) Stop(ctx context.Context, platformDir string, service string) error {
	composeFile := filepath.Join(platformDir, "docker-compose.yaml")
	args := append([]string{}, c.args...)
	args = append(args, "-f", composeFile, "stop", service)
	//nolint:gosec // dev command runs a local binary from a controlled path
	cmd := exec.CommandContext(ctx, c.command, args...)
	cmd.Dir = platformDir
	cmd.Env = composeEnv()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func composeEnv() []string {
	env := os.Environ()
	if _, ok := os.LookupEnv("JAVA_OPTS_APPEND"); !ok {
		env = append(env, "JAVA_OPTS_APPEND=")
	}
	return env
}

func ProcessStatus(pidPath string) (int, bool, error) {
	pid, err := readPID(pidPath)
	if err != nil {
		return 0, false, err
	}
	if pid == 0 {
		return 0, false, nil
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return pid, false, err
	}
	return pid, isProcessRunning(proc), nil
}

func CheckHTTP(ctx context.Context, url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client := http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
