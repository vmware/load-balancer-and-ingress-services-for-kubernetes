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

// UserActivityClient is a client for avi UserActivity resource
type UserActivityClient struct {
	aviSession *session.AviSession
}

// NewUserActivityClient creates a new client for UserActivity resource
func NewUserActivityClient(aviSession *session.AviSession) *UserActivityClient {
	return &UserActivityClient{aviSession: aviSession}
}

func (client *UserActivityClient) getAPIPath(uuid string) string {
	path := "api/useractivity"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of UserActivity objects
func (client *UserActivityClient) GetAll(options ...session.ApiOptionsParams) ([]*models.UserActivity, error) {
	var plist []*models.UserActivity
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing UserActivity by uuid
func (client *UserActivityClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.UserActivity, error) {
	var obj *models.UserActivity
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing UserActivity by name
func (client *UserActivityClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.UserActivity, error) {
	var obj *models.UserActivity
	err := client.aviSession.GetObjectByName("useractivity", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing UserActivity by filters like name, cloud, tenant
// Api creates UserActivity object with every call.
func (client *UserActivityClient) GetObject(options ...session.ApiOptionsParams) (*models.UserActivity, error) {
	var obj *models.UserActivity
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("useractivity", newOptions...)
	return obj, err
}

// Create a new UserActivity object
func (client *UserActivityClient) Create(obj *models.UserActivity, options ...session.ApiOptionsParams) (*models.UserActivity, error) {
	var robj *models.UserActivity
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing UserActivity object
func (client *UserActivityClient) Update(obj *models.UserActivity, options ...session.ApiOptionsParams) (*models.UserActivity, error) {
	var robj *models.UserActivity
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing UserActivity object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.UserActivity
// or it should be json compatible of form map[string]interface{}
func (client *UserActivityClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.UserActivity, error) {
	var robj *models.UserActivity
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing UserActivity object with a given UUID
func (client *UserActivityClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing UserActivity object with a given name
func (client *UserActivityClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *UserActivityClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
