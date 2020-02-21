package cmd

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run torpedo with given options",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ginkgoArgs := []string{"--trace", "--timeout", fmt.Sprintf("%s", timeout)}
		if verbose {
			ginkgoArgs = append(ginkgoArgs, "-v")
		}
		if failFast {
			ginkgoArgs = append(ginkgoArgs, "--failFast")
		} else {
			ginkgoArgs = append(ginkgoArgs, "--keepGoing")
		}

		if dryRun {
			ginkgoArgs = append(ginkgoArgs, "--dryRun")
		}

		testSuiteSet := make(map[string]bool)
		if len(focusTests) != 0 {
			focusTests = fmt.Sprintf("{%s}", strings.Replace(focusTests, ",", "}|{", -1))
			ginkgoArgs = append(ginkgoArgs, fmt.Sprintf("--focus=%s", focusTests))
			for _, focusTest := range strings.Split(focusTests, "|") {
				if val, ok := testRegistry[focusTest]; ok {
					testSuiteSet[val] = true
				}
			}
		}
		if len(skipTests) != 0 {
			skipTests = fmt.Sprintf("{%s}", strings.Replace(skipTests, ",", "}|{", -1))
			ginkgoArgs = append(ginkgoArgs, fmt.Sprintf("--skip=%s", skipTests))
		}

		if len(testSuite) == 0 && len(testSuiteSet) != 0 {
			for k, _ := range testSuiteSet {
				ginkgoArgs = append(ginkgoArgs, k)
			}
		} else {
			ginkgoArgs = append(ginkgoArgs, strings.Replace(testSuite, ",", " ", -1))
		}

		testArgs := []string{"--"}

		if len(specDir) != 0 {
			testArgs = append(testArgs, fmt.Sprintf("--spec-dir=%s", specDir))
		}

		if len(appList) != 0 {
			testArgs = append(testArgs, fmt.Sprintf("--app-list=%s", appList))
		}

		if len(scheduler) != 0 {
			testArgs = append(testArgs, fmt.Sprintf("--scheduler=%s", scheduler))
		}

		if len(loglevel) != 0 {
			testArgs = append(testArgs, fmt.Sprintf("--log-level=%s", loglevel))
		}

		if len(nodeDriver) != 0 {
			testArgs = append(testArgs, fmt.Sprintf("--node-driver=%s", nodeDriver))
		}

		//if len(azureTenantID) != 0 {
		//	testArgs = append(testArgs, fmt.Sprintf("--azure-tenantid=%s", azureTenantID))
		//}
		//
		//if len(azureClientID) != 0 {
		//	testArgs = append(testArgs, fmt.Sprintf("--azure-clientid=%s", azureClientID))
		//}
		//
		//if len(azureClientSecret) != 0 {
		//	testArgs = append(testArgs, fmt.Sprintf("--azure-clientsecret=%s", azureClientSecret))
		//}

		testArgs = append(testArgs, fmt.Sprintf("--scale-factor=%d", scaleFactor))
		testArgs = append(testArgs, fmt.Sprintf("--minimum-runtime-mins=%d", minRunTime))
		testArgs = append(testArgs, fmt.Sprintf("--destroy-app-timeout=%s", appDestroyTimeout))
		testArgs = append(testArgs, fmt.Sprintf("--driver-start-timeout=%s", driverStartTimeout))
		testArgs = append(testArgs, fmt.Sprintf("--storagenode-recovery-timeout=%s", storagenodeRecoveryTimeout))
		testArgs = append(testArgs, fmt.Sprintf("--chaos-level=%d", chaosLevel))
		testArgs = append(testArgs, fmt.Sprintf("--log-location=%s", logLocation))
		testArgs = append(testArgs, fmt.Sprintf("--test-results-path=%s", testResultsPath))

		ginkgoArgs = append(ginkgoArgs, testArgs...)
		fmt.Printf("ginkgo %v\n", ginkgoArgs)

		ginkgoCmd := exec.Command("ginkgo", ginkgoArgs...)
		stdout, _ := ginkgoCmd.StdoutPipe()
		stderr, _ := ginkgoCmd.StderrPipe()
		ginkgoCmd.Start()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println(scanner.Text())

		}
		scanner = bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println(scanner.Text())

		}
		if err := ginkgoCmd.Wait(); err != nil {
			println(err.Error())
		}
	},
}

var (
	specDir         string
	logLocation     string
	testResultsPath string
	dryRun          bool
)

func init() {
	rootCmd.AddCommand(runCmd)

	// ginkgo parameters
	runCmd.Flags().DurationVarP(&timeout, "timeout", "", 720*time.Hour, "")
	runCmd.Flags().BoolVarP(&dryRun, "dry-run", "", false, "")
	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "")
	runCmd.Flags().BoolVarP(&failFast, "fail-fast", "", false, "")
	runCmd.Flags().StringVarP(&skipTests, "skip-tests", "", "", "")
	runCmd.Flags().StringVarP(&focusTests, "focus-tests", "", "", "")
	runCmd.Flags().StringVarP(&testSuite, "test-suite", "", "", "")

	// test specific parameters
	runCmd.Flags().StringVarP(&specDir, "--spec-dir", "", "../drivers/scheduler/k8s/specs", "")
	runCmd.Flags().StringVarP(&logLocation, "log-location", "", "/mnt/torpedo_support_dir", "")
	runCmd.Flags().StringVarP(&testResultsPath, "test-results-path", "", "", "/testresults")
	runCmd.Flags().StringVarP(&appList, "app-list", "", "", "")
	runCmd.Flags().IntVarP(&scaleFactor, "scale-factor", "s", 1, "")
	runCmd.Flags().StringVarP(&scheduler, "scheduler", "S", "k8s", "")
	runCmd.Flags().StringVarP(&provisioner, "provisioner", "P", "portworx", "")
	runCmd.Flags().StringVarP(&storageDriver, "storage-driver", "D", "pxd", "")
	runCmd.Flags().StringVarP(&nodeDriver, "node-driver", "", "ssh", "")
	runCmd.Flags().StringVarP(&loglevel, "log-level", "L", "debug", "")
	runCmd.Flags().IntVarP(&chaosLevel, "chaos-level", "C", 5, "")
	runCmd.Flags().IntVarP(&minRunTime, "minimum-runtime-mins", "m", 0, "")
	runCmd.Flags().StringVarP(&upgradeEndpointUrl, "storage-upgrade-endpoint-url", "", "", "")
	runCmd.Flags().StringVarP(&upgradeEndpointVersion, "storage-upgrade-endpoint-version", "", "", "")
	runCmd.Flags().StringVarP(&configmap, "config-map", "", "", "")
	runCmd.Flags().DurationVarP(&appDestroyTimeout, "destroy-app-timeout", "", 5*time.Minute, "")
	runCmd.Flags().DurationVarP(&driverStartTimeout, "driver-start-timeout", "", 5*time.Minute, "")
	runCmd.Flags().DurationVarP(&storagenodeRecoveryTimeout, "storagenode-recovery-timeout", "", 35*time.Minute, "")
	runCmd.Flags().StringVarP(&customAppConfig, "custom-config", "", "", "")
	//runCmd.Flags().StringVarP(&azureTenantID, "azure-tenantid", "", "", "")
	//runCmd.Flags().StringVarP(&azureClientID, "azure-clientid", "", "", "")
	//runCmd.Flags().StringVarP(&azureClientSecret, "azure-clientsecret", "", "", "")
}
