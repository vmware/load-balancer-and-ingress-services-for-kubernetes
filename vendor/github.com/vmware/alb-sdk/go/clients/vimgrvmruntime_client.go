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

// VIMgrVMRuntimeClient is a client for avi VIMgrVMRuntime resource
type VIMgrVMRuntimeClient struct {
	aviSession *session.AviSession
}

// NewVIMgrVMRuntimeClient creates a new client for VIMgrVMRuntime resource
func NewVIMgrVMRuntimeClient(aviSession *session.AviSession) *VIMgrVMRuntimeClient {
	return &VIMgrVMRuntimeClient{aviSession: aviSession}
}

func (client *VIMgrVMRuntimeClient) getAPIPath(uuid string) string {
	path := "api/vimgrvmruntime"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VIMgrVMRuntime objects
func (client *VIMgrVMRuntimeClient) GetAll(options ...session.ApiOptionsParams) ([]*models.VIMgrVMRuntime, error) {
	var plist []*models.VIMgrVMRuntime
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing VIMgrVMRuntime by uuid
func (client *VIMgrVMRuntimeClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.VIMgrVMRuntime, error) {
	var obj *models.VIMgrVMRuntime
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing VIMgrVMRuntime by name
func (client *VIMgrVMRuntimeClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.VIMgrVMRuntime, error) {
	var obj *models.VIMgrVMRuntime
	err := client.aviSession.GetObjectByName("vimgrvmruntime", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing VIMgrVMRuntime by filters like name, cloud, tenant
// Api creates VIMgrVMRuntime object with every call.
func (client *VIMgrVMRuntimeClient) GetObject(options ...session.ApiOptionsParams) (*models.VIMgrVMRuntime, error) {
	var obj *models.VIMgrVMRuntime
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vimgrvmruntime", newOptions...)
	return obj, err
}

// Create a new VIMgrVMRuntime object
func (client *VIMgrVMRuntimeClient) Create(obj *models.VIMgrVMRuntime, options ...session.ApiOptionsParams) (*models.VIMgrVMRuntime, error) {
	var robj *models.VIMgrVMRuntime
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing VIMgrVMRuntime object
func (client *VIMgrVMRuntimeClient) Update(obj *models.VIMgrVMRuntime, options ...session.ApiOptionsParams) (*models.VIMgrVMRuntime, error) {
	var robj *models.VIMgrVMRuntime
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing VIMgrVMRuntime object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VIMgrVMRuntime
// or it should be json compatible of form map[string]interface{}
func (client *VIMgrVMRuntimeClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.VIMgrVMRuntime, error) {
	var robj *models.VIMgrVMRuntime
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing VIMgrVMRuntime object with a given UUID
func (client *VIMgrVMRuntimeClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing VIMgrVMRuntime object with a given name
func (client *VIMgrVMRuntimeClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *VIMgrVMRuntimeClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
