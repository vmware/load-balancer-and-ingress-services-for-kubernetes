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

// ObjectAccessPolicyClient is a client for avi ObjectAccessPolicy resource
type ObjectAccessPolicyClient struct {
	aviSession *session.AviSession
}

// NewObjectAccessPolicyClient creates a new client for ObjectAccessPolicy resource
func NewObjectAccessPolicyClient(aviSession *session.AviSession) *ObjectAccessPolicyClient {
	return &ObjectAccessPolicyClient{aviSession: aviSession}
}

func (client *ObjectAccessPolicyClient) getAPIPath(uuid string) string {
	path := "api/objectaccesspolicy"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ObjectAccessPolicy objects
func (client *ObjectAccessPolicyClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ObjectAccessPolicy, error) {
	var plist []*models.ObjectAccessPolicy
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ObjectAccessPolicy by uuid
func (client *ObjectAccessPolicyClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ObjectAccessPolicy, error) {
	var obj *models.ObjectAccessPolicy
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ObjectAccessPolicy by name
func (client *ObjectAccessPolicyClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ObjectAccessPolicy, error) {
	var obj *models.ObjectAccessPolicy
	err := client.aviSession.GetObjectByName("objectaccesspolicy", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ObjectAccessPolicy by filters like name, cloud, tenant
// Api creates ObjectAccessPolicy object with every call.
func (client *ObjectAccessPolicyClient) GetObject(options ...session.ApiOptionsParams) (*models.ObjectAccessPolicy, error) {
	var obj *models.ObjectAccessPolicy
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("objectaccesspolicy", newOptions...)
	return obj, err
}

// Create a new ObjectAccessPolicy object
func (client *ObjectAccessPolicyClient) Create(obj *models.ObjectAccessPolicy, options ...session.ApiOptionsParams) (*models.ObjectAccessPolicy, error) {
	var robj *models.ObjectAccessPolicy
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ObjectAccessPolicy object
func (client *ObjectAccessPolicyClient) Update(obj *models.ObjectAccessPolicy, options ...session.ApiOptionsParams) (*models.ObjectAccessPolicy, error) {
	var robj *models.ObjectAccessPolicy
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ObjectAccessPolicy object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ObjectAccessPolicy
// or it should be json compatible of form map[string]interface{}
func (client *ObjectAccessPolicyClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ObjectAccessPolicy, error) {
	var robj *models.ObjectAccessPolicy
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ObjectAccessPolicy object with a given UUID
func (client *ObjectAccessPolicyClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ObjectAccessPolicy object with a given name
func (client *ObjectAccessPolicyClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ObjectAccessPolicyClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
