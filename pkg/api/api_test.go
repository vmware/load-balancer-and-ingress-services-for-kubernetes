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

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api/models"
)

func TestMain(m *testing.M) {
	akoApi := &ApiServer{
		Port:   "12345",
		Models: []models.ApiModel{},
	}

	go akoApi.InitApi()

	os.Exit(m.Run())
}

// TestApiServerStatusModel tests InitApi and the StatusModel feature
func TestApiServerStatusModel(t *testing.T) {
	resp, err := http.Get("http://localhost:12345/api/status")
	if err != nil {
		t.Fail()
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fail()
	}

	var status models.StatusModel
	if err = json.Unmarshal(body, &status); err != nil {
		t.Fail()
	}

	if status.AviApi.ConnectionStatus != "INITIATING" {
		t.Fail()
	}
}
