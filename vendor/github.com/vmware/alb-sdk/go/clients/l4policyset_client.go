// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// L4PolicySetClient is a client for avi L4PolicySet resource
type L4PolicySetClient struct {
	aviSession *session.AviSession
}

// NewL4PolicySetClient creates a new client for L4PolicySet resource
func NewL4PolicySetClient(aviSession *session.AviSession) *L4PolicySetClient {
	return &L4PolicySetClient{aviSession: aviSession}
}

func (client *L4PolicySetClient) getAPIPath(uuid string) string {
	path := "api/l4policyset"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of L4PolicySet objects
func (client *L4PolicySetClient) GetAll(options ...session.ApiOptionsParams) ([]*models.L4PolicySet, error) {
	var plist []*models.L4PolicySet
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing L4PolicySet by uuid
func (client *L4PolicySetClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.L4PolicySet, error) {
	var obj *models.L4PolicySet
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing L4PolicySet by name
func (client *L4PolicySetClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.L4PolicySet, error) {
	var obj *models.L4PolicySet
	err := client.aviSession.GetObjectByName("l4policyset", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing L4PolicySet by filters like name, cloud, tenant
// Api creates L4PolicySet object with every call.
func (client *L4PolicySetClient) GetObject(options ...session.ApiOptionsParams) (*models.L4PolicySet, error) {
	var obj *models.L4PolicySet
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("l4policyset", newOptions...)
	return obj, err
}

// Create a new L4PolicySet object
func (client *L4PolicySetClient) Create(obj *models.L4PolicySet, options ...session.ApiOptionsParams) (*models.L4PolicySet, error) {
	var robj *models.L4PolicySet
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing L4PolicySet object
func (client *L4PolicySetClient) Update(obj *models.L4PolicySet, options ...session.ApiOptionsParams) (*models.L4PolicySet, error) {
	var robj *models.L4PolicySet
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing L4PolicySet object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.L4PolicySet
// or it should be json compatible of form map[string]interface{}
func (client *L4PolicySetClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.L4PolicySet, error) {
	var robj *models.L4PolicySet
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing L4PolicySet object with a given UUID
func (client *L4PolicySetClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing L4PolicySet object with a given name
func (client *L4PolicySetClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *L4PolicySetClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
