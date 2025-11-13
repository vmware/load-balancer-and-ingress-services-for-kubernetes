// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// ServiceEngineInventoryClient is a client for avi ServiceEngineInventory resource
type ServiceEngineInventoryClient struct {
	aviSession *session.AviSession
}

// NewServiceEngineInventoryClient creates a new client for ServiceEngineInventory resource
func NewServiceEngineInventoryClient(aviSession *session.AviSession) *ServiceEngineInventoryClient {
	return &ServiceEngineInventoryClient{aviSession: aviSession}
}

func (client *ServiceEngineInventoryClient) getAPIPath(uuid string) string {
	path := "api/serviceengineinventory"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ServiceEngineInventory objects
func (client *ServiceEngineInventoryClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ServiceEngineInventory, error) {
	var plist []*models.ServiceEngineInventory
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ServiceEngineInventory by uuid
func (client *ServiceEngineInventoryClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ServiceEngineInventory, error) {
	var obj *models.ServiceEngineInventory
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ServiceEngineInventory by name
func (client *ServiceEngineInventoryClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ServiceEngineInventory, error) {
	var obj *models.ServiceEngineInventory
	err := client.aviSession.GetObjectByName("serviceengineinventory", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ServiceEngineInventory by filters like name, cloud, tenant
// Api creates ServiceEngineInventory object with every call.
func (client *ServiceEngineInventoryClient) GetObject(options ...session.ApiOptionsParams) (*models.ServiceEngineInventory, error) {
	var obj *models.ServiceEngineInventory
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("serviceengineinventory", newOptions...)
	return obj, err
}

// Create a new ServiceEngineInventory object
func (client *ServiceEngineInventoryClient) Create(obj *models.ServiceEngineInventory, options ...session.ApiOptionsParams) (*models.ServiceEngineInventory, error) {
	var robj *models.ServiceEngineInventory
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ServiceEngineInventory object
func (client *ServiceEngineInventoryClient) Update(obj *models.ServiceEngineInventory, options ...session.ApiOptionsParams) (*models.ServiceEngineInventory, error) {
	var robj *models.ServiceEngineInventory
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ServiceEngineInventory object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ServiceEngineInventory
// or it should be json compatible of form map[string]interface{}
func (client *ServiceEngineInventoryClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ServiceEngineInventory, error) {
	var robj *models.ServiceEngineInventory
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ServiceEngineInventory object with a given UUID
func (client *ServiceEngineInventoryClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ServiceEngineInventory object with a given name
func (client *ServiceEngineInventoryClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ServiceEngineInventoryClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
