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

package hatests

import (
	"context"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateHostRule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	hrname := "samplehr-foo"
	integrationtest.SetupHostRule(t, hrname, "foo.com", true)

	// AKO is running as follower, hence the status won't be updated.
	g.Eventually(func() string {
		hostrule, _ := v1beta1CRDClient.AkoV1beta1().HostRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return hostrule.Status.Status
	}, 30*time.Second).Should(gomega.Equal(""))
}

func TestCreateHTTPRule(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	hrname := "samplehr-foo"
	integrationtest.SetupHTTPRule(t, hrname, "foo.com", "/foo")

	// AKO is running as follower, hence the status won't be updated.
	g.Eventually(func() string {
		httprule, _ := v1beta1CRDClient.AkoV1beta1().HTTPRules("default").Get(context.TODO(), hrname, metav1.GetOptions{})
		return httprule.Status.Status
	}, 30*time.Second).Should(gomega.Equal(""))
}

func TestCreateAviInfraSetting(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	aviInfraSettingName := "sample-aviinfra"
	integrationtest.SetupAviInfraSetting(t, aviInfraSettingName, "LARGE")

	// AKO is running as follower, hence the status won't be updated.
	g.Eventually(func() string {
		aviInfraSetting, _ := v1beta1CRDClient.AkoV1beta1().AviInfraSettings().Get(context.TODO(), aviInfraSettingName, metav1.GetOptions{})
		return aviInfraSetting.Status.Status
	}, 30*time.Second).Should(gomega.Equal(""))
}
