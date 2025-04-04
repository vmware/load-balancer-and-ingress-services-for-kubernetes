// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// GslbServiceInventoryClient is a client for avi GslbServiceInventory resource
type GslbServiceInventoryClient struct {
	aviSession *session.AviSession
}

// NewGslbServiceInventoryClient creates a new client for GslbServiceInventory resource
func NewGslbServiceInventoryClient(aviSession *session.AviSession) *GslbServiceInventoryClient {
	return &GslbServiceInventoryClient{aviSession: aviSession}
}

func (client *GslbServiceInventoryClient) getAPIPath(uuid string) string {
	path := "api/gslbserviceinventory"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of GslbServiceInventory objects
func (client *GslbServiceInventoryClient) GetAll(options ...session.ApiOptionsParams) ([]*models.GslbServiceInventory, error) {
	var plist []*models.GslbServiceInventory
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing GslbServiceInventory by uuid
func (client *GslbServiceInventoryClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.GslbServiceInventory, error) {
	var obj *models.GslbServiceInventory
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing GslbServiceInventory by name
func (client *GslbServiceInventoryClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.GslbServiceInventory, error) {
	var obj *models.GslbServiceInventory
	err := client.aviSession.GetObjectByName("gslbserviceinventory", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing GslbServiceInventory by filters like name, cloud, tenant
// Api creates GslbServiceInventory object with every call.
func (client *GslbServiceInventoryClient) GetObject(options ...session.ApiOptionsParams) (*models.GslbServiceInventory, error) {
	var obj *models.GslbServiceInventory
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("gslbserviceinventory", newOptions...)
	return obj, err
}

// Create a new GslbServiceInventory object
func (client *GslbServiceInventoryClient) Create(obj *models.GslbServiceInventory, options ...session.ApiOptionsParams) (*models.GslbServiceInventory, error) {
	var robj *models.GslbServiceInventory
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing GslbServiceInventory object
func (client *GslbServiceInventoryClient) Update(obj *models.GslbServiceInventory, options ...session.ApiOptionsParams) (*models.GslbServiceInventory, error) {
	var robj *models.GslbServiceInventory
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing GslbServiceInventory object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.GslbServiceInventory
// or it should be json compatible of form map[string]interface{}
func (client *GslbServiceInventoryClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.GslbServiceInventory, error) {
	var robj *models.GslbServiceInventory
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing GslbServiceInventory object with a given UUID
func (client *GslbServiceInventoryClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing GslbServiceInventory object with a given name
func (client *GslbServiceInventoryClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *GslbServiceInventoryClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
