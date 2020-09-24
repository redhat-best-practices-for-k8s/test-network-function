package cnf

import (
	"fmt"
	expect "github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/cnf-specific/casa/cnf/nrf"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/cnf-specific/casa/configuration"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	testTimeout = time.Second * 10
)

var _ = Describe("casa-cnf", func() {

	var config *configuration.CasaCNFConfiguration
	var err error
	config, err = configuration.GetCasaCNFTestConfiguration()
	log.Info("Casa CNF Specific Configuration: %s", config)
	Expect(err).To(BeNil())
	Expect(config).ToNot(BeNil())

	cnfTypes := config.CNFTypes
	nrfName := config.NRFName
	namespace := config.Namespace

	var context *interactive.Context
	When("A local shell is spawned", func() {
		goExpectSpawner := interactive.NewGoExpectSpawner()
		var spawner interactive.Spawner = goExpectSpawner
		context, err = interactive.SpawnShell(&spawner, testTimeout, expect.Verbose(true))
		Expect(err).To(BeNil())
		Expect(context).ToNot(BeNil())
		Expect(context.GetExpecter()).ToNot(BeNil())
	})

	var nrfs map[string]*nrf.NRFID
	When("Registrations are polled from the \"nfregistrations.mgmt.casa.io\" Custom Resource", func() {
		It("The appropriate registrations should be reported", func() {
			registrationTest := nrf.NewRegistration(testTimeout, namespace)
			Expect(registrationTest).ToNot(BeNil())
			test, err := tnf.NewTest(context.GetExpecter(), registrationTest, []reel.Handler{registrationTest}, context.GetErrorChannel())
			Expect(err).To(BeNil())
			Expect(test).ToNot(BeNil())
			testResult, err := test.Run()
			Expect(err).To(BeNil())
			Expect(testResult).To(Equal(tnf.SUCCESS))
			nrfs = registrationTest.GetRegisteredNRFs()
			log.Infof("nrfs=%s", nrfs)
			for _, cnfType := range cnfTypes {
				nrfInstalled := getNRF(nrfs, cnfType)
				Expect(nrfInstalled).ToNot(BeNil())
			}
		})
	})

	for _, cnfType := range cnfTypes {
		When(fmt.Sprintf("%s(%s) is checked for registration", nrfName, cnfType), func() {
			It("Should be registered", func() {
				for _, cnfType := range cnfTypes {
					specificNrf := getNRF(nrfs, cnfType)
					Expect(specificNrf).ToNot(BeNil())
					checkRegistrationTest := nrf.NewCheckRegistration(namespace, testTimeout, specificNrf)
					test, err := tnf.NewTest(context.GetExpecter(), checkRegistrationTest, []reel.Handler{checkRegistrationTest}, context.GetErrorChannel())
					Expect(err).To(BeNil())
					Expect(test).ToNot(BeNil())
					result, err := test.Run()
					Expect(err).To(BeNil())
					Expect(result).To(Equal(tnf.SUCCESS))
				}
			})
		})
	}
})

// getNRF returns the first NRFID for a given CNF Type (i.e., AMF, SMF, etc).
func getNRF(nrfs map[string]*nrf.NRFID, cnfType string) *nrf.NRFID {
	for _, nrf := range nrfs {
		if nrf.GetType() == cnfType {
			return nrf
		}
	}
	return nil
}
