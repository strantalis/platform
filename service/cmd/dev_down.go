package cmd

import (
	"context"
	"fmt"
	"os"

	devpkg "github.com/opentdf/platform/service/internal/dev"
	"github.com/spf13/cobra"
)

type downOptions struct {
	dataDir     string
	platformDir string
}

func newDevDownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "down",
		Short:        "Stop the local dev stack",
		SilenceUsage: true,
		RunE:         devDownRun,
	}
	cmd.Flags().String("data-dir", "", "Path for dev configs, keys, and binaries (default is ~/.opentdf/dev)")
	cmd.Flags().String("platform-dir", "", "Path to the platform repo (auto-detected if omitted)")
	return cmd
}

func devDownRun(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	opts := downOptions{
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

	fmt.Fprintln(out, sectionText("Stopping dev stack..."))

	_ = runSpinnerStep(out, "Stopping platform", func() error {
		return devpkg.StopProcess(layout.PlatformPIDPath)
	})
	_ = runSpinnerStep(out, "Stopping dev IdP", func() error {
		return devpkg.StopProcess(layout.IDPPIDPath)
	})

	if compose, err := devpkg.FindCompose(); err == nil {
		_ = runSpinnerStep(out, "Stopping Postgres", func() error {
			return compose.Stop(context.Background(), platformDir, "opentdfdb")
		})
	}

	return nil
}
