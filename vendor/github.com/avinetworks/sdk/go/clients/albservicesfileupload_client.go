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

// ALBServicesFileUploadClient is a client for avi ALBServicesFileUpload resource
type ALBServicesFileUploadClient struct {
	aviSession *session.AviSession
}

// NewALBServicesFileUploadClient creates a new client for ALBServicesFileUpload resource
func NewALBServicesFileUploadClient(aviSession *session.AviSession) *ALBServicesFileUploadClient {
	return &ALBServicesFileUploadClient{aviSession: aviSession}
}

func (client *ALBServicesFileUploadClient) getAPIPath(uuid string) string {
	path := "api/albservicesfileupload"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ALBServicesFileUpload objects
func (client *ALBServicesFileUploadClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ALBServicesFileUpload, error) {
	var plist []*models.ALBServicesFileUpload
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ALBServicesFileUpload by uuid
func (client *ALBServicesFileUploadClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ALBServicesFileUpload, error) {
	var obj *models.ALBServicesFileUpload
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ALBServicesFileUpload by name
func (client *ALBServicesFileUploadClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ALBServicesFileUpload, error) {
	var obj *models.ALBServicesFileUpload
	err := client.aviSession.GetObjectByName("albservicesfileupload", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ALBServicesFileUpload by filters like name, cloud, tenant
// Api creates ALBServicesFileUpload object with every call.
func (client *ALBServicesFileUploadClient) GetObject(options ...session.ApiOptionsParams) (*models.ALBServicesFileUpload, error) {
	var obj *models.ALBServicesFileUpload
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("albservicesfileupload", newOptions...)
	return obj, err
}

// Create a new ALBServicesFileUpload object
func (client *ALBServicesFileUploadClient) Create(obj *models.ALBServicesFileUpload, options ...session.ApiOptionsParams) (*models.ALBServicesFileUpload, error) {
	var robj *models.ALBServicesFileUpload
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ALBServicesFileUpload object
func (client *ALBServicesFileUploadClient) Update(obj *models.ALBServicesFileUpload, options ...session.ApiOptionsParams) (*models.ALBServicesFileUpload, error) {
	var robj *models.ALBServicesFileUpload
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ALBServicesFileUpload object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ALBServicesFileUpload
// or it should be json compatible of form map[string]interface{}
func (client *ALBServicesFileUploadClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ALBServicesFileUpload, error) {
	var robj *models.ALBServicesFileUpload
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ALBServicesFileUpload object with a given UUID
func (client *ALBServicesFileUploadClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ALBServicesFileUpload object with a given name
func (client *ALBServicesFileUploadClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ALBServicesFileUploadClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
