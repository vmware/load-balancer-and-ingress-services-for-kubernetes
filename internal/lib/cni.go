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
	"os"
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
)

// NewDynamicClientSet initializes dynamic client set instance
func NewDynamicClientSet(config *rest.Config) (dynamic.Interface, error) {
	// do not instantiate the dynamic client set if the CNI being used is NOT calico
	if GetCNIPlugin() != CALICO_CNI && GetCNIPlugin() != OPENSHIFT_CNI {
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
		utils.AviLog.Infof("Skipped initializing dynamic informers %s \n", GetCNIPlugin())
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
		podCIDR = node.Spec.PodCIDR
		if podCIDR == "" {
			utils.AviLog.Errorf("Error in fetching Pod CIDR from NodeSpec %v", node.ObjectMeta.Name)
			return nil, errors.New("podcidr not found")
		}

		podCIDRs = append(podCIDRs, node.Spec.PodCIDR)
	}

	return podCIDRs, nil
}

// GetCNIPlugin returns the user provided CNI plugin - oneof (calico|canal|flannel)
func GetCNIPlugin() string {
	return strings.ToLower(os.Getenv(CNI_PLUGIN))
}
