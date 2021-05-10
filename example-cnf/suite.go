package example_cnf

import (
	"bytes"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/example-cnf/cr"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"path"
	"time"
)

const (
	testCustomResourceCRType           = "trexapp"
	testCustomResourceCRName           = "trex-app-cnf-cert"
	testCustomResourceFileName         = "test-cr.yaml"
	testCustomResourceNamespace        = "example-cnf"
	testCustomResourceTemplateFileName = "test-cr.yaml.tpl"
	testCustomResourceValuesFileName   = "test-cr.values.yaml"
)

var (
	testTimeout = time.Second * 20
)

func getPath(filename string) string {
	return path.Join("..", "example-cnf", filename)
}

func createCustomResource() ([]byte, error) {
	tplBytes, err := ioutil.ReadFile(getPath(testCustomResourceValuesFileName))
	if err != nil {
		return nil, err
	}

	values := make(map[string]interface{})
	err = yaml.Unmarshal(tplBytes, values)

	if err != nil {
		return nil, err
	}

	templateBytes, err := ioutil.ReadFile(getPath(testCustomResourceTemplateFileName))
	if err != nil {
		return nil, err
	}
	// Note: "tpl" just names the template.  It is arbitrary, and doesn't really matter.
	t, err := template.New("tpl").Option("missingkey=error").Parse(string(templateBytes))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "tpl", values)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

var _ = ginkgo.Context("example-cnf", func() {
	ginkgo.When("a test Custom Resource is rendered", func() {
		ginkgo.It("should render correctly", func() {
			inputBytes, err := createCustomResource()
			log.Error(string(inputBytes))
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(inputBytes).ToNot(gomega.BeNil())
			err = ioutil.WriteFile(getPath(testCustomResourceFileName), inputBytes, 0644)
			gomega.Expect(err).To(gomega.BeNil())

			ginkgo.When("a test Custom Resource is created", func() {
				var spawner = interactive.NewGoExpectSpawner()
				var goExpectSpawner interactive.Spawner = spawner
				context, err := interactive.SpawnShell(&goExpectSpawner, testTimeout, interactive.Verbose(true))
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(context).ToNot(gomega.BeNil())
				tester := cr.NewCreate(getPath(testCustomResourceFileName), testTimeout)
				test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(test).ToNot(gomega.BeNil())
				result, err := test.Run()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(result).To(gomega.Equal(tnf.SUCCESS))

				time.Sleep(time.Second * 40)

				trafficPassed := cr.NewTraffic(testCustomResourceNamespace, testCustomResourceCRType, testCustomResourceCRName, testTimeout)
				test, err = tnf.NewTest(context.GetExpecter(), trafficPassed, []reel.Handler{trafficPassed}, context.GetErrorChannel())
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(test).ToNot(gomega.BeNil())
				result, err = test.Run()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(test).ToNot(gomega.BeNil())
				gomega.Expect(result).To(gomega.Equal(tnf.SUCCESS))
			})
		})
	})

})