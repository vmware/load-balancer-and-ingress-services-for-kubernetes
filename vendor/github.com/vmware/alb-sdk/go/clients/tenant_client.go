/***************************************************************************
 * ------------------------------------------------------------------------
 * Copyright 2020 VMware, Inc.  All rights reserved. VMware Confidential
 * ------------------------------------------------------------------------
 */

package clients

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// TenantClient is a client for avi Tenant resource
type TenantClient struct {
	aviSession *session.AviSession
}

// NewTenantClient creates a new client for Tenant resource
func NewTenantClient(aviSession *session.AviSession) *TenantClient {
	return &TenantClient{aviSession: aviSession}
}

func (client *TenantClient) getAPIPath(uuid string) string {
	path := "api/tenant"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of Tenant objects
func (client *TenantClient) GetAll(options ...session.ApiOptionsParams) ([]*models.Tenant, error) {
	var plist []*models.Tenant
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing Tenant by uuid
func (client *TenantClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.Tenant, error) {
	var obj *models.Tenant
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing Tenant by name
func (client *TenantClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.Tenant, error) {
	var obj *models.Tenant
	err := client.aviSession.GetObjectByName("tenant", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing Tenant by filters like name, cloud, tenant
// Api creates Tenant object with every call.
func (client *TenantClient) GetObject(options ...session.ApiOptionsParams) (*models.Tenant, error) {
	var obj *models.Tenant
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("tenant", newOptions...)
	return obj, err
}

// Create a new Tenant object
func (client *TenantClient) Create(obj *models.Tenant, options ...session.ApiOptionsParams) (*models.Tenant, error) {
	var robj *models.Tenant
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing Tenant object
func (client *TenantClient) Update(obj *models.Tenant, options ...session.ApiOptionsParams) (*models.Tenant, error) {
	var robj *models.Tenant
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing Tenant object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.Tenant
// or it should be json compatible of form map[string]interface{}
func (client *TenantClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.Tenant, error) {
	var robj *models.Tenant
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing Tenant object with a given UUID
func (client *TenantClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing Tenant object with a given name
func (client *TenantClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *TenantClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
