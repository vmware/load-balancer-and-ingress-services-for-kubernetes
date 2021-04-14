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

package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/avinetworks/sdk/go/models"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AviVsEvhSniModel : High Level interfaces that should be implemented by
// AviEvhVsNode and  AviVsNode
type AviVsEvhSniModel interface {
	GetName() string
	SetName(string)

	GetPoolRefs() []*AviPoolNode
	SetPoolRefs([]*AviPoolNode)

	GetPoolGroupRefs() []*AviPoolGroupNode
	SetPoolGroupRefs([]*AviPoolGroupNode)

	GetSSLKeyCertRefs() []*AviTLSKeyCertNode
	SetSSLKeyCertRefs([]*AviTLSKeyCertNode)

	GetHttpPolicyRefs() []*AviHttpPolicySetNode
	SetHttpPolicyRefs([]*AviHttpPolicySetNode)

	GetServiceMetadata() avicache.ServiceMetadataObj
	SetServiceMetadata(avicache.ServiceMetadataObj)

	GetSSLKeyCertAviRef() string
	SetSSLKeyCertAviRef(string)

	GetWafPolicyRef() string
	SetWafPolicyRef(string)

	GetHttpPolicySetRefs() []string
	SetHttpPolicySetRefs([]string)

	GetAppProfileRef() string
	SetAppProfileRef(string)

	GetAnalyticsProfileRef() string
	SetAnalyticsProfileRef(string)

	GetErrorPageProfileRef() string
	SetErrorPageProfileRef(string)

	GetSSLProfileRef() string
	SetSSLProfileRef(string)

	GetVsDatascriptRefs() []string
	SetVsDatascriptRefs([]string)

	GetEnabled() *bool
	SetEnabled(*bool)
}

type AviEvhVsNode struct {
	EVHParent     bool
	VHParentName  string
	VHDomainNames []string
	EvhNodes      []*AviEvhVsNode
	EvhHostName   string
	// props from avi vs node
	Name                string
	Tenant              string
	ServiceEngineGroup  string
	ApplicationProfile  string
	NetworkProfile      string
	EnableRhi           *bool
	Enabled             *bool
	PortProto           []AviPortHostProtocol // for listeners
	DefaultPool         string
	CloudConfigCksum    uint32
	DefaultPoolGroup    string
	HTTPChecksum        uint32
	PoolGroupRefs       []*AviPoolGroupNode
	PoolRefs            []*AviPoolNode
	HTTPDSrefs          []*AviHTTPDataScriptNode
	SharedVS            bool
	CACertRefs          []*AviTLSKeyCertNode
	SSLKeyCertRefs      []*AviTLSKeyCertNode
	HttpPolicyRefs      []*AviHttpPolicySetNode
	VSVIPRefs           []*AviVSVIPNode
	TLSType             string
	ServiceMetadata     avicache.ServiceMetadataObj
	VrfContext          string
	WafPolicyRef        string
	AppProfileRef       string
	AnalyticsProfileRef string
	ErrorPageProfileRef string
	HttpPolicySetRefs   []string
	VsDatascriptRefs    []string
	SSLProfileRef       string
	SSLKeyCertAviRef    string
}

// Implementing AviVsEvhSniModel

func (v *AviEvhVsNode) GetName() string {
	return v.Name
}

func (v *AviEvhVsNode) SetName(name string) {
	v.Name = name
}

func (v *AviEvhVsNode) GetPoolRefs() []*AviPoolNode {
	return v.PoolRefs
}

func (v *AviEvhVsNode) SetPoolRefs(poolRefs []*AviPoolNode) {
	v.PoolRefs = poolRefs
}

func (v *AviEvhVsNode) GetPoolGroupRefs() []*AviPoolGroupNode {
	return v.PoolGroupRefs
}

func (v *AviEvhVsNode) SetPoolGroupRefs(poolGroupRefs []*AviPoolGroupNode) {
	v.PoolGroupRefs = poolGroupRefs
}

func (v *AviEvhVsNode) GetSSLKeyCertRefs() []*AviTLSKeyCertNode {
	return v.SSLKeyCertRefs
}

func (v *AviEvhVsNode) SetSSLKeyCertRefs(sslKeyCertRefs []*AviTLSKeyCertNode) {
	v.SSLKeyCertRefs = sslKeyCertRefs
}

func (v *AviEvhVsNode) GetHttpPolicyRefs() []*AviHttpPolicySetNode {
	return v.HttpPolicyRefs
}

func (v *AviEvhVsNode) SetHttpPolicyRefs(httpPolicyRefs []*AviHttpPolicySetNode) {
	v.HttpPolicyRefs = httpPolicyRefs
}

func (v *AviEvhVsNode) GetServiceMetadata() avicache.ServiceMetadataObj {
	return v.ServiceMetadata
}

func (v *AviEvhVsNode) SetServiceMetadata(serviceMetadata avicache.ServiceMetadataObj) {
	v.ServiceMetadata = serviceMetadata
}

func (v *AviEvhVsNode) GetSSLKeyCertAviRef() string {
	return v.SSLKeyCertAviRef
}

func (v *AviEvhVsNode) SetSSLKeyCertAviRef(sslKeyCertAviRef string) {
	v.SSLKeyCertAviRef = sslKeyCertAviRef
}

func (v *AviEvhVsNode) GetWafPolicyRef() string {
	return v.WafPolicyRef
}

func (v *AviEvhVsNode) SetWafPolicyRef(wafPolicyRef string) {
	v.WafPolicyRef = wafPolicyRef
}

func (v *AviEvhVsNode) GetHttpPolicySetRefs() []string {
	return v.HttpPolicySetRefs
}

func (v *AviEvhVsNode) SetHttpPolicySetRefs(httpPolicySetRefs []string) {
	v.HttpPolicySetRefs = httpPolicySetRefs
}

func (v *AviEvhVsNode) GetAppProfileRef() string {
	return v.AppProfileRef
}

func (v *AviEvhVsNode) SetAppProfileRef(appProfileRef string) {
	v.AppProfileRef = appProfileRef
}

func (v *AviEvhVsNode) GetAnalyticsProfileRef() string {
	return v.AnalyticsProfileRef
}

func (v *AviEvhVsNode) SetAnalyticsProfileRef(AnalyticsProfileRef string) {
	v.AnalyticsProfileRef = AnalyticsProfileRef
}

func (v *AviEvhVsNode) GetErrorPageProfileRef() string {
	return v.ErrorPageProfileRef
}

func (v *AviEvhVsNode) SetErrorPageProfileRef(ErrorPageProfileRef string) {
	v.ErrorPageProfileRef = ErrorPageProfileRef
}

func (v *AviEvhVsNode) GetSSLProfileRef() string {
	return v.SSLProfileRef
}

func (v *AviEvhVsNode) SetSSLProfileRef(SSLProfileRef string) {
	v.SSLProfileRef = SSLProfileRef
}

func (v *AviEvhVsNode) GetVsDatascriptRefs() []string {
	return v.VsDatascriptRefs
}

func (v *AviEvhVsNode) SetVsDatascriptRefs(VsDatascriptRefs []string) {
	v.VsDatascriptRefs = VsDatascriptRefs
}

func (v *AviEvhVsNode) GetEnabled() *bool {
	return v.Enabled
}

func (v *AviEvhVsNode) SetEnabled(Enabled *bool) {
	v.Enabled = Enabled
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

func (o *AviEvhVsNode) DeleteSSLRefInEVHNode(sslKeyCertName, key string) {
	for i, sslKeyCertRefs := range o.SSLKeyCertRefs {
		if sslKeyCertRefs.Name == sslKeyCertName {
			o.SSLKeyCertRefs = append(o.SSLKeyCertRefs[:i], o.SSLKeyCertRefs[i+1:]...)
			utils.AviLog.Debugf("key: %s, msg: replaced SSLKeyCertRefs for evh in model: %s sslKeyCertRefs name: %s", key, o.Name, sslKeyCertRefs.Name)
			return
		}
	}
}

func (v *AviEvhVsNode) GetNodeType() string {
	return "VirtualServiceNode"
}

func (v *AviEvhVsNode) CalculateCheckSum() {
	portproto := v.PortProto
	sort.Slice(portproto, func(i, j int) bool {
		return portproto[i].Name < portproto[j].Name
	})

	var dsChecksum, httppolChecksum, evhChecksum, sslkeyChecksum, vsvipChecksum uint32

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
		utils.Hash(v.ServiceEngineGroup) +
		sslkeyChecksum +
		vsvipChecksum +
		utils.Hash(vsRefs) +
		utils.Hash(v.EvhHostName)

	if v.Enabled != nil {
		checksum += utils.Hash(utils.Stringify(v.Enabled))
	}
	checksum += lib.GetClusterLabelChecksum()

	if v.EnableRhi != nil {
		checksum += utils.Hash(utils.Stringify(*v.EnableRhi))
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

func (o *AviEvhVsNode) CheckHttpPolNameNChecksumForEvh(httpNodeName string, checksum uint32) bool {
	for _, http := range o.HttpPolicyRefs {
		if http.Name == httpNodeName {
			//Check if their checksums are same
			if http.GetCheckSum() == checksum {
				return false
			}
		}
	}
	return true
}

func (o *AviEvhVsNode) ReplaceHTTPRefInNodeForEvh(newHttpNode *AviHttpPolicySetNode, key string) {
	for i, http := range o.HttpPolicyRefs {
		if http.Name == newHttpNode.Name {
			o.HttpPolicyRefs = append(o.HttpPolicyRefs[:i], o.HttpPolicyRefs[i+1:]...)
			o.HttpPolicyRefs = append(o.HttpPolicyRefs, newHttpNode)
			utils.AviLog.Infof("key: %s, msg: replaced Evh http in model: %s Pool name: %s", key, o.Name, http.Name)
			return
		}
	}
	// If we have reached here it means we haven't found a match. Just append.
	o.HttpPolicyRefs = append(o.HttpPolicyRefs, newHttpNode)
	return
}

// Insecure ingress graph functions below

func (o *AviObjectGraph) ConstructAviL7SharedVsNodeForEvh(vsName, key string, routeIgrObj RouteIngressModel) *AviEvhVsNode {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	// This is a shared VS - always created in the admin namespace for now.
	avi_vs_meta := &AviEvhVsNode{Name: vsName, Tenant: lib.GetTenant(),
		SharedVS: true}
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
	avi_vs_meta.ApplicationProfile = utils.DEFAULT_L7_APP_PROFILE
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

	vsVipNode := &AviVSVIPNode{
		Name:       lib.GetVsVipName(vsName),
		Tenant:     lib.GetTenant(),
		FQDNs:      fqdns,
		EastWest:   false,
		VrfContext: vrfcontext,
	}

	if lib.GetSubnetIP() != "" {
		vsVipNode.SubnetIP = lib.GetSubnetIP()
		vsVipNode.SubnetPrefix = lib.GetSubnetPrefixInt()
	}

	if networkNames, err := lib.GetNetworkNamesForVsVipNode(); err != nil {
		utils.AviLog.Warnf("key: %s, msg: error when getting vipNetworkList: ", key, err)
		return nil
	} else {
		vsVipNode.NetworkNames = networkNames
	}

	if infraSetting := routeIgrObj.GetAviInfraSetting(); infraSetting != nil {
		buildWithInfraSettingForEvh(key, avi_vs_meta, vsVipNode, infraSetting)
	}

	avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)
	return avi_vs_meta
}

func (o *AviObjectGraph) BuildPolicyPGPoolsForEVH(vsNode []*AviEvhVsNode, childNode *AviEvhVsNode, namespace, ingName, key, host, infraSettingName string, paths []IngressHostPathSvc) {
	localPGList := make(map[string]*AviPoolGroupNode)

	// Update the VSVIP with the host information.
	if !utils.HasElem(vsNode[0].VSVIPRefs[0].FQDNs, host) {
		vsNode[0].VSVIPRefs[0].FQDNs = append(vsNode[0].VSVIPRefs[0].FQDNs, host)
	}
	if !utils.HasElem(childNode.VHDomainNames, host) {
		childNode.VHDomainNames = append(childNode.VHDomainNames, host)
	}
	var allFqdns []string
	allFqdns = append(allFqdns, host)
	for _, path := range paths {
		var httpPolicySet []AviHostPathPortPoolPG

		httpPGPath := AviHostPathPortPoolPG{Host: allFqdns}

		if path.PathType == networkingv1beta1.PathTypeExact {
			httpPGPath.MatchCriteria = "EQUALS"
		} else {
			// PathTypePrefix and PathTypeImplementationSpecific
			// default behaviour for AKO set be Prefix match on the path
			httpPGPath.MatchCriteria = "BEGINS_WITH"
		}

		if path.Path != "" {
			httpPGPath.Path = append(httpPGPath.Path, path.Path)
		}

		pgName := lib.GetEvhVsPoolNPgName(ingName, namespace, host, path.Path, infraSettingName)
		var pgNode *AviPoolGroupNode
		// There can be multiple services for the same path in case of alternate backend.
		// In that case, make sure we are creating only one PG per path
		pgNode, pgfound := localPGList[pgName]
		if !pgfound {
			pgNode = &AviPoolGroupNode{Name: pgName, Tenant: lib.GetTenant()}
			localPGList[pgName] = pgNode
			httpPGPath.PoolGroup = pgNode.Name
			httpPGPath.Host = allFqdns
			httpPolicySet = append(httpPolicySet, httpPGPath)
		}

		var poolName string
		poolName = lib.GetEvhVsPoolNPgName(ingName, namespace, host, path.Path, infraSettingName, path.ServiceName)
		poolNode := &AviPoolNode{
			Name:       poolName,
			PortName:   path.PortName,
			Tenant:     lib.GetTenant(),
			VrfContext: lib.GetVrf(),
		}

		serviceType := lib.GetServiceType()
		if serviceType == lib.NodePortLocal {
			if servers := PopulateServersForNPL(poolNode, namespace, path.ServiceName, true, key); servers != nil {
				poolNode.Servers = servers
			}
		} else if serviceType == lib.NodePort {
			if servers := PopulateServersForNodePort(poolNode, namespace, path.ServiceName, true, key); servers != nil {
				poolNode.Servers = servers
			}
		} else {
			if servers := PopulateServers(poolNode, namespace, path.ServiceName, true, key); servers != nil {
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
		if !pgfound {
			httppolname := lib.GetSniHttpPolName(ingName, namespace, host, path.Path, infraSettingName)
			policyNode := &AviHttpPolicySetNode{Name: httppolname, HppMap: httpPolicySet, Tenant: lib.GetTenant()}
			if childNode.CheckHttpPolNameNChecksumForEvh(httppolname, policyNode.GetCheckSum()) {
				childNode.ReplaceHTTPRefInNodeForEvh(policyNode, key)
			}
		}
	}
	for _, path := range paths {
		BuildPoolHTTPRule(host, path.Path, ingName, namespace, key, childNode, true)
	}

	utils.AviLog.Infof("key: %s, msg: added pools and poolgroups. childNodeChecksum for childNode :%s is :%v", key, childNode.Name, childNode.Name)

}

func ProcessInsecureHostsForEVH(routeIgrObj RouteIngressModel, key string, parsedIng IngressConfig, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost, hostsMap map[string]*objects.RouteIngrhost) {
	utils.AviLog.Debugf("key: %s, msg: Storedhosts before  processing insecurehosts: %s", key, utils.Stringify(Storedhosts))
	var infraSettingName string
	if aviInfraSetting := routeIgrObj.GetAviInfraSetting(); aviInfraSetting != nil {
		infraSettingName = aviInfraSetting.Name
	}

	for host, pathsvcmap := range parsedIng.IngressHostMap {
		// Remove this entry from storedHosts. First check if the host exists in the stored map or not.
		hostData, found := Storedhosts[host]
		if found && hostData.InsecurePolicy != lib.PolicyNone {
			// Verify the paths and take out the paths that are not need.
			pathSvcDiff := routeIgrObj.GetDiffPathSvc(hostData.PathSvc, pathsvcmap.ingressHPSvc, true)
			utils.AviLog.Debugf("key: %s, msg: pathSvcDiff %s", key, utils.Stringify(pathSvcDiff))
			if len(pathSvcDiff) == 0 {
				// Marking the entry as None to handle delete stale config
				utils.AviLog.Debugf("key: %s, msg: Marking the entry as None to handle delete stale config %s", key, utils.Stringify(pathSvcDiff))
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
		hostsMap[host].PathSvc = getPathSvc(pathsvcmap.ingressHPSvc)

		_, shardVsName := DeriveShardVSForEvh(host, key, routeIgrObj)
		modelName := lib.GetModelName(lib.GetTenant(), shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Infof("key: %s, msg: model not found, generating new model with name: %s", key, modelName)
			aviModel = NewAviObjectGraph()
			aviModel.(*AviObjectGraph).ConstructAviL7SharedVsNodeForEvh(shardVsName, key, routeIgrObj)
		}

		vsNode := aviModel.(*AviObjectGraph).GetAviEvhVS()
		if len(vsNode) > 0 && found {
			// if vsNode already exists, check for updates via AviInfraSetting
			if infraSetting := routeIgrObj.GetAviInfraSetting(); infraSetting != nil {
				buildWithInfraSettingForEvh(key, vsNode[0], vsNode[0].VSVIPRefs[0], infraSetting)
			}
		}

		// Create one evh child per host and associate http policies for each path.
		ingName := routeIgrObj.GetName()
		namespace := routeIgrObj.GetNamespace()
		evhNodeName := lib.GetEvhNodeName(ingName, namespace, host, infraSettingName)
		evhNode := vsNode[0].GetEvhNodeForName(evhNodeName)
		hostSlice := []string{host}

		// Populate the hostmap with empty secret for insecure ingress
		hostMap := HostNamePathSecrets{paths: getPaths(pathsvcmap.ingressHPSvc), secretName: ""}
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

		if evhNode == nil {
			evhNode = &AviEvhVsNode{
				Name:         evhNodeName,
				VHParentName: vsNode[0].Name,
				Tenant:       lib.GetTenant(),
				EVHParent:    false,
				EvhHostName:  host,
				ServiceMetadata: avicache.ServiceMetadataObj{
					NamespaceIngressName: ingressHostMap.GetIngressesForHostName(host),
					Namespace:            namespace,
					HostNames:            hostSlice,
				},
			}

			if lib.GetSEGName() != lib.DEFAULT_SE_GROUP {
				evhNode.ServiceEngineGroup = lib.GetSEGName()
			}
			evhNode.VrfContext = lib.GetVrf()

			foundEvhModel := FindAndReplaceEvhInModel(evhNode, vsNode, key)
			if !foundEvhModel {
				vsNode[0].EvhNodes = append(vsNode[0].EvhNodes, evhNode)
			}
		} else {
			// The evh node exists, just update the svc metadata
			evhNode.ServiceMetadata.NamespaceIngressName = ingressHostMap.GetIngressesForHostName(host)
			evhNode.ServiceMetadata.Namespace = namespace
			evhNode.ServiceMetadata.HostNames = hostSlice
		}

		evhNode.ApplicationProfile = utils.DEFAULT_L7_APP_PROFILE
		// Remove the redirect for secure to insecure transition
		RemoveRedirectHTTPPolicyInModelForEvh(evhNode, host, key)

		// build poolgroup and pool
		aviModel.(*AviObjectGraph).BuildPolicyPGPoolsForEVH(vsNode, evhNode, namespace, ingName, key, host, infraSettingName, pathsvcmap.ingressHPSvc)
		foundEvhModel := FindAndReplaceEvhInModel(evhNode, vsNode, key)
		if !foundEvhModel {
			vsNode[0].EvhNodes = append(vsNode[0].EvhNodes, evhNode)
		}
		// build host rule for insecure ingress in evh
		BuildL7HostRule(host, namespace, ingName, key, evhNode)
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

func (o *AviObjectGraph) BuildTlsCertNodeForEvh(svcLister *objects.SvcLister, tlsNode *AviEvhVsNode, namespace string, tlsData TlsSettings, key, infraSettingName, host string) bool {
	mClient := utils.GetInformers().ClientSet
	secretName := tlsData.SecretName
	secretNS := tlsData.SecretNS
	if secretNS == "" {
		secretNS = namespace
	}

	certNode := &AviTLSKeyCertNode{
		Name:   lib.GetTLSKeyCertNodeName(infraSettingName, host),
		Tenant: lib.GetTenant(),
		Type:   lib.CertTypeVS,
	}

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
	if tlsNode.CheckSSLCertNodeNameNChecksum(lib.GetTLSKeyCertNodeName(infraSettingName, host), certNode.GetCheckSum()) {
		tlsNode.ReplaceEvhSSLRefInEVHNode(certNode, key)
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
			if found && hostData.InsecurePolicy == lib.PolicyAllow {
				// this is transitioning from insecure to secure host
				Storedhosts[host].InsecurePolicy = lib.PolicyNone
			}
			if found && hostData.SecurePolicy == lib.PolicyEdgeTerm {
				// Verify the paths and take out the paths that are not need.
				pathSvcDiff := routeIgrObj.GetDiffPathSvc(hostData.PathSvc, newPathSvc, true)

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
	var infraSettingName string
	if aviInfraSetting := routeIgrObj.GetAviInfraSetting(); aviInfraSetting != nil {
		infraSettingName = aviInfraSetting.Name
	}

	for host, paths := range tlssetting.Hosts {
		var hosts []string
		hostPathSvcMap[host] = paths.ingressHPSvc
		hostMap := HostNamePathSecrets{paths: getPaths(paths.ingressHPSvc), secretName: tlssetting.SecretName}
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
		_, shardVsName := DeriveShardVSForEvh(host, key, routeIgrObj)
		// For each host, create a EVH node with the secret giving us the key and cert.
		// construct a EVH child VS node per tls setting which corresponds to one secret
		model_name := lib.GetModelName(lib.GetTenant(), shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(model_name)
		if !found || aviModel == nil {
			utils.AviLog.Infof("key: %s, msg: model not found, generating new model with name: %s", key, model_name)
			aviModel = NewAviObjectGraph()
			aviModel.(*AviObjectGraph).ConstructAviL7SharedVsNodeForEvh(shardVsName, key, routeIgrObj)
		}

		vsNode := aviModel.(*AviObjectGraph).GetAviEvhVS()
		if len(vsNode) < 1 {
			return nil
		}

		if found {
			// if vsNode already exists, check for updates via AviInfraSetting
			if infraSetting := routeIgrObj.GetAviInfraSetting(); infraSetting != nil {
				buildWithInfraSettingForEvh(key, vsNode[0], vsNode[0].VSVIPRefs[0], infraSetting)
			}
		}

		certsBuilt := false
		evhSecretName := tlssetting.SecretName
		re := regexp.MustCompile(fmt.Sprintf(`^%s.*`, lib.DummySecret))
		if re.MatchString(evhSecretName) {
			certsBuilt = true
		}

		evhNode := vsNode[0].GetEvhNodeForName(lib.GetEvhNodeName(ingName, namespace, host, infraSettingName))
		if evhNode == nil {
			evhNode = &AviEvhVsNode{
				Name:         lib.GetEvhNodeName(ingName, namespace, host, infraSettingName),
				VHParentName: vsNode[0].Name,
				Tenant:       lib.GetTenant(),
				EVHParent:    false,
				EvhHostName:  host,
				ServiceMetadata: avicache.ServiceMetadataObj{
					NamespaceIngressName: ingressHostMap.GetIngressesForHostName(host),
					Namespace:            namespace,
					HostNames:            hosts,
				},
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
		evhNode.ApplicationProfile = utils.DEFAULT_L7_APP_PROFILE
		evhNode.VrfContext = lib.GetVrf()
		if !certsBuilt {
			certsBuilt = aviModel.(*AviObjectGraph).BuildTlsCertNodeForEvh(routeIgrObj.GetSvcLister(), vsNode[0], namespace, tlssetting, key, infraSettingName, host)
		}
		if certsBuilt {
			aviModel.(*AviObjectGraph).BuildPolicyPGPoolsForEVH(vsNode, evhNode, namespace, ingName, key, host, infraSettingName, paths.ingressHPSvc)
			foundEvhModel := FindAndReplaceEvhInModel(evhNode, vsNode, key)
			if !foundEvhModel {
				vsNode[0].EvhNodes = append(vsNode[0].EvhNodes, evhNode)
			}

			RemoveRedirectHTTPPolicyInModelForEvh(evhNode, host, key)

			if tlssetting.redirect == true {
				aviModel.(*AviObjectGraph).BuildPolicyRedirectForVSForEvh(evhNode, host, namespace, ingName, key)
			}
			// Enable host rule
			BuildL7HostRule(host, namespace, ingName, key, evhNode)
		} else {
			hostMapOk, ingressHostMap := SharedHostNameLister().Get(host)
			if hostMapOk {
				// Replace the ingress map for this host.
				keyToRemove := namespace + "/" + ingName
				delete(ingressHostMap.HostNameMap, keyToRemove)
				SharedHostNameLister().Save(host, ingressHostMap)
			}
			// Since the cert couldn't be built, check if this EVH is affected by only in ingress if so remove the EVH node from the model
			if len(ingressHostMap.GetIngressesForHostName(host)) == 0 {
				vsNode[0].DeleteSSLRefInEVHNode(lib.GetTLSKeyCertNodeName(infraSettingName, host), key)
				RemoveEvhInModel(evhNode.Name, vsNode, key)
				RemoveRedirectHTTPPolicyInModelForEvh(evhNode, host, key)
			}

		}
		// Only add this node to the list of models if the checksum has changed.
		utils.AviLog.Debugf("key: %s, Saving Model: %v", key, utils.Stringify(vsNode))
		modelChanged := saveAviModel(model_name, aviModel.(*AviObjectGraph), key)
		if !utils.HasElem(*modelList, model_name) && modelChanged {
			*modelList = append(*modelList, model_name)
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

//DeleteStaleData : delete pool, EVH VS and redirect policy which are present in the object store but no longer required.
func DeleteStaleDataForEvh(routeIgrObj RouteIngressModel, key string, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost, hostsMap map[string]*objects.RouteIngrhost) {
	utils.AviLog.Debugf("key: %s, msg: About to delete stale data EVH Stored hosts: %v, hosts map: %v", key, utils.Stringify(Storedhosts), utils.Stringify(hostsMap))
	var infraSettingName string
	if aviInfraSetting := routeIgrObj.GetAviInfraSetting(); aviInfraSetting != nil {
		infraSettingName = aviInfraSetting.Name
	}

	for host, hostData := range Storedhosts {
		utils.AviLog.Debugf("host to del: %s, data : %s", host, utils.Stringify(hostData))
		_, shardVsName := DeriveShardVSForEvh(host, key, routeIgrObj)

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
		utils.AviLog.Debugf("key: %s, hostsMap: %s", key, utils.Stringify(hostsMap))
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
			aviModel.(*AviObjectGraph).DeletePoolForHostnameForEvh(shardVsName, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, removeFqdn, removeRedir, true)
		}
		if hostData.InsecurePolicy != lib.PolicyNone {
			aviModel.(*AviObjectGraph).DeletePoolForHostnameForEvh(shardVsName, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, removeFqdn, removeRedir, false)
		}
		changedModel := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if !utils.HasElem(modelList, modelName) && changedModel {
			*modelList = append(*modelList, modelName)
		}
	}
}

func DeriveShardVSForEvh(hostname, key string, routeIgrObj RouteIngressModel) (string, string) {
	// Read the value of the num_shards from the environment variable.
	utils.AviLog.Debugf("key: %s, msg: hostname for sharding: %s", key, hostname)

	// get stored infrasetting from ingress/route
	// figure out the current infrasetting via class/annotation
	oldSetting, newSetting := AviInfraSettingChange(routeIgrObj)
	var oldInfraPrefix, newInfraPrefix string

	oldShardSize, newShardSize := lib.GetshardSize(), lib.GetshardSize()
	if oldSetting != nil {
		if oldSetting.Spec.L7Settings != (akov1alpha1.AviInfraL7Settings{}) {
			oldShardSize = lib.ShardSizeMap[oldSetting.Spec.L7Settings.ShardSize]
		}
		oldInfraPrefix = oldSetting.Name
	}

	if !routeIgrObj.Exists() {
		// get the old ones.
		newShardSize = oldShardSize
		newInfraPrefix = oldInfraPrefix
	} else if newSetting != nil {
		if newSetting.Spec.L7Settings != (akov1alpha1.AviInfraL7Settings{}) {
			newShardSize = lib.ShardSizeMap[newSetting.Spec.L7Settings.ShardSize]
		}
		newInfraPrefix = newSetting.Name
	}

	shardVsPrefix := lib.GetNamePrefix() + lib.ShardVSPrefix + "-EVH-"
	oldVsName, newVsName := shardVsPrefix, shardVsPrefix
	if oldInfraPrefix != "" {
		oldVsName += "-" + oldInfraPrefix + "-"
	}
	if newInfraPrefix != "" {
		newVsName += "-" + newInfraPrefix + "-"
	}
	oldVsName += strconv.Itoa(int(utils.Bkt(hostname, oldShardSize)))
	newVsName += strconv.Itoa(int(utils.Bkt(hostname, newShardSize)))

	utils.AviLog.Infof("key: %s, msg: ShardVSNames: %s %s", key, oldVsName, newVsName)
	return oldVsName, newVsName
}

func (o *AviObjectGraph) RemovePoolNodeRefsFromEvh(poolName string, evhNode *AviEvhVsNode) {

	for i, pool := range evhNode.PoolRefs {
		if pool.Name == poolName {
			utils.AviLog.Debugf("Removing pool ref: %s", poolName)
			evhNode.PoolRefs = append(evhNode.PoolRefs[:i], evhNode.PoolRefs[i+1:]...)
			break
		}
	}
	utils.AviLog.Debugf("After removing the pool ref nodes are: %s", utils.Stringify(evhNode.PoolRefs))

}

func (o *AviObjectGraph) RemoveHTTPRefsFromEvh(httpPol string, evhNode *AviEvhVsNode) {

	for i, pol := range evhNode.HttpPolicyRefs {
		if pol.Name == httpPol {
			utils.AviLog.Debugf("Removing http pol ref: %s", httpPol)
			evhNode.HttpPolicyRefs = append(evhNode.HttpPolicyRefs[:i], evhNode.HttpPolicyRefs[i+1:]...)
			break
		}
	}
	utils.AviLog.Debugf("After removing the http policy nodes are: %s", utils.Stringify(evhNode.HttpPolicyRefs))

}

func (o *AviObjectGraph) RemovePGNodeRefsForEvh(pgName string, vsNode *AviEvhVsNode) {

	for i, pg := range vsNode.PoolGroupRefs {
		if pg.Name == pgName {
			utils.AviLog.Debugf("Removing pgRef: %s", pgName)
			vsNode.PoolGroupRefs = append(vsNode.PoolGroupRefs[:i], vsNode.PoolGroupRefs[i+1:]...)
			break
		}
	}
	utils.AviLog.Debugf("After removing the pg nodes are: %s", utils.Stringify(vsNode.PoolGroupRefs))

}

func (o *AviObjectGraph) ManipulateEvhNode(currentEvhNodeName, ingName, namespace, hostname string, pathSvc map[string][]string, vsNode []*AviEvhVsNode, infraSettingName, key string) bool {
	for _, modelEvhNode := range vsNode[0].EvhNodes {
		if currentEvhNodeName != modelEvhNode.Name {
			continue
		}

		for path, services := range pathSvc {
			pgName := lib.GetEvhVsPoolNPgName(ingName, namespace, hostname, path, infraSettingName)
			pgNode := modelEvhNode.GetPGForVSByName(pgName)
			for _, svc := range services {
				var evhPool string
				evhPool = lib.GetEvhVsPoolNPgName(ingName, namespace, hostname, path, infraSettingName, svc)
				o.RemovePoolNodeRefsFromEvh(evhPool, modelEvhNode)
				o.RemovePoolRefsFromPG(evhPool, pgNode)

				// Remove the EVH PG if it has no member
				if pgNode != nil {
					if len(pgNode.Members) == 0 {
						o.RemovePGNodeRefsForEvh(pgName, modelEvhNode)
						httppolname := lib.GetEvhVsPoolNPgName(ingName, namespace, hostname, path, infraSettingName)
						o.RemoveHTTPRefsFromEvh(httppolname, modelEvhNode)
					}
				}
			}
		}
		// After going through the paths, if the EVH node does not have any PGs - then delete it.
		if len(modelEvhNode.PoolGroupRefs) == 0 {
			RemoveEvhInModel(currentEvhNodeName, vsNode, key)
			// Remove the evhhost mapping
			SharedHostNameLister().Delete(hostname)
			return false
		}
	}

	return true
}

func (o *AviObjectGraph) GetAviPoolNodesByIngressForEvh(tenant string, ingName string) []*AviPoolNode {
	var aviPool []*AviPoolNode
	for _, model := range o.modelNodes {
		if model.GetNodeType() == "VirtualServiceNode" {
			for _, pool := range model.(*AviEvhVsNode).PoolRefs {
				if pool.IngressName == ingName && tenant == pool.ServiceMetadata.Namespace {
					utils.AviLog.Debugf("Found Pool with name: %s Adding...", pool.IngressName)
					aviPool = append(aviPool, pool)
				}
			}
		}
	}
	return aviPool
}

func (o *AviObjectGraph) DeletePoolForHostnameForEvh(vsName, hostname string, routeIgrObj RouteIngressModel, pathSvc map[string][]string, key, infraSettingName string, removeFqdn, removeRedir, secure bool) {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	namespace := routeIgrObj.GetNamespace()
	ingName := routeIgrObj.GetName()
	vsNode := o.GetAviEvhVS()
	keepEvh := false
	hostMapOk, ingressHostMap := SharedHostNameLister().Get(hostname)
	if hostMapOk {
		// Replace the ingress map for this host.
		keyToRemove := namespace + "/" + ingName
		delete(ingressHostMap.HostNameMap, keyToRemove)
		SharedHostNameLister().Save(hostname, ingressHostMap)
	}

	evhNodeName := lib.GetEvhNodeName(ingName, namespace, hostname, infraSettingName)
	utils.AviLog.Infof("key: %s, msg: EVH node to delete: %s", key, evhNodeName)
	keepEvh = o.ManipulateEvhNode(evhNodeName, ingName, namespace, hostname, pathSvc, vsNode, infraSettingName, key)
	if !keepEvh {
		// Delete the cert ref for the host
		vsNode[0].DeleteSSLRefInEVHNode(lib.GetTLSKeyCertNodeName(infraSettingName, hostname), key)
	}
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

func (o *AviObjectGraph) BuildPolicyRedirectForVSForEvh(vsNode *AviEvhVsNode, hostname string, namespace, ingName, key string) {
	policyname := lib.GetL7HttpRedirPolicy(vsNode.Name)
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

	if policyFound := FindAndReplaceRedirectHTTPPolicyInModelforEvh(vsNode, redirectPolicy, hostname, key); !policyFound {
		redirectPolicy.CalculateCheckSum()
		vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs, redirectPolicy)
	}

}

// RouteIngrDeletePoolsByHostname : Based on DeletePoolsByHostname, delete pools and policies that are no longer required
func RouteIngrDeletePoolsByHostnameForEvh(routeIgrObj RouteIngressModel, namespace, objname, key string, fullsync bool, sharedQueue *utils.WorkerQueue) {
	ok, hostMap := routeIgrObj.GetSvcLister().IngressMappings(namespace).GetRouteIngToHost(objname)
	if !ok {
		utils.AviLog.Warnf("key: %s, msg: nothing to delete for route: %s", key, objname)
		return
	}

	var infraSettingName string
	if aviInfraSetting := routeIgrObj.GetAviInfraSetting(); aviInfraSetting != nil {
		infraSettingName = aviInfraSetting.Name
	}

	utils.AviLog.Debugf("key: %s, msg: hosts to delete are :%s", key, utils.Stringify(hostMap))
	for host, hostData := range hostMap {
		_, shardVsName := DeriveShardVSForEvh(host, key, routeIgrObj)
		if hostData.SecurePolicy == lib.PolicyPass {
			shardVsName = lib.GetPassthroughShardVSName(host, key)
		}

		modelName := lib.GetModelName(lib.GetTenant(), shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Warnf("key: %s, msg: model not found during delete: %s", key, modelName)
			continue
		}

		// Delete the pool corresponding to this host
		if hostData.SecurePolicy == lib.PolicyEdgeTerm {
			aviModel.(*AviObjectGraph).DeletePoolForHostnameForEvh(shardVsName, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, true, true, true)
		} else if hostData.SecurePolicy == lib.PolicyPass {
			aviModel.(*AviObjectGraph).DeleteObjectsForPassthroughHost(shardVsName, host, routeIgrObj, hostData.PathSvc, key, true, true, true)
		}
		if hostData.InsecurePolicy == lib.PolicyAllow {
			aviModel.(*AviObjectGraph).DeletePoolForHostnameForEvh(shardVsName, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, true, true, false)
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
	updateHostPathCache(namespace, objname, hostMap, nil)
}

func DeleteStaleDataForModelChangeForEvh(routeIgrObj RouteIngressModel, namespace, objname, key string, fullsync bool, sharedQueue *utils.WorkerQueue) {
	ok, hostMap := routeIgrObj.GetSvcLister().IngressMappings(namespace).GetRouteIngToHost(objname)
	if !ok {
		utils.AviLog.Warnf("key: %s, msg: nothing to delete for route: %s", key, objname)
		return
	}

	utils.AviLog.Debugf("key: %s, msg: hosts to delete %s", key, utils.Stringify(hostMap))
	for host, hostData := range hostMap {
		shardVsName, newShardVsName := DeriveShardVSForEvh(host, key, routeIgrObj)
		if shardVsName == newShardVsName {
			continue
		}

		_, infraSettingName := objects.InfraSettingL7Lister().GetIngRouteToInfraSetting(routeIgrObj.GetNamespace() + "/" + routeIgrObj.GetName())
		modelName := lib.GetModelName(lib.GetTenant(), shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Warnf("key: %s, msg: model not found during delete: %s", key, modelName)
			continue
		}

		// Delete the pool corresponding to this host
		if hostData.SecurePolicy == lib.PolicyEdgeTerm {
			aviModel.(*AviObjectGraph).DeletePoolForHostnameForEvh(shardVsName, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, true, true, true)
		}
		if hostData.InsecurePolicy != lib.PolicyNone {
			aviModel.(*AviObjectGraph).DeletePoolForHostnameForEvh(shardVsName, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, true, true, false)
		}

		ok := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if ok && len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 && !fullsync {
			PublishKeyToRestLayer(modelName, key, sharedQueue)
		}
	}
}

func buildWithInfraSettingForEvh(key string, vs *AviEvhVsNode, vsvip *AviVSVIPNode, infraSetting *akov1alpha1.AviInfraSetting) {
	if infraSetting != nil && infraSetting.Status.Status == lib.StatusAccepted {
		if infraSetting.Spec.SeGroup.Name != "" {
			// This assumes that the SeGroup has the appropriate labels configured
			vs.ServiceEngineGroup = infraSetting.Spec.SeGroup.Name
		} else {
			vs.ServiceEngineGroup = lib.GetSEGName()
		}

		if infraSetting.Spec.Network.EnableRhi != nil {
			vs.EnableRhi = infraSetting.Spec.Network.EnableRhi
		} else {
			enableRhi := lib.GetEnableRHI()
			vs.EnableRhi = &enableRhi
		}

		if vsvip.NetworkNames != nil && len(vsvip.NetworkNames) > 0 {
			vsvip.NetworkNames = infraSetting.Spec.Network.Names
			vsvip.SubnetIP = ""
		} else {
			if networkNames, err := lib.GetNetworkNamesForVsVipNode(); err != nil {
				utils.AviLog.Warnf("key: %s, msg: error when getting vipNetworkList: ", key, err)
				return
			} else {
				vsvip.NetworkNames = networkNames
			}
		}
	}
	utils.AviLog.Debugf("key: %s, msg: Applied AviInfraSetting configuration over VSNode %s", key, vs.Name)
}
