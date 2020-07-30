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

// FileObjectClient is a client for avi FileObject resource
type FileObjectClient struct {
	aviSession *session.AviSession
}

// NewFileObjectClient creates a new client for FileObject resource
func NewFileObjectClient(aviSession *session.AviSession) *FileObjectClient {
	return &FileObjectClient{aviSession: aviSession}
}

func (client *FileObjectClient) getAPIPath(uuid string) string {
	path := "api/fileobject"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of FileObject objects
func (client *FileObjectClient) GetAll(options ...session.ApiOptionsParams) ([]*models.FileObject, error) {
	var plist []*models.FileObject
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing FileObject by uuid
func (client *FileObjectClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.FileObject, error) {
	var obj *models.FileObject
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing FileObject by name
func (client *FileObjectClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.FileObject, error) {
	var obj *models.FileObject
	err := client.aviSession.GetObjectByName("fileobject", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing FileObject by filters like name, cloud, tenant
// Api creates FileObject object with every call.
func (client *FileObjectClient) GetObject(options ...session.ApiOptionsParams) (*models.FileObject, error) {
	var obj *models.FileObject
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("fileobject", newOptions...)
	return obj, err
}

// Create a new FileObject object
func (client *FileObjectClient) Create(obj *models.FileObject, options ...session.ApiOptionsParams) (*models.FileObject, error) {
	var robj *models.FileObject
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing FileObject object
func (client *FileObjectClient) Update(obj *models.FileObject, options ...session.ApiOptionsParams) (*models.FileObject, error) {
	var robj *models.FileObject
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing FileObject object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.FileObject
// or it should be json compatible of form map[string]interface{}
func (client *FileObjectClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.FileObject, error) {
	var robj *models.FileObject
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing FileObject object with a given UUID
func (client *FileObjectClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing FileObject object with a given name
func (client *FileObjectClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *FileObjectClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
