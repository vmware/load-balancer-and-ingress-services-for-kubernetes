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

package utils

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"
)

type AviRestClientPool struct {
	AviClient []*clients.AviClient
}

var AviClientInstance *AviRestClientPool

func NewAviRestClientPool(num uint32, api_ep, username,
	password, authToken, controllerVersion, ctrlCAData string) (*AviRestClientPool, string, error) {
	var clientPool AviRestClientPool
	var wg sync.WaitGroup
	var globalErr error

	rootPEMCerts := ctrlCAData
	var transport *http.Transport
	if rootPEMCerts != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(rootPEMCerts))

		transport =
			&http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: caCertPool,
				},
			}
	}

	clientPool.AviClient = make([]*clients.AviClient, num)
	for i := uint32(0); i < num; i++ {
		wg.Add(1)
		go func(i uint32) {
			defer wg.Done()
			if globalErr != nil {
				return
			}

			var aviClient *clients.AviClient
			var err error

			if authToken == "" {
				if rootPEMCerts != "" {
					aviClient, err = clients.NewAviClient(api_ep, username,
						session.SetPassword(password), session.SetNoControllerStatusCheck, session.SetTransport(transport))
				} else {
					aviClient, err = clients.NewAviClient(api_ep, username,
						session.SetPassword(password), session.SetNoControllerStatusCheck, session.SetTransport(transport), session.SetInsecure)
				}
			} else {
				if rootPEMCerts != "" {
					aviClient, err = clients.NewAviClient(api_ep, username,
						session.SetAuthToken(authToken), session.SetRefreshAuthTokenCallbackV2(GetAuthtokenFromCache), session.SetNoControllerStatusCheck, session.SetTransport(transport))
				} else {
					aviClient, err = clients.NewAviClient(api_ep, username,
						session.SetAuthToken(authToken), session.SetRefreshAuthTokenCallbackV2(GetAuthtokenFromCache), session.SetNoControllerStatusCheck, session.SetTransport(transport), session.SetInsecure)
				}
			}
			if err != nil {
				AviLog.Warnf("NewAviClient returned err %v", err)
				globalErr = err
				return
			}
			clientPool.AviClient[i] = aviClient
		}(i)
	}

	wg.Wait()

	if globalErr != nil {
		return &clientPool, controllerVersion, globalErr
	}

	// Get the controller version if it is not present in env variable.
	if controllerVersion == "" {
		version, err := clientPool.AviClient[0].AviSession.GetControllerVersion()
		if err != nil {
			return &clientPool, controllerVersion, err
		}
		AviLog.Infof("Setting the client version to the current controller version %v", version)
		CtrlVersion = version
		// For AKO-1.9.1, if controller version is 22.1.3, set api version as 22.1.2
		if CtrlVersion == CTRL_VERSION_22_1_3 {
			CtrlVersion = CTRL_VERSION_22_1_2
		}
		AviLog.Infof("Setting the client version to the current controller version %v", version)
		controllerVersion = version
	}

	return &clientPool, controllerVersion, nil
}

func (p *AviRestClientPool) AviRestOperate(c *clients.AviClient, rest_ops []*RestOp) error {
	for i, op := range rest_ops {
		SetTenant := session.SetTenant(op.Tenant)
		SetTenant(c.AviSession)
		SetVersion := session.SetVersion(op.Version)
		SetVersion(c.AviSession)
		switch op.Method {
		case RestPost:
			op.Err = c.AviSession.Post(op.Path, op.Obj, &op.Response)
		case RestPut:
			op.Err = c.AviSession.Put(op.Path, op.Obj, &op.Response)
		case RestGet:
			op.Err = c.AviSession.Get(op.Path, &op.Response)
		case RestPatch:
			op.Err = c.AviSession.Patch(op.Path, op.Obj, op.PatchOp,
				&op.Response)
		case RestDelete:
			op.Err = c.AviSession.Delete(op.Path)
		default:
			AviLog.Errorf("Unknown RestOp %v", op.Method)
			op.Err = fmt.Errorf("Unknown RestOp %v", op.Method)
		}
		if op.Err != nil {
			AviLog.Warnf(`RestOp method %v path %v tenant %v Obj %s returned err %s with response %s`,
				op.Method, op.Path, op.Tenant, Stringify(op.Obj), Stringify(op.Err), Stringify(op.Response))
			for j := i + 1; j < len(rest_ops); j++ {
				rest_ops[j].Err = errors.New("Aborted due to prev error")
			}
			// Wrap the error into a websync error.
			err := &WebSyncError{Err: op.Err, Operation: string(op.Method)}
			return err
		} else {
			AviLog.Debugf(`RestOp method %v path %v tenant %v response %v`,
				op.Method, op.Path, op.Tenant, Stringify(op.Response))
		}
	}
	return nil
}

func AviModelToUrl(model string) string {
	switch model {
	case "Pool":
		return "/api/pool"
	case "VirtualService":
		return "/api/virtualservice"
	case "PoolGroup":
		return "/api/poolgroup"
	case "SSLKeyAndCertificate":
		return "/api/sslkeyandcertificate"
	case "HTTPPolicySet":
		return "/api/httppolicyset"
	case "GSLBService":
		return "/api/gslbservice"
	case "VsVip":
		return "/api/vsvip"
	case "VSDataScriptSet":
		return "/api/vsdatascriptset"
	default:
		AviLog.Warnf("Unknown model %v", model)
		return ""
	}
}

func GetAuthTokenWithRetry(c *clients.AviClient, retryCount int) (interface{}, error) {
	tokenPath := "api/user-token"
	var robj interface{}
	var err error
	for retry := 0; retry < retryCount; retry++ {
		err = c.AviSession.Get(tokenPath, &robj)
		if err == nil {
			return robj, nil
		}
		AviLog.Warnf("Failed to get authtoken, retry count:%d, err: %+v", retry, err)
	}
	return robj, err
}

func CreateAuthTokenWithRetry(c *clients.AviClient, retryCount int) (interface{}, error) {
	tokenPath := "api/user-token"
	var robj interface{}
	var err error
	data := make(map[string]string)
	data["hours"] = strconv.Itoa(AuthTokenExpiry)
	for retry := 0; retry < retryCount; retry++ {
		err = c.AviSession.Post(tokenPath, data, &robj)
		if err == nil {
			return robj, nil
		}
		AviLog.Warnf("Failed to create authtoken, retry count:%d, err: %+v", retry, err)
	}
	return robj, err
}

func DeleteAuthTokenWithRetry(c *clients.AviClient, tokenID string, retryCount int) error {
	tokenPath := "api/user-token"
	var err error
	for retry := 0; retry < retryCount; retry++ {
		err = c.AviSession.Delete(tokenPath + "/" + tokenID)
		if err == nil {
			return nil
		}
		AviLog.Warnf("Failed to delete authtoken, retry count:%d, err: %+v", retry, err)
	}
	return err
}
