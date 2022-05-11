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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"os"
	"strings"

	apimodels "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api/models"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"

	corev1 "k8s.io/api/core/v1"
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
		utils.AviLog.Warnf("msg: Unable to fetch collection data from uri %s %v", uri, err)
		checkForInvalidCredentials(uri, err)
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			return session.AviCollectionResult{}, err
		}
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
		utils.AviLog.Warnf("msg: Unable to fetch data from uri %s %v", uri, err)
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			utils.AviLog.Debugf("Switching to admin context from %s", GetTenant())
			SetAdminTenant := session.SetTenant(GetAdminTenant())
			SetTenant := session.SetTenant(GetTenant())
			SetAdminTenant(client.AviSession)
			defer SetTenant(client.AviSession)
			if err := AviGet(client, uri, response); err != nil {
				utils.AviLog.Warnf("msg: Unable to fetch data from uri %s %v after context switch", uri, err)
				return err
			}
		}
		checkForInvalidCredentials(uri, err)
		apimodels.RestStatus.UpdateAviApiRestStatus("", err)
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			return err
		}
		return AviGet(client, uri, response, retry+1)
	}

	apimodels.RestStatus.UpdateAviApiRestStatus(utils.AVIAPI_CONNECTED, nil)
	return nil
}

func AviGetRaw(client *clients.AviClient, uri string, retryNum ...int) ([]byte, error) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			err := errors.New("msg: AviGetRaw retried 3 times, aborting")
			return nil, err
		}
	}

	rawData, err := client.AviSession.GetRaw(uri)
	if err != nil {
		utils.AviLog.Warnf("msg: Unable to fetch data from uri %s %v", uri, err)
		checkForInvalidCredentials(uri, err)
		apimodels.RestStatus.UpdateAviApiRestStatus("", err)
		if aviError, ok := err.(session.AviError); ok {
			if aviError.HttpStatusCode == 403 ||
				strings.Contains(aviError.Error(), VSVIPNotFoundError) {
				return nil, err
			}
		}
		return AviGetRaw(client, uri, retry+1)
	}

	apimodels.RestStatus.UpdateAviApiRestStatus(utils.AVIAPI_CONNECTED, nil)
	return rawData, nil
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
		utils.AviLog.Warnf("msg: Unable to execute Put on uri %s %v", uri, err)
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			utils.AviLog.Debugf("Switching to admin context from %s", GetTenant())
			SetAdminTenant := session.SetTenant(GetAdminTenant())
			SetTenant := session.SetTenant(GetTenant())
			SetAdminTenant(client.AviSession)
			defer SetTenant(client.AviSession)
			if err := AviPut(client, uri, payload, response); err != nil {
				utils.AviLog.Warnf("msg: Unable to execute Put on uri %s %v after context switch", uri, err)
				return err
			}
		}
		checkForInvalidCredentials(uri, err)
		apimodels.RestStatus.UpdateAviApiRestStatus("", err)
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 400 {
			return err
		}
		return AviPut(client, uri, payload, response, retry+1)
	}

	apimodels.RestStatus.UpdateAviApiRestStatus(utils.AVIAPI_CONNECTED, nil)
	return nil
}

func AviPost(client *clients.AviClient, uri string, payload interface{}, response interface{}, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			err := errors.New("msg: AviPost retried 3 times, aborting")
			return err
		}
	}

	err := client.AviSession.Post(uri, payload, &response)
	if err != nil {
		utils.AviLog.Warnf("msg: Unable to execute Post on uri %s %v", uri, err)
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			utils.AviLog.Debugf("Switching to admin context from %s", GetTenant())
			SetAdminTenant := session.SetTenant(GetAdminTenant())
			SetTenant := session.SetTenant(GetTenant())
			SetAdminTenant(client.AviSession)
			defer SetTenant(client.AviSession)
			if err := AviPost(client, uri, payload, response); err != nil {
				utils.AviLog.Warnf("msg: Unable to execute Post on uri %s %v after context switch", uri, err)
				return err
			}
		}
		checkForInvalidCredentials(uri, err)
		apimodels.RestStatus.UpdateAviApiRestStatus("", err)
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			return err
		}
		return AviPost(client, uri, payload, response, retry+1)
	}

	apimodels.RestStatus.UpdateAviApiRestStatus(utils.AVIAPI_CONNECTED, nil)
	return nil
}

func AviDelete(client *clients.AviClient, uri string, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			err := errors.New("msg: AviDelete retried 3 times, aborting")
			return err
		}
	}

	err := client.AviSession.Delete(uri)
	if err != nil {
		utils.AviLog.Warnf("msg: Unable to execute Delete on uri %s %v", uri, err)
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			utils.AviLog.Debugf("Switching to admin context from %s", GetTenant())
			SetAdminTenant := session.SetTenant(GetAdminTenant())
			SetTenant := session.SetTenant(GetTenant())
			SetAdminTenant(client.AviSession)
			defer SetTenant(client.AviSession)
			if err := AviDelete(client, uri); err != nil {
				utils.AviLog.Warnf("msg: Unable to execute Post on uri %s %v after context switch", uri, err)
				return err
			}
		}
		checkForInvalidCredentials(uri, err)
		apimodels.RestStatus.UpdateAviApiRestStatus("", err)
		if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 403 {
			return err
		}
		return AviDelete(client, uri, retry+1)
	}

	apimodels.RestStatus.UpdateAviApiRestStatus(utils.AVIAPI_CONNECTED, nil)
	return nil
}

func checkForInvalidCredentials(uri string, err error) {
	if err == nil {
		return
	}

	if utils.IsVCFCluster() {
		WaitForInitSecretRecreateAndReboot()
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
}

func NewAviRestClientWithToken(api_ep string, username string, authToken string) *clients.AviClient {
	var aviClient *clients.AviClient
	var transport *http.Transport
	var err error

	ctrlIpAddress := GetControllerIP()
	if username == "" || authToken == "" || ctrlIpAddress == "" {
		var authTokenLog string
		if authToken != "" {
			authTokenLog = "<sensitive>"
		}
		AKOControlConfig().PodEventf(
			corev1.EventTypeWarning,
			AKOShutdown, "Avi Controller information missing (username: %s, password: %s, authToken: %s, controller: %s)",
			username, authTokenLog, ctrlIpAddress,
		)
		utils.AviLog.Fatalf("Avi Controller information missing (username: %s, authToken: %s, controller: %s). Update them in avi-secret.", username, authTokenLog, ctrlIpAddress)
	}

	rootPEMCerts := os.Getenv("CTRL_CA_DATA")
	if rootPEMCerts != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(rootPEMCerts))

		transport =
			&http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: caCertPool,
				},
			}
		aviClient, err = clients.NewAviClient(api_ep, username, session.SetAuthToken(authToken), session.SetNoControllerStatusCheck, session.SetTransport(transport))
	} else {
		aviClient, err = clients.NewAviClient(api_ep, username, session.SetAuthToken(authToken), session.SetNoControllerStatusCheck, session.SetTransport(transport), session.SetInsecure)
	}
	if err != nil {
		utils.AviLog.Warnf("NewAviClient returned err %v", err)
		return nil
	}

	controllerVersion := AKOControlConfig().ControllerVersion()
	SetTenant := session.SetTenant(GetTenant())
	SetTenant(aviClient.AviSession)
	SetVersion := session.SetVersion(controllerVersion)
	SetVersion(aviClient.AviSession)
	return aviClient
}
