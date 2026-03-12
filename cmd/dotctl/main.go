package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rommelporras/dotfiles/internal/collector"
	"github.com/rommelporras/dotfiles/internal/config"
	"github.com/rommelporras/dotfiles/internal/display"
	"github.com/rommelporras/dotfiles/internal/model"
	"github.com/rommelporras/dotfiles/internal/push"
	"github.com/rommelporras/dotfiles/internal/query"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dotctl",
		Short: "Dotfiles status dashboard",
	}

	// dotctl status
	var live bool
	var machine string
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show dotfiles status across all machines",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _ := config.Load(config.DefaultPath())
			if live {
				return runStatusLive(cfg, machine)
			}
			return runStatusRemote(cfg, machine)
		},
	}
	statusCmd.Flags().BoolVar(&live, "live", false, "Collect fresh data locally instead of querying Prometheus")
	statusCmd.Flags().StringVar(&machine, "machine", "", "Filter to a single machine")
	rootCmd.AddCommand(statusCmd)

	// dotctl collect
	var container string
	var verbose bool
	collectCmd := &cobra.Command{
		Use:   "collect",
		Short: "Collect status and push to OTel Collector",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _ := config.Load(config.DefaultPath())
			return runCollect(cfg, container, verbose)
		},
	}
	collectCmd.Flags().StringVar(&container, "container", "", "Collect from a single container only")
	collectCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	rootCmd.AddCommand(collectCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runStatusLive(cfg *config.Config, filterMachine string) error {
	runner := &collector.ExecRunner{}
	platform := collector.DetectPlatform()
	result, err := collector.CollectAll(runner, cfg.Hostname, platform)
	if err != nil {
		return err
	}

	machines := result.Machines
	if filterMachine != "" {
		machines = filterByHostname(machines, filterMachine)
	}

	fmt.Println(display.RenderAll(machines, result.Containers))
	return nil
}

func runStatusRemote(cfg *config.Config, filterMachine string) error {
	promClient := query.NewPrometheusClient(cfg.PrometheusURL)
	promMachines, err := promClient.QueryMachines()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Prometheus unreachable (%v), falling back to --live\n", err)
		return runStatusLive(cfg, filterMachine)
	}

	driftTotals, _ := promClient.QueryDriftTotals()
	toolsMap, _ := promClient.QueryToolsInstalled()
	credsMap, _ := promClient.QueryCredentials()

	lokiClient := query.NewLokiClient(cfg.LokiURL)
	lokiStates, _ := lokiClient.QueryLatestStates()

	lokiByHost := make(map[string]*model.MachineState)
	for i := range lokiStates {
		lokiByHost[lokiStates[i].Hostname] = &lokiStates[i]
	}

	var machines []model.MachineState
	for _, pm := range promMachines {
		ms := model.MachineState{
			Hostname: pm.Hostname,
			Platform: pm.Platform,
			Context:  pm.Context,
		}

		if ls, ok := lokiByHost[pm.Hostname]; ok {
			ms.DriftFiles = ls.DriftFiles
			ms.TemplateData = ls.TemplateData
			ms.SSHAgent = ls.SSHAgent
			ms.SetupCreds = ls.SetupCreds
			ms.AtuinSync = ls.AtuinSync
		} else if count, ok := driftTotals[pm.Hostname]; ok && count > 0 {
			ms.DriftFiles = make([]model.DriftFile, count)
		}

		ms.Tools = make(map[string]string)
		if tools, ok := toolsMap[pm.Hostname]; ok {
			for tool, installed := range tools {
				if installed {
					ms.Tools[tool] = "installed"
				}
			}
		}

		if creds, ok := credsMap[pm.Hostname]; ok {
			if creds["ssh_agent"] && ms.SSHAgent == "" {
				ms.SSHAgent = "active"
			}
			if creds["setup_creds"] && ms.SetupCreds == "" {
				ms.SetupCreds = "ran"
			}
			if creds["atuin_sync"] && ms.AtuinSync == "" {
				ms.AtuinSync = "synced"
			}
		}

		machines = append(machines, ms)
	}

	if filterMachine != "" {
		machines = filterByHostname(machines, filterMachine)
	}

	fmt.Println(display.RenderAll(machines, nil))
	return nil
}

func runCollect(cfg *config.Config, container string, verbose bool) error {
	runner := &collector.ExecRunner{}
	platform := collector.DetectPlatform()

	result, err := collector.CollectAll(runner, cfg.Hostname, platform)
	if err != nil {
		return err
	}

	ctx := context.Background()
	var pushErrors []string

	for _, ms := range result.Machines {
		msCopy := ms
		if container != "" && ms.Hostname != container {
			continue
		}
		if verbose {
			fmt.Printf("Collecting %s (%s/%s): %d drift files, tools probed\n",
				ms.Hostname, ms.Platform, ms.Context, len(ms.DriftFiles))
		}
		if err := push.Push(ctx, cfg.OTelEndpoint, &msCopy); err != nil {
			pushErrors = append(pushErrors, fmt.Sprintf("%s: %v", ms.Hostname, err))
			fmt.Fprintf(os.Stderr, "Warning: failed to push metrics for %s: %v\n", ms.Hostname, err)
		} else if verbose {
			fmt.Printf("  ✓ metrics pushed for %s\n", ms.Hostname)
		}
		if err := push.PushLog(ctx, cfg.OTelEndpoint, &msCopy); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "  Warning: failed to push log for %s: %v\n", ms.Hostname, err)
			}
		} else if verbose {
			fmt.Printf("  ✓ log pushed for %s\n", ms.Hostname)
		}
	}

	if verbose {
		fmt.Printf("\nCollection complete: %d machines\n", len(result.Machines))
		if len(pushErrors) > 0 {
			fmt.Printf("Push errors: %d\n", len(pushErrors))
		}
	}
	return nil
}

func filterByHostname(machines []model.MachineState, name string) []model.MachineState {
	var filtered []model.MachineState
	for _, m := range machines {
		if m.Hostname == name {
			filtered = append(filtered, m)
		}
	}
	return filtered
}
