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
	"encoding/json"
	"testing"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
)

var cmJson = `
{
	"apiVersion": "v1",
	"kind": "ConfigMap",
	"metadata": {
		"name": "avi-k8s-config",
		"namespace": "avi-system",
		"creationTimestamp": null
	},
	"data": {
		"apiServerPort": "8080",
		"autoFQDN": "default",
		"cloudName": "test-cloud",
		"clusterName": "test-cluster",
		"cniPlugin": "test-cni",
		"enableEVH": "false",
		"layer7Only": "false",
		"servicesAPI": "false",
		"vipPerNamespace": "false",
		"controllerIP": "10.10.10.11",
		"controllerVersion": "1.1",
		"defaultDomain": "test.com",
		"defaultIngController": "true",
		"deleteConfig": "false",
		"disableStaticRouteSync": "false",
		"fullSyncFrequency": "1800",
		"enableEvents": "true",
		"logLevel": "INFO",
		"nsxtT1LR": "",
		"nodeNetworkList": "[{\"cidrs\":[\"10.10.10.0/24\"],\"networkName\":\"test-nw\"}]",
		"passhtroughShardSize": "SMALL",
		"serviceEngineGroupName": "test-group",
		"serviceType": "ClusterIP",
		"shardVSSize": "LARGE",
		"noPGForSni": "false",
		"bgpPeerLabels": "[]",
		"vipNetworkList": "[{\"cidr\":\"10.10.10.0/24\",\"networkName\":\"test-nw\"}]",
		"nsSyncLabelKey": "",
		"nsSyncLabelValue": "",
		"enableMCI": "false",
		"enableRHI": "false",
		"tenantName": "admin",
		"primaryInstance": "true",
		"ipFamily": "V4",
		"istioEnabled": "false",
		"blockedNamespaceList": "[]",
		"useDefaultSecretsOnly": "false"
	}
}
`

func getTestDefaultConfigMap() (corev1.ConfigMap, error) {
	defCm := corev1.ConfigMap{}
	err := json.Unmarshal([]byte(cmJson), &defCm)
	return defCm, err
}

func buildConfigMapAndVerify(existingCm corev1.ConfigMap, akoConfig akov1alpha1.AKOConfig,
	rebootRequiredValue, shouldCksumMatch bool, t *testing.T) corev1.ConfigMap {

	g := gomega.NewGomegaWithT(t)
	// will send an empty string, as this is anyway verified during reboot required check
	existingCmCksum := getChecksum(existingCm, []string{})
	newCm, err := BuildConfigMap(akoConfig)

	newCksum := getChecksum(newCm, []string{})

	match := existingCmCksum == newCksum
	g.Expect(match).To(gomega.Equal(shouldCksumMatch))

	g.Expect(err).To(gomega.BeNil())
	SetIfRebootRequired(newCm, existingCm)
	g.Expect(rebootRequired).To(gomega.Equal(rebootRequiredValue))
	if rebootRequired {
		// reset the reboot required value
		rebootRequired = false
	}
	return newCm
}
