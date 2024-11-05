/*
 * Copyright 2023-2024 VMware, Inc.
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

package dedicatedvstests

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega"

	avinodes "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/integrationtest"
)

func TestProfilesAttachedToDedicatedSecureVS(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// middleware verifies the application and network profiles attached to the VS
	integrationtest.AddMiddleware(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.EscapedPath()
		if r.Method == http.MethodPost &&
			strings.Contains(url, "/api/virtualservice") {
			var resp map[string]interface{}
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &resp)
			g.Expect(resp["application_profile_ref"]).Should(gomega.HaveSuffix("System-Secure-HTTP"))
			g.Expect(resp["network_profile_ref"]).Should(gomega.HaveSuffix("System-TCP-Proxy"))
			return
		}
		integrationtest.NormalControllerServer(w, r)
	})

	modelName := "admin/cluster--foo.com-L7-dedicated"
	secretName := objNameMap.GenerateName("my-secret")
	ingressName := objNameMap.GenerateName("foo-with-targets")
	svcName := objNameMap.GenerateName("avisvc")
	ingTestObj := IngressTestObject{
		ingressName: ingressName,
		isTLS:       true,
		withSecret:  true,
		secretName:  secretName,
		serviceName: svcName,
		modelNames:  []string{modelName},
	}
	ingTestObj.FillParams()
	SetUpIngressForCacheSyncCheck(t, ingTestObj)
	g.Eventually(func() int {
		_, aviModel := objects.SharedAviGraphLister().Get(modelName)
		nodes, ok := aviModel.(*avinodes.AviObjectGraph)
		if !ok {
			return 0
		}
		return len(nodes.GetAviVS())
	}, 30*time.Second).Should(gomega.Equal(1))

	TearDownIngressForCacheSyncCheck(t, secretName, ingressName, modelName)

	integrationtest.ResetMiddleware()

}
