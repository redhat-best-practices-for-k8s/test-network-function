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

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/google/goterm/term"
	expect "github.com/ryandgoulding/goexpect"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	// ocCmdMandatoryNumArgs is the number of positional arguments expected for "jsontest run oc"
	ocCmdMandatoryNumArgs = 4

	// shellCmdMandatoryNumArgs is the number of positional arguments expected for "jsontest run shell"
	shellCmdMandatoryNumArgs = 1

	// sshCmdMandatoryNumArgs is the number of positional arguments expected for "jsontest run ssh"
	sshCmdMandatoryNumArgs = 3

	// testCreationErrorExitCode is the Unix return code used when a test fails to instantiate due to issues with the
	// underlying Unix subprocess.
	testCreationErrorExitCode = iota + 100

	// testDidNotParseExitCode is the Unix return code used when an improperly formatted or JSON file is provided to the
	// program.
	testDidNotParseExitCode

	// testDoesNotConformToSchemaExitCode is the Unix return code used when the JSON file provided does not adhere to the
	// generic-test.schema.json JSON schema.
	testDoesNotConformToSchemaExitCode

	// testExpecterCreationFailedExitCode is the Unix return code when an expect.Expecter fails to properly instantiate.
	testExpecterCreationFailedExitCode

	// testMarshalErrorExitCode is the Unix return code used when the JSON file fails to Marshal correctly.  This may
	// be caused by issues in the JSON payload.
	testMarshalErrorExitCode

	// testRunErrorExitCode is the Unix return code used when there is an error encountered while running the finite
	// state machine.
	testRunErrorExitCode

	// inappropriateArguments is the Unix return code used when the supplied arguments are inappropriate for the given
	// context.
	inappropriateArguments
)

var (
	// rootCmd is the jsontest executable root.  Currently, the jsontest entrypoint has only one sub-command called
	// "run", which runs a generic JSON test.
	rootCmd = &cobra.Command{
		Use:   "jsontest-cli",
		Short: "A CLI for creating, validating, and running JSON test-network-function tests.",
		Long:  `jsontest is a CLI library included in test-network-function used to prototype JSON test cases.`,
	}

	// runCmd is the json test executable option to run a JSON test.
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "run a JSON test case",
		Long:  `run is a CLI library included in test-network-function used to run a JSON test case.  The JSON test case can be run using oc, ssh, or local shell.`,
	}

	// shellCmd is the entrypoint for running a test case on the local shell.
	shellCmd = &cobra.Command{
		Use:   "shell [jsonTestFile]",
		Short: "run in a local shell context",
		Args:  cobra.ExactValidArgs(shellCmdMandatoryNumArgs),
		Run:   runShellCmd,
	}

	// sshCmd is the entrypoint for running a test case in an ssh shell.
	sshCmd = &cobra.Command{
		Use:   "ssh [user] [host] [jsonTestFile]",
		Short: "run in an ssh context",
		Args:  cobra.ExactValidArgs(sshCmdMandatoryNumArgs),
		Run:   runSSHCmd,
	}

	// ocCmd is the entrypoint for running a test case in an interactive oc shell.
	ocCmd = &cobra.Command{
		Use:   "oc [namespace] [pod] [container] [jsonTestFile]",
		Short: "run in an oc context",
		Args:  cobra.ExactValidArgs(ocCmdMandatoryNumArgs),
		Run:   runOcCmd,
	}

	// schemaPath is the path to the generic-test.schema.json JSON schema relative to the program entrypoint.
	schemaPath = path.Join("schemas", generic.TestSchemaFileName)
)

// fatalError reports a fatal error to stdout and exits.
func fatalError(msg string, err error, exitCode int) {
	log.Errorf("Fatal Error: %s caused by: %s", msg, err)
	os.Exit(exitCode)
}

// createTest creates a tnf.Test using tester and handlers in the expecter context.
func createTest(expecter *expect.Expecter, tester *tnf.Tester, handlers []reel.Handler, ch <-chan error) (*tnf.Test, error) {
	return tnf.NewTest(expecter, *tester, handlers, ch)
}

// reportResults is a helper function used to log results to the console.
func reportResults(tester *tnf.Tester, result int) {
	log.Infof("Test Result: %d", result)
	log.Info("Test Payload:")
	testOutput, err := json.MarshalIndent(tester, "", "    ")
	if err != nil {
		fatalError("could not Marshal the test result", err, testMarshalErrorExitCode)
	}
	log.Info(string(testOutput))
}

// runTest creates and runs the tnf.Test implementation.
func runTest(expecter *expect.Expecter, tester *tnf.Tester, handlers []reel.Handler, ch <-chan error) {
	// Instantiate the test.
	test, err := createTest(expecter, tester, handlers, ch)
	if err != nil {
		fatalError("could not create the test", err, testCreationErrorExitCode)
	}

	// Actually run the test in the expecter context.
	result, err := test.Run()
	if err != nil {
		fatalError("could not run the test", err, testRunErrorExitCode)
	}

	// Human readable results output to the console.
	reportResults(tester, result)
}

// setupAndRunTest is a helper function to run a JSON test in a given expecter context.
func setupTest(file string) (*tnf.Tester, []reel.Handler) {
	// Instantiate the test from JSON, ensuring that JSON validation succeeds.
	tester, handlers, result, err := generic.NewGenericFromJSONFile(file, schemaPath)
	if err != nil {
		fatalError("the supplied JSON test could not be parsed correctly", err, testDidNotParseExitCode)
	}

	// If the given JSON file does not parse against the generic test schema, report all of the problems in a human
	// readable format.
	if !result.Valid() {
		log.Error("The supplied JSON does not conform to the generic-test.schema.json JSON schema.  Here are the problems we found:")
		for _, e := range result.Errors() {
			log.Errorf("- %v\n", e)
		}
		fatalError("the supplied JSON does not conform to the expected JSON schema",
			fmt.Errorf("schema validation error"), testDoesNotConformToSchemaExitCode)
	}
	return tester, handlers
}

func runSSHCmd(_ *cobra.Command, args []string) {
	// Extra careful check since cobra checks this for us.
	if len(args) < sshCmdMandatoryNumArgs {
		errString := fmt.Sprintf("oc requires %d arguments", sshCmdMandatoryNumArgs)
		fatalError(errString, errors.New(errString), inappropriateArguments)
	}

	// Report what is being run to the user.
	user, host, file := args[0], args[1], args[2]
	log.Info(term.Bluef("Running %s from ssh interactive shell [user=\"%s\" host=\"%s\"]", file, user, host))

	// setup / parse the input test.  Must be done prior to creating the Expecter to derive the test timeout.
	tester, handlers := setupTest(file)

	// SSH shell creation.
	goExpectSpawner := interactive.NewGoExpectSpawner()
	var spawnContext interactive.Spawner = goExpectSpawner
	context, err := interactive.SpawnSSH(&spawnContext, user, host, (*tester).Timeout(), expect.Verbose(true))
	if err != nil {
		fatalError("could not create the ssh expecter", err, testExpecterCreationFailedExitCode)
	}

	// actually run the test.
	runTest(context.GetExpecter(), tester, handlers, context.GetErrorChannel())
}

// runOcCmd is a helper function used to run JSON tests in an oc context.
func runOcCmd(_ *cobra.Command, args []string) {
	// Extra careful check since cobra checks this for us.
	if len(args) < ocCmdMandatoryNumArgs {
		errString := fmt.Sprintf("oc requires %d arguments", ocCmdMandatoryNumArgs)
		fatalError(errString, errors.New(errString), inappropriateArguments)
	}

	// Report what is being run to the user.
	namespace, pod, container, file := args[0], args[1], args[2], args[3]
	log.Info(term.Bluef("Running %s from oc interactive shell [namespace=\"%s\" pod=\"%s\" container=\"%s\"]", file, namespace, pod, container))

	// setup / parse the input test.  Must be done prior to creating the Expecter to derive the test timeout.
	tester, handlers := setupTest(file)

	// oc shell creation.
	goExpectSpawner := interactive.NewGoExpectSpawner()
	var spawnContext interactive.Spawner = goExpectSpawner
	oc, ch, err := interactive.SpawnOc(&spawnContext, pod, container, namespace, (*tester).Timeout(), expect.Verbose(true))
	if err != nil {
		fatalError("could not create the oc expecter", err, testExpecterCreationFailedExitCode)
	}

	// actually run the test.
	runTest(oc.GetExpecter(), tester, handlers, ch)
}

// runShellCmd is a helper function used to run JSON tests in a shell context.
func runShellCmd(_ *cobra.Command, args []string) {
	// Extra careful check since cobra checks this for us.
	if len(args) < shellCmdMandatoryNumArgs {
		errString := fmt.Sprintf("shell requires %d arguments", shellCmdMandatoryNumArgs)
		fatalError(errString, errors.New(errString), inappropriateArguments)
	}

	// Report what is being run to the user.
	file := args[0]
	log.Info(term.Bluef("Running %s from a local shell context", file))

	// setup / parse the input test.  Must be done prior to creating the Expecter to derive the test timeout.
	tester, handlers := setupTest(file)

	// Shell creation.
	goExpectSpawner := interactive.NewGoExpectSpawner()
	var spawnContext interactive.Spawner = goExpectSpawner
	context, err := interactive.SpawnShell(&spawnContext, (*tester).Timeout(), expect.Verbose(true))
	if err != nil {
		fatalError("could not create the shell expecter", err, testExpecterCreationFailedExitCode)
	}

	// actually run the test.
	runTest(context.GetExpecter(), tester, handlers, context.GetErrorChannel())
}

// Execute executes the jsontest program, returning any applicable errors.
func Execute() error {
	runCmd.AddCommand(ocCmd, sshCmd, shellCmd)
	rootCmd.AddCommand(runCmd)
	return rootCmd.Execute()
}
