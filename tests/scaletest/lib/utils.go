/*
 * Copyright 2019-2020 VMware, Inc.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package lib

import (
	"encoding/json"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

const (
	OPER_DOWN = "OPER_DOWN"
	OPER_UP   = "OPER_UP"
)

type AviRestClientPool struct {
	AviClient []*clients.AviClient
}

var AviClientInstance *AviRestClientPool
var clientonce sync.Once

func NewError(text string) error {
	return &errorString{text}
}

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

func GetUriEncoded(uri string) string {
	if uriSplit := strings.SplitN(uri, "?", 2); len(uriSplit) == 2 {
		return uriSplit[0] + "?" + url.QueryEscape(uriSplit[1])
	}
	return uri
}

func SharedAVIClients(numClients uint32) ([]*clients.AviClient, error) {
	ctrlUsername := os.Getenv("CTRL_USERNAME")
	ctrlPassword := os.Getenv("CTRL_PASSWORD")
	ctrlIpAddress := os.Getenv("CTRL_IPADDRESS")
	err := NewError("")
	if ctrlUsername == "" || ctrlPassword == "" || ctrlIpAddress == "" {
		err = NewError("AVI controller information missing.")
		return AviClientInstance.AviClient, err
	}
	clientonce.Do(func() {
		AviClientInstance, err = NewAviRestClientPool(numClients, ctrlIpAddress, ctrlUsername, ctrlPassword)
		if err != nil {
			err = NewError("NewAviClient returned err ")
		}
	})
	return AviClientInstance.AviClient, err
}

func NewAviRestClientPool(num uint32, api_ep string, username string,
	password string) (*AviRestClientPool, error) {
	var p AviRestClientPool
	for i := uint32(0); i < num; i++ {
		aviClient, err := clients.NewAviClient(api_ep, username,
			session.SetPassword(password), session.SetControllerStatusCheckLimits(25, 15), session.SetInsecure, session.SetTimeout(120*time.Second))
		if err != nil {
			return &p, err
		}

		p.AviClient = append(p.AviClient, aviClient)
	}
	return &p, nil
}

func FetchVirtualServices(t *testing.T, AviClient *clients.AviClient, Nextpage ...int) []models.VirtualService {
	virtualServices := []models.VirtualService{}
	var page_num int
	if len(Nextpage) == 1 {
		page_num = Nextpage[0]
	} else {
		page_num = 1
	}

	uri := "/api/virtualservice?page=" + strconv.Itoa(page_num)
	result, err := AviClient.AviSession.GetCollectionRaw(GetUriEncoded(uri))
	if err != nil {
		t.Errorf("Get uri %v returned err for VS %v", uri, err)
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		t.Errorf("Failed to unmarshal VS data, err: %v", err)
	}
	for _, elem := range elems {
		vs := models.VirtualService{}
		err = json.Unmarshal(elem, &vs)
		if err != nil {
			t.Errorf("Failed to unmarshal VS data, err: %v", err)
		}
		virtualServices = append(virtualServices, vs)
	}
	if result.Next != "" {
		virtualServices = append(virtualServices, FetchVirtualServices(t, AviClient, page_num+1)...)
	}
	return virtualServices
}

func FetchPoolGroup(t *testing.T, AviClient *clients.AviClient, Nextpage ...int) []models.PoolGroup {
	poolGroups := []models.PoolGroup{}
	var page_num int
	if len(Nextpage) == 1 {
		page_num = Nextpage[0]
	} else {
		page_num = 1
	}

	uri := "/api/poolgroup?page=" + strconv.Itoa(page_num)
	result, err := AviClient.AviSession.GetCollectionRaw(GetUriEncoded(uri))
	if err != nil {
		t.Errorf("Get uri %v returned err for pg %v", uri, err)
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		t.Errorf("Failed to unmarshal pg data, err: %v", err)
	}
	for _, elem := range elems {
		pg := models.PoolGroup{}
		err = json.Unmarshal(elem, &pg)
		if err != nil {
			t.Errorf("Failed to unmarshal pg data, err: %v", err)
		}
		poolGroups = append(poolGroups, pg)
	}
	if result.Next != "" {
		poolGroups = append(poolGroups, FetchPoolGroup(t, AviClient, page_num+1)...)
	}
	return poolGroups
}

func FetchPools(t *testing.T, AviClient *clients.AviClient, Nextpage ...int) []models.Pool {
	pools := []models.Pool{}
	var page_num int
	if len(Nextpage) == 1 {
		page_num = Nextpage[0]
	} else {
		page_num = 1
	}
	uri := "/api/pool?page=" + strconv.Itoa(page_num)

	result, err := AviClient.AviSession.GetCollectionRaw(GetUriEncoded(uri))
	if err != nil {
		t.Errorf("Get uri %v returned err for pool %v", uri, err)
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		t.Errorf("Failed to unmarshal pool data, err: %v", err)
	}
	for _, elem := range elems {
		pool := models.Pool{}
		err = json.Unmarshal(elem, &pool)
		if err != nil {
			t.Errorf("Failed to unmarshal pool data, err: %v", err)
		}
		pools = append(pools, pool)
	}
	if result.Next != "" {
		pools = append(pools, FetchPools(t, AviClient, page_num+1)...)
	}
	return pools
}

func FetchDNSARecordsFQDN(t *testing.T, AviClient *clients.AviClient, Nextpage ...int) []string {
	FQDNList := []string{}
	var page_num int
	if len(Nextpage) == 1 {
		page_num = Nextpage[0]
	} else {
		page_num = 1
	}
	uri := "/api/virtualservice?page=" + strconv.Itoa(page_num)
	result, err := AviClient.AviSession.GetCollectionRaw(GetUriEncoded(uri))
	if err != nil {
		t.Errorf("Get uri %v returned err for VS %v", uri, err)
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		t.Errorf("Failed to unmarshal VS data, err: %v", err)
	}
	for _, elem := range elems {
		vs := models.VirtualService{}
		err = json.Unmarshal(elem, &vs)
		if err != nil {
			t.Errorf("Failed to unmarshal VS data, err: %v", err)
		}
		for _, dnsInfo := range vs.DNSInfo {
			FQDNList = append(FQDNList, *dnsInfo.Fqdn)
		}
	}
	if result.Next != "" {
		FQDNList = append(FQDNList, FetchDNSARecordsFQDN(t, AviClient, page_num+1)...)
	}
	return FQDNList
}

func FetchVirtualServiceOperStatus(t *testing.T, AviClient *clients.AviClient) []VirtualServiceInventoryRuntime {
	OperStatus := []VirtualServiceInventoryRuntime{}
	uri := "/api/virtualservice-inventory?page=1"
	page_num := 1
	result, err := AviClient.AviSession.GetCollectionRaw(GetUriEncoded(uri))
	if err != nil {
		t.Errorf("Get uri %v returned err for VS %v", uri, err)
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		t.Errorf("Failed to unmarshal VS data, err: %v", err)
	}
	for _, elem := range elems {
		vs := VirtualServiceInventoryResult{}
		err = json.Unmarshal(elem, &vs)
		if err != nil {
			t.Errorf("Failed to unmarshal VS data, err: %v", err)
		}
		operstate := VirtualServiceInventoryRuntime{
			Name:  vs.Config.Name,
			UUID:  vs.Config.UUID,
			State: vs.Runtime.OperStatus.State,
		}
		OperStatus = append(OperStatus, operstate)
	}
	for result.Next != "" {
		page_num = page_num + 1
		uri := "/api/virtualservice-inventory?page=" + strconv.Itoa(page_num)
		result, err = AviClient.AviSession.GetCollectionRaw(GetUriEncoded(uri))
		if err != nil {
			t.Errorf("Get uri %v returned err for VS %v", uri, err)
		}
		elems := make([]json.RawMessage, result.Count)
		err = json.Unmarshal(result.Results, &elems)
		if err != nil {
			t.Errorf("Failed to unmarshal VS data, err: %v", err)
		}
		for _, elem := range elems {
			vs := VirtualServiceInventoryResult{}
			err = json.Unmarshal(elem, &vs)
			if err != nil {
				t.Errorf("Failed to unmarshal VS data, err: %v", err)
			}
			operstate := VirtualServiceInventoryRuntime{
				Name:  vs.Config.Name,
				UUID:  vs.Config.UUID,
				State: vs.Runtime.OperStatus.State,
			}
			OperStatus = append(OperStatus, operstate)
		}
	}
	return OperStatus
}

func FetchOPERDownVirtualService(t *testing.T, AviClient *clients.AviClient) []VirtualServiceInventoryRuntime {
	OperDownVS := []VirtualServiceInventoryRuntime{}
	VSOperStatus := FetchVirtualServiceOperStatus(t, AviClient)
	for _, vs := range VSOperStatus {
		if vs.State != OPER_UP {
			OperDownVS = append(OperDownVS, vs)
		}
	}
	return OperDownVS
}

func CleanResourceData(data string) string {
	data = strings.Replace(data, "\\", "", -1)
	data = strings.Replace(data, " ", "", -1)
	return data
}

func CompareVirtualServiceResources(t *testing.T, eventLog models.EventLog) bool {
	var new, old models.VirtualService
	err := json.Unmarshal([]byte(CleanResourceData(*eventLog.EventDetails.ConfigUpdateDetails.NewResourceData)), &new)
	if err != nil {
		t.Fatalf("Error unmarshalling data into VS. Error : %v", err)
	}
	err = json.Unmarshal([]byte(CleanResourceData(*eventLog.EventDetails.ConfigUpdateDetails.OldResourceData)), &old)
	if err != nil {
		t.Fatalf("Error unmarshalling data into VS. Error : %v", err)
	}
	if *new.LastModified == *old.LastModified {
		return true
	}
	// Check if all fields other than LastModified are equal
	// Set the LastModified field of Old VS to LastModified of New VS to Mask the difference from DeepEqual
	old.LastModified = new.LastModified
	if reflect.DeepEqual(new, old) {
		// Old and New VS are same. Only the LastModified field has been updated -> Unnecessary API call by AKO
		return false
	}
	// Check if all fields other than LastModified and CloudConfigCksum are equal
	// Set the CloudConfigCksum field of Old VS to CloudConfigCksum of New VS to Mask the difference from DeepEqual
	old.CloudConfigCksum = new.CloudConfigCksum
	if reflect.DeepEqual(new, old) {
		// Old and New VS are same. Only the LastModified and CloudConfigCksum field has been updated -> Unnecessary API call by AKO
		return false
	}
	// Genuine update of VS object
	return true
}

func ComparePoolResources(t *testing.T, eventLog models.EventLog) bool {
	var new, old models.Pool
	err := json.Unmarshal([]byte(CleanResourceData(*eventLog.EventDetails.ConfigUpdateDetails.NewResourceData)), &new)
	if err != nil {
		t.Fatalf("Error unmarshalling data into Pool. Error : %v", err)
	}
	err = json.Unmarshal([]byte(CleanResourceData(*eventLog.EventDetails.ConfigUpdateDetails.OldResourceData)), &old)
	if err != nil {
		t.Fatalf("Error unmarshalling data into Pool. Error : %v", err)
	}
	if *new.LastModified == *old.LastModified {
		return true
	}
	// Check if all fields other than LastModified are equal
	// Set the LastModified field of Old Pool to LastModified of New Pool to Mask the difference from DeepEqual
	old.LastModified = new.LastModified
	if reflect.DeepEqual(new, old) {
		// Old and New Pool are same. Only the LastModified field has been updated -> Unnecessary API call by AKO
		return false
	}
	// Check if all fields other than LastModified and CloudConfigCksum are equal
	// Set the CloudConfigCksum field of Old Pool to CloudConfigCksum of New Pool to Mask the difference from DeepEqual
	old.CloudConfigCksum = new.CloudConfigCksum
	if reflect.DeepEqual(new, old) {
		// Old and New Pool are same. Only the LastModified and CloudConfigCksum field has been updated -> Unnecessary API call by AKO
		return false
	}
	// Genuine update of Pool object
	return true
}

func ComparePoolGroupResources(t *testing.T, eventLog models.EventLog) bool {
	var new, old models.PoolGroup
	err := json.Unmarshal([]byte(CleanResourceData(*eventLog.EventDetails.ConfigUpdateDetails.NewResourceData)), &new)
	if err != nil {
		t.Fatalf("Error unmarshalling data into PoolGroup. Error : %v", err)
	}
	err = json.Unmarshal([]byte(CleanResourceData(*eventLog.EventDetails.ConfigUpdateDetails.OldResourceData)), &old)
	if err != nil {
		t.Fatalf("Error unmarshalling data into PoolGroup. Error : %v", err)
	}

	if *new.LastModified == *old.LastModified {
		return true
	}
	// Check if all fields other than LastModified are equal
	// Set the LastModified field of Old PoolGroup to LastModified of New PoolGroup to Mask the difference from DeepEqual
	old.LastModified = new.LastModified
	if reflect.DeepEqual(new, old) {
		// Old and New PoolGroup are same. Only the LastModified field has been updated -> Unnecessary API call by AKO
		return false
	}
	// Check if all fields other than LastModified and CloudConfigCksum are equal
	// Set the CloudConfigCksum field of Old PoolGroup to CloudConfigCksum of New PoolGroup to Mask the difference from DeepEqual
	old.CloudConfigCksum = new.CloudConfigCksum
	if reflect.DeepEqual(new, old) {
		// Old and New PoolGroup are same. Only the LastModified and CloudConfigCksum field has been updated -> Unnecessary API call by AKO
		return false
	}
	// Genuine update of PoolGroup object
	return true
}

func CompareVsVipResources(t *testing.T, eventLog models.EventLog) bool {
	var new, old models.VsVip
	err := json.Unmarshal([]byte(CleanResourceData(*eventLog.EventDetails.ConfigUpdateDetails.NewResourceData)), &new)
	if err != nil {
		t.Fatalf("Error unmarshalling data into VsVip. Error : %v", err)
	}
	err = json.Unmarshal([]byte(CleanResourceData(*eventLog.EventDetails.ConfigUpdateDetails.OldResourceData)), &old)
	if err != nil {
		t.Fatalf("Error unmarshalling data into VsVip. Error : %v", err)
	}
	if *new.LastModified == *old.LastModified {
		return true
	}
	// Check if all fields other than LastModified are equal
	// Set the LastModified field of Old VsVip to LastModified of New VsVip to Mask the difference from DeepEqual
	old.LastModified = new.LastModified
	if reflect.DeepEqual(new, old) {
		// Old and New VsVip are same. Only the LastModified field has been updated -> Unnecessary API call by AKO
		return false
	}
	// Genuine update of VsVip object
	return true
}

func CompareHTTPPolicySet(t *testing.T, eventLog models.EventLog) bool {
	var new, old models.HTTPPolicySet
	err := json.Unmarshal([]byte(CleanResourceData(*eventLog.EventDetails.ConfigUpdateDetails.NewResourceData)), &new)
	if err != nil {
		t.Fatalf("Error unmarshalling data into HttpPolicySet. Error : %v", err)
	}
	err = json.Unmarshal([]byte(CleanResourceData(*eventLog.EventDetails.ConfigUpdateDetails.OldResourceData)), &old)
	if err != nil {
		t.Fatalf("Error unmarshalling data into HttpPolicySet. Error : %v", err)
	}
	if *new.LastModified == *old.LastModified {
		return true
	}
	// Check if all fields other than LastModified are equal
	// Set the LastModified field of Old HttpPolicySet to LastModified of New HttpPolicySet to Mask the difference from DeepEqual
	old.LastModified = new.LastModified
	if reflect.DeepEqual(new, old) {
		// Old and New HttpPolicySet are same. Only the LastModified field has been updated -> Unnecessary API call by AKO
		return false
	}
	// Check if all fields other than LastModified and CloudConfigCksum are equal
	// Set the CloudConfigCksum field of Old HttpPolicySet to CloudConfigCksum of New HttpPolicySet to Mask the difference from DeepEqual
	old.CloudConfigCksum = new.CloudConfigCksum
	if reflect.DeepEqual(new, old) {
		// Old and New HttpPolicySet are same. Only the LastModified and CloudConfigCksum field has been updated -> Unnecessary API call by AKO
		return false
	}
	// Genuine update of HttpPolicySet object
	return true
}

func CompareSSLKeyCertificate(t *testing.T, eventLog models.EventLog) bool {
	var new, old models.SSLKeyAndCertificate
	err := json.Unmarshal([]byte(CleanResourceData(*eventLog.EventDetails.ConfigUpdateDetails.NewResourceData)), &new)
	if err != nil {
		t.Fatalf("Error unmarshalling data into SSLKeyCertificate. Error : %v", err)
	}
	err = json.Unmarshal([]byte(CleanResourceData(*eventLog.EventDetails.ConfigUpdateDetails.OldResourceData)), &old)
	if err != nil {
		t.Fatalf("Error unmarshalling data into SSLKeyCertificate. Error : %v", err)
	}
	if *new.LastModified == *old.LastModified {
		return true
	}
	// Check if all fields other than LastModified are equal
	// Set the LastModified field of Old SSLKeyCertificate to LastModified of New SSLKeyCertificate to Mask the difference from DeepEqual
	old.LastModified = new.LastModified
	if reflect.DeepEqual(new, old) {
		// Old and New SSLKeyCertificate are same. Only the LastModified field has been updated -> Unnecessary API call by AKO
		return false
	}
	// Genuine update of SSLKeyCertificate object
	return true
}

func CheckForUnwantedAPICallsToController(t *testing.T, AviClient *clients.AviClient, start string, end string, Nextpage ...int) bool {
	var page_num int
	if len(Nextpage) == 1 {
		page_num = Nextpage[0]
	} else {
		page_num = 1
	}
	uri := "/api/analytics/logs/" +
		"?type=2" +
		"&filter=ne(internal,EVENT_INTERNAL)" +
		"&filter=co(event_id,CONFIG_UPDATE)" +
		"&orderby=-report_timestamp" +
		"&start=" + start +
		"&end=" + end +
		"&page=" + strconv.Itoa(page_num)

	result, err := AviClient.AviSession.GetCollectionRaw(GetUriEncoded(uri))
	if err != nil {
		t.Errorf("Get uri %v returned err for Event log %v", uri, err)
	}
	elems := make([]json.RawMessage, result.Count)
	t.Logf("Found %d config updates between %s and %s", result.Count, start, end)

	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		t.Fatalf("Failed to unmarshal Event log data, err: %v", err)
	}
	for _, elem := range elems {
		eventLog := models.EventLog{}
		err = json.Unmarshal(elem, &eventLog)
		if err != nil {
			t.Logf("Failed to unmarshal Event log data, err: %v", err)
		}
		objectType := eventLog.ObjType

		if *objectType == "VIRTUALSERVICE" {
			if !CompareVirtualServiceResources(t, eventLog) {
				return false
			}
		} else if *objectType == "POOL" {
			if !ComparePoolResources(t, eventLog) {
				return false
			}
		} else if *objectType == "POOLGROUP" {
			if !ComparePoolGroupResources(t, eventLog) {
				return false
			}
		} else if *objectType == "VSVIP" {
			if !CompareVsVipResources(t, eventLog) {
				return false
			}
		} else if *objectType == "HTTPPOLICYSET" {
			if !CompareHTTPPolicySet(t, eventLog) {
				return false
			}
		} else if *objectType == "SSLKEYANDCERTIFICATE" {
			if !CompareSSLKeyCertificate(t, eventLog) {
				return false
			}
		}
	}
	if result.Next != "" {
		return CheckForUnwantedAPICallsToController(t, AviClient, start, end, page_num+1)
	}
	return true

}
