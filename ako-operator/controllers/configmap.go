package controllers

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/go-logr/logr"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func SetIfRebootRequired(oldCm corev1.ConfigMap, newCm corev1.ConfigMap) {
	skipList := []string{DeleteConfig, LogLevel}
	oldCksum := getChecksum(oldCm, skipList)
	newCksum := getChecksum(newCm, skipList)

	if oldCksum != newCksum {
		// reboot is required
		rebootRequired = true
	}
}

func createOrUpdateConfigMap(ctx context.Context, ako akov1alpha1.AKOConfig, log logr.Logger, r *AKOConfigReconciler) error {
	log.V(1).Info("building a new configmap for AKO")

	var oldCM corev1.ConfigMap

	if err := r.Get(ctx, getConfigMapName(), &oldCM); err != nil {
		log.V(0).Info("error getting a configmap with name", "name", ConfigMapName, "err", err)
	} else {
		log.V(1).Info("old configmap", "old cm", oldCM)
		if oldCM.ObjectMeta.GetName() != "" {
			log.V(0).Info("a configmap with the same name already exists, it will be updated", "name",
				oldCM.ObjectMeta.GetName())
		}
	}

	cm, err := BuildConfigMap(ako)
	if err != nil {
		log.Error(err, "error in building configmap")
	}
	err = ctrl.SetControllerReference(&ako, &cm, r.Scheme)
	if err != nil {
		log.Error(err, "error in setting controller reference, configmap changes would be ignored")
	}

	if oldCM.ObjectMeta.GetName() != "" {
		SetIfRebootRequired(oldCM, cm)
		// "avi-k8s-config" configmap already exists, we just need to update that
		// updating shouldn't change the existing finalizers
		existingFinalizers := oldCM.ObjectMeta.Finalizers
		for _, f := range oldCM.GetObjectMeta().GetFinalizers() {
			if !utils.HasElem(cm.ObjectMeta.Finalizers, f) {
				cm.ObjectMeta.Finalizers = append(cm.ObjectMeta.Finalizers, f)
			}
		}
		cm.ObjectMeta.Finalizers = existingFinalizers
		err := r.Update(ctx, &cm)
		if err != nil {
			log.Error(err, "unable to update configmap", "namespace", cm.ObjectMeta.GetNamespace(),
				"name", cm.ObjectMeta.GetName())
			return err
		}
	} else {
		err := r.Create(ctx, &cm)
		if err != nil {
			log.Error(err, "unable to create configmap", "namespace", cm.ObjectMeta.GetNamespace(),
				"name", cm.ObjectMeta.GetName())
			return err
		}
	}

	var newCM corev1.ConfigMap
	err = r.Get(ctx, getConfigMapName(), &newCM)
	if err != nil {
		log.V(0).Info("error getting a configmap with name", "name", ConfigMapName, "err", err)
		return err
	}
	// update this object in the global list
	objList := getObjectList()
	objList[types.NamespacedName{
		Name:      cm.ObjectMeta.GetName(),
		Namespace: cm.ObjectMeta.GetNamespace(),
	}] = &newCM

	return nil
}

func BuildConfigMap(ako akov1alpha1.AKOConfig) (corev1.ConfigMap, error) {
	cm := corev1.ConfigMap{ObjectMeta: v1.ObjectMeta{
		Name:      ConfigMapName,
		Namespace: AviSystemNS,
	}}

	cm.Data = make(map[string]string)
	cm.Data[ControllerIP] = ako.Spec.ControllerSettings.ControllerIP
	cm.Data[ControllerVersion] = ako.Spec.ControllerSettings.ControllerVersion
	cm.Data[CniPlugin] = ako.Spec.AKOSettings.CNIPlugin
	cm.Data[ShardVSSize] = string(ako.Spec.L7Settings.ShardVSSize)
	cm.Data[PassthroughShardSize] = string(ako.Spec.L7Settings.PassthroughShardSize)
	fullSyncFreq := ako.Spec.AKOSettings.FullSyncFrequency
	cm.Data[FullSyncFrequency] = fullSyncFreq
	cm.Data[CloudName] = ako.Spec.ControllerSettings.CloudName
	cm.Data[ClusterName] = ako.Spec.AKOSettings.ClusterName
	cm.Data[DefaultDomain] = ako.Spec.L4Settings.DefaultDomain
	disableStaticRouteSync := "false"
	if ako.Spec.AKOSettings.DisableStaticRouteSync {
		disableStaticRouteSync = "true"
	}
	cm.Data[DisableStaticRouteSync] = disableStaticRouteSync

	defaultIngController := "false"
	if ako.Spec.L7Settings.DefaultIngController {
		defaultIngController = "true"
	}
	cm.Data[DefaultIngController] = defaultIngController

	cm.Data[SubnetIP] = ako.Spec.NetworkSettings.SubnetIP
	cm.Data[SubnetPrefix] = ako.Spec.NetworkSettings.SubnetPrefix
	cm.Data[NetworkName] = ako.Spec.NetworkSettings.NetworkName
	cm.Data[L7ShardingScheme] = ako.Spec.L7Settings.ShardingScheme
	cm.Data[LogLevel] = string(ako.Spec.LogLevel)

	deleteConfig := "false"
	if ako.Spec.AKOSettings.DeleteConfig {
		deleteConfig = "true"
	}
	cm.Data[DeleteConfig] = deleteConfig

	advancedL4 := "false"
	if ako.Spec.L4Settings.AdvancedL4 {
		advancedL4 = "true"
	}
	cm.Data[AdvancedL4] = advancedL4
	cm.Data[ServiceType] = string(ako.Spec.L7Settings.ServiceType)
	cm.Data[NodeKey] = ako.Spec.NodePortSelector.Key
	cm.Data[NodeValue] = ako.Spec.NodePortSelector.Value
	cm.Data[ServiceEngineGroupName] = ako.Spec.ControllerSettings.ServiceEngineGroupName
	apiServerPort := ako.Spec.AKOSettings.APIServerPort
	if apiServerPort > 0 {
		cm.Data[APIServerPort] = strconv.Itoa(apiServerPort)
	} else {
		cm.Data[APIServerPort] = "8080"
	}

	type NodeNetworkListRow struct {
		Cidrs       []string `json:"cidrs"`
		NetworkName string   `json:"networkName"`
	}
	nwListRows := []NodeNetworkListRow{}

	for _, row := range ako.Spec.NetworkSettings.NodeNetworkList {
		nwListRows = append(nwListRows, NodeNetworkListRow{
			Cidrs:       row.Cidrs,
			NetworkName: row.NetworkName,
		})
	}
	if len(nwListRows) != 0 {
		nwListBytes, err := json.Marshal(nwListRows)
		if err != nil {
			return cm, err
		}
		cm.Data[NodeNetworkList] = string(nwListBytes)
	}
	return cm, nil
}
