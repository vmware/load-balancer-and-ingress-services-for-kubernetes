// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// VsvipInventoryClient is a client for avi VsvipInventory resource
type VsvipInventoryClient struct {
	aviSession *session.AviSession
}

// NewVsvipInventoryClient creates a new client for VsvipInventory resource
func NewVsvipInventoryClient(aviSession *session.AviSession) *VsvipInventoryClient {
	return &VsvipInventoryClient{aviSession: aviSession}
}

func (client *VsvipInventoryClient) getAPIPath(uuid string) string {
	path := "api/vsvipinventory"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VsvipInventory objects
func (client *VsvipInventoryClient) GetAll(options ...session.ApiOptionsParams) ([]*models.VsvipInventory, error) {
	var plist []*models.VsvipInventory
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing VsvipInventory by uuid
func (client *VsvipInventoryClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.VsvipInventory, error) {
	var obj *models.VsvipInventory
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing VsvipInventory by name
func (client *VsvipInventoryClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.VsvipInventory, error) {
	var obj *models.VsvipInventory
	err := client.aviSession.GetObjectByName("vsvipinventory", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing VsvipInventory by filters like name, cloud, tenant
// Api creates VsvipInventory object with every call.
func (client *VsvipInventoryClient) GetObject(options ...session.ApiOptionsParams) (*models.VsvipInventory, error) {
	var obj *models.VsvipInventory
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vsvipinventory", newOptions...)
	return obj, err
}

// Create a new VsvipInventory object
func (client *VsvipInventoryClient) Create(obj *models.VsvipInventory, options ...session.ApiOptionsParams) (*models.VsvipInventory, error) {
	var robj *models.VsvipInventory
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing VsvipInventory object
func (client *VsvipInventoryClient) Update(obj *models.VsvipInventory, options ...session.ApiOptionsParams) (*models.VsvipInventory, error) {
	var robj *models.VsvipInventory
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing VsvipInventory object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VsvipInventory
// or it should be json compatible of form map[string]interface{}
func (client *VsvipInventoryClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.VsvipInventory, error) {
	var robj *models.VsvipInventory
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing VsvipInventory object with a given UUID
func (client *VsvipInventoryClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing VsvipInventory object with a given name
func (client *VsvipInventoryClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *VsvipInventoryClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
