// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// PoolGroupInventoryClient is a client for avi PoolGroupInventory resource
type PoolGroupInventoryClient struct {
	aviSession *session.AviSession
}

// NewPoolGroupInventoryClient creates a new client for PoolGroupInventory resource
func NewPoolGroupInventoryClient(aviSession *session.AviSession) *PoolGroupInventoryClient {
	return &PoolGroupInventoryClient{aviSession: aviSession}
}

func (client *PoolGroupInventoryClient) getAPIPath(uuid string) string {
	path := "api/poolgroupinventory"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of PoolGroupInventory objects
func (client *PoolGroupInventoryClient) GetAll(options ...session.ApiOptionsParams) ([]*models.PoolGroupInventory, error) {
	var plist []*models.PoolGroupInventory
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing PoolGroupInventory by uuid
func (client *PoolGroupInventoryClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.PoolGroupInventory, error) {
	var obj *models.PoolGroupInventory
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing PoolGroupInventory by name
func (client *PoolGroupInventoryClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.PoolGroupInventory, error) {
	var obj *models.PoolGroupInventory
	err := client.aviSession.GetObjectByName("poolgroupinventory", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing PoolGroupInventory by filters like name, cloud, tenant
// Api creates PoolGroupInventory object with every call.
func (client *PoolGroupInventoryClient) GetObject(options ...session.ApiOptionsParams) (*models.PoolGroupInventory, error) {
	var obj *models.PoolGroupInventory
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("poolgroupinventory", newOptions...)
	return obj, err
}

// Create a new PoolGroupInventory object
func (client *PoolGroupInventoryClient) Create(obj *models.PoolGroupInventory, options ...session.ApiOptionsParams) (*models.PoolGroupInventory, error) {
	var robj *models.PoolGroupInventory
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing PoolGroupInventory object
func (client *PoolGroupInventoryClient) Update(obj *models.PoolGroupInventory, options ...session.ApiOptionsParams) (*models.PoolGroupInventory, error) {
	var robj *models.PoolGroupInventory
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing PoolGroupInventory object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.PoolGroupInventory
// or it should be json compatible of form map[string]interface{}
func (client *PoolGroupInventoryClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.PoolGroupInventory, error) {
	var robj *models.PoolGroupInventory
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing PoolGroupInventory object with a given UUID
func (client *PoolGroupInventoryClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing PoolGroupInventory object with a given name
func (client *PoolGroupInventoryClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *PoolGroupInventoryClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
