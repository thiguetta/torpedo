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

var rootCmd = &cobra.Command{
	Use:   "torpedo",
	Short: "A brief description of your application",
	Long: `Torpedo is a test suite to qualify storage providers for stateful containers 
running in a distributed environment.

It tests various scenarios that applications encounter when running in Linux containers 
and deployed via schedulers such as Kubernetes, Marathon or Swarm.`,
}

var (
	testRegistry map[string]string
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generateTestRegistry() {
	files, err := ioutil.ReadDir("bin/")
	if err != nil && !os.IsNotExist(err) {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, file := range files {
		if !file.IsDir() {
			ginkgoCmd := exec.Command("ginkgo", "--trace", "--dryRun", "-v",
				fmt.Sprintf("bin/%s", file.Name()), "--", "--log-location", ".")
			stdout, _ := ginkgoCmd.StdoutPipe()
			ginkgoCmd.Start()
			specNameRegex := regexp.MustCompile("{.*}")
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				if specNameRegex.MatchString(line) {
					testRegistry[specNameRegex.FindString(line)] = "bin/" + file.Name()
				}
			}
			ginkgoCmd.Wait()
		}
	}
}

func init() {
	testRegistry = make(map[string]string)
	generateTestRegistry()
}
