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

// JWTProfileClient is a client for avi JWTProfile resource
type JWTProfileClient struct {
	aviSession *session.AviSession
}

// NewJWTProfileClient creates a new client for JWTProfile resource
func NewJWTProfileClient(aviSession *session.AviSession) *JWTProfileClient {
	return &JWTProfileClient{aviSession: aviSession}
}

func (client *JWTProfileClient) getAPIPath(uuid string) string {
	path := "api/jwtprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of JWTProfile objects
func (client *JWTProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.JWTProfile, error) {
	var plist []*models.JWTProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing JWTProfile by uuid
func (client *JWTProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.JWTProfile, error) {
	var obj *models.JWTProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing JWTProfile by name
func (client *JWTProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.JWTProfile, error) {
	var obj *models.JWTProfile
	err := client.aviSession.GetObjectByName("jwtprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing JWTProfile by filters like name, cloud, tenant
// Api creates JWTProfile object with every call.
func (client *JWTProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.JWTProfile, error) {
	var obj *models.JWTProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("jwtprofile", newOptions...)
	return obj, err
}

// Create a new JWTProfile object
func (client *JWTProfileClient) Create(obj *models.JWTProfile, options ...session.ApiOptionsParams) (*models.JWTProfile, error) {
	var robj *models.JWTProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing JWTProfile object
func (client *JWTProfileClient) Update(obj *models.JWTProfile, options ...session.ApiOptionsParams) (*models.JWTProfile, error) {
	var robj *models.JWTProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing JWTProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.JWTProfile
// or it should be json compatible of form map[string]interface{}
func (client *JWTProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.JWTProfile, error) {
	var robj *models.JWTProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing JWTProfile object with a given UUID
func (client *JWTProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing JWTProfile object with a given name
func (client *JWTProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *JWTProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
