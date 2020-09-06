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

package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api/models"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/gorilla/mux"
)

type ApiServer struct {
	http.Server
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

func (a *ApiServer) ShutDown() {
	//Create shutdown context with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	utils.AviLog.Infof("Shutting down the API server")
	//shutdown the server
	err := a.Shutdown(ctx)
	if err != nil {
		utils.AviLog.Warnf("Error Shutting down the API server :%s", err)
	}
}

func (a *ApiServer) initModels() {
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

func NewServer(port string, models []models.ApiModel) *ApiServer {

	s := &ApiServer{
		Server: http.Server{
			Addr:         ":" + port,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
	s.Models = models
	s.initModels()
	router := s.SetRouter()

	//set http server handler
	s.Handler = router

	return s
}

func (a *ApiServer) InitApi() {
	go func() {
		utils.AviLog.Infof("Starting API server at %s", a.Server.Addr)
		err := a.ListenAndServe()
		if err != nil {
			utils.AviLog.Infof("API server shutdown: %v", err)
		}
	}()
}

// InitFakeApi only initializes the models, and does not run the server on a port
func (a *ApiServer) InitFakeApi() {
	a.initModels()
	utils.AviLog.Infof("Fake API server now running on port %s", a.Port)
	return
}
