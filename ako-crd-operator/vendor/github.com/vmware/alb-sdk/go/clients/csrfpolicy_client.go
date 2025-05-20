// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// CSRFPolicyClient is a client for avi CSRFPolicy resource
type CSRFPolicyClient struct {
	aviSession *session.AviSession
}

// NewCSRFPolicyClient creates a new client for CSRFPolicy resource
func NewCSRFPolicyClient(aviSession *session.AviSession) *CSRFPolicyClient {
	return &CSRFPolicyClient{aviSession: aviSession}
}

func (client *CSRFPolicyClient) getAPIPath(uuid string) string {
	path := "api/csrfpolicy"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of CSRFPolicy objects
func (client *CSRFPolicyClient) GetAll(options ...session.ApiOptionsParams) ([]*models.CSRFPolicy, error) {
	var plist []*models.CSRFPolicy
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing CSRFPolicy by uuid
func (client *CSRFPolicyClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.CSRFPolicy, error) {
	var obj *models.CSRFPolicy
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing CSRFPolicy by name
func (client *CSRFPolicyClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.CSRFPolicy, error) {
	var obj *models.CSRFPolicy
	err := client.aviSession.GetObjectByName("csrfpolicy", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing CSRFPolicy by filters like name, cloud, tenant
// Api creates CSRFPolicy object with every call.
func (client *CSRFPolicyClient) GetObject(options ...session.ApiOptionsParams) (*models.CSRFPolicy, error) {
	var obj *models.CSRFPolicy
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("csrfpolicy", newOptions...)
	return obj, err
}

// Create a new CSRFPolicy object
func (client *CSRFPolicyClient) Create(obj *models.CSRFPolicy, options ...session.ApiOptionsParams) (*models.CSRFPolicy, error) {
	var robj *models.CSRFPolicy
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing CSRFPolicy object
func (client *CSRFPolicyClient) Update(obj *models.CSRFPolicy, options ...session.ApiOptionsParams) (*models.CSRFPolicy, error) {
	var robj *models.CSRFPolicy
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing CSRFPolicy object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.CSRFPolicy
// or it should be json compatible of form map[string]interface{}
func (client *CSRFPolicyClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.CSRFPolicy, error) {
	var robj *models.CSRFPolicy
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing CSRFPolicy object with a given UUID
func (client *CSRFPolicyClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing CSRFPolicy object with a given name
func (client *CSRFPolicyClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *CSRFPolicyClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
