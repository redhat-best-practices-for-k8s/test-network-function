// Copyright (C) 2021 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package interactive

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"time"

	expect "github.com/ryandgoulding/goexpect"
	"github.com/test-network-function/test-network-function/pkg/jsonschema"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
)

const (
	// PTYSchemaFileName is the filename for the generic pty JSON schema.
	PTYSchemaFileName = "generic-pty.schema.json"
)

// SpawnGenericPTYFromYAMLFile attempts to spawn an interactive PTY given the supplied file based on
// generic-pty.schema.json.  schemaPath should always be the path to generic-pty.schema.json relative to the execution
// entry-point, which will vary for unit tests, executables, and test suites.  If the supplied file does not conform to
// the generic-pty.schema.json schema, creation fails and the result is returned to the caller for further inspection.
func SpawnGenericPTYFromYAMLFile(ptyPath, schemaPath string, spawner *Spawner) (*Context, *gojsonschema.Result, error) {
	ptyBytes, err := ioutil.ReadFile(ptyPath)
	if err != nil {
		return nil, nil, err
	}
	return SpawnGenericPTYFromYAML(ptyBytes, schemaPath, spawner)
}

// SpawnGenericPTYFromYAML attempts to spawn an interactive PTY given the supplied bytes based on
// generic-pty.schema.json.  schemaPath should always be the path to generic-pty.schema.json relative to the execution
// entry-point, which will vary for unit tests, executables, and test suites.  If the supplied file does not conform to
// the generic-pty.schema.json schema, creation fails and the result is returned to the caller for further inspection.
func SpawnGenericPTYFromYAML(inputBytes []byte, schemaPath string, spawner *Spawner) (*Context, *gojsonschema.Result, error) {
	ptyCommandConfig, result, err := newPTYCommandFromYAML(inputBytes, schemaPath)

	// Only continue if there were no errors and the input is valid.
	if err != nil || !result.Valid() {
		return nil, result, err
	}

	context, err := spawnGenericPTY(spawner, ptyCommandConfig.Command, ptyCommandConfig.Args, ptyCommandConfig.Timeout)
	return context, result, err
}

// SpawnGenericPTYFromYAMLTemplate attempts to spawn an interactive PTY by rendering the supplied template/values and
// validating schema conformance based on generic-pty.schema.json.  schemaPath should always be the path to
// generic-pty.schema.json relative to the execution entry-point, which will vary for unit tests, executables, and test
// suites.  If the supplied template/values do not conform to the generic-pty.schema.json schema, creation fails and the
// result is returned to the caller for further inspection.
func SpawnGenericPTYFromYAMLTemplate(templateFile, valuesFile, schemaPath string, spawner *Spawner) (*Context, *gojsonschema.Result, error) {
	tplBytes, err := ioutil.ReadFile(valuesFile)
	if err != nil {
		return nil, nil, err
	}

	values := make(map[string]interface{})
	err = yaml.Unmarshal(tplBytes, values)

	if err != nil {
		return nil, nil, err
	}

	templateBytes, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return nil, nil, err
	}
	// Note: "tpl" just names the template.  It is arbitrary, and doesn't really matter.
	t, err := template.New("tpl").Option("missingkey=error").Parse(string(templateBytes))
	if err != nil {
		return nil, nil, err
	}

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "tpl", values)
	if err != nil {
		return nil, nil, err
	}

	return SpawnGenericPTYFromYAML(buf.Bytes(), schemaPath, spawner)
}

// spawnGenericPTY spawns a generic command as a pseudo-terminal (PTY).
func spawnGenericPTY(spawner *Spawner, command string, args []string, timeout time.Duration, opts ...expect.Option) (*Context, error) {
	return (*spawner).Spawn(command, args, timeout, opts...)
}

// PTYCommand represents any PTY command.
type PTYCommand struct {
	// Command represents any PTY command.  Commands are not necessarily unix-based.
	Command string `json:"command" yaml:"command"`

	// Args are the optional arguments to Command.
	Args []string `json:"args" yaml:"args"`

	// Timeout is the command timeout for the PTY Expecter.
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
}

// newPTYCommandFromYAML creates a PTYCommand from YAML configuration bytes.
func newPTYCommandFromYAML(configurationBytes []byte, schemaPath string) (*PTYCommand, *gojsonschema.Result, error) {
	result, err := jsonschema.ValidateJSONAgainstSchema(configurationBytes, schemaPath)

	if err != nil || !result.Valid() {
		return nil, result, err
	}

	command := &PTYCommand{}
	err = yaml.Unmarshal(configurationBytes, &command)

	return command, result, err
}
