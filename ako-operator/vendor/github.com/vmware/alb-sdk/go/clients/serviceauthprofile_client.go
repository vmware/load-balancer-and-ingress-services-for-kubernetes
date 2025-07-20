// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// ServiceAuthProfileClient is a client for avi ServiceAuthProfile resource
type ServiceAuthProfileClient struct {
	aviSession *session.AviSession
}

// NewServiceAuthProfileClient creates a new client for ServiceAuthProfile resource
func NewServiceAuthProfileClient(aviSession *session.AviSession) *ServiceAuthProfileClient {
	return &ServiceAuthProfileClient{aviSession: aviSession}
}

func (client *ServiceAuthProfileClient) getAPIPath(uuid string) string {
	path := "api/serviceauthprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ServiceAuthProfile objects
func (client *ServiceAuthProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ServiceAuthProfile, error) {
	var plist []*models.ServiceAuthProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ServiceAuthProfile by uuid
func (client *ServiceAuthProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ServiceAuthProfile, error) {
	var obj *models.ServiceAuthProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ServiceAuthProfile by name
func (client *ServiceAuthProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ServiceAuthProfile, error) {
	var obj *models.ServiceAuthProfile
	err := client.aviSession.GetObjectByName("serviceauthprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ServiceAuthProfile by filters like name, cloud, tenant
// Api creates ServiceAuthProfile object with every call.
func (client *ServiceAuthProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.ServiceAuthProfile, error) {
	var obj *models.ServiceAuthProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("serviceauthprofile", newOptions...)
	return obj, err
}

// Create a new ServiceAuthProfile object
func (client *ServiceAuthProfileClient) Create(obj *models.ServiceAuthProfile, options ...session.ApiOptionsParams) (*models.ServiceAuthProfile, error) {
	var robj *models.ServiceAuthProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ServiceAuthProfile object
func (client *ServiceAuthProfileClient) Update(obj *models.ServiceAuthProfile, options ...session.ApiOptionsParams) (*models.ServiceAuthProfile, error) {
	var robj *models.ServiceAuthProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ServiceAuthProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ServiceAuthProfile
// or it should be json compatible of form map[string]interface{}
func (client *ServiceAuthProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ServiceAuthProfile, error) {
	var robj *models.ServiceAuthProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ServiceAuthProfile object with a given UUID
func (client *ServiceAuthProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ServiceAuthProfile object with a given name
func (client *ServiceAuthProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ServiceAuthProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
