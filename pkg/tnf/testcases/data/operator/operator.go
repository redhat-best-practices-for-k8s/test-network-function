package operator

//OperatorJSON test templates for collecting operator status
var OperatorJSON = string(`{
  "testcase": [
    {
      "name": "CSV_INSTALLED",
      "skiptest": true,
      "command": "oc get csv %s -n %s -o json | jq -r '.status.phase'",
      "action": "allow",
      "resulttype": "string",
      "expectedstatus": [
        "Succeeded"
      ]
    }
  ]
}`)
