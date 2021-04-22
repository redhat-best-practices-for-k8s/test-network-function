// Copyright (C) 2020-2021 Red Hat, Inc.
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

package junit

import (
	"bytes"
	j "encoding/json"
	"os"

	xj "github.com/basgys/goxml2json"
)

// ExportJUnitAsJSON attempts to read a JUnit XML file and converts it to a generic JSON map.
func ExportJUnitAsJSON(junitFilename string) (map[string]interface{}, error) {
	xmlReader, err := os.Open(junitFilename)
	// An error is encountered reading the file.
	if err != nil {
		return nil, err
	}

	junitJSONBuffer, err := xj.Convert(xmlReader)
	// An error is encountered translating from XML to JSON.
	if err != nil {
		return nil, err
	}

	jsonMap, err := convertJSONBytesToMap(junitJSONBuffer)
	// An error is encountered unmarshalling the data.
	if err != nil {
		return nil, err
	}

	return jsonMap, err
}

// convertJSONBytesToMap is a utility function to convert a bytes.Buffer to a generic JSON map.
func convertJSONBytesToMap(junitJSONBuffer *bytes.Buffer) (map[string]interface{}, error) {
	jsonMap := make(map[string]interface{})
	err := j.Unmarshal(junitJSONBuffer.Bytes(), &jsonMap)
	return jsonMap, err
}
