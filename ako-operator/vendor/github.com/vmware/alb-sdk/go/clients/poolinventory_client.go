// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// PoolInventoryClient is a client for avi PoolInventory resource
type PoolInventoryClient struct {
	aviSession *session.AviSession
}

// NewPoolInventoryClient creates a new client for PoolInventory resource
func NewPoolInventoryClient(aviSession *session.AviSession) *PoolInventoryClient {
	return &PoolInventoryClient{aviSession: aviSession}
}

func (client *PoolInventoryClient) getAPIPath(uuid string) string {
	path := "api/poolinventory"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of PoolInventory objects
func (client *PoolInventoryClient) GetAll(options ...session.ApiOptionsParams) ([]*models.PoolInventory, error) {
	var plist []*models.PoolInventory
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing PoolInventory by uuid
func (client *PoolInventoryClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.PoolInventory, error) {
	var obj *models.PoolInventory
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing PoolInventory by name
func (client *PoolInventoryClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.PoolInventory, error) {
	var obj *models.PoolInventory
	err := client.aviSession.GetObjectByName("poolinventory", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing PoolInventory by filters like name, cloud, tenant
// Api creates PoolInventory object with every call.
func (client *PoolInventoryClient) GetObject(options ...session.ApiOptionsParams) (*models.PoolInventory, error) {
	var obj *models.PoolInventory
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("poolinventory", newOptions...)
	return obj, err
}

// Create a new PoolInventory object
func (client *PoolInventoryClient) Create(obj *models.PoolInventory, options ...session.ApiOptionsParams) (*models.PoolInventory, error) {
	var robj *models.PoolInventory
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing PoolInventory object
func (client *PoolInventoryClient) Update(obj *models.PoolInventory, options ...session.ApiOptionsParams) (*models.PoolInventory, error) {
	var robj *models.PoolInventory
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing PoolInventory object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.PoolInventory
// or it should be json compatible of form map[string]interface{}
func (client *PoolInventoryClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.PoolInventory, error) {
	var robj *models.PoolInventory
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing PoolInventory object with a given UUID
func (client *PoolInventoryClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing PoolInventory object with a given name
func (client *PoolInventoryClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *PoolInventoryClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
