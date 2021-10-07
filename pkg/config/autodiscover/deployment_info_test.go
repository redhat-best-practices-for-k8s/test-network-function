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
	"encoding/json"
	"log"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDeploymentFile = "testdeployment.json"
)

var (
	testDeploymentFilePath = path.Join(filePath, testDeploymentFile)
)

func loadDeployment(filePath string) (deployment DeploymentResource) {
	contents, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("error (%s) loading DeploymentResource %s for testing", err, filePath)
	}
	err = json.Unmarshal(contents, &deployment)
	if err != nil {
		log.Fatalf("error (%s) unmarshalling DeploymentResource %s for testing", err, filePath)
	}
	return
}

func TestPodGetAnnotationValue1(t *testing.T) {
	deployment := loadDeployment(testDeploymentFilePath)

	assert.Equal(t, "test", deployment.GetName())
	assert.Equal(t, "tnf", deployment.GetNamespace())
	assert.Equal(t, 2, deployment.GetReplicas())

	labels := deployment.GetLabels()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, "test", labels["app"])
}
