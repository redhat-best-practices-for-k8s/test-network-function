// Copyright (C) 2020-2021 Red Hat, Inc.
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

package common

import (
	"errors"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
)

var (
	// PathRelativeToRoot is used to calculate relative filepaths for the `test-network-function` executable entrypoint.
	PathRelativeToRoot = path.Join("..")

	// RelativeSchemaPath is the relative path to the generic-test.schema.json JSON schema.
	RelativeSchemaPath = path.Join(PathRelativeToRoot, schemaPath)

	// schemaPath is the path to the generic-test.schema.json JSON schema relative to the project root.
	schemaPath = path.Join("schemas", "generic-test.schema.json")
)

const (
	logLevelTraceString = "trace"
	logLevelDebugString = "debug"
	logLevelInfoString  = "info"
	logLevelWarnString  = "warn"
	logLevelErrorString = "error"
	logLevelFatalString = "fatal"
	logLevelPanicString = "panic"
	logLevelEmptyString = ""
	logLevelDefault     = logrus.InfoLevel
	errorEmpty          = "empty"
	errorInvalid        = "invalid"
)

// DefaultTimeout for creating new interactive sessions (oc, ssh, tty)
var DefaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second

// GetContext spawns a new shell session and returns its context
func GetContext() *interactive.Context {
	context, err := interactive.SpawnShell(interactive.CreateGoExpectSpawner(), DefaultTimeout, interactive.Verbose(true))
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(context).ToNot(gomega.BeNil())
	gomega.Expect(context.GetExpecter()).ToNot(gomega.BeNil())
	return context
}

// RunAndValidateTest runs the test and checks the result
func RunAndValidateTest(test *tnf.Test) {
	testResult, err := test.Run()
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	gomega.Expect(err).To(gomega.BeNil())
}

// IsMinikube returns true when the env var is set, OCP only test would be skipped based on this flag
func IsMinikube() bool {
	b, _ := strconv.ParseBool(os.Getenv("TNF_MINIKUBE_ONLY"))
	return b
}

// Intrusive is for running tests that can impact the CNF or test environment in an intrusive way
func Intrusive() bool {
	b, _ := strconv.ParseBool(os.Getenv("TNF_NON_INTRUSIVE_ONLY"))
	return !b
}

// logLevel retrieves the LOG_LEVEL environement vaiable
func logLevel() string {
	return os.Getenv("LOG_LEVEL")
}

// stringToLogLevel converts a string to a log logrus level
func stringToLogLevel(logLevelString string) (logrus.Level, error) {
	logLevelString = strings.ToLower(logLevelString)
	switch logLevelString {
	case logLevelTraceString:
		return logrus.TraceLevel, nil
	case logLevelDebugString:
		return logrus.DebugLevel, nil
	case logLevelInfoString:
		return logrus.InfoLevel, nil
	case logLevelWarnString:
		return logrus.WarnLevel, nil
	case logLevelErrorString:
		return logrus.ErrorLevel, nil
	case logLevelFatalString:
		return logrus.FatalLevel, nil
	case logLevelPanicString:
		return logrus.PanicLevel, nil
	case logLevelEmptyString:
		return logLevelDefault, errors.New(errorEmpty)
	}
	return logLevelDefault, errors.New(errorInvalid)
}

// SetLogLevel sets the log level for logrus based on the "LOG_LEVEL" environment variable
func SetLogLevel() {
	var aLogLevel, err = stringToLogLevel(logLevel())

	if err != nil {
		if err.Error() == errorInvalid {
			logrus.Info("LOG_LEVEL environment set with an invalid value, defaulting to", logLevelDefault, ".\n Valid values are:  trace, debug, info, warn, error, fatal, panic")
		}
		if err.Error() == errorEmpty {
			logrus.Info("LOG_LEVEL environment variable not set, defaulting to: ", logLevelDefault, ".\n Valid values are:  trace, debug, info, warn, error, fatal, panic")
		}
	}
	logrus.Info("Log level set to:", aLogLevel)
	logrus.SetLevel(aLogLevel)
}
