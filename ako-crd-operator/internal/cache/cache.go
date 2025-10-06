package cache

import (
	"context"
	"encoding/json"
	"sync"

	avisession "github.com/vmware/alb-sdk/go/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/types"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

//go:generate mockgen -source=cache.go -destination=../../test/mock/cache_mock.go -package=mock
type cache struct {
	dataStore   sync.Map
	session     session.AviClientInterface
	clusterName string
}

type CacheOperation interface {
	PopulateCache(context.Context, ...string) error
	GetObjectByUUID(context.Context, string) (types.DataMap, bool)
}

func NewCache(session session.AviClientInterface, clusterName string) CacheOperation {
	return &cache{
		dataStore:   sync.Map{},
		session:     session,
		clusterName: clusterName,
	}
}

func (c *cache) PopulateCache(ctx context.Context, urls ...string) error {
	log := utils.LoggerFromContext(ctx)
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
			result, err := c.session.AviSessionGetCollectionRaw(url, params, avisession.SetOptTenant(lib.GetQueryTenant()))
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
				log.Infof("populating cache for url: [%s] with uuid: [%s]", url, UUID)
				c.dataStore.Store(UUID, data)
			}
		}
	}
	log.Infof("populated cache successfully for urls: [%v]", urls)
	return nil
}

func (c *cache) GetObjectByUUID(ctx context.Context, UUID string) (types.DataMap, bool) {
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
