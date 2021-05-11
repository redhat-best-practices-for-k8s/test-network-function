## Test Configuration

Configuration is accomplished with `tnf_config.yml` by default. An alternative configuration can be provided using the
`TNF_CONFIGURATION_PATH` environment variable.

This config file contains several sections, each of which configures one or more test specs:

Config Section|Purpose
---|---
generic|Describes containers to be tested with the `generic` and `multus` specs, if they are run.
cnfs|Defines which containers are to be tested by the `container` spec.
operators|Defines which containers are to be tested by the `operator` spec.
certifiedcontainerinfo|Describes cnf names and repositories to be checked for certification status.
certifiedoperatorinfo|Describes operator names and organisations to be checked for certification status.

`testconfigure.yml` defines roles, and which tests are appropriate for which roles. It should not be necessary to modify this.


### generic

The `generic` section contains three subsections:

* `containersUnderTest:` describes the CNFs that will be tested.  Each container is defined by the combination of its
`namespace`, `podName`, and `containerName`, which are also used to connect to the container when required.

  * Each entry for `containersUnderTest` must also define the `defaultNetworkDevice` of that container.  There is also
  an optional `multusIpAddresses` that can be omitted if the multus tests are not run.

* `partnerContainers:` describes the containers that support the testing.  Multiple `partnerContainers` allows
for more complex testing scenarios.  At the time of writing, only one is used, which will also be the test
orchestrator.

* `testOrchestrator:` references a partner containers that is used for the generic test suite.  The test partner is used
to send various types of traffic to each container under test.  For example the orchestrator is used to ping a container
under test, and to be the ping target of a container under test.

The [included default](test-network-function/tnf_config.yml) defines a single container to be tested,
and a single partner to do the testing.

### cnfs and operators

The `cnfs` and `operators` sections define the roles under which operators and containers are to be tested.

[The default config](test-network-function/tnf_config.yml) is set up with some examples of this:
It will run the `"OPERATOR_STATUS"` tests (as defined in `testconfigure.yml`) against an etcd operator, and the
`"PRIVILEGED_POD"` and `"PRIVILEGED_ROLE"` tests against an nginx container.

A more extensive example of all these sections is provided in [example/example_config.yaml](example/example_config.yaml)

### certifiedcontainerinfo and certifiedoperatorinfo

The `certifiedcontainerinfo` and `certifiedoperatorinfo` sections contain information about CNFs and Operators that are
to be checked for certification status on Red Hat catalogs.
