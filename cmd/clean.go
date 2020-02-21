package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/autopilot"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean all torpedo resources",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		if err := core.Instance().DeleteNamespace("chaos-testing"); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		exec.Command("kubectl", "apply", "-f", "manifests/").Output()

		namespaces, err := core.Instance().ListNamespaces(map[string]string{"creator": "torpedo"})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, ns := range namespaces.Items {
			fmt.Printf("Cleaning up torpedo from namespace: %s\n", ns.Name)
			if err = core.Instance().DeleteNamespace(ns.Name); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		err = apps.Instance().DeleteDaemonSet("debug", "kube-system")
		// apps doesn't return proper grpc error, so we need to dig error message
		if err != nil && !strings.Contains(err.Error(), "not found") {
			fmt.Println(err)
			os.Exit(1)
		}

		autopilotRules, err := autopilot.Instance().ListAutopilotRules()
		// autopilot doesn't return proper grpc error, so we need to dig error message
		if err != nil && !strings.Contains(err.Error(), "could not find") {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, autopilotRule := range autopilotRules.Items {
			for labelKey, labelValue := range autopilotRule.Labels {
				if labelKey == "creator" && labelValue == "torpedo" {
					fmt.Printf("Cleaning up autopilot rule %s\n", autopilotRule.Name)
					if err = autopilot.Instance().DeleteAutopilotRule(autopilotRule.Name); err != nil {
						logrus.Errorf("failed to delete autopilot rule %s", autopilotRule.Name)
					}
					break
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
