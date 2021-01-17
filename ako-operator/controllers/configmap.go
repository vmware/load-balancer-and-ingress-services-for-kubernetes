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

	enableRHI := "false"
	if ako.Spec.NetworkSettings.EnableRHI {
		enableRHI = "true"
	}
	cm.Data[EnableRHI] = enableRHI
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
	nwListBytes := []byte{}
	var err error
	for _, row := range ako.Spec.NetworkSettings.NodeNetworkList {
		nwListRows = append(nwListRows, NodeNetworkListRow{
			Cidrs:       row.Cidrs,
			NetworkName: row.NetworkName,
		})
	}
	if len(nwListRows) != 0 {
		nwListBytes, err = json.Marshal(nwListRows)
		if err != nil {
			return cm, err
		}
	}
	cm.Data[NodeNetworkList] = string(nwListBytes)
	cm.Data[SyncNamespace] = ako.Spec.L7Settings.SyncNamespace

	cm.Data[NSSyncLabelKey] = ako.Spec.AKOSettings.NSSelector.LabelKey
	cm.Data[NSSyncLabelValue] = ako.Spec.AKOSettings.NSSelector.LabelValue

	tenantsPerCluster := "false"
	if ako.Spec.ControllerSettings.TenantsPerCluster {
		tenantsPerCluster = "true"
	}
	cm.Data[TenantsPerCluster] = tenantsPerCluster
	cm.Data[TenantName] = ako.Spec.ControllerSettings.TenantName

	return cm, nil
}
