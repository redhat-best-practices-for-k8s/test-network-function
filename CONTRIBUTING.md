# Contributing

## Peer review

Although this is an open source project, a review is required from one of the following committers prior to merging a
Pull Request:

* Ryan Goulding (rgoulding@redhat.com)
* David Spence (dspence@redhat.com)

This list is expected to grow over time.

*No Self Review is allowed.*  Each Pull Request should be peer reviewed prior to merge.

## Documentation guidelines

Each exported API must have proper documentation.  This documentation should adhere to `gofmt` rules.

Each exported global variable or constant must have proper documentation which adheres to `gofmt`.

## Style guidelines

The `test-network-function` project committers expect all Pull Requests to adhere to `gofmt` auto-styling tool.  To
run `gofmt`, run the following command:

```bash
make lint
```

Prior to submitting a Pull Request, please ensure you have run the above command.

## Test guidelines

Each `tnf.Test` implementation must have unit tests.  Ideally, each `tnf.Test` implementation should strive for 100%
line coverage when possible.  For some examples of existing tests, consult:

* pkg/tnf/handlers/base/version_test.go
* pkg/tnf/handlers/hostname/hostname_test.go
* pkg/tnf/handlers/ipaddr/ipaddr_test.go
* pkg/tnf/handlers/ping/ping_test.go

As always, you should ensure that tests should pass prior to submitting a Pull Request.  To run the unit tests issue the
following command:

```bash
make unit-tests
```

## Mock guidelines

The `test-network-function` project embraces GoLang `interface`s.  As such, pull requests should contain mock
implementations whenever necessary.  If an upstream dependency does not contain an interface or mock implementation,
the code author is expected to create a minimal shim interface that represents the needed functionality.  Tests that
utilize actual implementations will not be considered for merge.  Traditionally, we suggest using `mockgen`, though
other mock generators will be considered if they provided technology above and beyond what mockgen provides.

As an example, consult `spawner.SpawnFunc`.  For unit test case contexts, Unix commands should not be executed, so a
mock implementation is provided in `mock_interactive.MockSpawnFunc`.

Mocks for `test-network-function` interfaces should be auto-generated when possible.  Please add any necessary `mockgen`
implementations to the `mocks` Makefile target.

To make mocks, issue the following command:

```bash
make mocks
```

For some interfaces, such as `expect.Expecter`, generate the mock using `mockgen` externally and add it to source.

## Test guidelines

Each contributed test is expected to implement the `reel.Handler` and `tnf.Test` interfaces.  Additionally, each test
must be based on CLI commands.  No tests should utilize OpenShift client.  The choice to avoid OpenShift client is
deliberate, and was decided to aid in support of all versions of OpenShift despite the API(s) changing.  Generally
speaking, the CLI API changes much less quickly.

## Configuration guidelines

Most tests will require some form of configuration.  All configuration must implement or inherit a working `MarshalJSON`
and `UnmarshalJSON` interface.  This is due to the fact that a
[test-network-function-claim](https://github.com/redhat-nfvpe/test-network-function-claim) is output as JSON.

Additionally, each configuration type must be registered with `pkg/config/pool`.  In order to register with the
configuration pool, use code similar to the following:

```
(*configpool.GetInstance()).RegisterConfiguration(configurationKey, config)
```

Any configuration that adheres to these two requirements will automatically be included in the claim.
