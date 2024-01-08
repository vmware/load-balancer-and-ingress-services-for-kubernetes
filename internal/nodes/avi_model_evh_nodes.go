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
	"sort"
	"strconv"
	"strings"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/vmware/alb-sdk/go/models"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// AviVsEvhSniModel : High Level interfaces that should be implemented by
// AviEvhVsNode and  AviVsNode
type AviVsEvhSniModel interface {
	GetName() string
	SetName(string)

	IsSharedVS() bool
	IsDedicatedVS() bool
	IsSecure() bool

	GetPortProtocols() []AviPortHostProtocol
	SetPortProtocols([]AviPortHostProtocol)

	GetPoolRefs() []*AviPoolNode
	SetPoolRefs([]*AviPoolNode)

	GetPoolGroupRefs() []*AviPoolGroupNode
	SetPoolGroupRefs([]*AviPoolGroupNode)

	GetSSLKeyCertRefs() []*AviTLSKeyCertNode
	SetSSLKeyCertRefs([]*AviTLSKeyCertNode)

	GetHttpPolicyRefs() []*AviHttpPolicySetNode
	SetHttpPolicyRefs([]*AviHttpPolicySetNode)

	GetServiceMetadata() lib.ServiceMetadataObj
	SetServiceMetadata(lib.ServiceMetadataObj)

	GetSslKeyAndCertificateRefs() []string
	SetSslKeyAndCertificateRefs([]string)

	GetWafPolicyRef() *string
	SetWafPolicyRef(*string)

	GetHttpPolicySetRefs() []string
	SetHttpPolicySetRefs([]string)

	GetAppProfileRef() *string
	SetAppProfileRef(*string)

	GetICAPProfileRefs() []string
	SetICAPProfileRefs([]string)

	GetAnalyticsProfileRef() *string
	SetAnalyticsProfileRef(*string)

	GetErrorPageProfileRef() string
	SetErrorPageProfileRef(string)

	GetSSLProfileRef() *string
	SetSSLProfileRef(*string)

	GetVsDatascriptRefs() []string
	SetVsDatascriptRefs([]string)

	GetEnabled() *bool
	SetEnabled(*bool)

	GetAnalyticsPolicy() *avimodels.AnalyticsPolicy
	SetAnalyticsPolicy(*avimodels.AnalyticsPolicy)

	GetVSVIPLoadBalancerIP() string
	SetVSVIPLoadBalancerIP(string)

	GetVHDomainNames() []string
	SetVHDomainNames([]string)

	GetGeneratedFields() *AviVsNodeGeneratedFields
	GetCommonFields() *AviVsNodeCommonFields

	GetNetworkSecurityPolicyRef() *string
	SetNetworkSecurityPolicyRef(*string)
	GetTenant() string
}

type AviEvhVsNode struct {
	EVHParent     bool
	VHParentName  string
	VHDomainNames []string
	EvhNodes      []*AviEvhVsNode
	EvhHostName   string
	AviMarkers    utils.AviObjectMarkers
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
	ServiceMetadata     lib.ServiceMetadataObj
	VrfContext          string
	ICAPProfileRefs     []string
	ErrorPageProfileRef string
	HttpPolicySetRefs   []string
	Paths               []string
	IngressNames        []string
	Dedicated           bool
	VHMatches           []*avimodels.VHMatch
	Secure              bool
	Caller              string

	AviVsNodeCommonFields

	AviVsNodeGeneratedFields
}

// Implementing AviVsEvhSniModel

func (v *AviEvhVsNode) GetName() string {
	return v.Name
}

func (v *AviEvhVsNode) SetName(name string) {
	v.Name = name
}

func (v *AviEvhVsNode) IsSharedVS() bool {
	return v.SharedVS
}

func (v *AviEvhVsNode) IsDedicatedVS() bool {
	return v.Dedicated
}

func (v *AviEvhVsNode) IsSecure() bool {
	return v.Secure
}

func (v *AviEvhVsNode) GetPortProtocols() []AviPortHostProtocol {
	return v.PortProto
}

func (v *AviEvhVsNode) SetPortProtocols(portProto []AviPortHostProtocol) {
	v.PortProto = portProto
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

func (v *AviEvhVsNode) GetServiceMetadata() lib.ServiceMetadataObj {
	return v.ServiceMetadata
}

func (v *AviEvhVsNode) SetServiceMetadata(serviceMetadata lib.ServiceMetadataObj) {
	v.ServiceMetadata = serviceMetadata
}

func (v *AviEvhVsNode) GetSslKeyAndCertificateRefs() []string {
	return v.SslKeyAndCertificateRefs
}

func (v *AviEvhVsNode) SetSslKeyAndCertificateRefs(sslKeyAndCertificateRefs []string) {
	v.SslKeyAndCertificateRefs = sslKeyAndCertificateRefs
}

func (v *AviEvhVsNode) GetWafPolicyRef() *string {
	return v.WafPolicyRef
}

func (v *AviEvhVsNode) SetWafPolicyRef(wafPolicyRef *string) {
	v.WafPolicyRef = wafPolicyRef
}

func (v *AviEvhVsNode) GetHttpPolicySetRefs() []string {
	return v.HttpPolicySetRefs
}

func (v *AviEvhVsNode) SetHttpPolicySetRefs(httpPolicySetRefs []string) {
	v.HttpPolicySetRefs = httpPolicySetRefs
}

func (v *AviEvhVsNode) GetAppProfileRef() *string {
	return v.ApplicationProfileRef
}

func (v *AviEvhVsNode) SetAppProfileRef(applicationProfileRef *string) {
	v.ApplicationProfileRef = applicationProfileRef
}

func (v *AviEvhVsNode) GetICAPProfileRefs() []string {
	return v.ICAPProfileRefs
}

func (v *AviEvhVsNode) SetICAPProfileRefs(ICAPProfileRef []string) {
	v.ICAPProfileRefs = ICAPProfileRef
}

func (v *AviEvhVsNode) GetAnalyticsProfileRef() *string {
	return v.AnalyticsProfileRef
}

func (v *AviEvhVsNode) SetAnalyticsProfileRef(analyticsProfileRef *string) {
	v.AnalyticsProfileRef = analyticsProfileRef
}

func (v *AviEvhVsNode) GetErrorPageProfileRef() string {
	return v.ErrorPageProfileRef
}

func (v *AviEvhVsNode) SetErrorPageProfileRef(errorPageProfileRef string) {
	v.ErrorPageProfileRef = errorPageProfileRef
}

func (v *AviEvhVsNode) GetSSLProfileRef() *string {
	return v.SslProfileRef
}

func (v *AviEvhVsNode) SetSSLProfileRef(SSLProfileRef *string) {
	v.SslProfileRef = SSLProfileRef
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

func (v *AviEvhVsNode) GetAnalyticsPolicy() *avimodels.AnalyticsPolicy {
	return v.AnalyticsPolicy
}

func (v *AviEvhVsNode) SetAnalyticsPolicy(policy *avimodels.AnalyticsPolicy) {
	v.AnalyticsPolicy = policy
}

func (v *AviEvhVsNode) GetVSVIPLoadBalancerIP() string {
	if len(v.VSVIPRefs) > 0 {
		return v.VSVIPRefs[0].IPAddress
	}
	return ""
}

func (v *AviEvhVsNode) SetVSVIPLoadBalancerIP(ip string) {
	if len(v.VSVIPRefs) > 0 {
		v.VSVIPRefs[0].IPAddress = ip
	}
}

func (v *AviEvhVsNode) GetVHDomainNames() []string {
	return v.VHDomainNames
}

func (v *AviEvhVsNode) SetVHDomainNames(domainNames []string) {
	v.VHDomainNames = domainNames
}

func (v *AviEvhVsNode) GetGeneratedFields() *AviVsNodeGeneratedFields {
	return &v.AviVsNodeGeneratedFields
}

func (v *AviEvhVsNode) GetCommonFields() *AviVsNodeCommonFields {
	return &v.AviVsNodeCommonFields
}

func (v *AviEvhVsNode) GetNetworkSecurityPolicyRef() *string {
	return v.NetworkSecurityPolicyRef
}

func (v *AviEvhVsNode) SetNetworkSecurityPolicyRef(networkSecuirtyPolicyRef *string) {
	v.NetworkSecurityPolicyRef = networkSecuirtyPolicyRef
}

func (v *AviEvhVsNode) GetTenant() string {
	return v.Tenant
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
}

func (o *AviEvhVsNode) DeleteCACertRefInEVHNode(cacertNodeName, key string) {
	for i, cacert := range o.CACertRefs {
		if cacert.Name == cacertNodeName {
			o.CACertRefs = append(o.CACertRefs[:i], o.CACertRefs[i+1:]...)
			utils.AviLog.Debugf("key: %s, msg: deleted cacert for evh in model: %s Pool name: %s", key, o.Name, cacert.Name)
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

func (vsNode *AviEvhVsNode) AddSSLPort(key string) {
	for _, port := range vsNode.PortProto {
		if port.Port == lib.SSLPort {
			return
		}
	}
	httpsPort := AviPortHostProtocol{Port: 443, Protocol: utils.HTTP, EnableSSL: true}
	vsNode.PortProto = append(vsNode.PortProto, httpsPort)
}

// TODO: Next PR Opt: make part of Avivsevhsni model interface
func (vsNode *AviEvhVsNode) DeleteSSLPort(key string) {
	for i, port := range vsNode.PortProto {
		if port.Port == lib.SSLPort {
			vsNode.PortProto = append(vsNode.PortProto[:i], vsNode.PortProto[i+1:]...)
		}
	}
}

// TODO: Next PR opt: make part of Avivs model interface
func (vsNode *AviEvhVsNode) DeletSSLRefInDedicatedNode(key string) {
	vsNode.SSLKeyCertRefs = []*AviTLSKeyCertNode{}
	vsNode.SslProfileRef = nil
	vsNode.CACertRefs = []*AviTLSKeyCertNode{}
}

func (vsNode *AviEvhVsNode) DeleteSecureAppProfile(key string) {
	if vsNode.ApplicationProfile == utils.DEFAULT_L7_SECURE_APP_PROFILE {
		vsNode.ApplicationProfile = utils.DEFAULT_L7_APP_PROFILE
	}
}
func (v *AviEvhVsNode) GetNodeType() string {
	return "VirtualServiceNode"
}

func (o *AviEvhVsNode) AddFQDNAliasesToHTTPPolicy(hosts []string, key string) {

	// Find the hppMap and redirectPorts that matches the host
	for _, policy := range o.HttpPolicyRefs {
		for j := range policy.HppMap {
			policy.HppMap[j].Host = make([]string, len(hosts))
			copy(policy.HppMap[j].Host, hosts)
		}
		for j := range policy.RedirectPorts {
			policy.RedirectPorts[j].Hosts = make([]string, len(hosts))
			copy(policy.RedirectPorts[j].Hosts, hosts)
		}
		policy.AviMarkers.Host = make([]string, len(hosts))
		copy(policy.AviMarkers.Host, hosts)
	}

	utils.AviLog.Debugf("key: %s, msg: Added hosts %v to HTTP policy for VS %s", key, hosts, o.Name)
}

func (o *AviEvhVsNode) RemoveFQDNAliasesFromHTTPPolicy(hosts []string, key string) {

	for _, host := range hosts {
		// Find the hppMap and redirectPorts that matches the host and remove the hosts
		for _, policy := range o.HttpPolicyRefs {
			for j := range policy.HppMap {
				policy.HppMap[j].Host = utils.Remove(policy.HppMap[j].Host, host)
			}
			for j := range policy.RedirectPorts {
				policy.RedirectPorts[j].Hosts = utils.Remove(policy.RedirectPorts[j].Hosts, host)
			}
			policy.AviMarkers.Host = utils.Remove(policy.AviMarkers.Host, host)
		}
	}

	utils.AviLog.Debugf("key: %s, msg: Removed hosts %v from HTTP policy for VS %s", key, hosts, o.Name)
}

func (o *AviEvhVsNode) AddFQDNsToModel(hosts []string, gsFqdn, key string) {
	if len(o.VSVIPRefs) == 0 {
		return
	}
	for _, host := range hosts {
		if host != gsFqdn &&
			!utils.HasElem(o.VSVIPRefs[0].FQDNs, host) {
			o.VSVIPRefs[0].FQDNs = append(o.VSVIPRefs[0].FQDNs, host)
		}
	}
	utils.AviLog.Debugf("key: %s, msg: Added hosts %v to model for VS %s", key, hosts, o.Name)
}

func (o *AviEvhVsNode) RemoveFQDNsFromModel(hosts []string, key string) {
	if len(o.VSVIPRefs) == 0 {
		return
	}
	for i := 0; i < len(o.VSVIPRefs[0].FQDNs); i++ {
		for _, host := range hosts {
			if host == o.VSVIPRefs[0].FQDNs[i] {
				o.VSVIPRefs[0].FQDNs = append(o.VSVIPRefs[0].FQDNs[:i], o.VSVIPRefs[0].FQDNs[i+1:]...)
				i--
				break
			}
		}
	}
	utils.AviLog.Debugf("key: %s, msg: Removed hosts %v from VS %s", key, hosts, o.Name)
}

func (v *AviEvhVsNode) CalculateCheckSum() {
	portproto := v.PortProto
	sort.Slice(portproto, func(i, j int) bool {
		return portproto[i].Name < portproto[j].Name
	})

	var checksumStringSlice []string

	for _, ds := range v.HTTPDSrefs {
		checksumStringSlice = append(checksumStringSlice, "HTTPDS"+ds.Name)
	}

	for _, httppol := range v.HttpPolicyRefs {
		checksumStringSlice = append(checksumStringSlice, "HttpPolicy"+httppol.Name)
	}

	for _, cacert := range v.CACertRefs {
		checksumStringSlice = append(checksumStringSlice, "CACert"+cacert.Name)
	}

	for _, sslkeycert := range v.SSLKeyCertRefs {
		checksumStringSlice = append(checksumStringSlice, "SSLKeyCert"+sslkeycert.Name)
	}

	for _, vsvipref := range v.VSVIPRefs {
		checksumStringSlice = append(checksumStringSlice, "VSVIP"+vsvipref.Name)
	}
	for _, vhdomain := range v.VHDomainNames {
		checksumStringSlice = append(checksumStringSlice, "VHDomain"+vhdomain)
	}

	for _, evhnode := range v.EvhNodes {
		checksumStringSlice = append(checksumStringSlice, "EVHNode"+evhnode.Name)
		for _, evhcert := range evhnode.SslKeyAndCertificateRefs {
			checksumStringSlice = append(checksumStringSlice, "EVHNodeSSL"+evhcert)

		}
	}

	// Note: Changing the order of strings being appended, while computing vsRefs and checksum,
	// will change the eventual checksum Hash.

	// keep the order of these policies
	policies := v.HttpPolicySetRefs
	scripts := v.VsDatascriptRefs
	icaprefs := v.ICAPProfileRefs
	sslKeyAndCertificateRefs := v.SslKeyAndCertificateRefs

	var vsRefs string

	if v.WafPolicyRef != nil {
		vsRefs += *v.WafPolicyRef
	}

	if v.ApplicationProfileRef != nil {
		vsRefs += *v.ApplicationProfileRef
	}

	if v.AnalyticsProfileRef != nil {
		vsRefs += *v.AnalyticsProfileRef
	}

	vsRefs += v.ErrorPageProfileRef

	if v.SslProfileRef != nil {
		vsRefs += *v.SslProfileRef
	}

	if len(scripts) > 0 {
		vsRefs += utils.Stringify(scripts)
	}

	if len(policies) > 0 {
		vsRefs += utils.Stringify(policies)
	}

	if len(icaprefs) > 0 {
		vsRefs += utils.Stringify(icaprefs)
	}

	if len(sslKeyAndCertificateRefs) > 0 {
		vsRefs += utils.Stringify(sslKeyAndCertificateRefs)
	}

	sort.Strings(checksumStringSlice)
	checksum := utils.Hash(strings.Join(checksumStringSlice, delim) +
		v.ApplicationProfile +
		v.ServiceEngineGroup +
		v.NetworkProfile +
		utils.Stringify(portproto) +
		v.EvhHostName)

	if vsRefs != "" {
		checksum += utils.Hash(vsRefs)
	}

	if v.Enabled != nil {
		checksum += utils.Hash(utils.Stringify(v.Enabled))
	}

	checksum += lib.GetMarkersChecksum(v.AviMarkers)

	if v.EnableRhi != nil {
		checksum += utils.Hash(utils.Stringify(*v.EnableRhi))
	}

	if v.AnalyticsPolicy != nil {
		checksum += lib.GetAnalyticsPolicyChecksum(v.AnalyticsPolicy)
	}

	checksum += v.AviVsNodeGeneratedFields.CalculateCheckSumOfGeneratedCode()

	if v.VHMatches != nil {
		checksum += utils.Hash(utils.Stringify(v.VHMatches))
	}

	if v.DefaultPoolGroup != "" {
		checksum += utils.Hash(v.DefaultPoolGroup)
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

func (o *AviEvhVsNode) CheckHttpPolNameNChecksumForEvh(httpNodeName, hppMapName string, checksum uint32) bool {
	for i, http := range o.HttpPolicyRefs {
		if http.Name == httpNodeName {
			for _, hppMap := range o.HttpPolicyRefs[i].HppMap {
				if hppMap.Name == hppMapName {
					if http.GetCheckSum() == checksum {
						return false
					} else {
						return true
					}
				}
			}
		}
	}
	return true
}

func (o *AviEvhVsNode) ReplaceHTTPRefInNodeForEvh(httpPGPath AviHostPathPortPoolPG, httpPolName, key string) {
	for i, http := range o.HttpPolicyRefs {
		if http.Name == httpPolName {
			for j, hppMap := range o.HttpPolicyRefs[i].HppMap {
				if hppMap.Name == httpPGPath.Name {
					o.HttpPolicyRefs[i].HppMap = append(o.HttpPolicyRefs[i].HppMap[:j], o.HttpPolicyRefs[i].HppMap[j+1:]...)
					o.HttpPolicyRefs[i].HppMap = append(o.HttpPolicyRefs[i].HppMap, httpPGPath)

					utils.AviLog.Infof("key: %s, msg: replaced Evh httpmap in model: %s Pool name: %s", key, o.Name, hppMap.Name)
					return
				}
			}
			// If we have reached here it means we haven't found a match. Just append.
			o.HttpPolicyRefs[i].HppMap = append(o.HttpPolicyRefs[i].HppMap, httpPGPath)
		}
	}
}

// Insecure ingress graph functions below

func (o *AviObjectGraph) ConstructAviL7SharedVsNodeForEvh(vsName, key string, routeIgrObj RouteIngressModel, dedicated, secure bool) {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	infraSetting := routeIgrObj.GetAviInfraSetting()
	tenant := lib.GetTenant()
	if infraSetting != nil && infraSetting.Spec.NSXSettings.Project != nil {
		tenant = *infraSetting.Spec.NSXSettings.Project
	}
	// This is a shared VS - always created in the admin namespace for now.
	// Default case
	avi_vs_meta := &AviEvhVsNode{
		Name:               vsName,
		Tenant:             tenant,
		ServiceEngineGroup: lib.GetSEGName(),
		PortProto: []AviPortHostProtocol{
			{Port: 80, Protocol: utils.HTTP},
		},
		ApplicationProfile: utils.DEFAULT_L7_APP_PROFILE,
		NetworkProfile:     utils.DEFAULT_TCP_NW_PROFILE,
	}

	avi_vs_meta.Secure = secure

	if !dedicated || secure {
		httpsPort := AviPortHostProtocol{Port: 443, Protocol: utils.HTTP, EnableSSL: true}
		avi_vs_meta.PortProto = append(avi_vs_meta.PortProto, httpsPort)
	}
	if !dedicated {
		avi_vs_meta.SharedVS = true
		avi_vs_meta.EVHParent = true
	} else {
		avi_vs_meta.Dedicated = true
		if secure {
			avi_vs_meta.ApplicationProfile = utils.DEFAULT_L7_SECURE_APP_PROFILE
		}
	}

	var vrfcontext string
	t1lr := lib.GetT1LRPath()
	if infraSetting != nil && infraSetting.Spec.NSXSettings.T1LR != nil {
		t1lr = *infraSetting.Spec.NSXSettings.T1LR
	}
	if t1lr == "" {
		vrfcontext = lib.GetVrf()
		avi_vs_meta.VrfContext = vrfcontext
	}
	o.AddModelNode(avi_vs_meta)

	shardSize := lib.GetShardSizeFromAviInfraSetting(routeIgrObj.GetAviInfraSetting())
	subDomains := GetDefaultSubDomain()
	fqdns, fqdn := lib.GetFqdns(vsName, key, tenant, subDomains, shardSize)
	configuredSharedVSFqdn := fqdn

	vsVipNode := &AviVSVIPNode{
		Name:        lib.GetVsVipName(vsName),
		Tenant:      tenant,
		FQDNs:       fqdns,
		VrfContext:  vrfcontext,
		VipNetworks: utils.GetVipNetworkList(),
	}

	if t1lr != "" {
		vsVipNode.T1Lr = t1lr
	}

	if avi_vs_meta.EnableRhi != nil && *avi_vs_meta.EnableRhi {
		vsVipNode.BGPPeerLabels = lib.GetGlobalBgpPeerLabels()
	}

	buildWithInfraSettingForEvh(key, routeIgrObj.GetNamespace(), avi_vs_meta, vsVipNode, infraSetting)

	avi_vs_meta.VSVIPRefs = append(avi_vs_meta.VSVIPRefs, vsVipNode)

	if avi_vs_meta.SharedVS && configuredSharedVSFqdn != "" {
		BuildL7HostRule(configuredSharedVSFqdn, key, avi_vs_meta)
	}
}

func (o *AviObjectGraph) BuildPolicyPGPoolsForEVH(vsNode []*AviEvhVsNode, childNode *AviEvhVsNode, namespace, ingName, key string, infraSetting *akov1beta1.AviInfraSetting, hosts []string, paths []IngressHostPathSvc, tlsSettings *TlsSettings, modelType string) {
	localPGList := make(map[string]*AviPoolGroupNode)
	var httppolname string
	var policyNode *AviHttpPolicySetNode
	pathSet := sets.NewString(childNode.Paths...)

	var infraSettingName string
	if infraSetting != nil && !lib.IsInfraSettingNSScoped(infraSetting.Name, namespace) {
		infraSettingName = infraSetting.Name
	}

	ingressNameSet := sets.NewString(childNode.IngressNames...)
	ingressNameSet.Insert(ingName)
	// Update the VSVIP with the host information.
	if !utils.HasElem(vsNode[0].VSVIPRefs[0].FQDNs, hosts[0]) {
		vsNode[0].VSVIPRefs[0].FQDNs = append(vsNode[0].VSVIPRefs[0].FQDNs, hosts[0])
	}

	childNode.VHDomainNames = hosts
	httppolname = lib.GetSniHttpPolName(namespace, hosts[0], infraSettingName)

	for i, http := range childNode.HttpPolicyRefs {
		if http.Name == httppolname {
			policyNode = childNode.HttpPolicyRefs[i]
		}
	}
	if policyNode == nil {
		policyNode = &AviHttpPolicySetNode{Name: httppolname, Tenant: vsNode[0].Tenant}
		childNode.HttpPolicyRefs = append(childNode.HttpPolicyRefs, policyNode)
	}

	var allFqdns []string
	allFqdns = append(allFqdns, hosts...)
	for _, path := range paths {
		httpPGPath := AviHostPathPortPoolPG{Host: allFqdns}

		if path.PathType == networkingv1.PathTypeExact {
			httpPGPath.MatchCriteria = "EQUALS"
		} else {
			// PathTypePrefix and PathTypeImplementationSpecific
			// default behaviour for AKO set be Prefix match on the path
			httpPGPath.MatchCriteria = "BEGINS_WITH"
		}

		if path.Path != "" {
			httpPGPath.Path = append(httpPGPath.Path, path.Path)
		}

		pgName := lib.GetEvhPGName(ingName, namespace, hosts[0], path.Path, infraSettingName, vsNode[0].Dedicated)
		var pgNode *AviPoolGroupNode
		// There can be multiple services for the same path in case of alternate backend.
		// In that case, make sure we are creating only one PG per path
		pgNode, pgfound := localPGList[pgName]
		if !pgfound {
			pgNode = &AviPoolGroupNode{Name: pgName, Tenant: vsNode[0].Tenant}
			localPGList[pgName] = pgNode
			httpPGPath.PoolGroup = pgNode.Name
			httpPGPath.Host = allFqdns
		}
		pgNode.AviMarkers = lib.PopulatePGNodeMarkers(namespace, hosts[0], infraSettingName, []string{ingName}, []string{path.Path})
		poolName := lib.GetEvhPoolName(ingName, namespace, hosts[0], path.Path, infraSettingName, path.ServiceName, vsNode[0].Dedicated)
		hostslice := []string{hosts[0]}
		poolNode := &AviPoolNode{
			Name:       poolName,
			PortName:   path.PortName,
			Tenant:     vsNode[0].Tenant,
			VrfContext: lib.GetVrf(),
			Port:       path.Port,
			TargetPort: path.TargetPort,
			ServiceMetadata: lib.ServiceMetadataObj{
				IngressName: ingName,
				Namespace:   namespace,
				HostNames:   hostslice,
				PoolRatio:   path.weight,
			},
		}

		poolNode.NetworkPlacementSettings = lib.GetNodeNetworkMap()
		t1lr := lib.GetT1LRPath()
		if infraSetting != nil && infraSetting.Spec.NSXSettings.T1LR != nil {
			t1lr = *infraSetting.Spec.NSXSettings.T1LR
		}
		if t1lr != "" {
			poolNode.T1Lr = t1lr
			// Unset the poolnode's vrfcontext.
			poolNode.VrfContext = ""
		}

		poolNode.AviMarkers = lib.PopulatePoolNodeMarkers(namespace, hosts[0],
			infraSettingName, path.ServiceName, []string{ingName}, []string{path.Path})
		if tlsSettings != nil && tlsSettings.reencrypt {
			o.BuildPoolSecurity(poolNode, *tlsSettings, key, poolNode.AviMarkers)
		}

		serviceType := lib.GetServiceType()
		if serviceType == lib.NodePortLocal {
			if servers := PopulateServersForNPL(poolNode, namespace, path.ServiceName, true, key); servers != nil {
				poolNode.Servers = servers
			}
		} else if modelType == lib.MultiClusterIngress {
			if serviceType == lib.NodePort {
				poolNode.ServiceMetadata.IsMCIIngress = true
				// incase of multi-cluster ingress, the servers are created using service import CRD
				if servers := PopulateServersForMultiClusterIngress(poolNode, namespace, path.clusterContext, path.svcNamespace, path.ServiceName, key); servers != nil {
					poolNode.Servers = servers
				}
			} else {
				utils.AviLog.Errorf("key: %s, msg: Multi-cluster ingress is only supported for serviceType NodePort, not adding the servers", key)
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

		buildPoolWithInfraSetting(key, poolNode, infraSetting)
		if lib.IsIstioEnabled() {
			poolNode.UpdatePoolNodeForIstio()
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

		if !pgfound {
			pathSet.Insert(path.Path)
			hppMapName := lib.GetSniHppMapName(ingName, namespace, hosts[0], path.Path, infraSettingName, vsNode[0].Dedicated)
			httpPGPath.Name = hppMapName
			httpPGPath.IngName = ingName
			httpPGPath.CalculateCheckSum()
			policyNode.AviMarkers = lib.PopulateHTTPPolicysetNodeMarkers(namespace, hosts[0], infraSettingName, ingressNameSet.List(), pathSet.List())
			if childNode.CheckHttpPolNameNChecksumForEvh(httppolname, hppMapName, httpPGPath.Checksum) {
				childNode.ReplaceHTTPRefInNodeForEvh(httpPGPath, httppolname, key)
			}
		}
	}
	childNode.Paths = pathSet.List()
	childNode.IngressNames = ingressNameSet.List()
	for _, path := range paths {
		BuildPoolHTTPRule(hosts[0], path.Path, ingName, namespace, infraSettingName, key, childNode, true, vsNode[0].Dedicated)
	}

	utils.AviLog.Infof("key: %s, msg: added pools and poolgroups. childNodeChecksum for childNode :%s is :%v", key, childNode.Name, childNode.GetCheckSum())

}

func ProcessInsecureHostsForEVH(routeIgrObj RouteIngressModel, key string, parsedIng IngressConfig, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost, hostsMap map[string]*objects.RouteIngrhost) {
	utils.AviLog.Debugf("key: %s, msg: Storedhosts before  processing insecurehosts: %s", key, utils.Stringify(Storedhosts))
	infraSetting := routeIgrObj.GetAviInfraSetting()
	tenant := lib.GetTenant()
	if infraSetting != nil && infraSetting.Spec.NSXSettings.Project != nil {
		tenant = *infraSetting.Spec.NSXSettings.Project
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
		modelName := lib.GetModelName(tenant, shardVsName.Name)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Infof("key: %s, msg: model not found, generating new model with name: %s", key, modelName)
			aviModel = NewAviObjectGraph()
			aviModel.(*AviObjectGraph).ConstructAviL7SharedVsNodeForEvh(shardVsName.Name, key, routeIgrObj, shardVsName.Dedicated, false)
		}
		vsNode := aviModel.(*AviObjectGraph).GetAviEvhVS()
		infraSetting := routeIgrObj.GetAviInfraSetting()

		// Create one evh child per host and associate http policies for each path.
		modelGraph := aviModel.(*AviObjectGraph)
		modelGraph.BuildModelGraphForInsecureEVH(routeIgrObj, host, infraSetting, key, pathsvcmap)

		if len(vsNode) > 0 && found {
			// if vsNode already exists, check for updates via AviInfraSetting
			if infraSetting != nil {
				buildWithInfraSettingForEvh(key, routeIgrObj.GetNamespace(), vsNode[0], vsNode[0].VSVIPRefs[0], infraSetting)
			}
		}
		changedModel := saveAviModel(modelName, modelGraph, key)
		if !utils.HasElem(modelList, modelName) && changedModel {
			*modelList = append(*modelList, modelName)
		}
	}

	utils.AviLog.Debugf("key: %s, msg: Storedhosts after processing insecurehosts: %s", key, utils.Stringify(Storedhosts))
}

func (o *AviObjectGraph) BuildModelGraphForInsecureEVH(routeIgrObj RouteIngressModel, host string, infraSetting *akov1beta1.AviInfraSetting, key string, pathsvcmap HostMetadata) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	var evhNode *AviEvhVsNode
	vsNode := o.GetAviEvhVS()
	ingName := routeIgrObj.GetName()
	namespace := routeIgrObj.GetNamespace()
	isDedicated := vsNode[0].Dedicated
	var infraSettingName string
	if infraSetting != nil && !lib.IsInfraSettingNSScoped(infraSetting.Name, namespace) {
		infraSettingName = infraSetting.Name
	}

	hostSlice := []string{host}
	// Populate the hostmap with empty secret for insecure ingress
	PopulateIngHostMap(namespace, host, ingName, "", pathsvcmap)
	_, ingressHostMap := SharedHostNameLister().Get(host)

	if lib.VIPPerNamespace() {
		SharedHostNameLister().SaveNamespace(host, routeIgrObj.GetNamespace())
	}
	if !isDedicated {
		evhNodeName := lib.GetEvhNodeName(host, infraSettingName)
		evhNode = vsNode[0].GetEvhNodeForName(evhNodeName)
		if evhNode == nil {
			evhNode = &AviEvhVsNode{
				Name:         evhNodeName,
				VHParentName: vsNode[0].Name,
				Tenant:       vsNode[0].Tenant,
				EVHParent:    false,
				EvhHostName:  host,
				ServiceMetadata: lib.ServiceMetadataObj{
					NamespaceIngressName: ingressHostMap.GetIngressesForHostName(host),
					Namespace:            namespace,
					HostNames:            hostSlice,
				},
			}
		} else {
			// The evh node exists, just update the svc metadata
			evhNode.ServiceMetadata.NamespaceIngressName = ingressHostMap.GetIngressesForHostName(host)
			evhNode.ServiceMetadata.Namespace = namespace
			evhNode.ServiceMetadata.HostNames = hostSlice
		}
		evhNode.ServiceEngineGroup = lib.GetSEGName()
		evhNode.VrfContext = lib.GetVrf()
		evhNode.ApplicationProfile = utils.DEFAULT_L7_APP_PROFILE
		evhNode.AviMarkers = lib.PopulateVSNodeMarkers(namespace, host, infraSettingName)
	} else {
		vsNode[0].ServiceMetadata.NamespaceIngressName = ingressHostMap.GetIngressesForHostName(host)
		vsNode[0].ServiceMetadata.Namespace = namespace
		vsNode[0].ServiceMetadata.HostNames = hostSlice
		vsNode[0].AviMarkers = lib.PopulateVSNodeMarkers(namespace, host, infraSettingName)
	}

	hosts := []string{host}
	found, gsFqdnCache := objects.SharedCRDLister().GetLocalFqdnToGSFQDNMapping(host)
	if pathsvcmap.gslbHostHeader != "" {
		if !utils.HasElem(hosts, pathsvcmap.gslbHostHeader) {
			hosts = append(hosts, pathsvcmap.gslbHostHeader)
		}
		if vsNode[0].Dedicated {
			RemoveFqdnFromEVHVIP(vsNode[0], []string{gsFqdnCache}, key)
		}
		objects.SharedCRDLister().UpdateLocalFQDNToGSFqdnMapping(host, pathsvcmap.gslbHostHeader)
	} else {
		if found {
			objects.SharedCRDLister().DeleteLocalFqdnToGsFqdnMap(host)
			if vsNode[0].Dedicated {
				RemoveFqdnFromEVHVIP(vsNode[0], []string{gsFqdnCache}, key)
			}
		}
	}

	if isDedicated {
		evhNode = vsNode[0]
		evhNode.DeletSSLRefInDedicatedNode(key)
		evhNode.DeleteSSLPort(key)
		evhNode.Secure = false
		evhNode.DeleteSecureAppProfile(key)
	} else {
		vsNode[0].DeleteSSLRefInEVHNode(lib.GetTLSKeyCertNodeName(infraSettingName, host, ""), key)
	}
	// Remove the redirect for secure to insecure transition
	RemoveRedirectHTTPPolicyInModelForEvh(evhNode, hosts, key)
	// build poolgroup and pool
	o.BuildPolicyPGPoolsForEVH(vsNode, evhNode, namespace, ingName, key, infraSetting, hosts, pathsvcmap.ingressHPSvc, nil, routeIgrObj.GetType())
	if !isDedicated {
		foundEvhModel := FindAndReplaceEvhInModel(evhNode, vsNode, key)
		if !foundEvhModel {
			vsNode[0].EvhNodes = append(vsNode[0].EvhNodes, evhNode)
		}
	}
	// build host rule for insecure ingress in evh
	BuildL7HostRule(host, key, evhNode)
	// build SSORule for insecure ingress in evh
	BuildL7SSORule(host, key, evhNode)
	if !isDedicated {
		manipulateEvhNodeForSSL(key, vsNode[0], evhNode)
	}

	// Remove the deleted aliases from the FQDN list
	var hostsToRemove []string
	_, oldFQDNAliases := objects.SharedCRDLister().GetFQDNToAliasesMapping(host)
	for _, host := range oldFQDNAliases {
		if !utils.HasElem(evhNode.VHDomainNames, host) {
			hostsToRemove = append(hostsToRemove, host)
		}
	}
	vsNode[0].RemoveFQDNsFromModel(hostsToRemove, key)
	evhNode.RemoveFQDNAliasesFromHTTPPolicy(hostsToRemove, key)

	// Add FQDN aliases in the hostrule CRD to parent and child VSes
	vsNode[0].AddFQDNsToModel(evhNode.VHDomainNames, pathsvcmap.gslbHostHeader, key)
	evhNode.AddFQDNAliasesToHTTPPolicy(evhNode.VHDomainNames, key)
	evhNode.AviMarkers.Host = evhNode.VHDomainNames
	objects.SharedCRDLister().UpdateFQDNToAliasesMappings(host, evhNode.VHDomainNames)
}

// secure ingress graph functions

// BuildCACertNode : Build a new node to store CA cert, this would be referred by the corresponding keycert
func (o *AviObjectGraph) BuildCACertNodeForEvh(tlsNode *AviEvhVsNode, cacert, infraSettingName, host, key string) string {
	cacertNode := &AviTLSKeyCertNode{Name: lib.GetCACertNodeName(infraSettingName, host), Tenant: tlsNode.Tenant}
	cacertNode.Type = lib.CertTypeCA
	cacertNode.Cert = []byte(cacert)
	cacertNode.AviMarkers = lib.PopulateTLSKeyCertNode(host, infraSettingName)
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

	if lib.IsSecretK8sSecretRef(secretName) {
		secretName = strings.Split(secretName, "/")[2]
	}
	var altCertNode *AviTLSKeyCertNode
	var certNode *AviTLSKeyCertNode

	//for default cert, use existing node if it exists
	foundTLSKeyCertNode := false
	if tlsData.SecretName == lib.GetDefaultSecretForRoutes() {
		for _, ssl := range tlsNode.SSLKeyCertRefs {
			if ssl.Name == lib.GetTLSKeyCertNodeName(infraSettingName, host, tlsData.SecretName) {
				certNode = ssl
				foundTLSKeyCertNode = true
				break
			}
		}
		if foundTLSKeyCertNode {
			keyCertRefsSet := sets.NewString(certNode.AviMarkers.Host...)
			keyCertRefsSet.Insert(host)
			certNode.AviMarkers.Host = keyCertRefsSet.List()
		}
	}
	if !foundTLSKeyCertNode {
		certNode = &AviTLSKeyCertNode{
			Name:   lib.GetTLSKeyCertNodeName(infraSettingName, host, tlsData.SecretName),
			Tenant: tlsNode.Tenant,
			Type:   lib.CertTypeVS,
		}
		certNode.AviMarkers = lib.PopulateTLSKeyCertNode(host, infraSettingName)
	}

	// Openshift Routes do not refer to a secret, instead key/cert values are mentioned in the route.
	// Routes can refer to secrets only in case of using default secret in ako NS or using hostrule secret.
	if strings.HasPrefix(secretName, lib.RouteSecretsPrefix) {
		if tlsData.cert != "" && tlsData.key != "" {
			certNode.Cert = []byte(tlsData.cert)
			certNode.Key = []byte(tlsData.key)
			if tlsData.cacert != "" {
				certNode.CACert = o.BuildCACertNodeForEvh(tlsNode, tlsData.cacert, infraSettingName, host, key)
			} else {
				tlsNode.DeleteCACertRefInEVHNode(lib.GetCACertNodeName(infraSettingName, host), key)
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
			utils.AviLog.Warnf("key: %s, msg: secret: %s has been deleted, err: %s", key, secretName, err)
			return false
		}
		keycertMap := secretObj.Data
		cert, ok := keycertMap[utils.K8S_TLS_SECRET_CERT]
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
		altCert, ok := keycertMap[utils.K8S_TLS_SECRET_ALT_CERT]
		if ok {
			altKey, ok := keycertMap[utils.K8S_TLS_SECRET_ALT_KEY]
			if ok {
				foundTLSKeyCertNode := false
				for _, ssl := range tlsNode.SSLKeyCertRefs {
					if ssl.Name == lib.GetTLSKeyCertNodeName(infraSettingName, host, tlsData.SecretName+"-alt") {
						altCertNode = ssl
						altCertNode.AviMarkers = certNode.AviMarkers
						foundTLSKeyCertNode = true
						break
					}
				}
				if !foundTLSKeyCertNode {
					altCertNode = &AviTLSKeyCertNode{
						Name:       lib.GetTLSKeyCertNodeName(infraSettingName, host, tlsData.SecretName+"-alt"),
						Tenant:     tlsNode.Tenant,
						Type:       lib.CertTypeVS,
						AviMarkers: certNode.AviMarkers,
						Cert:       altCert,
						Key:        altKey,
					}
				}
			}
		}
		utils.AviLog.Infof("key: %s, msg: Added the secret object to tlsnode: %s", key, secretObj.Name)
	}
	// If this SSLCertRef is already present don't add it.
	if tlsNode.CheckSSLCertNodeNameNChecksum(lib.GetTLSKeyCertNodeName(infraSettingName, host, tlsData.SecretName), certNode.GetCheckSum()) {
		tlsNode.ReplaceEvhSSLRefInEVHNode(certNode, key)

	}
	if altCertNode != nil && tlsNode.CheckSSLCertNodeNameNChecksum(lib.GetTLSKeyCertNodeName(infraSettingName, host, tlsData.SecretName+"-alt"), altCertNode.GetCheckSum()) {
		tlsNode.ReplaceEvhSSLRefInEVHNode(altCertNode, key)

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
			if tlssetting.redirect {
				hostsMap[host].InsecurePolicy = lib.PolicyRedirect
			}
			hostsMap[host].PathSvc = getPathSvc(newPathSvc)
		}
	}
	utils.AviLog.Debugf("key: %s, msg: Storedhosts after processing securehosts: %s", key, utils.Stringify(Storedhosts))
}

func evhNodeHostName(routeIgrObj RouteIngressModel, tlssetting TlsSettings, ingName, namespace, key string, fullsync bool, sharedQueue *utils.WorkerQueue, modelList *[]string) map[string][]IngressHostPathSvc {
	hostPathSvcMap := make(map[string][]IngressHostPathSvc)
	infraSetting := routeIgrObj.GetAviInfraSetting()
	tenant := lib.GetTenant()
	if infraSetting != nil && infraSetting.Spec.NSXSettings.Project != nil {
		tenant = *infraSetting.Spec.NSXSettings.Project
	}

	for host, paths := range tlssetting.Hosts {
		var hosts []string
		hostPathSvcMap[host] = paths.ingressHPSvc

		PopulateIngHostMap(namespace, host, ingName, tlssetting.SecretName, paths)
		_, ingressHostMap := SharedHostNameLister().Get(host)

		if lib.VIPPerNamespace() {
			SharedHostNameLister().SaveNamespace(host, routeIgrObj.GetNamespace())
		}
		hosts = append(hosts, host)
		_, shardVsName := DeriveShardVSForEvh(host, key, routeIgrObj)
		// For each host, create a EVH node with the secret giving us the key and cert.
		// construct a EVH child VS node per tls setting which corresponds to one secret
		model_name := lib.GetModelName(tenant, shardVsName.Name)
		found, aviModel := objects.SharedAviGraphLister().Get(model_name)
		if !found || aviModel == nil {
			utils.AviLog.Infof("key: %s, msg: model not found, generating new model with name: %s", key, model_name)
			aviModel = NewAviObjectGraph()
			aviModel.(*AviObjectGraph).ConstructAviL7SharedVsNodeForEvh(shardVsName.Name, key, routeIgrObj, shardVsName.Dedicated, true)
		}

		vsNode := aviModel.(*AviObjectGraph).GetAviEvhVS()
		if len(vsNode) < 1 {
			return nil
		}
		modelGraph := aviModel.(*AviObjectGraph)
		modelGraph.BuildModelGraphForSecureEVH(routeIgrObj, ingressHostMap, hosts, tlssetting, ingName, namespace, infraSetting, host, key, paths)

		if found {
			// if vsNode already exists, check for updates via AviInfraSetting
			if infraSetting != nil {
				buildWithInfraSettingForEvh(key, namespace, vsNode[0], vsNode[0].VSVIPRefs[0], infraSetting)
			}
		}

		// Only add this node to the list of models if the checksum has changed.
		utils.AviLog.Debugf("key: %s, Saving Model: %v", key, utils.Stringify(vsNode))
		modelChanged := saveAviModel(model_name, modelGraph, key)
		if !utils.HasElem(*modelList, model_name) && modelChanged {
			*modelList = append(*modelList, model_name)
		}

	}

	return hostPathSvcMap
}

func (o *AviObjectGraph) BuildModelGraphForSecureEVH(routeIgrObj RouteIngressModel, ingressHostMap SecureHostNameMapProp, hosts []string, tlssetting TlsSettings, ingName, namespace string, infraSetting *akov1beta1.AviInfraSetting, host, key string, paths HostMetadata) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	var evhNode *AviEvhVsNode
	vsNode := o.GetAviEvhVS()
	isDedicated := vsNode[0].Dedicated
	certsBuilt := false
	evhSecretName := tlssetting.SecretName
	if lib.IsSecretAviCertRef(evhSecretName) {
		certsBuilt = true
	}

	var infraSettingName string
	if infraSetting != nil && !lib.IsInfraSettingNSScoped(infraSetting.Name, namespace) {
		infraSettingName = infraSetting.Name
	}
	if !isDedicated {
		childVSName := lib.GetEvhNodeName(host, infraSettingName)
		evhNode = vsNode[0].GetEvhNodeForName(childVSName)
		if evhNode == nil {
			evhNode = &AviEvhVsNode{
				Name:         childVSName,
				VHParentName: vsNode[0].Name,
				Tenant:       vsNode[0].Tenant,
				EVHParent:    false,
				EvhHostName:  host,
				ServiceMetadata: lib.ServiceMetadataObj{
					NamespaceIngressName: ingressHostMap.GetIngressesForHostName(host),
					Namespace:            namespace,
					HostNames:            hosts,
				},
			}
		} else {
			// The evh node exists, just update the svc metadata
			evhNode.ServiceMetadata.NamespaceIngressName = ingressHostMap.GetIngressesForHostName(host)
			evhNode.ServiceMetadata.Namespace = namespace
			evhNode.ServiceMetadata.HostNames = hosts
		}
		evhNode.ApplicationProfile = utils.DEFAULT_L7_APP_PROFILE
		evhNode.ServiceEngineGroup = lib.GetSEGName()
		evhNode.VrfContext = lib.GetVrf()
		evhNode.AviMarkers = lib.PopulateVSNodeMarkers(namespace, host, infraSettingName)
	} else {
		vsNode[0].ServiceMetadata.NamespaceIngressName = ingressHostMap.GetIngressesForHostName(host)
		vsNode[0].ServiceMetadata.Namespace = namespace
		vsNode[0].ServiceMetadata.HostNames = hosts
		vsNode[0].AddSSLPort(key)
		vsNode[0].Secure = true
		vsNode[0].ApplicationProfile = utils.DEFAULT_L7_SECURE_APP_PROFILE
		vsNode[0].AviMarkers = lib.PopulateVSNodeMarkers(namespace, host, infraSettingName)
	}

	var hostsToRemove []string
	hostsToRemove = append(hostsToRemove, host)
	found, gsFqdnCache := objects.SharedCRDLister().GetLocalFqdnToGSFQDNMapping(host)
	if paths.gslbHostHeader == "" {
		// If the gslbHostHeader is empty but it is present in the in memory cache, then add it as a candidate for removal and  remove the in memory cache relationship
		if found {
			hostsToRemove = append(hostsToRemove, gsFqdnCache)
			objects.SharedCRDLister().DeleteLocalFqdnToGsFqdnMap(host)
			if vsNode[0].Dedicated {
				RemoveFqdnFromEVHVIP(vsNode[0], []string{gsFqdnCache}, key)
			}
		}
	} else {
		if paths.gslbHostHeader != gsFqdnCache {
			hostsToRemove = append(hostsToRemove, gsFqdnCache)
		}
		if vsNode[0].Dedicated {
			RemoveFqdnFromEVHVIP(vsNode[0], []string{gsFqdnCache}, key)
		}
		objects.SharedCRDLister().UpdateLocalFQDNToGSFqdnMapping(host, paths.gslbHostHeader)
	}

	if !certsBuilt {
		certsBuilt = o.BuildTlsCertNodeForEvh(routeIgrObj.GetSvcLister(), vsNode[0], namespace, tlssetting, key, infraSettingName, host)
	} else {
		//Delete sslcertref object if host crd sslcertref (sslcertAviRef) is present for given host
		secretName := tlssetting.SecretName
		if strings.HasPrefix(secretName, lib.RouteSecretsPrefix) {
			//Openshift: Remove CA cert if present
			vsNode[0].DeleteCACertRefInEVHNode(lib.GetCACertNodeName(infraSettingName, host), key)
		}
		vsNode[0].DeleteSSLRefInEVHNode(lib.GetTLSKeyCertNodeName(infraSettingName, host, tlssetting.SecretName), key)
		vsNode[0].DeleteSSLRefInEVHNode(lib.GetTLSKeyCertNodeName(infraSettingName, host, tlssetting.SecretName+"-alt"), key)
	}
	if isDedicated {
		evhNode = vsNode[0]
	}
	if certsBuilt {
		hosts := []string{host}
		if paths.gslbHostHeader != "" {
			if !utils.HasElem(hosts, paths.gslbHostHeader) {
				hosts = append(hosts, paths.gslbHostHeader)
			}
		}

		o.BuildPolicyPGPoolsForEVH(vsNode, evhNode, namespace, ingName, key, infraSetting, hosts, paths.ingressHPSvc, &tlssetting, routeIgrObj.GetType())
		if !isDedicated {
			foundEvhModel := FindAndReplaceEvhInModel(evhNode, vsNode, key)
			if !foundEvhModel {
				vsNode[0].EvhNodes = append(vsNode[0].EvhNodes, evhNode)
			}
		}
		//EVH VS node (For route): either will have redirect policy or None policy
		RemoveRedirectHTTPPolicyInModelForEvh(evhNode, hostsToRemove, key)

		if tlssetting.redirect {
			o.BuildPolicyRedirectForVSForEvh(evhNode, hosts, namespace, ingName, key, infraSettingName)
		} else if tlssetting.blockHTTPTraffic {
			//Add drop rule to block traffic on 80
			o.BuildHTTPSecurityPolicyForVSForEvh(evhNode, hosts, namespace, ingName, key, infraSettingName)
		}
		// Enable host rule
		BuildL7HostRule(host, key, evhNode)
		// build SSORule for secure ingress in evh
		BuildL7SSORule(host, key, evhNode)
		if !isDedicated {
			manipulateEvhNodeForSSL(key, vsNode[0], evhNode)
		}

		// Remove the deleted aliases from the FQDN list
		var hostsToRemove []string
		_, oldFQDNAliases := objects.SharedCRDLister().GetFQDNToAliasesMapping(host)
		for _, host := range oldFQDNAliases {
			if !utils.HasElem(evhNode.VHDomainNames, host) {
				hostsToRemove = append(hostsToRemove, host)
			}
		}
		vsNode[0].RemoveFQDNsFromModel(hostsToRemove, key)
		evhNode.RemoveFQDNAliasesFromHTTPPolicy(hostsToRemove, key)

		// Add FQDN aliases in the hostrule CRD to parent and child VSes
		vsNode[0].AddFQDNsToModel(evhNode.VHDomainNames, paths.gslbHostHeader, key)
		evhNode.AddFQDNAliasesToHTTPPolicy(evhNode.VHDomainNames, key)
		evhNode.AviMarkers.Host = evhNode.VHDomainNames
		objects.SharedCRDLister().UpdateFQDNToAliasesMappings(host, evhNode.VHDomainNames)

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
			hostsToRemove = append(hostsToRemove, evhNode.VHDomainNames...)
			if vsNode[0].Dedicated {
				DeleteDedicatedEvhVSNode(vsNode[0], key, hostsToRemove)
			} else {
				vsNode[0].DeleteSSLRefInEVHNode(lib.GetTLSKeyCertNodeName(infraSettingName, host, tlssetting.SecretName), key)
				vsNode[0].DeleteSSLRefInEVHNode(lib.GetTLSKeyCertNodeName(infraSettingName, host, tlssetting.SecretName+"-alt"), key)
				RemoveEvhInModel(evhNode.Name, vsNode, key)
				RemoveRedirectHTTPPolicyInModelForEvh(evhNode, hostsToRemove, key)
			}
			vsNode[0].RemoveFQDNsFromModel(hostsToRemove, key)
		}
	}
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

// As either HttpSecurityPolicy or HttpRedirect policy exists, using same function for both.
func FindAndReplaceRedirectHTTPPolicyInModelforEvh(vsNode *AviEvhVsNode, httpPolicy *AviHttpPolicySetNode, hostnames []string, key string) bool {
	var policyFound bool
	for _, hostname := range hostnames {
		for _, policy := range vsNode.HttpPolicyRefs {
			//For existing HttpPolicyset, AviMarkers will be empty.
			if policy.Name == httpPolicy.Name && policy.AviMarkers.Namespace != "" {
				//No action for httpsecurity policy as port and actions are currently constant.
				if policy.RedirectPorts != nil && !utils.HasElem(policy.RedirectPorts[0].Hosts, hostname) {
					policy.RedirectPorts[0].Hosts = append(policy.RedirectPorts[0].Hosts, hostname)
					utils.AviLog.Debugf("key: %s, msg: replaced host %s for policy %s in model", key, hostname, policy.Name)
				}
				policyFound = true
			}
		}
	}
	return policyFound
}

// As either HttpSecurity policy or http redirect policy exists, using same function for both.
func RemoveRedirectHTTPPolicyInModelForEvh(vsNode *AviEvhVsNode, hostnames []string, key string) {
	policyName := lib.GetL7HttpRedirPolicy(vsNode.Name)
	for _, hostname := range hostnames {
		for i, policy := range vsNode.HttpPolicyRefs {
			if policy.Name == policyName {
				if policy.RedirectPorts != nil {
					// one redirect policy per child EVH vs
					if utils.HasElem(policy.RedirectPorts[0].Hosts, hostname) {
						policy.RedirectPorts[0].Hosts = utils.Remove(policy.RedirectPorts[0].Hosts, hostname)
						utils.AviLog.Debugf("key: %s, msg: removed host %s from policy %s in model", key, hostname, policy.Name)
					}
					if len(policy.RedirectPorts[0].Hosts) == 0 {
						vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs[:i], vsNode.HttpPolicyRefs[i+1:]...)
						utils.AviLog.Infof("key: %s, msg: removed redirect policy %s in model", key, policy.Name)
					}
				} else if policy.SecurityRules != nil {
					//Remove security policy
					vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs[:i], vsNode.HttpPolicyRefs[i+1:]...)
					utils.AviLog.Infof("key: %s, msg: removed security policy %s in model", key, policy.Name)
				}
			}
		}
	}
}

// DeleteStaleData : delete pool, EVH VS and redirect policy which are present in the object store but no longer required.
func DeleteStaleDataForEvh(routeIgrObj RouteIngressModel, key string, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost, hostsMap map[string]*objects.RouteIngrhost) {
	utils.AviLog.Debugf("key: %s, msg: About to delete stale data EVH Stored hosts: %v, hosts map: %v", key, utils.Stringify(Storedhosts), utils.Stringify(hostsMap))
	var infraSettingName string
	tenant := lib.GetTenant()
	if aviInfraSetting := routeIgrObj.GetAviInfraSetting(); aviInfraSetting != nil {
		if !lib.IsInfraSettingNSScoped(aviInfraSetting.Name, routeIgrObj.GetNamespace()) {
			infraSettingName = aviInfraSetting.Name
		}
		if aviInfraSetting.Spec.NSXSettings.Project != nil {
			tenant = *aviInfraSetting.Spec.NSXSettings.Project
		}
	}

	for host, hostData := range Storedhosts {
		utils.AviLog.Debugf("host to del: %s, data : %s", host, utils.Stringify(hostData))
		_, shardVsName := DeriveShardVSForEvh(host, key, routeIgrObj)
		if hostData.SecurePolicy == lib.PolicyPass {
			_, shardVsName.Name = DerivePassthroughVS(host, key, routeIgrObj)
		}
		modelName := lib.GetModelName(tenant, shardVsName.Name)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Warnf("key: %s, msg: model not found during delete: %s", key, modelName)
			continue
		}
		// By default remove both redirect and fqdn. So if the host isn't transitioning, then we will remove both.
		removeFqdn := true
		removeRedir := true
		removeRouteIngData := true
		currentData, ok := hostsMap[host]
		utils.AviLog.Debugf("key: %s, hostsMap: %s", key, utils.Stringify(hostsMap))
		// if route is transitioning from/to passthrough route, then always remove fqdn
		if ok && hostData.SecurePolicy != lib.PolicyPass && currentData.SecurePolicy != lib.PolicyPass {
			if currentData.InsecurePolicy == lib.PolicyRedirect {
				removeRedir = false
			}
			utils.AviLog.Infof("key: %s, host: %s, currentData: %v", key, host, currentData)
			removeFqdn = false
			if routeIgrObj.GetType() == utils.OshiftRoute {
				diff := lib.GetDiffPath(hostData.PathSvc, currentData.PathSvc)
				if len(diff) == 0 {
					removeRouteIngData = false
				}
			}
		}
		// Delete the pool corresponding to this host
		isPassthroughVS := false
		if hostData.SecurePolicy == lib.PolicyEdgeTerm {
			aviModel.(*AviObjectGraph).DeletePoolForHostnameForEvh(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, removeFqdn, removeRedir, removeRouteIngData, true)
		} else if hostData.SecurePolicy == lib.PolicyPass {
			isPassthroughVS = true
			aviModel.(*AviObjectGraph).DeleteObjectsForPassthroughHost(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, infraSettingName, key, true, true, true)
		}
		if hostData.InsecurePolicy != lib.PolicyNone {
			if isPassthroughVS {
				aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, true, true, false)
			} else {
				aviModel.(*AviObjectGraph).DeletePoolForHostnameForEvh(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, removeFqdn, removeRedir, removeRouteIngData, false)
			}
		}
		changedModel := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if !utils.HasElem(modelList, modelName) && changedModel {
			*modelList = append(*modelList, modelName)
		}
	}
}

func DeriveShardVSForEvh(hostname, key string, routeIgrObj RouteIngressModel) (lib.VSNameMetadata, lib.VSNameMetadata) {
	// Read the value of the num_shards from the environment variable.
	utils.AviLog.Debugf("key: %s, msg: hostname for sharding: %s", key, hostname)
	var newInfraPrefix, oldInfraPrefix string
	oldTenant, newTenant := lib.GetTenant(), lib.GetTenant()
	oldShardSize, newShardSize := lib.GetshardSize(), lib.GetshardSize()
	var oldVSNameMeta lib.VSNameMetadata
	var newVSNameMeta lib.VSNameMetadata
	// get stored infrasetting from ingress/route
	// figure out the current infrasetting via class/annotation
	var oldSettingName string
	var found bool
	if found, oldSettingName = objects.InfraSettingL7Lister().GetIngRouteToInfraSetting(routeIgrObj.GetNamespace() + "/" + routeIgrObj.GetName()); found {
		if found, shardSize := objects.InfraSettingL7Lister().GetInfraSettingToShardSize(oldSettingName); found && shardSize != "" {
			oldShardSize = lib.ShardSizeMap[shardSize]
		}
		tenant := objects.InfraSettingL7Lister().GetAviInfraSettingToTenant(oldSettingName)
		if tenant != "" {
			oldTenant = tenant
		}
		if !lib.IsInfraSettingNSScoped(oldSettingName, routeIgrObj.GetNamespace()) {
			oldInfraPrefix = oldSettingName
		}
	} else {
		utils.AviLog.Debugf("AviInfraSetting %s not found in cache", oldSettingName)
	}

	newSetting := routeIgrObj.GetAviInfraSetting()
	if !routeIgrObj.Exists() {
		// get the old ones.
		newShardSize = oldShardSize
		newInfraPrefix = oldInfraPrefix
		newTenant = oldTenant
	} else if newSetting != nil {
		if newSetting.Spec.L7Settings != (akov1beta1.AviInfraL7Settings{}) {
			newShardSize = lib.ShardSizeMap[newSetting.Spec.L7Settings.ShardSize]
		}
		if newSetting.Spec.NSXSettings.Project != nil {
			newTenant = *newSetting.Spec.NSXSettings.Project
		}
		if !lib.IsInfraSettingNSScoped(newSetting.Name, routeIgrObj.GetNamespace()) {
			newInfraPrefix = newSetting.Name
		}
	}
	shardVsPrefix := lib.GetNamePrefix() + lib.GetAKOIDPrefix() + lib.ShardEVHVSPrefix
	oldVsName, newVsName := shardVsPrefix, shardVsPrefix
	if oldInfraPrefix != "" {
		oldVsName += oldInfraPrefix + "-"
	}
	if newInfraPrefix != "" {
		newVsName += newInfraPrefix + "-"
	}

	if lib.VIPPerNamespace() {
		oldVsName += "NS-" + routeIgrObj.GetNamespace()
		newVsName += "NS-" + routeIgrObj.GetNamespace()
	} else {
		if oldShardSize != 0 {
			oldVsName += strconv.Itoa(int(utils.Bkt(hostname, oldShardSize)))
		} else {
			//Dedicated VS
			oldVsName = GetDedicatedVSName(hostname, oldInfraPrefix)
			oldVSNameMeta.Dedicated = true
		}
		if newShardSize != 0 {
			newVsName += strconv.Itoa(int(utils.Bkt(hostname, newShardSize)))
		} else {
			//Dedicated VS
			newVsName = GetDedicatedVSName(hostname, newInfraPrefix)
			newVSNameMeta.Dedicated = true
		}
	}
	oldVSNameMeta.Name = oldVsName
	oldVSNameMeta.Tenant = oldTenant
	newVSNameMeta.Name = newVsName
	newVSNameMeta.Tenant = newTenant
	utils.AviLog.Infof("key: %s, msg: ShardVSNames: %s %s", key, oldVsName, newVsName)
	return oldVSNameMeta, newVSNameMeta
}
func GetDedicatedVSName(host, infrasettingName string) string {
	var name string
	//For Dedicated Vs Name: Not encoded Suffix -EVH for process by evh model during dequeue node operation
	if infrasettingName != "" {
		name = lib.Encode(lib.GetNamePrefix()+infrasettingName+"-"+host+lib.DedicatedSuffix, lib.EVHVS) + lib.EVHSuffix
		return name
	}
	name = lib.Encode(lib.GetNamePrefix()+host+lib.DedicatedSuffix, lib.EVHVS) + lib.EVHSuffix
	return name
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
func RemoveFqdnFromEVHVIP(vsNode *AviEvhVsNode, hostsToRemove []string, key string) {
	if len(vsNode.VSVIPRefs) > 0 {
		for _, fqdn := range hostsToRemove {
			for i, vipFqdn := range vsNode.VSVIPRefs[0].FQDNs {
				if vipFqdn == fqdn {
					utils.AviLog.Debugf("key: %s, msg: Removed FQDN %s from vs node %s", key, fqdn, vsNode.Name)
					vsNode.VSVIPRefs[0].FQDNs = append(vsNode.VSVIPRefs[0].FQDNs[:i], vsNode.VSVIPRefs[0].FQDNs[i+1:]...)
				}
			}
		}
	}
}
func (o *AviObjectGraph) RemoveHTTPRefsFromEvh(httpPol, hppmapName string, evhNode *AviEvhVsNode) {

	for i, pol := range evhNode.HttpPolicyRefs {
		if pol.Name == httpPol {
			for j, hppmap := range evhNode.HttpPolicyRefs[i].HppMap {
				if hppmap.Name == hppmapName {
					evhNode.HttpPolicyRefs[i].HppMap = append(evhNode.HttpPolicyRefs[i].HppMap[:j], evhNode.HttpPolicyRefs[i].HppMap[j+1:]...)
					break
				}
			}
			if len(pol.HppMap) == 0 {
				utils.AviLog.Debugf("Removing http pol ref: %s", httpPol)
				evhNode.HttpPolicyRefs = append(evhNode.HttpPolicyRefs[:i], evhNode.HttpPolicyRefs[i+1:]...)
				break
			}
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
func (o *AviObjectGraph) manipulateEVHVsNode(vsNode *AviEvhVsNode, ingName, namespace, hostname string, pathSvc map[string][]string, infraSettingName, key string) {
	for path, services := range pathSvc {
		pgName := lib.GetEvhPGName(ingName, namespace, hostname, path, infraSettingName, vsNode.Dedicated)
		pgNode := vsNode.GetPGForVSByName(pgName)
		for _, svc := range services {
			evhPool := lib.GetEvhPoolName(ingName, namespace, hostname, path, infraSettingName, svc, vsNode.Dedicated)
			o.RemovePoolNodeRefsFromEvh(evhPool, vsNode)
			o.RemovePoolRefsFromPG(evhPool, pgNode)

			// Remove the EVH PG if it has no member
			if pgNode != nil {
				if len(pgNode.Members) == 0 {
					o.RemovePGNodeRefsForEvh(pgName, vsNode)
					httppolname := lib.GetSniHttpPolName(namespace, hostname, infraSettingName)
					hppmapname := lib.GetEvhPGName(ingName, namespace, hostname, path, infraSettingName, vsNode.Dedicated)
					o.RemoveHTTPRefsFromEvh(httppolname, hppmapname, vsNode)
				}
			}
		}
	}
}
func (o *AviObjectGraph) ManipulateEvhNode(currentEvhNodeName, ingName, namespace, hostname string, pathSvc map[string][]string, vsNode []*AviEvhVsNode, infraSettingName, key string) bool {
	if vsNode[0].Dedicated {
		o.manipulateEVHVsNode(vsNode[0], ingName, namespace, hostname, pathSvc, infraSettingName, key)
		if len(vsNode[0].PoolGroupRefs) == 0 {
			// Remove the evhhost mapping
			SharedHostNameLister().Delete(hostname)
			vsNode[0].DeletSSLRefInDedicatedNode(key)
			return false
		}
	} else {
		for _, modelEvhNode := range vsNode[0].EvhNodes {
			if currentEvhNodeName != modelEvhNode.Name {
				continue
			}
			o.manipulateEVHVsNode(modelEvhNode, ingName, namespace, hostname, pathSvc, infraSettingName, key)
			// After going through the paths, if the EVH node does not have any PGs - then delete it.
			if len(modelEvhNode.PoolGroupRefs) == 0 {
				RemoveEvhInModel(currentEvhNodeName, vsNode, key)
				// Remove the evhhost mapping
				SharedHostNameLister().Delete(hostname)
				return false
			}
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

func (o *AviObjectGraph) DeletePoolForHostnameForEvh(vsName, hostname string, routeIgrObj RouteIngressModel, pathSvc map[string][]string, key, infraSettingName string, removeFqdn, removeRedir, removeRouteIngData, secure bool) bool {
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

	evhNodeName := lib.GetEvhNodeName(hostname, infraSettingName)
	utils.AviLog.Infof("key: %s, msg: EVH node to delete: %s", key, evhNodeName)
	if removeRouteIngData {
		keepEvh = o.ManipulateEvhNode(evhNodeName, ingName, namespace, hostname, pathSvc, vsNode, infraSettingName, key)
	}
	if !keepEvh {
		// Delete the cert ref for the host
		vsNode[0].DeleteSSLRefInEVHNode(lib.GetTLSKeyCertNodeName(infraSettingName, hostname, ""), key)
	}
	_, FQDNAliases := objects.SharedCRDLister().GetFQDNToAliasesMapping(hostname)
	if removeFqdn && !keepEvh {
		var hosts []string
		found, gsFqdnCache := objects.SharedCRDLister().GetLocalFqdnToGSFQDNMapping(hostname)
		if found {
			hosts = append(hosts, gsFqdnCache)
		}
		hosts = append(hosts, hostname)
		hosts = append(hosts, FQDNAliases...)
		// Remove these hosts from the overall FQDN list
		vsNode[0].RemoveFQDNsFromModel(hosts, key)
	}
	if removeRedir && !keepEvh {
		var hostnames []string
		found, gsFqdnCache := objects.SharedCRDLister().GetLocalFqdnToGSFQDNMapping(hostname)
		if found {
			hostnames = append(hostnames, gsFqdnCache)
		}
		hostnames = append(hostnames, hostname)
		hostnames = append(hostnames, FQDNAliases...)
		RemoveRedirectHTTPPolicyInModelForEvh(vsNode[0], hostnames, key)
	}
	if vsNode[0].Dedicated && !keepEvh {
		return true
	}
	return false
}

func (o *AviObjectGraph) BuildPolicyRedirectForVSForEvh(vsNode *AviEvhVsNode, hostnames []string, namespace, ingName, key, infraSettingName string) {
	policyname := lib.GetL7HttpRedirPolicy(vsNode.Name)
	myHppMap := AviRedirectPort{
		Hosts:        hostnames,
		RedirectPort: 443,
		StatusCode:   lib.STATUS_REDIRECT,
		VsPort:       80,
	}

	redirectPolicy := &AviHttpPolicySetNode{
		Tenant:        vsNode.Tenant,
		Name:          policyname,
		RedirectPorts: []AviRedirectPort{myHppMap},
	}
	redirectPolicy.AviMarkers = lib.PopulateVSNodeMarkers(namespace, hostnames[0], infraSettingName)
	if policyFound := FindAndReplaceRedirectHTTPPolicyInModelforEvh(vsNode, redirectPolicy, hostnames, key); !policyFound {
		redirectPolicy.CalculateCheckSum()
		vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs, redirectPolicy)
	}

}

func (o *AviObjectGraph) BuildHTTPSecurityPolicyForVSForEvh(vsNode *AviEvhVsNode, hostnames []string, namespace, ingName, key, infrasettingName string) {
	policyname := lib.GetL7HttpRedirPolicy(vsNode.Name)
	// Close Connection.
	securityRule := AviHTTPSecurity{
		Action:        lib.CLOSE_CONNECTION,
		MatchCriteria: lib.IS_IN,
		Enable:        true,
		Port:          80,
	}

	securityPolicy := &AviHttpPolicySetNode{
		Tenant:        vsNode.Tenant,
		Name:          policyname,
		SecurityRules: []AviHTTPSecurity{securityRule},
	}
	securityPolicy.AviMarkers = lib.PopulateVSNodeMarkers(namespace, hostnames[0], infrasettingName)
	if policyFound := FindAndReplaceRedirectHTTPPolicyInModelforEvh(vsNode, securityPolicy, hostnames, key); !policyFound {
		securityPolicy.CalculateCheckSum()
		vsNode.HttpPolicyRefs = append(vsNode.HttpPolicyRefs, securityPolicy)
	}

}

// RouteIngrDeletePoolsByHostname : Based on DeletePoolsByHostname, delete pools and policies that are no longer required
func RouteIngrDeletePoolsByHostnameForEvh(routeIgrObj RouteIngressModel, namespace, objname, key string, fullsync bool, sharedQueue *utils.WorkerQueue) {
	ok, hostMap := routeIgrObj.GetSvcLister().IngressMappings(namespace).GetRouteIngToHost(objname)
	if !ok {
		utils.AviLog.Warnf("key: %s, msg: nothing to delete for route: %s", key, objname)
		return
	}

	_, infraSettingName := objects.InfraSettingL7Lister().GetIngRouteToInfraSetting(routeIgrObj.GetNamespace() + "/" + routeIgrObj.GetName())
	tenant := objects.InfraSettingL7Lister().GetAviInfraSettingToTenant(infraSettingName)
	if tenant == "" {
		tenant = lib.GetTenant()
	}
	if lib.IsInfraSettingNSScoped(infraSettingName, namespace) {
		infraSettingName = ""
	}

	utils.AviLog.Debugf("key: %s, msg: hosts to delete are :%s", key, utils.Stringify(hostMap))
	for host, hostData := range hostMap {
		shardVsName, _ := DeriveShardVSForEvh(host, key, routeIgrObj)
		deleteVS := false
		if hostData.SecurePolicy == lib.PolicyPass {
			shardVsName.Name, _ = DerivePassthroughVS(host, key, routeIgrObj)
		}

		modelName := lib.GetModelName(tenant, shardVsName.Name)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Warnf("key: %s, msg: model not found during delete: %s", key, modelName)
			continue
		}

		// Delete the pool corresponding to this host
		if hostData.SecurePolicy == lib.PolicyEdgeTerm {
			deleteVS = aviModel.(*AviObjectGraph).DeletePoolForHostnameForEvh(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, true, true, true, true)
		} else if hostData.SecurePolicy == lib.PolicyPass {
			aviModel.(*AviObjectGraph).DeleteObjectsForPassthroughHost(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, infraSettingName, key, true, true, true)
		}
		if hostData.InsecurePolicy == lib.PolicyAllow {
			deleteVS = aviModel.(*AviObjectGraph).DeletePoolForHostnameForEvh(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, true, true, true, false)
		}
		if !deleteVS {
			ok := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
			if ok && len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 && !fullsync {
				PublishKeyToRestLayer(modelName, key, sharedQueue)
			}
		} else {
			utils.AviLog.Debugf("Setting up model name :[%v] to nil", modelName)
			objects.SharedAviGraphLister().Save(modelName, nil)
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
	var shardVsName lib.VSNameMetadata
	var newShardVsName lib.VSNameMetadata
	utils.AviLog.Debugf("key: %s, msg: hosts to delete %s", key, utils.Stringify(hostMap))
	for host, hostData := range hostMap {

		shardVsName, newShardVsName = DeriveShardVSForEvh(host, key, routeIgrObj)
		if hostData.SecurePolicy == lib.PolicyPass {
			shardVsName.Name, newShardVsName.Name = DerivePassthroughVS(host, key, routeIgrObj)
		}
		if shardVsName == newShardVsName {
			continue
		}

		_, infraSettingName := objects.InfraSettingL7Lister().GetIngRouteToInfraSetting(routeIgrObj.GetNamespace() + "/" + routeIgrObj.GetName())
		if lib.IsInfraSettingNSScoped(infraSettingName, namespace) {
			infraSettingName = ""
		}
		modelName := lib.GetModelName(shardVsName.Tenant, shardVsName.Name)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Warnf("key: %s, msg: model not found during delete: %s", key, modelName)
			continue
		}

		// Delete the pool corresponding to this host
		isPassthroughVS := false
		if hostData.SecurePolicy == lib.PolicyEdgeTerm {
			aviModel.(*AviObjectGraph).DeletePoolForHostnameForEvh(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, true, true, true, true)
		} else if hostData.SecurePolicy == lib.PolicyPass {
			isPassthroughVS = true
			aviModel.(*AviObjectGraph).DeleteObjectsForPassthroughHost(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, infraSettingName, key, true, true, true)
		}
		if hostData.InsecurePolicy != lib.PolicyNone {
			if isPassthroughVS {
				aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, true, true, false)
			} else {
				aviModel.(*AviObjectGraph).DeletePoolForHostnameForEvh(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, true, true, true, false)
			}
		}

		ok := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if ok && len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 && !fullsync {
			PublishKeyToRestLayer(modelName, key, sharedQueue)
		}
	}
}

func buildWithInfraSettingForEvh(key, namespace string, vs *AviEvhVsNode, vsvip *AviVSVIPNode, infraSetting *akov1beta1.AviInfraSetting) {
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

		if vs.EnableRhi != nil && *vs.EnableRhi {
			if infraSetting.Spec.Network.BgpPeerLabels != nil {
				vsvip.BGPPeerLabels = infraSetting.Spec.Network.BgpPeerLabels
			} else {
				vsvip.BGPPeerLabels = lib.GetGlobalBgpPeerLabels()
			}
		} else {
			vsvip.BGPPeerLabels = nil
		}

		if infraSetting.Spec.Network.VipNetworks != nil && len(infraSetting.Spec.Network.VipNetworks) > 0 {
			vsvip.VipNetworks = lib.GetVipInfraNetworkList(infraSetting.Name)
		} else {
			vsvip.VipNetworks = utils.GetVipNetworkList()
		}
		if lib.IsPublicCloud() {
			vsvip.EnablePublicIP = infraSetting.Spec.Network.EnablePublicIP
		}
		if (vs.EVHParent || vs.Dedicated) && (infraSetting.Spec.Network.Listeners != nil && len(infraSetting.Spec.Network.Listeners) > 0) {
			portProto := buildListenerPortsWithInfraSetting(infraSetting, vs.PortProto)
			vs.SetPortProtocols(portProto)
		}
		if infraSetting.Spec.NSXSettings.T1LR != nil {
			vsvip.T1Lr = *infraSetting.Spec.NSXSettings.T1LR
		}
		utils.AviLog.Debugf("key: %s, msg: Applied AviInfraSetting configuration over VSNode %s", key, vs.Name)
	}
}
func DeleteDedicatedEvhVSNode(vsNode *AviEvhVsNode, key string, hostsToRemove []string) {
	vsNode.PoolGroupRefs = []*AviPoolGroupNode{}
	vsNode.PoolRefs = []*AviPoolNode{}
	vsNode.HttpPolicyRefs = []*AviHttpPolicySetNode{}
	vsNode.DeletSSLRefInDedicatedNode(key)
	RemoveFqdnFromEVHVIP(vsNode, hostsToRemove, key)
	utils.AviLog.Infof("key: %s, msg: Deleted Dedicated node vs: %s", key, vsNode.Name)
}
func manipulateEvhNodeForSSL(key string, vsNode *AviEvhVsNode, evhNode *AviEvhVsNode) {
	oldSSLProfile := vsNode.GetSSLProfileRef()
	newSSLProfile := evhNode.GetSSLProfileRef()
	if oldSSLProfile != nil &&
		*oldSSLProfile != "" &&
		newSSLProfile != nil &&
		*oldSSLProfile != *newSSLProfile {
		utils.AviLog.Warnf("key: %s msg: overwriting old ssl profile %s with new ssl profile %s", key, *oldSSLProfile, *newSSLProfile)
	}
	vsNode.SetSSLProfileRef(newSSLProfile)
	evhNode.SetSSLProfileRef(nil)
}
