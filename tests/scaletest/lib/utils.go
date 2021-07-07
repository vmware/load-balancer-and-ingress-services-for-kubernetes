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
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/avinetworks/sdk/go/clients"
	"github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
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
			session.SetPassword(password), session.SetControllerStatusCheckLimits(25, 15), session.SetInsecure)
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
	result, err := AviClient.AviSession.GetCollectionRaw(uri)
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
	result, err := AviClient.AviSession.GetCollectionRaw(uri)
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

	result, err := AviClient.AviSession.GetCollectionRaw(uri)
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
	result, err := AviClient.AviSession.GetCollectionRaw(uri)
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
	result, err := AviClient.AviSession.GetCollectionRaw(uri)
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
		uri := "/api/virtualservice-inventory??page=" + strconv.Itoa(page_num)
		result, err = AviClient.AviSession.GetCollectionRaw(uri)
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
