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

package scaletest

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	crdfake "ako/internal/client/clientset/versioned/fake"
	"ako/internal/k8s"
	"ako/internal/lib"
	avinodes "ako/internal/nodes"
	"ako/internal/objects"
	"ako/tests/integrationtest"

	utils "github.com/avinetworks/container-lib/utils"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestMain(m *testing.M) {
	os.Setenv("INGRESS_API", "extensionv1")
	os.Setenv("NETWORK_NAME", "net123")
	os.Setenv("CLUSTER_NAME", "cluster")
	os.Setenv("CLOUD_NAME", "Default-Cloud")
	KubeClient = k8sfake.NewSimpleClientset()
	CRDClient = crdfake.NewSimpleClientset()
	lib.SetCRDClientset(CRDClient)

	registeredInformers := []string{
		utils.ServiceInformer,
		utils.EndpointInformer,
		utils.IngressInformer,
		utils.SecretInformer,
		utils.NSInformer,
		utils.NodeInformer,
		utils.ConfigMapInformer,
	}
	utils.NewInformers(utils.KubeClientIntf{KubeClient}, registeredInformers)
	informers := k8s.K8sinformers{Cs: KubeClient}
	k8s.NewCRDInformers(CRDClient)

	integrationtest.InitializeFakeAKOAPIServer()

	integrationtest.NewAviFakeClientInstance()
	defer integrationtest.AviFakeClientInstance.Close()

	ctrl = k8s.SharedAviController()
	stopCh := utils.SetupSignalHandler()
	k8s.PopulateCache()
	ctrlCh := make(chan struct{})
	quickSyncCh := make(chan struct{})
	ctrl.HandleConfigMap(informers, ctrlCh, stopCh, quickSyncCh)
	go ctrl.InitController(informers, registeredInformers, ctrlCh, stopCh, quickSyncCh)
	AddConfigMap()
	integrationtest.KubeClient = KubeClient

	os.Exit(m.Run())
}

var KubeClient *k8sfake.Clientset
var CRDClient *crdfake.Clientset
var ctrl *k8s.AviController

func AddConfigMap() {
	aviCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "avi-system",
			Name:      "avi-k8s-config",
		},
	}
	KubeClient.CoreV1().ConfigMaps("avi-system").Create(aviCM)

	integrationtest.PollForSyncStart(ctrl, 10)
}

func SetUpHostnameShardTestforIngress(t *testing.T, ns string) {
	os.Setenv("L7_SHARD_SCHEME", "hostname")
	os.Setenv("USE_PVC", "true")
	integrationtest.CreateSVC(t, ns, "avisvc", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, ns, "avisvc", false, false, "1.1.1")
}

func TearDownHostnameShardTestForIngress(t *testing.T, ns string, modeltoPools map[string]map[string]poolData) {
	integrationtest.DelSVC(t, ns, "avisvc")
	integrationtest.DelEP(t, ns, "avisvc")
	for modelName := range modeltoPools {
		ok, _ := objects.SharedAviGraphLister().Get(modelName)
		if ok {
			objects.SharedAviGraphLister().Delete(modelName)
		}
	}
}

func SetUpTestForIngress(t *testing.T, ns string) string {
	modelName := "admin/" + avinodes.DeriveNamespacedShardVS(ns, "test")
	integrationtest.CreateSVC(t, ns, "avisvc", corev1.ServiceTypeClusterIP, false)
	integrationtest.CreateEP(t, ns, "avisvc", false, false, "1.1.1")
	return modelName
}

func TearDownTestForIngress(t *testing.T, ns, modelName string) {
	integrationtest.DelSVC(t, ns, "avisvc")
	integrationtest.DelEP(t, ns, "avisvc")
	ok, _ := objects.SharedAviGraphLister().Get(modelName)
	if ok {
		objects.SharedAviGraphLister().Delete(modelName)
	}
}

// Pool information for an ingress
// Can be used to match pools from model graph
type poolData struct {
	priority    string
	ingressName string
	hostName    string
	path        string
}

type ingrData struct {
	ingressName string
	hosts       []string
	paths       []string
}

func createIngresses(t *testing.T, ns string, ingrdata map[string]ingrData) {
	for _, ingr := range ingrdata {
		ingrFake := (integrationtest.FakeIngress{
			Name:        ingr.ingressName,
			Namespace:   ns,
			DnsNames:    ingr.hosts,
			Paths:       ingr.paths,
			ServiceName: "avisvc",
		}).Ingress()

		_, err := KubeClient.ExtensionsV1beta1().Ingresses(ns).Create(ingrFake)
		if err != nil {
			t.Fatalf("error in adding Ingress: %v", err)
		}
		time.Sleep(10 * time.Microsecond)
	}
}

func createIngressesMultiPath(t *testing.T, ns string, ingrdata map[string]ingrData) {
	for _, ingr := range ingrdata {
		ingrFake := (integrationtest.FakeIngress{
			Name:        ingr.ingressName,
			Namespace:   ns,
			DnsNames:    ingr.hosts,
			Paths:       ingr.paths,
			ServiceName: "avisvc",
		}).IngressMultiPath()

		_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Create(ingrFake)
		if err != nil {
			t.Fatalf("error in adding Ingress: %v", err)
		}
		time.Sleep(10 * time.Microsecond)
	}
}

func updateIngresses(t *testing.T, ns string, ingrdata map[string]ingrData) {
	for _, ingr := range ingrdata {
		ingrFake := (integrationtest.FakeIngress{
			Name:        ingr.ingressName,
			Namespace:   ns,
			DnsNames:    ingr.hosts,
			Paths:       ingr.paths,
			ServiceName: "avisvc",
		}).Ingress()
		ingrFake.ResourceVersion = "2"

		_, err := KubeClient.ExtensionsV1beta1().Ingresses(ns).Update(ingrFake)
		if err != nil {
			t.Fatalf("error in adding Ingress: %v", err)
		}
		time.Sleep(10 * time.Microsecond)
	}
}

func updateIngressesMultiPath(t *testing.T, ns string, ingrdata map[string]ingrData) {
	for _, ingr := range ingrdata {
		ingrFake := (integrationtest.FakeIngress{
			Name:        ingr.ingressName,
			Namespace:   ns,
			DnsNames:    ingr.hosts,
			Paths:       ingr.paths,
			ServiceName: "avisvc",
		}).IngressMultiPath()
		ingrFake.ResourceVersion = "2"

		_, err := KubeClient.ExtensionsV1beta1().Ingresses("default").Update(ingrFake)
		if err != nil {
			t.Fatalf("error in adding Ingress: %v", err)
		}
		time.Sleep(10 * time.Microsecond)
	}
}

func delIngresses(t *testing.T, ns string, ingrdata map[string]ingrData) {
	for _, ingr := range ingrdata {
		err := KubeClient.ExtensionsV1beta1().Ingresses(ns).Delete(ingr.ingressName, nil)
		time.Sleep(10 * time.Microsecond)
		if err != nil {
			t.Fatalf("error in deleting Ingress: %v", err)
		}
	}
}

// constructs list of ingresses and corresponsing pool data
func generatePoolData(t *testing.T, ns string, count int, hostSuffix, ingrSuffix, path string) (map[string]poolData, map[string]ingrData) {
	if hostSuffix == "" {
		hostSuffix = "foo.com"
	}
	if ingrSuffix == "" {
		ingrSuffix = "ingr"
	}
	if path == "" {
		path = "/bar"
	}

	pooldata := make(map[string]poolData)
	ingrdata := make(map[string]ingrData)
	for i := 0; i < count; i++ {
		host := fmt.Sprintf("%d-%s", i, hostSuffix)
		ingrname := fmt.Sprintf("%d-%s", i, ingrSuffix)
		poolname := "cluster--" + host + strings.Replace(path, "/", "_", 1) + "-" + ns + "-" + ingrname
		priority := host + path
		pooldata[poolname] = poolData{
			ingressName: ingrname,
			priority:    priority,
			hostName:    host,
			path:        path,
		}
		ingrdata[ingrname] = ingrData{
			ingressName: ingrname,
			hosts:       []string{host},
			paths:       []string{path},
		}
	}
	return pooldata, ingrdata
}

// constructs list of ingresses with multiple paths and corresponsing pool data
func generatePoolDataMultiPath(t *testing.T, ns string, count, pathCount int, hostSuffix, ingrSuffix string, pathPrefix string) (map[string]poolData, map[string]ingrData) {
	if hostSuffix == "" {
		hostSuffix = "foo.com"
	}
	if ingrSuffix == "" {
		ingrSuffix = "ingr"
	}
	if pathPrefix == "" {
		pathPrefix = "/bar"
	}

	var paths []string
	for i := 0; i < pathCount; i++ {
		path := fmt.Sprintf("%s-%d", pathPrefix, i)
		paths = append(paths, path)
	}

	pooldata := make(map[string]poolData)
	ingrdata := make(map[string]ingrData)
	for i := 0; i < count; i++ {
		host := fmt.Sprintf("%d-%s", i, hostSuffix)
		ingrname := fmt.Sprintf("%d-%s", i, ingrSuffix)
		ingrdata[ingrname] = ingrData{ingressName: ingrname, hosts: []string{host}}
		for _, path := range paths {
			poolname := "cluster--" + host + strings.Replace(path, "/", "_", 1) + "-" + ns + "-" + ingrname
			priority := host + path
			pooldata[poolname] = poolData{
				ingressName: ingrname,
				priority:    priority,
				hostName:    host,
				path:        path,
			}
		}
		ingrdata[ingrname] = ingrData{
			ingressName: ingrname,
			hosts:       []string{host},
			paths:       paths,
		}
	}
	return pooldata, ingrdata
}

// constructs list of ingresses with multiple hosts and corresponsing pool data
func generatePoolDataMultiHost(t *testing.T, ns string, count, hostCount int, hostSuffix, ingrSuffix, path string) (map[string]poolData, map[string]ingrData) {
	if hostSuffix == "" {
		hostSuffix = "foo.com"
	}
	if ingrSuffix == "" {
		ingrSuffix = "ingr"
	}
	if path == "" {
		path = "/bar"
	}

	pooldata := make(map[string]poolData)
	ingrdata := make(map[string]ingrData)
	for i := 0; i < count; i++ {
		ingrname := fmt.Sprintf("%d-%s", i, ingrSuffix)
		baseHost := fmt.Sprintf("%d-%s", i, hostSuffix)
		hosts := []string{}
		paths := []string{}
		for j := 0; j < hostCount; j++ {
			host := fmt.Sprintf("%s-%d", baseHost, j)
			hosts = append(hosts, host)
			paths = append(paths, path)
			poolname := "cluster--" + host + strings.Replace(path, "/", "_", 1) + "-" + ns + "-" + ingrname
			priority := host + path
			pooldata[poolname] = poolData{
				ingressName: ingrname,
				priority:    priority,
				hostName:    host,
				path:        path,
			}
		}
		ingrdata[ingrname] = ingrData{
			ingressName: ingrname,
			hosts:       hosts,
			paths:       paths,
		}
	}
	return pooldata, ingrdata
}

// given a model and list of pooldata verify if the model has correct pools
func verifyModel(t *testing.T, g *gomega.GomegaWithT, count int, modelName string, pooldata map[string]poolData, timeout int) {
	if timeout <= 0 {
		timeout = 10
	}

	integrationtest.PollForCompletion(t, modelName, timeout)
	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Could not find model: %s", modelName)
	}

	t.Logf("processing model: %s\n", modelName)

	g.Eventually(func() int {
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return len(nodes[0].PoolRefs)
	}, timeout).Should(gomega.Equal(count))

	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	for _, pool := range nodes[0].PoolRefs {
		data, ok := pooldata[pool.Name]
		if !ok {
			t.Fatalf("Unexpected Poolname: %s", pool.Name)
		}
		if data.priority != pool.PriorityLabel {
			t.Fatalf("Unexpected priority %s for Pool %s, expected %s", pool.PriorityLabel, pool.Name, data.priority)
		}
	}
	for _, pool := range nodes[0].PoolGroupRefs[0].Members {
		poolname := strings.TrimPrefix(*pool.PoolRef, "/api/pool?name=")
		data, ok := pooldata[poolname]
		if !ok {
			t.Fatalf("Unexpected Poolname: %s", poolname)
		}
		if data.priority != *pool.PriorityLabel {
			t.Fatalf("Unexpected priority %s for Pool %s, expected %s", *pool.PriorityLabel, poolname, data.priority)
		}
	}
}

// given a model and list of pooldata verify if the model has correct pools, to be retried if fails
func verifyModelWithRetry(t *testing.T, g *gomega.GomegaWithT, count int, modelName string, pooldata map[string]poolData, timeout, retry int) {
	if timeout <= 0 {
		timeout = 10
	}
	var success bool

	integrationtest.PollForCompletion(t, modelName, timeout)
	found, _ := objects.SharedAviGraphLister().Get(modelName)
	if !found {
		t.Fatalf("Could not find model: %s", modelName)
	}

	t.Logf("processing model: %s\n", modelName)

	for i := 0; i < retry; i++ {
		success = true
		g.Eventually(func() int {
			_, aviModel := objects.SharedAviGraphLister().Get(modelName)
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs)
		}, timeout).Should(gomega.Equal(count))

		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		for _, pool := range nodes[0].PoolRefs {
			data, ok := pooldata[pool.Name]
			if !ok {
				t.Logf("Unexpected Poolname: %s for model: %s", pool.Name, modelName)
				success = false
				time.Sleep(1 * time.Second)
				break
			}
			if data.priority != pool.PriorityLabel {
				t.Fatalf("Unexpected priority %s for Pool %s, expected %s", pool.PriorityLabel, pool.Name, data.priority)
			}
		}
		if !success {
			continue
		}
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			poolname := strings.TrimPrefix(*pool.PoolRef, "/api/pool?name=")
			data, ok := pooldata[poolname]
			if !ok {
				t.Fatalf("Unexpected Poolname: %s", poolname)
			}
			if data.priority != *pool.PriorityLabel {
				t.Fatalf("Unexpected priority %s for Pool %s, expected %s", *pool.PriorityLabel, poolname, data.priority)
			}
		}
	}

	if !success {
		t.Fatalf("model verification failed for: %s", modelName)
	}
}

// verify that a model has no pool
func verifyIngressDelete(t *testing.T, g *gomega.GomegaWithT, modelName string, timeout int) {
	if timeout <= 0 {
		timeout = 10
	}

	t.Logf("processing model: %s\n", modelName)

	found, aviModel := objects.SharedAviGraphLister().Get(modelName)
	if found {
		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs)
		}, timeout).Should(gomega.Equal(0))

		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolGroupRefs[0].Members)
		}, timeout).Should(gomega.Equal(0))
	} else {
		t.Fatalf("Could not find model: %s", modelName)
	}
}

// given multiple models and list of pooldata for each model, verify if the models have correct pools
func verifyMultiModel(t *testing.T, g *gomega.GomegaWithT, modeltoPools map[string]map[string]poolData, timeout int) {
	if timeout <= 0 {
		timeout = 10
	}

	var count int
	for modelName, pooldata := range modeltoPools {
		count = len(pooldata)
		t.Logf("processing model: %s\n", modelName)
		integrationtest.PollForCompletion(t, modelName, timeout)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			t.Fatalf("Could not find model: %s", modelName)
		}

		g.Eventually(func() int {
			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs)
		}, timeout).Should(gomega.Equal(count))

		nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		for _, pool := range nodes[0].PoolRefs {
			data, ok := pooldata[pool.Name]
			if !ok {
				t.Fatalf("Unexpected Poolname: %s", pool.Name)
			}
			if data.priority != pool.PriorityLabel {
				t.Fatalf("Unexpected priority %s for Pool %s, expected %s", pool.PriorityLabel, pool.Name, data.priority)
			}
		}
		for _, pool := range nodes[0].PoolGroupRefs[0].Members {
			poolname := strings.TrimPrefix(*pool.PoolRef, "/api/pool?name=")
			data, ok := pooldata[poolname]
			if !ok {
				t.Fatalf("Unexpected Poolname: %s", poolname)
			}
			if data.priority != *pool.PriorityLabel {
				t.Fatalf("Unexpected priority %s for Pool %s, expected %s", *pool.PriorityLabel, poolname, data.priority)
			}
		}
	}
}

// given multiple models and list of pooldata for each model, verify if the models have correct pools
// to be retried if fails
func verifyMultiModelWithRetry(t *testing.T, g *gomega.GomegaWithT, modeltoPools map[string]map[string]poolData, timeout, retry int) {
	if timeout <= 0 {
		timeout = 10
	}

	var success bool
	var count int
	for modelName, pooldata := range modeltoPools {
		count = len(pooldata)
		t.Logf("processing model: %s\n", modelName)
		for i := 0; i < retry; i++ {
			success = true
			integrationtest.PollForCompletion(t, modelName, timeout)
			found, aviModel := objects.SharedAviGraphLister().Get(modelName)
			if !found {
				t.Fatalf("Could not find model: %s", modelName)
			}

			g.Eventually(func() int {
				nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
				return len(nodes[0].PoolRefs)
			}, timeout).Should(gomega.Equal(count))

			nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			for _, pool := range nodes[0].PoolRefs {
				data, ok := pooldata[pool.Name]
				if !ok {
					//t.Fatalf("Unexpected Poolname: %s", pool.Name)
					t.Logf("Unexpected Poolname: %s for model: %s", pool.Name, modelName)
					success = false
					time.Sleep(1 * time.Second)
					break
				}
				if data.priority != pool.PriorityLabel {
					t.Fatalf("Unexpected priority %s for Pool %s, expected %s", pool.PriorityLabel, pool.Name, data.priority)
				}
			}
			if !success {
				continue
			}
			for _, pool := range nodes[0].PoolGroupRefs[0].Members {
				poolname := strings.TrimPrefix(*pool.PoolRef, "/api/pool?name=")
				data, ok := pooldata[poolname]
				if !ok {
					t.Fatalf("Unexpected Poolname: %s", poolname)
				}
				if data.priority != *pool.PriorityLabel {
					t.Fatalf("Unexpected priority %s for Pool %s, expected %s", *pool.PriorityLabel, poolname, data.priority)
				}
			}
		}
		if !success {
			t.Fatalf("model verification failed for: %s", modelName)
		}
	}
}

// verify that a list of model have no pool
func verifyMultiModelIngressDelete(t *testing.T, g *gomega.GomegaWithT, modeltoPools map[string]map[string]poolData, timeout int) {
	if timeout <= 0 {
		timeout = 10
	}

	for modelName := range modeltoPools {
		t.Logf("processing model: %s\n", modelName)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if found {
			g.Eventually(func() int {
				nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
				return len(nodes[0].PoolRefs)
			}, timeout).Should(gomega.Equal(0))

			g.Eventually(func() int {
				nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
				return len(nodes[0].PoolGroupRefs[0].Members)
			}, timeout).Should(gomega.Equal(0))
		} else {
			t.Fatalf("Could not find model: %s", modelName)
		}
	}
}

// test ingress create in a single namespace
func nIngressCreateTest(t *testing.T, count, timeout int) {
	g := gomega.NewGomegaWithT(t)
	ns := "default"
	modelName := SetUpTestForIngress(t, ns)

	pooldata, ingrdata := generatePoolData(t, ns, count, "", "", "")
	createIngresses(t, ns, ingrdata)
	verifyModel(t, g, count, modelName, pooldata, timeout)

	delIngresses(t, ns, ingrdata)
	verifyIngressDelete(t, g, modelName, timeout)
	TearDownTestForIngress(t, ns, modelName)
}

// test ingress update in a single namespace
func nIngressUpdatePathTest(t *testing.T, count, timeout, retry int) {
	g := gomega.NewGomegaWithT(t)
	ns := "default"
	modelName := SetUpTestForIngress(t, ns)

	_, ingrdata := generatePoolData(t, ns, count, "", "", "")
	createIngresses(t, ns, ingrdata)
	pooldataUpdated, ingrdataUpdated := generatePoolData(t, ns, count, "", "", "/foobar")
	updateIngresses(t, ns, ingrdataUpdated)
	verifyModelWithRetry(t, g, count, modelName, pooldataUpdated, timeout, retry)

	delIngresses(t, ns, ingrdata)
	verifyIngressDelete(t, g, modelName, timeout)
	TearDownTestForIngress(t, ns, modelName)
}

// test ingress with multipath create in a single namespace
func nIngressMultipathCreateTest(t *testing.T, ingrCount, pathCount, timeout int) {
	g := gomega.NewGomegaWithT(t)
	ns := "default"
	modelName := SetUpTestForIngress(t, ns)

	pooldata, ingrdata := generatePoolDataMultiPath(t, ns, ingrCount, pathCount, "", "", "")
	createIngressesMultiPath(t, ns, ingrdata)
	verifyModel(t, g, pathCount*ingrCount, modelName, pooldata, timeout)

	delIngresses(t, ns, ingrdata)
	verifyIngressDelete(t, g, modelName, timeout)
	TearDownTestForIngress(t, ns, modelName)
}

// test ingress with multipath update in a single namespace
func nIngressMultipathUpdatePathTest(t *testing.T, ingrCount, pathCount, timeout, retry int) {
	g := gomega.NewGomegaWithT(t)
	ns := "default"
	modelName := SetUpTestForIngress(t, ns)

	pooldata, ingrdata := generatePoolDataMultiPath(t, ns, ingrCount, pathCount, "", "", "")
	createIngressesMultiPath(t, ns, ingrdata)
	verifyModel(t, g, pathCount*ingrCount, modelName, pooldata, timeout)

	pooldataUpdated, ingrdataUpdated := generatePoolDataMultiPath(t, ns, ingrCount, pathCount, "", "", "/foobar")
	updateIngressesMultiPath(t, ns, ingrdataUpdated)
	verifyModelWithRetry(t, g, pathCount*ingrCount, modelName, pooldataUpdated, timeout, retry)

	delIngresses(t, ns, ingrdata)
	verifyIngressDelete(t, g, modelName, timeout)
	TearDownTestForIngress(t, ns, modelName)
}

// test ingress with multihost create in a single namespace
func nIngressMultihostCreateTest(t *testing.T, ingrCount, hostCount, timeout int) {
	g := gomega.NewGomegaWithT(t)
	ns := "default"
	modelName := SetUpTestForIngress(t, ns)

	pooldata, ingrdata := generatePoolDataMultiHost(t, ns, ingrCount, hostCount, "", "", "")
	createIngresses(t, ns, ingrdata)
	verifyModel(t, g, hostCount*ingrCount, modelName, pooldata, timeout)

	delIngresses(t, ns, ingrdata)
	verifyIngressDelete(t, g, modelName, timeout)
	TearDownTestForIngress(t, ns, modelName)
}

// test ingress with multihost update in a single namespace
func nIngressMultihostUpdatePathTest(t *testing.T, ingrCount, hostCount, timeout, retry int) {
	g := gomega.NewGomegaWithT(t)
	ns := "default"
	modelName := SetUpTestForIngress(t, ns)

	_, ingrdata := generatePoolDataMultiHost(t, ns, ingrCount, hostCount, "", "", "")
	createIngresses(t, ns, ingrdata)
	pooldataUpdated, ingrdataUpdated := generatePoolDataMultiHost(t, ns, ingrCount, hostCount, "", "", "/foobar")
	updateIngresses(t, ns, ingrdataUpdated)
	verifyModelWithRetry(t, g, hostCount*ingrCount, modelName, pooldataUpdated, timeout, retry)

	delIngresses(t, ns, ingrdata)
	verifyIngressDelete(t, g, modelName, timeout)
	TearDownTestForIngress(t, ns, modelName)
}

// test ingress create in multiple namespaces
func nIngressMultiNSIngressCreateTest(t *testing.T, ingrCount, nsCount, timeout int) {
	g := gomega.NewGomegaWithT(t)
	nsPrefix := "red"
	var nsList []string
	for i := 1; i <= nsCount; i++ {
		ns := fmt.Sprintf("%s-%d", nsPrefix, i)
		nsList = append(nsList, ns)
	}

	modelToPools := make(map[string]map[string]poolData)
	allIngrData := make(map[string]map[string]ingrData)
	allModels := make(map[string]string)

	for _, ns := range nsList {
		modelName := SetUpTestForIngress(t, ns)
		allModels[ns] = modelName
		pooldata, ingrdata := generatePoolData(t, ns, ingrCount, "", "", "")
		allIngrData[ns] = ingrdata
		createIngresses(t, ns, ingrdata)
		if modelToPools[modelName] == nil {
			modelToPools[modelName] = make(map[string]poolData)
		}
		for poolname := range pooldata {
			modelToPools[modelName][poolname] = pooldata[poolname]
		}
	}

	verifyMultiModel(t, g, modelToPools, timeout)
	for _, ns := range nsList {
		delIngresses(t, ns, allIngrData[ns])
	}
	verifyMultiModelIngressDelete(t, g, modelToPools, timeout)

	for _, ns := range nsList {
		TearDownTestForIngress(t, ns, allModels[ns])
	}
}

// test ingress update in multiple namespaces
func nIngressMultiNSIngressUpdatePathTest(t *testing.T, ingrCount, nsCount, timeout, retry int) {
	g := gomega.NewGomegaWithT(t)
	nsPrefix := "red"
	var nsList []string
	for i := 1; i <= nsCount; i++ {
		ns := fmt.Sprintf("%s-%d", nsPrefix, i)
		nsList = append(nsList, ns)
	}

	modelToPools := make(map[string]map[string]poolData)
	allIngrData := make(map[string]map[string]ingrData)
	allModels := make(map[string]string)

	for _, ns := range nsList {
		modelName := SetUpTestForIngress(t, ns)
		allModels[ns] = modelName
		_, ingrdata := generatePoolData(t, ns, ingrCount, "", "", "")
		allIngrData[ns] = ingrdata
		createIngresses(t, ns, ingrdata)

		pooldataUpdated, ingrdataUpdated := generatePoolData(t, ns, ingrCount, "", "", "foobar")
		updateIngresses(t, ns, ingrdataUpdated)
		if modelToPools[modelName] == nil {
			modelToPools[modelName] = make(map[string]poolData)
		}
		for poolname := range pooldataUpdated {
			modelToPools[modelName][poolname] = pooldataUpdated[poolname]
		}
		allIngrData[ns] = ingrdataUpdated
	}

	verifyMultiModelWithRetry(t, g, modelToPools, timeout, retry)
	for _, ns := range nsList {
		delIngresses(t, ns, allIngrData[ns])
	}
	verifyMultiModelIngressDelete(t, g, modelToPools, timeout)

	for _, ns := range nsList {
		TearDownTestForIngress(t, ns, allModels[ns])
	}
}

func Test100IngressCreate(t *testing.T) {
	nIngressCreateTest(t, 100, 100)
}

func Test100IngressUpdatePath(t *testing.T) {
	nIngressUpdatePathTest(t, 100, 100, 20)
}

func Test100MultipathIngressCreate(t *testing.T) {
	nIngressMultipathCreateTest(t, 100, 10, 100)
}

func Test100MultipathIngressUpdatePath(t *testing.T) {
	nIngressMultipathUpdatePathTest(t, 100, 10, 100, 200)
}

func Test100MultihostIngressCreate(t *testing.T) {
	nIngressMultihostCreateTest(t, 100, 10, 100)
}

func Test100MultihostIngressUpdatePath(t *testing.T) {
	nIngressMultihostUpdatePathTest(t, 100, 10, 100, 200)
}

func Test10X10MultiNSIngressCreate(t *testing.T) {
	nIngressMultiNSIngressCreateTest(t, 10, 10, 100)
}

func Test10X10MultiNSIngressUpdatePath(t *testing.T) {
	nIngressMultiNSIngressUpdatePathTest(t, 10, 10, 100, 200)
}

func Test1X100MultiNSIngressCreate(t *testing.T) {
	nIngressMultiNSIngressCreateTest(t, 1, 100, 100)
}

func Test1X100MultiNSIngressUpdatePath(t *testing.T) {
	nIngressMultiNSIngressUpdatePathTest(t, 1, 100, 100, 200)
}

// hostname shard: test ingress create with single host
func nIngressCreateTestHostnameShard(t *testing.T, count, timeout int) {
	g := gomega.NewGomegaWithT(t)
	ns := "default"
	SetUpHostnameShardTestforIngress(t, ns)

	modelToPools := make(map[string]map[string]poolData)
	pooldata, ingrdata := generatePoolData(t, ns, count, "", "", "")
	for ingrname, ingr := range ingrdata {
		host := ingr.hosts[0]
		path := ingr.paths[0]
		modelname := "admin/" + avinodes.DeriveHostNameShardVS(host, "test")
		poolname := "cluster--" + host + strings.Replace(path, "/", "_", 1) + "-" + ns + "-" + ingrname
		if modelToPools[modelname] == nil {
			modelToPools[modelname] = make(map[string]poolData)
		}
		modelToPools[modelname][poolname] = pooldata[poolname]
	}
	createIngresses(t, ns, ingrdata)
	verifyMultiModel(t, g, modelToPools, timeout)

	delIngresses(t, ns, ingrdata)
	verifyMultiModelIngressDelete(t, g, modelToPools, timeout)
	TearDownHostnameShardTestForIngress(t, ns, modelToPools)
}

// hostname shard: test ingress update with single host
func nIngressUpdateTestHostnameShard(t *testing.T, count, timeout, retry int) {
	g := gomega.NewGomegaWithT(t)
	ns := "default"
	SetUpHostnameShardTestforIngress(t, ns)
	modelToPools := make(map[string]map[string]poolData)

	pooldata, ingrdata := generatePoolData(t, ns, count, "", "", "")
	createIngresses(t, ns, ingrdata)

	pooldata, ingrdata = generatePoolData(t, ns, count, "", "", "/foobar")
	for ingrname, ingr := range ingrdata {
		host := ingr.hosts[0]
		path := ingr.paths[0]
		modelname := "admin/" + avinodes.DeriveHostNameShardVS(host, "test")
		poolname := "cluster--" + host + strings.Replace(path, "/", "_", 1) + "-" + ns + "-" + ingrname
		if modelToPools[modelname] == nil {
			modelToPools[modelname] = make(map[string]poolData)
		}
		modelToPools[modelname][poolname] = pooldata[poolname]
	}
	updateIngresses(t, ns, ingrdata)
	verifyMultiModelWithRetry(t, g, modelToPools, timeout, retry)

	delIngresses(t, ns, ingrdata)
	verifyMultiModelIngressDelete(t, g, modelToPools, timeout)
	TearDownHostnameShardTestForIngress(t, ns, modelToPools)
}

// hostname shard: test ingress create with multiple host
func nIngressMultihostHostnameShardCreateTest(t *testing.T, ingrCount, hostCount, timeout int) {
	g := gomega.NewGomegaWithT(t)
	ns := "default"

	SetUpHostnameShardTestforIngress(t, ns)

	modelToPools := make(map[string]map[string]poolData)
	pooldata, ingrdata := generatePoolDataMultiHost(t, ns, ingrCount, hostCount, "", "", "")
	for ingrname, ingr := range ingrdata {
		for i := range ingr.hosts {
			host := ingr.hosts[i]
			path := ingr.paths[i]
			modelname := "admin/" + avinodes.DeriveHostNameShardVS(host, "test")
			poolname := "cluster--" + host + strings.Replace(path, "/", "_", 1) + "-" + ns + "-" + ingrname
			if modelToPools[modelname] == nil {
				modelToPools[modelname] = make(map[string]poolData)
			}
			modelToPools[modelname][poolname] = pooldata[poolname]
		}
	}
	createIngresses(t, ns, ingrdata)
	verifyMultiModel(t, g, modelToPools, timeout)

	delIngresses(t, ns, ingrdata)
	verifyMultiModelIngressDelete(t, g, modelToPools, timeout)
	TearDownHostnameShardTestForIngress(t, ns, modelToPools)
}

// hostname shard: test ingress update with multiple host
func nIngressMultihostHostnameShardUpdateTest(t *testing.T, ingrCount, hostCount, timeout, retry int) {
	g := gomega.NewGomegaWithT(t)
	ns := "default"

	os.Setenv("L7_SHARD_SCHEME", "hostname")
	SetUpHostnameShardTestforIngress(t, ns)

	modelToPools := make(map[string]map[string]poolData)
	pooldata, ingrdata := generatePoolDataMultiHost(t, ns, ingrCount, hostCount, "", "", "")
	createIngresses(t, ns, ingrdata)
	pooldata, ingrdata = generatePoolDataMultiHost(t, ns, ingrCount, hostCount, "", "", "/foobar")
	for ingrname, ingr := range ingrdata {
		for i := range ingr.hosts {
			host := ingr.hosts[i]
			path := ingr.paths[i]
			modelname := "admin/" + avinodes.DeriveHostNameShardVS(host, "test")
			poolname := "cluster--" + host + strings.Replace(path, "/", "_", 1) + "-" + ns + "-" + ingrname
			if modelToPools[modelname] == nil {
				modelToPools[modelname] = make(map[string]poolData)
			}
			modelToPools[modelname][poolname] = pooldata[poolname]
		}
	}
	updateIngresses(t, ns, ingrdata)
	verifyMultiModelWithRetry(t, g, modelToPools, timeout, retry)

	delIngresses(t, ns, ingrdata)
	verifyMultiModelIngressDelete(t, g, modelToPools, timeout)
	TearDownHostnameShardTestForIngress(t, ns, modelToPools)
}

func Test100IngressCreateHostnameShard(t *testing.T) {
	nIngressCreateTestHostnameShard(t, 100, 100)
}

func Test100IngressUpdateHostnameShard(t *testing.T) {
	nIngressUpdateTestHostnameShard(t, 100, 100, 200)
}

func Test100X10IngressCreateMultiHostHostnameShard(t *testing.T) {
	nIngressMultihostHostnameShardCreateTest(t, 100, 10, 100)
}

func Test100X10IngressUpdateMultiHostHostnameShard(t *testing.T) {
	nIngressMultihostHostnameShardUpdateTest(t, 100, 10, 100, 200)
}
