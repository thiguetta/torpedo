package tests

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
	"math/rand"
)

func TestStopScheduler(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_StopScheduler.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : StopScheduler", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

var _ = Describe("{StopScheduler}", func() {
	var contexts []*scheduler.Context

	testName := "stopscheduler"
	It("has to stop scheduler service and check if applications are fine", func() {
		var err error
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("%s-%d", testName, i))...)
		}

		ValidateApplications(contexts)

		Step("get nodes for all apps in test and induce scheduler service to stop on one of the nodes", func() {
			for _, ctx := range contexts {
				var appNodes []node.Node

				Step(fmt.Sprintf("get nodes where %s app is running", ctx.App.Key), func() {
					appNodes, err = Inst().S.GetNodesForApp(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(appNodes).NotTo(BeEmpty())
				})
				randNode := rand.Intn(len(appNodes))
				appNode := appNodes[randNode]
				Step(fmt.Sprintf("stop scheduler service"), func() {
					err := Inst().S.StopSchedOnNode(appNode)
					Expect(err).NotTo(HaveOccurred())
					Step("wait for the service to stop and reschedule apps", func() {
						time.Sleep(6 * time.Minute)
					})

					Step(fmt.Sprintf("check if apps are running"), func() {
						ValidateContext(ctx)
					})
				})

				Step(fmt.Sprintf("start scheduler service"), func() {
					err := Inst().S.StartSchedOnNode(appNode)
					Expect(err).NotTo(HaveOccurred())
				})
			}
		})

		ValidateAndDestroy(contexts, nil)
	})
	JustAfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			CollectSupport()
			DescribeNamespace(contexts)
		}
	})
})

var _ = AfterSuite(func() {
	PerformSystemCheck()
	ValidateCleanup()
})

func init() {
	ParseFlags()
}
