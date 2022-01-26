package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/test-network-function/test-network-function/pkg/config/configsections"
)

// Endpoints document can be found here
// https://docs.engineering.redhat.com/pages/viewpage.action?spaceKey=EXD&title=Pyxis
// There are external and internal endpoints. External doesn't need authentication
// Here we are using only External endpoint to collect published containers and operator information

const apiContainerCatalogExternalBaseEndPoint = "https://catalog.redhat.com/api/containers/v1"
const apiOperatorCatalogExternalBaseEndPoint = "https://catalog.redhat.com/api/containers/v1/operators"
const apiCatalogByRepositoriesBaseEndPoint = "https://catalog.redhat.com/api/containers/v1/repositories/registry/registry.access.redhat.com/repository"

var (
	dataKey = "data"
	idKey   = "_id"
)

// HTTPClient Client interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// CertAPIClient is http client to handle `pyxis` rest api
type CertAPIClient struct {
	Client HTTPClient
}

// NewHTTPClient return new http client
func NewHTTPClient() CertAPIClient {
	return CertAPIClient{Client: &http.Client{}}
}

type catalogQueryResponse struct {
	Page     uint          `json:"page"`
	PageSize uint          `json:"page_size"`
	Total    uint          `json:"total"`
	Data     []interface{} `json:"data"`
}

// IsContainerCertified get container image info by repo/name and checks if container details is present
// If present then returns `true` as certified operators.
func (api CertAPIClient) IsContainerCertified(id configsections.ContainerImageIdentifier) (bool, error) {
	responseData, err := api.getRequest(CreateContainerCatalogQueryURL(id))
	if err == nil {
		var response catalogQueryResponse
		err = json.Unmarshal(responseData, &response)
		if err == nil {
			return len(response.Data) > 0, nil
		}
	}
	return false, err
}

func CreateContainerCatalogQueryURL(id configsections.ContainerImageIdentifier) string {
	var url string
	const defaultTag = "latest"
	if id.Digest == "" {
		if id.Tag == "" {
			id.Tag = defaultTag
		}
		url = fmt.Sprintf("%s/%s/%s/images?filter=repositories.repository==%s/%s;repositories.tags.name==%s", apiCatalogByRepositoriesBaseEndPoint, id.Repository, id.Name, id.Repository, id.Name, id.Tag)
	} else {
		url = fmt.Sprintf("%s/%s/%s/images?filter=image_id==%s", apiCatalogByRepositoriesBaseEndPoint, id.Repository, id.Name, id.Digest)
	}
	return url
}

// IsOperatorCertified get operator bundle by package name and check if package details is present
// If present then returns `true` as certified operators.
func (api CertAPIClient) IsOperatorCertified(org, packageName, version string) bool {
	if imageID, err := api.GetOperatorBundleIDByPackageName(org, packageName, version); err != nil || imageID == "" {
		return false
	}
	return false, err
}

// GetImageByID get container image data for the given container Id.  Returns (response, error).
func (api CertAPIClient) GetImageByID(id string) (string, error) {
	var response string
	url := fmt.Sprintf("%s/images/id/%s", apiContainerCatalogExternalBaseEndPoint, id)
	responseData, err := api.getRequest(url)
	if err == nil {
		response = string(responseData)
	}
	return response, err
}

// GetOperatorBundleIDByPackageName get published operator bundle Id by organization and package name.
// Returns (ImageID, error).
func (api CertAPIClient) GetOperatorBundleIDByPackageName(org, name, vsersion string) (string, error) {
	var imageID string
	url := fmt.Sprintf("%s/bundles?page_size=1&filter=organization==%s;csv_name==%s;ocp_version==%s", apiOperatorCatalogExternalBaseEndPoint, org, name, vsersion)
	responseData, err := api.getRequest(url)
	if err == nil {
		imageID, err = api.getIDFromResponse(responseData)
	}
	return imageID, err
}

// getRequest a http call to rest api, returns byte array or error. Returns (response, error).
func (api CertAPIClient) getRequest(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody) //nolint:noctx
	if err != nil {
		return nil, err
	}
	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// getIDFromResponse searches for first occurrence of id and return. Returns (id and error).
func (api CertAPIClient) getIDFromResponse(response []byte) (string, error) {
	var data interface{}
	var id string
	if err := json.Unmarshal(response, &data); err != nil {
		return id, fmt.Errorf("error unmarshalling payload in API Response %v", err.Error())
	}
	m := data.(map[string]interface{})
	for k, v := range m {
		if k == dataKey {
			// if the value is an array, search recursively
			// from each element
			if va, ok := v.([]interface{}); ok {
				for _, a := range va {
					if res, ok := api.Find(a, idKey); ok {
						id = fmt.Sprintf("%v", res)
						break
					}
				}
			}
		}
	}

	return id, nil
}

// Find key in interface (recursively) and return value as interface
func (api CertAPIClient) Find(obj interface{}, key string) (interface{}, bool) {
	// if the argument is not a map, ignore it
	mobj, ok := obj.(map[string]interface{})
	if !ok {
		return nil, false
	}
	for k, v := range mobj {
		// key match, return value
		if k == key {
			return v, true
		}
		// if the value is a map, search recursively
		if m, ok := v.(map[string]interface{}); ok {
			if res, ok := api.Find(m, key); ok {
				return res, true
			}
		}
		// if the value is an array, search recursively
		// from each element
		if va, ok := v.([]interface{}); ok {
			for _, a := range va {
				if res, ok := api.Find(a, key); ok {
					return res, true
				}
			}
		}
	}
	// element not found
	return nil, false
}
