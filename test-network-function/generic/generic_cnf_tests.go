package generic

import (
	"fmt"
	expect "github.com/google/goexpect"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/ipaddr"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/ping"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/configuration"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

const (
	defaultNumPings       = 10
	defaultTimeoutSeconds = 20
	multusTestsKey        = "multus"
	testsKey              = "generic"
)

// The default expect.Expecter arguments;  for our purposes just enabling verbosity is enough.
var defaultExpectArgs = expect.Verbose(true)

// The default test timeout.
var defaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second

// Helper used to instantiate an OpenShift Client Session.
func getOcSession(pod, container, namespace string, timeout time.Duration, options ...expect.Option) *interactive.Oc {
	// Spawn an interactive OC shell using a goroutine (needed to avoid cross expect.Expecter interaction).  Extract the
	// Oc reference from the goroutine through a channel.  Performs basic sanity checking that the Oc session is set up
	// correctly.
	var containerOc *interactive.Oc
	ocChan := make(chan *interactive.Oc)
	var chOut <-chan error

	goExpectSpawner := interactive.NewGoExpectSpawner()
	var spawner interactive.Spawner = goExpectSpawner

	go func(chOut <-chan error) {
		oc, chOut, err := interactive.SpawnOc(&spawner, pod, container, namespace, timeout, options...)
		gomega.Expect(chOut).ToNot(gomega.BeNil())
		gomega.Expect(err).To(gomega.BeNil())
		ocChan <- oc
	}(chOut)

	// Set up a go routine which reads from the error channel
	go func() {
		err := <-chOut
		gomega.Expect(err).To(gomega.BeNil())
	}()

	containerOc = <-ocChan

	gomega.Expect(containerOc).ToNot(gomega.BeNil())

	return containerOc
}

func getOcSessions(podUnderTestName, podUnderTestContainerName, podUnderTestNamespace, partnerPodName, partnerPodContainerName, partnerPodNamespace string, timeout time.Duration, options ...expect.Option) (*interactive.Oc, *interactive.Oc) {
	podUnderTestOc := getOcSession(podUnderTestName, podUnderTestContainerName, podUnderTestNamespace, timeout, options...)
	partnerPodOc := getOcSession(partnerPodName, partnerPodContainerName, partnerPodNamespace, timeout, options...)
	return podUnderTestOc, partnerPodOc
}

// Runs the cnf-certification-generic-tests CNF test cases.
var _ = ginkgo.Describe(testsKey, func() {
	config := GetTestConfiguration()
	log.Infof("Test Configuration: %s", config)

	podUnderTestNamespace := config.PodUnderTest.Namespace
	podUnderTestName := config.PodUnderTest.Name
	podUnderTestContainerName := config.PodUnderTest.ContainerConfiguration.Name
	podUnderTestDefaultNetworkDevice := config.PodUnderTest.ContainerConfiguration.DefaultNetworkDevice

	partnerPodNamespace := config.PartnerPod.Namespace
	partnerPodName := config.PartnerPod.Name
	partnerPodContainerName := config.PartnerPod.ContainerConfiguration.Name
	partnerPodDefaultNetworkDevice := config.PartnerPod.ContainerConfiguration.DefaultNetworkDevice

	podUnderTestOc, partnerPodOc := getOcSessions(podUnderTestName, podUnderTestContainerName, podUnderTestNamespace,
		partnerPodName, partnerPodContainerName, partnerPodNamespace, defaultTimeout, defaultExpectArgs)

	// Extract the ip addresses for the pod under test and the test partner
	podUnderTestIPAddress := getContainerDefaultNetworkIPAddress(podUnderTestOc, podUnderTestDefaultNetworkDevice)
	log.Infof("%s(%s) IP Address: %s", podUnderTestName, podUnderTestContainerName, podUnderTestIPAddress)

	partnerPodIPAddress := getContainerDefaultNetworkIPAddress(partnerPodOc, partnerPodDefaultNetworkDevice)
	log.Infof("%s(%s) IP Address: %s", partnerPodName, partnerPodContainerName, partnerPodIPAddress)

	ginkgo.Context("Both Pods are on the Default network", func() {
		testNetworkConnectivity(podUnderTestOc, partnerPodOc, partnerPodIPAddress, defaultNumPings)
		testNetworkConnectivity(partnerPodOc, podUnderTestOc, podUnderTestIPAddress, defaultNumPings)
	})
})

// TODO: Multus is not applicable to every CNF, so in some regards it is CNF-specific.  On the other hand, it is likely
// a useful test across most CNFs.  Should "multus" be considered generic, cnf-specific, or somewhere in between.
var _ = ginkgo.Describe(multusTestsKey, func() {
	config := GetTestConfiguration()
	log.Infof("Test Configuration: %s", config)

	podUnderTestNamespace := config.PodUnderTest.Namespace
	podUnderTestName := config.PodUnderTest.Name
	podUnderTestContainerName := config.PodUnderTest.ContainerConfiguration.Name
	podUnderTestMultusIPAddress := config.PodUnderTest.ContainerConfiguration.MultusIPAddresses[0]

	partnerPodNamespace := config.PartnerPod.Namespace
	partnerPodName := config.PartnerPod.Name
	partnerPodContainerName := config.PartnerPod.ContainerConfiguration.Name

	podUnderTestOc, partnerPodOc := getOcSessions(podUnderTestName, podUnderTestContainerName, podUnderTestNamespace,
		partnerPodName, partnerPodContainerName, partnerPodNamespace, defaultTimeout, defaultExpectArgs)

	ginkgo.Context("Both Pods are connected via a Multus Overlay Network", func() {
		testNetworkConnectivity(partnerPodOc, podUnderTestOc, podUnderTestMultusIPAddress, defaultNumPings)
	})
})

// Helper to test that a container can ping a target IP address, and report through Ginkgo.
func testNetworkConnectivity(initiatingPodOc *interactive.Oc, targetPodOc *interactive.Oc, targetPodIPAddress string, count int) {
	ginkgo.When(fmt.Sprintf("a Ping is issued from %s(%s) to %s(%s) %s", initiatingPodOc.GetPodName(),
		initiatingPodOc.GetPodContainerName(), targetPodOc.GetPodName(), targetPodOc.GetPodContainerName(),
		targetPodIPAddress), func() {
		ginkgo.It(fmt.Sprintf("%s(%s) should reply", targetPodOc.GetPodName(), targetPodOc.GetPodContainerName()), func() {
			testPing(initiatingPodOc, targetPodIPAddress, count)
		})
	})
}

// Test that a container can ping a target IP address.
func testPing(initiatingPodOc *interactive.Oc, targetPodIPAddress string, count int) {
	log.Infof("Sending ICMP traffic(%s to %s)", initiatingPodOc.GetPodName(), targetPodIPAddress)
	pingTester := ping.NewPing(defaultTimeout, targetPodIPAddress, count)
	test, err := tnf.NewTest(initiatingPodOc.GetExpecter(), pingTester, []reel.Handler{pingTester}, initiatingPodOc.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	testResult, err := test.Run()
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	gomega.Expect(err).To(gomega.BeNil())
	transmitted, received, errors := pingTester.GetStats()
	gomega.Expect(received).To(gomega.Equal(transmitted))
	gomega.Expect(errors).To(gomega.BeZero())
}

// Extract a container ip address for a particular device.  This is needed since container default network IP address
// is served by dhcp, and thus is ephemeral.
func getContainerDefaultNetworkIPAddress(oc *interactive.Oc, dev string) string {
	ipTester := ipaddr.NewIPAddr(defaultTimeout, dev)
	test, err := tnf.NewTest(oc.GetExpecter(), ipTester, []reel.Handler{ipTester}, oc.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	testResult, err := test.Run()
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	gomega.Expect(err).To(gomega.BeNil())
	return ipTester.GetIPv4Address()
}

// GetTestConfiguration returns the cnf-certification-generic-tests test configuration.
func GetTestConfiguration() *configuration.TestConfiguration {
	config := &configuration.TestConfiguration{}
	ginkgo.Context("Instantiate some configuration information from the environment", func() {
		yamlFile, err := ioutil.ReadFile(configuration.GetConfigurationFilePathFromEnvironment())
		gomega.Expect(err).To(gomega.BeNil())
		err = yaml.Unmarshal(yamlFile, config)
		gomega.Expect(err).To(gomega.BeNil())
	})
	return config
}
