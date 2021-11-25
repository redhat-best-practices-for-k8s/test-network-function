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
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"

	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
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

// DefaultTimeout for creating new interactive sessions (oc, ssh, tty)
var DefaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second

// LogLevelTraceEnabled is saved to filter some debug trace logs (e.g. expecters Sent/Match)
var LogLevelTraceEnabled = false

// GetContext spawns a new shell session and returns its context
func GetContext() *interactive.Context {
	context, err := interactive.SpawnShell(interactive.CreateGoExpectSpawner(), DefaultTimeout, interactive.Verbose(LogLevelTraceEnabled), interactive.SendTimeout(DefaultTimeout))
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(context).ToNot(gomega.BeNil())
	gomega.Expect(context.GetExpecter()).ToNot(gomega.BeNil())
	return context
}

// IsNonOcpCluster returns true when the env var is set, OCP only test would be skipped based on this flag
func IsNonOcpCluster() bool {
	b, _ := strconv.ParseBool(os.Getenv("TNF_NON_OCP_CLUSTER"))
	return b
}

// Intrusive is for running tests that can impact the CNF or test environment in an intrusive way
func Intrusive() bool {
	b, _ := strconv.ParseBool(os.Getenv("TNF_NON_INTRUSIVE_ONLY"))
	return !b
}

// GetOcDebugImageID is for running oc debug commands in a disconnected environment with a specific oc debug pod image mirrored
func GetOcDebugImageID() string {
	return os.Getenv("TNF_OC_DEBUG_IMAGE_ID")
}

// logLevel retrieves the LOG_LEVEL environment variable
func logLevel() string {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		log.Info("LOG_LEVEL environment is not set, defaulting to DEBUG")
		logLevel = "debug" //nolint:goconst
	}

	return logLevel
}

// SetLogLevel sets the log level for logrus based on the "LOG_LEVEL" environment variable
func SetLogLevel() {
	var aLogLevel, err = log.ParseLevel(logLevel())

	if err != nil {
		log.Error("LOG_LEVEL environment set with an invalid value, defaulting to DEBUG \n Valid values are:  trace, debug, info, warn, error, fatal, panic")
		aLogLevel = log.DebugLevel
	}

	if aLogLevel == log.TraceLevel {
		LogLevelTraceEnabled = true
	}

	log.Info("Log level set to:", aLogLevel)
	log.SetLevel(aLogLevel)
}

// SetLogFormat sets the log format for logrus
func SetLogFormat() {
	log.Info("debug format initialization: start")
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = time.StampMilli
	customFormatter.PadLevelText = true
	customFormatter.FullTimestamp = true
	customFormatter.ForceColors = true
	log.SetReportCaller(true)
	customFormatter.CallerPrettyfier = func(f *runtime.Frame) (string, string) {
		_, filename := path.Split(f.File)
		return strconv.Itoa(f.Line) + "]", fmt.Sprintf("[%s:", filename)
	}
	log.SetFormatter(customFormatter)
	log.Info("debug format initialization: done")
}
