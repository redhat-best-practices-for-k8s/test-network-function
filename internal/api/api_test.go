package api_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/internal/api"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
)

const (
	id                 = "5ea8cf595a13466876a10215"
	imageName          = "nginx-120"
	marketPlaceOrg     = "redhat-marketplace"
	packageName        = "amq-streams"
	redHatOrg          = "redhat-operators"
	repository         = "rhel8"
	unKnownRepository  = "wrong_repo"
	unKnownImageName   = "wrong_id"
	unknownPackageName = "unknownPackage"
	version            = "4.8"
	jsonResponseFound  = `{
		"data": [
		  {
			"_id": "61ba0db5d095e30ed5db6330",
			"_links": {
			  "rpm_manifest": {
				"href": "/v1/images/id/61ba0db5d095e30ed5db6330/rpm-manifest"
			  },
			  "vulnerabilities": {
				"href": "/v1/images/id/61ba0db5d095e30ed5db6330/vulnerabilities"
			  }
			},
			"architecture": "amd64",
			"brew": {
			  "build": "nginx-120-container-1-7",
			  "completion_date": "2021-12-15T15:39:17+00:00",
			  "nvra": "nginx-120-container-1-7.amd64",
			  "package": "nginx-120-container"
			},
			"certified": false,
			"content_sets": [
			  "rhel-8-for-x86_64-baseos-rpms",
			  "rhel-8-for-x86_64-appstream-rpms"
			],
			"cpe_ids": [
			  "cpe:/a:redhat:enterprise_linux:8::appstream",
			  "cpe:/o:redhat:enterprise_linux:8::baseos",
			  "cpe:/o:redhat:rhel:8.3::baseos",
			  "cpe:/a:redhat:rhel:8.3::appstream"
			],
			"creation_date": "2021-12-15T15:45:57.616000+00:00",
			"docker_image_id": "sha256:b9dbffacfeb14acf36c7da686a0874be4484a473a8993e4aaf72a68b80ea4cf6",
			"freshness_grades": [
			  {
				"creation_date": "2021-12-15T15:46:07.851000+00:00",
				"grade": "A",
				"start_date": "2021-12-15T15:46:00+00:00"
			  }
			],
			"image_id": "sha256:aa34453a6417f8f76423ffd2cf874e9c4a1a5451ac872b78dc636ab54a0ebbc3",
			"last_update_date": "2021-12-21T12:10:34.397000+00:00",
			"object_type": "containerImage",
			"parsed_data": {
			  "architecture": "amd64",
			  "command": "['/bin/sh', '-c', '$STI_SCRIPTS_PATH/usage']",
			  "comment": "",
			  "created": "2021-12-15T15:31:51.564422Z",
			  "docker_version": "1.13.1",
			  "env_variables": [
				"PATH=/opt/app-root/src/bin:/opt/app-root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
				"container=oci",
				"SUMMARY=Platform for running nginx 1.20 or building nginx-based application",
				"DESCRIPTION=Nginx is a web server and a reverse proxy server for HTTP, SMTP, POP3 and IMAP protocols, <snip>",
				"STI_SCRIPTS_URL=image:///usr/libexec/s2i",
				"STI_SCRIPTS_PATH=/usr/libexec/s2i",
				"APP_ROOT=/opt/app-root",
				"HOME=/opt/app-root/src",
				"PLATFORM=el8",
				"NAME=nginx",
				"NGINX_VERSION=1.20",
				"NGINX_SHORT_VER=120",
				"VERSION=0",
				"NGINX_CONFIGURATION_PATH=/opt/app-root/etc/nginx.d",
				"NGINX_CONF_PATH=/etc/nginx/nginx.conf",
				"NGINX_DEFAULT_CONF_PATH=/opt/app-root/etc/nginx.default.d",
				"NGINX_CONTAINER_SCRIPTS_PATH=/usr/share/container-scripts/nginx",
				"NGINX_APP_ROOT=/opt/app-root",
				"NGINX_LOG_PATH=/var/log/nginx",
				"NGINX_PERL_MODULE_PATH=/opt/app-root/etc/perl"
			  ],
			  "labels": [
				{
				  "name": "architecture",
				  "value": "x86_64"
				},
				{
				  "name": "build-date",
				  "value": "2021-12-15T15:30:27.746206"
				},
				{
				  "name": "com.redhat.build-host",
				  "value": "cpt-1003.osbs.prod.upshift.rdu2.redhat.com"
				},
				{
				  "name": "com.redhat.component",
				  "value": "nginx-120-container"
				},
				{
				  "name": "com.redhat.license_terms",
				  "value": "https://www.redhat.com/en/about/red-hat-end-user-license-agreements#UBI"
				},
				{
				  "name": "description",
				  "value": "Nginx is a web server and a reverse proxy server for HTTP, SMTP, POP3 and IMAP protocols, <snip>"
				},
				{
				  "name": "distribution-scope",
				  "value": "public"
				},
				{
				  "name": "help",
				  "value": "For more information visit https://github.com/sclorg/nginx-container"
				},
				{
				  "name": "io.k8s.description",
				  "value": "Nginx is a web server and a reverse proxy server for HTTP, SMTP, POP3 and IMAP protocols, <snip>"
				},
				{
				  "name": "io.k8s.display-name",
				  "value": "Nginx 1.20"
				},
				{
				  "name": "io.openshift.expose-services",
				  "value": "8443:https"
				},
				{
				  "name": "io.openshift.s2i.scripts-url",
				  "value": "image:///usr/libexec/s2i"
				},
				{
				  "name": "io.openshift.tags",
				  "value": "builder,nginx,nginx-120"
				},
				{
				  "name": "io.s2i.scripts-url",
				  "value": "image:///usr/libexec/s2i"
				},
				{
				  "name": "maintainer",
				  "value": "SoftwareCollections.org <sclorg@redhat.com>"
				},
				{
				  "name": "name",
				  "value": "ubi8/nginx-120"
				},
				{
				  "name": "release",
				  "value": "7"
				},
				{
				  "name": "summary",
				  "value": "Platform for running nginx 1.20 or building nginx-based application"
				},
				{
				  "name": "url",
				  "value": "https://access.redhat.com/containers/#/registry.access.redhat.com/ubi8/nginx-120/images/1-7"
				},
				{
				  "name": "usage",
				  "value": "s2i build <SOURCE-REPOSITORY> ubi8/nginx-120:latest <APP-NAME>"
				},
				{
				  "name": "vcs-ref",
				  "value": "ee2f1c913a5a96f9680f7414a932a7c79558cbaa"
				},
				{
				  "name": "vcs-type",
				  "value": "git"
				},
				{
				  "name": "vendor",
				  "value": "Red Hat, Inc."
				},
				{
				  "name": "version",
				  "value": "1"
				}
			  ],
			  "layers": [
				"sha256:26c599acaaef776aada58962ad763a181101d2f49e204763ab276e67b37d5d88",
				"sha256:0661f10c38ccb1007a5937fd652f834283d016642264a0e031028979fcfb2dbf",
				"sha256:adffa69631469a649556cee5b8456f184928818064aac82106bd08bd62e51d4e",
				"sha256:26f1167feaf74177f9054bf26ac8775a4b188f25914e23bda9574ef2a759cce4"
			  ],
			  "os": "linux",
			  "ports": "[\\\"8080/tcp\\\", \\\"8443/tcp\\\"]",
			  "size": 0,
			  "uncompressed_layer_sizes": [
				{
				  "layer_id": "sha256:ec1a38375a3346f8d00b725aaab417725bad949ee387422689c69d72ef5d940e",
				  "size_bytes": 165025139
				},
				{
				  "layer_id": "sha256:558b534f4e1baf7b63f0a54e8926e2e4ea4a582be73120fa7f6a3b86d7070328",
				  "size_bytes": 56483974
				},
				{
				  "layer_id": "sha256:3ba8c926eef966b75b9545c1c2d990d3d114a4063ab71801dcaaf53165a2b130",
				  "size_bytes": 4719
				},
				{
				  "layer_id": "sha256:352ba846236b2af884cab10c53aa37d82bba9d9fb0f8797d5af211ccf317e236",
				  "size_bytes": 215755840
				}
			  ],
			  "uncompressed_size_bytes": 437269672,
			  "user": "1001"
			},
			"repositories": [
			  {
				"_links": {
				  "image_advisory": {
					"href": "/v1/advisories/redhat/id/RHBA-2021:5260"
				  },
				  "repository": {
					"href": "/v1/repositories/registry/registry.access.redhat.com/repository/ubi8/nginx-120"
				  }
				},
				"comparison": {
				  "advisory_rpm_mapping": [
					{
					  "advisory_ids": [
						"RHBA-2021:5229"
					  ],
					  "nvra": "systemd-239-51.el8_5.2.x86_64"
					},
					{
					  "advisory_ids": [
						"RHBA-2021:5229"
					  ],
					  "nvra": "systemd-pam-239-51.el8_5.2.x86_64"
					},
					{
					  "advisory_ids": [
						"RHSA-2021:5226"
					  ],
					  "nvra": "openssl-libs-1.1.1k-5.el8_5.x86_64"
					},
					{
					  "advisory_ids": [
						"RHBA-2021:5229"
					  ],
					  "nvra": "systemd-libs-239-51.el8_5.2.x86_64"
					},
					{
					  "advisory_ids": [
						"RHSA-2021:5226"
					  ],
					  "nvra": "openssl-1.1.1k-5.el8_5.x86_64"
					}
				  ],
				  "reason": "OK",
				  "reason_text": "No error",
				  "rpms": {
					"downgrade": [],
					"new": [],
					"remove": [],
					"upgrade": [
					  "systemd-239-51.el8_5.2.x86_64",
					  "systemd-pam-239-51.el8_5.2.x86_64",
					  "openssl-libs-1.1.1k-5.el8_5.x86_64",
					  "systemd-libs-239-51.el8_5.2.x86_64",
					  "openssl-1.1.1k-5.el8_5.x86_64"
					]
				  },
				  "with_nvr": "nginx-120-container-1-5.1638356804"
				},
				"content_advisory_ids": [
				  "RHSA-2021:5226",
				  "RHBA-2021:5229"
				],
				"image_advisory_id": "RHBA-2021:5260",
				"manifest_list_digest": "sha256:53f454b7894a3f4c4afea398c881e84bf9f3a375c41b119ba86f732f6eba1f92",
				"manifest_schema2_digest": "sha256:aa34453a6417f8f76423ffd2cf874e9c4a1a5451ac872b78dc636ab54a0ebbc3",
				"published": true,
				"published_date": "2021-12-21T12:04:49+00:00",
				"push_date": "2021-12-21T11:48:39+00:00",
				"registry": "registry.access.redhat.com",
				"repository": "ubi8/nginx-120",
				"signatures": [
				  {
					"key_long_id": "199E2F91FD431D51",
					"tags": [
					  "1",
					  "1-7",
					  "latest"
					]
				  }
				],
				"tags": [
				  {
					"_links": {
					  "tag_history": {
						"href": "/v1/tag-history/registry/registry.access.redhat.com/repository/ubi8/nginx-120/tag/latest"
					  }
					},
					"added_date": "2021-12-21T12:10:34.397000+00:00",
					"name": "latest"
				  },
				  {
					"_links": {
					  "tag_history": {
						"href": "/v1/tag-history/registry/registry.access.redhat.com/repository/ubi8/nginx-120/tag/1"
					  }
					},
					"added_date": "2021-12-21T12:10:34.397000+00:00",
					"name": "1"
				  },
				  {
					"_links": {
					  "tag_history": {
						"href": "/v1/tag-history/registry/registry.access.redhat.com/repository/ubi8/nginx-120/tag/1-7"
					  }
					},
					"added_date": "2021-12-21T12:10:34.397000+00:00",
					"name": "1-7"
				  }
				]
			  },
			  {
				"_links": {
				  "image_advisory": {
					"href": "/v1/advisories/redhat/id/RHBA-2021:5260"
				  },
				  "repository": {
					"href": "/v1/repositories/registry/registry.access.redhat.com/repository/rhel8/nginx-120"
				  }
				},
				"comparison": {
				  "advisory_rpm_mapping": [
					{
					  "advisory_ids": [
						"RHBA-2021:5229"
					  ],
					  "nvra": "systemd-239-51.el8_5.2.x86_64"
					},
					{
					  "advisory_ids": [
						"RHBA-2021:5229"
					  ],
					  "nvra": "systemd-pam-239-51.el8_5.2.x86_64"
					},
					{
					  "advisory_ids": [
						"RHSA-2021:5226"
					  ],
					  "nvra": "openssl-libs-1.1.1k-5.el8_5.x86_64"
					},
					{
					  "advisory_ids": [
						"RHBA-2021:5229"
					  ],
					  "nvra": "systemd-libs-239-51.el8_5.2.x86_64"
					},
					{
					  "advisory_ids": [
						"RHSA-2021:5226"
					  ],
					  "nvra": "openssl-1.1.1k-5.el8_5.x86_64"
					}
				  ],
				  "reason": "OK",
				  "reason_text": "No error",
				  "rpms": {
					"downgrade": [],
					"new": [],
					"remove": [],
					"upgrade": [
					  "systemd-239-51.el8_5.2.x86_64",
					  "systemd-pam-239-51.el8_5.2.x86_64",
					  "openssl-libs-1.1.1k-5.el8_5.x86_64",
					  "systemd-libs-239-51.el8_5.2.x86_64",
					  "openssl-1.1.1k-5.el8_5.x86_64"
					]
				  },
				  "with_nvr": "nginx-120-container-1-5.1638356804"
				},
				"content_advisory_ids": [
				  "RHSA-2021:5226",
				  "RHBA-2021:5229"
				],
				"image_advisory_id": "RHBA-2021:5260",
				"manifest_list_digest": "sha256:53f454b7894a3f4c4afea398c881e84bf9f3a375c41b119ba86f732f6eba1f92",
				"manifest_schema2_digest": "sha256:aa34453a6417f8f76423ffd2cf874e9c4a1a5451ac872b78dc636ab54a0ebbc3",
				"published": true,
				"published_date": "2021-12-21T12:04:52+00:00",
				"push_date": "2021-12-21T11:48:39+00:00",
				"registry": "registry.access.redhat.com",
				"repository": "rhel8/nginx-120",
				"signatures": [
				  {
					"key_long_id": "199E2F91FD431D51",
					"tags": [
					  "1",
					  "1-7",
					  "latest"
					]
				  }
				],
				"tags": [
				  {
					"_links": {
					  "tag_history": {
						"href": "/v1/tag-history/registry/registry.access.redhat.com/repository/rhel8/nginx-120/tag/1"
					  }
					},
					"added_date": "2021-12-21T12:08:52.904000+00:00",
					"name": "1"
				  },
				  {
					"_links": {
					  "tag_history": {
						"href": "/v1/tag-history/registry/registry.access.redhat.com/repository/rhel8/nginx-120/tag/latest"
					  }
					},
					"added_date": "2021-12-21T12:08:52.904000+00:00",
					"name": "latest"
				  },
				  {
					"_links": {
					  "tag_history": {
						"href": "/v1/tag-history/registry/registry.access.redhat.com/repository/rhel8/nginx-120/tag/1-7"
					  }
					},
					"added_date": "2021-12-21T12:08:52.904000+00:00",
					"name": "1-7"
				  }
				]
			  }
			],
			"sum_layer_size_bytes": 162814999,
			"top_layer_id": "sha256:26c599acaaef776aada58962ad763a181101d2f49e204763ab276e67b37d5d88",
			"uncompressed_top_layer_id": "sha256:ec1a38375a3346f8d00b725aaab417725bad949ee387422689c69d72ef5d940e"
		  }
		],
		"page": 0,
		"page_size": 100,
		"total": 1
	  }`
	jsonResponseNotFound = `{
		"data": [],
		"page": 0,
		"page_size": 1,
		"total": 0
	  }
	  `
)

var (
	client = api.CertAPIClient{}
	// GetDoFunc fetches the mock client's `Do` func
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

// MockClient is the mock client
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

//nolint:gochecknoinits
func init() {
	client.Client = &MockClient{}
}

var (
	containerTestCases = []struct {
		repository     string
		name           string
		id             string
		expectedError  error
		expectedResult *api.ContainerCatalogEntry
		responseData   string
		responseStatus int
	}{
		{repository: repository, name: imageName, expectedError: nil, id: "",
			expectedResult: &api.ContainerCatalogEntry{ID: "61ba0db5d095e30ed5db6330",
				FreshnessGrades: []api.ContainerImageFreshnessGrade{{Grade: "A"}}},
			responseData: jsonResponseFound, responseStatus: http.StatusAccepted},
		{repository: unKnownRepository, name: unKnownImageName, expectedError: nil, id: "", expectedResult: nil,
			responseData: jsonResponseNotFound, responseStatus: http.StatusAccepted},
	}

	operatorTestCases = []struct {
		packageName    string
		org            string
		id             string
		expectedError  error
		expectedResult bool
		responseData   string
		responseStatus int
		version        string
	}{
		{packageName: packageName, org: redHatOrg, expectedError: nil, id: "", expectedResult: true,
			responseData: jsonResponseFound, responseStatus: http.StatusAccepted, version: version},
		{packageName: unknownPackageName, org: marketPlaceOrg, expectedError: nil, id: "", expectedResult: false,
			responseData: jsonResponseNotFound, responseStatus: http.StatusNotFound, version: version},
	}

	containerQueryURLTestCases = []struct {
		id  configsections.ContainerImageIdentifier
		url string
	}{
		{id: configsections.ContainerImageIdentifier{Repository: "rhel8", Name: "nginx-120", Tag: "1-7"},
			url: "https://catalog.redhat.com/api/containers/v1/repositories/registry/registry.access.redhat.com/repository/rhel8/nginx-120/" +
				"images?filter=architecture==amd64;repositories.repository==rhel8/nginx-120;repositories.tags.name==1-7"},
		{id: configsections.ContainerImageIdentifier{Repository: "rhel8", Name: "nginx-120", Digest: "sha256:aa34453a6417f8f76423ffd2cf874e9c4a1a5451ac872b78dc636ab54a0ebbc3"},
			url: "https://catalog.redhat.com/api/containers/v1/repositories/registry/registry.access.redhat.com/repository/rhel8/nginx-120/" +
				"images?filter=architecture==amd64;image_id==sha256:aa34453a6417f8f76423ffd2cf874e9c4a1a5451ac872b78dc636ab54a0ebbc3"},
		{id: configsections.ContainerImageIdentifier{Repository: "rhel8", Name: "nginx-120"},
			url: "https://catalog.redhat.com/api/containers/v1/repositories/registry/registry.access.redhat.com/repository/rhel8/nginx-120/" +
				"images?filter=architecture==amd64;repositories.repository==rhel8/nginx-120;repositories.tags.name==latest"},
	}
)

// Do is the mock client's `Do` func
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

func getDoFunc(data string, status int) func(req *http.Request) (*http.Response, error) {
	response := io.NopCloser(bytes.NewReader([]byte(data)))
	defer response.Close()
	return func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Body:       response,
		}, nil
	}
}
func TestApiClient_GetContainerCatalogEntry(t *testing.T) {
	for _, c := range containerTestCases {
		GetDoFunc = getDoFunc(c.responseData, c.responseStatus) //nolint:bodyclose
		result, err := client.GetContainerCatalogEntry(configsections.ContainerImageIdentifier{Repository: c.repository, Name: c.name})
		assert.Equal(t, c.expectedResult, result)
		assert.Equal(t, c.expectedError, err)
	}
}

func TestApiClient_IsOperatorCertified(t *testing.T) {
	for _, c := range operatorTestCases {
		GetDoFunc = getDoFunc(c.responseData, c.responseStatus) //nolint:bodyclose
		result, err := client.IsOperatorCertified(c.org, c.packageName, c.version)
		assert.Equal(t, c.expectedResult, result)
		assert.Equal(t, c.expectedError, err)
	}
}

func TestCreateContainerCatalogQueryURL(t *testing.T) {
	for _, c := range containerQueryURLTestCases {
		url := api.CreateContainerCatalogQueryURL(c.id)
		assert.Equal(t, url, c.url)
	}
}

func TestApiClient_GetImageById(t *testing.T) {
	containerTestCases[0].id = id
	for _, c := range containerTestCases {
		GetDoFunc = getDoFunc(c.responseData, c.responseStatus) //nolint:bodyclose
		result, err := client.GetImageByID(c.id)
		assert.Equal(t, c.expectedError, err)
		if err == nil {
			assert.True(t, len(result) > 0)
		}
	}
}

func TestCertApiClient_Find(t *testing.T) {
	testData := []struct {
		data map[string]interface{}
	}{
		{data: map[string]interface{}{"_id": id}},
		{data: map[string]interface{}{"index": "index.html",
			"specs": map[string]interface{}{
				"_id":  id,
				"edit": "edit.html",
			}}},
	}
	for _, c := range testData {
		val, found := client.Find(c.data, "_id")
		assert.True(t, found)
		assert.Equal(t, id, fmt.Sprintf("%v", val))
	}
}
