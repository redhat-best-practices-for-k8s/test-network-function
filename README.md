# Test Network Function

This repository contains a set of network function test cases.

## Dependencies

At a minimum, the following dependencies must be installed *prior* to running `make dependencies`.

Dependency|Minimum Version
---|---
[GoLang](https://golang.org/dl/)|1.14
[golangci-lint](https://golangci-lint.run/usage/install/)|1.32.2
[jq](https://stedolan.github.io/jq/)|1.6
[OpenShift Client](https://docs.openshift.com/container-platform/4.4/welcome/index.html)|4.4

All other dependencies required to run tests can be installed using the following command:

```shell-script
make dependencies
```

*Note*:  Efforts to containerize this offering are considered a work in progress.

## Available Test Specs

There are two categories for CNF tests;  'General' and 'CNF-specific'.

The 'General' tests are designed to test any commodity CNF running on OpenShift, and include specifications such as
'Default' network connectivity.

'CNF-specific' tests are designed to test some unique aspects of the CNF under test are behaving correctly.  This could
include specifications such as issuing a `GET` request to a web server, or passing traffic through an IPSEC tunnel.

### General

The general-purpose category covers most tests. It consists of multiple suites that can be run in any combination as is
appropriate for the CNF(s) under test:

Suite|Test Spec Description|Minimum OpenShift Version
---|---|---
diagnostic|The diagnostic test suite is used to gather node information from an OpenShift cluster.  The diagnostic test suite should be run whenever generating a claim.json file.|4.4.3
generic|The generic test suite is used to test `Default` network connectivity between containers.  It also checks that the base container image is based on `RHEL`.|4.4.3
multus|The multus test suite is used to test SR-IOV network connectivity between containers.|4.4.3
operator|The operator test suite is designed basic Kubernetes Operator functionality.|4.4.3
container|The container test suite is designed to test container functionality and configuration|4.4.3

## Performing Tests

Currently, all available tests are part of the "CNF Certification Test Suite" test suite, which serves as the entrypoint
to run all test specs.  `CNF Certification 1.0` is not containerized, and involves pulling, building, then running the
tests.

By default, `test-network-function` emits results to `test-network-function/cnf-certification-tests_junit.xml`.

The included default configuration is for running `generic` and `multus` suites on the trivial example at
[cnf-certification-test-partner](https://github.com/redhat-nfvpe/cnf-certification-test-partner). To configure for your
own environment, please see the Test Configuration section, below.

### Pulling The Code

In order to pull the code, issue the following command:

```shell-script
mkdir ~/workspace
cd ~/workspace
git clone git@github.com:redhat-nfvpe/test-network-function.git
cd test-network-function
```

### Building the Tests

In order to build the test executable, first make sure you have satisfied the [dependencies](#dependencies).

```shell-script
make build-cnf-tests
```

If build fails after `go get github.com/onsi/ginkgo/ginkgo` Add ginkgo location to the PATH: `export PATH=$PATH:~/go/bin`

*Gotcha:* The `make build` command runs the unit tests for the framework, it does NOT test the CNF.

### Testing a CNF

Once the executable is built, a CNF can be tested by specifying which suites to run using the `run-cnf-suites.sh` helper
script.
Any combintation of the suites listed above can be run, e.g.

```shell-script
./run-cnf-suites.sh diagnostic
./run-cnf-suites.sh diagnostic generic
./run-cnf-suites.sh diagnostic generic multus
./run-cnf-suites.sh diagnostic operator
./run-cnf-suites.sh diagnostic generic multus container operator
```

*Gotcha:* The generic test suite requires that the CNF has both `ping` and `ip` binaries installed. Please add them
manually if the CNF under test does not include these. Automated installation of missing dependencies is targetted
for a future version.

## Test Configuration

There are a few pieces of configuration required to allow the test framework to access and test the CNF:

Config File|Purpose
---|---
test-configuration.yaml|Describes the CNF or CNFs that are to be tested, the container that will run the tests, and the test orchestrator.
config.yml|Defines which operators are to be tested.
testconfigure.yml|Defines operator tests are appropriate for which roles.

Combining these configuration files is a near-term goal.

### test-configuration.yaml

The config file `test-configuration.yaml` contains three sections:

* `containersUnderTest:` describes the CNFs that will be tested. Each container is defined by the combination of its
`namespace`, `podName`, and `containerName`, which are also used to connect to the container when required.

  * Each entry for `containersUnderTest` must also define the `defaultNetworkDevice` of that container. There is also an
  optional `multusIpAddresses` that can be omitted if the multus tests are not run.

* `partnerContainers:` describes the containers that support the testing. Multiple `partnerContainers` allows
for more complex testing scenarios. At the time of writing, only one is used, which will also be the test
orchestrator.

* `testOrchestrator:` references a partner containers that is used for the generic test suite. The test partner is used to send various types of traffic to each container under test. For example the orchestrator is used to ping a container under test, and to be the ping target of a container under test.

The [included example](test-network-function/test-configuration.yaml) defines a single container to be tested, and a
single partner to do the testing.

### Operator Test Configuration

Testing operators is currently configured separately from the generic tests.

#### config.yml

You can either edit the provided config `config.yml` TODO: or pass a different config by using the `-config` flag.

Sample config.yml

```yaml
cnfs:
  - name: "crole-test-pod"
    namespace: "default"
    status: "Running"
    tests:
      - "PRIVILEGED_POD"
      - "PRIVILEGED_ROLE"
  - name: "nginx"
    namespace: "default"
    status: "Running"
    tests:
      - "PRIVILEGED_POD"
      - "PRIVILEGED_ROLE"
```

This example config is set to run the `"PRIVILEGED_POD"` and `"PRIVILEGED_ROLE"` tests on two operators: `nginx` and
`crole-test-pod`

#### testconfigure.yml

By default, the test suite will run all the default test cases defined by `testconfigure.yml` file.  You can change
which tests run by modifying `testconfigure.yml`.
Example testconfigure.yml:

```yaml
 operatortest:
   - name: "OPERATOR_STATUS"
     tests:
       - "CSV_INSTALLED"
 cnftest:
   - name: "PRIVILEGED_POD"
     tests:
     - "HOST_NETWORK_CHECK"
     - "HOST_PORT_CHECK"
     - "HOST_PATH_CHECK"
     - "HOST_IPC_CHECK"
     - "HOST_PID_CHECK"
     - "CAPABILITY_CHECK"
     - "ROOT_CHECK"
     - "PRIVILEGE_ESCALATION"
   - name: "PRIVILEGED_ROLE"
     tests:
     - "CLUSTER_ROLE_BINDING_BY_SA"
 ```

#### Container Test Configuration

You can either edit the provided config `config.yml`  or pass a different config by using the `-config` flag to the
test suite.

Sample config.yml

```yaml
cnfs:
  - name: "crole-test-pod"
    namespace: "default"
    status: "Running"
    tests:
      - "PRIVILEGED_POD"
      - "PRIVILEGED_ROLE"
  - name: "nginx"
    namespace: "default"
    status: "Running"
    tests:
      - "PRIVILEGED_POD"
      - "PRIVILEGED_ROLE"
```

## Test Output

The test suite generates a "claim" file, which describes the system(s) under test, the tests that were run, and the
outcome of all of the tests. This claim file is the proof of the test run that is evaluated by Red Hat when
certified status is being granted. For more information about the contents of the claim file please see the
[schema](https://github.com/redhat-nfvpe/test-network-function-claim/blob/master/claim.schema.json). For more
information about the purpose of the claim file see the docs.
!!! TODO: link to docs when published.
