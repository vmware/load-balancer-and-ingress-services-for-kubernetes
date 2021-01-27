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

// UserClient is a client for avi User resource
type UserClient struct {
	aviSession *session.AviSession
}

// NewUserClient creates a new client for User resource
func NewUserClient(aviSession *session.AviSession) *UserClient {
	return &UserClient{aviSession: aviSession}
}

func (client *UserClient) getAPIPath(uuid string) string {
	path := "api/user"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of User objects
func (client *UserClient) GetAll(options ...session.ApiOptionsParams) ([]*models.User, error) {
	var plist []*models.User
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing User by uuid
func (client *UserClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.User, error) {
	var obj *models.User
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing User by name
func (client *UserClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.User, error) {
	var obj *models.User
	err := client.aviSession.GetObjectByName("user", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing User by filters like name, cloud, tenant
// Api creates User object with every call.
func (client *UserClient) GetObject(options ...session.ApiOptionsParams) (*models.User, error) {
	var obj *models.User
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("user", newOptions...)
	return obj, err
}

// Create a new User object
func (client *UserClient) Create(obj *models.User, options ...session.ApiOptionsParams) (*models.User, error) {
	var robj *models.User
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing User object
func (client *UserClient) Update(obj *models.User, options ...session.ApiOptionsParams) (*models.User, error) {
	var robj *models.User
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing User object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.User
// or it should be json compatible of form map[string]interface{}
func (client *UserClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.User, error) {
	var robj *models.User
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing User object with a given UUID
func (client *UserClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing User object with a given name
func (client *UserClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *UserClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
