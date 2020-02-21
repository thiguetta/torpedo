package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "gets logs from torpedo latest run",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		torpedoLogsCmd, err := exec.Command("kubectl", "logs", "torpedo").Output()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Print(string(torpedoLogsCmd))
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}
