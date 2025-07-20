// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// GslbInventoryClient is a client for avi GslbInventory resource
type GslbInventoryClient struct {
	aviSession *session.AviSession
}

// NewGslbInventoryClient creates a new client for GslbInventory resource
func NewGslbInventoryClient(aviSession *session.AviSession) *GslbInventoryClient {
	return &GslbInventoryClient{aviSession: aviSession}
}

func (client *GslbInventoryClient) getAPIPath(uuid string) string {
	path := "api/gslbinventory"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of GslbInventory objects
func (client *GslbInventoryClient) GetAll(options ...session.ApiOptionsParams) ([]*models.GslbInventory, error) {
	var plist []*models.GslbInventory
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing GslbInventory by uuid
func (client *GslbInventoryClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.GslbInventory, error) {
	var obj *models.GslbInventory
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing GslbInventory by name
func (client *GslbInventoryClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.GslbInventory, error) {
	var obj *models.GslbInventory
	err := client.aviSession.GetObjectByName("gslbinventory", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing GslbInventory by filters like name, cloud, tenant
// Api creates GslbInventory object with every call.
func (client *GslbInventoryClient) GetObject(options ...session.ApiOptionsParams) (*models.GslbInventory, error) {
	var obj *models.GslbInventory
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("gslbinventory", newOptions...)
	return obj, err
}

// Create a new GslbInventory object
func (client *GslbInventoryClient) Create(obj *models.GslbInventory, options ...session.ApiOptionsParams) (*models.GslbInventory, error) {
	var robj *models.GslbInventory
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing GslbInventory object
func (client *GslbInventoryClient) Update(obj *models.GslbInventory, options ...session.ApiOptionsParams) (*models.GslbInventory, error) {
	var robj *models.GslbInventory
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing GslbInventory object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.GslbInventory
// or it should be json compatible of form map[string]interface{}
func (client *GslbInventoryClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.GslbInventory, error) {
	var robj *models.GslbInventory
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing GslbInventory object with a given UUID
func (client *GslbInventoryClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing GslbInventory object with a given name
func (client *GslbInventoryClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *GslbInventoryClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
