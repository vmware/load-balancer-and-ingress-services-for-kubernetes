// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// TenantSystemConfigurationClient is a client for avi TenantSystemConfiguration resource
type TenantSystemConfigurationClient struct {
	aviSession *session.AviSession
}

// NewTenantSystemConfigurationClient creates a new client for TenantSystemConfiguration resource
func NewTenantSystemConfigurationClient(aviSession *session.AviSession) *TenantSystemConfigurationClient {
	return &TenantSystemConfigurationClient{aviSession: aviSession}
}

func (client *TenantSystemConfigurationClient) getAPIPath(uuid string) string {
	path := "api/tenantsystemconfiguration"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of TenantSystemConfiguration objects
func (client *TenantSystemConfigurationClient) GetAll(options ...session.ApiOptionsParams) ([]*models.TenantSystemConfiguration, error) {
	var plist []*models.TenantSystemConfiguration
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing TenantSystemConfiguration by uuid
func (client *TenantSystemConfigurationClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.TenantSystemConfiguration, error) {
	var obj *models.TenantSystemConfiguration
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing TenantSystemConfiguration by name
func (client *TenantSystemConfigurationClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.TenantSystemConfiguration, error) {
	var obj *models.TenantSystemConfiguration
	err := client.aviSession.GetObjectByName("tenantsystemconfiguration", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing TenantSystemConfiguration by filters like name, cloud, tenant
// Api creates TenantSystemConfiguration object with every call.
func (client *TenantSystemConfigurationClient) GetObject(options ...session.ApiOptionsParams) (*models.TenantSystemConfiguration, error) {
	var obj *models.TenantSystemConfiguration
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("tenantsystemconfiguration", newOptions...)
	return obj, err
}

// Create a new TenantSystemConfiguration object
func (client *TenantSystemConfigurationClient) Create(obj *models.TenantSystemConfiguration, options ...session.ApiOptionsParams) (*models.TenantSystemConfiguration, error) {
	var robj *models.TenantSystemConfiguration
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing TenantSystemConfiguration object
func (client *TenantSystemConfigurationClient) Update(obj *models.TenantSystemConfiguration, options ...session.ApiOptionsParams) (*models.TenantSystemConfiguration, error) {
	var robj *models.TenantSystemConfiguration
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing TenantSystemConfiguration object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.TenantSystemConfiguration
// or it should be json compatible of form map[string]interface{}
func (client *TenantSystemConfigurationClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.TenantSystemConfiguration, error) {
	var robj *models.TenantSystemConfiguration
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing TenantSystemConfiguration object with a given UUID
func (client *TenantSystemConfigurationClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing TenantSystemConfiguration object with a given name
func (client *TenantSystemConfigurationClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *TenantSystemConfigurationClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
