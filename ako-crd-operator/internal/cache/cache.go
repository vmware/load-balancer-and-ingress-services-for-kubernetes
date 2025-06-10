package cache

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	avisession "github.com/vmware/alb-sdk/go/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type cache struct {
	dataStore   sync.Map
	session     *session.Session
	clusterName string
}

type CacheOperation interface {
	PopulateCache(context.Context, ...string) error
	GetObjectByUUID(context.Context, string) (dataMap, bool)
}

func NewCache(session *session.Session, clusterName string) CacheOperation {
	return &cache{
		dataStore:   sync.Map{},
		session:     session,
		clusterName: clusterName,
	}
}

type dataMap map[string]interface{}

func (d dataMap) GetLastModifiedTimeStamp() time.Time {
	timestamp, ok := d["_last_modified"]
	if !ok {
		return time.Unix(0, 0)
	}
	timeInt, _ := strconv.ParseInt(timestamp.(string), 10, 64)
	return time.UnixMicro(timeInt).UTC()
}

func (c *cache) PopulateCache(ctx context.Context, urls ...string) error {
	log := utils.LoggerFromContext(ctx)
	setTenant := avisession.SetTenant("*")
	aviSession := c.session.GetAviClients().AviClient[0].AviSession
	_ = setTenant(aviSession)
	// TODO: get from env variable
	setDefaultTenant := avisession.SetTenant("admin")
	defer setDefaultTenant(aviSession)

	params := avisession.SetParams(map[string]string{
		"fields":      "_last_modified,uuid",
		"page_size":   "100",
		"label_key":   "clustername",
		"label_value": c.clusterName,
	},
	)
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
					log.Warnf("unable to find uuid in object :[%v]", data)
					continue
				}
				c.dataStore.Store(UUID, data)
			}
		}
	}
	log.Infof("populated cache successfully for urls: [%v]", urls)
	return nil
}

func (c *cache) GetObjectByUUID(ctx context.Context, UUID string) (dataMap, bool) {
	log := utils.LoggerFromContext(ctx)
	data, ok := c.dataStore.Load(UUID)
	if !ok {
		log.Warnf("UUID: [%s] not found in cache", UUID)
		return nil, false
	}
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		log.Warnf("dataMap not type of map interface. type: %T", data)
		return nil, false
	}
	return dataMap, true
}
