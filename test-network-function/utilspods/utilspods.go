package utilspods

import(
"time"
"github.com/test-network-function/test-network-function/pkg/config/configsections"
"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
"github.com/onsi/gomega"
)

const (
	defaultTimeoutSeconds         = 10
	defaultTerminationGracePeriod = 30
	drainTimeoutMinutes           = 5
)


// containersToExcludeFromConnectivityTests is a set used for storing the containers that should be excluded from
// connectivity testing.
var containersToExcludeFromConnectivityTests = make(map[configsections.ContainerIdentifier]interface{})


// container is an internal construct which follows the Container design pattern.  Essentially, a container holds the
// pertinent information to perform a test against or using an Operating System container.  This includes facets such
// as the reference to the interactive.Oc instance, the reference to the test configuration, and the default network
// IP address.
type container struct {
	containerConfiguration  configsections.Container
	oc                      *interactive.Oc
	defaultNetworkIPAddress string
	containerIdentifier     configsections.ContainerIdentifier
}

// The default test timeout.
var defaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second

var drainTimeout = time.Duration(drainTimeoutMinutes) * time.Minute

// Helper used to instantiate an OpenShift Client Session.
func getOcSession(pod, container, namespace string, timeout time.Duration, options ...interactive.Option) *interactive.Oc {
	// Spawn an interactive OC shell using a goroutine (needed to avoid cross expect.Expecter interaction).  Extract the
	// Oc reference from the goroutine through a channel.  Performs basic sanity checking that the Oc session is set up
	// correctly.
	var containerOc *interactive.Oc
	ocChan := make(chan *interactive.Oc)
	var chOut <-chan error

	goExpectSpawner := interactive.NewGoExpectSpawner()
	var spawner interactive.Spawner = goExpectSpawner

	go func() {
		oc, outCh, err := interactive.SpawnOc(&spawner, pod, container, namespace, timeout, options...)
		gomega.Expect(outCh).ToNot(gomega.BeNil())
		gomega.Expect(err).To(gomega.BeNil())
		ocChan <- oc
	}()

	// Set up a go routine which reads from the error channel
	go func() {
		err := <-chOut
		gomega.Expect(err).To(gomega.BeNil())
	}()

	containerOc = <-ocChan

	gomega.Expect(containerOc).ToNot(gomega.BeNil())

	return containerOc
}
// createContainers contains the general steps involved in creating "oc" sessions and other configuration. A map of the
// aggregate information is returned.
func createContainers(containerDefinitions []configsections.Container) map[configsections.ContainerIdentifier]*container {
	createdContainers := make(map[configsections.ContainerIdentifier]*container)
	for _, c := range containerDefinitions {
		oc := getOcSession(c.PodName, c.ContainerName, c.Namespace, defaultTimeout, interactive.Verbose(true))
		var defaultIPAddress = "UNKNOWN"
		if _, ok := containersToExcludeFromConnectivityTests[c.ContainerIdentifier]; !ok {
			defaultIPAddress = getContainerDefaultNetworkIPAddress(oc, c.DefaultNetworkDevice)
		}
		createdContainers[c.ContainerIdentifier] = &container{
			containerConfiguration:  c,
			oc:                      oc,
			defaultNetworkIPAddress: defaultIPAddress,
			containerIdentifier:     c.ContainerIdentifier,
		}
	}
	return createdContainers
}

// createContainersUnderTest sets up the test containers.
func createContainersUnderTest(conf *configsections.TestConfiguration) map[configsections.ContainerIdentifier]*container {
	return createContainers(conf.ContainersUnderTest)
}

// createPartnerContainers sets up the partner containers.
func createPartnerContainers(conf *configsections.TestConfiguration) map[configsections.ContainerIdentifier]*container {
	return createContainers(conf.PartnerContainers)
}

// GetTestConfiguration returns the cnf-certification-generic-tests test configuration.
func GetTestConfiguration() *configsections.TestConfiguration {
	conf := config.GetConfigInstance()
	return &conf.Generic
}


func getContext() *interactive.Context {
	context, err := interactive.SpawnShell(interactive.CreateGoExpectSpawner(), defaultTimeout, interactive.Verbose(true))
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(context).ToNot(gomega.BeNil())
	return context
}
func isMinikube() bool {
	b, _ := strconv.ParseBool(os.Getenv("TNF_MINIKUBE_ONLY"))
	return b
}