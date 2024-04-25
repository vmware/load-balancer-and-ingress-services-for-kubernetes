// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// BotMappingClient is a client for avi BotMapping resource
type BotMappingClient struct {
	aviSession *session.AviSession
}

// NewBotMappingClient creates a new client for BotMapping resource
func NewBotMappingClient(aviSession *session.AviSession) *BotMappingClient {
	return &BotMappingClient{aviSession: aviSession}
}

func (client *BotMappingClient) getAPIPath(uuid string) string {
	path := "api/botmapping"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of BotMapping objects
func (client *BotMappingClient) GetAll(options ...session.ApiOptionsParams) ([]*models.BotMapping, error) {
	var plist []*models.BotMapping
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing BotMapping by uuid
func (client *BotMappingClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.BotMapping, error) {
	var obj *models.BotMapping
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing BotMapping by name
func (client *BotMappingClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.BotMapping, error) {
	var obj *models.BotMapping
	err := client.aviSession.GetObjectByName("botmapping", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing BotMapping by filters like name, cloud, tenant
// Api creates BotMapping object with every call.
func (client *BotMappingClient) GetObject(options ...session.ApiOptionsParams) (*models.BotMapping, error) {
	var obj *models.BotMapping
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("botmapping", newOptions...)
	return obj, err
}

// Create a new BotMapping object
func (client *BotMappingClient) Create(obj *models.BotMapping, options ...session.ApiOptionsParams) (*models.BotMapping, error) {
	var robj *models.BotMapping
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing BotMapping object
func (client *BotMappingClient) Update(obj *models.BotMapping, options ...session.ApiOptionsParams) (*models.BotMapping, error) {
	var robj *models.BotMapping
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing BotMapping object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.BotMapping
// or it should be json compatible of form map[string]interface{}
func (client *BotMappingClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.BotMapping, error) {
	var robj *models.BotMapping
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing BotMapping object with a given UUID
func (client *BotMappingClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing BotMapping object with a given name
func (client *BotMappingClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *BotMappingClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
