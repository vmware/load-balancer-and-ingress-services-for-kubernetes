package controllers

import (
	"encoding/json"
	"testing"

	"github.com/onsi/gomega"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
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
		"advancedL4": "false",
		"apiServerPort": "8080",
		"cloudName": "test-cloud",
		"clusterName": "test-cluster",
		"cniPlugin": "test-cni",
		"controllerIP": "10.10.10.11",
		"controllerVersion": "1.1",
		"defaultDomain": "test.com",
		"defaultIngController": "true",
		"deleteConfig": "false",
		"disableStaticRouteSync": "false",
		"fullSyncFrequency": "1800",
		"l7ShardingScheme": "hostname",
		"logLevel": "INFO",
		"networkName": "test-nw",
		"nodeKey": "key",
		"nodeNetworkList": "[{\"cidrs\":[\"10.10.10.0/24\"],\"networkName\":\"test-nw\"}]",
		"nodeValue": "value",
		"passhtroughShardSize": "SMALL",
		"serviceEngineGroupName": "test-group",
		"serviceType": "ClusterIP",
		"shardVSSize": "LARGE",
		"subnetIP": "10.10.10.1",
		"subnetPrefix": "24"
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
