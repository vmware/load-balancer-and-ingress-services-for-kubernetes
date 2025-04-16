// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// VsGsClient is a client for avi VsGs resource
type VsGsClient struct {
	aviSession *session.AviSession
}

// NewVsGsClient creates a new client for VsGs resource
func NewVsGsClient(aviSession *session.AviSession) *VsGsClient {
	return &VsGsClient{aviSession: aviSession}
}

func (client *VsGsClient) getAPIPath(uuid string) string {
	path := "api/vsgs"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VsGs objects
func (client *VsGsClient) GetAll(options ...session.ApiOptionsParams) ([]*models.VsGs, error) {
	var plist []*models.VsGs
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing VsGs by uuid
func (client *VsGsClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.VsGs, error) {
	var obj *models.VsGs
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing VsGs by name
func (client *VsGsClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.VsGs, error) {
	var obj *models.VsGs
	err := client.aviSession.GetObjectByName("vsgs", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing VsGs by filters like name, cloud, tenant
// Api creates VsGs object with every call.
func (client *VsGsClient) GetObject(options ...session.ApiOptionsParams) (*models.VsGs, error) {
	var obj *models.VsGs
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vsgs", newOptions...)
	return obj, err
}

// Create a new VsGs object
func (client *VsGsClient) Create(obj *models.VsGs, options ...session.ApiOptionsParams) (*models.VsGs, error) {
	var robj *models.VsGs
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing VsGs object
func (client *VsGsClient) Update(obj *models.VsGs, options ...session.ApiOptionsParams) (*models.VsGs, error) {
	var robj *models.VsGs
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing VsGs object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VsGs
// or it should be json compatible of form map[string]interface{}
func (client *VsGsClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.VsGs, error) {
	var robj *models.VsGs
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing VsGs object with a given UUID
func (client *VsGsClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing VsGs object with a given name
func (client *VsGsClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *VsGsClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
