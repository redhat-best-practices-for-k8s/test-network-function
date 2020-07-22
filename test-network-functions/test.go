package setup

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	testConfigurationFileName = "conf.yaml"
)

var _ = ginkgo.Describe("Test ICMP On Default Network", func() {
	config := &TestConfiguration{}
	ginkgo.Context("Grab some configuration information from the environment", func() {
		yamlFile, err := ioutil.ReadFile(testConfigurationFileName)
		gomega.Expect(err).To(gomega.BeNil())
		err = yaml.Unmarshal(yamlFile, config)
		gomega.Expect(err).To(gomega.BeNil())
	})

	//result := tnf.ERROR
	//logfile := ""
	ipaddr := tnf.NewIpAddr(2, "eth0")
	//printer := reel.NewPrinter("")
	//test, err := tnf.NewTest(logfile, ipaddr, []reel.Handler{printer, ipaddr})
	//if err == nil {
	//	result, err = test.Run()
	//}
	//gomega.Expect(err).To(gomega.BeNil())
	//gomega.Expect(result).To(gomega.Equal(tnf.SUCCESS))
	//ipAddress := ipaddr.GetAddr()
	//log.Info(ipAddress)

	ginkgo.Context("Both Pods are on the Default network", func() {
		ginkgo.When("When a Ping is initiated from IKEster to the container under test on the Default network", func() {
			ginkgo.It("The Container Under Test should reply", func() {
				args := []string{"-n", config.PartnerContainer.Namespace}
				oc := tnf.NewOc(10, config.PartnerContainer.Name, args)
				printer := reel.NewPrinter(" \r\n")
				//file, err := os.Open("ping_multus.json")
				//if err != nil {
				//	log.Fatal(err)
				//}
				feeder := ipaddr //tnf.NewTestFeeder(10, tnf.OcPrompt, bufio.NewScanner(file))
				chain := []reel.Handler{printer, feeder, oc}
				test, err := tnf.NewTest("", oc, chain)
				if err == nil {
					result, err := test.Run()
					gomega.Expect(err).Should(gomega.BeNil())
					gomega.Expect(result).Should(gomega.BeZero())
				}
			})
		})
	})
})
