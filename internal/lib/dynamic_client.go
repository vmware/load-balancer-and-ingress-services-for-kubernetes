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
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
)

var dynamicInformerInstance *DynamicInformers
var dynamicClientSet dynamic.Interface

var vcfDynamicInformerInstance *VCFDynamicInformers
var vcfDynamicClientSet dynamic.Interface
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

	BootstrapGVR = schema.GroupVersionResource{
		Group:    "ncp.vmware.com",
		Version:  "v1alpha1",
		Resource: "akobootstrapconditions",
	}

	NetworkInfoGVR = schema.GroupVersionResource{
		Group:    "nsx.vmware.com",
		Version:  "v1alpha1",
		Resource: "namespacenetworkinfos",
	}
)

type BootstrapCRData struct {
	SecretName, SecretNamespace, UserName, TZPath, AviURL string
}

// NewDynamicClientSet initializes dynamic client set instance
func NewDynamicClientSet(config *rest.Config) (dynamic.Interface, error) {
	// do not instantiate the dynamic client set if the CNI being used is NOT calico
	if GetCNIPlugin() != CALICO_CNI && GetCNIPlugin() != OPENSHIFT_CNI && !utils.IsVCFCluster() {
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

// GetDynamicClientSet returns dynamic client set instance
func GetDynamicClientSet() dynamic.Interface {
	if dynamicClientSet == nil {
		utils.AviLog.Warn("Cannot retrieve the dynamic informers since it's not initialized yet.")
		return nil
	}
	return dynamicClientSet
}

// DynamicInformers holds third party generic informers
type DynamicInformers struct {
	CalicoBlockAffinityInformer informers.GenericInformer
	HostSubnetInformer          informers.GenericInformer
	NCPBootstrapInformer        informers.GenericInformer
}

// NewDynamicInformers initializes the DynamicInformers struct
func NewDynamicInformers(client dynamic.Interface) *DynamicInformers {
	informers := &DynamicInformers{}
	f := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, v1.NamespaceAll, nil)

	switch GetCNIPlugin() {
	case CALICO_CNI:
		informers.CalicoBlockAffinityInformer = f.ForResource(CalicoBlockaffinityGVR)
	case OPENSHIFT_CNI:
		informers.HostSubnetInformer = f.ForResource(HostSubnetGVR)
	default:
		utils.AviLog.Infof("Skipped initializing dynamic informers %s ", GetCNIPlugin())
	}

	if utils.IsVCFCluster() {
		informers.NCPBootstrapInformer = f.ForResource(BootstrapGVR)
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

func NewVCFDynamicClientSet(config *rest.Config) (dynamic.Interface, error) {
	ds, err := dynamic.NewForConfig(config)
	if err != nil {
		utils.AviLog.Infof("Error while creating dynamic client %v", err)
		return nil, err
	}
	if vcfDynamicClientSet == nil {
		vcfDynamicClientSet = ds
	}
	return vcfDynamicClientSet, nil
}

func SetVCFVCFDynamicClientSet(dc dynamic.Interface) {
	vcfDynamicClientSet = dc
}

func GetVCFDynamicClientSet() dynamic.Interface {
	if vcfDynamicClientSet == nil {
		utils.AviLog.Warn("Cannot retrieve the dynamic informers since it's not initialized yet.")
		return nil
	}
	return vcfDynamicClientSet
}

type VCFDynamicInformers struct {
	NCPBootstrapInformer informers.GenericInformer
	NetworkInfoInformer  informers.GenericInformer
}

func NewVCFDynamicInformers(client dynamic.Interface) *VCFDynamicInformers {
	informers := &VCFDynamicInformers{}
	f := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, v1.NamespaceAll, nil)

	informers.NCPBootstrapInformer = f.ForResource(BootstrapGVR)
	informers.NetworkInfoInformer = f.ForResource(NetworkInfoGVR)

	vcfDynamicInformerInstance = informers
	return vcfDynamicInformerInstance
}

func GetVCFDynamicInformers() *VCFDynamicInformers {
	if vcfDynamicInformerInstance == nil {
		utils.AviLog.Warn("Cannot retrieve the dynamic informers since it's not initialized yet.")
		return nil
	}
	return vcfDynamicInformerInstance
}

func GetBootstrapCRData() (BootstrapCRData, bool) {
	var boostrapdata BootstrapCRData
	dynamicClient := GetVCFDynamicClientSet()
	crdClient := dynamicClient.Resource(BootstrapGVR)
	crdList, err := crdClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		utils.AviLog.Errorf("Error getting CRD %v", err)
		return boostrapdata, false
	}

	if len(crdList.Items) != 1 {
		utils.AviLog.Errorf("Expected only one object for NCP bootstrap but found: %d", len(crdList.Items))
		return boostrapdata, false
	}

	obj := crdList.Items[0]
	spec, ok := obj.Object["spec"].(map[string]interface{})
	if !ok {
		utils.AviLog.Errorf("spec is not found in NCP bootstrap object")
		return boostrapdata, false
	}
	secretref, ok := spec["albCredentialSecretRef"].(map[string]interface{})
	if !ok {
		utils.AviLog.Errorf("albCredentialSecretRef is not found in NCP bootstrap object")
		return boostrapdata, false
	}
	albtoken, ok := spec["albTokenProperty"].(map[string]interface{})
	if !ok {
		utils.AviLog.Errorf("albTokenProperty is not found in NCP bootstrap object")
		return boostrapdata, false
	}

	secretName, ok := secretref["name"].(string)
	if !ok {
		utils.AviLog.Errorf("secretName is not of type string")
		return boostrapdata, false
	}
	secretNamespace, ok := secretref["namespace"].(string)
	if !ok {
		utils.AviLog.Errorf("secretNamespace is not of type string")
		return boostrapdata, false
	}
	userName, ok := albtoken["userName"].(string)
	if !ok {
		utils.AviLog.Errorf("userName is not of type string")
		return boostrapdata, false
	}

	status, ok := obj.Object["status"].(map[string]interface{})
	if !ok {
		utils.AviLog.Errorf("Status not found in bootstrap CR")
		return boostrapdata, false
	}
	tzPath, ok := status["transportZone"].(string)
	if !ok {
		utils.AviLog.Errorf("transportZone path not found in status of bootstrap CR")
		return boostrapdata, false
	}

	albEndpoint, ok := status["albEndpoint"].(map[string]interface{})
	if !ok {
		utils.AviLog.Errorf("albEndpoint not found in status of bootstrap CR")
		return boostrapdata, false
	}
	hostUrl, ok := albEndpoint["hostUrl"].(string)
	if !ok {
		utils.AviLog.Errorf("hostUrl path not found in status of bootstrap CR")
		return boostrapdata, false
	}
	boostrapdata.SecretName = secretName
	boostrapdata.SecretNamespace = secretNamespace
	boostrapdata.UserName = userName
	boostrapdata.TZPath = tzPath
	boostrapdata.AviURL = hostUrl

	return boostrapdata, true
}

func GetControllerURLFromBootstrapCR() string {
	dynamicClient := GetVCFDynamicClientSet()
	crdClient := dynamicClient.Resource(BootstrapGVR)
	crdList, err := crdClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		utils.AviLog.Errorf("Error getting CRD %v", err)
		return ""
	}

	if len(crdList.Items) != 1 {
		utils.AviLog.Errorf("Expected only one object for NCP bootstrap but found: %d", len(crdList.Items))
		return ""
	}

	obj := crdList.Items[0]

	status, ok := obj.Object["status"].(map[string]interface{})
	if !ok {
		utils.AviLog.Errorf("Status not found in bootstrap CR")
		return ""
	}

	albEndpoint, ok := status["albEndpoint"].(map[string]interface{})
	if !ok {
		utils.AviLog.Errorf("albEndpoint not found in status of bootstrap CR")
		return ""
	}
	hostUrl, ok := albEndpoint["hostUrl"].(string)
	if !ok {
		utils.AviLog.Errorf("hostUrl path not found in status of bootstrap CR")
		return ""
	}

	return hostUrl
}

func GetNetinfoCRData() (map[string]string, map[string]struct{}) {
	lslrMap := make(map[string]string)
	cidrs := make(map[string]struct{})
	dynamicClient := GetVCFDynamicClientSet()
	crdClient := dynamicClient.Resource(NetworkInfoGVR)
	crList, err := crdClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		utils.AviLog.Errorf("Error getting Networkinfo CR %v", err)
		return lslrMap, cidrs
	}

	if len(crList.Items) == 0 {
		utils.AviLog.Infof("No Networkinfo CRs found.")
		return lslrMap, cidrs
	}

	for _, obj := range crList.Items {
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
		cidrIntf, ok := spec["ingressCIDRs"].([]interface{})
		if !ok {
			utils.AviLog.Infof("cidr not found in networkinfo object")
			continue
		} else {
			for _, cidr := range cidrIntf {
				cidrs[cidr.(string)] = struct{}{}
			}
		}
	}

	return lslrMap, cidrs
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
