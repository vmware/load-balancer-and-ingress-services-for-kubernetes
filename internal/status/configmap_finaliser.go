package status

import (
	"encoding/json"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// RemoveConfigmapFinalizer : Remove the ako finaliser from configmap. After this the configmap can be deleted by the user
// This can be used to notify the user that all AVI objects have been deleted by AKO.
func RemoveConfigmapFinalizer() {
	currConfig, err := utils.GetInformers().ConfigMapInformer.Lister().ConfigMaps(lib.AviNS).Get(lib.AviConfigMap)
	if err != nil {
		utils.AviLog.Warnf("Error in getting configmap: %v", err)
		return
	}
	currConfig.SetFinalizers([]string{})
	UpdateConfigmapFinalizer(currConfig, []string{})
	utils.AviLog.Infof("Removed the finalizer %s from avi CM", lib.ConfigmapFinalizer)
}

// SetConfigmapFinalizer : update from configmap with ako finaliser.
// After this the configmap cannot be deleted by the user without clearing the finaliser
func AddConfigmapFinalizer() {
	currConfig, err := utils.GetInformers().ConfigMapInformer.Lister().ConfigMaps(lib.AviNS).Get(lib.AviConfigMap)
	if err != nil {
		utils.AviLog.Warnf("Error in getting configmap: %v", err)
		return
	}

	if lib.ContainsFinalizer(currConfig, lib.ConfigmapFinalizer) {
		utils.AviLog.Warnf("Avi configmap already has the finaliser: %s", lib.ConfigmapFinalizer)
		return
	}

	UpdateConfigmapFinalizer(currConfig, []string{lib.ConfigmapFinalizer})
	utils.AviLog.Infof("Successfully patched the CM with finalizers: %v", currConfig.GetFinalizers())
}

func UpdateConfigmapFinalizer(currConfig *v1.ConfigMap, finalizerStr []string) {
	currConfig.SetFinalizers(finalizerStr)
	patchPayload, _ := json.Marshal(map[string]interface{}{
		"metadata": map[string][]string{
			"finalizers": finalizerStr,
		},
	})

	_, err := utils.GetInformers().ClientSet.CoreV1().ConfigMaps(lib.AviNS).Patch(lib.AviConfigMap, types.MergePatchType, patchPayload)
	if err != nil {
		utils.AviLog.Warnf("Error in updating configmap: %v", err)
	}
}
