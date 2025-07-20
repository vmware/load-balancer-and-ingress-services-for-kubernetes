// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

import (
	"encoding/json"
	"fmt"
	"github.com/vmware/alb-sdk/go/session"
	"net/http"
)

// VirtualServiceClient is a client for avi VirtualService resource
type GenericClient struct {
	aviSession *session.AviSession
}

func NewGenericClient(aviSession *session.AviSession) *GenericClient {
	return &GenericClient{aviSession: aviSession}
}

// NewVirtualServiceClient creates a new client for VirtualService resource
func (client *GenericClient) AviResponse(payloadJSON string, url string, method string) *http.Response {
	var payload interface{}
	var err interface{}
	if err = json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		fmt.Println("Unable to decode payload: %v", err)
	}
	resp, err := client.aviSession.RestRequest(method, url, payload, "admin", nil)
	aviError, _ := err.(session.AviError)
	if err != nil {
		fmt.Println("Request error: %v %v", resp, aviError.Error())
	}
	fmt.Println("Response: %v ", resp)
	return resp
}
