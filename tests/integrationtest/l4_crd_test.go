/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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

	"github.com/fatih/structs"
	"github.com/onsi/gomega"
	"github.com/vmware/alb-sdk/go/models"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1alpha2 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha2"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
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
			if ref2 == nil {
				g.Expect(ref1).To(gomega.BeNil())
				g.Expect(ref2).To(gomega.BeNil())
			} else {
				g.Expect(ref1).NotTo(gomega.BeNil())
				g.Expect(ref2).NotTo(gomega.BeNil())
				if *ref2 == "System-L4-Application" {
					g.Expect(*ref1).To(gomega.HaveSuffix("System-SSL-Application"))
				} else {
					g.Expect(*ref1).To(gomega.HaveSuffix(*ref2))
				}
			}
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

			// Try to type cast valueFromGraphLayer to pointer types first
			if actualAnalyticsPolicyStruct, ok := valueFromGraphLayer.(*models.AnalyticsPolicy); ok {
				// Handle AVI models.AnalyticsPolicy struct (from VS model)
				if valuefromCRD == nil {
					// When L4Rule is invalid/default, AnalyticsPolicy should be nil
					g.Expect(actualAnalyticsPolicyStruct).To(gomega.BeNil())
				} else if expectedAnalyticsPolicy, ok := valuefromCRD.(*akov1alpha2.AnalyticsPolicy); ok {
					// VS-level AnalyticsPolicy comparison
					if expectedAnalyticsPolicy != nil && expectedAnalyticsPolicy.FullClientLogs != nil && actualAnalyticsPolicyStruct != nil && actualAnalyticsPolicyStruct.FullClientLogs != nil {
						g.Expect(expectedAnalyticsPolicy.FullClientLogs.Enabled).To(gomega.Equal(actualAnalyticsPolicyStruct.FullClientLogs.Enabled))
						g.Expect(expectedAnalyticsPolicy.FullClientLogs.Duration).To(gomega.Equal(actualAnalyticsPolicyStruct.FullClientLogs.Duration))
						g.Expect(expectedAnalyticsPolicy.FullClientLogs.Throttle).To(gomega.Equal(actualAnalyticsPolicyStruct.FullClientLogs.Throttle))
					}
				}
			} else if actualPoolAnalyticsPolicyStruct, ok := valueFromGraphLayer.(*models.PoolAnalyticsPolicy); ok {
				// Handle AVI models.PoolAnalyticsPolicy struct (from Pool model)
				if valuefromCRD == nil {
					// When L4Rule is invalid/default, PoolAnalyticsPolicy should be nil
					g.Expect(actualPoolAnalyticsPolicyStruct).To(gomega.BeNil())
				} else if expectedPoolAnalyticsPolicy, ok := valuefromCRD.(*akov1alpha2.PoolAnalyticsPolicy); ok {
					// Pool-level AnalyticsPolicy comparison
					if expectedPoolAnalyticsPolicy != nil && expectedPoolAnalyticsPolicy.EnableRealtimeMetrics != nil && actualPoolAnalyticsPolicyStruct != nil {
						g.Expect(expectedPoolAnalyticsPolicy.EnableRealtimeMetrics).To(gomega.Equal(actualPoolAnalyticsPolicyStruct.EnableRealtimeMetrics))
					}
				}
			} else if actualAnalyticsPolicyMap, ok := valueFromGraphLayer.(map[string]interface{}); ok {
				// Fallback to map[string]interface{} handling for backward compatibility
				if valuefromCRD == nil {
					// When L4Rule is invalid/default, AnalyticsPolicy should be nil
					g.Expect(actualAnalyticsPolicyMap).To(gomega.BeNil())
				} else {
					expectedAnalyticsPolicy := valuefromCRD.(map[string]interface{})
					// For backendProperties, the PoolAnalyticsPolicy only have one field EnableRealtimeMetrics
					if _, ok := expectedAnalyticsPolicy["EnableRealtimeMetrics"]; ok {
						g.Expect(expectedAnalyticsPolicy["EnableRealtimeMetrics"]).To(gomega.Equal(actualAnalyticsPolicyMap["EnableRealtimeMetrics"]))
					} else {
						g.Expect(expectedAnalyticsPolicy["FullClientLogs"]).To(gomega.Equal(actualAnalyticsPolicyMap["FullClientLogs"]))
					}
				}
			} else {
				// Direct comparison for other types (including nil)
				g.Expect(valuefromCRD).To(gomega.Equal(valueFromGraphLayer))
			}
		} else {
			g.Expect(utils.Stringify(valuefromCRD)).To(gomega.Equal(utils.Stringify(valueFromGraphLayer)))
		}
	}
}

func TestCreateDeleteL4RuleInvalidLBClass(t *testing.T) {
	// this test checks the following scenario:
	// adding a valid l4rule crd annotation to an invalid LBSvc should not cause the VS to come up
	g := gomega.NewGomegaWithT(t)
	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080}
	lib.AKOControlConfig().SetDefaultLBController(true)
	// test invalid service spec.LoadBalancerClass != ako.vmware.com/avi-lb for DefaultLBContoller == true
	SetUpTestForSvcLBWithLBClass(t, INVALID_LB_CLASS, svcName)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(false))

	SetupL4Rule(t, L4RuleName, NAMESPACE, ports)
	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	svcObj := (FakeService{
		Name:         svcName,
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
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(false))
	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)

	// test invalid service spec.LoadBalancerClass != ako.vmware.com/avi-lb for DefaultLBContoller == false
	lib.AKOControlConfig().SetDefaultLBController(false)
	SetUpTestForSvcLBWithLBClass(t, INVALID_LB_CLASS, svcName)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(false))

	SetupL4Rule(t, L4RuleName, NAMESPACE, ports)
	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	svcObj = (FakeService{
		Name:         svcName,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcObj.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName}
	svcObj.ResourceVersion = "2"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(false))
	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)

	// test invalid service with spec.LoadBalancerClass empty for defaultLBController == false
	SetUpTestForSvcLB(t, svcName)
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(false))

	SetupL4Rule(t, L4RuleName, NAMESPACE, ports)
	PollForCompletion(t, modelName, 15)
	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	svcObj = (FakeService{
		Name:         svcName,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcObj.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName}
	svcObj.ResourceVersion = "2"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(false))
	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)

	lib.AKOControlConfig().SetDefaultLBController(true)
}
func TestCreateDeleteL4Rule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080}

	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
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
		Name:         svcName,
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
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

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestUpdateDeleteL4Rule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080}

	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
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
		Name:         svcName,
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
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

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	validateCRDValues(t, g, obj.Spec, nodes[0].AviVsNodeGeneratedFields, nodes[0].AviVsNodeCommonFields)

	validateCRDValues(t, g, obj.Spec.BackendProperties[0],
		nodes[0].PoolRefs[0].AviPoolGeneratedFields, nodes[0].PoolRefs[0].AviPoolCommonFields)

	// Update the L4Rule object
	obj.Spec.PerformanceLimits.MaxConcurrentConnections = proto.Int32(100)
	obj.Spec.PerformanceLimits.MaxThroughput = proto.Int32(30)
	obj.Spec.VsDatascriptRefs = []string{"thisisaviref--new-ds1", "thisisaviref-new-ds2"}
	obj.Spec.BackendProperties[0].MinServersUp = proto.Uint32(2)
	obj.Spec.BackendProperties[0].HealthMonitorRefs = []string{"thisisaviref-new-hm1", "thisisaviref-new-hm2"}
	obj.Spec.BackendProperties[0].Enabled = proto.Bool(false)
	obj.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Update(context.TODO(), obj, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in updating L4Rule: %v", err)
	}

	// Adding a sleep since the model will be present.
	time.Sleep(5 * time.Second)

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestL4RuleWithWrongPortInBackendProperties(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8081}

	// Create the service
	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
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
		Name:         svcName,
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
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

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestCreateDeleteL4RuleMultiportSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(MULTIPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080, 8081, 8082}

	SetUpTestForSvcLBMultiport(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
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
	svcObj, err := KubeClient.CoreV1().Services(NAMESPACE).Get(context.TODO(), svcName, metav1.GetOptions{})
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
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

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
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

	TearDownTestForSvcLBMultiport(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestUpdateDeleteL4RuleMultiportSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(MULTIPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080, 8081, 8082}

	SetUpTestForSvcLBMultiport(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
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
	svcObj, err := KubeClient.CoreV1().Services(NAMESPACE).Get(context.TODO(), svcName, metav1.GetOptions{})
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
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

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].PoolRefs[1].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[1].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 60*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
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

	TearDownTestForSvcLBMultiport(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestInvalidToValidL4Rule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080}

	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
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
		Name:         svcName,
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
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

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	validateCRDValues(t, g, obj.Spec, nodes[0].AviVsNodeGeneratedFields, nodes[0].AviVsNodeCommonFields)

	validateCRDValues(t, g, obj.Spec.BackendProperties[0],
		nodes[0].PoolRefs[0].AviPoolGeneratedFields, nodes[0].PoolRefs[0].AviPoolCommonFields)

	// Update the L4Rule with an invalid application profile reference.
	obj.Spec.BackendProperties[0].ApplicationPersistenceProfileRef = proto.String("invalid-profile-ref")
	obj.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Update(context.TODO(), obj, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in adding L4Rule: %v", err)
	}

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Rejected"))

	time.Sleep(5 * time.Second)

	// Verify that the properties revert to defaults when L4Rule is invalid
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	// Create empty/default L4Rule spec to validate that VS reverts to default properties
	defaultL4RuleSpec := akov1alpha2.L4RuleSpec{}
	defaultBackendProperties := akov1alpha2.BackendProperties{}

	validateCRDValues(t, g, defaultL4RuleSpec, nodes[0].AviVsNodeGeneratedFields, nodes[0].AviVsNodeCommonFields)

	validateCRDValues(t, g, defaultBackendProperties,
		nodes[0].PoolRefs[0].AviPoolGeneratedFields, nodes[0].PoolRefs[0].AviPoolCommonFields)

	// Update L4Rule with a valid application persistence profile.
	obj.Spec.BackendProperties[0].ApplicationPersistenceProfileRef = proto.String("thisisaviref-applicationpersistenceprofileref")
	obj.ResourceVersion = "3"
	if _, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Update(context.TODO(), obj, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in adding L4Rule: %v", err)
	}

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	time.Sleep(5 * time.Second)

	// Re-verify the models
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestL4RuleLbAlgorithm(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080}

	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// Create an L4Rule with LbAlgorithm as LB_ALGORITHM_ROUND_ROBIN
	l4Rule := FakeL4Rule{
		Name:      L4RuleName,
		Namespace: NAMESPACE,
		Ports:     ports,
	}
	obj := l4Rule.L4Rule()
	obj.Spec.BackendProperties[0].LbAlgorithm = proto.String("LB_ALGORITHM_ROUND_ROBIN")
	obj.Spec.BackendProperties[0].LbAlgorithmHash = nil
	obj.Spec.BackendProperties[0].LbAlgorithmConsistentHashHdr = nil
	if _, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Create(context.TODO(), obj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding L4Rule: %v", err)
	}

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// Apply the  L4Rule to Service
	svcObj := (FakeService{
		Name:         svcName,
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).NotTo(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	validateCRDValues(t, g, obj.Spec, nodes[0].AviVsNodeGeneratedFields, nodes[0].AviVsNodeCommonFields)

	validateCRDValues(t, g, obj.Spec.BackendProperties[0],
		nodes[0].PoolRefs[0].AviPoolGeneratedFields, nodes[0].PoolRefs[0].AviPoolCommonFields)

	// Update the L4Rule with LbAlgorithm as LB_ALGORITHM_CONSISTENT_HASH without lbAlgorithmHash
	obj.Spec.BackendProperties[0].LbAlgorithm = proto.String("LB_ALGORITHM_CONSISTENT_HASH")
	obj.ResourceVersion = "2"
	if _, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Update(context.TODO(), obj, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in adding L4Rule: %v", err)
	}

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Rejected"))

	// Update the L4Rule with LbAlgorithm as LB_ALGORITHM_CONSISTENT_HASH with lbAlgorithmHash
	// as LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER and without lbAlgorithmConsistentHashHdr
	obj.Spec.BackendProperties[0].LbAlgorithm = proto.String("LB_ALGORITHM_CONSISTENT_HASH")
	obj.Spec.BackendProperties[0].LbAlgorithmHash = proto.String("LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER")
	obj.ResourceVersion = "3"
	if _, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Update(context.TODO(), obj, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in adding L4Rule: %v", err)
	}

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Rejected"))

	time.Sleep(5 * time.Second)

	// Verify that the properties revert to defaults when L4Rule is invalid
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	// Create empty/default L4Rule spec to validate that VS reverts to default properties
	defaultL4RuleSpec := akov1alpha2.L4RuleSpec{}
	defaultBackendProperties := akov1alpha2.BackendProperties{}

	validateCRDValues(t, g, defaultL4RuleSpec, nodes[0].AviVsNodeGeneratedFields, nodes[0].AviVsNodeCommonFields)

	validateCRDValues(t, g, defaultBackendProperties,
		nodes[0].PoolRefs[0].AviPoolGeneratedFields, nodes[0].PoolRefs[0].AviPoolCommonFields)

	// Update the L4Rule with LbAlgorithm as LB_ALGORITHM_CONSISTENT_HASH and lbAlgorithmHash as
	// LB_ALGORITHM_CONSISTENT_HASH_URI.
	obj.Spec.BackendProperties[0].LbAlgorithm = proto.String("LB_ALGORITHM_CONSISTENT_HASH")
	obj.Spec.BackendProperties[0].LbAlgorithmHash = proto.String("LB_ALGORITHM_CONSISTENT_HASH_URI")
	obj.ResourceVersion = "4"
	if _, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Update(context.TODO(), obj, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("error in adding L4Rule: %v", err)
	}

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	time.Sleep(5 * time.Second)

	// Verify whether the properties are updated.
	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestSharedVIPSvcWithL4Rule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	ports := []int{8080}
	modelName := MODEL_REDNS_PREFIX + SHAREDVIPKEY

	SetUpTestForSharedVIPSvcLB(t, corev1.ProtocolTCP, corev1.ProtocolTCP)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SHAREDVIPKEY)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(2))
	addresses := map[string]struct{}{
		"1.1.1.1": {},
		"2.1.1.1": {},
	}
	for _, poolRef := range nodes[0].PoolRefs {
		ipAddr := poolRef.Servers[0].Ip.Addr
		delete(addresses, *ipAddr)
	}
	g.Expect(addresses).To(gomega.HaveLen(0))

	// Create the L4Rule
	SetupL4Rule(t, L4RuleName, NAMESPACE, ports)

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// Apply the L4Rule to first Service
	svcObj01 := (FakeService{
		Name:         SHAREDVIPSVC01,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcObj01.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName, lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	svcObj01.ResourceVersion = "2"
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj01, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	// Apply the L4Rule to second Service
	svcObj02 := (FakeService{
		Name:         SHAREDVIPSVC02,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcObj01.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName, lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	svcObj02.ResourceVersion = "2"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj02, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
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

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
	nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()

	validateCRDValues(t, g, obj.Spec, nodes[0].AviVsNodeGeneratedFields, nodes[0].AviVsNodeCommonFields)

	validateCRDValues(t, g, obj.Spec.BackendProperties[0],
		nodes[0].PoolRefs[0].AviPoolGeneratedFields, nodes[0].PoolRefs[0].AviPoolCommonFields)

	TearDownTestForSharedVIPSvcLB(t, g)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestSharedVIPSvcWithL4RuleTransition(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	modelName := MODEL_REDNS_PREFIX + SHAREDVIPKEY
	L4RuleName := objNameMap.GenerateName("test-l4rule")

	SetUpTestForSharedVIPSvcLB(t, corev1.ProtocolTCP, corev1.ProtocolUDP)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	VerfiyL4Node(nodes[0], g, "TCP", "UDP")

	l4Rule := &akov1alpha2.L4Rule{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       NAMESPACE,
			Name:            L4RuleName,
			ResourceVersion: "1",
		},
		Spec: akov1alpha2.L4RuleSpec{
			ApplicationProfileRef: proto.String("thisisaviref-l4-appprofile"),
		}}

	if _, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Create(context.TODO(), l4Rule, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding L4Rule: %v", err)
	}

	svcObj := ConstructService(NAMESPACE, SHAREDVIPSVC01, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	svcObj.ResourceVersion = "2"
	svcObj.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName, lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	svcObj = ConstructService(NAMESPACE, SHAREDVIPSVC02, corev1.ProtocolUDP, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	svcObj.ResourceVersion = "2"
	svcObj.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName, lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	g.Eventually(func() bool {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			if nodes[0].AviVsNodeCommonFields.ApplicationProfileRef != nil {
				return strings.Contains(*nodes[0].AviVsNodeCommonFields.ApplicationProfileRef, "thisisaviref-l4-appprofile")
			}
		}
		return false
	}, 30*time.Second).Should(gomega.Equal(true))
	VerfiyL4Node(nodes[0], g, "TCP", "UDP")

	// Initiating transition for one shared vip LB svc to type ClusterIP so the corresponfing pool and l4policyset should be deleted
	svcObj = ConstructService(NAMESPACE, SHAREDVIPSVC01, corev1.ProtocolTCP, corev1.ServiceTypeClusterIP, false, make(map[string]string), "")
	svcObj.ResourceVersion = "3"
	svcObj.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName, lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs)
		}
		return -1
	}, 30*time.Second).Should(gomega.Equal(1))
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SHAREDVIPKEY)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	g.Expect(nodes[0].PortProto[0].Protocol).To(gomega.Equal("UDP"))
	g.Expect(nodes[0].NetworkProfile).To(gomega.Equal(utils.SYSTEM_UDP_FAST_PATH))
	g.Expect(*nodes[0].AviVsNodeCommonFields.ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-l4-appprofile"))

	// Initiating transition for same shared vip ClusterIP svc back to LB so the corresponfing pool and l4policyset should be re-created
	svcObj = ConstructService(NAMESPACE, SHAREDVIPSVC01, corev1.ProtocolTCP, corev1.ServiceTypeLoadBalancer, false, make(map[string]string), "")
	svcObj.ResourceVersion = "4"
	svcObj.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName, lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in adding Service: %v", err)
	}

	g.Eventually(func() int {
		if found, aviModel := objects.SharedAviGraphLister().Get(modelName); found {
			nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
			return len(nodes[0].PoolRefs)
		}
		return -1
	}, 30*time.Second).Should(gomega.Equal(2))
	VerfiyL4Node(nodes[0], g, "TCP", "UDP")
	g.Expect(*nodes[0].AviVsNodeCommonFields.ApplicationProfileRef).To(gomega.ContainSubstring("thisisaviref-l4-appprofile"))

	TearDownTestForSharedVIPSvcLB(t, g)
}

func TestCreateDeleteL4RuleSSLCustomValues(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// setting license to enterprise
	SetupLicense(lib.LicenseTypeEnterprise)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080}

	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	applicationProfileRef := proto.String("thisisaviref-l4-ssl-appprofile")
	networkProfileRef := proto.String("thisisaviref-networkprofile-tcp-proxy")
	sslKeyAndCertificateRefs := []string{"thisisaviref-sslkeyandcertref"}
	sslProfileRef := proto.String("thisisaviref-sslprofileref")
	// Create the L4Rule
	SetupL4RuleSSL(t, L4RuleName, NAMESPACE, ports, applicationProfileRef, networkProfileRef, sslProfileRef, sslKeyAndCertificateRefs...)

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// Apply the  L4Rule to Service
	svcObj := (FakeService{
		Name:         svcName,
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

	expectedDefaultPoolName := fmt.Sprintf("cluster--%s-%s-%s-%d", NAMESPACE, svcName, "TCP", ports[0])
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(0)) &&
			g.Expect(nodes[0].DefaultPool).To(gomega.Equal(expectedDefaultPoolName))
	}, 30*time.Second).Should(gomega.Equal(true))

	l4Rule := FakeL4Rule{
		Name:       L4RuleName,
		Namespace:  NAMESPACE,
		Ports:      ports,
		SSLEnabled: true,
	}
	obj := l4Rule.L4Rule()
	convertL4RuleToSSL(obj, ports, applicationProfileRef, networkProfileRef, sslProfileRef, sslKeyAndCertificateRefs...)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)

	// setting license back to basic
	SetupLicense("BASIC")
	ResetMiddleware()
}

func TestCreateDeleteL4RuleSSLDefaultValues(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// setting license to enterprise
	SetupLicense(lib.LicenseTypeEnterprise)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080}

	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// Create the L4Rule
	SetupL4RuleSSL(t, L4RuleName, NAMESPACE, ports, nil, nil, nil)

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// Apply the  L4Rule to Service
	svcObj := (FakeService{
		Name:         svcName,
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

	expectedDefaultPoolName := fmt.Sprintf("cluster--%s-%s-%s-%d", NAMESPACE, svcName, "TCP", ports[0])
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(0)) &&
			g.Expect(nodes[0].DefaultPool).To(gomega.Equal(expectedDefaultPoolName))
	}, 30*time.Second).Should(gomega.Equal(true))

	l4Rule := FakeL4Rule{
		Name:       L4RuleName,
		Namespace:  NAMESPACE,
		Ports:      ports,
		SSLEnabled: true,
	}
	obj := l4Rule.L4Rule()
	convertL4RuleToSSL(obj, ports, nil, nil, nil)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)

	// setting license back to basic
	SetupLicense("BASIC")
	ResetMiddleware()
}

func TestL4RuleSSLCustomValuesLicenseCloudServices(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// setting license to enterprise with cloud services
	SetupLicense(lib.LicenseTypeEnterpriseCloudServices)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080}

	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	applicationProfileRef := proto.String("thisisaviref-l4-ssl-appprofile")
	networkProfileRef := proto.String("thisisaviref-networkprofile-tcp-proxy")
	sslKeyAndCertificateRefs := []string{"thisisaviref-sslkeyandcertref"}
	sslProfileRef := proto.String("thisisaviref-sslprofileref")
	// Create the L4Rule
	SetupL4RuleSSL(t, L4RuleName, NAMESPACE, ports, applicationProfileRef, networkProfileRef, sslProfileRef, sslKeyAndCertificateRefs...)

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// Apply the  L4Rule to Service
	svcObj := (FakeService{
		Name:         svcName,
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

	expectedDefaultPoolName := fmt.Sprintf("cluster--%s-%s-%s-%d", NAMESPACE, svcName, "TCP", ports[0])
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(0)) &&
			g.Expect(nodes[0].DefaultPool).To(gomega.Equal(expectedDefaultPoolName))
	}, 30*time.Second).Should(gomega.Equal(true))

	l4Rule := FakeL4Rule{
		Name:       L4RuleName,
		Namespace:  NAMESPACE,
		Ports:      ports,
		SSLEnabled: true,
	}
	obj := l4Rule.L4Rule()
	convertL4RuleToSSL(obj, ports, applicationProfileRef, networkProfileRef, sslProfileRef, sslKeyAndCertificateRefs...)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)

	// setting license back to basic
	SetupLicense("BASIC")
	ResetMiddleware()
}

func TestL4RuleSSLDefaultValuesLicenseCloudServices(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// setting license to enterprise with cloud services
	SetupLicense(lib.LicenseTypeEnterpriseCloudServices)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080}

	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// Create the L4Rule
	SetupL4RuleSSL(t, L4RuleName, NAMESPACE, ports, nil, nil, nil)

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// Apply the  L4Rule to Service
	svcObj := (FakeService{
		Name:         svcName,
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

	expectedDefaultPoolName := fmt.Sprintf("cluster--%s-%s-%s-%d", NAMESPACE, svcName, "TCP", ports[0])
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].L4PolicyRefs).To(gomega.HaveLen(0)) &&
			g.Expect(nodes[0].DefaultPool).To(gomega.Equal(expectedDefaultPoolName))
	}, 30*time.Second).Should(gomega.Equal(true))

	l4Rule := FakeL4Rule{
		Name:       L4RuleName,
		Namespace:  NAMESPACE,
		Ports:      ports,
		SSLEnabled: true,
	}
	obj := l4Rule.L4Rule()
	convertL4RuleToSSL(obj, ports, nil, nil, nil)

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)

	// setting license back to basic
	SetupLicense("BASIC")
	ResetMiddleware()
}

func TestCreateDeleteL4RuleSSLWrongAppProfile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080}

	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// setting appProfileRef to some unknown non ssl app profile
	appProfileRefIncorrect := proto.String("thisisaviref-L4-custom-profile")
	// Create the L4Rule
	SetupL4RuleSSL(t, L4RuleName, NAMESPACE, ports, appProfileRefIncorrect, nil, nil)

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Rejected"))

	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestCreateDeleteL4RuleNoSSLInSvc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080}

	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// setting appProfileRef to valid ssl app profile
	appProfileRef := proto.String("thisisaviref-l4-ssl-appprofile")

	// Create the L4Rule
	l4Rule := FakeL4Rule{
		Name:       L4RuleName,
		Namespace:  NAMESPACE,
		Ports:      ports,
		SSLEnabled: false,
	}
	obj := l4Rule.L4Rule()
	convertL4RuleToSSL(obj, []int{}, appProfileRef, nil, nil)
	if _, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Create(context.TODO(), obj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding L4Rule: %v", err)
	}

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Rejected"))

	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestCreateDeleteL4RuleSSLProfileWithWrongAppProfile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// this test case can also be written for ssk key and cert
	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080}

	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// setting appProfileRef to some unknown non ssl app profile
	appProfileRefIncorrect := proto.String("thisisaviref-L4-custom-profile")
	sslProfileRef := proto.String("thisisaviref-sslprofileref")
	// Create the L4Rule
	SetupL4RuleSSL(t, L4RuleName, NAMESPACE, ports, appProfileRefIncorrect, nil, sslProfileRef)

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Rejected"))

	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestCreateDeleteL4RuleSSLWrongNetworkProfile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080}

	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// setting networkProfileRef to some unknown non tcp network profile
	networkProfileRefIncorrect := proto.String("thisisaviref-network-custom-profile")
	// Create the L4Rule
	SetupL4RuleSSL(t, L4RuleName, NAMESPACE, ports, networkProfileRefIncorrect, nil, nil)

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Rejected"))

	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}

func TestCreateDeleteL4RuleInOtherNS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	L4RuleName := objNameMap.GenerateName("test-l4rule")
	svcName := objNameMap.GenerateName(SINGLEPORTSVC)
	modelName := MODEL_REDNS_PREFIX + svcName
	ports := []int{8080}

	SetUpTestForSvcLB(t, svcName)

	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, svcName)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(1))
	address := "1.1.1.1"
	g.Expect(nodes[0].PoolRefs[0].Servers[0].Ip.Addr).To(gomega.Equal(&address))

	// Create the L4Rule
	SetupL4Rule(t, L4RuleName, "default", ports)

	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules("default").Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// Apply the  L4Rule to Service
	svcObj := (FakeService{
		Name:         svcName,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	l4Ruleannotation := fmt.Sprintf("default/%s", L4RuleName)
	svcObj.Annotations = map[string]string{lib.L4RuleAnnotation: l4Ruleannotation}
	svcObj.ResourceVersion = "2"
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
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
		Namespace: "default",
		Ports:     ports,
	}
	obj := l4Rule.L4Rule()

	_, aviModel = objects.SharedAviGraphLister().Get(modelName)
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
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes[0].AviVsNodeCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).To(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSvcLB(t, g, svcName)
	TeardownL4Rule(t, L4RuleName, "default")
}

// validates L4Rule RevokeVipRoute's behaviour w.r.t. ako.vmware.com/enable-shared-vip
// 1. checks that L4Rule is accepted with revokeviproute present if cloud is NSX-T
// 2. checks that revokeviproute and enable-shared-vip combination results in
// revokeviproute being omitted out from vsnode.
// 3. checks that on removing enable-shared-vip annotation from svc1 and sv2 results in
// two different vsnodes getting created with each having revokeviproute set back
// in vsnode with its definied value (true) in l4rule.
func TestSharedVIPSvcWithRevokeVipRoute(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// L4Rule with revokeviproute field is only supported in nsx-t
	lib.SetCloudType(lib.CLOUD_NSXT)

	L4RuleName := objNameMap.GenerateName("test-rvr-l4rule")
	ports := []int{8080}
	modelName := MODEL_REDNS_PREFIX + SHAREDVIPKEY

	SetUpTestForSharedVIPSvcLB(t, corev1.ProtocolTCP, corev1.ProtocolTCP)
	// initial validation
	g.Eventually(func() bool {
		found, _ := objects.SharedAviGraphLister().Get(modelName)
		return found
	}, 30*time.Second).Should(gomega.Equal(true))
	_, aviModel := objects.SharedAviGraphLister().Get(modelName)
	nodes := aviModel.(*avinodes.AviObjectGraph).GetAviVS()
	g.Expect(nodes).To(gomega.HaveLen(1))
	g.Expect(nodes[0].Name).To(gomega.Equal(fmt.Sprintf("cluster--%s-%s", NAMESPACE, SHAREDVIPKEY)))
	g.Expect(nodes[0].Tenant).To(gomega.Equal(AVINAMESPACE))
	g.Expect(nodes[0].PortProto[0].Port).To(gomega.Equal(int32(8080)))

	// Check for the pools
	g.Expect(nodes[0].PoolRefs).To(gomega.HaveLen(2))
	addresses := map[string]struct{}{
		"1.1.1.1": {},
		"2.1.1.1": {},
	}
	for _, poolRef := range nodes[0].PoolRefs {
		ipAddr := poolRef.Servers[0].Ip.Addr
		delete(addresses, *ipAddr)
	}
	g.Expect(addresses).To(gomega.HaveLen(0))

	// Create the L4Rule with "RevokeVipRoute"
	l4Rule := FakeL4Rule{
		Name:      L4RuleName,
		Namespace: NAMESPACE,
		Ports:     ports,
	}
	obj := l4Rule.L4Rule()
	obj.Spec.RevokeVipRoute = proto.Bool(true)
	if _, err := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Create(context.TODO(), obj, metav1.CreateOptions{}); err != nil {
		t.Fatalf("error in adding L4Rule: %v", err)
	}
	// validate if L4Rule is accepted
	g.Eventually(func() string {
		l4Rule, _ := lib.AKOControlConfig().V1alpha2CRDClientset().AkoV1alpha2().L4Rules(NAMESPACE).Get(context.TODO(), L4RuleName, metav1.GetOptions{})
		return l4Rule.Status.Status
	}, 30*time.Second).Should(gomega.Equal("Accepted"))

	// Apply the L4Rule to first Service
	svcObj01 := (FakeService{
		Name:         SHAREDVIPSVC01,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcObj01.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName, lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	svcObj01.ResourceVersion = "2"
	_, err := KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj01, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	// Apply the L4Rule to second Service
	svcObj02 := (FakeService{
		Name:         SHAREDVIPSVC02,
		Namespace:    NAMESPACE,
		Type:         corev1.ServiceTypeLoadBalancer,
		ServicePorts: []Serviceport{{PortName: "foo1", Protocol: "TCP", PortNumber: 8080, TargetPort: intstr.FromInt(8080)}},
	}).Service()
	svcObj02.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName, lib.SharedVipSvcLBAnnotation: SHAREDVIPKEY}
	svcObj02.ResourceVersion = "2"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj02, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	// validate that revokeviproute is indeed not applied to the vsnode.
	g.Eventually(func() bool {
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()

		return g.Expect(nodes).To(gomega.HaveLen(1)) &&
			g.Expect(nodes[0].AviVsNodeCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].RevokeVipRoute).To(gomega.BeZero())
	}, 30*time.Second).Should(gomega.Equal(true))

	// *no shared vip* case
	// remove "enable-shared-vip" annotation from both the services
	svcObj01.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName}
	svcObj01.ResourceVersion = "3"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj01, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	svcObj02.Annotations = map[string]string{lib.L4RuleAnnotation: L4RuleName}
	svcObj02.ResourceVersion = "3"
	_, err = KubeClient.CoreV1().Services(NAMESPACE).Update(context.TODO(), svcObj02, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error in updating Service: %v", err)
	}

	// validate both that two nodes are created, with different modelnames for each
	// rather than 1 with shared vip (model name changes once shared vip annotation is removed)
	// also validate RevokeVipRoute is applied as per L4Rule

	// check for shared-vip-svc-01
	g.Eventually(func() bool {
		modelNameSvc1 := MODEL_REDNS_PREFIX + SHAREDVIPSVC01
		found, aviModel := objects.SharedAviGraphLister().Get(modelNameSvc1)
		if !found || aviModel == nil {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes).To(gomega.HaveLen(1)) &&
			g.Expect(nodes[0].AviVsNodeCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].RevokeVipRoute).ToNot(gomega.BeNil()) &&
			g.Expect(*nodes[0].RevokeVipRoute).To(gomega.BeTrue())
	}, 30*time.Second).Should(gomega.Equal(true))

	// check for shared-vip-svc-02
	g.Eventually(func() bool {
		modelNameSvc2 := MODEL_REDNS_PREFIX + SHAREDVIPSVC02
		found, aviModel := objects.SharedAviGraphLister().Get(modelNameSvc2)
		if !found || aviModel == nil {
			return false
		}
		nodes = aviModel.(*avinodes.AviObjectGraph).GetAviVS()
		return g.Expect(nodes).To(gomega.HaveLen(1)) &&
			g.Expect(nodes[0].AviVsNodeCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].AviVsNodeGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolCommonFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].PoolRefs[0].AviPoolGeneratedFields).NotTo(gomega.BeZero()) &&
			g.Expect(nodes[0].RevokeVipRoute).ToNot(gomega.BeNil()) &&
			g.Expect(*nodes[0].RevokeVipRoute).To(gomega.BeTrue())
	}, 30*time.Second).Should(gomega.Equal(true))

	TearDownTestForSharedVIPSvcLB(t, g)
	TeardownL4Rule(t, L4RuleName, NAMESPACE)
}
