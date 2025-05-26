package cache

import (
	"encoding/json"
	avisession "github.com/vmware/alb-sdk/go/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"strconv"
	"sync"
	"time"
)

type cache struct {
	dataStore sync.Map
	session   *session.Session
}

type CacheOperation interface {
	PopulateCache(...string) error
	GetObjectByUUID(string) (dataMap, bool)
}

func NewCache(session *session.Session) CacheOperation {
	return &cache{
		dataStore: sync.Map{},
		session:   session}
}

type dataMap map[string]interface{}

func (d dataMap) GetLastModifiedTimeStamp() time.Time {
	timestamp, ok := d["_last_modified"]
	if !ok {
		return time.Unix(0, 0)
	}
	timeInt, _ := strconv.ParseInt(timestamp.(string), 10, 64)
	return time.UnixMicro(timeInt)
}

func (c *cache) PopulateCache(urls ...string) error {
	setTenant := avisession.SetTenant("*")
	aviSession := c.session.GetAviClients().AviClient[0].AviSession
	_ = setTenant(aviSession)
	// TODO: get from env variable
	setDefaultTenant := avisession.SetTenant("admin")
	defer setDefaultTenant(aviSession)

	params := avisession.SetParams(map[string]string{"fields": "_last_modified,uuid", "page_size": "100"})
	for _, url := range urls {
		for url != "" {
			dataList := []map[string]interface{}{}
			// TODO: use ako-crd-operator session object interface instead directly accessing
			result, err := c.session.GetAviClients().AviClient[0].AviSession.GetCollectionRaw(url, params)
			url = result.Next
			if err != nil {
				return err
			}
			if err := json.Unmarshal(result.Results, &dataList); err != nil {
				return err
			}
			for _, data := range dataList {
				UUID, ok := data["uuid"].(string)
				if !ok {
					utils.AviLog.Warnf("unable to find uuid in object :[%v]", data)
				}
				c.dataStore.Store(UUID, data)
			}
		}
	}
	utils.AviLog.Infof("populated cache successfully for urls: [%v]", urls)
	return nil
}

func (c *cache) GetObjectByUUID(UUID string) (dataMap, bool) {
	data, ok := c.dataStore.Load(UUID)
	if !ok {
		utils.AviLog.Warnf("UUID: [%s] not found in cache", UUID)
		return nil, false
	}
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		utils.AviLog.Warnf("dataMap not type of map interface. type: %T", data)
		return nil, false
	}
	return dataMap, true
}
