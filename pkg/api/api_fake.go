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
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api/models"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type FakeApiServer struct {
	Models   []models.ApiModel
	Port     string
	Shutdown bool
}

func (a *FakeApiServer) initModels() {
	// add common models in ApiServer
	genericModels := []models.ApiModel{
		models.RestStatus,
	}
	a.Models = append(a.Models, genericModels...)

	// initialize all models
	for _, model := range a.Models {
		model.InitModel()
	}
}

func (a *FakeApiServer) InitApi() {
	a.initModels()
	utils.AviLog.Infof("Fake API server now running on port %s", a.Port)
	return
}

func (a *FakeApiServer) SetRouter(prometheusEnavbled bool, reg *prometheus.Registry) *mux.Router {
	return nil
}

func (a *FakeApiServer) ShutDown() {
	a.Shutdown = true
}
