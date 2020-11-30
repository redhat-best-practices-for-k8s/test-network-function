// Copyright (C) 2020 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package cnf

import (
	"fmt"
	expect "github.com/google/goexpect"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/cnf_specific/casa/cnf/nrf"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/cnf_specific/casa/configuration"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	testTimeout = time.Second * 10
)

var _ = ginkgo.Describe("casa_cnf", func() {

	var config *configuration.CasaCNFConfiguration
	var err error
	config, err = configuration.GetCasaCNFTestConfiguration()
	log.Info("Casa CNF Specific Configuration: %s", config)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(config).ToNot(gomega.BeNil())

	cnfTypes := config.CNFTypes
	nrfName := config.NRFName
	namespace := config.Namespace

	var context *interactive.Context
	ginkgo.When("A local shell is spawned", func() {
		goExpectSpawner := interactive.NewGoExpectSpawner()
		var spawner interactive.Spawner = goExpectSpawner
		context, err = interactive.SpawnShell(&spawner, testTimeout, expect.Verbose(true))
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(context).ToNot(gomega.BeNil())
		gomega.Expect(context.GetExpecter()).ToNot(gomega.BeNil())
	})

	var nrfs map[string]*nrf.ID
	ginkgo.When("Registrations are polled from the \"nfregistrations.mgmt.casa.io\" Custom Resource", func() {
		ginkgo.It("The appropriate registrations should be reported", func() {
			registrationTest := nrf.NewRegistration(testTimeout, namespace)
			gomega.Expect(registrationTest).ToNot(gomega.BeNil())
			test, err := tnf.NewTest(context.GetExpecter(), registrationTest, []reel.Handler{registrationTest}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(test).ToNot(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			nrfs = registrationTest.GetRegisteredNRFs()
			log.Infof("nrfs=%s", nrfs)
			for _, cnfType := range cnfTypes {
				nrfInstalled := getNRF(nrfs, cnfType)
				gomega.Expect(nrfInstalled).ToNot(gomega.BeNil())
			}
		})
	})

	for _, cnfType := range cnfTypes {
		ginkgo.When(fmt.Sprintf("%s(%s) is checked for registration", nrfName, cnfType), func() {
			ginkgo.It("Should be registered", func() {
				for _, cnfType := range cnfTypes {
					specificNrf := getNRF(nrfs, cnfType)
					gomega.Expect(specificNrf).ToNot(gomega.BeNil())
					checkRegistrationTest := nrf.NewCheckRegistration(namespace, testTimeout, specificNrf)
					test, err := tnf.NewTest(context.GetExpecter(), checkRegistrationTest, []reel.Handler{checkRegistrationTest}, context.GetErrorChannel())
					gomega.Expect(err).To(gomega.BeNil())
					gomega.Expect(test).ToNot(gomega.BeNil())
					result, err := test.Run()
					gomega.Expect(err).To(gomega.BeNil())
					gomega.Expect(result).To(gomega.Equal(tnf.SUCCESS))
				}
			})
		})
	}
})

// getNRF returns the first ID for a given CNF Type (i.e., AMF, SMF, etc).
func getNRF(nrfs map[string]*nrf.ID, cnfType string) *nrf.ID {
	for _, nrf := range nrfs {
		if nrf.GetType() == cnfType {
			return nrf
		}
	}
	return nil
}
