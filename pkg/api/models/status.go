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

package models

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

// AviApiRestStatus holds status details for AKO/AMKO <-> AVI connection
type AviApiRestStatus struct {
	sync.Mutex
	ConnectionStatus string            `json:"connection_status"`
	Errors           []RestStatusError `json:"errors"`
}

type RestStatusError struct {
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

var RestStatus *StatusModel
var reststatusonce sync.Once

// StatusModel implements ApiModel
type StatusModel struct {
	AviApi AviApiRestStatus `json:"avi_api"`
}

func (a *StatusModel) InitModel() {
	reststatusonce.Do(func() {
		RestStatus = &StatusModel{
			AviApi: AviApiRestStatus{
				ConnectionStatus: utils.AVIAPI_INITIATING,
				Errors:           []RestStatusError{},
			},
		}
	})
}

func (a *StatusModel) ApiOperationMap() []OperationMap {
	var operationMapList []OperationMap

	get := OperationMap{
		Route:  "/api/status",
		Method: "GET",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			response := &RestStatus
			utils.Respond(w, response)
		},
	}

	operationMapList = append(operationMapList, get)
	return operationMapList
}

// utility function to be used by modules to update RestStatus.AviApi
func (a *StatusModel) UpdateAviApiRestStatus(connectionStatus string, err error) {
	a.AviApi.Lock()
	defer a.AviApi.Unlock()
	aviApiRestStatus := a.AviApi
	var setConnectionStatus string

	if connectionStatus != "" {
		setConnectionStatus = connectionStatus
	}

	if err != nil {
		if strings.Contains(err.Error(), "Client.Timeout") {
			setConnectionStatus = utils.AVIAPI_DISCONNECTED
		}

		// cyclic slice, shows last 10 errors
		if len(aviApiRestStatus.Errors) == 10 {
			aviApiRestStatus.Errors = aviApiRestStatus.Errors[1:]
		}
		aviApiRestStatus.Errors = append(aviApiRestStatus.Errors, RestStatusError{
			Timestamp: time.Now(),
			Error:     err.Error(),
		})
	}

	if setConnectionStatus != "" {
		aviApiRestStatus.ConnectionStatus = setConnectionStatus
	}

	a.AviApi = aviApiRestStatus

	return
}
