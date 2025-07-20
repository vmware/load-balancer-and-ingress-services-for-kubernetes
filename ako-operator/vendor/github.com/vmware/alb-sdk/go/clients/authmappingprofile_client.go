// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// AuthMappingProfileClient is a client for avi AuthMappingProfile resource
type AuthMappingProfileClient struct {
	aviSession *session.AviSession
}

// NewAuthMappingProfileClient creates a new client for AuthMappingProfile resource
func NewAuthMappingProfileClient(aviSession *session.AviSession) *AuthMappingProfileClient {
	return &AuthMappingProfileClient{aviSession: aviSession}
}

func (client *AuthMappingProfileClient) getAPIPath(uuid string) string {
	path := "api/authmappingprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of AuthMappingProfile objects
func (client *AuthMappingProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.AuthMappingProfile, error) {
	var plist []*models.AuthMappingProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing AuthMappingProfile by uuid
func (client *AuthMappingProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.AuthMappingProfile, error) {
	var obj *models.AuthMappingProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing AuthMappingProfile by name
func (client *AuthMappingProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.AuthMappingProfile, error) {
	var obj *models.AuthMappingProfile
	err := client.aviSession.GetObjectByName("authmappingprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing AuthMappingProfile by filters like name, cloud, tenant
// Api creates AuthMappingProfile object with every call.
func (client *AuthMappingProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.AuthMappingProfile, error) {
	var obj *models.AuthMappingProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("authmappingprofile", newOptions...)
	return obj, err
}

// Create a new AuthMappingProfile object
func (client *AuthMappingProfileClient) Create(obj *models.AuthMappingProfile, options ...session.ApiOptionsParams) (*models.AuthMappingProfile, error) {
	var robj *models.AuthMappingProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing AuthMappingProfile object
func (client *AuthMappingProfileClient) Update(obj *models.AuthMappingProfile, options ...session.ApiOptionsParams) (*models.AuthMappingProfile, error) {
	var robj *models.AuthMappingProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing AuthMappingProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.AuthMappingProfile
// or it should be json compatible of form map[string]interface{}
func (client *AuthMappingProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.AuthMappingProfile, error) {
	var robj *models.AuthMappingProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing AuthMappingProfile object with a given UUID
func (client *AuthMappingProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing AuthMappingProfile object with a given name
func (client *AuthMappingProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *AuthMappingProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
