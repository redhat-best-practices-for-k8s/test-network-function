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

package autodiscover

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/test-network-function/test-network-function/pkg/config/configsections"
)

const (
	resourceTypeCSV = "csv"
)

// CSVList holds the data from an `oc get csv -o json` command
type CSVList struct {
	Items []CSVResource `json:"items"`
}

// CSVResource is a single entry from an `oc get csv -o json` command
type CSVResource struct {
	Metadata struct {
		Name        string            `json:"name"`
		Namespace   string            `json:"namespace"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
	} `json:"metadata"`
}

func (csv *CSVResource) hasAnnotation(annotationKey string) (present bool) {
	_, present = csv.Metadata.Annotations[annotationKey]
	return
}

// GetAnnotationValue will get the value stored in the given annotation and
// Unmarshal it into the given var `v`.
func (csv *CSVResource) GetAnnotationValue(annotationKey string, v interface{}) (err error) {
	if !csv.hasAnnotation(annotationKey) {
		return fmt.Errorf("failed to find annotation '%s' on CSV '%s/%s'", annotationKey, csv.Metadata.Namespace, csv.Metadata.Name)
	}
	val := csv.Metadata.Annotations[annotationKey]
	err = jsonUnmarshal([]byte(val), v)
	if err != nil {
		return csv.annotationUnmarshalError(annotationKey, err)
	}
	return
}

func (csv *CSVResource) annotationUnmarshalError(annotationKey string, err error) error {
	return fmt.Errorf("error (%s) attempting to unmarshal value of annotation '%s' on CSV '%s/%s'",
		err, annotationKey, csv.Metadata.Namespace, csv.Metadata.Name)
}

// GetCSVsByLabel will return all CSVs with a given label value. If `labelValue` is an empty string, all CSVs with that
// label will be returned, regardless of the labels value.
func GetCSVsByLabel(labelName, labelValue, namespace string) (*CSVList, error) {
	out := executeOcGetCommand(resourceTypeCSV, buildLabelQuery(configsections.Label{Prefix: tnfLabelPrefix, Name: labelName, Value: labelValue}), namespace)

	log.Debug("JSON output for all pods labeled with: ", labelName)
	log.Debug("Command: ", out)

	var csvList CSVList
	err := jsonUnmarshal([]byte(out), &csvList)
	if err != nil {
		return nil, err
	}

	return &csvList, nil
}
