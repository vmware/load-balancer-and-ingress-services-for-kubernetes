/*
 * [2013] - [2020] Avi Networks Incorporated
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
	"fmt"
	"net/http"

	"github.com/avinetworks/container-lib/api/models"
	"github.com/avinetworks/container-lib/utils"
	"github.com/gorilla/mux"
)

type ApiServer struct {
	Port   string
	Models []models.ApiModel
}

func (a *ApiServer) SetRouter() *mux.Router {
	router := mux.NewRouter()
	routerMap := make(map[string]bool)

	for _, model := range a.Models {
		opermaps := model.ApiOperationMap()
		for _, o := range opermaps {
			routerMapKey := fmt.Sprintf("%s:%s", o.Method, o.Route)

			if _, ok := routerMap[routerMapKey]; !ok {
				routerMap[routerMapKey] = true
				utils.AviLog.Infof("Setting route for %s %s", o.Method, o.Route)
				router.HandleFunc(o.Route, o.Handler).Methods(o.Method)
			} else {
				utils.AviLog.Warnf("Route for %s %s already exists", o.Method, o.Route)
			}
		}
	}

	router.Use(utils.LogApi)
	return router
}

func (a *ApiServer) InitApi() {
	// add common models in ApiServer
	genericModels := []models.ApiModel{
		&models.RestStatus,
	}
	a.Models = append(a.Models, genericModels...)

	// initialize all models
	for _, model := range a.Models {
		model.InitModel()
	}

	router := a.SetRouter()
	port := a.Port

	utils.AviLog.Infof("API server now running on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		utils.AviLog.Errorf("Error initializing AKO api server: %+v", err)
		return
	}
}
