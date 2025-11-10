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

package avirest

import (
	"fmt"
	"sync"
	"time"

	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/session"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

const (
	// VKS Management Service APIs were introduced in 32.1.1
	VKSAviVersion = "32.1.1"
)

var infraAviClientInstance *clients.AviClient
var ctrlClientOnce sync.Once

var vksAviClientInstance *clients.AviClient

func InfraAviClientInstance(c ...*clients.AviClient) *clients.AviClient {
	if len(c) > 0 {
		ctrlClientOnce.Do(func() {
			infraAviClientInstance = c[0]
		})
	}
	return infraAviClientInstance
}

// VKSAviClientInstance returns a VKS-specific AVI client with API version
// This client is specifically for VKS Management Service and Grant APIs
func VKSAviClientInstance(c ...*clients.AviClient) *clients.AviClient {
	if len(c) > 0 && vksAviClientInstance == nil {
		vksAviClientInstance = c[0]
	}
	return vksAviClientInstance
}

// IsVKSAviClientAvailable returns true if VKS AVI client has been successfully initialized
// This indicates that the controller version supports VKS Management Service APIs
func IsVKSAviClientAvailable() bool {
	return vksAviClientInstance != nil
}

// CreateVKSAviClient creates a new AVI client specifically for VKS operations
// with API version set to support Management Service APIs
func CreateVKSAviClient(controllerIP, username, authToken, caData string) (*clients.AviClient, error) {
	if controllerIP == "" {
		return nil, fmt.Errorf("VKS: Controller IP not available for VKS client initialization")
	}

	if username == "" || authToken == "" {
		return nil, fmt.Errorf("VKS: Controller credentials not available for VKS client initialization")
	}

	// Set X-Avi-UserAgent header for rate limiting identification
	userHeaders := utils.SharedCtrlProp().GetCtrlUserHeader()
	userHeaders[utils.XAviUserAgentHeader] = "AKO"

	transport, isSecure := utils.GetHTTPTransportWithCert(caData)
	options := []func(*session.AviSession) error{
		session.SetAuthToken(authToken),
		session.DisableControllerStatusCheckOnFailure(true),
		session.SetTransport(transport),
		session.SetTimeout(120 * time.Second),
		session.SetVersion(VKSAviVersion),
		session.SetUserHeader(userHeaders),
	}

	if !isSecure {
		options = append(options, session.SetInsecure)
	}

	aviClient, err := clients.NewAviClient(controllerIP, username, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create VKS AVI client: %v", err)
	}

	return aviClient, nil
}
