package lib

import (
	"errors"

	"github.com/avinetworks/container-lib/utils"

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
	CalicoBlockaffinityGVR = schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "blockaffinities",
	}
)

func NewDynamicClientSet(config *rest.Config) (dynamic.Interface, error) {
	ds, err := dynamic.NewForConfig(config)
	if err != nil {
		utils.AviLog.Warning.Printf("Error while creating dynamic client %v", err)
		return nil, err
	}
	if dynamicClientSet == nil {
		dynamicClientSet = ds
	}
	return dynamicClientSet, nil
}

func GetDynamicClientSet() dynamic.Interface {
	if dynamicClientSet == nil {
		utils.AviLog.Warning.Print("Cannot retrieve the dynamic informers since it's not initialized yet.")
		return nil
	}
	return dynamicClientSet
}

// DynamicInformers holds third party generic informers
type DynamicInformers struct {
	CalicoBlockAffinityInformer informers.GenericInformer
}

func NewDynamicInformers(client dynamic.Interface) *DynamicInformers {
	informers := &DynamicInformers{}
	f := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, v1.NamespaceAll, nil)
	informers.CalicoBlockAffinityInformer = f.ForResource(CalicoBlockaffinityGVR)
	return informers
}

func GetDynamicInformers() *DynamicInformers {
	if dynamicInformerInstance == nil {
		utils.AviLog.Warning.Print("Cannot retrieve the dynamic informers since it's not initialized yet.")
		return nil
	}
	return dynamicInformerInstance
}

// GetPodCIDR returns the node's configured PodCIDR
func GetPodCIDR(node *v1.Node) (string, error) {
	nodename := node.ObjectMeta.Name
	podCIDR := node.Spec.PodCIDR // default
	dynamicClient := GetDynamicClientSet()

	if dynamicClientSet != nil {
		crdClient := dynamicClient.Resource(CalicoBlockaffinityGVR)
		crdList, err := crdClient.List(metav1.ListOptions{})
		if err != nil {
			utils.AviLog.Warning.Printf("Error getting CRD %v", err)
		}

		for _, i := range crdList.Items {
			crdSpec := (i.Object["spec"]).(map[string]interface{})
			crdNodeName := crdSpec["node"].(string)
			if crdNodeName == nodename {
				podCIDR = crdSpec["cidr"].(string)
				break
			}
		}

	}

	if podCIDR == "" {
		utils.AviLog.Error.Printf("Error in fetching Pod CIDR for %v", node.ObjectMeta.Name)
		return "", errors.New("podcidr not found")
	}
	return podCIDR, nil
}
