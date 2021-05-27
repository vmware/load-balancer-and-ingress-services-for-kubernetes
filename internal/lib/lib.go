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
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/avinetworks/sdk/go/models"
	routev1 "github.com/openshift/api/route/v1"
	oshiftclient "github.com/openshift/client-go/route/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
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

var fqdnEnum = map[string]int32{
	"default": 1,
	"flat":    2,
}

var NamePrefix string

func SetNamePrefix() {
	NamePrefix = GetClusterName() + "--"
}

func GetNamePrefix() string {
	return NamePrefix
}

func Encode(s string) string {
	if !IsEvhEnabled() {
		return s
	}
	hash := sha1.Sum([]byte(s))
	return GetNamePrefix() + hex.EncodeToString(hash[:]) + AKO_ENCODED
}

var DisableSync bool
var layer7Only bool
var noPGForSNI, gRBAC bool

func SetDisableSync(state bool) {
	DisableSync = state
	utils.AviLog.Infof("Setting Disable Sync to: %v", state)
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

func SetGRBACSupport(val string) {
	if boolVal, err := strconv.ParseBool(val); err == nil {
		gRBAC = boolVal
	}
	controllerVersion := utils.CtrlVersion
	if gRBAC && CheckControllerVersionCompatibility(controllerVersion, "<", ControllerVersion2015) {
		// GRBAC is supported from 20.1.5 and above
		utils.AviLog.Infof("Disabling GRBAC as current controller version %s is less than %s.", controllerVersion, ControllerVersion2015)
		gRBAC = false
	}
}

func GetGRBACSupport() bool {
	return gRBAC
}

func GetNoPGForSNI() bool {
	return noPGForSNI
}

func GetLayer7Only() bool {
	return layer7Only
}

var AKOUser string

func SetAKOUser() {
	AKOUser = "ako-" + GetClusterName()
	utils.AviLog.Infof("Setting AKOUser: %s for Avi Objects", AKOUser)
}

func GetAKOUser() string {
	return AKOUser
}

var enableCtrl2014Features bool

func SetEnableCtrl2014Features(controllerVersion string) {
	enableCtrl2014Features = CheckControllerVersionCompatibility(controllerVersion, ">=", ControllerVersion2014)
}

func GetEnableCtrl2014Features() bool {
	return enableCtrl2014Features
}

func GetshardSize() uint32 {
	if GetAdvancedL4() {
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

func GetL4FqdnFormat() int32 {
	fqdnFormat := os.Getenv("AUTO_L4_FQDN")
	enumVal, ok := fqdnEnum[fqdnFormat]
	if ok {
		return enumVal
	} else {
		// If no match then disable FQDNs for L4.
		return 3
	}
}

func GetModelName(namespace, objectName string) string {
	return namespace + "/" + objectName
}

// All L4 object names.
func GetL4VSName(svcName, namespace string) string {
	return Encode(NamePrefix + namespace + "-" + svcName)
}

func GetL4VSVipName(svcName, namespace string) string {
	return Encode(NamePrefix + namespace + "-" + svcName)
}

func GetL4PoolName(svcName, namespace string, port int32) string {
	poolName := NamePrefix + namespace + "-" + svcName + "--" + strconv.Itoa(int(port))
	return Encode(poolName)
}

func GetAdvL4PoolName(svcName, namespace, gwName string, port int32) string {
	poolName := NamePrefix + namespace + "-" + svcName + "-" + gwName + "--" + strconv.Itoa(int(port))
	return Encode(poolName)
}

// All L7 object names.
func GetVsVipName(vsName string) string {
	return vsName
}

func GetL7InsecureDSName(vsName string) string {
	return vsName
}

func GetL7SharedPGName(vsName string) string {
	return vsName
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
	return Encode(poolName)
}

func GetL7HttpRedirPolicy(vsName string) string {
	return vsName
}

func GetHeaderRewritePolicy(vsName, localHost string) string {
	return vsName + "--host-hdr-re-write" + "--" + localHost
}

func GetSniNodeName(ingName, infrasetting, sniHostName string) string {
	namePrefix := NamePrefix
	if infrasetting != "" {
		namePrefix += infrasetting + "-"
	}
	return Encode(namePrefix + sniHostName)
}

func GetSniPoolName(ingName, namespace, host, path, infrasetting string, args ...string) string {
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
	return poolName
}

func GetSniHttpPolName(ingName, namespace, host, path, infrasetting string) string {
	path = strings.ReplaceAll(path, "/", "_")
	if infrasetting != "" {
		return Encode(NamePrefix + infrasetting + "-" + namespace + "-" + host + path + "-" + ingName)
	}
	return Encode(NamePrefix + namespace + "-" + host + path + "-" + ingName)
}

func GetSniPGName(ingName, namespace, host, path, infrasetting string) string {
	path = strings.ReplaceAll(path, "/", "_")
	if infrasetting != "" {
		return NamePrefix + infrasetting + "-" + namespace + "-" + host + path + "-" + ingName
	}
	return NamePrefix + namespace + "-" + host + path + "-" + ingName
}

// evh child
func GetEvhVsPoolNPgName(ingName, namespace, host, path, infrasetting string, args ...string) string {
	path = strings.ReplaceAll(path, "/", "_")
	namePrefix := NamePrefix
	if infrasetting != "" {
		namePrefix += infrasetting + "-"
	}
	poolName := namePrefix + namespace + "-" + host + path + "-" + ingName
	if len(args) > 0 {
		svcName := args[0]
		poolName = poolName + "-" + svcName
	}
	return Encode(poolName)
}
func GetEvhPoolName(ingName, namespace, host, path, infrasetting, svcName string) string {
	path = strings.ReplaceAll(path, "/", "_")
	namePrefix := NamePrefix
	if infrasetting != "" {
		namePrefix += infrasetting + "-"
	}
	poolName := namePrefix + namespace + "-" + host + path + "-" + ingName + "-" + svcName
	return poolName
}

func GetEvhNodeName(ingName, namespace, host, infrasetting string) string {
	if infrasetting != "" {
		return Encode(NamePrefix + infrasetting + "-" + namespace + "-" + host)
	}
	return Encode(NamePrefix + namespace + "-" + host)
}

func GetEvhPGName(ingName, namespace, host, path, infrasetting string) string {
	path = strings.ReplaceAll(path, "/", "_")
	if infrasetting != "" {
		return Encode(NamePrefix + infrasetting + "-" + namespace + "-" + host + path + "-" + ingName)
	}
	return Encode(NamePrefix + namespace + "-" + host + path + "-" + ingName)
}

func GetTLSKeyCertNodeName(infrasetting, sniHostName string) string {
	namePrefix := NamePrefix
	if infrasetting != "" {
		namePrefix += infrasetting + "-"
	}
	return Encode(namePrefix + sniHostName)
}

func GetCACertNodeName(infrasetting, sniHostName string) string {
	namePrefix := NamePrefix
	if infrasetting != "" {
		namePrefix += infrasetting + "-"
	}
	keycertname := namePrefix + sniHostName
	return Encode(keycertname + "-cacert")
}

func GetPoolPKIProfileName(poolName string) string {
	return Encode(poolName + "-pkiprofile")
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

func GetTenantsPerCluster() bool {
	tpc := os.Getenv("TENANTS_PER_CLUSTER")
	if tpc == "true" {
		return true
	}
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

func GetSubnetIP() string {
	subnetIP := os.Getenv(SUBNET_IP)
	if subnetIP != "" {
		return subnetIP
	}
	return ""
}

func GetSubnetPrefix() string {
	subnetPrefix := os.Getenv(SUBNET_PREFIX)
	if subnetPrefix != "" {
		return subnetPrefix
	}
	return ""
}

func GetSubnetPrefixInt() int32 {
	// check if subnetPrefix value is a valid integer value
	defaultCidr := int32(24)
	intCidr, err := strconv.ParseInt(GetSubnetPrefix(), 10, 32)
	if err != nil {
		utils.AviLog.Warnf("The value of subnetPrefix couldn't be converted to int32, defaulting to /24, %v", err)
		return defaultCidr
	}
	return int32(intCidr)
}

func GetVipNetworkList() ([]string, error) {
	var vipNetworkList []string
	type Row struct {
		NetworkName string `json:"networkName"`
	}
	type vipNetworkListRow []Row

	vipNetworkListStr := os.Getenv(VIP_NETWORK_LIST)
	if vipNetworkListStr == "" || vipNetworkListStr == "null" {
		return vipNetworkList, fmt.Errorf("vipNetworkList not set in values yaml")
	}

	var vipNetworkListObj vipNetworkListRow
	err := json.Unmarshal([]byte(vipNetworkListStr), &vipNetworkListObj)
	if err != nil {
		return vipNetworkList, fmt.Errorf("Unable to unmarshall json for vipNetworkListMap")
	}
	for _, subnet := range vipNetworkListObj {
		vipNetworkList = append(vipNetworkList, subnet.NetworkName)
	}
	if len(vipNetworkList) > 1 {
		// Only AWS cloud supports multiple VIP networks
		if GetCloudType() != CLOUD_AWS {
			return nil, fmt.Errorf("More than one network specified in VIP Network List and Cloud type is not AWS")
		}
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

func GetT1LRPath() string {
	t1LrPath := os.Getenv("NSXT_T1_LR")
	return t1LrPath
}

func GetSEGName() string {
	segName := os.Getenv(SEG_NAME)
	if segName != "" {
		return segName
	}
	if GetAdvancedL4() {
		return DEFAULT_SE_GROUP
	}
	return DEFAULT_SE_GROUP
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
		return nodeNetworkMap, fmt.Errorf("Maximum of %v entries are allowed for nodeNetworkMap", string(NODE_NETWORK_MAX_ENTRIES))
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
	if advanceL4 == "true" {

		return true
	}
	return false
}

// This utility returns true if AKO is configured to create
// VS with Enhanced Virtual Hosting
func IsEvhEnabled() bool {
	evh := os.Getenv(ENABLE_EVH)
	if evh == "true" {
		return true
	}
	return false
}

// If this flag is set to true, then AKO uses services API. Currently the support is limited for layer 4 Virtualservices
func UseServicesAPI() bool {
	if ok, _ := strconv.ParseBool(os.Getenv(SERVICES_API)); ok {
		return true
	}
	return false
}

//Here v1 is compared against v2
func CheckControllerVersionCompatibility(v1, cmpSign, v2 string) bool {
	if c, err := semver.NewConstraint(cmpSign + v2); err == nil {
		if currentVersion, err := semver.NewVersion(v1); err == nil && c.Check(currentVersion) {
			return true
		}
	}
	return false
}

func IsValidCni() bool {
	// if serviceType is set as NodePortLocal, then the CNI must be of type 'antrea'
	if GetServiceType() == NodePortLocal && GetCNIPlugin() != ANTREA_CNI {
		utils.AviLog.Warnf("ServiceType is set as a NodePortLocal, but the CNI is not set as antrea")
		return false
	}
	return true
}

func GetDisableStaticRoute() bool {
	if GetAdvancedL4() {
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
	if GetAdvancedL4() {
		return GetClusterID()
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
		clusterName := strings.Split(clusterID, ":")
		if len(clusterName) > 1 {
			return clusterName[0]
		}
	}
	return ""
}

func IsClusterNameValid() bool {
	clusterName := GetClusterName()
	re := regexp.MustCompile("^[a-zA-Z0-9-_]*$")
	if clusterName == "" {
		utils.AviLog.Error("Required param clusterName not specified, syncing will be disabled")
		return false
	} else if !re.MatchString(clusterName) {
		utils.AviLog.Error("clusterName must consist of alphanumeric characters or '-'/'_' (max 32 chars), syncing will be disabled")
		return false
	}
	return true
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
	if GetEnableCtrl2014Features() {
		labels := GetLabels()
		clusterKey = *labels[0].Key
		clusterValue = *labels[0].Value
		clusterLabelChecksum = utils.Hash(clusterKey + clusterValue)
	}
}

func GetClusterLabelChecksum() uint32 {
	return clusterLabelChecksum
}

func ObjectLabelChecksum(objectLabels []*models.RoleFilterMatchLabel) uint32 {
	var objChecksum uint32

	for _, label := range objectLabels {
		if *label.Key == clusterKey && label.Values != nil && len(label.Values) > 0 && label.Values[0] == clusterValue {
			objChecksum = clusterLabelChecksum
			break
		}
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
	if GetGRBACSupport() {
		if populateCache {
			if markers != nil {
				checksum += ObjectLabelChecksum(markers)
			}
			return checksum
		}
		checksum += GetClusterLabelChecksum()
	}
	return checksum
}

func InformersToRegister(oclient *oshiftclient.Clientset, kclient *kubernetes.Clientset) ([]string, error) {
	allInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.SecretInformer,
		utils.ConfigMapInformer,
		utils.PodInformer,
	}

	if GetServiceType() == NodePortLocal {
		allInformers = append(allInformers, utils.PodInformer)
	}

	if !GetAdvancedL4() {
		allInformers = append(allInformers, utils.NSInformer)
		allInformers = append(allInformers, utils.NodeInformer)

		informerTimeout := int64(120)
		_, err := kclient.CoreV1().Services(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{TimeoutSeconds: &informerTimeout})
		if err != nil {
			return allInformers, errors.New("Error in fetching services: " + err.Error())
		}
		_, err = oclient.RouteV1().Routes(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{TimeoutSeconds: &informerTimeout})
		if err == nil {
			// Openshift cluster with route support, we will just add route informer
			allInformers = append(allInformers, utils.RouteInformer)
		} else {
			// Kubernetes cluster
			allInformers = append(allInformers, utils.IngressInformer)
			allInformers = append(allInformers, utils.IngressClassInformer)
		}
	}
	return allInformers, nil
}

func SSLKeyCertChecksum(sslName, certificate, cacert string, markers []*models.RoleFilterMatchLabel, populateCache bool) uint32 {
	checksum := utils.Hash(sslName + certificate + cacert)
	if GetGRBACSupport() {
		if populateCache {
			if markers != nil {
				checksum += ObjectLabelChecksum(markers)
			}
			return checksum
		}
		checksum += GetClusterLabelChecksum()
	}
	return checksum
}

func L4PolicyChecksum(ports []int64, protocol string, markers []*models.RoleFilterMatchLabel, populateCache bool) uint32 {
	var portsInt []int
	for _, port := range ports {
		portsInt = append(portsInt, int(port))
	}
	sort.Ints(portsInt)
	checksum := utils.Hash(utils.Stringify(portsInt)) + utils.Hash(protocol)
	if GetGRBACSupport() {
		if populateCache {
			if markers != nil {
				checksum += ObjectLabelChecksum(markers)
			}
			return checksum
		}
		checksum += GetClusterLabelChecksum()
	}
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

var CloudType string

func SetCloudType(cloudType string) {
	CloudType = cloudType
}

func GetCloudType() string {
	if CloudType == "" {
		return CLOUD_VCENTER
	}
	return CloudType
}

var IsCloudInAdminTenant = true

func SetIsCloudInAdminTenant(isCloudInAdminTenant bool) {
	IsCloudInAdminTenant = isCloudInAdminTenant
}

func IsPublicCloud() bool {
	cloudType := GetCloudType()
	if cloudType == CLOUD_AZURE || cloudType == CLOUD_AWS ||
		cloudType == CLOUD_GCP {
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

func GetPassthroughShardVSName(s string, key string) string {
	var vsNum uint32
	shardSize := PassthroughShardSize()
	shardVsPrefix := GetClusterName() + "--" + PassthroughPrefix
	vsNum = utils.Bkt(s, shardSize)
	vsName := shardVsPrefix + strconv.Itoa(int(vsNum))
	utils.AviLog.Infof("key: %s, msg: ShardVSName: %s", key, vsName)
	return Encode(vsName)
}

// GetLabels returns the key value pair used for tagging the segroups and routes in vrfcontext
func GetLabels() []*models.KeyValue {
	clusterName := GetClusterName()
	labelKey := SeGroupLabelKey
	kv := &models.KeyValue{
		Key:   &labelKey,
		Value: &clusterName,
	}
	labels := []*models.KeyValue{}
	labels = append(labels, kv)
	return labels
}

// GetMarkers returns the key values pair used for tagging the segroups and routes in vrfcontext
func GetMarkers() []*models.RoleFilterMatchLabel {
	clusterName := GetClusterName()
	labelKey := SeGroupLabelKey
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

func VSVipDelRequired() bool {
	c, err := semver.NewConstraint(">= " + VSVIPDELCTRLVER)
	if err == nil {
		currVersion, verErr := semver.NewVersion(utils.CtrlVersion)
		if verErr == nil && c.Check(currVersion) {
			return true
		}
	}
	return false
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

func ValidateIngressForClass(key string, ingress *networkingv1beta1.Ingress) bool {
	// see whether ingress class resources are present or not
	if !utils.GetIngressClassEnabled() {
		return filterIngressOnClassAnnotation(key, ingress)
	}

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

	ingClassObj, err := utils.GetInformers().IngressClassInformer.Lister().Get(*ingress.Spec.IngressClassName)
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

func filterIngressOnClassAnnotation(key string, ingress *networkingv1beta1.Ingress) bool {
	// If Avi is not the default ingress, then filter on ingress class.
	if !GetDefaultIngController() {
		annotations := ingress.GetAnnotations()
		ingClass, ok := annotations[INGRESS_CLASS_ANNOT]
		if ok && ingClass == AVI_INGRESS_CLASS {
			return true
		} else {
			utils.AviLog.Infof("key: %s, msg: AKO is not running as the default ingress controller. Not processing the ingress: %s. Please annotate the ingress class as 'avi'", key, ingress.Name)
			return false
		}
	} else {
		// If Avi is the default ingress controller, sync everything than the ones that are annotated with ingress class other than 'avi'
		annotations := ingress.GetAnnotations()
		ingClass, ok := annotations[INGRESS_CLASS_ANNOT]
		if ok && ingClass != AVI_INGRESS_CLASS {
			utils.AviLog.Infof("key: %s, msg: AKO is the default ingress controller but not processing the ingress: %s since ingress class is set to : %s", key, ingress.Name, ingClass)
			return false
		} else {
			return true
		}
	}
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

func IsAviLBDefaultIngressClassWithClient(kc kubernetes.Interface) (string, bool) {
	ingClassObjs, _ := kc.NetworkingV1beta1().IngressClasses().List(context.TODO(), metav1.ListOptions{})
	for _, ingClass := range ingClassObjs.Items {
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

func GetAviSecretWithRetry(kc kubernetes.Interface, retryCount int) (*v1.Secret, error) {
	var aviSecret *v1.Secret
	var err error
	for retry := 0; retry < retryCount; retry++ {
		aviSecret, err = kc.CoreV1().Secrets(utils.GetAKONamespace()).Get(context.TODO(), AviSecret, metav1.GetOptions{})
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
	ctrlIpAddress := os.Getenv(utils.ENV_CTRL_IPADDRESS)

	aviClient := utils.NewAviRestClientWithToken(ctrlIpAddress, ctrlUsername, ctrlAuthToken)
	if aviClient == nil {
		utils.AviLog.Errorf("Failed to initialize AVI client")
		return
	}
	userTokensListResp, err := utils.GetAuthTokenWithRetry(aviClient, retryCount)
	if err != nil {
		utils.AviLog.Errorf("Failed to get existing tokens from controller, err: %+v", err)
		return
	}
	oldTokenID, refresh, err := utils.GetTokenFromRestObj(userTokensListResp, ctrlAuthToken)
	if err != nil {
		utils.AviLog.Errorf("Failed to find token on controller, err: %+v", err)
		return
	}
	if !refresh {
		utils.AviLog.Infof("Skipping AuthToken Refresh")
		return
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
	aviSecret, err := GetAviSecretWithRetry(kc, retryCount)
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
	return ctrlProps, nil
}
