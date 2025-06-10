/*
Copyright 2020 VMware, Inc.
All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
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
		if oldCM.GetName() != "" {
			log.V(0).Info("a configmap with the same name already exists, it will be updated", "name",
				oldCM.GetName())
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

	if oldCM.GetName() != "" {
		SetIfRebootRequired(oldCM, cm)
		// "avi-k8s-config" configmap already exists, we just need to update that
		// updating shouldn't change the existing finalizers
		existingFinalizers := oldCM.GetFinalizers()
		for _, f := range existingFinalizers {
			if !utils.HasElem(cm.GetFinalizers(), f) {
				cm.Finalizers = append(cm.Finalizers, f)
			}
		}
		err := r.Update(ctx, &cm)
		if err != nil {
			log.Error(err, "unable to update configmap", "namespace", cm.GetNamespace(), "name",
				cm.GetName())
			return err
		}
	} else {
		err := r.Create(ctx, &cm)
		if err != nil {
			log.Error(err, "unable to create configmap", "namespace", cm.GetNamespace(), "name",
				cm.GetName())
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
		Name:      cm.GetName(),
		Namespace: cm.GetNamespace(),
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

	enableEVH := "false"
	if ako.Spec.AKOSettings.EnableEVH {
		enableEVH = "true"
	}
	cm.Data[EnableEVH] = enableEVH

	layer7Only := "false"
	if ako.Spec.AKOSettings.Layer7Only {
		layer7Only = "true"
	}
	cm.Data[Layer7Only] = layer7Only

	servicesAPI := "false"
	if ako.Spec.AKOSettings.ServicesAPI {
		servicesAPI = "true"
	}
	cm.Data[ServicesAPI] = servicesAPI

	vipPerNamespace := "false"
	if ako.Spec.AKOSettings.VipPerNamespace {
		vipPerNamespace = "true"
	}
	cm.Data[VipPerNamespace] = vipPerNamespace

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

	cm.Data[LogLevel] = string(ako.Spec.LogLevel)

	deleteConfig := "false"
	if ako.Spec.AKOSettings.DeleteConfig {
		deleteConfig = "true"
	}
	cm.Data[DeleteConfig] = deleteConfig

	enableRHI := "false"
	if ako.Spec.NetworkSettings.EnableRHI {
		enableRHI = "true"
	}
	cm.Data[EnableRHI] = enableRHI
	cm.Data[NsxtT1LR] = ako.Spec.NetworkSettings.NsxtT1LR

	var err error
	type VipNetworkListRow struct {
		Cidr        string `json:"cidr,omitempty"`
		NetworkName string `json:"networkName,omitempty"`
		V6Cidr      string `json:"v6cidr,omitempty"`
		NetworkUUID string `json:"networkUUID,omitempty"`
	}

	vipListRows := []VipNetworkListRow{}
	vipListBytes := []byte{}
	for _, row := range ako.Spec.NetworkSettings.VipNetworkList {
		vipListRows = append(vipListRows, VipNetworkListRow{
			Cidr:        row.Cidr,
			NetworkName: row.NetworkName,
			V6Cidr:      row.V6Cidr,
			NetworkUUID: row.NetworkUUID,
		})
	}
	if len(vipListRows) != 0 {
		vipListBytes, err = json.Marshal(vipListRows)
		if err != nil {
			return cm, err
		}
	}
	cm.Data[VipNetworkList] = string(vipListBytes)

	bgpPeerLabelsBytes, err := json.Marshal(ako.Spec.NetworkSettings.BGPPeerLabels)
	if err != nil {
		return cm, err
	}
	cm.Data[BgpPeerLabels] = string(bgpPeerLabelsBytes)

	serviceType := string(ako.Spec.L7Settings.ServiceType)
	cm.Data[ServiceType] = serviceType
	if serviceType == "NodePort" {
		cm.Data[NodeKey] = ako.Spec.NodePortSelector.Key
		cm.Data[NodeValue] = ako.Spec.NodePortSelector.Value
	}

	cm.Data[ServiceEngineGroupName] = ako.Spec.ControllerSettings.ServiceEngineGroupName
	apiServerPort := ako.Spec.AKOSettings.APIServerPort
	if apiServerPort > 0 {
		cm.Data[APIServerPort] = strconv.Itoa(apiServerPort)
	} else {
		cm.Data[APIServerPort] = "8080"
	}

	type NodeNetworkListRow struct {
		Cidrs       []string `json:"cidrs,omitempty"`
		NetworkName string   `json:"networkName,omitempty"`
		NetworkUUID string   `json:"networkUUID,omitempty"`
	}

	nwListRows := []NodeNetworkListRow{}
	nwListBytes := []byte{}
	for _, row := range ako.Spec.NetworkSettings.NodeNetworkList {
		nwListRows = append(nwListRows, NodeNetworkListRow{
			Cidrs:       row.Cidrs,
			NetworkName: row.NetworkName,
			NetworkUUID: row.NetworkUUID,
		})
	}
	if len(nwListRows) != 0 {
		nwListBytes, err = json.Marshal(nwListRows)
		if err != nil {
			return cm, err
		}
	}
	cm.Data[NodeNetworkList] = string(nwListBytes)

	noPGForSni := "false"
	if ako.Spec.L7Settings.NoPGForSNI {
		noPGForSni = "true"
	}
	cm.Data[NoPGForSni] = noPGForSni

	cm.Data[NSSyncLabelKey] = ako.Spec.AKOSettings.NSSelector.LabelKey
	cm.Data[NSSyncLabelValue] = ako.Spec.AKOSettings.NSSelector.LabelValue

	cm.Data[TenantName] = ako.Spec.ControllerSettings.TenantName
	cm.Data[AutoFQDN] = ako.Spec.L4Settings.AutoFQDN

	enableEvents := "true"
	if !ako.Spec.AKOSettings.EnableEvents {
		enableEvents = "false"
	}
	cm.Data[EnableEvents] = enableEvents

	primaryInstance := "true"
	cm.Data[PrimaryInstance] = primaryInstance

	istioEnabled := "false"
	if ako.Spec.AKOSettings.IstioEnabled {
		istioEnabled = "true"
	}
	cm.Data[IstioEnabled] = istioEnabled

	blockedNamespaceListBytes, err := json.Marshal(ako.Spec.AKOSettings.BlockedNamespaceList)
	if err != nil {
		return cm, err
	}
	cm.Data[BlockedNamespaceList] = string(blockedNamespaceListBytes)
	cm.Data[IPFamily] = ako.Spec.AKOSettings.IPFamily

	enableMCI := "false"
	cm.Data[EnableMCI] = enableMCI

	useDefaultSecretsOnly := "false"
	if ako.Spec.AKOSettings.UseDefaultSecretsOnly {
		useDefaultSecretsOnly = "true"
	}
	cm.Data[UseDefaultSecretsOnly] = useDefaultSecretsOnly

	cm.Data[VPCMode] = strconv.FormatBool(ako.Spec.AKOSettings.VPCMode)

	return cm, nil
}

func checkDeprecatedFields(ako akov1alpha1.AKOConfig, log logr.Logger) {
	if ako.Spec.L4Settings.AdvancedL4 {
		log.V(0).Info("", "WARN: ", "akoconfig.Spec.L4Settings.AdvancedL4 will be deprecated")
	}

	if ako.Spec.L7Settings.SyncNamespace != "" {
		log.V(0).Info("", "WARN: ", "akoconfig.Spec.L7Settings.SyncNamespace will be deprecated")
	}

	if ako.Spec.ControllerSettings.TenantsPerCluster {
		log.V(0).Info("", "WARN: ", "akoconfig.Spec.ControllerSettings.TenantsPerCluster will be deprecated")
	}
}
