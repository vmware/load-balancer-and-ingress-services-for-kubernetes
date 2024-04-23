// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// InventoryFaultConfigClient is a client for avi InventoryFaultConfig resource
type InventoryFaultConfigClient struct {
	aviSession *session.AviSession
}

// NewInventoryFaultConfigClient creates a new client for InventoryFaultConfig resource
func NewInventoryFaultConfigClient(aviSession *session.AviSession) *InventoryFaultConfigClient {
	return &InventoryFaultConfigClient{aviSession: aviSession}
}

func (client *InventoryFaultConfigClient) getAPIPath(uuid string) string {
	path := "api/inventoryfaultconfig"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of InventoryFaultConfig objects
func (client *InventoryFaultConfigClient) GetAll(options ...session.ApiOptionsParams) ([]*models.InventoryFaultConfig, error) {
	var plist []*models.InventoryFaultConfig
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing InventoryFaultConfig by uuid
func (client *InventoryFaultConfigClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.InventoryFaultConfig, error) {
	var obj *models.InventoryFaultConfig
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing InventoryFaultConfig by name
func (client *InventoryFaultConfigClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.InventoryFaultConfig, error) {
	var obj *models.InventoryFaultConfig
	err := client.aviSession.GetObjectByName("inventoryfaultconfig", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing InventoryFaultConfig by filters like name, cloud, tenant
// Api creates InventoryFaultConfig object with every call.
func (client *InventoryFaultConfigClient) GetObject(options ...session.ApiOptionsParams) (*models.InventoryFaultConfig, error) {
	var obj *models.InventoryFaultConfig
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("inventoryfaultconfig", newOptions...)
	return obj, err
}

// Create a new InventoryFaultConfig object
func (client *InventoryFaultConfigClient) Create(obj *models.InventoryFaultConfig, options ...session.ApiOptionsParams) (*models.InventoryFaultConfig, error) {
	var robj *models.InventoryFaultConfig
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing InventoryFaultConfig object
func (client *InventoryFaultConfigClient) Update(obj *models.InventoryFaultConfig, options ...session.ApiOptionsParams) (*models.InventoryFaultConfig, error) {
	var robj *models.InventoryFaultConfig
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing InventoryFaultConfig object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.InventoryFaultConfig
// or it should be json compatible of form map[string]interface{}
func (client *InventoryFaultConfigClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.InventoryFaultConfig, error) {
	var robj *models.InventoryFaultConfig
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing InventoryFaultConfig object with a given UUID
func (client *InventoryFaultConfigClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing InventoryFaultConfig object with a given name
func (client *InventoryFaultConfigClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *InventoryFaultConfigClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
