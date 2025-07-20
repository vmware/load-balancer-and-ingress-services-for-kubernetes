// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// ServiceEngineGroupInventoryClient is a client for avi ServiceEngineGroupInventory resource
type ServiceEngineGroupInventoryClient struct {
	aviSession *session.AviSession
}

// NewServiceEngineGroupInventoryClient creates a new client for ServiceEngineGroupInventory resource
func NewServiceEngineGroupInventoryClient(aviSession *session.AviSession) *ServiceEngineGroupInventoryClient {
	return &ServiceEngineGroupInventoryClient{aviSession: aviSession}
}

func (client *ServiceEngineGroupInventoryClient) getAPIPath(uuid string) string {
	path := "api/serviceenginegroupinventory"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ServiceEngineGroupInventory objects
func (client *ServiceEngineGroupInventoryClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ServiceEngineGroupInventory, error) {
	var plist []*models.ServiceEngineGroupInventory
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ServiceEngineGroupInventory by uuid
func (client *ServiceEngineGroupInventoryClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ServiceEngineGroupInventory, error) {
	var obj *models.ServiceEngineGroupInventory
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ServiceEngineGroupInventory by name
func (client *ServiceEngineGroupInventoryClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ServiceEngineGroupInventory, error) {
	var obj *models.ServiceEngineGroupInventory
	err := client.aviSession.GetObjectByName("serviceenginegroupinventory", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ServiceEngineGroupInventory by filters like name, cloud, tenant
// Api creates ServiceEngineGroupInventory object with every call.
func (client *ServiceEngineGroupInventoryClient) GetObject(options ...session.ApiOptionsParams) (*models.ServiceEngineGroupInventory, error) {
	var obj *models.ServiceEngineGroupInventory
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("serviceenginegroupinventory", newOptions...)
	return obj, err
}

// Create a new ServiceEngineGroupInventory object
func (client *ServiceEngineGroupInventoryClient) Create(obj *models.ServiceEngineGroupInventory, options ...session.ApiOptionsParams) (*models.ServiceEngineGroupInventory, error) {
	var robj *models.ServiceEngineGroupInventory
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ServiceEngineGroupInventory object
func (client *ServiceEngineGroupInventoryClient) Update(obj *models.ServiceEngineGroupInventory, options ...session.ApiOptionsParams) (*models.ServiceEngineGroupInventory, error) {
	var robj *models.ServiceEngineGroupInventory
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ServiceEngineGroupInventory object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ServiceEngineGroupInventory
// or it should be json compatible of form map[string]interface{}
func (client *ServiceEngineGroupInventoryClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ServiceEngineGroupInventory, error) {
	var robj *models.ServiceEngineGroupInventory
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ServiceEngineGroupInventory object with a given UUID
func (client *ServiceEngineGroupInventoryClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ServiceEngineGroupInventory object with a given name
func (client *ServiceEngineGroupInventoryClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ServiceEngineGroupInventoryClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
