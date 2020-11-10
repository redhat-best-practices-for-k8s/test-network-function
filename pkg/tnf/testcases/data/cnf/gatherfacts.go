package cnf

// GatherPodFactsJSON test templates for collecting facts
var GatherPodFactsJSON = string(`{
  "testcase": [
    {
      "name": "CONTAINER_COUNT",
      "skiptest": false,
      "command": "oc get pod %s -n %s -o json | jq -r '.spec.containers | length'",
      "action": "allow",
      "resulttype": "integer",
      "expectedType": "regex",
      "expectedstatus": [
        "DIGIT"
      ]
    },
    {
      "name": "SERVICE_ACCOUNT_NAME",
      "skiptest": false,
      "command": "oc get pod %s -n %s -o json | jq -r '.spec.serviceAccountName'",
      "action": "allow",
      "resulttype": "string",
      "expectedType": "regex",
      "expectedstatus": [
        "ALLOW_ALL"
      ]
    }
  ]
}`)
