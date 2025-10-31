// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// LogControllerMappingClient is a client for avi LogControllerMapping resource
type LogControllerMappingClient struct {
	aviSession *session.AviSession
}

// NewLogControllerMappingClient creates a new client for LogControllerMapping resource
func NewLogControllerMappingClient(aviSession *session.AviSession) *LogControllerMappingClient {
	return &LogControllerMappingClient{aviSession: aviSession}
}

func (client *LogControllerMappingClient) getAPIPath(uuid string) string {
	path := "api/logcontrollermapping"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of LogControllerMapping objects
func (client *LogControllerMappingClient) GetAll(options ...session.ApiOptionsParams) ([]*models.LogControllerMapping, error) {
	var plist []*models.LogControllerMapping
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing LogControllerMapping by uuid
func (client *LogControllerMappingClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.LogControllerMapping, error) {
	var obj *models.LogControllerMapping
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing LogControllerMapping by name
func (client *LogControllerMappingClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.LogControllerMapping, error) {
	var obj *models.LogControllerMapping
	err := client.aviSession.GetObjectByName("logcontrollermapping", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing LogControllerMapping by filters like name, cloud, tenant
// Api creates LogControllerMapping object with every call.
func (client *LogControllerMappingClient) GetObject(options ...session.ApiOptionsParams) (*models.LogControllerMapping, error) {
	var obj *models.LogControllerMapping
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("logcontrollermapping", newOptions...)
	return obj, err
}

// Create a new LogControllerMapping object
func (client *LogControllerMappingClient) Create(obj *models.LogControllerMapping, options ...session.ApiOptionsParams) (*models.LogControllerMapping, error) {
	var robj *models.LogControllerMapping
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing LogControllerMapping object
func (client *LogControllerMappingClient) Update(obj *models.LogControllerMapping, options ...session.ApiOptionsParams) (*models.LogControllerMapping, error) {
	var robj *models.LogControllerMapping
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing LogControllerMapping object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.LogControllerMapping
// or it should be json compatible of form map[string]interface{}
func (client *LogControllerMappingClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.LogControllerMapping, error) {
	var robj *models.LogControllerMapping
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing LogControllerMapping object with a given UUID
func (client *LogControllerMappingClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing LogControllerMapping object with a given name
func (client *LogControllerMappingClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *LogControllerMappingClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
