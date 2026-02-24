package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	devpkg "github.com/opentdf/platform/service/internal/dev"
	"github.com/spf13/cobra"
)

const (
	statusTimeout = 2 * time.Second
)

type statusOptions struct {
	dataDir     string
	platformDir string
}

func newDevStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "status",
		Short:        "Show dev stack status",
		SilenceUsage: true,
		RunE:         devStatusRun,
	}
	cmd.Flags().String("data-dir", "", "Path for dev configs, keys, and binaries (default is ~/.opentdf/dev)")
	cmd.Flags().String("platform-dir", "", "Path to the platform repo (auto-detected if omitted)")
	return cmd
}

func devStatusRun(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	opts := statusOptions{
		dataDir:     cmd.Flags().Lookup("data-dir").Value.String(),
		platformDir: cmd.Flags().Lookup("platform-dir").Value.String(),
	}

	if opts.dataDir == "" {
		defaultDir, err := devpkg.DefaultDataDir()
		if err != nil {
			return fmt.Errorf("resolve default data dir: %w", err)
		}
		opts.dataDir = defaultDir
	}

	layout, err := devpkg.EnsureLayout(opts.dataDir)
	if err != nil {
		return fmt.Errorf("prepare dev layout: %w", err)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}

	platformDir, err := devpkg.FindPlatformDir(workingDir, opts.platformDir)
	if err != nil {
		return err
	}

	idpConfig, idpErr := devpkg.LoadIDPConfig(layout.IDPConfigPath)
	rows := [][]string{
		{"Platform Config", layout.PlatformConfigPath},
		{"Dev IdP Config", layout.IDPConfigPath},
	}

	if idpErr == nil {
		rows = append(rows,
			[]string{"Dev IdP", idpConfig.Issuer},
			[]string{"Platform", idpConfig.Audience},
		)
		if len(idpConfig.Clients) > 0 {
			rows = append(rows,
				[]string{"Client ID", idpConfig.Clients[0].ID},
				[]string{"Client Secret", idpConfig.Clients[0].Secret},
			)
		}
	} else {
		rows = append(rows, []string{"Dev IdP", warningText("not configured")})
	}

	platformPID, platformRunning, _ := devpkg.ProcessStatus(layout.PlatformPIDPath)
	idpPID, idpRunning, _ := devpkg.ProcessStatus(layout.IDPPIDPath)

	rows = append(rows,
		[]string{"Platform PID", formatPID(platformPID, platformRunning)},
		[]string{"Dev IdP PID", formatPID(idpPID, idpRunning)},
	)

	if idpErr == nil {
		idpStatus := warningText("unreachable")
		if err := devpkg.CheckHTTP(context.Background(), idpConfig.Issuer+"/healthz", statusTimeout); err == nil {
			idpStatus = okText("ok")
		}
		rows = append(rows, []string{"Dev IdP Health", idpStatus})

		platformStatus := warningText("unreachable")
		if err := devpkg.CheckHTTP(context.Background(), idpConfig.Audience+"/healthz", statusTimeout); err == nil {
			platformStatus = okText("ok")
		}
		rows = append(rows, []string{"Platform Health", platformStatus})
	}

	if compose, err := devpkg.FindCompose(); err == nil {
		rows = append(rows, []string{"Docker Compose", fmt.Sprintf("%s %s", compose.Command(), strings.Join(compose.Args(), " "))})
		rows = append(rows, []string{"Compose File", platformDir + "/docker-compose.yaml"})
	}

	fmt.Fprintf(out, "%s %s\n", okText("[status]"), "Dev stack")
	fmt.Fprint(out, formatRows(rows))
	return nil
}

func formatPID(pid int, running bool) string {
	if pid == 0 {
		return warningText("not running")
	}
	if running {
		return fmt.Sprintf("%d %s", pid, okText("(running)"))
	}
	return fmt.Sprintf("%d %s", pid, warningText("(stopped)"))
}
