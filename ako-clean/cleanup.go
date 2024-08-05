package akoclean

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/avirest"
	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/rest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var (
	SidecarProxyEndpoint = "localhost:1080"
	UseExternalCert      = "external-cert"
	ServerCertHeader     = "x-vmware-server-tls-cert"
	AviMinVersion        = "30.1.1"
	NamePrefix           = ""
	AKOuser              = ""
	AdminTenant          = "admin"
	Cloud                = ""
	SEGroupUUID          = ""
	SEGroupNotFoundError = "SEGroup does not exist"
	Referer              = "Referer"
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
	useEnvoy  bool
}

func NewAKOCleanupConfig(host, user, password, authToken, caCert, clusterID string, useEnvoy bool) *AKOCleanupConfig {
	return &AKOCleanupConfig{
		aviControllerConfig: aviControllerConfig{
			host:      host,
			user:      user,
			password:  password,
			authToken: authToken,
			caCert:    caCert,
		},
		clusterID: clusterID,
		useEnvoy:  useEnvoy,
	}
}

func (cfg *AKOCleanupConfig) Cleanup(ctx context.Context) error {
	err := cfg.validate()
	if err != nil {
		return err
	}

	lib.SetClusterID(cfg.clusterID)
	referer := "https://" + cfg.host
	if cfg.useEnvoy {
		cfg.host = fmt.Sprintf("%s/%s/http1/%s/443", SidecarProxyEndpoint, UseExternalCert, cfg.host)
	}
	lib.SetControllerIP(cfg.host)
	os.Setenv(utils.VCF_CLUSTER, "true")
	akoControlConfig := lib.AKOControlConfig()
	akoControlConfig.SetAKOInstanceFlag(true)
	lib.SetNamePrefix("")
	lib.SetAKOUser(lib.AKOPrefix)
	akoControlConfig.SetIsLeaderFlag(true)

	populateControllerProperties(cfg, referer)

	var aviRestClientPool *utils.AviRestClientPool
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		aviRestClientPool = avicache.SharedAVIClients(lib.GetTenant())
		if aviRestClientPool == nil {
			return fmt.Errorf("avi client not initialized")
		}
	}

	avirest.InfraAviClientInstance(aviRestClientPool.AviClient[0])

	ops := []func() error{
		setCloudName,
		populateCache,
		cleanupVirtualServices,
		cleanupVsVips,
		cleanupVSDatascripts,
		cleanupHTTPPolicySets,
		cleanupL4PolicySets,
		cleanupPoolGroups,
		cleanupPools,
		func() error { return avirest.DeleteServiceEngines() },
		avirest.DeleteServiceEngineGroup,
		cleanupVIPNetwork,
	}

	for _, op := range ops {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = op()
			if err != nil {
				if strings.Contains(err.Error(), SEGroupNotFoundError) {
					return nil
				}
				return err
			}
		}
	}
	return nil
}

func setCloudName() error {
	uri := "/api/serviceenginegroup/?name=" + lib.GetClusterID() + "&include_name=True"
	aviRestClientPool := avicache.SharedAVIClients(lib.GetTenant())
	response := models.ServiceEngineGroupAPIResponse{}
	err := lib.AviGet(aviRestClientPool.AviClient[0], uri, &response)
	if err != nil {
		return err
	}
	if len(response.Results) == 0 {
		return fmt.Errorf(SEGroupNotFoundError)
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

func populateControllerProperties(cfg *AKOCleanupConfig, referer string) {
	ctrlProps := map[string]string{
		utils.ENV_CTRL_USERNAME:  cfg.user,
		utils.ENV_CTRL_PASSWORD:  cfg.password,
		utils.ENV_CTRL_AUTHTOKEN: cfg.authToken,
		utils.ENV_CTRL_CADATA:    cfg.caCert,
	}
	ctrlPropCache := utils.SharedCtrlProp()
	ctrlPropCache.PopulateCtrlProp(ctrlProps)
	if cfg.useEnvoy {
		ctrlPropCache.PopulateCtrlAPIScheme("http")
		header := map[string]string{
			ServerCertHeader: convertPemToDer(cfg.caCert),
			Referer:          referer,
		}
		ctrlPropCache.PopulateCtrlAPIUserHeaders(header)
	}
}

func populateCache() error {
	adminRestClientPool := avicache.SharedAVIClients(lib.GetTenant())
	tenants := make(map[string]struct{})
	err := lib.GetAllTenants(adminRestClientPool.AviClient[0], tenants)
	if err != nil {
		return err
	}
	for tenant := range tenants {
		aviRestClientPool := avicache.SharedAVIClients(tenant)
		aviObjCache := avicache.SharedAviObjCache()
		// Randomly pickup a client.
		if aviRestClientPool != nil && len(aviRestClientPool.AviClient) > 0 {
			_, _, err = aviObjCache.AviObjCachePopulate(aviRestClientPool.AviClient, lib.AKOControlConfig().ControllerVersion(), utils.CloudName, tenant)
			if err != nil {
				utils.AviLog.Warnf("failed to populate avi cache with error: %v", err.Error())
				return err
			}

		}
	}
	if err = avicache.SetControllerClusterUUID(adminRestClientPool); err != nil {
		utils.AviLog.Warnf("Failed to set the controller cluster uuid with error: %v", err)
	}

	return nil
}

func waitTillDeletion(uri string, client *clients.AviClient, retry int) error {
	if retry == 0 {
		return fmt.Errorf("resource not deleted under expected time")
	}
	var response interface{}
	err := client.AviSession.Get(utils.GetUriEncoded(uri), &response)
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
		aviClient := avicache.SharedAVIClients(tenant).AviClient[0]
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

/*
Below section is only applicable for T1 based Supervisor deployments
*/
func cleanupVIPNetwork() error {
	err := sanitzeAviCloud()
	if err != nil {
		return err
	}

	aviClient := avicache.SharedAVIClients(lib.GetAdminTenant()).AviClient[0]
	avirest.AviNetCachePopulate(aviClient, utils.CloudName)
	if len(avirest.NetCache) == 0 {
		return nil
	}

	err = checkAndUpdateIPAM()
	if err != nil {
		return err
	}

	networks := map[string][]string{
		lib.GetAdminTenant(): {},
	}

	for _, network := range avirest.NetCache {
		networks[lib.GetAdminTenant()] = append(networks[lib.GetAdminTenant()], *network.UUID)
	}
	return deleteAviResource("/api/network", networks)
}

func sanitzeAviCloud() error {
	aviClient := avicache.SharedAVIClients(lib.GetAdminTenant()).AviClient[0]
	err := avirest.AviCloudCachePopulate(aviClient, utils.CloudName, true)
	if err != nil {
		return err
	}
	dataNetworkTier1Lrs := make([]*models.Tier1LogicalRouterInfo, 0)
	cloudTier1Lrs := avirest.CloudCache.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs
	cloudLRLSMap := make(map[string]string)
	t1Handler := avirest.T1LRNetworking{}
	err = t1Handler.GetClusterSpecificNSXTSegmentsinCloud(aviClient, cloudLRLSMap)
	if err != nil {
		return err
	}

	utils.AviLog.Infof("Number of stale entries to be remvoved: %d", len(cloudLRLSMap))
	for i := range cloudTier1Lrs {
		if _, ok := cloudLRLSMap[*cloudTier1Lrs[i].SegmentID]; ok {
			continue
		}
		dataNetworkTier1Lrs = append(dataNetworkTier1Lrs, cloudTier1Lrs[i])
	}
	avirest.CloudCache.NsxtConfiguration.DataNetworkConfig.Tier1SegmentConfig.Manual.Tier1Lrs = dataNetworkTier1Lrs
	path := "/api/cloud/" + *avirest.CloudCache.UUID
	restOp := utils.RestOp{
		ObjName: utils.CloudName,
		Path:    path,
		Method:  utils.RestPut,
		Obj:     &avirest.CloudCache,
		Tenant:  "admin",
		Model:   "cloud",
	}
	restLayer := rest.NewRestOperations(nil, true)
	return restLayer.AviRestOperateWrapper(aviClient, []*utils.RestOp{&restOp}, "aviCleanup")
}

func checkAndUpdateIPAM() error {
	aviClient := avicache.SharedAVIClients(lib.GetAdminTenant()).AviClient[0]
	if avirest.CloudCache.IPAMProviderRef == nil || *avirest.CloudCache.IPAMProviderRef == "" {
		return nil
	}
	avirest.AviIPAMCachePopulate(aviClient, strings.Split(*avirest.CloudCache.IPAMProviderRef, "#")[1])
	ipam := avirest.IPAMCache
	updateIPAM := false
	usableNetworks := make([]*models.IPAMUsableNetwork, 0)
	if ipam.InternalProfile != nil && len(ipam.InternalProfile.UsableNetworks) > 0 {
		for _, nw := range ipam.InternalProfile.UsableNetworks {
			netName := strings.Split(*nw.NwRef, "#")[1]
			if strings.HasPrefix(netName, lib.GetVCFNetworkName()) {
				utils.AviLog.Infof("Removing VIP Network: %s from the Avi IPAM profile: %s", netName, *ipam.Name)
				updateIPAM = true
				continue
			}
			networkRef := "/api/network/?name=" + netName
			usableNetworks = append(usableNetworks, &models.IPAMUsableNetwork{NwRef: &networkRef})
		}
	}
	if !updateIPAM {
		return nil
	}
	ipam.InternalProfile.UsableNetworks = usableNetworks
	path := strings.Split(*avirest.CloudCache.IPAMProviderRef, "/ipamdnsproviderprofile/")[1]
	restOp := utils.RestOp{
		Path:   "/api/ipamdnsproviderprofile/" + path,
		Method: utils.RestPut,
		Obj:    &ipam,
		Tenant: "admin",
		Model:  "ipamdnsproviderprofile",
	}
	restLayer := rest.NewRestOperations(nil, true)
	return restLayer.AviRestOperateWrapper(aviClient, []*utils.RestOp{&restOp}, "aviCleanup")
}

func convertPemToDer(cert string) string {
	cert = strings.TrimPrefix(cert, "-----BEGIN CERTIFICATE-----")
	cert = strings.TrimSuffix(cert, "-----END CERTIFICATE-----")
	return strings.ReplaceAll(cert, "\n", "")
}
