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

package lib

import (
	"errors"
	"strings"

	apimodels "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api/models"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/avinetworks/sdk/go/clients"
	"github.com/avinetworks/sdk/go/session"
)

func AviGetCollectionRaw(client *clients.AviClient, uri string, retryNum ...int) (session.AviCollectionResult, error) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			err := errors.New("msg: AviGetCollectionRaw retried 3 times, aborting")
			return session.AviCollectionResult{}, err
		}
	}

	result, err := client.AviSession.GetCollectionRaw(uri)

	if err != nil {
		checkForInvalidCredentials(uri, err)
		apimodels.RestStatus.UpdateAviApiRestStatus("", err)
		return AviGetCollectionRaw(client, uri, retry+1)
	}

	apimodels.RestStatus.UpdateAviApiRestStatus(utils.AVIAPI_CONNECTED, nil)
	return result, nil
}

func AviGet(client *clients.AviClient, uri string, response interface{}, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			err := errors.New("msg: AviGet retried 3 times, aborting")
			return err
		}
	}

	err := client.AviSession.Get(uri, &response)
	if err != nil {
		checkForInvalidCredentials(uri, err)
		apimodels.RestStatus.UpdateAviApiRestStatus("", err)
		return AviGet(client, uri, response, retry+1)
	}

	apimodels.RestStatus.UpdateAviApiRestStatus(utils.AVIAPI_CONNECTED, nil)
	return nil
}

func AviPut(client *clients.AviClient, uri string, payload interface{}, response interface{}, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			err := errors.New("msg: AviPut retried 3 times, aborting")
			return err
		}
	}

	err := client.AviSession.Put(uri, payload, &response)
	if err != nil {
		checkForInvalidCredentials(uri, err)
		apimodels.RestStatus.UpdateAviApiRestStatus("", err)
		return AviPut(client, uri, payload, response, retry+1)
	}

	apimodels.RestStatus.UpdateAviApiRestStatus(utils.AVIAPI_CONNECTED, nil)
	return nil
}

func checkForInvalidCredentials(uri string, err error) {
	if err == nil {
		return
	}

	if webSyncErr, ok := err.(*utils.WebSyncError); ok {
		aviError, ok := webSyncErr.GetWebAPIError().(session.AviError)
		if ok && aviError.HttpStatusCode == 401 {
			if strings.Contains(*aviError.Message, "Invalid credentials") {
				utils.AviLog.Errorf("msg: Invalid credentials error for API request: %s, Shutting down API Server", uri)
				ShutdownApi()
			}
		}
	}
	return
}
