package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/katyella/lazyoc/internal/ui"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	ctx := context.Background()

	var debugMode bool
	var noAltScreen bool
	var kubeconfigPath string
	var mouseSupport bool

	rootCmd := &cobra.Command{
		Use:   "lazyoc",
		Short: "LazyOC - A lazy terminal UI for OpenShift/Kubernetes clusters",
		Long: `LazyOC is a terminal-based user interface for managing OpenShift and Kubernetes clusters.
It provides an intuitive, vim-like interface for viewing and managing cluster resources.

Press ? for help once inside the application.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		Run: func(cmd *cobra.Command, args []string) {
			runTUI(debugMode, !noAltScreen, kubeconfigPath, mouseSupport)
		},
	}

	// Add flags
	rootCmd.Flags().BoolP("version", "v", false, "Print version information")
	rootCmd.Flags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug mode (logs to lazyoc.log)")
	rootCmd.Flags().BoolVar(&noAltScreen, "no-alt-screen", false, "Disable alternate screen buffer")
	rootCmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "", "Path to kubeconfig file (defaults to $HOME/.kube/config)")
	rootCmd.Flags().BoolVar(&mouseSupport, "mouse", true, "Enable mouse support (click tabs, select resources, scroll)")

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Fatalf("Error executing command: %v", err)
		os.Exit(1)
	}
}

// runTUI starts the terminal user interface
func runTUI(debug bool, altScreen bool, kubeconfigPath string, mouseSupport bool) {
	opts := ui.ProgramOptions{
		Version:      fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		Debug:        debug,
		AltScreen:    altScreen,
		MouseSupport: mouseSupport,
		KubeConfig:   kubeconfigPath,
	}

	if err := ui.RunTUI(opts); err != nil {
		log.Fatalf("Error running TUI: %v", err)
		os.Exit(1)
	}
}
