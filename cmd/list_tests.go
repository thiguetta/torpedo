package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"

	"github.com/spf13/cobra"
)

var testsCmd = &cobra.Command{
	Use:   "tests",
	Short: "list all available test suites",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		files, err := ioutil.ReadDir("bin/")
		if err != nil && os.IsNotExist(err) {
			fmt.Println("no test suites available")
			os.Exit(0)
		} else if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, file := range files {
			if !file.IsDir() {
				fmt.Printf("File: %s\n", file.Name())
				ginkgoCmd := exec.Command("ginkgo", "--trace", "--dryRun", "-v",
					fmt.Sprintf("bin/%s", file.Name()), "--", "--log-location", ".")
				stdout, _ := ginkgoCmd.StdoutPipe()
				ginkgoCmd.Start()
				availableSpecs := make([]string, 0)
				specNameRegex := regexp.MustCompile("{.*}")
				suiteNameRegex := regexp.MustCompile(`Running Suite: Torpedo : (\w+)`)
				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					line := scanner.Text()
					if specNameRegex.MatchString(line) {
						availableSpecs = append(availableSpecs, specNameRegex.FindString(line))
					}
					if suiteNameRegex.MatchString(line) {
						fmt.Printf("Suite: %s\n", suiteNameRegex.FindStringSubmatch(line)[1])
					}
				}
				ginkgoCmd.Wait()
				fmt.Println("Specs: ")
				for _, availableSpec := range availableSpecs {
					fmt.Printf("\t- %s\n", availableSpec)
				}
				fmt.Println()

			}
		}
	},
}

func init() {
	listCmd.AddCommand(testsCmd)
}
