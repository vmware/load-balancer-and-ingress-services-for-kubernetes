// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// VsInventoryClient is a client for avi VsInventory resource
type VsInventoryClient struct {
	aviSession *session.AviSession
}

// NewVsInventoryClient creates a new client for VsInventory resource
func NewVsInventoryClient(aviSession *session.AviSession) *VsInventoryClient {
	return &VsInventoryClient{aviSession: aviSession}
}

func (client *VsInventoryClient) getAPIPath(uuid string) string {
	path := "api/vsinventory"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VsInventory objects
func (client *VsInventoryClient) GetAll(options ...session.ApiOptionsParams) ([]*models.VsInventory, error) {
	var plist []*models.VsInventory
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing VsInventory by uuid
func (client *VsInventoryClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.VsInventory, error) {
	var obj *models.VsInventory
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing VsInventory by name
func (client *VsInventoryClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.VsInventory, error) {
	var obj *models.VsInventory
	err := client.aviSession.GetObjectByName("vsinventory", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing VsInventory by filters like name, cloud, tenant
// Api creates VsInventory object with every call.
func (client *VsInventoryClient) GetObject(options ...session.ApiOptionsParams) (*models.VsInventory, error) {
	var obj *models.VsInventory
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vsinventory", newOptions...)
	return obj, err
}

// Create a new VsInventory object
func (client *VsInventoryClient) Create(obj *models.VsInventory, options ...session.ApiOptionsParams) (*models.VsInventory, error) {
	var robj *models.VsInventory
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing VsInventory object
func (client *VsInventoryClient) Update(obj *models.VsInventory, options ...session.ApiOptionsParams) (*models.VsInventory, error) {
	var robj *models.VsInventory
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing VsInventory object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VsInventory
// or it should be json compatible of form map[string]interface{}
func (client *VsInventoryClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.VsInventory, error) {
	var robj *models.VsInventory
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing VsInventory object with a given UUID
func (client *VsInventoryClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing VsInventory object with a given name
func (client *VsInventoryClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *VsInventoryClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
