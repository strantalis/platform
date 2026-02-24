package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	devpkg "github.com/opentdf/platform/service/internal/dev"
	"github.com/spf13/cobra"
)

const (
	defaultIDPPort         = 8082
	defaultPlatformPort    = 8080
	defaultPostgresTimeout = 45 * time.Second
	defaultIDPTimeout      = 20 * time.Second
	defaultPlatformTimeout = 30 * time.Second
)

type upOptions struct {
	dataDir      string
	platformDir  string
	idpPort      int
	platformPort int
	regenerate   bool
	detach       bool
	skipCompose  bool
}

func newDevUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "up",
		Short:        "Start a local dev stack",
		SilenceUsage: true,
		RunE:         devUpRun,
	}
	cmd.Flags().String("data-dir", "", "Path for dev configs, keys, and binaries (default is ~/.opentdf/dev)")
	cmd.Flags().String("platform-dir", "", "Path to the platform repo (auto-detected if omitted)")
	cmd.Flags().Int("idp-port", defaultIDPPort, "Port for the dev IdP")
	cmd.Flags().Int("platform-port", defaultPlatformPort, "Port for the platform server")
	cmd.Flags().Bool("regen", false, "Regenerate keys and configs")
	cmd.Flags().Bool("detach", false, "Run in background and return immediately")
	cmd.Flags().Bool("skip-compose", false, "Skip docker compose startup for Postgres")
	return cmd
}

func devUpRun(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	opts := upOptions{
		dataDir:      cmd.Flags().Lookup("data-dir").Value.String(),
		platformDir:  cmd.Flags().Lookup("platform-dir").Value.String(),
		idpPort:      mustIntFlag(cmd, "idp-port"),
		platformPort: mustIntFlag(cmd, "platform-port"),
		regenerate:   mustBoolFlag(cmd, "regen"),
		detach:       mustBoolFlag(cmd, "detach"),
		skipCompose:  mustBoolFlag(cmd, "skip-compose"),
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

	idpConfig, err := devpkg.EnsureConfigs(layout, devpkg.Options{
		IDPPort:      opts.idpPort,
		PlatformPort: opts.platformPort,
		Regenerate:   opts.regenerate,
	})
	if err != nil {
		return fmt.Errorf("generate dev config: %w", err)
	}

	fmt.Fprintln(out, sectionText("Starting dev stack..."))

	if !opts.skipCompose {
		compose, err := devpkg.FindCompose()
		if err != nil {
			return fmt.Errorf("docker compose not available: %w", err)
		}
		if err := compose.Up(context.Background(), platformDir, "opentdfdb"); err != nil {
			return fmt.Errorf("start postgres: %w", err)
		}
		if err := runSpinnerStep(out, "Waiting for Postgres", func() error {
			return devpkg.WaitForTCP(context.Background(), "127.0.0.1:5432", defaultPostgresTimeout)
		}); err != nil {
			return fmt.Errorf("postgres did not become ready: %w", err)
		}
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve platform binary: %w", err)
	}

	runCtx := context.Background()

	idpCmd := exec.CommandContext(runCtx, execPath, "dev-idp", "--config-file", layout.IDPConfigPath)
	idpCmd.Stdout = out
	idpCmd.Stderr = errOut
	idpCmd.Dir = platformDir

	if err := devpkg.StartProcess(idpCmd, layout.IDPPIDPath); err != nil {
		return fmt.Errorf("start dev idp: %w", err)
	}
	if err := runSpinnerStep(out, "Waiting for dev IdP", func() error {
		return devpkg.WaitForHTTP(context.Background(), idpConfig.Issuer+"/.well-known/openid-configuration", defaultIDPTimeout)
	}); err != nil {
		return fmt.Errorf("dev idp did not become ready: %w", err)
	}

	platformCmd := exec.CommandContext(runCtx, execPath, "start", "--config-file", layout.PlatformConfigPath)
	platformCmd.Stdout = out
	platformCmd.Stderr = errOut
	platformCmd.Dir = platformDir

	if err := devpkg.StartProcess(platformCmd, layout.PlatformPIDPath); err != nil {
		return fmt.Errorf("start platform: %w", err)
	}

	if err := runSpinnerStep(out, "Waiting for platform", func() error {
		return devpkg.WaitForHTTP(context.Background(), idpConfig.Audience+"/.well-known/opentdf-configuration", defaultPlatformTimeout)
	}); err != nil {
		return fmt.Errorf("platform did not become ready: %w", err)
	}

	if len(idpConfig.Clients) == 0 {
		return errors.New("no dev idp clients configured")
	}
	rootKey, err := devpkg.LoadRootKey(layout.RootKeyPath)
	if err != nil {
		return fmt.Errorf("read root key: %w", err)
	}

	var kasSeed *devpkg.KasSeedResult
	if err := runSpinnerStep(out, "Seeding KAS keys", func() error {
		var seedErr error
		kasSeed, seedErr = devpkg.EnsureKasRegistryKeys(context.Background(), devpkg.KasSeedOptions{
			PlatformEndpoint: idpConfig.Audience,
			KasURI:           idpConfig.Audience,
			ClientID:         idpConfig.Clients[0].ID,
			ClientSecret:     idpConfig.Clients[0].Secret,
			RootKey:          rootKey,
			Regenerate:       opts.regenerate,
		})
		return seedErr
	}); err != nil {
		return fmt.Errorf("seed kas keys: %w", err)
	}

	rows := [][]string{
		{"Dev IdP", idpConfig.Issuer},
		{"Platform", idpConfig.Audience},
		{"Platform Config", layout.PlatformConfigPath},
		{"Dev IdP Config", layout.IDPConfigPath},
		{"KAS Registry", fmt.Sprintf("%s (%s)", kasSeed.KasURI, kasSeed.KasID)},
	}
	if len(idpConfig.Clients) > 0 {
		rows = append(rows,
			[]string{"Client ID", idpConfig.Clients[0].ID},
			[]string{"Client Secret", idpConfig.Clients[0].Secret},
		)
	}
	for _, key := range kasSeed.Keys {
		status := existingText("existing")
		if key.Created {
			status = createdText("created")
		}
		rows = append(rows, []string{"KAS Key " + key.Kid, status})
	}

	fmt.Fprintf(out, "%s %s\n", okText("[ready]"), "Dev stack is running")
	fmt.Fprint(out, formatRows(rows))
	fmt.Fprintln(out, hintText("Press Ctrl+C to stop the dev stack (or re-run with --detach)."))

	if opts.detach {
		return nil
	}

	sig := waitForShutdown()
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Received %s. Shutting down dev stack...\n", sig)

	_ = runSpinnerStep(out, "Stopping platform", func() error {
		return devpkg.StopProcess(layout.PlatformPIDPath)
	})
	_ = runSpinnerStep(out, "Stopping dev IdP", func() error {
		return devpkg.StopProcess(layout.IDPPIDPath)
	})

	if !opts.skipCompose {
		if compose, err := devpkg.FindCompose(); err == nil {
			_ = runSpinnerStep(out, "Stopping Postgres", func() error {
				return compose.Stop(context.Background(), platformDir, "opentdfdb")
			})
		}
	}

	return nil
}

func waitForShutdown() os.Signal {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	return <-sigCh
}

func mustIntFlag(cmd *cobra.Command, name string) int {
	v, err := cmd.Flags().GetInt(name)
	if err != nil {
		panic(err)
	}
	return v
}

func mustBoolFlag(cmd *cobra.Command, name string) bool {
	v, err := cmd.Flags().GetBool(name)
	if err != nil {
		panic(err)
	}
	return v
}
