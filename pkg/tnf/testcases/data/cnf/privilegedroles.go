package cnf

//RolesJSON test templates for testing permission
var RolesJSON = string(`{
  "testcase": [
    {
      "name": "CLUSTER_ROLE_BINDING_BY_SA",
      "skiptest": true,
      "command": "oc get clusterrolebinding -n %s -o json | jq --arg name 'ServiceAccount' --arg null ',null,' --arg subjects 'subjects' --arg ns '%s' --arg sa '%s' -jr 'if (.items|length)>0 then .items[] | if (has($subjects)) then .subjects[] | select((.namespace==$ns) and (.kind==$name) and (.name==$sa)).name else $null end else $null end'",
      "action": "deny",
	  "loop": 0,
      "resulttype": "array",
      "expectedtype": "function",
      "expectedstatus": [
        "FN_SERVICE_ACCOUNT_NAME"
      ]
    },
    {
      "name": "ROLE_BINDING_BY_SA",
      "skiptest": true,
      "loop": 0,
      "command": "oc get rolebinding -n %s -o json | jq --arg name 'ServiceAccount' --arg null ',null,' --arg ns '%s' --arg subjects 'subjects' --arg sa '%s' -jr 'if (.items|length)>0 then .items[] | if (has($subjects)) then .subjects[] | select((.namespace==$ns) and (.kind==$name) and (.name==$sa)).name else $null end else $null end'",
      "action": "allow",
      "resulttype": "array",
      "expectedtype": "function",
      "expectedstatus": [
        "FN_SERVICE_ACCOUNT_NAME",
        "null"
      ]
    }
  ]
}`)
