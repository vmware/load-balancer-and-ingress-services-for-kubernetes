// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// NetworkInventoryClient is a client for avi NetworkInventory resource
type NetworkInventoryClient struct {
	aviSession *session.AviSession
}

// NewNetworkInventoryClient creates a new client for NetworkInventory resource
func NewNetworkInventoryClient(aviSession *session.AviSession) *NetworkInventoryClient {
	return &NetworkInventoryClient{aviSession: aviSession}
}

func (client *NetworkInventoryClient) getAPIPath(uuid string) string {
	path := "api/networkinventory"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of NetworkInventory objects
func (client *NetworkInventoryClient) GetAll(options ...session.ApiOptionsParams) ([]*models.NetworkInventory, error) {
	var plist []*models.NetworkInventory
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing NetworkInventory by uuid
func (client *NetworkInventoryClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.NetworkInventory, error) {
	var obj *models.NetworkInventory
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing NetworkInventory by name
func (client *NetworkInventoryClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.NetworkInventory, error) {
	var obj *models.NetworkInventory
	err := client.aviSession.GetObjectByName("networkinventory", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing NetworkInventory by filters like name, cloud, tenant
// Api creates NetworkInventory object with every call.
func (client *NetworkInventoryClient) GetObject(options ...session.ApiOptionsParams) (*models.NetworkInventory, error) {
	var obj *models.NetworkInventory
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("networkinventory", newOptions...)
	return obj, err
}

// Create a new NetworkInventory object
func (client *NetworkInventoryClient) Create(obj *models.NetworkInventory, options ...session.ApiOptionsParams) (*models.NetworkInventory, error) {
	var robj *models.NetworkInventory
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing NetworkInventory object
func (client *NetworkInventoryClient) Update(obj *models.NetworkInventory, options ...session.ApiOptionsParams) (*models.NetworkInventory, error) {
	var robj *models.NetworkInventory
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing NetworkInventory object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.NetworkInventory
// or it should be json compatible of form map[string]interface{}
func (client *NetworkInventoryClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.NetworkInventory, error) {
	var robj *models.NetworkInventory
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing NetworkInventory object with a given UUID
func (client *NetworkInventoryClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing NetworkInventory object with a given name
func (client *NetworkInventoryClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *NetworkInventoryClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
