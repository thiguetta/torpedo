package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "attach to torpedo pod stdout",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		attachTorpedoStdout()
	},
}

func attachTorpedoStdout() {
	pod, err := core.Instance().GetPodByName("torpedo", "default")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if pod.Status.Phase != corev1.PodRunning {
		fmt.Printf("torpedo pod is not running. State: %s\n", pod.Status.Phase)
		os.Exit(1)
	}
	torpedoLogsCmd := exec.Command("kubectl", "logs", "-f", "torpedo")
	stdout, _ := torpedoLogsCmd.StdoutPipe()
	if err = torpedoLogsCmd.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err = torpedoLogsCmd.Wait(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(attachCmd)
}
