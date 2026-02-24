package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/opentdf/platform/service/internal/devidp"
	"github.com/spf13/cobra"
)

func init() {
	devIDPCmd := cobra.Command{
		Use:   "dev-idp",
		Short: "Start the dev OIDC IdP (client credentials only)",
		RunE:  runDevIDP,
	}
	devIDPCmd.SilenceUsage = true

	rootCmd.AddCommand(&devIDPCmd)
}

func runDevIDP(cmd *cobra.Command, _ []string) error {
	configFile, _ := cmd.Flags().GetString(configFileFlag)
	if configFile == "" {
		return errors.New("config-file is required")
	}

	cfg, err := devidp.LoadConfig(configFile)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return devidp.Run(ctx, cfg)
}
