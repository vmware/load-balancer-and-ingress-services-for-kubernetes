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

// SystemLimitsClient is a client for avi SystemLimits resource
type SystemLimitsClient struct {
	aviSession *session.AviSession
}

// NewSystemLimitsClient creates a new client for SystemLimits resource
func NewSystemLimitsClient(aviSession *session.AviSession) *SystemLimitsClient {
	return &SystemLimitsClient{aviSession: aviSession}
}

func (client *SystemLimitsClient) getAPIPath(uuid string) string {
	path := "api/systemlimits"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of SystemLimits objects
func (client *SystemLimitsClient) GetAll(options ...session.ApiOptionsParams) ([]*models.SystemLimits, error) {
	var plist []*models.SystemLimits
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing SystemLimits by uuid
func (client *SystemLimitsClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.SystemLimits, error) {
	var obj *models.SystemLimits
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing SystemLimits by name
func (client *SystemLimitsClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.SystemLimits, error) {
	var obj *models.SystemLimits
	err := client.aviSession.GetObjectByName("systemlimits", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing SystemLimits by filters like name, cloud, tenant
// Api creates SystemLimits object with every call.
func (client *SystemLimitsClient) GetObject(options ...session.ApiOptionsParams) (*models.SystemLimits, error) {
	var obj *models.SystemLimits
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("systemlimits", newOptions...)
	return obj, err
}

// Create a new SystemLimits object
func (client *SystemLimitsClient) Create(obj *models.SystemLimits, options ...session.ApiOptionsParams) (*models.SystemLimits, error) {
	var robj *models.SystemLimits
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing SystemLimits object
func (client *SystemLimitsClient) Update(obj *models.SystemLimits, options ...session.ApiOptionsParams) (*models.SystemLimits, error) {
	var robj *models.SystemLimits
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing SystemLimits object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.SystemLimits
// or it should be json compatible of form map[string]interface{}
func (client *SystemLimitsClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.SystemLimits, error) {
	var robj *models.SystemLimits
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing SystemLimits object with a given UUID
func (client *SystemLimitsClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing SystemLimits object with a given name
func (client *SystemLimitsClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *SystemLimitsClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
