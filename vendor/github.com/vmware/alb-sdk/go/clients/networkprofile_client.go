// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// NetworkProfileClient is a client for avi NetworkProfile resource
type NetworkProfileClient struct {
	aviSession *session.AviSession
}

// NewNetworkProfileClient creates a new client for NetworkProfile resource
func NewNetworkProfileClient(aviSession *session.AviSession) *NetworkProfileClient {
	return &NetworkProfileClient{aviSession: aviSession}
}

func (client *NetworkProfileClient) getAPIPath(uuid string) string {
	path := "api/networkprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of NetworkProfile objects
func (client *NetworkProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.NetworkProfile, error) {
	var plist []*models.NetworkProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing NetworkProfile by uuid
func (client *NetworkProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.NetworkProfile, error) {
	var obj *models.NetworkProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing NetworkProfile by name
func (client *NetworkProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.NetworkProfile, error) {
	var obj *models.NetworkProfile
	err := client.aviSession.GetObjectByName("networkprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing NetworkProfile by filters like name, cloud, tenant
// Api creates NetworkProfile object with every call.
func (client *NetworkProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.NetworkProfile, error) {
	var obj *models.NetworkProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("networkprofile", newOptions...)
	return obj, err
}

// Create a new NetworkProfile object
func (client *NetworkProfileClient) Create(obj *models.NetworkProfile, options ...session.ApiOptionsParams) (*models.NetworkProfile, error) {
	var robj *models.NetworkProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing NetworkProfile object
func (client *NetworkProfileClient) Update(obj *models.NetworkProfile, options ...session.ApiOptionsParams) (*models.NetworkProfile, error) {
	var robj *models.NetworkProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing NetworkProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.NetworkProfile
// or it should be json compatible of form map[string]interface{}
func (client *NetworkProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.NetworkProfile, error) {
	var robj *models.NetworkProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing NetworkProfile object with a given UUID
func (client *NetworkProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing NetworkProfile object with a given name
func (client *NetworkProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *NetworkProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
