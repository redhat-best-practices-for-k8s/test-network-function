apiVersion: examplecnf.openshift.io/v1
kind: TRexApp
metadata:
  name: {{.CR_NAME}}
  namespace: {{.NAMESPACE}}
spec:
  duration: {{.DURATION}}
  enableLb: true
  imagePullPolicy: Always
  lbMacs:
    - 40:04:0f:f1:89:01
    - 40:04:0f:f1:89:02
  org: krsacme
  packetRate: {{.PKT.RATE}}
  packetSize: {{.PKT.SIZE}}
  registry: quay.io
  trexApp: false
  trexProfileConfigMap: null
  trexProfileName: null
  version: v0.2.1