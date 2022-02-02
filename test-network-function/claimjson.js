var initialjson={
  "claim": {
    "configurations": {
      "acceptedKernelTaints": [
        {
          "module": "vboxsf"
        },
        {
          "module": "vboxguest"
        }
      ],
      "certifiedcontainerinfo": [
        {
          "digest": "",
          "name": "nginx-116",
          "repository": "rhel8",
          "tag": "1-112"
        }
      ],
      "certifiedoperatorinfo": [
        {
          "name": "etcd",
          "organization": "community-operators"
        }
      ],
      "checkDiscoveredContainerCertificationStatus": false,
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
            "ImageSource": {
              "Registry": "quay.io",
              "digest": "",
              "name": "debug-partner",
              "repository": "testnetworkfunction",
              "tag": "latest"
            },
            "Oc": null,
            "containerName": "container-00",
            "containerRuntime": "docker",
            "containerUID": "37de9a5a47cf02cd81b9145ca6a7b8cd938057e843a98638bd0b7db3d3e151b9",
            "namespace": "default",
            "nodeName": "minikube-m03",
            "podName": "debug-pqq88"
          },
          {
            "ImageSource": {
              "Registry": "quay.io",
              "digest": "",
              "name": "debug-partner",
              "repository": "testnetworkfunction",
              "tag": "latest"
            },
            "Oc": null,
            "containerName": "container-00",
            "containerRuntime": "docker",
            "containerUID": "7b6f777c4cc1f20879dda149f6b67fa35140c07cbd570f5565bdc0e73fe8022f",
            "namespace": "default",
            "nodeName": "minikube",
            "podName": "debug-sc2cx"
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
            "ImageSource": {
              "Registry": "quay.io",
              "digest": "",
              "name": "cnf-test-partner",
              "repository": "testnetworkfunction",
              "tag": "latest"
            },
            "Oc": {},
            "containerName": "test",
            "containerRuntime": "docker",
            "containerUID": "2f34cb80222069d488db747bb6d5463c507f0a96daca78a0465faa3e89250d62",
            "namespace": "tnf",
            "nodeName": "minikube",
            "podName": "test-54bc4c6d7-qhtx6"
          },
          {
            "ImageSource": {
              "Registry": "quay.io",
              "digest": "",
              "name": "cnf-test-partner",
              "repository": "testnetworkfunction",
              "tag": "latest"
            },
            "Oc": {},
            "containerName": "test",
            "containerRuntime": "docker",
            "containerUID": "ef9a9cf2946713197b77396e7df83e0b41055f7c5f7c5bf6a136469e605d6eca",
            "namespace": "tnf",
            "nodeName": "minikube-m03",
            "podName": "test-54bc4c6d7-s6zt7"
          }
        ],
        "deploymentsUnderTest": [
          {
            "Hpa": {
              "HpaName": "",
              "MaxReplicas": 0,
              "MinReplicas": 0
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
            "Org": "nginx-operator-catalog",
            "Version": "v0.0.1",
            "name": "nginx-operator.v0.0.1",
            "namespace": "tnf",
            "packag": "nginx-operator.v0.0.1",
            "subscriptionName": "nginx-operator-v0-0-1-sub",
            "tests": [
              "OPERATOR_STATUS"
            ]
          }
        ],
        "podsUnderTest": [
          {
            "IsManaged": true,
            "containercount": 1,
            "containerfornettests": [
              {
                "ImageSource": {
                  "Registry": "quay.io",
                  "digest": "",
                  "name": "cnf-test-partner",
                  "repository": "testnetworkfunction",
                  "tag": "latest"
                },
                "Oc": null,
                "containerName": "test",
                "containerRuntime": "docker",
                "containerUID": "2f34cb80222069d488db747bb6d5463c507f0a96daca78a0465faa3e89250d62",
                "namespace": "tnf",
                "nodeName": "minikube",
                "podName": "test-54bc4c6d7-qhtx6"
              }
            ],
            "defaultNetworkDevice": "eth0",
            "defaultnetworkipaddress": "10.244.120.67",
            "multusIpAddressesPerNet": {
              "tnf/macvlan-conf": [
                "192.168.1.2"
              ]
            },
            "name": "test-54bc4c6d7-qhtx6",
            "namespace": "tnf",
            "serviceaccount": "default",
            "tests": [
              "PRIVILEGED_POD",
              "PRIVILEGED_ROLE"
            ]
          },
          {
            "IsManaged": true,
            "containercount": 1,
            "containerfornettests": [
              {
                "ImageSource": {
                  "Registry": "quay.io",
                  "digest": "",
                  "name": "cnf-test-partner",
                  "repository": "testnetworkfunction",
                  "tag": "latest"
                },
                "Oc": null,
                "containerName": "test",
                "containerRuntime": "docker",
                "containerUID": "ef9a9cf2946713197b77396e7df83e0b41055f7c5f7c5bf6a136469e605d6eca",
                "namespace": "tnf",
                "nodeName": "minikube-m03",
                "podName": "test-54bc4c6d7-s6zt7"
              }
            ],
            "defaultNetworkDevice": "eth0",
            "defaultnetworkipaddress": "10.244.151.2",
            "multusIpAddressesPerNet": {
              "tnf/macvlan-conf": [
                "192.168.1.3"
              ]
            },
            "name": "test-54bc4c6d7-s6zt7",
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
      "endTime": "2022-02-02T17:45:17+00:00",
      "startTime": "2022-02-02T17:44:50+00:00"
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
          "-tests": "6",
          "-time": "26.995914041",
          "testsuite": {
            "-disabled": "0",
            "-errors": "0",
            "-failures": "0",
            "-name": "CNF Certification Test Suite",
            "-package": "/home/speretz/test-network-function/test-network-function",
            "-skipped": "0",
            "-tests": "6",
            "-time": "26.995914041",
            "-timestamp": "2022-02-02T19:44:50",
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
                  "-name": "SuiteLabels",
                  "-value": "[]"
                },
                {
                  "-name": "RandomSeed",
                  "-value": "1643823890"
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
                },
                {
                  "-name": "OutputInterceptorMode",
                  "-value": ""
                }
              ]
            },
            "testcase": [
              {
                "-classname": "CNF Certification Test Suite",
                "-name": "[BeforeSuite]",
                "-status": "passed",
                "-time": "0.481533895"
              },
              {
                "-classname": "CNF Certification Test Suite",
                "-name": "[It] networking Both Pods are on the Default network Testing Default network connectivity networking-icmpv4-connectivity [networking-icmpv4-connectivity]",
                "-status": "passed",
                "-time": "21.724627579",
                "system-err": "Pod test-54bc4c6d7-qhtx6, container test selected to initiate ping tests\n***Test for Network attachment: default\nFrom initiating container: 10.244.120.67 ( node:minikube ns:tnf podName:test-54bc4c6d7-qhtx6 containerName:test containerUID:2f34cb80222069d488db747bb6d5463c507f0a96daca78a0465faa3e89250d62 containerRuntime:docker )\n--\u003e To target container: 10.244.151.2 ( node:minikube-m03 ns:tnf podName:test-54bc4c6d7-s6zt7 containerName:test containerUID:ef9a9cf2946713197b77396e7df83e0b41055f7c5f7c5bf6a136469e605d6eca containerRuntime:docker )\n\n\n�[1mSTEP:�[0m Ping tests on network default. Number of target IPs: 1 �[38;5;243m02/02/22 19:45:08.137�[0m\n�[1mSTEP:�[0m a Ping is issued from test-54bc4c6d7-qhtx6(test) 10.244.120.67 to test-54bc4c6d7-s6zt7(test) 10.244.151.2 �[38;5;243m02/02/22 19:45:08.137�[0m",
                "system-out": "Report Entries:\nBy Step\n/home/speretz/test-network-function/test-network-function/networking/suite.go:205\n2022-02-02T19:45:08.13726009+02:00\n\u0026{Text:Ping tests on network default. Number of target IPs: 1 Duration:0s}\n--\nBy Step\n/home/speretz/test-network-function/test-network-function/networking/suite.go:207\n2022-02-02T19:45:08.137328648+02:00\n\u0026{Text:a Ping is issued from test-54bc4c6d7-qhtx6(test) 10.244.120.67 to test-54bc4c6d7-s6zt7(test) 10.244.151.2 Duration:0s}"
              },
              {
                "-classname": "CNF Certification Test Suite",
                "-name": "[It] networking Both Pods are connected via a Multus Overlay Network Testing Multus network connectivity networking-icmpv4-connectivity-multus [networking-icmpv4-connectivity-multus]",
                "-status": "passed",
                "-time": "4.160147993",
                "system-err": "Pod test-54bc4c6d7-qhtx6, container test selected to initiate ping tests\n***Test for Network attachment: tnf/macvlan-conf\nFrom initiating container: 192.168.1.2 ( node:minikube ns:tnf podName:test-54bc4c6d7-qhtx6 containerName:test containerUID:2f34cb80222069d488db747bb6d5463c507f0a96daca78a0465faa3e89250d62 containerRuntime:docker )\n--\u003e To target container: 192.168.1.3 ( node:minikube-m03 ns:tnf podName:test-54bc4c6d7-s6zt7 containerName:test containerUID:ef9a9cf2946713197b77396e7df83e0b41055f7c5f7c5bf6a136469e605d6eca containerRuntime:docker )\n\n\n�[1mSTEP:�[0m Ping tests on network tnf/macvlan-conf. Number of target IPs: 1 �[38;5;243m02/02/22 19:45:12.287�[0m\n�[1mSTEP:�[0m a Ping is issued from test-54bc4c6d7-qhtx6(test) 192.168.1.2 to test-54bc4c6d7-s6zt7(test) 192.168.1.3 �[38;5;243m02/02/22 19:45:12.287�[0m",
                "system-out": "Report Entries:\nBy Step\n/home/speretz/test-network-function/test-network-function/networking/suite.go:205\n2022-02-02T19:45:12.287538369+02:00\n\u0026{Text:Ping tests on network tnf/macvlan-conf. Number of target IPs: 1 Duration:0s}\n--\nBy Step\n/home/speretz/test-network-function/test-network-function/networking/suite.go:207\n2022-02-02T19:45:12.287571492+02:00\n\u0026{Text:a Ping is issued from test-54bc4c6d7-qhtx6(test) 192.168.1.2 to test-54bc4c6d7-s6zt7(test) 192.168.1.3 Duration:0s}"
              },
              {
                "-classname": "CNF Certification Test Suite",
                "-name": "[It] networking Should not have type of nodePort networking-service-type [networking-service-type]",
                "-status": "passed",
                "-time": "0.079875961",
                "system-err": "�[1mSTEP:�[0m Testing services in namespace tnf �[38;5;243m02/02/22 19:45:16.448�[0m",
                "system-out": "Report Entries:\nBy Step\n/home/speretz/test-network-function/test-network-function/networking/suite.go:332\n2022-02-02T19:45:16.448445547+02:00\n\u0026{Text:Testing services in namespace tnf Duration:0s}"
              },
              {
                "-classname": "CNF Certification Test Suite",
                "-name": "[It] networking Should not have type of listen port and declared port networking-service-type",
                "-status": "passed",
                "-time": "0.366401437",
                "system-err": "8080\n8080"
              },
              {
                "-classname": "CNF Certification Test Suite",
                "-name": "[AfterSuite]",
                "-status": "passed",
                "-time": "0.182371245"
              }
            ]
          }
        }
      },
      "testsExtraInfo": []
    },
    "results": {
      "networking-Both_Pods_are_connected_via_a_Multus_Overlay_Network-Testing_Multus_network_connectivity-networking-icmpv4-connectivity-multus": [
        {
          "CapturedTestOutput": "Pod test-54bc4c6d7-qhtx6, container test selected to initiate ping tests\n***Test for Network attachment: tnf/macvlan-conf\nFrom initiating container: 192.168.1.2 ( node:minikube ns:tnf podName:test-54bc4c6d7-qhtx6 containerName:test containerUID:2f34cb80222069d488db747bb6d5463c507f0a96daca78a0465faa3e89250d62 containerRuntime:docker )\n--\u003e To target container: 192.168.1.3 ( node:minikube-m03 ns:tnf podName:test-54bc4c6d7-s6zt7 containerName:test containerUID:ef9a9cf2946713197b77396e7df83e0b41055f7c5f7c5bf6a136469e605d6eca containerRuntime:docker )\n\n\n\u001b[1mSTEP:\u001b[0m Ping tests on network tnf/macvlan-conf. Number of target IPs: 1 \u001b[38;5;243m02/02/22 19:45:12.287\u001b[0m\n\u001b[1mSTEP:\u001b[0m a Ping is issued from test-54bc4c6d7-qhtx6(test) 192.168.1.2 to test-54bc4c6d7-s6zt7(test) 192.168.1.3 \u001b[38;5;243m02/02/22 19:45:12.287\u001b[0m\n",
          "duration": 4160147993,
          "endTime": "2022-02-02 19:45:16.447612381 +0200 IST m=+26.368958306",
          "failureLineContent": "",
          "failureLocation": ":0",
          "failureReason": "",
          "startTime": "2022-02-02 19:45:12.287464372 +0200 IST m=+22.208810313",
          "state": "passed",
          "testID": {
            "url": "http://test-network-function.com/testcases/networking/icmpv4-connectivity-multus",
            "version": "v1.0.0"
          },
          "testText": "http://test-network-function.com/testcases/networking/icmpv4-connectivity-multus checks that each CNF Container is able to communicate via ICMPv4 on the Multus network(s).  This\ntest case requires the Deployment of the debug daemonset."
        }
      ],
      "networking-Both_Pods_are_on_the_Default_network-Testing_Default_network_connectivity-networking-icmpv4-connectivity": [
        {
          "CapturedTestOutput": "Pod test-54bc4c6d7-qhtx6, container test selected to initiate ping tests\n***Test for Network attachment: default\nFrom initiating container: 10.244.120.67 ( node:minikube ns:tnf podName:test-54bc4c6d7-qhtx6 containerName:test containerUID:2f34cb80222069d488db747bb6d5463c507f0a96daca78a0465faa3e89250d62 containerRuntime:docker )\n--\u003e To target container: 10.244.151.2 ( node:minikube-m03 ns:tnf podName:test-54bc4c6d7-s6zt7 containerName:test containerUID:ef9a9cf2946713197b77396e7df83e0b41055f7c5f7c5bf6a136469e605d6eca containerRuntime:docker )\n\n\n\u001b[1mSTEP:\u001b[0m Ping tests on network default. Number of target IPs: 1 \u001b[38;5;243m02/02/22 19:45:08.137\u001b[0m\n\u001b[1mSTEP:\u001b[0m a Ping is issued from test-54bc4c6d7-qhtx6(test) 10.244.120.67 to test-54bc4c6d7-s6zt7(test) 10.244.151.2 \u001b[38;5;243m02/02/22 19:45:08.137\u001b[0m\n",
          "duration": 21724627579,
          "endTime": "2022-02-02 19:45:12.287343266 +0200 IST m=+22.208689207",
          "failureLineContent": "",
          "failureLocation": ":0",
          "failureReason": "",
          "startTime": "2022-02-02 19:44:50.562715704 +0200 IST m=+0.484061628",
          "state": "passed",
          "testID": {
            "url": "http://test-network-function.com/testcases/networking/icmpv4-connectivity",
            "version": "v1.0.0"
          },
          "testText": "http://test-network-function.com/testcases/networking/icmpv4-connectivity checks that each CNF Container is able to communicate via ICMPv4 on the Default OpenShift network.  This\ntest case requires the Deployment of the debug daemonset."
        }
      ],
      "networking-Should_not_have_type_of_listen_port_and_declared_port-networking-service-type": [
        {
          "CapturedTestOutput": "8080\n8080\n",
          "duration": 366401437,
          "endTime": "2022-02-02 19:45:16.894393338 +0200 IST m=+26.815739270",
          "failureLineContent": "",
          "failureLocation": ":0",
          "failureReason": "",
          "startTime": "2022-02-02 19:45:16.527991896 +0200 IST m=+26.449337833",
          "state": "passed",
          "testID": {
            "url": "http://test-network-function.com/testcases/networking/service-type",
            "version": "v1.0.0"
          },
          "testText": "http://test-network-function.com/testcases/networking/service-type tests that each CNF Service does not utilize NodePort(s)."
        }
      ],
      "networking-Should_not_have_type_of_nodePort-networking-service-type": [
        {
          "CapturedTestOutput": "\u001b[1mSTEP:\u001b[0m Testing services in namespace tnf \u001b[38;5;243m02/02/22 19:45:16.448\u001b[0m\n",
          "duration": 79875961,
          "endTime": "2022-02-02 19:45:16.527873117 +0200 IST m=+26.449219042",
          "failureLineContent": "",
          "failureLocation": ":0",
          "failureReason": "",
          "startTime": "2022-02-02 19:45:16.447997139 +0200 IST m=+26.369343081",
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
      "tnfGitCommit": "9fb4686ac0e925e0861d3a95f307686c76b78a84"
    }
  }
}