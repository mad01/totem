package main

import (
	"fmt"

	"time"

	"github.com/spf13/cobra"
)

func cmdVersion() *cobra.Command {
	var command = &cobra.Command{
		Use:   "version",
		Short: "get version",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(getVersion())
		},
	}
	return command
}

func cmdRunController() *cobra.Command {
	var kubeconfig string
	var verbose bool
	var interval time.Duration
	var port int
	var command = &cobra.Command{
		Use:   "controller",
		Short: "run the controller",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			initLog(verbose)

			kube := newKube(kubeconfig)

			controller := newController(kube, interval, port)
			controller.Run()
		},
	}
	command.Flags().StringVarP(&kubeconfig, "kube.config", "k", "", "outside cluster path to kube config")
	command.Flags().IntVarP(&port, "http.port", "p", 8080, "http server port")
	command.Flags().DurationVarP(
		&interval, "interval.controller", "i", 10*time.Second, "controller update interaval for internal k8s caches"
		)
	command.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	return command
}

func runCmd() error {
	var rootCmd = &cobra.Command{Use: "totem"}
	rootCmd.AddCommand(cmdVersion())
	rootCmd.AddCommand(cmdRunController())

	err := rootCmd.Execute()
	if err != nil {
		return fmt.Errorf("%v", err.Error())
	}
	return nil
}
