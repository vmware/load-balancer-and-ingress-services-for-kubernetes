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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var dynamicInformerInstance *DynamicInformers
var dynamicClientSet dynamic.Interface

var (
	// CalicoBlockaffinityGVR : Calico's BlockAffinity CRD resource identifier
	CalicoBlockaffinityGVR = schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "blockaffinities",
	}

	// HostSubnetGVR : OpenShift's HostSubnet CRD resource identifier
	HostSubnetGVR = schema.GroupVersionResource{
		Group:    "network.openshift.io",
		Version:  "v1",
		Resource: "hostsubnets",
	}

	// CiliumNodeGVR : Cilium's CiliumNode CRD resource identifier
	CiliumNodeGVR = schema.GroupVersionResource{
		Group:    "cilium.io",
		Version:  "v2",
		Resource: "ciliumnodes",
	}

	NetworkInfoGVR = schema.GroupVersionResource{
		Group:    "nsx.vmware.com",
		Version:  "v1alpha1",
		Resource: "namespacenetworkinfos",
	}

	ClusterNetworkGVR = schema.GroupVersionResource{
		Group:    "nsx.vmware.com",
		Version:  "v1alpha1",
		Resource: "clusternetworkinfos",
	}

	AvailabilityZoneVR = schema.GroupVersionResource{
		Group:    "topology.tanzu.vmware.com",
		Version:  "v1alpha1",
		Resource: "availabilityzones",
	}

	VPCNetworkConfigurationGVR = schema.GroupVersionResource{
		Group:    "crd.nsx.vmware.com",
		Version:  "v1alpha1",
		Resource: "vpcnetworkconfigurations",
	}
)

type BootstrapCRData struct {
	SecretName, SecretNamespace, UserName, TZPath, AviURL string
}

// NewDynamicClientSet initializes dynamic client set instance
func NewDynamicClientSet(config *rest.Config) (dynamic.Interface, error) {
	// do not instantiate the dynamic client set if the CNI being used is NOT calico
	if !utils.IsVCFCluster() && GetCNIPlugin() != CALICO_CNI && GetCNIPlugin() != OPENSHIFT_CNI && GetCNIPlugin() != CILIUM_CNI {
		return nil, nil
	}

	ds, err := dynamic.NewForConfig(config)
	if err != nil {
		utils.AviLog.Infof("Error while creating dynamic client %v", err)
		return nil, err
	}
	if dynamicClientSet == nil {
		dynamicClientSet = ds
	}
	return dynamicClientSet, nil
}

// SetDynamicClientSet is used for Unit tests.
func SetDynamicClientSet(c dynamic.Interface) {
	dynamicClientSet = c
}

// GetDynamicClientSet returns dynamic client set instance
func GetDynamicClientSet() dynamic.Interface {
	if dynamicClientSet == nil {
		utils.AviLog.Warn("Cannot retrieve the dynamic clientset since it's not initialized yet.")
		return nil
	}
	return dynamicClientSet
}

// DynamicInformers holds third party generic informers
type DynamicInformers struct {
	CalicoBlockAffinityInformer informers.GenericInformer
	HostSubnetInformer          informers.GenericInformer
	CiliumNodeInformer          informers.GenericInformer

	VCFNetworkInfoInformer    informers.GenericInformer
	VCFClusterNetworkInformer informers.GenericInformer

	AvailabilityZoneInformer informers.GenericInformer

	VPCNetworkConfigurationInformer informers.GenericInformer
}

// NewDynamicInformers initializes the DynamicInformers struct
func NewDynamicInformers(client dynamic.Interface, akoInfra bool) *DynamicInformers {
	informers := &DynamicInformers{}
	f := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, v1.NamespaceAll, nil)

	switch GetCNIPlugin() {
	case CALICO_CNI:
		informers.CalicoBlockAffinityInformer = f.ForResource(CalicoBlockaffinityGVR)
	case OPENSHIFT_CNI:
		informers.HostSubnetInformer = f.ForResource(HostSubnetGVR)
	case CILIUM_CNI:
		informers.CiliumNodeInformer = f.ForResource(CiliumNodeGVR)
	default:
		utils.AviLog.Infof("Skipped initializing dynamic informers for cniPlugin %s", GetCNIPlugin())
	}

	if utils.IsVCFCluster() && akoInfra {
		informers.VCFNetworkInfoInformer = f.ForResource(NetworkInfoGVR)
		informers.VCFClusterNetworkInformer = f.ForResource(ClusterNetworkGVR)
		informers.AvailabilityZoneInformer = f.ForResource(AvailabilityZoneVR)
		informers.VPCNetworkConfigurationInformer = f.ForResource(VPCNetworkConfigurationGVR)
	}

	dynamicInformerInstance = informers
	return dynamicInformerInstance
}

// GetDynamicInformers returns DynamicInformers instance
func GetDynamicInformers() *DynamicInformers {
	if dynamicInformerInstance == nil {
		utils.AviLog.Warn("Cannot retrieve the dynamic informers since it's not initialized yet.")
		return nil
	}
	return dynamicInformerInstance
}

func GetNetworkInfoCRData() (map[string]string, map[string]string, map[string]map[string]struct{}) {
	clientSet := GetDynamicClientSet()
	lslrMap := make(map[string]string)
	nsLRMap := make(map[string]string)
	cidrs := make(map[string]map[string]struct{})

	crList, err := clientSet.Resource(NetworkInfoGVR).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		utils.AviLog.Errorf("Error getting Networkinfo CR %v", err)
		return lslrMap, nsLRMap, cidrs
	}

	if len(crList.Items) == 0 {
		utils.AviLog.Infof("No Networkinfo CRs found.")
		return lslrMap, nsLRMap, cidrs
	}

	if cidrIntf, clusterNetworkCIDRFound := GetClusterNetworkInfoCRData(clientSet); clusterNetworkCIDRFound {
		// Set the namespace to cluster name for the cluster ingress cidr
		ns := GetClusterName()
		cidrs[ns] = make(map[string]struct{})
		cidrMap := cidrs[ns]
		for _, cidr := range cidrIntf {
			cidrMap[cidr.(string)] = struct{}{}
		}
		utils.AviLog.Infof("Ingress CIDR found from Cluster Network Info %v", cidrIntf)
	}

	for _, obj := range crList.Items {
		ns := obj.GetNamespace()
		spec := obj.Object["topology"].(map[string]interface{})
		lr, ok := spec["gatewayPath"].(string)
		if !ok {
			utils.AviLog.Infof("lr not found in networkinfo object")
			continue
		}
		ls, ok := spec["aviSegmentPath"].(string)
		if !ok {
			utils.AviLog.Infof("ls not found in networkinfo object")
			continue
		}
		lslrMap[ls] = lr
		nsLRMap[ns] = lr
		cidrIntf, ok := spec["ingressCIDRs"].([]interface{})
		if !ok {
			continue
		}
		for _, cidr := range cidrIntf {
			if _, ok := cidrs[ns]; !ok {
				cidrs[ns] = make(map[string]struct{})
			}
			cidrMap := cidrs[ns]
			cidrMap[cidr.(string)] = struct{}{}
		}
	}

	return lslrMap, nsLRMap, cidrs
}

func GetAvailabilityZonesCRData(clientSet dynamic.Interface) ([]string, error) {
	clusterIDs := make([]string, 0)
	crList, err := clientSet.Resource(AvailabilityZoneVR).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		utils.AviLog.Errorf("Error getting Availability Zone CR %v", err)
		return clusterIDs, err
	}
	if len(crList.Items) == 0 {
		return clusterIDs, fmt.Errorf("no Availability Zone CRs found")
	}

	for _, obj := range crList.Items {
		spec, ok := obj.Object["spec"].(map[string]interface{})
		if !ok {
			utils.AviLog.Errorf("spec not found in the CR %+v", obj)
			continue
		}
		clusterID, ok := spec["clusterComputeResourceMoId"].(string)
		if !ok {
			utils.AviLog.Errorf("Cluster MoID not found in the CR %+v", obj)
			continue
		}
		clusterIDs = append(clusterIDs, clusterID)
	}
	return clusterIDs, nil
}

func GetClusterNetworkInfoCRData(clientSet dynamic.Interface) ([]interface{}, bool) {
	var cidrs []interface{}
	crList, err := clientSet.Resource(ClusterNetworkGVR).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		utils.AviLog.Errorf("Error getting Cluster Network Info CR %v", err)
		return cidrs, false
	}

	if len(crList.Items) == 0 {
		utils.AviLog.Errorf("No Cluster Network Info CRs found.")
		return cidrs, false
	}

	crObj := crList.Items[0]
	spec := crObj.Object["topology"].(map[string]interface{})
	cidrIntf, ok := spec["ingressCIDRs"].([]interface{})
	if !ok {
		utils.AviLog.Infof("cidr not found in Cluster Network Info object")
		return cidrs, false
	}
	return cidrIntf, true
}

// GetPodCIDR returns the node's configured PodCIDR
func GetPodCIDR(node *v1.Node) ([]string, error) {
	nodename := node.ObjectMeta.Name
	var podCIDR string
	var podCIDRs []string
	dynamicClient := GetDynamicClientSet()

	if GetCNIPlugin() == CALICO_CNI && dynamicClientSet != nil {
		crdClient := dynamicClient.Resource(CalicoBlockaffinityGVR)
		crdList, err := crdClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			utils.AviLog.Errorf("Error getting CRD %v", err)
			return nil, err
		}

		for _, i := range crdList.Items {
			crdSpec := (i.Object["spec"]).(map[string]interface{})
			crdNodeName := crdSpec["node"].(string)
			if crdNodeName == nodename {
				podCIDR = crdSpec["cidr"].(string)
				if podCIDR == "" {
					utils.AviLog.Errorf("Error in fetching Pod CIDR from BlockAffinity %v", node.ObjectMeta.Name)
					return nil, errors.New("podcidr not found")
				}

				if !utils.HasElem(podCIDRs, podCIDR) {
					podCIDRs = append(podCIDRs, podCIDR)
				}
			}
		}

	} else if GetCNIPlugin() == OPENSHIFT_CNI && dynamicClientSet != nil {
		crdClient := dynamicClient.Resource(HostSubnetGVR)
		crdList, err := crdClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			utils.AviLog.Errorf("Error getting CRD %v", err)
			return nil, err
		}

		for _, i := range crdList.Items {
			host, ok := (i.Object["host"]).(string)
			if !ok {
				utils.AviLog.Errorf("Error in parsing hostsubnets crd list")
				return nil, errors.New("Error in parsing hostsubnets crd list")
			}

			if host == nodename {
				podCIDR, ok := (i.Object["subnet"]).(string)
				if !ok {
					utils.AviLog.Errorf("Error in parsing hostsubnets crd list")
					return nil, errors.New("Error in parsing hostsubnets crd list")
				}

				if !utils.HasElem(podCIDRs, podCIDR) {
					podCIDRs = append(podCIDRs, podCIDR)
				}
			}
		}

	} else if GetCNIPlugin() == OVN_KUBERNETES_CNI {
		var nodeSubnets string
		var found bool
		if nodeSubnets, found = node.Annotations[OVNNodeSubnetAnnotation]; !found {
			return nil, errors.New("k8s.ovn.org/node-subnets annotation not found in Node Metadata")
		}
		var nodeSubnetJson map[string]interface{}
		err := json.Unmarshal([]byte(nodeSubnets), &nodeSubnetJson)
		if err != nil {
			return nil, errors.New("Error while unmarshalling k8s.ovn.org/node-subnets annotation in Node Metadata : " + err.Error())
		}
		if podCIDR, ok := nodeSubnetJson["default"].(string); ok {
			if podCIDR == "" {
				utils.AviLog.Errorf("Error in fetching Pod CIDR from Node Metadata %v", node.ObjectMeta.Name)
				return nil, errors.New("podcidr not found")
			}
			podCIDRs = append(podCIDRs, podCIDR)
		} else if podCIDRList, ok := nodeSubnetJson["default"].([]interface{}); ok {
			if len(podCIDRList) == 0 {
				utils.AviLog.Errorf("Error in fetching Pod CIDR from Node Metadata %v", node.ObjectMeta.Name)
				return nil, errors.New("podcidr not found")
			}
			for _, cidr := range podCIDRList {
				if podCIDR, ok := cidr.(string); ok {
					if podCIDR == "" {
						utils.AviLog.Errorf("Error in fetching Pod CIDR from Node Metadata %v", node.ObjectMeta.Name)
						return nil, errors.New("podcidr not found")
					}
					podCIDRs = append(podCIDRs, podCIDR)
				}
			}
		}
	} else if GetCNIPlugin() == CILIUM_CNI && dynamicClientSet != nil {
		crdClient := dynamicClient.Resource(CiliumNodeGVR)
		crdList, err := crdClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			utils.AviLog.Errorf("Error getting CRD %v", err)
			return nil, err
		}

		for _, i := range crdList.Items {
			crdMetadata := (i.Object["metadata"]).(map[string]interface{})
			crdNodeName := crdMetadata["name"].(string)
			if crdNodeName == nodename {
				crdSpec := (i.Object["spec"]).(map[string]interface{})
				crdIpam, ok := crdSpec["ipam"].(map[string]interface{})
				if !ok {
					utils.AviLog.Errorf("Error in fetching ipam from CiliumNode")
					return nil, errors.New("Error in parsing ciliumnode crd list")
				}
				crdPodCidrs, ok := crdIpam["podCIDRs"].([]interface{})
				if !ok {
					utils.AviLog.Errorf("Error in fetching Pod CIDR from CiliumNode")
					return nil, errors.New("Error in parsing ciliumnode crd list")
				}
				for _, podCIDR := range crdPodCidrs {
					podCIDRString := podCIDR.(string)
					if !utils.HasElem(podCIDRs, podCIDRString) {
						podCIDRs = append(podCIDRs, podCIDRString)
					}
				}
			}
		}

	} else {
		if podCidrsFromAnnotation, ok := node.Annotations[StaticRouteAnnotation]; ok {
			podCidrSlice := strings.Split(strings.TrimSpace(podCidrsFromAnnotation), ",")
			for _, podCidr := range podCidrSlice {
				if podCidr == "" {
					continue
				}
				cidr := strings.TrimSpace(podCidr)
				re := regexp.MustCompile(IPCIDRRegex)
				if !re.MatchString(cidr) {
					return nil, fmt.Errorf("CIDR value %s in annotation %v is of incorrect format", cidr, podCidrsFromAnnotation)
				}
				podCIDRs = append(podCIDRs, cidr)
			}
		} else {
			if node.Spec.PodCIDR == "" {
				utils.AviLog.Errorf("Error in fetching Pod CIDR from NodeSpec %v", node.ObjectMeta.Name)
				return nil, errors.New("podcidr not found")
			}
			podCIDRs = append(podCIDRs, node.Spec.PodCIDRs...)
		}
	}
	return podCIDRs, nil
}

// GetCNIPlugin returns the user provided CNI plugin - oneof (calico|canal|flannel)
func GetCNIPlugin() string {
	return strings.ToLower(os.Getenv(CNI_PLUGIN))
}

// WaitForInitSecretRecreateAndReboot Deletes the avi-init-secret provided by NCP,
// in order for NCP to generate the token and recreate the Secret. After Secret deletion,
// once AKO finds a new Secret created, it reboots in order to refresh the Client and
// Session to the Avi Controller.
// This can be further improved to update Avi Controller Session during runtime, but
// is complicated business right now.
func WaitForInitSecretRecreateAndReboot() {
	cs := utils.GetInformers().ClientSet
	if err := cs.CoreV1().Secrets(utils.VMWARE_SYSTEM_AKO).Delete(context.TODO(), AviInitSecret, metav1.DeleteOptions{}); err != nil {
		utils.AviLog.Errorf("Error while deleting the init Secret for Secret refresh.")
		return
	}

	var checkForInitSecretRecreate = func(cs kubernetes.Interface) error {
		_, err := cs.CoreV1().Secrets(utils.VMWARE_SYSTEM_AKO).Get(context.TODO(), AviInitSecret, metav1.GetOptions{})
		return err
	}

	defer utils.AviLog.Fatalf("Found new init secret, rebooting AKO")
	// This waits for AKO to get a new refreshed Secret for a total of 75 seconds.
	for retry := 0; retry < 15; retry++ {
		err := checkForInitSecretRecreate(cs)
		if err == nil {
			return
		}
		if k8serrors.IsNotFound(err) {
			utils.AviLog.Infof("init Secret not found, retrying...")
		} else {
			utils.AviLog.Fatalf(err.Error())
		}
		time.Sleep(5 * time.Second)
	}
}

func GetVPCs() (map[string]string, map[string]string, error) {
	clientSet := GetDynamicClientSet()
	vpcToSubnetMap := make(map[string]string)
	nsToVPCMap := make(map[string]string)
	vpcNetworkConfigCRs, err := clientSet.Resource(VPCNetworkConfigurationGVR).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return vpcToSubnetMap, nsToVPCMap, err
	}
	namespaces, err := utils.GetInformers().NSInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		return vpcToSubnetMap, nsToVPCMap, err
	}
	vpcNetworkConfigToNamespaceMap := make(map[string][]string)
	for _, ns := range namespaces {
		vpcNetCR := ns.Annotations["nsx.vmware.com/vpc_network_config"]
		if vpcNetCR == "" {
			continue
		}
		vpcNetworkConfigToNamespaceMap[vpcNetCR] = append(vpcNetworkConfigToNamespaceMap[vpcNetCR], ns.GetName())
	}

	for _, obj := range vpcNetworkConfigCRs.Items {
		status, ok := obj.Object["status"].(map[string]interface{})
		if !ok {
			continue
		}
		vpcs, ok := status["vpcs"].([]interface{})
		if !ok {
			continue
		}
		for _, vpc := range vpcs {
			vpcInfo, ok := vpc.(map[string]interface{})
			if !ok {
				continue
			}
			aviSubnetPath, ok := vpcInfo["lbSubnetPath"].(string)
			if !ok {
				continue
			}
			if aviSubnetPath == "" {
				utils.AviLog.Warnf("invalid value for aviSubnetPath: %s in the VPCNetworkConfig CR %s", aviSubnetPath, obj.GetName())
				continue
			}
			vpcPath := strings.Split(aviSubnetPath, "/subnets/")[0]
			vpcToSubnetMap[vpcPath] = aviSubnetPath
			for _, ns := range vpcNetworkConfigToNamespaceMap[obj.GetName()] {
				nsToVPCMap[ns] = vpcPath
			}
		}
	}
	return vpcToSubnetMap, nsToVPCMap, nil
}
