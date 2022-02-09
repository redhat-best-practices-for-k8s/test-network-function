var initialjson={
  "claim": {
    "configurations": {
      "certifiedcontainerinfo": [
        {
          "name": "nginx-116",
          "repository": "rhel8"
        }
      ],
      "certifiedoperatorinfo": [
        {
          "name": "etcd",
          "organization": "community-operators"
        }
      ],
      "targetCrdFilters": [
        {
          "nameSuffix": "group1.test.com"
        },
        {
          "nameSuffix": "test-network-function.com"
        }
      ],
      "targetNameSpaces": [
        {
          "name": "tnf"
        }
      ],
      "targetPodLabels": [
        {
          "name": "generic",
          "prefix": "test-network-function.com",
          "value": "target"
        }
      ],
      "testPartner": {
        "debugContainers": [
          {
            "containerName": "container-00",
            "containerRuntime": "docker",
            "containerUID": "fcd1ac9ce91b96a7e957351257be8c46cd62fd2e796e352f588acc49a016a791",
            "defaultNetworkDevice": "",
            "namespace": "default",
            "nodeName": "minikube",
            "podName": "debug-7d8f7"
          },
          {
            "containerName": "container-00",
            "containerRuntime": "docker",
            "containerUID": "0afa1a9a57e6578dc869241141485820e49b914fdaca08ca37aacb8290c53fc4",
            "defaultNetworkDevice": "",
            "namespace": "default",
            "nodeName": "minikube-m03",
            "podName": "debug-jxqw7"
          }
        ]
      },
      "testTarget": {
        "Nodes": {
          "minikube": {
            "Labels": [
              "node-role.kubernetes.io/master",
              "node-role.kubernetes.io/worker"
            ],
            "Name": "minikube"
          },
          "minikube-m02": {
            "Labels": [
              "node-role.kubernetes.io/worker"
            ],
            "Name": "minikube-m02"
          },
          "minikube-m03": {
            "Labels": [
              "node-role.kubernetes.io/worker"
            ],
            "Name": "minikube-m03"
          }
        },
        "NonValidPods": null,
        "containersUnderTest": [
          {
            "containerName": "test",
            "containerRuntime": "docker",
            "containerUID": "09acaabb89c51902e623943d3e91e690072f215e1bca231e2c0fc569d9f828ce",
            "defaultNetworkDevice": "eth0",
            "multusIpAddressesPerNet": {
              "tnf/macvlan-conf": [
                "192.168.1.3"
              ]
            },
            "namespace": "tnf",
            "nodeName": "minikube-m03",
            "podName": "test-5b76f77fd5-7d6tb"
          },
          {
            "containerName": "test",
            "containerRuntime": "docker",
            "containerUID": "97ece3466ed2381ac4f9ca90c69ea5bf3dd9f594735b7d70faa1ad245f7f9ccd",
            "defaultNetworkDevice": "eth0",
            "multusIpAddressesPerNet": {
              "tnf/macvlan-conf": [
                "192.168.1.2"
              ]
            },
            "namespace": "tnf",
            "nodeName": "minikube",
            "podName": "test-5b76f77fd5-cg2ll"
          }
        ],
        "deploymentsUnderTest": [
          {
            "Hpa": {
              "HpaName": "test",
              "MaxReplicas": 3,
              "MinReplicas": 2
            },
            "Name": "test",
            "Namespace": "tnf",
            "Replicas": 2,
            "Type": "deployment"
          }
        ],
        "excludeContainersFromConnectivityTests": null,
        "operators": [
          {
            "name": "nginx-operator.v0.0.1",
            "namespace": "tnf",
            "subscriptionName": "nginx-operator-v0-0-1-sub",
            "tests": [
              "OPERATOR_STATUS"
            ]
          }
        ],
        "podsUnderTest": [
          {
            "containercount": 1,
            "containerfornettests": {
              "Oc": null,
              "containerName": "test",
              "containerRuntime": "docker",
              "containerUID": "09acaabb89c51902e623943d3e91e690072f215e1bca231e2c0fc569d9f828ce",
              "defaultNetworkDevice": "eth0",
              "defaultnetworkipaddress": "10.244.151.4",
              "multusIpAddressesPerNet": {
                "tnf/macvlan-conf": [
                  "192.168.1.3"
                ]
              },
              "namespace": "tnf",
              "nodeName": "minikube-m03",
              "podName": "test-5b76f77fd5-7d6tb"
            },
            "name": "test-5b76f77fd5-7d6tb",
            "namespace": "tnf",
            "serviceaccount": "default",
            "tests": [
              "PRIVILEGED_POD",
              "PRIVILEGED_ROLE"
            ]
          },
          {
            "containercount": 1,
            "containerfornettests": {
              "Oc": null,
              "containerName": "test",
              "containerRuntime": "docker",
              "containerUID": "97ece3466ed2381ac4f9ca90c69ea5bf3dd9f594735b7d70faa1ad245f7f9ccd",
              "defaultNetworkDevice": "eth0",
              "defaultnetworkipaddress": "10.244.120.67",
              "multusIpAddressesPerNet": {
                "tnf/macvlan-conf": [
                  "192.168.1.2"
                ]
              },
              "namespace": "tnf",
              "nodeName": "minikube",
              "podName": "test-5b76f77fd5-cg2ll"
            },
            "name": "test-5b76f77fd5-cg2ll",
            "namespace": "tnf",
            "serviceaccount": "default",
            "tests": [
              "PRIVILEGED_POD",
              "PRIVILEGED_ROLE"
            ]
          }
        ],
        "stateFulSetUnderTest": null
      }
    },
    "metadata": {
      "endTime": "2022-01-05T15:05:56+00:00",
      "startTime": "2022-01-05T15:05:39+00:00"
    },
    "nodes": {
      "cniPlugins": [],
      "csiDriver": {},
      "nodeSummary": {},
      "nodesHwInfo": {
        "Master": {
          "NodeName": "",
          "Lscpu": null,
          "IPconfig": null,
          "Lsblk": null,
          "Lspci": null
        },
        "Worker": {
          "NodeName": "",
          "Lscpu": null,
          "IPconfig": null,
          "Lsblk": null,
          "Lspci": null
        }
      }
    },
    "rawResults": {
      "cnf-certification-test": {
        "testsuites": {
          "-disabled": "0",
          "-errors": "0",
          "-failures": "0",
          "-tests": "5",
          "-time": "16.443838635",
          "testsuite": {
            "-disabled": "0",
            "-errors": "0",
            "-failures": "0",
            "-name": "CNF Certification Test Suite",
            "-package": "/home/deliedit/redhat/github/david/cnftest/test-network-function",
            "-skipped": "0",
            "-tests": "5",
            "-time": "16.443838635",
            "-timestamp": "2022-01-05T09:05:39",
            "properties": {
              "property": [
                {
                  "-name": "SuiteSucceeded",
                  "-value": "true"
                },
                {
                  "-name": "SuiteHasProgrammaticFocus",
                  "-value": "false"
                },
                {
                  "-name": "SpecialSuiteFailureReason",
                  "-value": ""
                },
                {
                  "-name": "RandomSeed",
                  "-value": "1641395139"
                },
                {
                  "-name": "RandomizeAllSpecs",
                  "-value": "false"
                },
                {
                  "-name": "LabelFilter",
                  "-value": ""
                },
                {
                  "-name": "FocusStrings",
                  "-value": "networking,networking"
                },
                {
                  "-name": "SkipStrings",
                  "-value": ""
                },
                {
                  "-name": "FocusFiles",
                  "-value": ""
                },
                {
                  "-name": "SkipFiles",
                  "-value": ""
                },
                {
                  "-name": "FailOnPending",
                  "-value": "false"
                },
                {
                  "-name": "FailFast",
                  "-value": "false"
                },
                {
                  "-name": "FlakeAttempts",
                  "-value": "0"
                },
                {
                  "-name": "EmitSpecProgress",
                  "-value": "false"
                },
                {
                  "-name": "DryRun",
                  "-value": "false"
                },
                {
                  "-name": "ParallelTotal",
                  "-value": "1"
                }
              ]
            },
            "testcase": [
              {
                "-classname": "CNF Certification Test Suite",
                "-name": "[BeforeSuite]",
                "-status": "passed",
                "-time": "0.383125948"
              },
              {
                "-classname": "CNF Certification Test Suite",
                "-name": "[It] networking Both Pods are on the Default network Testing Default network connectivity networking-icmpv4-connectivity",
                "-status": "passed",
                "-time": "11.571668866",
                "system-err": "Pod test-5b76f77fd5-7d6tb, container test selected to initiate ping tests\n***Test for Network attachment: default\nFrom initiating container: 10.244.151.4 ( node:minikube-m03 ns:tnf podName:test-5b76f77fd5-7d6tb containerName:test containerUID:09acaabb89c51902e623943d3e91e690072f215e1bca231e2c0fc569d9f828ce containerRuntime:docker )\n--\u003e To target container: 10.244.120.67 ( node:minikube ns:tnf podName:test-5b76f77fd5-cg2ll containerName:test containerUID:97ece3466ed2381ac4f9ca90c69ea5bf3dd9f594735b7d70faa1ad245f7f9ccd containerRuntime:docker )\n\n\n�[1mSTEP:�[0m Ping tests on network default. Number of target IPs: 1 �[38;5;243m01/05/22 09:05:47.776�[0m\n�[1mSTEP:�[0m a Ping is issued from test-5b76f77fd5-7d6tb(test) 10.244.151.4 to test-5b76f77fd5-cg2ll(test) 10.244.120.67 �[38;5;243m01/05/22 09:05:47.776�[0m",
                "system-out": "Report Entries:\nBy Step\n/home/deliedit/redhat/github/david/cnftest/test-network-function/networking/suite.go:180\n2022-01-05T09:05:47.776676106-06:00\n\u0026{Text:Ping tests on network default. Number of target IPs: 1 Duration:0s}\n--\nBy Step\n/home/deliedit/redhat/github/david/cnftest/test-network-function/networking/suite.go:182\n2022-01-05T09:05:47.77673356-06:00\n\u0026{Text:a Ping is issued from test-5b76f77fd5-7d6tb(test) 10.244.151.4 to test-5b76f77fd5-cg2ll(test) 10.244.120.67 Duration:0s}"
              },
              {
                "-classname": "CNF Certification Test Suite",
                "-name": "[It] networking Both Pods are connected via a Multus Overlay Network Testing Multus network connectivity networking-icmpv4-connectivity",
                "-status": "passed",
                "-time": "4.159608689",
                "system-err": "Pod test-5b76f77fd5-7d6tb, container test selected to initiate ping tests\n***Test for Network attachment: tnf/macvlan-conf\nFrom initiating container: 192.168.1.3 ( node:minikube-m03 ns:tnf podName:test-5b76f77fd5-7d6tb containerName:test containerUID:09acaabb89c51902e623943d3e91e690072f215e1bca231e2c0fc569d9f828ce containerRuntime:docker )\n--\u003e To target container: 192.168.1.2 ( node:minikube ns:tnf podName:test-5b76f77fd5-cg2ll containerName:test containerUID:97ece3466ed2381ac4f9ca90c69ea5bf3dd9f594735b7d70faa1ad245f7f9ccd containerRuntime:docker )\n\n\n�[1mSTEP:�[0m Ping tests on network tnf/macvlan-conf. Number of target IPs: 1 �[38;5;243m01/05/22 09:05:51.926�[0m\n�[1mSTEP:�[0m a Ping is issued from test-5b76f77fd5-7d6tb(test) 192.168.1.3 to test-5b76f77fd5-cg2ll(test) 192.168.1.2 �[38;5;243m01/05/22 09:05:51.926�[0m",
                "system-out": "Report Entries:\nBy Step\n/home/deliedit/redhat/github/david/cnftest/test-network-function/networking/suite.go:180\n2022-01-05T09:05:51.926185606-06:00\n\u0026{Text:Ping tests on network tnf/macvlan-conf. Number of target IPs: 1 Duration:0s}\n--\nBy Step\n/home/deliedit/redhat/github/david/cnftest/test-network-function/networking/suite.go:182\n2022-01-05T09:05:51.926216747-06:00\n\u0026{Text:a Ping is issued from test-5b76f77fd5-7d6tb(test) 192.168.1.3 to test-5b76f77fd5-cg2ll(test) 192.168.1.2 Duration:0s}"
              },
              {
                "-classname": "CNF Certification Test Suite",
                "-name": "[It] networking Should not have type of nodePort networking-service-type",
                "-status": "passed",
                "-time": "0.11867667",
                "system-err": "�[1mSTEP:�[0m Testing services in namespace tnf �[38;5;243m01/05/22 09:05:56.086�[0m",
                "system-out": "Report Entries:\nBy Step\n/home/deliedit/redhat/github/david/cnftest/test-network-function/networking/suite.go:303\n2022-01-05T09:05:56.086758034-06:00\n\u0026{Text:Testing services in namespace tnf Duration:0s}"
              },
              {
                "-classname": "CNF Certification Test Suite",
                "-name": "[AfterSuite]",
                "-status": "passed",
                "-time": "0.209926976"
              }
            ]
          }
        }
      },
      "testsExtraInfo": []
    },
    "results": {
      "networking-Both_Pods_are_connected_via_a_Multus_Overlay_Network-Testing_Multus_network_connectivity-networking-icmpv4-connectivity": [
        {
          "CapturedTestOutput": "Pod test-5b76f77fd5-7d6tb, container test selected to initiate ping tests\n***Test for Network attachment: tnf/macvlan-conf\nFrom initiating container: 192.168.1.3 ( node:minikube-m03 ns:tnf podName:test-5b76f77fd5-7d6tb containerName:test containerUID:09acaabb89c51902e623943d3e91e690072f215e1bca231e2c0fc569d9f828ce containerRuntime:docker )\n--\u003e To target container: 192.168.1.2 ( node:minikube ns:tnf podName:test-5b76f77fd5-cg2ll containerName:test containerUID:97ece3466ed2381ac4f9ca90c69ea5bf3dd9f594735b7d70faa1ad245f7f9ccd containerRuntime:docker )\n\n\n\u001b[1mSTEP:\u001b[0m Ping tests on network tnf/macvlan-conf. Number of target IPs: 1 \u001b[38;5;243m01/05/22 09:05:51.926\u001b[0m\n\u001b[1mSTEP:\u001b[0m a Ping is issued from test-5b76f77fd5-7d6tb(test) 192.168.1.3 to test-5b76f77fd5-cg2ll(test) 192.168.1.2 \u001b[38;5;243m01/05/22 09:05:51.926\u001b[0m\n",
          "duration": 4159608689,
          "endTime": "2022-01-05 09:05:56.085731342 -0600 CST m=+16.117725012",
          "failureLineContent": "",
          "failureLocation": ":0",
          "failureReason": "",
          "startTime": "2022-01-05 09:05:51.926122652 -0600 CST m=+11.958116323",
          "state": "passed",
          "testID": {
            "url": "http://test-network-function.com/testcases/networking/icmpv4-connectivity",
            "version": "v1.0.0"
          },
          "testText": "http://test-network-function.com/testcases/networking/icmpv4-connectivity checks that each CNF Container is able to communicate via ICMPv4 on the Default OpenShift network.  This\ntest case requires the Deployment of the debug daemonset.\n"
        }
      ],
      "networking-Both_Pods_are_on_the_Default_network-Testing_Default_network_connectivity-networking-icmpv4-connectivity": [
        {
          "CapturedTestOutput": "Pod test-5b76f77fd5-7d6tb, container test selected to initiate ping tests\n***Test for Network attachment: default\nFrom initiating container: 10.244.151.4 ( node:minikube-m03 ns:tnf podName:test-5b76f77fd5-7d6tb containerName:test containerUID:09acaabb89c51902e623943d3e91e690072f215e1bca231e2c0fc569d9f828ce containerRuntime:docker )\n--\u003e To target container: 10.244.120.67 ( node:minikube ns:tnf podName:test-5b76f77fd5-cg2ll containerName:test containerUID:97ece3466ed2381ac4f9ca90c69ea5bf3dd9f594735b7d70faa1ad245f7f9ccd containerRuntime:docker )\n\n\n\u001b[1mSTEP:\u001b[0m Ping tests on network default. Number of target IPs: 1 \u001b[38;5;243m01/05/22 09:05:47.776\u001b[0m\n\u001b[1mSTEP:\u001b[0m a Ping is issued from test-5b76f77fd5-7d6tb(test) 10.244.151.4 to test-5b76f77fd5-cg2ll(test) 10.244.120.67 \u001b[38;5;243m01/05/22 09:05:47.776\u001b[0m\n",
          "duration": 11571668866,
          "endTime": "2022-01-05 09:05:51.925974393 -0600 CST m=+11.957968062",
          "failureLineContent": "",
          "failureLocation": ":0",
          "failureReason": "",
          "startTime": "2022-01-05 09:05:40.354305518 -0600 CST m=+0.386299196",
          "state": "passed",
          "testID": {
            "url": "http://test-network-function.com/testcases/networking/icmpv4-connectivity",
            "version": "v1.0.0"
          },
          "testText": "http://test-network-function.com/testcases/networking/icmpv4-connectivity checks that each CNF Container is able to communicate via ICMPv4 on the Default OpenShift network.  This\ntest case requires the Deployment of the debug daemonset.\n"
        }
      ],
      "networking-Should_not_have_type_of_nodePort-networking-service-type": [
        {
          "CapturedTestOutput": "\u001b[1mSTEP:\u001b[0m Testing services in namespace tnf \u001b[38;5;243m01/05/22 09:05:56.086\u001b[0m\n",
          "duration": 118676670,
          "endTime": "2022-01-05 09:05:56.204577966 -0600 CST m=+16.236571646",
          "failureLineContent": "",
          "failureLocation": ":0",
          "failureReason": "",
          "startTime": "2022-01-05 09:05:56.085901306 -0600 CST m=+16.117894976",
          "state": "passed",
          "testID": {
            "url": "http://test-network-function.com/testcases/networking/service-type",
            "version": "v1.0.0"
          },
          "testText": "http://test-network-function.com/testcases/networking/service-type tests that each CNF Service does not utilize NodePort(s)."
        }
      ]
    },
    "versions": {
      "k8s": "",
      "ocClient": "",
      "ocp": "",
      "tnf": "Unreleased build post v3.2.0",
      "tnfGitCommit": "1d8dd62ef4dc09538a8299071231ebb366724527"
    }
  }
}