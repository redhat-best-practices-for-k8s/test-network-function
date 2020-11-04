package data

//PrivilegedPodJSON test templates for privileged pods
var PrivilegedPodJSON = string(`{
  "testcase": [
    {
      "name": "HOST_NETWORK_CHECK",
      "skiptest": true,
      "command": "oc get pod  %s  -n %s -o json  | jq -r '.spec.hostNetwork'",
      "action": "allow",
      "expectedstatus": [
        "NULL_FALSE"
      ]
    },
    {
      "name": "HOST_PORT_CHECK",
      "skiptest": true,
      "command": "oc get pod  %s  -n %s -o json  | jq -r '.spec.hostPort'",
      "action": "allow",
      "expectedstatus": [
        "NULL_FALSE"
      ]
    },
    {
      "name": "HOST_PATH_CHECK",
      "skiptest": true,
      "command": "oc get pod  %s  -n %s -o json  | jq -r '.spec.hostpath.path'",
      "action": "allow",
      "expectedstatus": [
        "NOT_SET"
      ]
    },
    {
      "name": "HOST_IPC_CHECK",
      "skiptest": true,
      "command": "oc get pod  %s  -n %s -o json  | jq -r '.spec.hostipc'",
      "action": "allow",
      "expectedstatus": [
        "NULL_FALSE"
      ]
    },
    {
      "name": "HOST_PID_CHECK",
      "skiptest": true,
      "command": "oc get pod  %s  -n %s -o json  | jq -r '.spec.hostpid'",
      "action": "allow",
      "expectedstatus": [
        "NULL_FALSE"
      ]
    },
    {
      "name": "CAPABILITY_CHECK",
      "skiptest": true,
      "command": "oc get pod  %s  -n %s -o json  | jq -r '.spec.containers[0].securityContext.capabilities.add'",
      "resultType": "array",
      "action": "deny",
      "expectedstatus": [
        "NET_ADMIN",
        "SYS_ADMIN"
      ]
    },
    {
      "name": "ROOT_CHECK",
      "skiptest": true,
      "command": "oc get pod  %s  -n %s -o json  | jq -r '.spec.containers[0].securityContext.runAsUser'",
      "resulttype": "string",
      "action": "allow",
      "expectedstatus": [
        "NON_ROOT_USER"
      ]
    },
    {
      "name": "PRIVILEGE_ESCALATION",
      "skiptest": true,
      "command": "oc get pod  %s  -n %s -o json  | jq -r '.spec.containers[0].securityContext.allowPrivilegeEscalation'",
      "action": "allow",
      "expectedstatus": [
        "NULL_FALSE"
      ]
    }
  ]
}`)
