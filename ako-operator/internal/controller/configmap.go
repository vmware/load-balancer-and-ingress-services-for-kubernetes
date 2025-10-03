/*
Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.

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

package controller

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1beta1"
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

func createOrUpdateConfigMap(ctx context.Context, ako akov1beta1.AKOConfig, log logr.Logger, r *AKOConfigReconciler) error {
	log.V(1).Info("building a new configmap for AKO")

	var oldCM corev1.ConfigMap

	if err := r.Get(ctx, getConfigMapName(), &oldCM); err != nil {
		if errors.IsNotFound(err) {
			log.V(0).Info("no existing configmap with name", "name", ConfigMapName)
		} else {
			log.Error(err, "unable to get existing configmap", "name", ConfigMapName)
			return err
		}
	} else {
		if oldCM.GetName() != "" {
			log.V(0).Info("a configmap with the same name already exists, it will be updated", "name",
				oldCM.GetName())
		}
	}

	desiredCM, err := BuildConfigMap(ako)
	if err != nil {
		log.Error(err, "error in building configmap")
		return err
	}

	if oldCM.GetName() != "" { // ConfigMap exists, update it
		if reflect.DeepEqual(oldCM.Data, desiredCM.Data) {
			log.V(0).Info("no updates required for configmap")
			// add this object in the global list
			objList := getObjectList()
			objList[types.NamespacedName{
				Name:      oldCM.GetName(),
				Namespace: oldCM.GetNamespace(),
			}] = &oldCM
			return nil
		}

		SetIfRebootRequired(oldCM, desiredCM)

		// Update the existing ConfigMap object with the new data
		oldCM.Data = desiredCM.Data

		// Set controller reference on the fetched object
		err = ctrl.SetControllerReference(&ako, &oldCM, r.Scheme)
		if err != nil {
			log.Error(err, "error in setting controller reference to configmap", "name", oldCM.GetName())
			return err
		}

		err := r.Update(ctx, &oldCM)
		if err != nil {
			log.Error(err, "unable to update configmap", "name", oldCM.GetName())
			return err
		}

	} else {
		// ConfigMap does not exist, create it
		// Set controller reference on the new object before creating
		err = ctrl.SetControllerReference(&ako, &desiredCM, r.Scheme)
		if err != nil {
			log.Error(err, "error in setting controller reference to configmap", "name", desiredCM.GetName())
			return err
		}

		err := r.Create(ctx, &desiredCM)
		if err != nil {
			log.Error(err, "unable to create configmap", "name", desiredCM.GetName())
			return err
		}
	}

	var newCM corev1.ConfigMap
	err = r.Get(ctx, getConfigMapName(), &newCM) // Fetch the latest state after create/update
	if err != nil {
		log.Error(err, "error getting a configmap with name", "name", ConfigMapName)
		return err
	}
	// update this object in the global list
	objList := getObjectList()
	objList[types.NamespacedName{
		Name:      newCM.GetName(),
		Namespace: newCM.GetNamespace(),
	}] = &newCM

	return nil
}

func BuildConfigMap(ako akov1beta1.AKOConfig) (corev1.ConfigMap, error) {
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
	if ako.Spec.NetworkSettings.DefaultDomain != "" {
		cm.Data[DefaultDomain] = ako.Spec.NetworkSettings.DefaultDomain
	} else {
		cm.Data[DefaultDomain] = ako.Spec.L4Settings.DefaultDomain
	}
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

	cm.Data[LogLevel] = string(ako.Spec.AKOSettings.LogLevel)

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

	autoFQDN := "default"
	if ako.Spec.L4Settings.AutoFQDN != "" {
		autoFQDN = ako.Spec.L4Settings.AutoFQDN
	}
	cm.Data[AutoFQDN] = autoFQDN

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
	cm.Data[VRFName] = ako.Spec.ControllerSettings.VRFName

	defaultLBController := "true"
	if !ako.Spec.L4Settings.DefaultLBController {
		defaultLBController = "false"
	}
	cm.Data[DefaultLBController] = defaultLBController

	enablePrometheus := "false"
	if ako.Spec.FeatureGates.EnablePrometheus {
		enablePrometheus = "true"
	}
	cm.Data[EnablePrometheus] = enablePrometheus

	fqdnReusePolicy := "InterNamespaceAllowed"
	if string(ako.Spec.L7Settings.FQDNReusePolicy) != "" {
		fqdnReusePolicy = string(ako.Spec.L7Settings.FQDNReusePolicy)
	}
	cm.Data[FQDNReusePolicy] = fqdnReusePolicy

	if ako.Spec.FeatureGates.EnableEndpointSlice {
		cm.Data[EnableEndpointSlice] = "true"
	} else {
		cm.Data[EnableEndpointSlice] = "false"
	}

	return cm, nil
}
