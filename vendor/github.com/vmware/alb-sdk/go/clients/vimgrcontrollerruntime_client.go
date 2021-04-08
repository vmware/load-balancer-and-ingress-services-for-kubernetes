/***************************************************************************
 * ------------------------------------------------------------------------
 * Copyright 2020 VMware, Inc.  All rights reserved. VMware Confidential
 * ------------------------------------------------------------------------
 */

package clients

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// VIMgrControllerRuntimeClient is a client for avi VIMgrControllerRuntime resource
type VIMgrControllerRuntimeClient struct {
	aviSession *session.AviSession
}

// NewVIMgrControllerRuntimeClient creates a new client for VIMgrControllerRuntime resource
func NewVIMgrControllerRuntimeClient(aviSession *session.AviSession) *VIMgrControllerRuntimeClient {
	return &VIMgrControllerRuntimeClient{aviSession: aviSession}
}

func (client *VIMgrControllerRuntimeClient) getAPIPath(uuid string) string {
	path := "api/vimgrcontrollerruntime"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VIMgrControllerRuntime objects
func (client *VIMgrControllerRuntimeClient) GetAll(options ...session.ApiOptionsParams) ([]*models.VIMgrControllerRuntime, error) {
	var plist []*models.VIMgrControllerRuntime
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing VIMgrControllerRuntime by uuid
func (client *VIMgrControllerRuntimeClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.VIMgrControllerRuntime, error) {
	var obj *models.VIMgrControllerRuntime
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing VIMgrControllerRuntime by name
func (client *VIMgrControllerRuntimeClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.VIMgrControllerRuntime, error) {
	var obj *models.VIMgrControllerRuntime
	err := client.aviSession.GetObjectByName("vimgrcontrollerruntime", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing VIMgrControllerRuntime by filters like name, cloud, tenant
// Api creates VIMgrControllerRuntime object with every call.
func (client *VIMgrControllerRuntimeClient) GetObject(options ...session.ApiOptionsParams) (*models.VIMgrControllerRuntime, error) {
	var obj *models.VIMgrControllerRuntime
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vimgrcontrollerruntime", newOptions...)
	return obj, err
}

// Create a new VIMgrControllerRuntime object
func (client *VIMgrControllerRuntimeClient) Create(obj *models.VIMgrControllerRuntime, options ...session.ApiOptionsParams) (*models.VIMgrControllerRuntime, error) {
	var robj *models.VIMgrControllerRuntime
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing VIMgrControllerRuntime object
func (client *VIMgrControllerRuntimeClient) Update(obj *models.VIMgrControllerRuntime, options ...session.ApiOptionsParams) (*models.VIMgrControllerRuntime, error) {
	var robj *models.VIMgrControllerRuntime
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing VIMgrControllerRuntime object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VIMgrControllerRuntime
// or it should be json compatible of form map[string]interface{}
func (client *VIMgrControllerRuntimeClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.VIMgrControllerRuntime, error) {
	var robj *models.VIMgrControllerRuntime
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing VIMgrControllerRuntime object with a given UUID
func (client *VIMgrControllerRuntimeClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing VIMgrControllerRuntime object with a given name
func (client *VIMgrControllerRuntimeClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *VIMgrControllerRuntimeClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
