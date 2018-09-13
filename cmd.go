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
	var kubeconfig, namespace, clusterName, config, clusterAddr string
	var verbose bool
	var interval, tokenLifetime time.Duration
	var port int
	var command = &cobra.Command{
		Use:   "controller",
		Short: "run the controller",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			initLog(verbose)

			kube := newKube(kubeconfig)
			kube.serviceAccountNamespace = namespace
			kube.cluster = clusterName
			kube.clusterDNS = clusterAddr

			cfg := &Config{}
			cfg.Load(config)
			cfg.Port = port
			log().Debugf("config: %q", cfg)
			controller := newController(kube, interval, tokenLifetime, cfg)
			controller.Run()
		},
	}
	command.Flags().StringVarP(&kubeconfig, "kube.config", "k", "", "outside cluster path to kube config")
	command.Flags().StringVarP(
		&config,
		"config", "u", "",
		"path to config. config contains user/role mapping",
	)
	command.Flags().StringVarP(&clusterAddr, "cluster.addr", "a", "", "public dns to api cluster")
	command.Flags().StringVarP(&clusterName, "cluster.name", "c", "default", "name of k8s cluster")
	command.Flags().StringVarP(
		&namespace,
		"namespace", "n",
		"default", "ns where the service accounts and cluster role bindings is created",
	)
	command.Flags().IntVarP(&port, "http.port", "p", 8080, "port to expose service on")
	command.Flags().DurationVarP(
		&interval,
		"interval", "i",
		60*time.Second,
		"the interval in which the cleanup of old token runs",
	)
	command.Flags().DurationVarP(
		&tokenLifetime,
		"token.lifetime", "l",
		1*time.Hour,
		"the time that a kube config is valid for",
	)
	command.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	command.MarkFlagRequired("cluster.addr")
	command.MarkFlagRequired("cluster.name")

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
