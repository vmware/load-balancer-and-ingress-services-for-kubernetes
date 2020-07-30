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

// ALBServicesConfigClient is a client for avi ALBServicesConfig resource
type ALBServicesConfigClient struct {
	aviSession *session.AviSession
}

// NewALBServicesConfigClient creates a new client for ALBServicesConfig resource
func NewALBServicesConfigClient(aviSession *session.AviSession) *ALBServicesConfigClient {
	return &ALBServicesConfigClient{aviSession: aviSession}
}

func (client *ALBServicesConfigClient) getAPIPath(uuid string) string {
	path := "api/albservicesconfig"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ALBServicesConfig objects
func (client *ALBServicesConfigClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ALBServicesConfig, error) {
	var plist []*models.ALBServicesConfig
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ALBServicesConfig by uuid
func (client *ALBServicesConfigClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ALBServicesConfig, error) {
	var obj *models.ALBServicesConfig
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ALBServicesConfig by name
func (client *ALBServicesConfigClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ALBServicesConfig, error) {
	var obj *models.ALBServicesConfig
	err := client.aviSession.GetObjectByName("albservicesconfig", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ALBServicesConfig by filters like name, cloud, tenant
// Api creates ALBServicesConfig object with every call.
func (client *ALBServicesConfigClient) GetObject(options ...session.ApiOptionsParams) (*models.ALBServicesConfig, error) {
	var obj *models.ALBServicesConfig
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("albservicesconfig", newOptions...)
	return obj, err
}

// Create a new ALBServicesConfig object
func (client *ALBServicesConfigClient) Create(obj *models.ALBServicesConfig, options ...session.ApiOptionsParams) (*models.ALBServicesConfig, error) {
	var robj *models.ALBServicesConfig
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ALBServicesConfig object
func (client *ALBServicesConfigClient) Update(obj *models.ALBServicesConfig, options ...session.ApiOptionsParams) (*models.ALBServicesConfig, error) {
	var robj *models.ALBServicesConfig
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ALBServicesConfig object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ALBServicesConfig
// or it should be json compatible of form map[string]interface{}
func (client *ALBServicesConfigClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ALBServicesConfig, error) {
	var robj *models.ALBServicesConfig
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ALBServicesConfig object with a given UUID
func (client *ALBServicesConfigClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ALBServicesConfig object with a given name
func (client *ALBServicesConfigClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ALBServicesConfigClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
