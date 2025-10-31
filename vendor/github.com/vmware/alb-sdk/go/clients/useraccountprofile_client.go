// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// UserAccountProfileClient is a client for avi UserAccountProfile resource
type UserAccountProfileClient struct {
	aviSession *session.AviSession
}

// NewUserAccountProfileClient creates a new client for UserAccountProfile resource
func NewUserAccountProfileClient(aviSession *session.AviSession) *UserAccountProfileClient {
	return &UserAccountProfileClient{aviSession: aviSession}
}

func (client *UserAccountProfileClient) getAPIPath(uuid string) string {
	path := "api/useraccountprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of UserAccountProfile objects
func (client *UserAccountProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.UserAccountProfile, error) {
	var plist []*models.UserAccountProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing UserAccountProfile by uuid
func (client *UserAccountProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.UserAccountProfile, error) {
	var obj *models.UserAccountProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing UserAccountProfile by name
func (client *UserAccountProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.UserAccountProfile, error) {
	var obj *models.UserAccountProfile
	err := client.aviSession.GetObjectByName("useraccountprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing UserAccountProfile by filters like name, cloud, tenant
// Api creates UserAccountProfile object with every call.
func (client *UserAccountProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.UserAccountProfile, error) {
	var obj *models.UserAccountProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("useraccountprofile", newOptions...)
	return obj, err
}

// Create a new UserAccountProfile object
func (client *UserAccountProfileClient) Create(obj *models.UserAccountProfile, options ...session.ApiOptionsParams) (*models.UserAccountProfile, error) {
	var robj *models.UserAccountProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing UserAccountProfile object
func (client *UserAccountProfileClient) Update(obj *models.UserAccountProfile, options ...session.ApiOptionsParams) (*models.UserAccountProfile, error) {
	var robj *models.UserAccountProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing UserAccountProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.UserAccountProfile
// or it should be json compatible of form map[string]interface{}
func (client *UserAccountProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.UserAccountProfile, error) {
	var robj *models.UserAccountProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing UserAccountProfile object with a given UUID
func (client *UserAccountProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing UserAccountProfile object with a given name
func (client *UserAccountProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *UserAccountProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
