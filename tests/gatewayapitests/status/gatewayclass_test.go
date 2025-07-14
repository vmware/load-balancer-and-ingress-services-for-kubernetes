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

package status

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapitests "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/gatewayapitests"
)

func TestGatewayClassValidation(t *testing.T) {

	gatewayClassName := "gateway-class-01"
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, akogatewayapilib.GatewayController)

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gatewayClass, err := akogatewayapitests.GatewayClient.GatewayV1().GatewayClasses().Get(context.TODO(), gatewayClassName, metav1.GetOptions{})
		if err != nil || gatewayClass == nil {
			t.Logf("Couldn't get the GatewayClass, err: %+v", err)
			return false
		}
		return apimeta.IsStatusConditionTrue(gatewayClass.Status.Conditions, string(gatewayv1.GatewayClassConditionStatusAccepted))
	}, 30*time.Second).Should(gomega.Equal(true))

	expectedStatus := &gatewayv1.GatewayClassStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayClassConditionStatusAccepted),
				Status:             metav1.ConditionTrue,
				Message:            "GatewayClass is valid",
				ObservedGeneration: 1,
				Reason:             string(gatewayv1.GatewayClassReasonAccepted),
			},
		},
	}

	gatewayClass, err := akogatewayapitests.GatewayClient.GatewayV1().GatewayClasses().Get(context.TODO(), gatewayClassName, metav1.GetOptions{})
	if err != nil || gatewayClass == nil {
		t.Fatalf("Couldn't get the GatewayClass, err: %+v", err)
	}

	akogatewayapitests.ValidateConditions(t, gatewayClass.Status.Conditions, expectedStatus.Conditions)
	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}

func TestGatewayClassWithNonAKOController(t *testing.T) {

	gatewayClassName := "gateway-class-01"
	akogatewayapitests.SetupGatewayClass(t, gatewayClassName, "other.com/non-ako-controller")

	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		gatewayClass, err := akogatewayapitests.GatewayClient.GatewayV1().GatewayClasses().Get(context.TODO(), gatewayClassName, metav1.GetOptions{})
		if err != nil || gatewayClass == nil {
			t.Logf("Couldn't get the GatewayClass, err: %+v", err)
			return false
		}
		return apimeta.FindStatusCondition(gatewayClass.Status.Conditions, string(gatewayv1.GatewayClassConditionStatusAccepted)) == nil
	}, 30*time.Second).Should(gomega.Equal(true))

	akogatewayapitests.TeardownGatewayClass(t, gatewayClassName)
}
