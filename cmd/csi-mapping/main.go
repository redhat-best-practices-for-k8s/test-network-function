package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

type catalog struct {
	Page     uint `json:"page"`
	PageSize uint `json:"page_size"`
	Total    uint `json:"total"`
	Data     []struct {
		CsvName    string `json:"csv_name"`
		OcpVersion string `json:"ocp_version"`
	} `json:"data"`
	NodeName string `json:"nodeName"`
}
type Key struct {
	OperatorName, OcpVersion string
}

var (
	filterCsi          = "&filter=csv_description=iregex=CSI;organization==certified-operators"
	driverNamesCommand = "./get-driver-names.sh"
)

func getHttpBody(url string) []uint8 {
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("Http request failed with error:%s", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatalf("Error reading body: %v", err)
	}
	return body
}

func getCatalogPage(url string, page uint, filter string) catalog {
	body := getHttpBody(fmt.Sprintf("%spage=%d%s", url, page, filter))
	var aCatalog catalog
	err := json.Unmarshal(body, &aCatalog)
	if err != nil {
		log.Fatalf("Error in unmarshaling body: %v", err)
	}
	return aCatalog
}
func removeDuplicateValues(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
func ListOperator(mapOperators map[Key][]string) []string {
	keys := make(map[string]bool)
	for key, _ := range mapOperators {
		keys[key.OperatorName] = true
	}
	var list []string
	for key, _ := range keys {
		list = append(list, key)
	}
	return list
}

func getDriverNames() map[string]string {
	out, err := exec.Command(driverNamesCommand).Output()
	if err != nil {
		log.Errorf("Command execution failed err:%v", err)
	}

	driverMap := make(map[string]string)

	for _, line := range strings.Split(strings.TrimSuffix(string(out), "\n"), "\n") {
		section := strings.Split(line, ",")
		driverMap[section[1]] = section[0]
	}
	return driverMap
}

func createMapping(driverMap map[string]string, operatorList []string) map[string]string {

	driverOperatorMapping := make(map[string]string)
	for _, operator := range operatorList {
		cleanedName := strings.ReplaceAll(operator, "-csi", "")
		cleanedName = strings.ReplaceAll(cleanedName, "-certified", "")
		cleanedName = strings.ReplaceAll(cleanedName, "-operator", "")
		cleanedName = strings.ReplaceAll(cleanedName, "-stable", "")
		cleanedName = strings.ReplaceAll(cleanedName, "k8s-", "")
		cleanedName = strings.ReplaceAll(cleanedName, "bundle-", "")
		cleanedName = strings.ReplaceAll(cleanedName, "-bundle", "")
		cleanedName = strings.ReplaceAll(cleanedName, "csioperator", "")
		cleanedName = strings.ReplaceAll(cleanedName, "-cluster", "")
		for key, driverMeta := range driverMap {
			matched, err := regexp.Match(cleanedName, []byte(driverMeta))
			if err == nil && matched {
				driverOperatorMapping[key] = operator
			}
		}

	}
	return driverOperatorMapping
}

func main() {

	var fullCatalog catalog
	firstPageCatalog := getCatalogPage("https://catalog.redhat.com/api/containers/v1/operators/bundles?", 0, filterCsi)
	totalPages := firstPageCatalog.Total / firstPageCatalog.PageSize
	for i := uint(0); i < totalPages+1; i++ {
		aCatalog := getCatalogPage("https://catalog.redhat.com/api/containers/v1/operators/bundles?", i, filterCsi)
		fullCatalog.Data = append(fullCatalog.Data, aCatalog.Data...)
		log.Debug(i)
	}

	mapOperators := make(map[Key][]string)
	for _, entry := range fullCatalog.Data {
		operatorName := strings.Split(entry.CsvName, ".")[0]
		version := strings.Split(entry.CsvName, operatorName+".")[1]
		aKey := Key{OperatorName: operatorName, OcpVersion: entry.OcpVersion}
		aList := mapOperators[aKey]
		aList = append(aList, version)
		mapOperators[aKey] = aList
	}
	for key, operator := range mapOperators {
		mapOperators[key] = removeDuplicateValues(operator)
	}

	log.Infof("%+v\n", mapOperators)

	aList := ListOperator(mapOperators)
	for _, str := range aList {
		log.Infof("%s\n", str)
	}

	aDriverMap := getDriverNames()
	aMapping := createMapping(aDriverMap, aList)
	fmt.Println(aMapping)
	out, err := json.MarshalIndent(aMapping,""," ")
	if err == nil {
		err = ioutil.WriteFile("csi-mapping.json", out, 0644)
		if err != nil {
			log.Errorf("%s", err)
		}
	}
}
