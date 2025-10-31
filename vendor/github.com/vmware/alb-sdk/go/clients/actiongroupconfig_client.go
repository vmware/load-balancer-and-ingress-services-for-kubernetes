// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// ActionGroupConfigClient is a client for avi ActionGroupConfig resource
type ActionGroupConfigClient struct {
	aviSession *session.AviSession
}

// NewActionGroupConfigClient creates a new client for ActionGroupConfig resource
func NewActionGroupConfigClient(aviSession *session.AviSession) *ActionGroupConfigClient {
	return &ActionGroupConfigClient{aviSession: aviSession}
}

func (client *ActionGroupConfigClient) getAPIPath(uuid string) string {
	path := "api/actiongroupconfig"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ActionGroupConfig objects
func (client *ActionGroupConfigClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ActionGroupConfig, error) {
	var plist []*models.ActionGroupConfig
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ActionGroupConfig by uuid
func (client *ActionGroupConfigClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ActionGroupConfig, error) {
	var obj *models.ActionGroupConfig
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ActionGroupConfig by name
func (client *ActionGroupConfigClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ActionGroupConfig, error) {
	var obj *models.ActionGroupConfig
	err := client.aviSession.GetObjectByName("actiongroupconfig", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ActionGroupConfig by filters like name, cloud, tenant
// Api creates ActionGroupConfig object with every call.
func (client *ActionGroupConfigClient) GetObject(options ...session.ApiOptionsParams) (*models.ActionGroupConfig, error) {
	var obj *models.ActionGroupConfig
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("actiongroupconfig", newOptions...)
	return obj, err
}

// Create a new ActionGroupConfig object
func (client *ActionGroupConfigClient) Create(obj *models.ActionGroupConfig, options ...session.ApiOptionsParams) (*models.ActionGroupConfig, error) {
	var robj *models.ActionGroupConfig
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ActionGroupConfig object
func (client *ActionGroupConfigClient) Update(obj *models.ActionGroupConfig, options ...session.ApiOptionsParams) (*models.ActionGroupConfig, error) {
	var robj *models.ActionGroupConfig
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ActionGroupConfig object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ActionGroupConfig
// or it should be json compatible of form map[string]interface{}
func (client *ActionGroupConfigClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ActionGroupConfig, error) {
	var robj *models.ActionGroupConfig
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ActionGroupConfig object with a given UUID
func (client *ActionGroupConfigClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ActionGroupConfig object with a given name
func (client *ActionGroupConfigClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ActionGroupConfigClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
