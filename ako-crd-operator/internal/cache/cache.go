package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	var result avisession.AviCollectionResult
	var err error
	for _, baseURL := range urls {
		currentURL := baseURL
		pageCount := 1
		for currentURL != "" {
			dataList := []map[string]interface{}{}
			// TODO: use ako-crd-operator session object interface instead directly accessing
			if pageCount == 1 {
				result, err = c.session.AviSessionGetCollectionRaw(currentURL, params, avisession.SetOptTenant(lib.GetQueryTenant()))
			} else {
				// result.Next will have all required params
				// parse the result.Next
				// base uri: /api/healthmonitor, /api/applicationprofile
				next_uri := strings.Split(currentURL, baseURL)
				utils.AviLog.Debugf("Found next page,  uri: %s", next_uri)
				if len(next_uri) == 1 {
					return fmt.Errorf("error while parsing next uri: [%s]", currentURL)
				}
				// next_uri[1] should contain query parameters
				// so now url should e.g. /api/healthmonitor?page=2&pag_size=100
				currentURL = baseURL + next_uri[1]
				result, err = c.session.AviSessionGetCollectionRaw(currentURL, avisession.SetOptTenant(lib.GetQueryTenant()))
			}
			if err != nil {
				return err
			}
			if err := json.Unmarshal(result.Results, &dataList); err != nil {
				return err
			}
			objectCount := 0
			for _, data := range dataList {
				UUID, ok := data["uuid"].(string)
				if !ok {
					log.Warnf("unable to find uuid in object :[%v]", data)
					continue
				}
				log.Debugf("populating cache for url: [%s] with uuid: [%s]", currentURL, UUID)
				c.dataStore.Store(UUID, data)
				objectCount++
			}
			log.Infof("cached %v objects for url: %s", objectCount, currentURL)
			currentURL = result.Next
			pageCount++
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
