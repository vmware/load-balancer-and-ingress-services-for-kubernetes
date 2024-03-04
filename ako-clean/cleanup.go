package akoclean

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/avirest"
	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type aviControllerConfig struct {
	host      string
	user      string
	password  string
	authToken string
	caCert    string
}

type AKOCleanupConfig struct {
	aviControllerConfig
	clusterID string
}

func NewAKOCleanupConfig(host, user, password, authToken, caCert, clusterID string) *AKOCleanupConfig {
	return &AKOCleanupConfig{
		aviControllerConfig: aviControllerConfig{
			host:      host,
			user:      user,
			password:  password,
			authToken: authToken,
			caCert:    caCert,
		},
		clusterID: clusterID,
	}
}

func populateControllerProperties(cfg *AKOCleanupConfig) {
	ctrlProps := map[string]string{
		utils.ENV_CTRL_USERNAME:  cfg.user,
		utils.ENV_CTRL_PASSWORD:  cfg.password,
		utils.ENV_CTRL_AUTHTOKEN: cfg.authToken,
		utils.ENV_CTRL_CADATA:    cfg.caCert,
	}
	ctrlPropCache := utils.SharedCtrlProp()
	ctrlPropCache.PopulateCtrlProp(ctrlProps)
}

func waitTillDeletion(uri string, client *clients.AviClient, retry int) error {
	if retry == 0 {
		return fmt.Errorf("resource not deleted under expected time")
	}
	var response interface{}
	err := client.AviSession.Get(uri, &response)
	if err != nil {
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 404 {
			return nil
		}
	}
	time.Sleep(1 * time.Second)
	return waitTillDeletion(uri, client, retry-1)
}

func deleteAviResource(prefix string, res map[string][]string) error {
	for tenant, uuids := range res {
		aviClient := avicache.SharedAVIClients(tenant).AviClient[tenant][0]
		for _, uuid := range uuids {
			uri := prefix + "/" + uuid
			utils.AviLog.Infof("Deleting %s in %s tenant", uuid, tenant)
			err := lib.AviDelete(aviClient, uri)
			if err != nil {
				return err
			}
			waitTillDeletion(uri, aviClient, 10)
		}
	}
	return nil
}

func cleanupVirtualServices() error {
	aviObjCache := avicache.SharedAviObjCache()
	parentVsKeys := aviObjCache.VsCacheMeta.AviCacheGetAllParentVSKeys()
	parentVsKeySet := make(map[string]struct{})
	for _, key := range parentVsKeys {
		parentVsKeySet[fmt.Sprintf("%s/%s", key.Namespace, key.Name)] = struct{}{}
	}

	parentVS := make(map[string][]string)
	otherVS := make(map[string][]string)
	for _, key := range aviObjCache.VsCacheMeta.AviGetAllKeys() {
		vsCache, _ := aviObjCache.VsCacheMeta.AviCacheGet(key)
		vsUuid := vsCache.(*avicache.AviVsCache).Uuid
		if vsUuid == "" {
			continue
		}
		if _, ok := parentVS[key.Namespace]; !ok {
			parentVS[key.Namespace] = []string{}
		}
		if _, ok := otherVS[key.Namespace]; !ok {
			otherVS[key.Namespace] = []string{}
		}
		if _, ok := parentVsKeySet[fmt.Sprintf("%s/%s", key.Namespace, key.Name)]; ok {
			parentVS[key.Namespace] = append(parentVS[key.Namespace], vsUuid)
		} else {
			otherVS[key.Namespace] = append(otherVS[key.Namespace], vsUuid)
		}
	}

	err := deleteAviResource("/api/virtualservice", otherVS)
	if err != nil {
		return err
	}
	return deleteAviResource("/api/virtualservice", parentVS)
}

func cleanupVsVips() error {
	aviObjCache := avicache.SharedAviObjCache()
	vsvips := make(map[string][]string)
	for _, key := range aviObjCache.VSVIPCache.AviGetAllKeys() {
		if _, ok := vsvips[key.Namespace]; !ok {
			vsvips[key.Namespace] = []string{}
		}
		vsvipCache, _ := aviObjCache.VSVIPCache.AviCacheGet(key)
		vsvips[key.Namespace] = append(vsvips[key.Namespace], vsvipCache.(*avicache.AviVSVIPCache).Uuid)
	}
	return deleteAviResource("/api/vsvip", vsvips)
}

func cleanupVSDatascripts() error {
	aviObjCache := avicache.SharedAviObjCache()
	dscripts := make(map[string][]string)
	for _, key := range aviObjCache.DSCache.AviGetAllKeys() {
		if _, ok := dscripts[key.Namespace]; !ok {
			dscripts[key.Namespace] = []string{}
		}
		dsCache, _ := aviObjCache.DSCache.AviCacheGet(key)
		dscripts[key.Namespace] = append(dscripts[key.Namespace], dsCache.(*avicache.AviDSCache).Uuid)

	}
	return deleteAviResource("/api/vsdatascriptset", dscripts)
}

func cleanupHTTPPolicySets() error {
	aviObjCache := avicache.SharedAviObjCache()
	httpsets := make(map[string][]string)
	for _, key := range aviObjCache.HTTPPolicyCache.AviGetAllKeys() {
		if _, ok := httpsets[key.Namespace]; !ok {
			httpsets[key.Namespace] = []string{}
		}
		httpCache, _ := aviObjCache.HTTPPolicyCache.AviCacheGet(key)
		httpsets[key.Namespace] = append(httpsets[key.Namespace], httpCache.(*avicache.AviHTTPPolicyCache).Uuid)

	}
	return deleteAviResource("/api/httppolicyset", httpsets)
}

func cleanupL4PolicySets() error {
	aviObjCache := avicache.SharedAviObjCache()
	l4sets := make(map[string][]string)
	for _, key := range aviObjCache.L4PolicyCache.AviGetAllKeys() {
		if _, ok := l4sets[key.Namespace]; !ok {
			l4sets[key.Namespace] = []string{}
		}
		l4Cache, _ := aviObjCache.L4PolicyCache.AviCacheGet(key)
		l4sets[key.Namespace] = append(l4sets[key.Namespace], l4Cache.(*avicache.AviL4PolicyCache).Uuid)
	}
	return deleteAviResource("/api/l4policyset", l4sets)
}

func cleanupPoolGroups() error {
	aviObjCache := avicache.SharedAviObjCache()
	pgroups := make(map[string][]string)
	for _, key := range aviObjCache.PgCache.AviGetAllKeys() {
		if _, ok := pgroups[key.Namespace]; !ok {
			pgroups[key.Namespace] = []string{}
		}
		pgCache, _ := aviObjCache.PgCache.AviCacheGet(key)
		pgroups[key.Namespace] = append(pgroups[key.Namespace], pgCache.(*avicache.AviPGCache).Uuid)
	}
	return deleteAviResource("/api/poolgroup", pgroups)
}

func cleanupPools() error {
	aviObjCache := avicache.SharedAviObjCache()
	pools := make(map[string][]string)
	for _, key := range aviObjCache.PoolCache.AviGetAllKeys() {
		if _, ok := pools[key.Namespace]; !ok {
			pools[key.Namespace] = []string{}
		}
		poolCache, _ := aviObjCache.PoolCache.AviCacheGet(key)
		pools[key.Namespace] = append(pools[key.Namespace], poolCache.(*avicache.AviPoolCache).Uuid)
	}
	return deleteAviResource("/api/pool", pools)
}

func (cfg *AKOCleanupConfig) Cleanup() error {
	err := cfg.validate()
	if err != nil {
		return err
	}

	lib.SetClusterID(cfg.clusterID)
	lib.SetControllerIP(cfg.host)
	os.Setenv(utils.VCF_CLUSTER, "true")
	akoControlConfig := lib.AKOControlConfig()
	akoControlConfig.SetAKOInstanceFlag(true)
	lib.SetNamePrefix("")
	lib.SetAKOUser(lib.AKOPrefix)
	//utils.AviLog.SetLevel("INFO")
	akoControlConfig.SetIsLeaderFlag(true)

	populateControllerProperties(cfg)

	aviRestClientPool := avicache.SharedAVIClients(lib.GetTenant())
	if aviRestClientPool == nil {
		return fmt.Errorf("avi client not initialized")
	}

	err = setCloudName()
	if err != nil {
		return err
	}

	tenants := make(map[string]struct{})
	err = lib.GetAllTenants(aviRestClientPool.AviClient[lib.GetTenant()][0], tenants)
	if err != nil {
		return err
	}

	for tenant := range tenants {
		err := k8s.PopulateCache(tenant)
		if err != nil {
			return err
		}
	}

	err = cleanupVirtualServices()
	if err != nil {
		return err
	}

	err = cleanupVsVips()
	if err != nil {
		return err
	}

	err = cleanupVSDatascripts()
	if err != nil {
		return err
	}

	err = cleanupHTTPPolicySets()
	if err != nil {
		return err
	}

	err = cleanupL4PolicySets()
	if err != nil {
		return err
	}

	err = cleanupPoolGroups()
	if err != nil {
		return err
	}

	err = cleanupPools()
	if err != nil {
		return err
	}

	avirest.InfraAviClientInstance(aviRestClientPool.AviClient[lib.GetTenant()][0])
	err = avirest.DeleteServiceEngines()
	if err != nil {
		return err
	}

	err = avirest.DeleteServiceEngineGroup()
	if err != nil {
		return err
	}

	return nil
}

func setCloudName() error {
	uri := "/api/serviceenginegroup/?name=" + lib.GetClusterID() + "&include_name=True"
	aviRestClientPool := avicache.SharedAVIClients(lib.GetTenant())
	response := models.ServiceEngineGroupAPIResponse{}
	err := lib.AviGet(aviRestClientPool.AviClient[lib.GetTenant()][0], uri, &response)
	if err != nil {
		return err
	}
	if len(response.Results) == 0 {
		return fmt.Errorf("segroup not found")
	}
	cloudName := strings.Split(*response.Results[0].CloudRef, "#")[1]
	utils.SetCloudName(cloudName)
	return nil
}

func (cfg *AKOCleanupConfig) validate() error {
	if cfg.host == "" || cfg.user == "" {
		return fmt.Errorf("invalid config: host/user is required")
	}
	if cfg.password == "" && cfg.authToken == "" {
		return fmt.Errorf("invalid config: one of password or authtoken is required")
	}
	if cfg.clusterID == "" {
		return fmt.Errorf("invalid config: cluster id is required")
	}
	return nil
}
