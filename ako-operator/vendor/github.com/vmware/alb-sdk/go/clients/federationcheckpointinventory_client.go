// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// FederationCheckpointInventoryClient is a client for avi FederationCheckpointInventory resource
type FederationCheckpointInventoryClient struct {
	aviSession *session.AviSession
}

// NewFederationCheckpointInventoryClient creates a new client for FederationCheckpointInventory resource
func NewFederationCheckpointInventoryClient(aviSession *session.AviSession) *FederationCheckpointInventoryClient {
	return &FederationCheckpointInventoryClient{aviSession: aviSession}
}

func (client *FederationCheckpointInventoryClient) getAPIPath(uuid string) string {
	path := "api/federationcheckpointinventory"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of FederationCheckpointInventory objects
func (client *FederationCheckpointInventoryClient) GetAll(options ...session.ApiOptionsParams) ([]*models.FederationCheckpointInventory, error) {
	var plist []*models.FederationCheckpointInventory
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing FederationCheckpointInventory by uuid
func (client *FederationCheckpointInventoryClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.FederationCheckpointInventory, error) {
	var obj *models.FederationCheckpointInventory
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing FederationCheckpointInventory by name
func (client *FederationCheckpointInventoryClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.FederationCheckpointInventory, error) {
	var obj *models.FederationCheckpointInventory
	err := client.aviSession.GetObjectByName("federationcheckpointinventory", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing FederationCheckpointInventory by filters like name, cloud, tenant
// Api creates FederationCheckpointInventory object with every call.
func (client *FederationCheckpointInventoryClient) GetObject(options ...session.ApiOptionsParams) (*models.FederationCheckpointInventory, error) {
	var obj *models.FederationCheckpointInventory
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("federationcheckpointinventory", newOptions...)
	return obj, err
}

// Create a new FederationCheckpointInventory object
func (client *FederationCheckpointInventoryClient) Create(obj *models.FederationCheckpointInventory, options ...session.ApiOptionsParams) (*models.FederationCheckpointInventory, error) {
	var robj *models.FederationCheckpointInventory
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing FederationCheckpointInventory object
func (client *FederationCheckpointInventoryClient) Update(obj *models.FederationCheckpointInventory, options ...session.ApiOptionsParams) (*models.FederationCheckpointInventory, error) {
	var robj *models.FederationCheckpointInventory
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing FederationCheckpointInventory object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.FederationCheckpointInventory
// or it should be json compatible of form map[string]interface{}
func (client *FederationCheckpointInventoryClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.FederationCheckpointInventory, error) {
	var robj *models.FederationCheckpointInventory
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing FederationCheckpointInventory object with a given UUID
func (client *FederationCheckpointInventoryClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing FederationCheckpointInventory object with a given name
func (client *FederationCheckpointInventoryClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *FederationCheckpointInventoryClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
