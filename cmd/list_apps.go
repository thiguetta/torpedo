package cmd

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/apps"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "list all apps created by torpedo",
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
			fmt.Printf("Listing torpedo app components from namespace: %s\n", ns.Name)

			fmt.Println("***** Pods ***** ")
			pods, err := core.Instance().GetPods(ns.Name, map[string]string{})
			if err != nil {
				fmt.Printf("failed to get pods for namespace: %s\n", ns.Name)
			}
			for _, pod := range pods.Items {
				fmt.Printf("%s %s %s %s\n", pod.Name, pod.Status.Phase, pod.Status.Reason, pod.Status.Message)
				fmt.Println("Events:")
				fields := fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s", pod.Name)
				events, err := core.Instance().ListEvents(ns.Name, metav1.ListOptions{FieldSelector: fields})
				if err != nil {
					fmt.Printf("failed to list events for pod %s. Cause: %v", pod.Name, err)
				}
				for _, event := range events.Items {
					fmt.Printf("%s\t%s\t%s\t%s\t%s\n", event.Type, event.Source.Component, event.Source.Host, event.Reason, event.Message)
				}
			}
			fmt.Printf("\n\n")

			fmt.Println("***** Deployments ***** ")
			deployments, err := apps.Instance().ListDeployments(ns.Name, metav1.ListOptions{})
			if err != nil {
				fmt.Printf("failed to get deployments for namespace: %s\n", ns.Name)
			}
			for _, deployment := range deployments.Items {
				fmt.Printf("%s %d %d\n", deployment.Name, deployment.Status.Replicas, deployment.Status.ReadyReplicas)
				fmt.Println("Events:")
				fields := fmt.Sprintf("involvedObject.kind=Deployment,involvedObject.name=%s", deployment.Name)
				events, err := core.Instance().ListEvents(ns.Name, metav1.ListOptions{FieldSelector: fields})
				if err != nil {
					fmt.Printf("failed to list events for pod %s. Cause: %v", deployment.Name, err)
				}
				for _, event := range events.Items {
					fmt.Printf("%s\t%s\t%s\t%s\t%s\n", event.Type, event.Source.Component, event.Source.Host, event.Reason, event.Message)
				}
			}
			fmt.Printf("\n\n")

			fmt.Println("***** Statefulsets ***** ")
			statefulsets, err := apps.Instance().ListStatefulSets(ns.Name)
			if err != nil {
				fmt.Printf("failed to get statefulsets for namespace: %s\n", ns.Name)
			}
			for _, statefulset := range statefulsets.Items {
				fmt.Printf("%s %d %d\n", statefulset.Name, statefulset.Status.Replicas, statefulset.Status.ReadyReplicas)
				fmt.Println("Events:")
				fields := fmt.Sprintf("involvedObject.kind=StatefulSet,involvedObject.name=%s", statefulset.Name)
				events, err := core.Instance().ListEvents(ns.Name, metav1.ListOptions{FieldSelector: fields})
				if err != nil {
					fmt.Printf("failed to list events for pod %s. Cause: %v", statefulset.Name, err)
				}
				for _, event := range events.Items {
					fmt.Printf("%s\t%s\t%s\t%s\t%s\n", event.Type, event.Source.Component, event.Source.Host, event.Reason, event.Message)
				}
			}
			fmt.Printf("\n\n")

			fmt.Println("***** Services ***** ")
			services, err := core.Instance().ListServices(ns.Name, metav1.ListOptions{})
			if err != nil {
				fmt.Printf("failed to get services for namespace: %s\n", ns.Name)
			}
			for _, service := range services.Items {
				fmt.Printf("%s %+v %+v\n", service.Name, service.Spec.Ports, service.Status.LoadBalancer)
				fmt.Println("Events:")
				fields := fmt.Sprintf("involvedObject.kind=Service,involvedObject.name=%s", service.Name)
				events, err := core.Instance().ListEvents(ns.Name, metav1.ListOptions{FieldSelector: fields})
				if err != nil {
					fmt.Printf("failed to list events for pod %s. Cause: %v", service.Name, err)
				}
				for _, event := range events.Items {
					fmt.Printf("%s\t%s\t%s\t%s\t%s\n", event.Type, event.Source.Component, event.Source.Host, event.Reason, event.Message)
				}
			}
			fmt.Printf("\n\n")
		}
	},
}

func init() {
	listCmd.AddCommand(appsCmd)
}
