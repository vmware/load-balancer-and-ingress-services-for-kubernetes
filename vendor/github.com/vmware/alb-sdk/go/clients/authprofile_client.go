// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// AuthProfileClient is a client for avi AuthProfile resource
type AuthProfileClient struct {
	aviSession *session.AviSession
}

// NewAuthProfileClient creates a new client for AuthProfile resource
func NewAuthProfileClient(aviSession *session.AviSession) *AuthProfileClient {
	return &AuthProfileClient{aviSession: aviSession}
}

func (client *AuthProfileClient) getAPIPath(uuid string) string {
	path := "api/authprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of AuthProfile objects
func (client *AuthProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.AuthProfile, error) {
	var plist []*models.AuthProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing AuthProfile by uuid
func (client *AuthProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.AuthProfile, error) {
	var obj *models.AuthProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing AuthProfile by name
func (client *AuthProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.AuthProfile, error) {
	var obj *models.AuthProfile
	err := client.aviSession.GetObjectByName("authprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing AuthProfile by filters like name, cloud, tenant
// Api creates AuthProfile object with every call.
func (client *AuthProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.AuthProfile, error) {
	var obj *models.AuthProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("authprofile", newOptions...)
	return obj, err
}

// Create a new AuthProfile object
func (client *AuthProfileClient) Create(obj *models.AuthProfile, options ...session.ApiOptionsParams) (*models.AuthProfile, error) {
	var robj *models.AuthProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing AuthProfile object
func (client *AuthProfileClient) Update(obj *models.AuthProfile, options ...session.ApiOptionsParams) (*models.AuthProfile, error) {
	var robj *models.AuthProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing AuthProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.AuthProfile
// or it should be json compatible of form map[string]interface{}
func (client *AuthProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.AuthProfile, error) {
	var robj *models.AuthProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing AuthProfile object with a given UUID
func (client *AuthProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing AuthProfile object with a given name
func (client *AuthProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *AuthProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
