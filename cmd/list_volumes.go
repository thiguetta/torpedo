package cmd

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/spf13/cobra"
)

var volumesCmd = &cobra.Command{
	Use:   "volumes",
	Short: "list all volumes created by torpedo apps",
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
			fmt.Printf("Listing torpedo volumes from namespace: %s\n", ns.Name)
			pvcs, err := core.Instance().GetPersistentVolumeClaims(ns.Name, map[string]string{})
			if err != nil {
				fmt.Printf("failed to get PVCs. cause: %v\n", err)
				os.Exit(1)
			}
			for _, pvc := range pvcs.Items {
				fmt.Printf("%s %s %v %s\n", pvc.Name, pvc.Status.Phase, pvc.Status.Capacity, pvc.Status.AccessModes)
				fields := fmt.Sprintf("involvedObject.kind=PersistentVolumeClaim,involvedObject.name=%s",
					pvc.Name)
				events, err := core.Instance().ListEvents(ns.Name, metav1.ListOptions{FieldSelector: fields})
				if err != nil {
					fmt.Printf("failed to list events for pod %s. Cause: %v", pvc.Name, err)
				}
				for _, event := range events.Items {
					fmt.Printf("%s\t%s\t%s\t%s\t%s\n", event.Type, event.Source.Component, event.Source.Host,
						event.Reason, event.Message)
				}
				pv, err := core.Instance().GetPersistentVolume(pvc.Spec.VolumeName)
				if err != nil {
					fmt.Printf("failed to get pv %s. cause: %v\n", pvc.Spec.VolumeName, err)
					os.Exit(1)
				}
				fmt.Printf("%s %s %s %s\n", pv.Name, pv.Status.Phase, pv.Status.Reason, pv.Status.Message)
				fields = fmt.Sprintf("involvedObject.kind=PersistentVolume,involvedObject.name=%s", pv.Name)
				events, err = core.Instance().ListEvents(ns.Name, metav1.ListOptions{FieldSelector: fields})
				if err != nil {
					fmt.Printf("failed to list events for pod %s. Cause: %v", pv.Name, err)
				}
				for _, event := range events.Items {
					fmt.Printf("%s\t%s\t%s\t%s\t%s\n", event.Type, event.Source.Component, event.Source.Host,
						event.Reason, event.Message)
				}
				fmt.Printf("\n\n\n")
			}
		}
	},
}

func init() {
	listCmd.AddCommand(volumesCmd)
}
