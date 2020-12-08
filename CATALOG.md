# `tnf.Test` Catalog

A number of `tnf.Test` implementations are included out of the box.  This is a summary of the available implementations:
## http://test-network-function.com/tests/casa/nrf/checkregistration
Property|Description
---|---
Version|v1.0.0
Description|A Casa cnf-specific test which checks the Registration status of the AMF and SMF from the NRF.  This is done by making sure the "nfStatus" field in the "nfregistrations.mgmt.casa.io" Custom Resource reports as "REGISTERED"
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`jq`, `oc`

## http://test-network-function.com/tests/hostname
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to check the hostname of a target machine/container.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`hostname`

## http://test-network-function.com/tests/ipaddr
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to derive the default network interface IP address of a target container.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`ip`

## http://test-network-function.com/tests/operator
Property|Description
---|---
Version|v1.0.0
Description|An operator-specific test used to exercise the behavior of a given operator.  In the current offering, we check if the operator ClusterServiceVersion (CSV) is installed properly.  A CSV is a YAML manifest created from Operator metadata that assists the Operator Lifecycle Manager (OLM) in running the Operator.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`jq`, `oc`

## http://test-network-function.com/tests/ping
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to test ICMP connectivity from a source machine/container to a target destination.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`ping`

## http://test-network-function.com/tests/container/pod
Property|Description
---|---
Version|v1.0.0
Description|A container-specific test suite used to verify various aspects of the underlying container.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`jq`, `oc`

## http://test-network-function.com/tests/generic/version
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to determine if a target container/machine is based on RHEL.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`cat`

## http://test-network-function.com/tests/casa/nrf/id
Property|Description
---|---
Version|v1.0.0
Description|A Casa cnf-specific test which checks for the existence of the AMF and SMF CNFs.  The UUIDs are gathered and stored by introspecting the "nfregistrations.mgmt.casa.io" Custom Resource.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`awk`, `oc`, `xargs`

