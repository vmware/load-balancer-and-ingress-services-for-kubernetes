// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

import (
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"
)

// AviClient -- an API Client for Avi Controller
type AviClient struct {
	AviSession *session.AviSession
}

// NewAviClient initiates an AviSession and returns an AviClient wrapping that session
func NewAviClient(host string, username string, options ...func(*session.AviSession) error) (*AviClient, error) {
	aviClient := AviClient{}
	aviSession, err := session.NewAviSession(host, username, options...)
	if err != nil {
		return &aviClient, err
	}
	aviClient.AviSession = aviSession
	return &aviClient, nil
}
