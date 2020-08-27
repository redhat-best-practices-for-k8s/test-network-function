package generic

import (
	"fmt"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/configuration"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	defaultNumPings = 10
	defaultTnfTimeout = 20
	GenericTestsKey = "generic"
	MultusTestsKey = "multus"
	openshiftNamespaceArgument = "-n"
)

// Runs the cnf-certification-generic-tests CNF test cases.
var _ = ginkgo.Describe(GenericTestsKey, func() {
	config := GetTestConfiguration()
	log.Infof("Test Configuration: %s", config)

	podUnderTestNamespaceArgs := CreateNamespaceArgs(config.PodUnderTest.Namespace)
	podUnderTestName := config.PodUnderTest.Name
	podUnderTestContainerName := config.PodUnderTest.ContainerConfiguration.Name
	podUnderTestDefaultNetworkDevice := config.PodUnderTest.ContainerConfiguration.DefaultNetworkDevice

	partnerPodNamespaceArgs := CreateNamespaceArgs(config.PartnerPod.Namespace)
	partnerPodName := config.PartnerPod.Name
	partnerPodContainerName := config.PartnerPod.ContainerConfiguration.Name
	partnerPodDefaultNetworkDevice := config.PartnerPod.ContainerConfiguration.DefaultNetworkDevice

	// Extract the ip addresses for the pod under test and the test partner
	podUnderTestIpAddress, err := getContainerDefaultNetworkIpAddress(podUnderTestName, podUnderTestContainerName,
		podUnderTestDefaultNetworkDevice, podUnderTestNamespaceArgs)
	gomega.Expect(err).To(gomega.BeNil())
	log.Infof("%s(%s) IP Address: %s", podUnderTestName, podUnderTestContainerName, podUnderTestIpAddress)

	partnerPodIpAddress, err := getContainerDefaultNetworkIpAddress(partnerPodName, partnerPodContainerName,
		partnerPodDefaultNetworkDevice, partnerPodNamespaceArgs)
	gomega.Expect(err).To(gomega.BeNil())
	log.Infof("%s(%s) IP Address: %s", partnerPodName, partnerPodContainerName, partnerPodIpAddress)

	ginkgo.Context("Both Pods are on the Default network", func() {
		testNetworkConnectivity(partnerPodName, partnerPodContainerName, podUnderTestName,
			podUnderTestContainerName, podUnderTestIpAddress, partnerPodNamespaceArgs, defaultNumPings)
		testNetworkConnectivity(podUnderTestName, podUnderTestContainerName, partnerPodName,
			partnerPodContainerName, partnerPodIpAddress, podUnderTestNamespaceArgs, defaultNumPings)
	})
})

// TODO: Multus is not applicable to every CNF, so in some regards it is CNF-specific.  On the other hand, it is likely
// a useful test across most CNFs.  Should "multus" be considered generic, cnf-specific, or somewhere in between.
var _ = ginkgo.Describe(MultusTestsKey, func() {
	config := GetTestConfiguration()
	log.Infof("Test Configuration: %s", config)

	podUnderTestName := config.PodUnderTest.Name
	podUnderTestContainerName := config.PodUnderTest.ContainerConfiguration.Name
	podUnderTestMultusIpAddress := config.PodUnderTest.ContainerConfiguration.MultusIpAddresses[0]

	partnerPodNamespaceArgs := CreateNamespaceArgs(config.PartnerPod.Namespace)
	partnerPodName := config.PartnerPod.Name
	partnerPodContainerName := config.PartnerPod.ContainerConfiguration.Name
	ginkgo.Context("Both Pods are connected via a Multus Overlay Network", func() {
		testNetworkConnectivity(partnerPodName, partnerPodContainerName, podUnderTestName,
			podUnderTestContainerName, podUnderTestMultusIpAddress, partnerPodNamespaceArgs, defaultNumPings)
	})
})

// Helper to test that a container can ping a target IP address, and report through Ginkgo.
func testNetworkConnectivity(initiatingPodName, initiatingPodContainerName, targetPodName,
	targetPodContainerName, targetPodIpAddress string, initiatingPodNamespaceArgs []string, count int) {
	ginkgo.When(fmt.Sprintf("a Ping is issued from %s(%s) to %s(%s) %s", initiatingPodName,
		initiatingPodContainerName, targetPodName, targetPodContainerName, targetPodIpAddress), func() {
		ginkgo.It(fmt.Sprintf("%s(%s) should reply", targetPodName, targetPodContainerName), func() {
			testPing(initiatingPodName, initiatingPodContainerName, targetPodIpAddress,
				initiatingPodNamespaceArgs, count)
		})
	})
}

// Test that a container can ping a target IP address.
func testPing(initiatingPodName, initiatingPodContainerName, targetPodIpAddress string,
	initiatingPodNamespaceArgs []string, count int) {
	ocPing := tnf.NewOcPing(defaultTnfTimeout, initiatingPodName, initiatingPodContainerName, targetPodIpAddress,
		initiatingPodNamespaceArgs, count)
	printer := reel.NewPrinter("")
	test, err := tnf.NewTest("", ocPing, []reel.Handler{printer, ocPing})
	gomega.Expect(err).To(gomega.BeNil())
	if err == nil {
		result, err := test.Run()
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result).To(gomega.Equal(tnf.SUCCESS))
	}
}

// Extract a container ip address for a particular device.  This is needed since container default network IP address
// is served by dhcp, and thus is ephemeral.
func getContainerDefaultNetworkIpAddress(pod, container string, device string, args []string) (string, error) {
	logfile := ""
	containerIpAddress := tnf.NewIpAddr(2, pod, container, device, args)
	printer := reel.NewPrinter("")
	test, err := tnf.NewTest(logfile, containerIpAddress, []reel.Handler{printer, containerIpAddress})
	gomega.Expect(err).To(gomega.BeNil())
	if err == nil {
		result, err := test.Run()
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result).To(gomega.Equal(tnf.SUCCESS))
		return containerIpAddress.GetAddr(), nil
	}
	return "", err
}

// Get cnf-certification-generic-tests test configuration.
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

// Helper to create namespace args for an OC command.
func CreateNamespaceArgs(namespace string) []string {
	return []string{openshiftNamespaceArgument, namespace}
}
