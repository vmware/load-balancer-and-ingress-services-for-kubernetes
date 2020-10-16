package lib

import (
	"os"
	"sync"
	"strconv"
	"encoding/json"
	"testing"
	"github.com/avinetworks/sdk/go/clients"
    "github.com/avinetworks/sdk/go/models"
    "github.com/avinetworks/sdk/go/session"
)

type AviRestClientPool struct {
	AviClient []*clients.AviClient
}

var AviClientInstance *AviRestClientPool
var clientonce sync.Once

func SharedAVIClients(t *testing.T, numClients uint32) *AviRestClientPool {
    ctrlUsername := os.Getenv("CTRL_USERNAME")
	ctrlPassword := os.Getenv("CTRL_PASSWORD")
	ctrlIpAddress := os.Getenv("CTRL_IPADDRESS")
	if ctrlUsername == "" || ctrlPassword == "" || ctrlIpAddress == "" {
		t.Logf(`AVI controller information missing. Update them in kubernetes secret or via environment variables.`)
	}
	clientonce.Do(func() {
		AviClientInstance, _ = NewAviRestClientPool(t, numClients, ctrlIpAddress, ctrlUsername, ctrlPassword)
	})
	return AviClientInstance
}

func NewAviRestClientPool(t *testing.T, num uint32, api_ep string, username string,
	password string) (*AviRestClientPool, error) {
	var p AviRestClientPool
	for i := uint32(0); i < num; i++ {
		t.Log("Error")
		aviClient, err := clients.NewAviClient(api_ep, username,
			session.SetPassword(password), session.SetControllerStatusCheckLimits(50, 10), session.SetInsecure)
		t.Log("Error")
		if err != nil {
			t.Logf("NewAviClient returned err %v", err)
			return &p, err
		}

		p.AviClient = append(p.AviClient, aviClient)
	}
	return &p, nil
}

func FetchVirtualServices(t *testing.T, AviClient *clients.AviClient) models.VirtualService{
    uri := "/api/virtualservice"
    result, err := AviClient.AviSession.GetCollectionRaw(uri)
	if err != nil {
		t.Logf("Get uri %v returned err for VS %v", uri, err)
	}
	elems := make([]json.RawMessage, result.Count)
    err = json.Unmarshal(result.Results, &elems)
    if err != nil {
		t.Logf("Failed to unmarshal VS data, err: %v", err)
	}
	vs := models.VirtualService{}
	for i := 0; i < len(elems); i++ {
        err = json.Unmarshal(elems[i], &vs)
        if err != nil {
			t.Logf("Failed to unmarshal VS data, err: %v", err)
        }
	}
	return vs
}

func FetchPoolGroup(t *testing.T, AviClient *clients.AviClient) models.PoolGroup{
	uri := "/api/poolgroup"
	result, err := AviClient.AviSession.GetCollectionRaw(uri)
	if err != nil {
		t.Logf("Get uri %v returned err for pg %v", uri, err)
	}
	elems := make([]json.RawMessage, result.Count)
    err = json.Unmarshal(result.Results, &elems)
    if err != nil {
		t.Logf("Failed to unmarshal pg data, err: %v", err)
	}
	pg := models.PoolGroup{}
	for i := 0; i < len(elems); i++ {
        err = json.Unmarshal(elems[i], &pg)
        if err != nil {
			t.Logf("Failed to unmarshal pg data, err: %v", err)
        }
	}
	return pg
}

func FetchPools(t *testing.T, AviClient *clients.AviClient) []models.Pool{
	pools := []models.Pool{}
	uri := "/api/pool?page=1"
	page_num := 1
	result, err := AviClient.AviSession.GetCollectionRaw(uri)
	if err != nil {
		t.Logf("Get uri %v returned err for pool %v", uri, err)
	}
	elems := make([]json.RawMessage, result.Count)
    err = json.Unmarshal(result.Results, &elems)
    if err != nil {
		t.Logf("Failed to unmarshal pool data, err: %v", err)
	}
	for i := 0; i < len(elems); i++ {
		pool := models.Pool{}
        err = json.Unmarshal(elems[i], &pool)
        if err != nil {
			t.Logf("Failed to unmarshal pool data, err: %v", err)
		}
		pools = append(pools, pool)
	}
	for result.Next != "" {
		page_num = page_num + 1
		uri = "/api/pool?page="+strconv.Itoa(page_num)
		result, err = AviClient.AviSession.GetCollectionRaw(uri)
		if err != nil {
			t.Logf("Get uri %v returned err for pool %v", uri, err)
		}
		elems = make([]json.RawMessage, result.Count)
		err = json.Unmarshal(result.Results, &elems)
		if err != nil {
			t.Logf("Failed to unmarshal pool data, err: %v", err)
		}
		
		for i := 0; i < len(elems); i++ {
			pool := models.Pool{}
			err = json.Unmarshal(elems[i], &pool)
			if err != nil {
				t.Logf("Failed to unmarshal pool data, err: %v", err)
			}
			pools = append(pools, pool)
		}
	}
	return pools
}