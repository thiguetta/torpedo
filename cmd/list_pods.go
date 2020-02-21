package cmd

import (
	"fmt"
	"os"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var podsCmd = &cobra.Command{
	Use:   "pods",
	Short: "list all pods created by torpedo",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		namespaces, err := core.Instance().ListNamespaces(map[string]string{"creator": "torpedo"})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, ns := range namespaces.Items {
			fmt.Printf("Listing torpedo pods from namespace: %s\n\n", ns.Name)
			pods, err := core.Instance().GetPods(ns.Name, map[string]string{})
			if err != nil {
				fmt.Printf("failed to get pods for namespace: %s\n", ns.Name)
			}
			for _, pod := range pods.Items {
				fmt.Printf("%s %s %s %s\n\n", pod.Name, pod.Status.Phase, pod.Status.Reason, pod.Status.Message)
				fmt.Println("Events:")
				fields := fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s", pod.Name)
				events, err := core.Instance().ListEvents(ns.Name, metav1.ListOptions{FieldSelector: fields})
				if err != nil {
					fmt.Printf("failed to list events for pod %s. Cause: %v", pod.Name, err)
				}
				for _, event := range events.Items {
					fmt.Printf("%s\t%s\t%s\t%s\t%s\n", event.Type, event.Source.Component, event.Source.Host, event.Reason, event.Message)
				}
				fmt.Printf("\n\n\n")
			}
		}
	},
}

func init() {
	listCmd.AddCommand(podsCmd)
}
