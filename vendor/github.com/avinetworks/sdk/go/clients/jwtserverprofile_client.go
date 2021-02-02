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

// JWTServerProfileClient is a client for avi JWTServerProfile resource
type JWTServerProfileClient struct {
	aviSession *session.AviSession
}

// NewJWTServerProfileClient creates a new client for JWTServerProfile resource
func NewJWTServerProfileClient(aviSession *session.AviSession) *JWTServerProfileClient {
	return &JWTServerProfileClient{aviSession: aviSession}
}

func (client *JWTServerProfileClient) getAPIPath(uuid string) string {
	path := "api/jwtserverprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of JWTServerProfile objects
func (client *JWTServerProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.JWTServerProfile, error) {
	var plist []*models.JWTServerProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing JWTServerProfile by uuid
func (client *JWTServerProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.JWTServerProfile, error) {
	var obj *models.JWTServerProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing JWTServerProfile by name
func (client *JWTServerProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.JWTServerProfile, error) {
	var obj *models.JWTServerProfile
	err := client.aviSession.GetObjectByName("jwtserverprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing JWTServerProfile by filters like name, cloud, tenant
// Api creates JWTServerProfile object with every call.
func (client *JWTServerProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.JWTServerProfile, error) {
	var obj *models.JWTServerProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("jwtserverprofile", newOptions...)
	return obj, err
}

// Create a new JWTServerProfile object
func (client *JWTServerProfileClient) Create(obj *models.JWTServerProfile, options ...session.ApiOptionsParams) (*models.JWTServerProfile, error) {
	var robj *models.JWTServerProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing JWTServerProfile object
func (client *JWTServerProfileClient) Update(obj *models.JWTServerProfile, options ...session.ApiOptionsParams) (*models.JWTServerProfile, error) {
	var robj *models.JWTServerProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing JWTServerProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.JWTServerProfile
// or it should be json compatible of form map[string]interface{}
func (client *JWTServerProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.JWTServerProfile, error) {
	var robj *models.JWTServerProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing JWTServerProfile object with a given UUID
func (client *JWTServerProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing JWTServerProfile object with a given name
func (client *JWTServerProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *JWTServerProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
