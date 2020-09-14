package ipaddr_test

import (
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/ipaddr"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path"
	"testing"
	"time"
)

const (
	testDataDirectory   = "testdata"
	testDataFileSuffix  = ".txt"
	testTimeoutDuration = time.Second * 2
)

type TestCase struct {
	device              string
	pattern             string
	expectedResult      int
	expectedIpv4Address string
}

var testCases = map[string]TestCase{
	"device_exists": {
		device:              "eth0",
		pattern:             ipaddr.SuccessfulOutputRegex,
		expectedResult:      tnf.SUCCESS,
		expectedIpv4Address: "172.17.0.7",
	},
	"device_does_not_exist": {
		device:              "dne",
		pattern:             ipaddr.DeviceDoesNotExistRegex,
		expectedResult:      tnf.ERROR,
		expectedIpv4Address: "",
	},
}

func getMockOutputFilename(testName string) string {
	return path.Join(testDataDirectory, fmt.Sprintf("%s%s", testName, testDataFileSuffix))
}

func getMockOutput(t *testing.T, testName string) string {
	fileName := getMockOutputFilename(testName)
	b, err := ioutil.ReadFile(fileName)
	assert.Nil(t, err)
	return string(b)
}

func TestNewIpAddr(t *testing.T) {
	for _, testCase := range testCases {
		ipAddr := ipaddr.NewIpAddr(testTimeoutDuration, testCase.device)
		assert.NotNil(t, ipAddr)
		assert.Equal(t, tnf.ERROR, ipAddr.Result())
		assert.Equal(t, []string{"ip", "addr", "show", "dev", testCase.device}, ipAddr.Args())
	}
}

func TestIpAddr_Args(t *testing.T) {
	for _, testCase := range testCases {
		ipAddr := ipaddr.NewIpAddr(testTimeoutDuration, testCase.device)
		assert.Equal(t, []string{"ip", "addr", "show", "dev", testCase.device}, ipAddr.Args())
	}
}

func TestIpAddr_Timeout(t *testing.T) {
	for _, testCase := range testCases {
		ipAddr := ipaddr.NewIpAddr(testTimeoutDuration, testCase.device)
		assert.Equal(t, testTimeoutDuration, ipAddr.Timeout())
	}
}

func TestIpAddr_ReelFirst(t *testing.T) {
	for _, testCase := range testCases {
		ipAddr := ipaddr.NewIpAddr(testTimeoutDuration, testCase.device)
		step := ipAddr.ReelFirst()
		assert.Equal(t, "", step.Execute)
		assert.Contains(t, step.Expect, ipaddr.SuccessfulOutputRegex)
		assert.Equal(t, testTimeoutDuration, step.Timeout)
	}
}

func TestIpAddr_Result(t *testing.T) {
	for testName, testCase := range testCases {
		ipAddr := ipaddr.NewIpAddr(testTimeoutDuration, testCase.device)
		assert.Equal(t, tnf.ERROR, ipAddr.Result())
		step := ipAddr.ReelMatch(testCase.pattern, "", getMockOutput(t, testName))
		assert.Nil(t, step)
		assert.Equal(t, testCase.expectedResult, ipAddr.Result())
	}
}

func TestIpAddr_GetIpv4Address(t *testing.T) {
	for testName, testCase := range testCases {
		ipAddr := ipaddr.NewIpAddr(testTimeoutDuration, testCase.device)
		step := ipAddr.ReelMatch(testCase.pattern, "", getMockOutput(t, testName))
		assert.Nil(t, step)
		assert.Equal(t, testCase.expectedIpv4Address, ipAddr.GetIpv4Address())
	}
}

func TestIpAddr_ReelTimeout(t *testing.T) {
	for _, testCase := range testCases {
		ipAddr := ipaddr.NewIpAddr(testTimeoutDuration, testCase.device)
		assert.Nil(t, ipAddr.ReelTimeout())
	}
}

// Ensure there are no panics.
func TestIpAddr_ReelEof(t *testing.T) {
	for _, testCase := range testCases {
		ipAddr := ipaddr.NewIpAddr(testTimeoutDuration, testCase.device)
		ipAddr.ReelEof()
	}
}
