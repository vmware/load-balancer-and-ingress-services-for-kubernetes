/***************************************************************************
 * ------------------------------------------------------------------------
 * Copyright 2020 VMware, Inc.  All rights reserved. VMware Confidential
 * ------------------------------------------------------------------------
 */

package clients

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

import (
	"github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
)

// SecurityManagerDataClient is a client for avi SecurityManagerData resource
type SecurityManagerDataClient struct {
	aviSession *session.AviSession
}

// NewSecurityManagerDataClient creates a new client for SecurityManagerData resource
func NewSecurityManagerDataClient(aviSession *session.AviSession) *SecurityManagerDataClient {
	return &SecurityManagerDataClient{aviSession: aviSession}
}

func (client *SecurityManagerDataClient) getAPIPath(uuid string) string {
	path := "api/securitymanagerdata"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of SecurityManagerData objects
func (client *SecurityManagerDataClient) GetAll(options ...session.ApiOptionsParams) ([]*models.SecurityManagerData, error) {
	var plist []*models.SecurityManagerData
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing SecurityManagerData by uuid
func (client *SecurityManagerDataClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.SecurityManagerData, error) {
	var obj *models.SecurityManagerData
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing SecurityManagerData by name
func (client *SecurityManagerDataClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.SecurityManagerData, error) {
	var obj *models.SecurityManagerData
	err := client.aviSession.GetObjectByName("securitymanagerdata", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing SecurityManagerData by filters like name, cloud, tenant
// Api creates SecurityManagerData object with every call.
func (client *SecurityManagerDataClient) GetObject(options ...session.ApiOptionsParams) (*models.SecurityManagerData, error) {
	var obj *models.SecurityManagerData
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("securitymanagerdata", newOptions...)
	return obj, err
}

// Create a new SecurityManagerData object
func (client *SecurityManagerDataClient) Create(obj *models.SecurityManagerData, options ...session.ApiOptionsParams) (*models.SecurityManagerData, error) {
	var robj *models.SecurityManagerData
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing SecurityManagerData object
func (client *SecurityManagerDataClient) Update(obj *models.SecurityManagerData, options ...session.ApiOptionsParams) (*models.SecurityManagerData, error) {
	var robj *models.SecurityManagerData
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing SecurityManagerData object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.SecurityManagerData
// or it should be json compatible of form map[string]interface{}
func (client *SecurityManagerDataClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.SecurityManagerData, error) {
	var robj *models.SecurityManagerData
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing SecurityManagerData object with a given UUID
func (client *SecurityManagerDataClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing SecurityManagerData object with a given name
func (client *SecurityManagerDataClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *SecurityManagerDataClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
