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

package lib

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"

	"github.com/Masterminds/semver"
	routev1 "github.com/openshift/api/route/v1"
	oshiftclient "github.com/openshift/client-go/route/clientset/versioned"
	"github.com/vmware/alb-sdk/go/models"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"

	"google.golang.org/protobuf/proto"
)

var ShardSchemeMap = map[string]string{
	"hostname":  "hostname",
	"namespace": "namespace",
}

var ShardSizeMap = map[string]uint32{
	"LARGE":     8,
	"MEDIUM":    4,
	"SMALL":     1,
	"DEDICATED": 0,
}

var fqdnMap = map[string]string{
	"default": AutoFQDNDefault,
	"flat":    AutoFQDNFlat,
}

type CRDMetadata struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	Status string `json:"status"`
}

type ServiceMetadataObj struct {
	NamespaceIngressName  []string    `json:"namespace_ingress_name"`
	IngressName           string      `json:"ingress_name"`
	Namespace             string      `json:"namespace"`
	HostNames             []string    `json:"hostnames"`
	NamespaceServiceName  []string    `json:"namespace_svc_name"` // []string{ns/name}
	CRDStatus             CRDMetadata `json:"crd_status"`
	PoolRatio             int32       `json:"pool_ratio"`
	PassthroughParentRef  string      `json:"passthrough_parent_ref"`
	PassthroughChildRef   string      `json:"passthrough_child_ref"`
	Gateway               string      `json:"gateway"` // ns/name
	InsecureEdgeTermAllow bool        `json:"insecureedgetermallow"`
	IsMCIIngress          bool        `json:"is_mci_ingress"`
}

type ServiceMetadataMappingObjType string

const (
	GatewayVS            ServiceMetadataMappingObjType = "GATEWAY_VS"
	ChildVS              ServiceMetadataMappingObjType = "CHILD_VS"
	ServiceTypeLBVS      ServiceMetadataMappingObjType = "SERVICELB_VS"
	GatewayPool          ServiceMetadataMappingObjType = "GATEWAY_POOL"
	SNIInsecureOrEVHPool ServiceMetadataMappingObjType = "SNI_INSECURE_OR_EVH_POOL"
)

func (c ServiceMetadataObj) ServiceMetadataMapping(objType string) ServiceMetadataMappingObjType {
	if c.Gateway != "" {
		// Check for `Gateway` in VS serviceMetadata. Present in case of
		// 1) Advl4 VS
		// 2) SvcApi VS
		return GatewayVS
	} else if len(c.NamespaceIngressName) > 0 || c.PassthroughChildRef != "" {
		// Check for `NamespaceIngressName` in VS serviceMetadata. Present in case of
		// 1) SNI Secure VS
		// 2) EVH Secure/Insecure VS
		// or Check Passthrough VS using child ref
		return ChildVS
	} else if objType == "VS" && len(c.NamespaceServiceName) > 0 {
		// Check for `NamesppaceServiceName` in VS serviceMetadata. Present in case of
		// 1) Service TypeLB L4VSes
		return ServiceTypeLBVS
	} else if objType == "Pool" && len(c.NamespaceServiceName) > 0 {
		// Check for `NamespaceServiceName` in Pool serviceMetadata. Present in case of
		// 1) Advl4 Pools: without hostname information
		// 2) SvcApi Pools: with hostname information
		// 3) SharedVip SvcLB Pools: with hostname information
		return GatewayPool
	} else if c.Namespace != "" && c.IngressName != "" {
		// Check for `Namespace` and `IngressName` in Pool serviceMetadata. Present in case of
		// 1) Insecure Pools (SNI)
		// 2) Secure/Insecure Pools (EVH)
		return SNIInsecureOrEVHPool
	}
	return ""
}

type VSNameMetadata struct {
	Name      string
	Dedicated bool
}

var NamePrefix string

func CheckObjectNameLength(objName, objType string) bool {
	if len(objName) > AVI_OBJ_NAME_MAX_LENGTH {
		utils.AviLog.Warnf("%s name %s exceeds maximum length limit of %d characters for AVI Object", objType, objName, AVI_OBJ_NAME_MAX_LENGTH)
		return true
	}
	return false
}

func SetNamePrefix() {
	NamePrefix = GetClusterName() + "--"
}

func GetNamePrefix() string {
	return NamePrefix
}

func Encode(s, objType string) string {
	if !IsEvhEnabled() || IsWCP() {
		CheckObjectNameLength(s, objType)
		return s
	}
	hash := sha1.Sum([]byte(s))
	encodedStr := GetNamePrefix() + hex.EncodeToString(hash[:])
	//Added this check to be safe side if encoded name becomes greater than limit set
	CheckObjectNameLength(encodedStr, objType)
	return encodedStr
}

func IsNameEncoded(name string) bool {
	split := strings.Split(name, "--")
	if len(split) == 2 {
		_, err := hex.DecodeString(split[1])
		if err == nil {
			return true
		}
	}
	return false
}

var DisableSync bool
var layer7Only bool
var noPGForSNI bool
var NsxTTzType string
var deleteConfigMap bool

func SetNSXTTransportZone(tzType string) {
	NsxTTzType = tzType
	utils.AviLog.Infof("Setting NSX-T transport zone to: %v", tzType)
}

func GetNSXTTransportZone() string {
	return NsxTTzType
}

func GetFqdns(vsName, key string, subDomains []string) ([]string, string) {
	var fqdns []string
	var fqdn string
	autoFQDN := true
	if GetL4FqdnFormat() == AutoFQDNDisabled {
		autoFQDN = false
	}
	if subDomains != nil && autoFQDN {
		// honour defaultSubDomain from values.yaml if specified
		defaultSubDomain := GetDomain()
		if defaultSubDomain != "" && utils.HasElem(subDomains, defaultSubDomain) {
			subDomains = []string{defaultSubDomain}
		}

		// subDomains[0] would either have the defaultSubDomain value
		// or would default to the first dns subdomain it gets from the dns profile
		subdomain := subDomains[0]
		if strings.HasPrefix(subDomains[0], ".") {
			subdomain = strings.Replace(subDomains[0], ".", "", 1)
		}
		if GetL4FqdnFormat() == AutoFQDNDefault {
			// Generate the FQDN based on the logic: <svc_name>.<namespace>.<sub-domain>
			fqdn = vsName + "." + GetTenant() + "." + subdomain
		} else if GetL4FqdnFormat() == AutoFQDNFlat {
			// Generate the FQDN based on the logic: <svc_name>-<namespace>.<sub-domain>
			fqdn = vsName + "-" + GetTenant() + "." + subdomain
		}
		objects.SharedCRDLister().UpdateFQDNSharedVSModelMappings(fqdn, GetModelName(GetTenant(), vsName))
		utils.AviLog.Infof("key: %s, msg: Configured the shared VS with default fqdn as: %s", key, fqdn)
		fqdns = append(fqdns, fqdn)
	}
	return fqdns, fqdn
}

func SetDisableSync(state bool) {
	DisableSync = state
	utils.AviLog.Infof("Setting Disable Sync to: %v", state)
}
func SetDeleteConfigMap(deleteCMFlag bool) {
	deleteConfigMap = deleteCMFlag
	utils.AviLog.Debugf("Setting deleteConfigMap flag to: [%v]", deleteConfigMap)
}
func SetLayer7Only(val string) {
	if boolVal, err := strconv.ParseBool(val); err == nil {
		layer7Only = boolVal
	}
	utils.AviLog.Infof("Setting the value for the layer7Only flag %v", layer7Only)
}

func SetNoPGForSNI(val string) {
	if boolVal, err := strconv.ParseBool(val); err == nil {
		noPGForSNI = boolVal
	}
	utils.AviLog.Infof("Setting the value for the noPGForSNI flag %v", noPGForSNI)
}

func IsShardVS(vsName string) bool {
	if deleteConfigMap {
		//delete configmap is set, do not save anything
		return false
	}
	if GetshardSize() == 0 {
		//Dedicated mode
		return false
	}
	if IsEvhEnabled() {
		if strings.Contains(vsName, ShardEVHVSPrefix) {
			return true
		}
	} else {
		//Second condition is for migration from evh -> sni
		if strings.Contains(vsName, ShardVSPrefix) && !strings.Contains(vsName, ShardEVHVSPrefix) {
			return true
		}
	}
	return false
}

func GetNoPGForSNI() bool {
	return noPGForSNI
}

func GetLayer7Only() bool {
	return layer7Only
}
func GetDeleteConfigMap() bool {
	return deleteConfigMap
}

var AKOUser string

func SetAKOUser() {
	AKOUser = "ako-" + GetClusterName()
	isPrimaryAKO := akoControlConfigInstance.GetAKOInstanceFlag()
	if !isPrimaryAKO {
		AKOUser = AKOUser + "-" + os.Getenv("POD_NAMESPACE")
	}
	utils.AviLog.Infof("Setting AKOUser: %s for Avi Objects", AKOUser)
}

func GetAKOUser() string {
	return AKOUser
}

func GetshardSize() uint32 {
	if IsWCP() {
		// shard to 8 go routines in the REST layer
		return ShardSizeMap["LARGE"]
	}
	shardVsSize := os.Getenv("SHARD_VS_SIZE")
	shardSize, ok := ShardSizeMap[shardVsSize]
	if ok {
		return shardSize
	} else {
		return 1
	}
}

func GetL4FqdnFormat() string {
	if IsWCP() {
		// disable for advancedL4
		return AutoFQDNDisabled
	}

	fqdnFormat := os.Getenv("AUTO_L4_FQDN")
	val, ok := fqdnMap[fqdnFormat]
	if ok {
		return val
	}

	// If no match then disable FQDNs for L4.
	utils.AviLog.Infof("No valid value provided for autoFQDN, disabling feature.")
	return AutoFQDNDisabled
}

func GetModelName(namespace, objectName string) string {
	return namespace + "/" + objectName
}

// All L4 object names.
func GetL4VSName(svcName, namespace string) string {
	return Encode(NamePrefix+namespace+"-"+svcName, L4VS)
}

func GetL4VSVipName(svcName, namespace string) string {
	return Encode(NamePrefix+namespace+"-"+svcName, L4VIP)
}

func GetL4PoolName(svcName, namespace, protocol string, port int32) string {
	poolName := NamePrefix + namespace + "-" + svcName + "-" + protocol + "-" + strconv.Itoa(int(port))
	return Encode(poolName, L4Pool)
}

func GetAdvL4PoolName(svcName, namespace, gwName string, port int32) string {
	poolName := NamePrefix + namespace + "-" + svcName + "-" + gwName + "--" + strconv.Itoa(int(port))
	return Encode(poolName, L4AdvPool)
}

func GetSvcApiL4PoolName(svcName, namespace, gwName, protocol string, port int32) string {
	poolName := NamePrefix + namespace + "-" + svcName + "-" + gwName + "-" + protocol + "-" + strconv.Itoa(int(port))
	return Encode(poolName, L4AdvPool)
}

// All L7 object names.
func GetVsVipName(vsName string) string {
	vsVipName := vsName
	CheckObjectNameLength(vsVipName, VIP)
	return vsName
}

func GetL7InsecureDSName(vsName string) string {
	l7DSName := vsName
	CheckObjectNameLength(l7DSName, DataScript)
	return l7DSName
}

func GetL7SharedPGName(vsName string) string {
	l7PGName := vsName
	CheckObjectNameLength(l7PGName, PG)
	return l7PGName
}

func GetPassthroughPGName(hostname, infrasettingName string) string {
	var pgName string
	if infrasettingName != "" {
		pgName = GetClusterName() + "--" + infrasettingName + "-" + hostname
	} else {
		pgName = GetClusterName() + "--" + hostname
	}
	return pgName
}

func GetPassthroughPoolName(hostname, serviceName, infrasettingName string) string {
	var poolName string
	if infrasettingName != "" {
		poolName = GetClusterName() + "--" + infrasettingName + "-" + hostname + "-" + serviceName
	} else {
		poolName = GetClusterName() + "--" + hostname + "-" + serviceName
	}
	return poolName
}

func GetL7PoolName(priorityLabel, namespace, ingName, infrasetting string, args ...string) string {
	priorityLabel = strings.ReplaceAll(priorityLabel, "/", "_")
	var poolName string
	if infrasetting != "" {
		poolName = NamePrefix + infrasetting + "-" + priorityLabel + "-" + namespace + "-" + ingName
	} else {
		poolName = NamePrefix + priorityLabel + "-" + namespace + "-" + ingName
	}
	if len(args) > 0 {
		svcName := args[0]
		poolName = poolName + "-" + svcName
	}
	return Encode(poolName, Pool)
}

func GetL7HttpRedirPolicy(vsName string) string {
	httpRedirectPolicy := vsName
	CheckObjectNameLength(httpRedirectPolicy, HTTPRedirectPolicy)
	return httpRedirectPolicy
}

func GetHeaderRewritePolicy(vsName, localHost string) string {
	headerWriterPolicy := vsName + "--host-hdr-re-write" + "--" + localHost
	CheckObjectNameLength(headerWriterPolicy, HeaderRewritePolicy)
	return headerWriterPolicy
}

func GetSniNodeName(infrasetting, sniHostName string) string {
	namePrefix := NamePrefix
	if infrasetting != "" {
		namePrefix += infrasetting + "-"
	}
	return Encode(namePrefix+sniHostName, SNIVS)
}

func GetSniPoolName(ingName, namespace, host, path, infrasetting string, dedicatedVS bool, args ...string) string {
	path = strings.ReplaceAll(path, "/", "_")
	var poolName string
	if infrasetting != "" {
		poolName = NamePrefix + infrasetting + "-" + namespace + "-" + host + path + "-" + ingName
	} else {
		poolName = NamePrefix + namespace + "-" + host + path + "-" + ingName
	}
	if len(args) > 0 {
		svcName := args[0]
		poolName = poolName + "-" + svcName
	}
	if dedicatedVS {
		poolName += DedicatedSuffix
	}
	CheckObjectNameLength(poolName, Pool)
	return poolName
}

func GetSniHttpPolName(namespace, host, infrasetting string) string {

	if infrasetting != "" {
		return Encode(NamePrefix+infrasetting+"-"+namespace+"-"+host, HTTPPS)
	}
	return Encode(NamePrefix+namespace+"-"+host, HTTPPS)
}
func GetSniHppMapName(ingName, namespace, host, path, infrasetting string, dedicatedVS bool) string {
	path = strings.ReplaceAll(path, "/", "_")
	hppmap := NamePrefix
	if infrasetting != "" {
		hppmap += infrasetting + "-" + namespace + "-" + host + path + "-" + ingName
	} else {
		hppmap += namespace + "-" + host + path + "-" + ingName
	}
	if dedicatedVS {
		hppmap += DedicatedSuffix
	}
	return Encode(hppmap, HPPMAP)
}

func GetSniPGName(ingName, namespace, host, path, infrasetting string, dedicatedVS bool) string {
	path = strings.ReplaceAll(path, "/", "_")
	var sniPGName string
	if infrasetting != "" {
		sniPGName = NamePrefix + infrasetting + "-" + namespace + "-" + host + path + "-" + ingName
	} else {
		sniPGName = NamePrefix + namespace + "-" + host + path + "-" + ingName
	}
	if dedicatedVS {
		sniPGName += DedicatedSuffix
	}
	CheckObjectNameLength(sniPGName, PG)
	return sniPGName
}

// evh child
func GetEvhPoolName(ingName, namespace, host, path, infrasetting, svcName string, dedicatedVS bool) string {
	poolName := GetEvhPoolNameNoEncoding(ingName, namespace, host, path, infrasetting, svcName, dedicatedVS)
	return Encode(poolName, Pool)
}

func GetEvhPoolNameNoEncoding(ingName, namespace, host, path, infrasetting, svcName string, dedicatedVS bool) string {
	path = strings.ReplaceAll(path, "/", "_")
	namePrefix := NamePrefix
	if infrasetting != "" {
		namePrefix += infrasetting + "-"
	}
	poolName := namePrefix + namespace + "-" + host + path + "-" + ingName + "-" + svcName
	if dedicatedVS {
		poolName += DedicatedSuffix
	}
	return poolName
}

func GetEvhNodeName(host, infrasetting string) string {
	if infrasetting != "" {
		return Encode(NamePrefix+infrasetting+"-"+host, EVHVS)
	}
	return Encode(NamePrefix+host, EVHVS)
}

func GetEvhPGName(ingName, namespace, host, path, infrasetting string, dedicatedVs bool) string {
	path = strings.ReplaceAll(path, "/", "_")

	evhPG := NamePrefix
	if infrasetting != "" {
		evhPG += infrasetting + "-" + namespace + "-" + host + path + "-" + ingName
	} else {
		evhPG += namespace + "-" + host + path + "-" + ingName
	}
	if dedicatedVs {
		evhPG += DedicatedSuffix
	}
	return Encode(evhPG, PG)
}

func IsSecretK8sSecretRef(secret string) bool {
	re := regexp.MustCompile(fmt.Sprintf(`^%s.*`, DummySecretK8s))
	if re.MatchString(secret) {
		return true
	}
	return false
}

func IsSecretAviCertRef(secret string) bool {
	re := regexp.MustCompile(fmt.Sprintf(`^%s.*`, DummySecret))
	if re.MatchString(secret) {
		return true
	}
	return false
}

func GetTLSKeyCertNodeName(infrasetting, sniHostName, secretName string) string {
	if IsSecretK8sSecretRef(secretName) {
		secretName = strings.Split(secretName, "/")[2]
		if secretName == GetDefaultSecretForRoutes() {
			return Encode(secretName, TLSKeyCert)
		}
	}
	namePrefix := NamePrefix
	if infrasetting != "" {
		namePrefix += infrasetting + "-"
	}
	if strings.HasSuffix(secretName, "-alt") {
		return Encode(namePrefix+sniHostName+"-alt", TLSKeyCert)
	}
	return Encode(namePrefix+sniHostName, TLSKeyCert)
}

func GetCACertNodeName(infrasetting, sniHostName string) string {
	namePrefix := NamePrefix
	if infrasetting != "" {
		namePrefix += infrasetting + "-"
	}
	keycertname := namePrefix + sniHostName
	return Encode(keycertname+"-cacert", CACert)
}

func GetPoolPKIProfileName(poolName string) string {
	return Encode(poolName+"-pkiprofile", PKIProfile)
}

var VRFContext string
var VRFUuid string

func SetVrf(vrf string) {
	VRFContext = vrf
}

func SetVrfUuid(uuid string) {
	VRFUuid = uuid
}

func GetVrfUuid() string {
	if VRFUuid == "" {
		utils.AviLog.Warnf("VRF uuid not set")
	}
	return VRFUuid
}

func GetVrf() string {
	if VRFContext == "" {
		return utils.GlobalVRF
	}
	return VRFContext
}

func GetAdminTenant() string {
	return utils.ADMIN_NS
}

func GetTenant() string {
	tenantName := os.Getenv("TENANT_NAME")
	if tenantName != "" {
		return tenantName
	}
	return utils.ADMIN_NS
}

func IsIstioEnabled() bool {
	if ok, _ := strconv.ParseBool(os.Getenv("ISTIO_ENABLED")); ok {
		utils.AviLog.Debugf("Istio is enabled")
		return true
	}
	utils.AviLog.Debugf("Istio is not enabled")
	return false
}

func GetDefaultIngController() bool {
	defaultIngCtrl := os.Getenv("DEFAULT_ING_CONTROLLER")
	if defaultIngCtrl != "false" {
		return true
	}
	return false
}

func GetNamespaceToSync() string {
	namespace := os.Getenv("SYNC_NAMESPACE")
	if namespace != "" {
		return namespace
	}
	return ""
}

func GetEnableRHI() bool {
	if ok, _ := strconv.ParseBool(os.Getenv(ENABLE_RHI)); ok {
		utils.AviLog.Debugf("Enable RHI set to true")
		return true
	}
	utils.AviLog.Debugf("Enable RHI set to false")
	return false
}

func GetLabelToSyncNamespace() (string, string) {
	labelKey := os.Getenv("NAMESPACE_SYNC_LABEL_KEY")
	labelValue := os.Getenv("NAMESPACE_SYNC_LABEL_VALUE")

	if strings.Trim(labelKey, " ") != "" && strings.Trim(labelValue, " ") != "" {
		return labelKey, labelValue
	}
	return "", ""
}

// The port to run the AKO API server on
func GetAkoApiServerPort() string {
	port := os.Getenv("AKO_API_PORT")
	if port != "" {
		return port
	}
	// Default case, if not specified.
	return "8080"
}

var VipNetworkList []akov1alpha1.AviInfraSettingVipNetwork

func SetVipNetworkList(vipNetworks []akov1alpha1.AviInfraSettingVipNetwork) {
	VipNetworkList = vipNetworks
}

func GetVipNetworkList() []akov1alpha1.AviInfraSettingVipNetwork {
	return VipNetworkList
}

func GetVipNetworkListEnv() ([]akov1alpha1.AviInfraSettingVipNetwork, error) {
	var vipNetworkList []akov1alpha1.AviInfraSettingVipNetwork
	if IsWCP() {
		// do not return error in case of WCP deployments.
		return vipNetworkList, nil
	}
	vipNetworkListStr := os.Getenv(VIP_NETWORK_LIST)
	if vipNetworkListStr == "" || vipNetworkListStr == "null" {
		return vipNetworkList, fmt.Errorf("vipNetworkList not set in values.yaml")
	}

	err := json.Unmarshal([]byte(vipNetworkListStr), &vipNetworkList)
	if err != nil {
		utils.AviLog.Warnf("Unable to unmarshall json for vipNetworkList :%v", err)
		return vipNetworkList, fmt.Errorf("unable to unmarshall json for vipNetworkList")
	}

	// Only AWS cloud supports multiple VIP networks
	if GetCloudType() != CLOUD_AWS && len(vipNetworkList) > 1 {
		return nil, fmt.Errorf("more than one network specified in VIP Network List and Cloud type is not AWS")
	}
	return vipNetworkList, nil
}

func GetGlobalBgpPeerLabels() []string {
	var bgpPeerLabels []string
	bgpPeerLabelsStr := os.Getenv(BGP_PEER_LABELS)
	err := json.Unmarshal([]byte(bgpPeerLabelsStr), &bgpPeerLabels)
	if err != nil {
		utils.AviLog.Warnf("Unable to fetch the BGP Peer labels from environment variables.")
	}
	return bgpPeerLabels
}

func GetGlobalBlockedNSList() []string {
	var blockedNs []string
	blockedNSStr := os.Getenv(BLOCKED_NS_LIST)
	err := json.Unmarshal([]byte(blockedNSStr), &blockedNs)
	if err != nil {
		utils.AviLog.Warnf("Unable to fetch Blocked namespaces from environment variables. %v", err)
	}
	return blockedNs
}

func GetT1LRPath() string {
	return os.Getenv("NSXT_T1_LR")
}

var SEGroupName string

func SetSEGName(seg string) {
	SEGroupName = seg
}

func GetSEGName() string {
	return SEGroupName
}

func GetSEGNameEnv() string {
	segName := os.Getenv(SEG_NAME)
	if segName != "" {
		return segName
	}
	return ""
}

func GetNodeNetworkMap() (map[string][]string, error) {
	nodeNetworkMap := make(map[string][]string)
	type Row struct {
		NetworkName string   `json:"networkName"`
		Cidrs       []string `json:"cidrs"`
	}
	type nodeNetworkList []Row

	nodeNetworkListStr := os.Getenv(NODE_NETWORK_LIST)
	if nodeNetworkListStr == "" || nodeNetworkListStr == "null" {
		return nodeNetworkMap, fmt.Errorf("nodeNetworkList not set in values yaml")
	}
	var nodeNetworkListObj nodeNetworkList
	err := json.Unmarshal([]byte(nodeNetworkListStr), &nodeNetworkListObj)
	if err != nil {
		return nodeNetworkMap, fmt.Errorf("Unable to unmarshall json for nodeNetworkMap")
	}

	if len(nodeNetworkListObj) > NODE_NETWORK_MAX_ENTRIES {
		return nodeNetworkMap, fmt.Errorf("Maximum of %v entries are allowed for nodeNetworkMap", strconv.Itoa(NODE_NETWORK_MAX_ENTRIES))
	}

	for _, nodeNetwork := range nodeNetworkListObj {
		nodeNetworkMap[nodeNetwork.NetworkName] = nodeNetwork.Cidrs
	}

	return nodeNetworkMap, nil
}

func GetDomain() string {
	subDomain := os.Getenv(DEFAULT_DOMAIN)
	if subDomain != "" {
		return subDomain
	}
	return ""
}

// This utility returns a true/false depending on whether
// the user requires advanced L4 functionality
func GetAdvancedL4() bool {
	advanceL4 := os.Getenv(ADVANCED_L4)
	return advanceL4 == "true"
}

// Wrapper function for AKO running in either VDS
// or VCF (WCP with NSX).
func IsWCP() bool {
	if GetAdvancedL4() || utils.IsVCFCluster() {
		return true
	}
	return false
}

type NextPage struct {
	NextURI    string
	Collection interface{}
}

func FetchSEGroupWithMarkerSet(client *clients.AviClient, overrideUri ...NextPage) (error, string) {
	var uri string
	if len(overrideUri) == 1 {
		uri = overrideUri[0].NextURI
	} else {
		uri = "/api/serviceenginegroup/?include_name&page_size=100&cloud_ref.name=" + utils.CloudName
	}
	var result session.AviCollectionResult
	result, err := AviGetCollectionRaw(client, uri)
	if err != nil {
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			//SE in provider context no read access
			utils.AviLog.Debugf("Switching to admin context from  %s", GetTenant())
			SetAdminTenant := session.SetTenant(GetAdminTenant())
			SetTenant := session.SetTenant(GetTenant())
			SetAdminTenant(client.AviSession)
			defer SetTenant(client.AviSession)
			result, err = AviGetCollectionRaw(client, uri)
			if err != nil {
				utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
				return err, ""

			}
		} else {
			utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
			return err, ""
		}
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Errorf("Failed to unmarshal data, err: %v", err)
		return err, ""
	}

	// Using clusterID for advl4.
	clusterName := GetClusterID()
	for _, elem := range elems {
		seg := models.ServiceEngineGroup{}
		err = json.Unmarshal(elem, &seg)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
			continue
		}

		if len(seg.Markers) == 1 &&
			*seg.Markers[0].Key == ClusterNameLabelKey &&
			len(seg.Markers[0].Values) == 1 &&
			seg.Markers[0].Values[0] == clusterName {
			utils.AviLog.Infof("Marker configuration found in Service Engine Group %s.", *seg.Name)
			return nil, *seg.Name
		}
	}

	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/serviceenginegroup")
		if len(next_uri) > 1 {
			overrideUri := "/api/serviceenginegroup" + next_uri[1]
			nextPage := NextPage{NextURI: overrideUri}
			return FetchSEGroupWithMarkerSet(client, nextPage)
		}
	}

	utils.AviLog.Infof("No Marker configured Service Engine Group found.")
	return nil, ""
}

// This utility returns true if AKO is configured to create
// VS with Enhanced Virtual Hosting
func IsEvhEnabled() bool {
	evh := os.Getenv(ENABLE_EVH)
	if evh == "true" {
		return true
	}
	return utils.IsVCFCluster()
}

// If this flag is set to true, then AKO uses services API. Currently the support is limited for layer 4 Virtualservices
func UseServicesAPI() bool {
	if ok, _ := strconv.ParseBool(os.Getenv(SERVICES_API)); ok {
		return true
	}
	return false
}

// CompareVersions compares version v1 against version v2.
func CompareVersions(v1, cmpSign, v2 string) bool {
	if c, err := semver.NewConstraint(cmpSign + v2); err == nil {
		if currentVersion, err := semver.NewVersion(v1); err == nil && c.Check(currentVersion) {
			return true
		}
	}
	return false
}

func IsValidCni(returnErr *error) bool {
	// if serviceType is set as NodePortLocal, then the CNI must be of type 'antrea'
	if GetServiceType() == NodePortLocal && GetCNIPlugin() != ANTREA_CNI {
		*returnErr = fmt.Errorf("ServiceType is set as a NodePortLocal, but the CNI is not set as antrea")
		return false
	}
	return true
}

func GetDisableStaticRoute() bool {
	// We don't need the static routes for NSX-T cloud
	if IsWCP() || (GetCloudType() == CLOUD_NSXT && GetCNIPlugin() == NCP_CNI) {
		return true
	}
	if ok, _ := strconv.ParseBool(os.Getenv(DISABLE_STATIC_ROUTE_SYNC)); ok {
		return true
	}
	if IsNodePortMode() || GetServiceType() == NodePortLocal {
		return true
	}
	return false
}

func GetClusterName() string {
	if IsWCP() {
		return GetClusterIDSplit()
	}
	clusterName := os.Getenv(CLUSTER_NAME)
	if clusterName != "" {
		return clusterName
	}
	return ""
}

func GetClusterID() string {
	clusterID := os.Getenv(CLUSTER_ID)
	// The clusterID is an internal field only in the advanced L4 mode and we expect the format to be: domain-c8:3fb16b38-55f0-49fb-997d-c117487cd98d
	// We want to truncate this string to just have the uuid.
	if clusterID != "" {
		return clusterID
	}
	return ""
}

func GetClusterIDSplit() string {
	clusterID := GetClusterID()
	if clusterID != "" {
		clusterName := strings.Split(clusterID, ":")
		if len(clusterName) > 1 {
			return clusterName[0]
		}
	}
	return ""
}

func IsClusterNameValid() (bool, error) {
	clusterName := GetClusterName()
	re := regexp.MustCompile(`^[a-zA-Z0-9-_]*$`)
	if clusterName == "" {
		return false, fmt.Errorf("Required param clusterName not specified, syncing will be disabled")
	} else if !re.MatchString(clusterName) {
		return false, fmt.Errorf("clusterName must consist of alphanumeric characters or '-'/'_' (max 32 chars), syncing will be disabled")
	}
	return true, nil
}

var StaticRouteSyncChan chan struct{}
var ConfigDeleteSyncChan chan struct{}

var akoApi api.ApiServerInterface

func SetStaticRouteSyncHandler() {
	StaticRouteSyncChan = make(chan struct{})
}
func SetConfigDeleteSyncChan() {
	ConfigDeleteSyncChan = make(chan struct{})
}

func SetApiServerInstance(akoApiInstance api.ApiServerInterface) {
	akoApi = akoApiInstance
}

func ShutdownApi() {
	akoApi.ShutDown()
}

var clusterLabelChecksum uint32
var clusterKey string
var clusterValue string

func SetClusterLabelChecksum() {
	labels := GetLabels()
	clusterKey = *labels[0].Key
	clusterValue = *labels[0].Value
	clusterLabelChecksum = utils.Hash(clusterKey + clusterValue)
}

func GetClusterLabelChecksum() uint32 {
	return clusterLabelChecksum
}

func GetMarkersChecksum(markers utils.AviObjectMarkers) uint32 {
	vals := reflect.ValueOf(markers)
	var j int
	var cksum uint32
	var combinedString string
	numMarkerFields := vals.NumField()
	typeOfVals := vals.Type()
	markersStr := make([]string, numMarkerFields)
	for i := 0; i < numMarkerFields; i++ {
		if vals.Field(i).Interface() != "" {
			field := typeOfVals.Field(i).Name
			var value string
			if field == "Path" || field == "IngressName" || field == "Host" {
				pathArr := vals.Field(i).Interface().([]string)
				sort.Strings(pathArr)
				value = strings.Join(pathArr, "-")
			} else {
				value = vals.Field(i).Interface().(string)
			}
			markersStr[j] = value
			j = j + 1
		}
	}
	sort.Strings(markersStr)
	for _, ele := range markersStr {
		combinedString += ele
	}
	if len(combinedString) != 0 {
		cksum = utils.Hash(combinedString)
	}
	cksum += clusterLabelChecksum
	return cksum
}

func ObjectLabelChecksum(objectLabels []*models.RoleFilterMatchLabel) uint32 {
	var objChecksum uint32
	//Assumption here is User is not adding additional marker fields from UI/CLI
	//other than internal structure defined.
	markersStr := make([]string, len(objectLabels))
	var j int
	var combinedString string
	//For shared objects, checksum will be of only cluster label
	for _, label := range objectLabels {
		if *label.Key == clusterKey {
			if label.Values != nil && len(label.Values) > 0 && label.Values[0] == clusterValue {
				objChecksum += clusterLabelChecksum
			}
		} else {
			if len(label.Values) != 0 {
				markersStr[j] = label.Values[0]
				j = j + 1
			}
		}
	}
	sort.Strings(markersStr)
	for _, ele := range markersStr {
		combinedString += ele
	}
	if len(combinedString) != 0 {
		objChecksum += utils.Hash(combinedString)
	}
	return objChecksum
}

func VrfChecksum(vrfName string, staticRoutes []*models.StaticRoute) uint32 {
	clusterName := GetClusterName()
	filteredStaticRoutes := []*models.StaticRoute{}
	for _, staticRoute := range staticRoutes {
		if strings.HasPrefix(*staticRoute.RouteID, clusterName) {
			filteredStaticRoutes = append(filteredStaticRoutes, staticRoute)
		}
	}
	return utils.Hash(utils.Stringify(filteredStaticRoutes))
}

func DSChecksum(pgrefs []string, markers []*models.RoleFilterMatchLabel, populateCache bool) uint32 {
	sort.Strings(pgrefs)
	checksum := utils.Hash(utils.Stringify(pgrefs))
	if populateCache {
		if markers != nil {
			checksum += ObjectLabelChecksum(markers)
		}
		return checksum
	}
	checksum += GetClusterLabelChecksum()
	return checksum
}

func GetAnalyticsPolicyChecksum(analyticsPolicy *models.AnalyticsPolicy) uint32 {
	return utils.Hash(utils.Stringify(analyticsPolicy)) + GetClusterLabelChecksum()
}

func PopulatePoolNodeMarkers(namespace, host, infraSettingName, serviceName string, ingName, path []string) utils.AviObjectMarkers {
	var markers utils.AviObjectMarkers
	markers.Namespace = namespace
	markers.Host = []string{host}
	markers.Path = path
	markers.IngressName = ingName
	markers.InfrasettingName = infraSettingName
	markers.ServiceName = serviceName
	return markers
}
func PopulateVSNodeMarkers(namespace, host, infraSettingName string) utils.AviObjectMarkers {
	var markers utils.AviObjectMarkers
	markers.Namespace = namespace
	markers.Host = []string{host}
	markers.InfrasettingName = infraSettingName
	return markers
}

func PopulateHTTPPolicysetNodeMarkers(namespace, host, infraSettingName string, ingName, path []string) utils.AviObjectMarkers {
	var markers utils.AviObjectMarkers
	markers.Namespace = namespace
	markers.Host = []string{host}
	markers.IngressName = ingName
	markers.Path = path
	markers.InfrasettingName = infraSettingName
	return markers
}

func PopulateL4VSNodeMarkers(namespace, serviceName string) utils.AviObjectMarkers {
	var markers utils.AviObjectMarkers
	markers.Namespace = namespace
	markers.ServiceName = serviceName
	return markers
}

func PopulateL4PolicysetMarkers(namespace, serviceName string, protocols string) utils.AviObjectMarkers {
	var markers utils.AviObjectMarkers
	markers.Namespace = namespace
	markers.ServiceName = serviceName
	markers.Protocol = protocols
	return markers
}

func PopulateAdvL4VSNodeMarkers(namespace, gatewayName string) utils.AviObjectMarkers {
	var markers utils.AviObjectMarkers
	markers.Namespace = namespace
	markers.GatewayName = gatewayName
	return markers
}

func PopulateAdvL4PoolNodeMarkers(namespace, svcName, gatewayName string, port int) utils.AviObjectMarkers {
	var markers utils.AviObjectMarkers
	markers.Namespace = namespace
	markers.GatewayName = gatewayName
	markers.ServiceName = svcName
	markers.Port = strconv.Itoa(port)
	return markers
}

func PopulateSvcApiL4PoolNodeMarkers(namespace, svcName, gatewayName, protocol string, port int) utils.AviObjectMarkers {
	var markers utils.AviObjectMarkers
	markers.Namespace = namespace
	markers.GatewayName = gatewayName
	markers.ServiceName = svcName
	markers.Protocol = protocol
	markers.Port = strconv.Itoa(port)
	return markers
}

func PopulatePGNodeMarkers(namespace, host, infraSettingName string, ingName, path []string) utils.AviObjectMarkers {
	var markers utils.AviObjectMarkers
	markers.Namespace = namespace
	markers.Host = []string{host}
	markers.Path = path
	markers.IngressName = ingName
	markers.InfrasettingName = infraSettingName
	return markers
}

func PopulateTLSKeyCertNode(host, infraSettingName string) utils.AviObjectMarkers {
	var markers utils.AviObjectMarkers
	markers.Host = []string{host}
	markers.InfrasettingName = infraSettingName
	return markers
}

func PopulateL4PoolNodeMarkers(namespace, svcName, port string) utils.AviObjectMarkers {
	var markers utils.AviObjectMarkers
	markers.Namespace = namespace
	markers.ServiceName = svcName
	markers.Port = port
	return markers
}

func PopulatePassthroughPGMarkers(host, infrasettingName string) utils.AviObjectMarkers {
	var markers utils.AviObjectMarkers
	markers.Host = []string{host}
	markers.InfrasettingName = infrasettingName
	return markers
}

func PopulatePassthroughPoolMarkers(host, svcName, infrasettingName string) utils.AviObjectMarkers {
	var markers utils.AviObjectMarkers
	markers.Host = []string{host}
	markers.ServiceName = svcName
	markers.InfrasettingName = infrasettingName
	return markers
}

func InformersToRegister(kclient *kubernetes.Clientset, oclient *oshiftclient.Clientset) ([]string, error) {
	var isOshift bool
	// Initialize the following informers in all AKO deployments. Provide AKO the ability to watch over
	// Services, Endpoints, Secrets, ConfigMaps and Namespaces.
	allInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.SecretInformer,
		utils.ConfigMapInformer,
		utils.NSInformer,
	}

	// AKO must watch over Pods in case of NodePortLocal, to get Antrea annotation values.
	if GetServiceType() == NodePortLocal {
		allInformers = append(allInformers, utils.PodInformer)
	}

	// Watch over Ingresses for AKO deployment in WCP with NSX.
	if utils.IsVCFCluster() {
		allInformers = append(allInformers, utils.IngressInformer)
		allInformers = append(allInformers, utils.IngressClassInformer)
	}

	// For all deployments excluding AKO in WCP, watch over
	// Nodes, Ingresses, IngressClasses, Routes, MultiClusterIngress and ServiceImports.
	// Routes should be watched over in Openshift environments only.
	// MultiClusterIngress and ServiceImport should be watched over only when MCI is enabled.
	if !IsWCP() {
		allInformers = append(allInformers, utils.NodeInformer)

		informerTimeout := int64(120)
		_, err := kclient.CoreV1().Services(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{TimeoutSeconds: &informerTimeout})
		if err != nil {
			return allInformers, errors.New("Error in fetching services: " + err.Error())
		}
		if oclient != nil {
			_, err = oclient.RouteV1().Routes(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{TimeoutSeconds: &informerTimeout})
			if err == nil {
				// Openshift cluster with route support, we will just add route informer
				allInformers = append(allInformers, utils.RouteInformer)
				isOshift = true
			}
		}
		if !isOshift {
			allInformers = append(allInformers, utils.IngressInformer)
			allInformers = append(allInformers, utils.IngressClassInformer)
		}

		// Add MultiClusterIngress and ServiceImport informers if enabled.
		if utils.IsMultiClusterIngressEnabled() {
			allInformers = append(allInformers, utils.MultiClusterIngressInformer)
			allInformers = append(allInformers, utils.ServiceImportInformer)
		}
	}

	return allInformers, nil
}

func GetDiffPath(storedPathSvc map[string][]string, currentPathSvc map[string][]string) map[string][]string {
	pathSvcCopy := make(map[string][]string)
	for k, v := range storedPathSvc {
		pathSvcCopy[k] = v
	}

	for path, services := range currentPathSvc {
		// for OshiftRouteModel service diff is always checked
		storedServices, ok := pathSvcCopy[path]
		if ok {
			pathSvcCopy[path] = Difference(storedServices, services)
			if len(pathSvcCopy[path]) == 0 {
				delete(pathSvcCopy, path)
			}
		}
	}
	return pathSvcCopy
}

func SSLKeyCertChecksum(sslName, certificate, cacert string, ingestionMarkers utils.AviObjectMarkers, markers []*models.RoleFilterMatchLabel, populateCache bool) uint32 {
	checksum := utils.Hash(sslName + certificate + cacert)
	if populateCache {
		if markers != nil {
			checksum += ObjectLabelChecksum(markers)
		}
		return checksum
	}
	checksum += GetMarkersChecksum(ingestionMarkers)
	return checksum
}

func L4PolicyChecksum(ports []int64, protocols []string, ingestionMarkers utils.AviObjectMarkers, markers []*models.RoleFilterMatchLabel, populateCache bool) uint32 {
	var portsInt []int
	for _, port := range ports {
		portsInt = append(portsInt, int(port))
	}
	sort.Ints(portsInt)
	sort.Strings(protocols)
	checksum := utils.Hash(utils.Stringify(portsInt)) + utils.Hash(utils.Stringify(protocols))
	if populateCache {
		if markers != nil {
			checksum += ObjectLabelChecksum(markers)
		}
		return checksum
	}
	checksum += GetMarkersChecksum(ingestionMarkers)
	return checksum
}

func IsNodePortMode() bool {
	nodePortType := os.Getenv(SERVICE_TYPE)
	if nodePortType == NODE_PORT {
		return true
	}
	return false
}

// ToDo: Set the Service Type only once. But this creates a problem in UTs,
// because different types of Services needs to be tested in the UTs.
func GetServiceType() string {
	return os.Getenv(SERVICE_TYPE)
}

// AutoAnnotateNPLSvc returns true if AKO is automatically annotating required Services instead of user for NPL
func AutoAnnotateNPLSvc() bool {
	autoAnnotateSvc := os.Getenv(autoAnnotateService)
	if GetServiceType() == NodePortLocal && !strings.EqualFold(autoAnnotateSvc, "false") {
		return true
	}
	return false
}

func GetNodePortsSelector() map[string]string {
	nodePortsSelectorLabels := make(map[string]string)
	if IsNodePortMode() {
		// If the key/values are kept empty then we select all nodes
		nodePortsSelectorLabels["key"] = os.Getenv(NODE_KEY)
		nodePortsSelectorLabels["value"] = os.Getenv(NODE_VALUE)
	}
	return nodePortsSelectorLabels
}

func IsValidLabelOnNode(labels map[string]string, key string) bool {
	nodeFilter := GetNodePortsSelector()
	//If nodefilter is not mentioned or key is empty in values.yaml, node is valid
	if len(nodeFilter) != 2 || nodeFilter["key"] == "" {
		return true
	}
	val, ok := labels[nodeFilter["key"]]
	if !ok || val != nodeFilter["value"] {
		utils.AviLog.Debugf("key: %s, msg: node does not have valid label", key)
		return false
	}
	return true
}

var CloudType string
var CloudUUID string

func SetCloudType(cloudType string) {
	CloudType = cloudType
}

func SetCloudUUID(cloudUUID string) {
	CloudUUID = cloudUUID
}

func GetCloudType() string {
	if CloudType == "" {
		return CLOUD_VCENTER
	}
	return CloudType
}

func GetCloudUUID() string {
	return CloudUUID
}

var IsCloudInAdminTenant = true

func SetIsCloudInAdminTenant(isCloudInAdminTenant bool) {
	IsCloudInAdminTenant = isCloudInAdminTenant
}

func IsPublicCloud() bool {
	cloudType := GetCloudType()
	if cloudType == CLOUD_AZURE || cloudType == CLOUD_AWS ||
		cloudType == CLOUD_GCP || cloudType == CLOUD_OPENSTACK {
		return true
	}
	return false
}

func IsNodeNetworkAllowedCloud() bool {
	cloudType := GetCloudType()
	if (cloudType == CLOUD_NSXT && GetNSXTTransportZone() == VLAN_TRANSPORT_ZONE) ||
		cloudType == CLOUD_VCENTER {
		return true
	}
	return false
}

func UsesNetworkRef() bool {
	cloudType := GetCloudType()
	if cloudType == CLOUD_AWS || cloudType == CLOUD_OPENSTACK ||
		cloudType == CLOUD_AZURE {
		return true
	}
	return false
}

func PassthroughShardSize() uint32 {
	shardVsSize := os.Getenv("PASSTHROUGH_SHARD_SIZE")
	shardSize, ok := ShardSizeMap[shardVsSize]
	if ok {
		return shardSize
	}
	return 1
}
func GetAKOIDPrefix() string {
	var akoID string
	isPrimaryAKO := AKOControlConfig().GetAKOInstanceFlag()
	if !isPrimaryAKO {
		akoID = os.Getenv("POD_NAMESPACE") + "-"
	}
	return akoID
}

// TODO: Optimize
func IsNamespaceBlocked(namespace string) bool {
	nsBlockedList := AKOControlConfig().GetAKOBlockedNSList()
	_, ok := nsBlockedList[namespace]
	return ok
}

func GetPassthroughShardVSName(s, aviInfraSettingName, key string, shardSize uint32) string {
	var vsNum uint32
	shardVsPrefix := GetClusterName() + "--" + GetAKOIDPrefix() + PassthroughPrefix
	vsNum = utils.Bkt(s, shardSize)
	if aviInfraSettingName != "" {
		shardVsPrefix += aviInfraSettingName + "-"
	}
	vsName := shardVsPrefix + strconv.Itoa(int(vsNum))
	utils.AviLog.Infof("key: %s, msg: Passthrough ShardVSName: %s", key, vsName)
	return vsName
}

// GetLabels returns the key value pair used for tagging the segroups and routes in vrfcontext
func GetLabels() []*models.KeyValue {
	clusterName := GetClusterName()
	labelKey := ClusterNameLabelKey
	kv := &models.KeyValue{
		Key:   &labelKey,
		Value: &clusterName,
	}
	labels := []*models.KeyValue{}
	labels = append(labels, kv)
	return labels
}

// GetMarkers returns the key values pair used for tagging the segroups and routes in vrfcontext
func GetAllMarkers(markers utils.AviObjectMarkers) []*models.RoleFilterMatchLabel {
	clusterName := GetClusterName()
	labelKey := ClusterNameLabelKey
	rfml := &models.RoleFilterMatchLabel{
		Key:    &labelKey,
		Values: []string{clusterName},
	}
	rfmls := []*models.RoleFilterMatchLabel{}
	rfmls = append(rfmls, rfml)

	vals := reflect.ValueOf(markers)

	typeOfVals := vals.Type()

	for i := 0; i < vals.NumField(); i++ {
		if vals.Field(i).Interface() != "" {
			field := typeOfVals.Field(i).Name
			value := vals.Field(i).Interface()
			if field == "Path" || field == "IngressName" || field == "Host" {
				values := value.([]string)
				if len(values) == 0 {
					continue
				}
				rfml = &models.RoleFilterMatchLabel{
					Key:    &field,
					Values: values,
				}

			} else {
				rfml = &models.RoleFilterMatchLabel{
					Key:    &field,
					Values: []string{value.(string)},
				}
			}
			rfmls = append(rfmls, rfml)
		}
	}
	return rfmls
}
func GetMarkers() []*models.RoleFilterMatchLabel {
	clusterName := GetClusterName()
	labelKey := ClusterNameLabelKey
	rfml := &models.RoleFilterMatchLabel{
		Key:    &labelKey,
		Values: []string{clusterName},
	}
	rfmls := []*models.RoleFilterMatchLabel{}
	rfmls = append(rfmls, rfml)
	return rfmls
}

func HasValidBackends(routeSpec routev1.RouteSpec, routeName, namespace, key string) bool {
	svcList := make(map[string]bool)
	toSvc := routeSpec.To.Name
	svcList[toSvc] = true
	for _, altBackend := range routeSpec.AlternateBackends {
		if _, found := svcList[altBackend.Name]; found {
			utils.AviLog.Warnf("key: %s, msg: multiple backends with name %s found for route: %s", key, altBackend.Name, routeName)
			return false
		}
		svcList[altBackend.Name] = true
	}
	return true
}

func ContainsFinalizer(o metav1.Object, finalizer string) bool {
	f := o.GetFinalizers()
	for _, e := range f {
		if e == finalizer {
			return true
		}
	}
	return false
}

func GetDefaultSecretForRoutes() string {
	return DefaultRouteCert
}

func ValidateIngressForClass(key string, ingress *networkingv1.Ingress) bool {
	if utils.IsVCFCluster() {
		return true
	}
	// see whether ingress class resources are present or not
	if ingress.Spec.IngressClassName == nil {
		// check whether avi-lb ingress class is set as the default ingress class
		if _, found := IsAviLBDefaultIngressClass(); found {
			utils.AviLog.Infof("key: %s, msg: ingress class name is not specified but ako.vmware.com/avi-lb is default ingress controller", key)
			return true
		} else {
			utils.AviLog.Warnf("key: %s, msg: ingress class name not specified for ingress %s and ako.vmware.com/avi-lb is not default ingress controller", key, ingress.Name)
			return false
		}
	}

	// If the key is "syncstatus" then use a clientset, else using the lister cache. This is because
	// the status sync happens before the informers are run and caches are synced.
	var ingClassObj *networkingv1.IngressClass
	var err error
	if key == SyncStatusKey {
		ingClassObj, err = utils.GetInformers().ClientSet.NetworkingV1().IngressClasses().Get(context.TODO(), *ingress.Spec.IngressClassName, metav1.GetOptions{})
	} else {
		ingClassObj, err = utils.GetInformers().IngressClassInformer.Lister().Get(*ingress.Spec.IngressClassName)
	}
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Unable to fetch corresponding networking.k8s.io/ingressclass %s %v",
			key, *ingress.Spec.IngressClassName, err)
		return false
	}

	// Additional check to see if the ingressclass is a valid avi ingress class or not.
	if ingClassObj.Spec.Controller != AviIngressController {
		// Return an error since this is not our object.
		utils.AviLog.Warnf("key: %s, msg: Unexpected controller in ingress class %s", key, *ingress.Spec.IngressClassName)
		return false
	}

	return true
}

func IsAviLBDefaultIngressClass() (string, bool) {
	ingClassObjs, _ := utils.GetInformers().IngressClassInformer.Lister().List(labels.Set(nil).AsSelector())
	for _, ingClass := range ingClassObjs {
		if ingClass.Spec.Controller == AviIngressController {
			annotations := ingClass.GetAnnotations()
			isDefaultClass, ok := annotations[DefaultIngressClassAnnotation]
			if ok && isDefaultClass == "true" {
				return ingClass.Name, true
			}
		}
	}

	utils.AviLog.Debugf("IngressClass with controller ako.vmware.com/avi-lb not found in the cluster")
	return "", false
}

func IsAviLBDefaultIngressClassWithClient() (string, bool) {
	ingClassObjs, _ := utils.GetInformers().IngressClassInformer.Lister().List(labels.Set(nil).AsSelector())
	for _, ingClass := range ingClassObjs {
		if ingClass.Spec.Controller == AviIngressController {
			annotations := ingClass.GetAnnotations()
			isDefaultClass, ok := annotations[DefaultIngressClassAnnotation]
			if ok && isDefaultClass == "true" {
				return ingClass.Name, true
			}
		}
	}

	utils.AviLog.Debugf("IngressClass with controller ako.vmware.com/avi-lb not found in the cluster")
	return "", false
}

func GetAviSecretWithRetry(kc kubernetes.Interface, retryCount int, secret string) (*v1.Secret, error) {
	var aviSecret *v1.Secret
	var err error
	for retry := 0; retry < retryCount; retry++ {
		aviSecret, err = kc.CoreV1().Secrets(utils.GetAKONamespace()).Get(context.TODO(), secret, metav1.GetOptions{})
		if err == nil {
			return aviSecret, nil
		}
		utils.AviLog.Warnf("Failed to get avi-secret, retry count:%d, err: %+v", retry, err)
	}
	return nil, err
}

func UpdateAviSecretWithRetry(kc kubernetes.Interface, aviSecret *v1.Secret, retryCount int) error {
	var err error
	for retry := 0; retry < retryCount; retry++ {
		_, err = kc.CoreV1().Secrets(utils.GetAKONamespace()).Update(context.TODO(), aviSecret, metav1.UpdateOptions{})
		if err == nil {
			return nil
		}
		utils.AviLog.Warnf("Failed to update avi-secret, retry count:%d, err: %+v", retry, err)
	}
	return err
}

func RefreshAuthToken(kc kubernetes.Interface) {
	retryCount := 5
	ctrlProp := utils.SharedCtrlProp().GetAllCtrlProp()
	ctrlUsername := ctrlProp[utils.ENV_CTRL_USERNAME]
	ctrlAuthToken := ctrlProp[utils.ENV_CTRL_AUTHTOKEN]
	ctrlCAData := ctrlProp[utils.ENV_CTRL_CADATA]
	ctrlIpAddress := GetControllerIP()
	oldTokenID := ""

	aviClient := NewAviRestClientWithToken(ctrlIpAddress, ctrlUsername, ctrlAuthToken, ctrlCAData)
	if aviClient == nil {
		utils.AviLog.Errorf("Failed to initialize AVI client")
		return
	}
	tokens := make(map[string]interface{})
	err := utils.GetAuthTokenMapWithRetry(aviClient, tokens, retryCount)
	if err != nil {
		utils.AviLog.Errorf("Failed to get existing tokens from controller, err: %+v", err)
		return
	}
	aviToken, ok := tokens[ctrlAuthToken]
	if ok {
		expiry, ok := aviToken.(map[string]interface{})["expires_at"].(string)
		if !ok {
			utils.AviLog.Errorf("Failed to parse token object")
			return
		}
		layout := "2006-01-02T15:04:05.000000+00:00"
		expiryTime, err := time.Parse(layout, expiry)
		if err != nil {
			utils.AviLog.Errorf("Unable to parse token expiry time, err: %+v", err)
			return
		}
		if time.Until(expiryTime) > (utils.RefreshAuthTokenPeriod*utils.AuthTokenExpiry)*time.Hour {
			utils.AviLog.Infof("Skipping AuthToken Refresh")
			return
		}
		oldTokenID, _ = aviToken.(map[string]interface{})["uuid"].(string)
	}

	newTokenResp, err := utils.CreateAuthTokenWithRetry(aviClient, retryCount)
	if err != nil {
		utils.AviLog.Errorf("Failed to post new token, err: %+v", err)
		return
	}
	if _, ok := newTokenResp.(map[string]interface{}); !ok {
		utils.AviLog.Errorf("Failed to parse new token, err: %+v", err)
		return
	}
	token := newTokenResp.(map[string]interface{})["token"].(string)
	var secrets []string
	if utils.IsVCFCluster() {
		secrets = []string{
			AviInitSecret,
			AviSecret,
		}
	} else {
		secrets = []string{
			AviSecret,
		}
	}
	for _, secret := range secrets {
		aviSecret, err := GetAviSecretWithRetry(kc, retryCount, secret)
		if err != nil {
			utils.AviLog.Errorf("Failed to get secret, err: %+v", err)
			return
		}
		aviSecret.Data["authtoken"] = []byte(token)

		err = UpdateAviSecretWithRetry(kc, aviSecret, retryCount)
		if err != nil {
			utils.AviLog.Errorf("Failed to update secret, err: %+v", err)
			return
		}
	}
	utils.AviLog.Infof("Successfully updated authtoken")
	if oldTokenID != "" {
		err = utils.DeleteAuthTokenWithRetry(aviClient, oldTokenID, retryCount)
		if err != nil {
			utils.AviLog.Warnf("Failed to delete old token %s, err: %+v", oldTokenID, err)
		}
	}
}

func GetControllerPropertiesFromSecret(cs kubernetes.Interface) (map[string]string, error) {
	ctrlProps := make(map[string]string)
	aviSecret, err := cs.CoreV1().Secrets(utils.GetAKONamespace()).Get(context.TODO(), AviSecret, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Error(err, err.Error())
		return ctrlProps, err
	}
	ctrlProps[utils.ENV_CTRL_USERNAME] = string(aviSecret.Data["username"])
	if aviSecret.Data["password"] != nil {
		ctrlProps[utils.ENV_CTRL_PASSWORD] = string(aviSecret.Data["password"])
	} else {
		ctrlProps[utils.ENV_CTRL_PASSWORD] = ""
	}
	if aviSecret.Data["authtoken"] != nil {
		ctrlProps[utils.ENV_CTRL_AUTHTOKEN] = string(aviSecret.Data["authtoken"])
	} else {
		ctrlProps[utils.ENV_CTRL_AUTHTOKEN] = ""
	}
	if aviSecret.Data["certificateAuthorityData"] != nil {
		ctrlProps[utils.ENV_CTRL_CADATA] = string(aviSecret.Data["certificateAuthorityData"])
	} else {
		ctrlProps[utils.ENV_CTRL_CADATA] = ""
	}
	return ctrlProps, nil
}

func GetVCFNetworkName() string {
	return VCF_NETWORK + "-" + GetClusterID()
}

var (
	aviMinVersion = ""
	aviMaxVersion = ""
	k8sMinVersion = ""
	k8sMaxVersion = ""
)

func GetAviMinSupportedVersion() string {
	return aviMinVersion
}

func GetAviMaxSupportedVersion() string {
	return aviMaxVersion
}

func GetK8sMinSupportedVersion() string {
	return k8sMinVersion
}

func GetK8sMaxSupportedVersion() string {
	return k8sMaxVersion
}

func GetControllerVersion() string {
	controllerVersion := AKOControlConfig().ControllerVersion()
	// Ensure that the controllerVersion is less than the supported Avi maxVersion and more than minVersion.
	if CompareVersions(controllerVersion, ">", GetAviMaxSupportedVersion()) {
		utils.AviLog.Infof("Setting the client version to AVI Max supported version %s", GetAviMaxSupportedVersion())
		controllerVersion = GetAviMaxSupportedVersion()
	}
	if CompareVersions(controllerVersion, "<", GetAviMinSupportedVersion()) {
		AKOControlConfig().PodEventf(
			corev1.EventTypeWarning,
			AKOShutdown, "AKO is running with unsupported Avi version %s",
			controllerVersion,
		)
		utils.AviLog.Fatalf("AKO is not supported for the Avi version %s, Avi must be %s or more", controllerVersion, GetAviMinSupportedVersion())
	}
	return controllerVersion
}

func VIPPerNamespace() bool {
	vipPerNS := os.Getenv(VIP_PER_NAMESPACE)
	if vipPerNS == "true" {
		return true
	}
	return utils.IsVCFCluster()
}

var controllerIP string

func GetControllerIP() string {
	if controllerIP == "" {
		SetControllerIP(os.Getenv(utils.ENV_CTRL_IPADDRESS))
	}
	return controllerIP
}

func SetControllerIP(ctrlIP string) {
	controllerIP = ctrlIP
}

var VCFInitialized bool
var AviSecretInitialized bool
var AviSEInitialized bool

var throttle = map[string]uint32{
	"HIGH":     10,
	"MEDIUM":   30,
	"LOW":      50,
	"DISABLED": 0,
}

func GetThrottle(key string) *int32 {
	throttle := int32(throttle[key])
	return &throttle
}

func UpdateV6(vip *models.Vip, vipNetwork *akov1alpha1.AviInfraSettingVipNetwork) {
	if vipNetwork.Cidr != "" {
		vip.AutoAllocateIPType = proto.String("V4_V6")
	} else {
		vip.AutoAllocateIPType = proto.String("V6_ONLY")
	}
}

var IPfamily string

func SetIPFamily() {
	ipFamily := os.Getenv(IP_FAMILY)
	if GetCloudType() == CLOUD_VCENTER {
		if ipFamily != "" {
			utils.AviLog.Debugf("ipFamily is set to %s", ipFamily)
			IPfamily = ipFamily
			return
		} else {
			utils.AviLog.Debugf("ipFamily is not set, default is V4")
			ipFamily = "V4"
		}
	} else {
		ipFamily = "V4"
	}
	IPfamily = ipFamily
}

func GetIPFamily() string {
	if IPfamily == "" {
		SetIPFamily()
	}
	return IPfamily
}

func IsValidV6Config(returnErr *error) bool {
	ipFamily := GetIPFamily()
	if !(ipFamily == "V4" || ipFamily == "V6") {
		*returnErr = fmt.Errorf("ipFamily is not one of (V4, V6)")
		return false
	}

	vipNetworkList := GetVipNetworkList()
	isCloudVCenter := (GetCloudType() == CLOUD_VCENTER)
	for _, vipNetwork := range vipNetworkList {
		if !isCloudVCenter && vipNetwork.V6Cidr != "" {
			*returnErr = fmt.Errorf("IPv6 CIDR is only supported for vCenter Clouds")
			return false
		}
	}
	return true
}

func CreateIstioSecretFromCert(name string, kc kubernetes.Interface) {

	fileData, err := os.ReadFile(name)
	if err != nil {
		utils.AviLog.Warnf("%s", err)
		return
	}
	data := make(map[string][]byte)
	var istioSecret *v1.Secret
	istioSecret, err = kc.CoreV1().Secrets(utils.GetAKONamespace()).Get(context.TODO(), IstioSecret, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Infof("%s not found, creating new empty secret", IstioSecret)
		istioSecret, err = kc.CoreV1().Secrets(utils.GetAKONamespace()).Create(context.TODO(), &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: IstioSecret,
			},
			Data: data,
			Type: corev1.SecretTypeOpaque,
		}, metav1.CreateOptions{})
		if err != nil {
			utils.AviLog.Warnf("Failed to create %s %s", IstioSecret, err.Error())
			return
		}
	}
	nameSplit := strings.Split(name, "/")
	dataName := nameSplit[len(nameSplit)-1]
	dataName = dataName[:len(dataName)-4]
	if istioSecret.Data == nil {
		istioSecret.Data = make(map[string][]byte)
	}
	istioSecret.Data[dataName] = fileData
	_, err = kc.CoreV1().Secrets(utils.GetAKONamespace()).Update(context.TODO(), istioSecret, metav1.UpdateOptions{})

	if err != nil {
		utils.AviLog.Warnf("Failed to update %s %s", IstioSecret, err.Error())
		return
	}
	updateIstioCertSet(name)
	utils.AviLog.Infof("Updated %s with resource %s", IstioSecret, name)
}

var istioInitialized bool

func SetIstioInitialized(b bool) {
	istioInitialized = b
}

func IsIstioInitialized() bool {
	return istioInitialized
}

func IsIstioKey(key string) bool {
	return strings.HasPrefix(key, "istio-pki-") || strings.HasPrefix(key, "istio-workload-")
}

func GetIstioPKIProfileName() string {
	return "istio-pki-" + GetClusterName() + "-" + utils.GetAKONamespace()
}

func GetIstioWorkloadCertificateName() string {
	return "istio-workload-" + GetClusterName() + "-" + utils.GetAKONamespace()
}

var istioCertSet sets.String

func updateIstioCertSet(s string) {
	if istioCertSet == nil {
		istioCertSet = sets.NewString()
	}
	istioCertSet.Insert(s)
}

func GetIstioCertSet() sets.String {
	return istioCertSet
}

func IsChanClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

func GetIPFromNode(node *v1.Node) (string, string) {
	var nodeV4, nodeV6 string
	nodeAddrs := node.Status.Addresses
	for _, addr := range nodeAddrs {
		if addr.Type == corev1.NodeInternalIP {
			nodeIP := addr.Address
			if utils.IsV4(nodeIP) {
				nodeV4 = nodeIP
			} else {
				nodeV6 = nodeIP
			}
		}
	}
	if GetIPFamily() == "V6" {
		if GetCNIPlugin() == CALICO_CNI {
			if nodeIP, ok := node.Annotations["projectcalico.org/IPv4Address"]; ok {
				nodeV4 = strings.Split(nodeIP, "/")[0]
			}
			if nodeIP, ok := node.Annotations["projectcalico.org/IPv6Address"]; ok {
				nodeV6 = strings.Split(nodeIP, "/")[0]
			}
		}
		if GetCNIPlugin() == ANTREA_CNI {
			if nodeIPstr, ok := node.Annotations["node.antrea.io/transport-addresses"]; ok {
				nodeIPlist := strings.Split(nodeIPstr, ",")
				for _, nodeIP := range nodeIPlist {
					if utils.IsV4(nodeIP) {
						nodeV4 = nodeIP
					} else {
						nodeV6 = nodeIP
					}
				}
			}
		}
	}
	return nodeV4, nodeV6
}

func init() {
	seGroupToUse := os.Getenv(SEG_NAME)
	if seGroupToUse == "" {
		seGroupToUse = DEFAULT_SE_GROUP
	}
	SEGroupName = seGroupToUse
}
