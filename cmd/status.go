package cmd

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "gets torpedo current status",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		pod, err := core.Instance().GetPodByName("torpedo", "default")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("%s status: %s reason: %s message: %s\n", pod.Name, pod.Status.Phase, pod.Status.Reason,
			pod.Status.Message)
		fmt.Println("Events:")
		fields := fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s", pod.Name)
		events, err := core.Instance().ListEvents("default", metav1.ListOptions{FieldSelector: fields})
		if err != nil {
			fmt.Printf("failed to list events for pod %s. Cause: %v", pod.Name, err)
		}
		for _, event := range events.Items {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", event.Type, event.Source.Component, event.Source.Host,
				event.Reason, event.Message)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
