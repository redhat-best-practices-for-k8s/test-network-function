package junit

import (
	"bytes"
	j "encoding/json"
	xj "github.com/basgys/goxml2json"
	"os"
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
