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

// CloudConnectorUserClient is a client for avi CloudConnectorUser resource
type CloudConnectorUserClient struct {
	aviSession *session.AviSession
}

// NewCloudConnectorUserClient creates a new client for CloudConnectorUser resource
func NewCloudConnectorUserClient(aviSession *session.AviSession) *CloudConnectorUserClient {
	return &CloudConnectorUserClient{aviSession: aviSession}
}

func (client *CloudConnectorUserClient) getAPIPath(uuid string) string {
	path := "api/cloudconnectoruser"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of CloudConnectorUser objects
func (client *CloudConnectorUserClient) GetAll(options ...session.ApiOptionsParams) ([]*models.CloudConnectorUser, error) {
	var plist []*models.CloudConnectorUser
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing CloudConnectorUser by uuid
func (client *CloudConnectorUserClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.CloudConnectorUser, error) {
	var obj *models.CloudConnectorUser
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing CloudConnectorUser by name
func (client *CloudConnectorUserClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.CloudConnectorUser, error) {
	var obj *models.CloudConnectorUser
	err := client.aviSession.GetObjectByName("cloudconnectoruser", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing CloudConnectorUser by filters like name, cloud, tenant
// Api creates CloudConnectorUser object with every call.
func (client *CloudConnectorUserClient) GetObject(options ...session.ApiOptionsParams) (*models.CloudConnectorUser, error) {
	var obj *models.CloudConnectorUser
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("cloudconnectoruser", newOptions...)
	return obj, err
}

// Create a new CloudConnectorUser object
func (client *CloudConnectorUserClient) Create(obj *models.CloudConnectorUser, options ...session.ApiOptionsParams) (*models.CloudConnectorUser, error) {
	var robj *models.CloudConnectorUser
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing CloudConnectorUser object
func (client *CloudConnectorUserClient) Update(obj *models.CloudConnectorUser, options ...session.ApiOptionsParams) (*models.CloudConnectorUser, error) {
	var robj *models.CloudConnectorUser
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing CloudConnectorUser object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.CloudConnectorUser
// or it should be json compatible of form map[string]interface{}
func (client *CloudConnectorUserClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.CloudConnectorUser, error) {
	var robj *models.CloudConnectorUser
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing CloudConnectorUser object with a given UUID
func (client *CloudConnectorUserClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing CloudConnectorUser object with a given name
func (client *CloudConnectorUserClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *CloudConnectorUserClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
