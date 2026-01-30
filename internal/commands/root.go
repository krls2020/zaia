package commands

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/auth"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

// Version variables set at build time via -ldflags
var (
	version = "dev"
	commit  = "none"
	built   = "unknown"
)

// defaultStoragePath returns empty string which causes Storage to use OS-default path.
func defaultStoragePath() string {
	return ""
}

// realClientFactory creates a real ZeropsClient from token and apiHost.
func realClientFactory(token, apiHost string) platform.Client {
	c, err := platform.NewZeropsClient(token, apiHost)
	if err != nil {
		return nil
	}
	return c
}

// RootDeps holds dependencies for building the root command.
// Used by NewRootForTest to inject mocks.
type RootDeps struct {
	StoragePath   string
	Client        platform.Client
	ClientFactory func(token, apiHost string) platform.Client
	LogFetcher    platform.LogFetcher
}

// NewRootForTest creates the root command with injected dependencies for testing.
func NewRootForTest(deps RootDeps) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "zaia",
		Short:         "ZAIA - AI agent CLI for Zerops",
		SilenceUsage:  true,
		SilenceErrors: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: noCommandRunE,
	}

	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug output on stderr")

	storagePath := deps.StoragePath
	client := deps.Client
	fetcher := deps.LogFetcher

	clientFactory := deps.ClientFactory
	if clientFactory == nil {
		clientFactory = func(token, apiHost string) platform.Client { return client }
	}

	rootCmd.AddCommand(NewLogin(storagePath, clientFactory))
	rootCmd.AddCommand(NewLogout(storagePath))
	rootCmd.AddCommand(NewStatus(storagePath))
	rootCmd.AddCommand(NewVersion())
	rootCmd.AddCommand(NewDiscover(storagePath, client))
	rootCmd.AddCommand(NewProcess(storagePath, client))
	rootCmd.AddCommand(NewCancel(storagePath, client))
	rootCmd.AddCommand(NewLogs(storagePath, client, fetcher))
	rootCmd.AddCommand(NewValidate())
	rootCmd.AddCommand(NewSearch())
	rootCmd.AddCommand(NewStart(storagePath, client))
	rootCmd.AddCommand(NewStop(storagePath, client))
	rootCmd.AddCommand(NewRestart(storagePath, client))
	rootCmd.AddCommand(NewScale(storagePath, client))
	rootCmd.AddCommand(NewEnv(storagePath, client))
	rootCmd.AddCommand(NewImport(storagePath, client))
	rootCmd.AddCommand(NewDelete(storagePath, client))
	rootCmd.AddCommand(NewSubdomain(storagePath, client))

	return rootCmd
}

func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "zaia",
		Short:         "ZAIA - AI agent CLI for Zerops",
		SilenceUsage:  true,
		SilenceErrors: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: noCommandRunE,
	}

	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug output on stderr")

	storagePath := defaultStoragePath()

	// Lazy client — creates ZeropsClient on first API call by reading stored credentials.
	client := platform.NewLazyClient(func() (string, string, error) {
		storage := auth.NewStorage(storagePath)
		mgr := auth.NewManager(storage, nil)
		creds, err := mgr.GetCredentials()
		if err != nil {
			return "", "", err
		}
		return creds.Token, creds.APIHost, nil
	})

	fetcher := platform.NewLogFetcher()

	rootCmd.AddCommand(NewLogin(storagePath, realClientFactory))
	rootCmd.AddCommand(NewLogout(storagePath))
	rootCmd.AddCommand(NewStatus(storagePath))
	rootCmd.AddCommand(NewVersion())
	rootCmd.AddCommand(NewDiscover(storagePath, client))
	rootCmd.AddCommand(NewProcess(storagePath, client))
	rootCmd.AddCommand(NewCancel(storagePath, client))
	rootCmd.AddCommand(NewLogs(storagePath, client, fetcher))
	rootCmd.AddCommand(NewValidate())
	rootCmd.AddCommand(NewSearch())
	rootCmd.AddCommand(NewStart(storagePath, client))
	rootCmd.AddCommand(NewStop(storagePath, client))
	rootCmd.AddCommand(NewRestart(storagePath, client))
	rootCmd.AddCommand(NewScale(storagePath, client))
	rootCmd.AddCommand(NewEnv(storagePath, client))
	rootCmd.AddCommand(NewImport(storagePath, client))
	rootCmd.AddCommand(NewDelete(storagePath, client))
	rootCmd.AddCommand(NewSubdomain(storagePath, client))

	return rootCmd
}

// Execute wraps cmd.Execute() — converts any non-JSON Cobra error to JSON envelope.
func Execute(cmd *cobra.Command) error {
	err := cmd.Execute()
	if err == nil {
		return nil
	}
	var zaiaErr *output.ZaiaError
	if errors.As(err, &zaiaErr) {
		return err // already JSON on stdout
	}
	// Cobra-level error (unknown command, missing args, unknown flag) — wrap as JSON
	return output.Err(platform.ErrInvalidUsage, err.Error(), "Run: zaia --help", nil)
}

// noCommandRunE is the RunE for parent commands that require a subcommand.
func noCommandRunE(cmd *cobra.Command, args []string) error {
	return output.Err(platform.ErrInvalidUsage,
		"No command specified",
		"Run: zaia <command>",
		map[string]interface{}{"availableCommands": commandNames(cmd)})
}

// commandNames returns a list of subcommand names for a command.
func commandNames(cmd *cobra.Command) []string {
	cmds := cmd.Commands()
	names := make([]string, 0, len(cmds))
	for _, c := range cmds {
		if !c.Hidden {
			names = append(names, c.Name())
		}
	}
	return names
}
