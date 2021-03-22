{
  "command": "ssh",
  "args": [{{range $index, $element := .SSH_ARGS}}{{if $index}},{{end}}"{{$element}}"{{end}}],
  "timeout": 2000000000
}