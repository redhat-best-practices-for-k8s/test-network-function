# Test Network Function

This repository contains a set of network function test cases.

## Dependencies

At a minimum, the following dependencies must be installed prior to running `make dependencies`.

Dependency|Minimum Version
---|---
[GoLang](https://golang.org/dl/)|1.14
[golangci-lint](https://golangci-lint.run/usage/install/)|1.32.2
[jq](https://stedolan.github.io/jq/)|1.6
[OpenShift Client](https://docs.openshift.com/container-platform/4.4/welcome/index.html)|4.4

All other dependencies required to run tests should be installed using the following command:

```shell script
make dependencies
```

*Note*:  Efforts to containerize this offering are considered a work in progress.

## Available Test Specs

There are two categories for CNF tests;  `generic` and `CNF-specific`.  `generic` tests are designed to test any
commodity CNF running on OpenShift, and include specifications such as `Default` network connectivity.  `CNF-specific`
tests are designed to test unique aspects of the CNF under test.  This could include specifications such as issuing a
`GET` request to a web server, or passing traffic through an IPSEC tunnel.

### Generic

Suite|Test Spec Description|Minimum OpenShift Version
---|---|---
generic|The generic test suite is used to test `Default` network connectivity between containers.  It also checks that the base container image is based on `RHEL`.|4.4.3
multus|The multus test suite is used to test SR-IOV network connectivity between containers.|4.4.3
operator|The operator test suite is designed basic Kubernetes Operator functionality.|4.4.3
container|The container test suite is designed to test container functionality and configuration|4.4.3

#### Generic Test Network Function Runtime Dependencies

The generic test suite requires that the NF has the following binary dependencies:
* `ping`
* `ip`

If you do not provide these dependencies in your NF, please add them manually using your platform-dependent package
manager.  Automated installation of missing dependencies is considered future work.

### CNF Specific

Suite|Test Spec Description|Minimum OpenShift Version
---|---|---
cisco_kiknos|Cisco Kiknos specific tests include establishing an IPSEC tunnel between Kiknos and test pod `ikester`, and then passing a minimum amount of `ICMP` and `UDP` traffic across the tunnel.|4.4.3
casa_cnf|Casa 5G Core specific tests include ensuring proper registration of AMF/SMF CNFs with the NRF.|4.4.3

## Performing Tests

Currently, all available tests are part of the "CNF Certification Test Suite" test suite, which serves as the entrypoint
to run all test specs.  `CNF Certification 1.0` is not containerized, and involves pulling, building, then running the
tests.  By default, `test-network-function` emits results to `test-network-function/cnf-certification-tests_junit.xml`.

### Pulling The Code

In order to pull the code, issue the following command:

```shell script
mkdir ~/workspace
cd ~/workspace
git clone git@github.com:redhat-nfvpe/test-network-function.git
cd test-network-function
```

### Building the Tests

In order to build the code, first make sure you have satisfied the [dependencies](#dependencies).

```shell script
make build-cnf-tests
```

### Run Generic Tests Only

In order to run the Generic CNF tests only, issue the following command:

```shell script
make generic-cnf-tests
```

### Run All CNF Certification Tests

In order to run all CNF tests, issue the following command:

```shell script
make cnf-tests
```

### Run a Specific Test Suite

In order to run a specific test suite, `cisco_kiknos` for example, issue the following command:

```shell script
make build build-cnf-tests
cd ./test-network-function && ./test-network-function.test -ginkgo.v -ginkgo.focus="cisco_kiknos" -junit . -report .
```

### Run the Operator Test Suite
In order to run the Operator test suite, issue the following command:

```shell script
make build build-cnf-tests
cd ./test-network-function && ./test-network-function.test -config=config.yml -ginkgo.v -ginkgo.focus="operator" -junit . -report .
```

#### Operator Test Configuration
 
You can either edit the provided config `config.yml` or pass a different config by using the `-config` flag.

Sample config.yml
```
operators:
  - name: etcdoperator.v0.9.4
    namespace: my-etcd
    status: Succeeded
    autogenerate: "false"
    tests:
      - "OPERATOR_STATUS"
cnfs:
```

Configuring tests

By default, the test suite will run all the default test cases defined by `testconfigure.yml` file.  You can change
which tests run by modifying `testconfigure.yml`.  The contents should match as shown below.

Example:
```
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

### Run the Container Test Suite

In order to run the container test suite, issue the following command:

```shell script
make build build-cnf-tests
cd ./test-network-function && ./test-network-function.test -config=config.yml -ginkgo.v -ginkgo.focus="container" -junit . -report .
```

#### Container Test Configuration
 
You can either edit the provided config `config.yml`  or pass a different config by using the `-config` flag to the
test suite.

Sample config.yml
```
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

Configuring test cases

By default, the test suite will run all the default test cases defined by `testconfigure.yml` file.  You can change
which tests run by modifying `testconfigure.yml`.  The contents should match as shown below.

Example:
```
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
