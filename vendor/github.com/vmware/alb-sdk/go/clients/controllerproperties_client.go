// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// ControllerPropertiesClient is a client for avi ControllerProperties resource
type ControllerPropertiesClient struct {
	aviSession *session.AviSession
}

// NewControllerPropertiesClient creates a new client for ControllerProperties resource
func NewControllerPropertiesClient(aviSession *session.AviSession) *ControllerPropertiesClient {
	return &ControllerPropertiesClient{aviSession: aviSession}
}

func (client *ControllerPropertiesClient) getAPIPath(uuid string) string {
	path := "api/controllerproperties"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ControllerProperties objects
func (client *ControllerPropertiesClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ControllerProperties, error) {
	var plist []*models.ControllerProperties
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ControllerProperties by uuid
func (client *ControllerPropertiesClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ControllerProperties, error) {
	var obj *models.ControllerProperties
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ControllerProperties by name
func (client *ControllerPropertiesClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ControllerProperties, error) {
	var obj *models.ControllerProperties
	err := client.aviSession.GetObjectByName("controllerproperties", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ControllerProperties by filters like name, cloud, tenant
// Api creates ControllerProperties object with every call.
func (client *ControllerPropertiesClient) GetObject(options ...session.ApiOptionsParams) (*models.ControllerProperties, error) {
	var obj *models.ControllerProperties
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("controllerproperties", newOptions...)
	return obj, err
}

// Create a new ControllerProperties object
func (client *ControllerPropertiesClient) Create(obj *models.ControllerProperties, options ...session.ApiOptionsParams) (*models.ControllerProperties, error) {
	var robj *models.ControllerProperties
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ControllerProperties object
func (client *ControllerPropertiesClient) Update(obj *models.ControllerProperties, options ...session.ApiOptionsParams) (*models.ControllerProperties, error) {
	var robj *models.ControllerProperties
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ControllerProperties object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ControllerProperties
// or it should be json compatible of form map[string]interface{}
func (client *ControllerPropertiesClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ControllerProperties, error) {
	var robj *models.ControllerProperties
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ControllerProperties object with a given UUID
func (client *ControllerPropertiesClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ControllerProperties object with a given name
func (client *ControllerPropertiesClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ControllerPropertiesClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
