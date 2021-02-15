package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/avinetworks/sdk/go/models"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AviEvhVsNode struct {
	EVHParent        bool
	VHParentName     string
	VHDomainNames    []string
	EvhNodes         []*AviEvhVsNode
	EvhHostName      string
	EvhPath          string
	EvhMatchCriteria string
	// props from avi vs node
	Name                  string
	Tenant                string
	ServiceEngineGroup    string
	ApplicationProfile    string
	NetworkProfile        string
	Enabled               *bool
	PortProto             []AviPortHostProtocol // for listeners
	DefaultPool           string
	EastWest              bool
	CloudConfigCksum      uint32
	DefaultPoolGroup      string
	HTTPChecksum          uint32
	PoolGroupRefs         []*AviPoolGroupNode
	PoolRefs              []*AviPoolNode
	TCPPoolGroupRefs      []*AviPoolGroupNode
	HTTPDSrefs            []*AviHTTPDataScriptNode
	PassthroughChildNodes []*AviEvhVsNode
	SharedVS              bool
	CACertRefs            []*AviTLSKeyCertNode
	SSLKeyCertRefs        []*AviTLSKeyCertNode
	HttpPolicyRefs        []*AviHttpPolicySetNode
	VSVIPRefs             []*AviVSVIPNode
	L4PolicyRefs          []*AviL4PolicyNode
	TLSType               string
	ServiceMetadata       avicache.ServiceMetadataObj
	VrfContext            string
	WafPolicyRef          string
	AppProfileRef         string
	AnalyticsProfileRef   string
	ErrorPageProfileRef   string
	HttpPolicySetRefs     []string
	SSLProfileRef         string
	VsDatascriptRefs      []string
	SSLKeyCertAviRef      string
}

func (o *AviObjectGraph) GetAviEvhVS() []*AviEvhVsNode {
	var aviVs []*AviEvhVsNode
	for _, model := range o.modelNodes {
		vs, ok := model.(*AviEvhVsNode)
		if ok {
			aviVs = append(aviVs, vs)
		}
	}
	return aviVs
}

func (v *AviEvhVsNode) GetCheckSum() uint32 {
	// Calculate checksum and return
	v.CalculateCheckSum()
	return v.CloudConfigCksum
}

func (v *AviEvhVsNode) GetEvhNodeForName(EVHNodeName string) *AviEvhVsNode {
	for _, evhNode := range v.EvhNodes {
		if evhNode.Name == EVHNodeName {
			return evhNode
		}
	}
	return nil
}

func (o *AviEvhVsNode) CheckCACertNodeNameNChecksum(cacertNodeName string, checksum uint32) bool {
	for _, caCert := range o.CACertRefs {
		if caCert.Name == cacertNodeName {
			//Check if their checksums are same
			if caCert.GetCheckSum() == checksum {
				return false
			}
		}
	}
	return true
}

func (o *AviEvhVsNode) CheckSSLCertNodeNameNChecksum(sslNodeName string, checksum uint32) bool {
	for _, sslCert := range o.SSLKeyCertRefs {
		if sslCert.Name == sslNodeName {
			//Check if their checksums are same
			if sslCert.GetCheckSum() == checksum {
				return false
			}
		}
	}
	return true
}

func (o *AviEvhVsNode) CheckPGNameNChecksum(pgNodeName string, checksum uint32) bool {
	for _, pg := range o.PoolGroupRefs {
		if pg.Name == pgNodeName {
			//Check if their checksums are same
			if pg.GetCheckSum() == checksum {
				return false
			} else {
				return true
			}
		}
	}
	return true
}

func (o *AviEvhVsNode) CheckPoolNChecksum(poolNodeName string, checksum uint32) bool {
	for _, pool := range o.PoolRefs {
		if pool.Name == poolNodeName {
			//Check if their checksums are same
			if pool.GetCheckSum() == checksum {
				return false
			}
		}
	}
	return true
}

func (o *AviEvhVsNode) GetPGForVSByName(pgName string) *AviPoolGroupNode {
	for _, pgNode := range o.PoolGroupRefs {
		if pgNode.Name == pgName {
			return pgNode
		}
	}
	return nil
}

func (o *AviEvhVsNode) ReplaceEvhPoolInEVHNode(newPoolNode *AviPoolNode, key string) {
	for i, pool := range o.PoolRefs {
		if pool.Name == newPoolNode.Name {
			o.PoolRefs = append(o.PoolRefs[:i], o.PoolRefs[i+1:]...)
			o.PoolRefs = append(o.PoolRefs, newPoolNode)
			utils.AviLog.Infof("key: %s, msg: replaced evh pool in model: %s Pool name: %s", key, o.Name, pool.Name)
			return
		}
	}
	// If we have reached here it means we haven't found a match. Just append the pool.
	o.PoolRefs = append(o.PoolRefs, newPoolNode)
	return
}

func (o *AviEvhVsNode) ReplaceEvhPGInEVHNode(newPGNode *AviPoolGroupNode, key string) {
	for i, pg := range o.PoolGroupRefs {
		if pg.Name == newPGNode.Name {
			o.PoolGroupRefs = append(o.PoolGroupRefs[:i], o.PoolGroupRefs[i+1:]...)
			o.PoolGroupRefs = append(o.PoolGroupRefs, newPGNode)
			utils.AviLog.Infof("key: %s, msg: replaced evh pg in model: %s Pool name: %s", key, o.Name, pg.Name)
			return
		}
	}
	// If we have reached here it means we haven't found a match. Just append.
	o.PoolGroupRefs = append(o.PoolGroupRefs, newPGNode)
	return
}

func (o *AviEvhVsNode) DeleteCACertRefInEVHNode(cacertNodeName, key string) {
	for i, cacert := range o.CACertRefs {
		if cacert.Name == cacertNodeName {
			o.CACertRefs = append(o.CACertRefs[:i], o.CACertRefs[i+1:]...)
			utils.AviLog.Infof("key: %s, msg: replaced cacert for evh in model: %s Pool name: %s", key, o.Name, cacert.Name)
			return
		}
	}
}

func (o *AviEvhVsNode) ReplaceCACertRefInEVHNode(cacertNode *AviTLSKeyCertNode, key string) {
	for i, cacert := range o.CACertRefs {
		if cacert.Name == cacertNode.Name {
			o.CACertRefs = append(o.CACertRefs[:i], o.CACertRefs[i+1:]...)
			o.CACertRefs = append(o.CACertRefs, cacertNode)
			utils.AviLog.Infof("key: %s, msg: replaced cacert for evh in model: %s Pool name: %s", key, o.Name, cacert.Name)
			return
		}
	}
	// If we have reached here it means we haven't found a match. Just append.
	o.CACertRefs = append(o.CACertRefs, cacertNode)
}

func (o *AviEvhVsNode) ReplaceEvhSSLRefInEVHNode(newSslNode *AviTLSKeyCertNode, key string) {
	for i, ssl := range o.SSLKeyCertRefs {
		if ssl.Name == newSslNode.Name {
			o.SSLKeyCertRefs = append(o.SSLKeyCertRefs[:i], o.SSLKeyCertRefs[i+1:]...)
			o.SSLKeyCertRefs = append(o.SSLKeyCertRefs, newSslNode)
			utils.AviLog.Infof("key: %s, msg: replaced evh ssl in model: %s Pool name: %s", key, o.Name, ssl.Name)
			return
		}
	}
	// If we have reached here it means we haven't found a match. Just append.
	o.SSLKeyCertRefs = append(o.SSLKeyCertRefs, newSslNode)
	return
}

func (v *AviEvhVsNode) GetNodeType() string {
	return "VirtualServiceNode"
}

func (v *AviEvhVsNode) CalculateCheckSum() {
	portproto := v.PortProto
	sort.Slice(portproto, func(i, j int) bool {
		return portproto[i].Name < portproto[j].Name
	})

	var dsChecksum, httppolChecksum, evhChecksum, sslkeyChecksum, l4policyChecksum, passthroughChecksum, vsvipChecksum uint32

	for _, ds := range v.HTTPDSrefs {
		dsChecksum += ds.GetCheckSum()
	}

	for _, httppol := range v.HttpPolicyRefs {
		httppolChecksum += httppol.GetCheckSum()
	}

	for _, EVHNode := range v.EvhNodes {
		evhChecksum += EVHNode.GetCheckSum()
	}

	for _, cacert := range v.CACertRefs {
		sslkeyChecksum += cacert.GetCheckSum()
	}

	for _, sslkeycert := range v.SSLKeyCertRefs {
		sslkeyChecksum += sslkeycert.GetCheckSum()
	}

	for _, vsvipref := range v.VSVIPRefs {
		vsvipChecksum += vsvipref.GetCheckSum()
	}

	for _, l4policy := range v.L4PolicyRefs {
		l4policyChecksum += l4policy.GetCheckSum()
	}

	for _, passthroughChild := range v.PassthroughChildNodes {
		passthroughChecksum += passthroughChild.GetCheckSum()
	}

	// keep the order of these policies
	policies := v.HttpPolicySetRefs
	scripts := v.VsDatascriptRefs

	vsRefs := v.WafPolicyRef +
		v.AppProfileRef +
		utils.Stringify(policies) +
		v.AnalyticsProfileRef +
		v.ErrorPageProfileRef +
		v.SSLProfileRef

	if len(scripts) > 0 {
		vsRefs += utils.Stringify(scripts)
	}

	checksum := dsChecksum +
		httppolChecksum +
		evhChecksum +
		utils.Hash(v.ApplicationProfile) +
		utils.Hash(v.NetworkProfile) +
		utils.Hash(utils.Stringify(portproto)) +
		sslkeyChecksum +
		vsvipChecksum +
		utils.Hash(vsRefs) +
		l4policyChecksum +
		passthroughChecksum +
		utils.Hash(v.EvhHostName) +
		utils.Hash(v.EvhPath)

	if v.Enabled != nil {
		checksum += utils.Hash(utils.Stringify(v.Enabled))
	}

	v.CloudConfigCksum = checksum
}

func (v *AviEvhVsNode) CopyNode() AviModelNode {
	newNode := AviEvhVsNode{}
	bytes, err := json.Marshal(v)
	if err != nil {
		utils.AviLog.Warnf("Unable to marshal AviEvhVsNode: %s", err)
	}
	err = json.Unmarshal(bytes, &newNode)
	if err != nil {
		utils.AviLog.Warnf("Unable to unmarshal AviEvhVsNode: %s", err)
	}
	return &newNode
}

// Insecure ingress graph functions below

func (o *AviObjectGraph) ConstructAviL7SharedVsNodeForEvh(vsName string, key string) *AviEvhVsNode {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	// This is a shared VS - always created in the admin namespace for now.
	avi_vs_meta := &AviEvhVsNode{Name: vsName, Tenant: lib.GetTenant(),
		EastWest: false, SharedVS: true}
	if lib.GetSEGName() != lib.DEFAULT_SE_GROUP {
		avi_vs_meta.ServiceEngineGroup = lib.GetSEGName()
	}
	// Hard coded ports for the shared VS
	var portProtocols []AviPortHostProtocol
	httpPort := AviPortHostProtocol{Port: 80, Protocol: utils.HTTP}
	httpsPort := AviPortHostProtocol{Port: 443, Protocol: utils.HTTP, EnableSSL: true}
	portProtocols = append(portProtocols, httpPort)
	portProtocols = append(portProtocols, httpsPort)
	avi_vs_meta.PortProto = portProtocols
	// Default case.
	avi_vs_meta.ApplicationProfile = utils.DEFAULT_L7_SECURE_APP_PROFILE
	avi_vs_meta.NetworkProfile = utils.DEFAULT_TCP_NW_PROFILE
	avi_vs_meta.EVHParent = true

	vrfcontext := lib.GetVrf()
	avi_vs_meta.VrfContext = vrfcontext

	o.AddModelNode(avi_vs_meta)

	var fqdns []string

	subDomains := GetDefaultSubDomain()
	if subDomains != nil {
		var fqdn string
		if strings.HasPrefix(subDomains[0], ".") {
			fqdn = vsName + "." + lib.GetTenant() + subDomains[0]
		} else {
			fqdn = vsName + "." + lib.GetTenant() + "." + subDomains[0]
		}
		fqdns = append(fqdns, fqdn)
	} else {
		utils.AviLog.Warnf("key: %s, msg: there is no nsipamdns configured in the cloud, not configuring the default fqdn", key)
	}
	vsVipNode := &AviVSVIPNode{Name: lib.GetVsVipName(vsName), Tenant: lib.GetTenant(), FQDNs: fqdns,
		EastWest: false, VrfContext: vrfcontext}
	avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
	return avi_vs_meta
}

func (o *AviObjectGraph) BuildPolicyPGPoolsForEVH(vsNode []*AviEvhVsNode, childNode *AviEvhVsNode, namespace string, ingName string, key string, isIngr bool, host string, path IngressHostPathSvc) {
	localPGList := make(map[string]*AviPoolGroupNode)

	pgName := lib.GetEvhVsPoolNPgName(ingName, namespace, host, path.Path)
	var pgNode *AviPoolGroupNode
	// There can be multiple services for the same path in case of alternate backend.
	// In that case, make sure we are creating only one PG per path
	pgNode, pgfound := localPGList[pgName]
	if !pgfound {
		pgNode = &AviPoolGroupNode{Name: pgName, Tenant: lib.GetTenant()}
		localPGList[pgName] = pgNode
	}

	var poolName string
	// Do not use serviceName in evh Pool Name for ingress for backward compatibility
	if isIngr {
		poolName = lib.GetEvhVsPoolNPgName(ingName, namespace, host, path.Path)
	} else {
		poolName = lib.GetEvhVsPoolNPgName(ingName, namespace, host, path.Path, path.ServiceName)
	}
	hostSlice := []string{host}
	poolNode := &AviPoolNode{
		Name:       poolName,
		PortName:   path.PortName,
		Tenant:     lib.GetTenant(),
		VrfContext: lib.GetVrf(),
		ServiceMetadata: avicache.ServiceMetadataObj{
			IngressName: ingName,
			Namespace:   namespace,
			HostNames:   hostSlice,
		},
	}

	if !lib.IsNodePortMode() {
		if servers := PopulateServers(poolNode, namespace, path.ServiceName, true, key); servers != nil {
			poolNode.Servers = servers
		}
	} else {
		if servers := PopulateServersForNodePort(poolNode, namespace, path.ServiceName, true, key); servers != nil {
			poolNode.Servers = servers
		}
	}
	pool_ref := fmt.Sprintf("/api/pool?name=%s", poolNode.Name)
	ratio := path.weight
	pgNode.Members = append(pgNode.Members, &avimodels.PoolGroupMember{PoolRef: &pool_ref, Ratio: &ratio})

	if childNode.CheckPGNameNChecksum(pgNode.Name, pgNode.GetCheckSum()) {
		childNode.ReplaceEvhPGInEVHNode(pgNode, key)
	}
	if childNode.CheckPoolNChecksum(poolNode.Name, poolNode.GetCheckSum()) {
		// Replace the poolNode.
		childNode.ReplaceEvhPoolInEVHNode(poolNode, key)
	}
	o.AddModelNode(poolNode)

	utils.AviLog.Infof("key: %s, msg: added pools and poolgroups. childNodeChecksum for childNode :%s is :%v", key, childNode.Name, childNode.GetCheckSum())

}

func ProcessInsecureHostsForEVH(routeIgrObj RouteIngressModel, key string, parsedIng IngressConfig, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost, hostsMap map[string]*objects.RouteIngrhost) {
	for host, pathsvcmap := range parsedIng.IngressHostMap {
		// Remove this entry from storedHosts. First check if the host exists in the stored map or not.
		hostData, found := Storedhosts[host]
		if found && hostData.InsecurePolicy != lib.PolicyNone {
			// Verify the paths and take out the paths that are not need.
			pathSvcDiff := routeIgrObj.GetDiffPathSvc(hostData.PathSvc, pathsvcmap)
			if len(pathSvcDiff) == 0 {
				// Marking the entry as None to handle delete stale config
				Storedhosts[host].InsecurePolicy = lib.PolicyNone
				Storedhosts[host].SecurePolicy = lib.PolicyNone
			} else {
				hostData.PathSvc = pathSvcDiff
			}
		}
		if _, ok := hostsMap[host]; !ok {
			hostsMap[host] = &objects.RouteIngrhost{
				SecurePolicy: lib.PolicyNone,
			}
		}
		hostsMap[host].InsecurePolicy = lib.PolicyAllow
		hostsMap[host].PathSvc = getPathSvc(pathsvcmap)

		shardVsName := DeriveHostNameShardVSForEvh(host, key)
		if shardVsName == "" {
			// If we aren't able to derive the ShardVS name, we should return
			return
		}
		modelName := lib.GetModelName(lib.GetTenant(), shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Infof("key: %s, msg: model not found, generating new model with name: %s", key, modelName)
			aviModel = NewAviObjectGraph()
			aviModel.(*AviObjectGraph).ConstructAviL7SharedVsNodeForEvh(shardVsName, key)
		}

		// fill PG pool and pool members for host + path

		// We create evh child vs, pg, pools and attach servers to them here.
		vsNode := aviModel.(*AviObjectGraph).GetAviEvhVS()
		ingName := routeIgrObj.GetName()
		namespace := routeIgrObj.GetNamespace()
		for _, path := range pathsvcmap {
			evhNodeName := lib.GetEvhVsPoolNPgName(ingName, namespace, host, path.Path)
			evhNode := vsNode[0].GetEvhNodeForName(evhNodeName)
			if evhNode == nil {
				evhNode = &AviEvhVsNode{
					Name:         evhNodeName,
					VHParentName: vsNode[0].Name,
					Tenant:       lib.GetTenant(),
					EVHParent:    false,
					EvhHostName:  host,
					EvhPath:      path.Path,
				}

				if path.PathType == networkingv1beta1.PathTypeExact {
					evhNode.EvhMatchCriteria = "EQUALS"
				} else {
					// PathTypePrefix and PathTypeImplementationSpecific
					// default behaviour for AKO set be Prefix match on the path
					evhNode.EvhMatchCriteria = "BEGINS_WITH"
				}
			}
			if lib.GetSEGName() != lib.DEFAULT_SE_GROUP {
				evhNode.ServiceEngineGroup = lib.GetSEGName()
			}
			evhNode.VrfContext = lib.GetVrf()
			isIngr := routeIgrObj.GetType() == utils.Ingress
			aviModel.(*AviObjectGraph).BuildPolicyPGPoolsForEVH(vsNode, evhNode, namespace, ingName, key, isIngr, host, path)
			foundEvhModel := FindAndReplaceEvhInModel(evhNode, vsNode, key)
			if !foundEvhModel {
				vsNode[0].EvhNodes = append(vsNode[0].EvhNodes, evhNode)
			}
		}

		utils.AviLog.Debugf("key: %s, Saving Model in ProcessInsecureHostsForEVH : %v", key, utils.Stringify(vsNode))
		changedModel := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if !utils.HasElem(modelList, modelName) && changedModel {
			*modelList = append(*modelList, modelName)
		}
	}

	utils.AviLog.Debugf("key: %s, msg: Storedhosts after processing insecurehosts: %s", key, utils.Stringify(Storedhosts))
}

// secure ingress graph functions

// BuildCACertNode : Build a new node to store CA cert, this would be referred by the corresponding keycert
func (o *AviObjectGraph) BuildCACertNodeForEvh(tlsNode *AviEvhVsNode, cacert, keycertname, key string) string {
	cacertNode := &AviTLSKeyCertNode{Name: lib.GetCACertNodeName(keycertname), Tenant: lib.GetTenant()}
	cacertNode.Type = lib.CertTypeCA
	cacertNode.Cert = []byte(cacert)

	if tlsNode.CheckCACertNodeNameNChecksum(cacertNode.Name, cacertNode.GetCheckSum()) {
		if len(tlsNode.CACertRefs) == 1 {
			tlsNode.CACertRefs[0] = cacertNode
			utils.AviLog.Warnf("key: %s, msg: duplicate cacerts detected for %s, overwriting", key, cacertNode.Name)
		} else {
			tlsNode.ReplaceCACertRefInEVHNode(cacertNode, key)
		}
	}
	return cacertNode.Name
}

func (o *AviObjectGraph) BuildTlsCertNodeForEvh(svcLister *objects.SvcLister, tlsNode *AviEvhVsNode, namespace string, tlsData TlsSettings, key string, host ...string) bool {
	mClient := utils.GetInformers().ClientSet
	secretName := tlsData.SecretName
	secretNS := tlsData.SecretNS
	if secretNS == "" {
		secretNS = namespace
	}

	var certNode *AviTLSKeyCertNode
	if len(host) > 0 {
		certNode = &AviTLSKeyCertNode{Name: lib.GetTLSKeyCertNodeName(namespace, secretName, host[0]), Tenant: lib.GetTenant()}
	} else {
		certNode = &AviTLSKeyCertNode{Name: lib.GetTLSKeyCertNodeName(namespace, secretName), Tenant: lib.GetTenant()}
	}
	certNode.Type = lib.CertTypeVS

	// Openshift Routes do not refer to a secret, instead key/cert values are mentioned in the route
	if strings.HasPrefix(secretName, lib.RouteSecretsPrefix) {
		if tlsData.cert != "" && tlsData.key != "" {
			certNode.Cert = []byte(tlsData.cert)
			certNode.Key = []byte(tlsData.key)
			if tlsData.cacert != "" {
				certNode.CACert = o.BuildCACertNodeForEvh(tlsNode, tlsData.cacert, certNode.Name, key)
			} else {
				tlsNode.DeleteCACertRefInEVHNode(lib.GetCACertNodeName(certNode.Name), key)
			}
		} else {
			ok, _ := svcLister.IngressMappings(namespace).GetSecretToIng(secretName)
			if ok {
				svcLister.IngressMappings(namespace).DeleteSecretToIngMapping(secretName)
			}
			utils.AviLog.Infof("key: %s, msg: no cert/key specified for TLS route")
			//To Do: use a Default secret if required
			return false
		}
	} else {
		secretObj, err := mClient.CoreV1().Secrets(secretNS).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil || secretObj == nil {
			// This secret has been deleted.
			ok, ingNames := svcLister.IngressMappings(namespace).GetSecretToIng(secretName)
			if ok {
				// Delete the secret key in the cache if it has no references
				if len(ingNames) == 0 {
					svcLister.IngressMappings(namespace).DeleteSecretToIngMapping(secretName)
				}
			}
			utils.AviLog.Infof("key: %s, msg: secret: %s has been deleted, err: %s", key, secretName, err)
			return false
		}
		keycertMap := secretObj.Data
		cert, ok := keycertMap[tlsCert]
		if ok {
			certNode.Cert = cert
		} else {
			utils.AviLog.Infof("key: %s, msg: certificate not found for secret: %s", key, secretObj.Name)
			return false
		}
		tlsKey, keyfound := keycertMap[utils.K8S_TLS_SECRET_KEY]
		if keyfound {
			certNode.Key = tlsKey
		} else {
			utils.AviLog.Infof("key: %s, msg: key not found for secret: %s", key, secretObj.Name)
			return false
		}
		utils.AviLog.Infof("key: %s, msg: Added the secret object to tlsnode: %s", key, secretObj.Name)
	}
	// If this SSLCertRef is already present don't add it.
	if len(host) > 0 {
		if tlsNode.CheckSSLCertNodeNameNChecksum(lib.GetTLSKeyCertNodeName(namespace, secretName, host[0]), certNode.GetCheckSum()) {
			tlsNode.ReplaceEvhSSLRefInEVHNode(certNode, key)
		}
	} else {
		tlsNode.SSLKeyCertRefs = append(tlsNode.SSLKeyCertRefs, certNode)
	}
	return true
}

func ProcessSecureHostsForEVH(routeIgrObj RouteIngressModel, key string, parsedIng IngressConfig, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost,
	hostsMap map[string]*objects.RouteIngrhost, fullsync bool, sharedQueue *utils.WorkerQueue) {
	utils.AviLog.Debugf("key: %s, msg: Storedhosts before processing securehosts: %v", key, utils.Stringify(Storedhosts))

	for _, tlssetting := range parsedIng.TlsCollection {
		locEvhHostMap := evhNodeHostName(routeIgrObj, tlssetting, routeIgrObj.GetName(), routeIgrObj.GetNamespace(), key, fullsync, sharedQueue, modelList)
		for host, newPathSvc := range locEvhHostMap {
			// Remove this entry from storedHosts. First check if the host exists in the stored map or not.
			hostData, found := Storedhosts[host]
			if found && hostData.SecurePolicy == lib.PolicyEdgeTerm {
				// Verify the paths and take out the paths that are not need.
				pathSvcDiff := routeIgrObj.GetDiffPathSvc(hostData.PathSvc, newPathSvc)

				if len(pathSvcDiff) == 0 {
					Storedhosts[host].SecurePolicy = lib.PolicyNone
					Storedhosts[host].InsecurePolicy = lib.PolicyNone
				} else {
					hostData.PathSvc = pathSvcDiff
				}
			}
			if _, ok := hostsMap[host]; !ok {
				hostsMap[host] = &objects.RouteIngrhost{
					InsecurePolicy: lib.PolicyNone,
				}
			}
			hostsMap[host].SecurePolicy = lib.PolicyEdgeTerm
			if tlssetting.redirect == true {
				hostsMap[host].InsecurePolicy = lib.PolicyRedirect
			}
			hostsMap[host].PathSvc = getPathSvc(newPathSvc)
		}
	}
	utils.AviLog.Debugf("key: %s, msg: Storedhosts after processing securehosts: %s", key, utils.Stringify(Storedhosts))
}

func evhNodeHostName(routeIgrObj RouteIngressModel, tlssetting TlsSettings, ingName, namespace, key string, fullsync bool, sharedQueue *utils.WorkerQueue, modelList *[]string) map[string][]IngressHostPathSvc {
	hostPathSvcMap := make(map[string][]IngressHostPathSvc)
	for host, paths := range tlssetting.Hosts {
		var hosts []string
		hostPathSvcMap[host] = paths
		hostMap := HostNamePathSecrets{paths: getPaths(paths), secretName: tlssetting.SecretName}
		found, ingressHostMap := SharedHostNameLister().Get(host)
		if found {
			// Replace the ingress map for this host.
			ingressHostMap.HostNameMap[namespace+"/"+ingName] = hostMap
			ingressHostMap.GetIngressesForHostName(host)
		} else {
			// Create the map
			ingressHostMap = NewSecureHostNameMapProp()
			ingressHostMap.HostNameMap[namespace+"/"+ingName] = hostMap
		}
		SharedHostNameLister().Save(host, ingressHostMap)
		hosts = append(hosts, host)
		shardVsName := DeriveHostNameShardVSForEvh(host, key)
		// For each host, create a EVH node with the secret giving us the key and cert.
		// construct a EVH child VS node per tls setting which corresponds to one secret
		if shardVsName == "" {
			// If we aren't able to derive the ShardVS name, we should return
			//return hostPathMap
			return hostPathSvcMap
		}
		model_name := lib.GetModelName(lib.GetTenant(), shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(model_name)
		if !found || aviModel == nil {
			utils.AviLog.Infof("key: %s, msg: model not found, generating new model with name: %s", key, model_name)
			aviModel = NewAviObjectGraph()
			aviModel.(*AviObjectGraph).ConstructAviL7SharedVsNodeForEvh(shardVsName, key)
		}
		vsNode := aviModel.(*AviObjectGraph).GetAviEvhVS()

		certsBuilt := false
		evhSecretName := tlssetting.SecretName
		re := regexp.MustCompile(fmt.Sprintf(`^%s.*`, lib.DummySecret))
		if re.MatchString(evhSecretName) {
			evhSecretName = strings.Split(evhSecretName, "/")[1]
			certsBuilt = true
		}

		for _, path := range paths {
			evhNode := vsNode[0].GetEvhNodeForName(lib.GetEvhTlsNodeName(ingName, namespace, evhSecretName, host, path.Path))
			if evhNode == nil {
				evhNode = &AviEvhVsNode{
					Name:         lib.GetEvhTlsNodeName(ingName, namespace, evhSecretName, host, path.Path),
					VHParentName: vsNode[0].Name,
					Tenant:       lib.GetTenant(),
					EVHParent:    false,
					EvhHostName:  host,
					EvhPath:      path.Path,
					ServiceMetadata: avicache.ServiceMetadataObj{
						NamespaceIngressName: ingressHostMap.GetIngressesForHostName(host),
						Namespace:            namespace,
						HostNames:            hosts,
					},
				}

				if path.PathType == networkingv1beta1.PathTypeExact {
					evhNode.EvhMatchCriteria = "EQUALS"
				} else {
					// PathTypePrefix and PathTypeImplementationSpecific
					// default behaviour for AKO set be Prefix match on the path
					evhNode.EvhMatchCriteria = "BEGINS_WITH"
				}
				if lib.GetSEGName() != lib.DEFAULT_SE_GROUP {
					evhNode.ServiceEngineGroup = lib.GetSEGName()
				}
			} else {
				// The evh node exists, just update the svc metadata
				evhNode.ServiceMetadata.NamespaceIngressName = ingressHostMap.GetIngressesForHostName(host)
				evhNode.ServiceMetadata.Namespace = namespace
				evhNode.ServiceMetadata.HostNames = hosts
				if evhNode.SSLKeyCertAviRef != "" {
					certsBuilt = true
				}
			}
			if lib.GetSEGName() != lib.DEFAULT_SE_GROUP {
				evhNode.ServiceEngineGroup = lib.GetSEGName()
			}
			evhNode.VrfContext = lib.GetVrf()
			if !certsBuilt {
				certsBuilt = aviModel.(*AviObjectGraph).BuildTlsCertNodeForEvh(routeIgrObj.GetSvcLister(), vsNode[0], namespace, tlssetting, key, host)
			}
			if certsBuilt {
				isIngr := routeIgrObj.GetType() == utils.Ingress
				aviModel.(*AviObjectGraph).BuildPolicyPGPoolsForEVH(vsNode, evhNode, namespace, ingName, key, isIngr, host, path)
				foundEvhModel := FindAndReplaceEvhInModel(evhNode, vsNode, key)
				if !foundEvhModel {
					vsNode[0].EvhNodes = append(vsNode[0].EvhNodes, evhNode)
				}

				RemoveRedirectHTTPPolicyInModelForEvh(vsNode[0], host, key)

				if tlssetting.redirect == true {
					aviModel.(*AviObjectGraph).BuildPolicyRedirectForVSForEvh(vsNode, host, namespace, ingName, key)
				}
				// TODO: Enable host rule
				// BuildL7HostRule(host, namespace, ingName, key, evhNode)
			} else {
				hostMapOk, ingressHostMap := SharedHostNameLister().Get(host)
				if hostMapOk {
					// Replace the ingress map for this host.
					keyToRemove := namespace + "/" + ingName
					delete(ingressHostMap.HostNameMap, keyToRemove)
					SharedHostNameLister().Save(host, ingressHostMap)
				}
				// Since the cert couldn't be built, remove the evh node from the model
				RemoveEvhInModel(evhNode.Name, vsNode, key)
				RemoveRedirectHTTPPolicyInModelForEvh(vsNode[0], host, key)

			}
			// Only add this node to the list of models if the checksum has changed.
			utils.AviLog.Debugf("key: %s, Saving Model: %v", key, utils.Stringify(vsNode))
			modelChanged := saveAviModel(model_name, aviModel.(*AviObjectGraph), key)
			if !utils.HasElem(*modelList, model_name) && modelChanged {
				*modelList = append(*modelList, model_name)
			}
		}
	}

	return hostPathSvcMap
}

// Util functions

func FindAndReplaceEvhInModel(currentEvhNode *AviEvhVsNode, modelEvhNodes []*AviEvhVsNode, key string) bool {
	for i, modelEvhNode := range modelEvhNodes[0].EvhNodes {
		if currentEvhNode.Name == modelEvhNode.Name {
			// Check if the checksums are same
			if !(modelEvhNode.GetCheckSum() == currentEvhNode.GetCheckSum()) {
				// The checksums are not same. Replace this evh node
				modelEvhNodes[0].EvhNodes = append(modelEvhNodes[0].EvhNodes[:i], modelEvhNodes[0].EvhNodes[i+1:]...)
				modelEvhNodes[0].EvhNodes = append(modelEvhNodes[0].EvhNodes, currentEvhNode)
				utils.AviLog.Infof("key: %s, msg: replaced evh node in model: %s", key, currentEvhNode.Name)
			}
			return true
		}
	}
	return false
}

func RemoveEvhInModel(currentEvhNodeName string, modelEvhNodes []*AviEvhVsNode, key string) {
	if len(modelEvhNodes[0].EvhNodes) > 0 {
		for i, modelEvhNode := range modelEvhNodes[0].EvhNodes {
			if currentEvhNodeName == modelEvhNode.Name {
				modelEvhNodes[0].EvhNodes = append(modelEvhNodes[0].EvhNodes[:i], modelEvhNodes[0].EvhNodes[i+1:]...)
				utils.AviLog.Infof("key: %s, msg: deleted evh node in model: %s", key, currentEvhNodeName)
				return
			}
		}
	}
}

func FindAndReplaceRedirectHTTPPolicyInModelforEvh(vsNode *AviEvhVsNode, httpPolicy *AviHttpPolicySetNode, hostname, key string) bool {
	for _, policy := range vsNode.HttpPolicyRefs {
		if policy.Name == httpPolicy.Name && policy.CloudConfigCksum != httpPolicy.CloudConfigCksum {
			if !utils.HasElem(policy.RedirectPorts[0].Hosts, hostname) {
				policy.RedirectPorts[0].Hosts = append(policy.RedirectPorts[0].Hosts, hostname)
				utils.AviLog.Infof("key: %s, msg: replaced host %s for policy %s in model", key, hostname, policy.Name)
			}
			return true
		}
	}
	return false
}

func RemoveRedirectHTTPPolicyInModelForEvh(vsNode *AviEvhVsNode, hostname, key string) {
	policyName := lib.GetL7HttpRedirPolicy(vsNode.Name)
	deletePolicy := false
	for i, policy := range vsNode.HttpPolicyRefs {
		if policy.Name == policyName {
			// one redirect policy per shard vs
			policy.RedirectPorts[0].Hosts = utils.Remove(policy.RedirectPorts[0].Hosts, hostname)
			utils.AviLog.Infof("key: %s, msg: removed host %s from policy %s in model %v", key, hostname, policy.Name, policy.RedirectPorts[0].Hosts)
			if len(policy.RedirectPorts[0].Hosts) == 0 {
				deletePolicy = true
			}

			if deletePolicy {
				vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs[:i], vsNode.HttpPolicyRefs[i+1:]...)
				utils.AviLog.Infof("key: %s, msg: removed policy %s in model", key, policy.Name)
			}
		}
	}
}

func RemoveFQDNsFromModelForEvh(vsNode *AviEvhVsNode, hosts []string, key string) {
	if len(vsNode.VSVIPRefs) > 0 {
		for i, fqdn := range vsNode.VSVIPRefs[0].FQDNs {
			if utils.HasElem(hosts, fqdn) {
				// remove logic conainer-lib candidate
				vsNode.VSVIPRefs[0].FQDNs[i] = vsNode.VSVIPRefs[0].FQDNs[len(vsNode.VSVIPRefs[0].FQDNs)-1]
				vsNode.VSVIPRefs[0].FQDNs[len(vsNode.VSVIPRefs[0].FQDNs)-1] = ""
				vsNode.VSVIPRefs[0].FQDNs = vsNode.VSVIPRefs[0].FQDNs[:len(vsNode.VSVIPRefs[0].FQDNs)-1]
			}
		}
	}
}

// RouteIngrDeletePoolsByHostname : Based on DeletePoolsByHostname, delete pools and policies that are no longer required
func RouteIngrDeletePoolsByHostnameForEvh(routeIgrObj RouteIngressModel, namespace, objname, key string, fullsync bool, sharedQueue *utils.WorkerQueue) {
	ok, hostMap := routeIgrObj.GetSvcLister().IngressMappings(namespace).GetRouteIngToHost(objname)
	if !ok {
		utils.AviLog.Warnf("key: %s, msg: nothing to delete for route: %s", key, objname)
		return
	}

	utils.AviLog.Debugf("key: %s, msg: hosts to delete are :%s", key, utils.Stringify(hostMap))
	for host, hostData := range hostMap {
		shardVsName := DeriveHostNameShardVSForEvh(host, key)
		if hostData.SecurePolicy == lib.PolicyPass {
			shardVsName = lib.GetPassthroughShardVSName(host, key)
		}
		if shardVsName == "" {
			// If we aren't able to derive the ShardVS name, we should return
			utils.AviLog.Infof("key: %s, shard vs ndoe not found for host: %s", host)
			return
		}
		modelName := lib.GetModelName(lib.GetTenant(), shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Warnf("key: %s, msg: model not found during delete: %s", key, modelName)
			continue
		}
		// Delete the pool corresponding to this host
		if hostData.SecurePolicy == lib.PolicyEdgeTerm {
			aviModel.(*AviObjectGraph).DeleteVsForHostnameForEvh(shardVsName, host, routeIgrObj, hostData.PathSvc, key, true, true, true)
		} else if hostData.SecurePolicy == lib.PolicyPass {
			aviModel.(*AviObjectGraph).DeleteObjectsForPassthroughHost(shardVsName, host, routeIgrObj, hostData.PathSvc, key, true, true, true)
		}
		if hostData.InsecurePolicy == lib.PolicyAllow {
			aviModel.(*AviObjectGraph).DeleteVsForHostnameForEvh(shardVsName, host, routeIgrObj, hostData.PathSvc, key, true, true, false)
		}
		ok := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if ok && len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 && !fullsync {
			PublishKeyToRestLayer(modelName, key, sharedQueue)
		}
	}
	// Now remove the secret relationship
	routeIgrObj.GetSvcLister().IngressMappings(namespace).RemoveIngressSecretMappings(objname)
	utils.AviLog.Infof("key: %s, removed ingress mapping for: %s", key, objname)

	// Remove the hosts mapping for this ingress
	routeIgrObj.GetSvcLister().IngressMappings(namespace).DeleteIngToHostMapping(objname)

	// remove hostpath mappings
	updateHostPathCacheV2(namespace, objname, hostMap, nil)
}

//DeleteStaleData : delete pool, EVH VS and redirect policy which are present in the object store but no longer required.
func DeleteStaleDataForEvh(routeIgrObj RouteIngressModel, key string, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost, hostsMap map[string]*objects.RouteIngrhost) {
	utils.AviLog.Debugf("key: %s, msg: About to delete stale data EVH Stored hosts: %v, hosts map: %v", key, utils.Stringify(Storedhosts), utils.Stringify(hostsMap))
	for host, hostData := range Storedhosts {
		utils.AviLog.Debugf("host to del: %s, data : %s", host, utils.Stringify(hostData))
		shardVsName := DeriveHostNameShardVSForEvh(host, key)

		if shardVsName == "" {
			// If we aren't able to derive the ShardVS name, we should return
			return
		}
		modelName := lib.GetModelName(lib.GetTenant(), shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Warnf("key: %s, msg: model not found during delete: %s", key, modelName)
			continue
		}
		// By default remove both redirect and fqdn. So if the host isn't transitioning, then we will remove both.
		removeFqdn := true
		removeRedir := true
		currentData, ok := hostsMap[host]
		utils.AviLog.Warnf("key: %s, hostsMap: %s", key, utils.Stringify(hostsMap))
		// if route is transitioning from/to passthrough route, then always remove fqdn
		if ok && hostData.SecurePolicy != lib.PolicyPass && currentData.SecurePolicy != lib.PolicyPass {
			if currentData.InsecurePolicy == lib.PolicyRedirect {
				removeRedir = false
			}
			utils.AviLog.Infof("key: %s, host: %s, currentData: %v", key, host, currentData)
			removeFqdn = false
		}
		// Delete the pool corresponding to this host
		if hostData.SecurePolicy == lib.PolicyEdgeTerm {
			aviModel.(*AviObjectGraph).DeleteVsForHostnameForEvh(shardVsName, host, routeIgrObj, hostData.PathSvc, key, removeFqdn, removeRedir, true)
		}
		if hostData.InsecurePolicy != lib.PolicyNone {
			aviModel.(*AviObjectGraph).DeleteVsForHostnameForEvh(shardVsName, host, routeIgrObj, hostData.PathSvc, key, removeFqdn, removeRedir, false)

		}
		changedModel := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if !utils.HasElem(modelList, modelName) && changedModel {
			*modelList = append(*modelList, modelName)
		}
	}
}

func DeriveHostNameShardVSForEvh(hostname string, key string) string {
	// Read the value of the num_shards from the environment variable.
	utils.AviLog.Debugf("key: %s, msg: hostname for sharding: %s", key, hostname)
	var vsNum uint32
	shardSize := lib.GetshardSize()
	shardVsPrefix := lib.GetNamePrefix() + lib.ShardVSPrefix + "-EVH-"
	if shardSize != 0 {
		vsNum = utils.Bkt(hostname, shardSize)
		utils.AviLog.Debugf("key: %s, msg: VS number: %v", key, vsNum)
	} else {
		utils.AviLog.Warnf("key: %s, msg: the value for shard_vs_size does not match the ENUM values", key)
		return ""
	}
	vsName := shardVsPrefix + fmt.Sprint(vsNum)
	utils.AviLog.Infof("key: %s, msg: ShardVSName: %s", key, vsName)
	return vsName
}

func (o *AviObjectGraph) DeleteVsForHostnameForEvh(vsName, hostname string, routeIgrObj RouteIngressModel, pathSvc map[string][]string, key string, removeFqdn, removeRedir, secure bool) {

	o.Lock.Lock()
	defer o.Lock.Unlock()

	namespace := routeIgrObj.GetNamespace()
	ingName := routeIgrObj.GetName()
	vsNode := o.GetAviEvhVS()
	keepEvh := false
	if !secure {
		// Fetch the ingress evh vs that are present in the model and delete them.
		for path := range pathSvc {
			evhVsName := lib.GetEvhVsPoolNPgName(ingName, namespace, hostname, path)
			_ = o.RemoveEvhVsNode(evhVsName, vsNode, key, hostname)

		}

	} else {
		// Remove the ingress from the hostmap
		hostMapOk, ingressHostMap := SharedHostNameLister().Get(hostname)
		if hostMapOk {
			// Replace the ingress map for this host.
			keyToRemove := namespace + "/" + ingName
			delete(ingressHostMap.HostNameMap, keyToRemove)
			SharedHostNameLister().Save(hostname, ingressHostMap)
		}

		for path := range pathSvc {
			evhVsName := lib.GetEvhTlsNodeName(ingName, namespace, "", hostname, path)
			utils.AviLog.Infof("key: %s, msg: evh node to delete: %s", key, evhVsName)
			keepEvh = o.RemoveEvhVsNode(evhVsName, vsNode, key, hostname)

			if removeFqdn && !keepEvh {
				var hosts []string
				hosts = append(hosts, hostname)
				// Remove these hosts from the overall FQDN list
				RemoveFQDNsFromModelForEvh(vsNode[0], hosts, key)
			}
			if removeRedir && !keepEvh {
				RemoveRedirectHTTPPolicyInModelForEvh(vsNode[0], hostname, key)
			}

		}
	}

}

func (o *AviObjectGraph) RemoveEvhVsNode(evhVsName string, vsNode []*AviEvhVsNode, key string, hostname string) bool {
	utils.AviLog.Debugf("Removing EVH vs: %s", evhVsName)
	for _, modelEvhNode := range vsNode[0].EvhNodes {
		if evhVsName != modelEvhNode.Name {
			continue
		}
		RemoveEvhInModel(evhVsName, vsNode, key)
		SharedHostNameLister().Delete(hostname)
		return false
	}

	return true
}

func (o *AviObjectGraph) BuildPolicyRedirectForVSForEvh(vsNode []*AviEvhVsNode, hostname string, namespace, ingName, key string) {
	policyname := lib.GetL7HttpRedirPolicy(vsNode[0].Name)
	myHppMap := AviRedirectPort{
		Hosts:        []string{hostname},
		RedirectPort: 443,
		StatusCode:   lib.STATUS_REDIRECT,
		VsPort:       80,
	}

	redirectPolicy := &AviHttpPolicySetNode{
		Tenant:        lib.GetTenant(),
		Name:          policyname,
		RedirectPorts: []AviRedirectPort{myHppMap},
	}

	if policyFound := FindAndReplaceRedirectHTTPPolicyInModelforEvh(vsNode[0], redirectPolicy, hostname, key); !policyFound {
		redirectPolicy.CalculateCheckSum()
		vsNode[0].HttpPolicyRefs = append(vsNode[0].HttpPolicyRefs, redirectPolicy)
	}

}
