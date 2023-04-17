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

package integrationtest

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/fatih/structs"
	"github.com/onsi/gomega"
	"golang.org/x/exp/maps"
)

func validateCRDValues(t *testing.T, g *gomega.GomegaWithT, expectedValues interface{}, actualValues ...interface{}) {
	valuesfromCRD := structs.Map(expectedValues)
	valuesFromGraphLayer := make(map[string]interface{})
	for _, actualValue := range actualValues {
		actualValueMap := structs.Map(actualValue)
		maps.Copy(valuesFromGraphLayer, actualValueMap)
	}

	for field, valuefromCRD := range valuesfromCRD {

		t.Logf("Current Field: %s", field)
		valueFromGraphLayer, ok := valuesFromGraphLayer[field]
		if !ok {
			continue
		}

		if strings.HasSuffix(field, "Ref") {
			ref1 := valueFromGraphLayer.(*string)
			ref2 := valuefromCRD.(*string)
			g.Expect(ref1).NotTo(gomega.BeNil())
			g.Expect(ref2).NotTo(gomega.BeNil())
			g.Expect(*ref1).To(gomega.HaveSuffix(*ref2))
		} else if strings.HasSuffix(field, "Refs") {
			refs1 := valueFromGraphLayer.([]string)
			refs2 := valuefromCRD.([]string)
			g.Expect(len(refs1)).To(gomega.Equal(len(refs2)))
			l := len(refs1)
			for i := 0; i < l; i++ {
				g.Expect(refs1[i]).To(gomega.HaveSuffix(refs2[i]))
			}
		} else if field == "AnalyticsPolicy" {
			// All fields of AnalyticsPolicy are not used currently. Hence
			// handling it separately.
			actualAnalyticsPolicy := valueFromGraphLayer.(map[string]interface{})
			expectedAnalyticsPolicy := valuefromCRD.(map[string]interface{})
			// For backendProperties, the PoolAnalyticsPolicy only have one field EnableRealtimeMetrics
			if _, ok := expectedAnalyticsPolicy["EnableRealtimeMetrics"]; ok {
				g.Expect(expectedAnalyticsPolicy["EnableRealtimeMetrics"]).To(gomega.Equal(actualAnalyticsPolicy["EnableRealtimeMetrics"]))
			} else {
				g.Expect(expectedAnalyticsPolicy["FullClientLogs"]).To(gomega.Equal(actualAnalyticsPolicy["FullClientLogs"]))
			}
		} else {
			g.Expect(utils.Stringify(valuefromCRD)).To(gomega.Equal(utils.Stringify(valueFromGraphLayer)))
		}
	}
}

func TestCreateDeleteL4Rule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := "test-l4rule"
	ports := []int{8080}

	SetUpTestForSvcLB(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// Create the L4Rule
	SetupL4Rule(t, L4RuleName, NAMESPACE, ports)

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// Apply the  L4Rule to Service
	svcObj := (FakeService{
		Name:         SINGLEPORTSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcObj.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName}
	svcObj.ResourceVersion = "2"
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).NotTo(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	l4Rule := FakeL4Rule{
		Name:      L4RuleName,
		Namespace: NAMESPACE,
		Ports:     ports,
	}
	obj := l4Rule.L4Rule()

	_, aviModel = objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	validateCRDValues(t, g, obj.Spec, nodes[0].AviVsNodeGeneratedFields, nodes[0].AviVsNodeCommonFields)

	validateCRDValues(t, g, obj.Spec.BackendProperties[0],
		nodes[0].PoolRefs[0].AviPoolGeneratedFields, nodes[0].PoolRefs[0].AviPoolCommonFields)

	// Remove the  L4Rule from Service
	svcObj.Annotations = nil
	svcObj.ResourceVersion = "3"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestUpdateDeleteL4Rule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := "test-l4rule"
	ports := []int{8080}

	SetUpTestForSvcLB(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// Create the L4Rule
	SetupL4Rule(t, L4RuleName, NAMESPACE, ports)

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// Apply the  L4Rule to Service
	svcObj := (FakeService{
		Name:         SINGLEPORTSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcObj.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName}
	svcObj.ResourceVersion = "2"
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).NotTo(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	l4Rule := FakeL4Rule{
		Name:      L4RuleName,
		Namespace: NAMESPACE,
		Ports:     ports,
	}
	obj := l4Rule.L4Rule()

	_, aviModel = objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	validateCRDValues(t, g, obj.Spec, nodes[0].AviVsNodeGeneratedFields, nodes[0].AviVsNodeCommonFields)

	validateCRDValues(t, g, obj.Spec.BackendProperties[0],
		nodes[0].PoolRefs[0].AviPoolGeneratedFields, nodes[0].PoolRefs[0].AviPoolCommonFields)

	// Update the L4Rule object
	obj.Spec.PerformanceLimits.MaxConcurrentConnections = proto.Int32(100)
	obj.Spec.PerformanceLimits.MaxThroughput = proto.Int32(30)
	obj.Spec.VsDatascriptRefs = []string{"thisisaviref--new-ds1", "thisisaviref-new-ds2"}
	obj.Spec.BackendProperties[0].InlineHealthMonitor = proto.Bool(false)
	obj.Spec.BackendProperties[0].HealthMonitorRefs = []string{"thisisaviref-new-hm1", "thisisaviref-new-hm2"}
	obj.Spec.BackendProperties[0].DefaultServerPort = proto.Int32(9090)
	obj.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Update(context.TODO(), obj, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating L4Rule: %v", err)
	}

	// Adding a sleep since the model will be present.
	time.Sleep(5 * time.Second)

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).NotTo(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	validateCRDValues(t, g, obj.Spec, nodes[0].AviVsNodeGeneratedFields, nodes[0].AviVsNodeCommonFields)

	validateCRDValues(t, g, obj.Spec.BackendProperties[0],
		nodes[0].PoolRefs[0].AviPoolGeneratedFields, nodes[0].PoolRefs[0].AviPoolCommonFields)

	// Remove the  L4Rule from Service
	svcObj.Annotations = nil
	svcObj.ResourceVersion = "3"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestL4RuleWithWrongPortInBackendProperties(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := "test-l4rule"
	ports := []int{8081}

	// Create the service
	SetUpTestForSvcLB(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SINGLEPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// Create the L4Rule with port 8081
	SetupL4Rule(t, L4RuleName, NAMESPACE, ports)

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// Apply the  L4Rule to Service
	svcObj := (FakeService{
		Name:         SINGLEPORTSVC,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcObj.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName}
	svcObj.ResourceVersion = "2"
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating the service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).NotTo(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	l4Rule := FakeL4Rule{
		Name:      L4RuleName,
		Namespace: NAMESPACE,
		Ports:     ports,
	}
	obj := l4Rule.L4Rule()

	_, aviModel = objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	validateCRDValues(t, g, obj.Spec, nodes[0].AviVsNodeGeneratedFields, nodes[0].AviVsNodeCommonFields)

	// Pool properties won't get added due to the mismatch of port and protocol
	g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero())
	g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero())

	// Remove the  L4Rule from Service
	svcObj.Annotations = nil
	svcObj.ResourceVersion = "3"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(SINGLEPORTMODEL)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestCreateDeleteL4RuleMultiportSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := "test-l4rule"
	ports := []int{8080, 8081, 8082}

	SetUpTestForSvcLBMultiport(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(MULTIPORTMODEL)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(MULTIPORTMODEL)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, MULTIPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(3))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].PortProto[1].Port).To(gomega.Equal(int32(8081)))
	g.Expect(nodes[0].PortProto[2].Port).To(gomega.Equal(int32(8082)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(3))
	for _, node := range nodes[0].PoolRefs {
		if node.Port == 8080 {
			address := "1.1.1.1"
			g.Expect(node.Servers).To(gomega.HaveLen(3))
			g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
		} else if node.Port == 8081 {
			address := "1.1.1.4"
			g.Expect(node.Servers).To(gomega.HaveLen(2))
			g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
		} else if node.Port == 8082 {
			address := "1.1.1.6"
			g.Expect(node.Servers).To(gomega.HaveLen(1))
			g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
		}
	}

	// Create the L4Rule for pools with port 8080, 8081, 8082
	SetupL4Rule(t, L4RuleName, NAMESPACE, ports)

	// Validate the CRD status
	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// Apply the  L4Rule to Service
	svcObj, err := KubeClient.CoreV1().Services(NAMESPACE).Get(context.TODO(), MULTIPORTSVC, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}
	svcObj.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName}
	svcObj.ResourceVersion = "2"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(MULTIPORTMODEL)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(3)) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[1].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[1].AviPoolGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[2].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[2].AviPoolGeneratedFields).NotTo(gomega.BeZero())
	}, 60*time.Second).Should(gomega.Equal(true))

	l4Rule := FakeL4Rule{
		Name:      L4RuleName,
		Namespace: NAMESPACE,
		Ports:     ports,
	}
	obj := l4Rule.L4Rule()

	_, aviModel = objects.SharedAviGraphLister().Get(MULTIPORTMODEL)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	validateCRDValues(t, g, obj.Spec, nodes[0].AviVsNodeGeneratedFields, nodes[0].AviVsNodeCommonFields)

	// Validate for pool with port 8080
	validateCRDValues(t, g, obj.Spec.BackendProperties[0],
		nodes[0].PoolRefs[0].AviPoolGeneratedFields, nodes[0].PoolRefs[0].AviPoolCommonFields)

	// Validate for pool with port 8081
	validateCRDValues(t, g, obj.Spec.BackendProperties[1],
		nodes[0].PoolRefs[1].AviPoolGeneratedFields, nodes[0].PoolRefs[1].AviPoolCommonFields)

	// Validate for pool with port 8082
	validateCRDValues(t, g, obj.Spec.BackendProperties[2],
		nodes[0].PoolRefs[2].AviPoolGeneratedFields, nodes[0].PoolRefs[2].AviPoolCommonFields)

	// Remove the  L4Rule from Service
	svcObj.Annotations = nil
	svcObj.ResourceVersion = "3"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(MULTIPORTMODEL)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(3)) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[1].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[1].AviPoolGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[2].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[2].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLBMultiport(t, g)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestUpdateDeleteL4RuleMultiportSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := "test-l4rule"
	ports := []int{8080, 8081, 8082}

	SetUpTestForSvcLBMultiport(t)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(MULTIPORTMODEL)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(MULTIPORTMODEL)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, MULTIPORTSVC)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto).To(gomega.HaveLen(3))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].PortProto[1].Port).To(gomega.Equal(int32(8081)))
	g.Expect(nodes[0].PortProto[2].Port).To(gomega.Equal(int32(8082)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(3))
	for _, node := range nodes[0].PoolRefs {
		if node.Port == 8080 {
			address := "1.1.1.1"
			g.Expect(node.Servers).To(gomega.HaveLen(3))
			g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
		} else if node.Port == 8081 {
			address := "1.1.1.4"
			g.Expect(node.Servers).To(gomega.HaveLen(2))
			g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
		} else if node.Port == 8082 {
			address := "1.1.1.6"
			g.Expect(node.Servers).To(gomega.HaveLen(1))
			g.Expect(node.Servers[0].Ip.Addr).To(gomega.Equal(&address))
		}
	}

	// Create the L4Rule for pools with port 8080, 8081, 8082
	SetupL4Rule(t, L4RuleName, NAMESPACE, ports)

	// Validate the CRD status
	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// Apply the  L4Rule to Service
	svcObj, err := KubeClient.CoreV1().Services(NAMESPACE).Get(context.TODO(), MULTIPORTSVC, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}
	svcObj.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName}
	svcObj.ResourceVersion = "2"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(MULTIPORTMODEL)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(3)) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[1].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[1].AviPoolGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[2].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[2].AviPoolGeneratedFields).NotTo(gomega.BeZero())
	}, 60*time.Second).Should(gomega.Equal(true))

	l4Rule := FakeL4Rule{
		Name:      L4RuleName,
		Namespace: NAMESPACE,
		Ports:     ports,
	}
	obj := l4Rule.L4Rule()

	_, aviModel = objects.SharedAviGraphLister().Get(MULTIPORTMODEL)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	validateCRDValues(t, g, obj.Spec, nodes[0].AviVsNodeGeneratedFields, nodes[0].AviVsNodeCommonFields)

	// Validate for pool with port 8080
	validateCRDValues(t, g, obj.Spec.BackendProperties[0],
		nodes[0].PoolRefs[0].AviPoolGeneratedFields, nodes[0].PoolRefs[0].AviPoolCommonFields)

	// Validate for pool with port 8081
	validateCRDValues(t, g, obj.Spec.BackendProperties[1],
		nodes[0].PoolRefs[1].AviPoolGeneratedFields, nodes[0].PoolRefs[1].AviPoolCommonFields)

	// Validate for pool with port 8082
	validateCRDValues(t, g, obj.Spec.BackendProperties[2],
		nodes[0].PoolRefs[2].AviPoolGeneratedFields, nodes[0].PoolRefs[2].AviPoolCommonFields)

	// Remove the properties corresponding to 8081 port
	l4Rule = FakeL4Rule{
		Name:      L4RuleName,
		Namespace: NAMESPACE,
		Ports:     []int{8080, 8082},
	}
	obj = l4Rule.L4Rule()
	obj.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Update(context.TODO(), obj, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating L4Rule: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(MULTIPORTMODEL)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].PoolRefs[1].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[1].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 60*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(MULTIPORTMODEL)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	validateCRDValues(t, g, obj.Spec, nodes[0].AviVsNodeGeneratedFields, nodes[0].AviVsNodeCommonFields)

	// Validate for pool with port 8080
	validateCRDValues(t, g, obj.Spec.BackendProperties[0],
		nodes[0].PoolRefs[0].AviPoolGeneratedFields, nodes[0].PoolRefs[0].AviPoolCommonFields)

	// Validate for pool with port 8082
	validateCRDValues(t, g, obj.Spec.BackendProperties[1],
		nodes[0].PoolRefs[2].AviPoolGeneratedFields, nodes[0].PoolRefs[2].AviPoolCommonFields)

	// Remove the  L4Rule from Service
	svcObj.Annotations = nil
	svcObj.ResourceVersion = "3"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(MULTIPORTMODEL)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(3)) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[1].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[1].AviPoolGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[2].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[2].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLBMultiport(t, g)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}
