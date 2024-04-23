// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// GslbGeoDbProfileClient is a client for avi GslbGeoDbProfile resource
type GslbGeoDbProfileClient struct {
	aviSession *session.AviSession
}

// NewGslbGeoDbProfileClient creates a new client for GslbGeoDbProfile resource
func NewGslbGeoDbProfileClient(aviSession *session.AviSession) *GslbGeoDbProfileClient {
	return &GslbGeoDbProfileClient{aviSession: aviSession}
}

func (client *GslbGeoDbProfileClient) getAPIPath(uuid string) string {
	path := "api/gslbgeodbprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of GslbGeoDbProfile objects
func (client *GslbGeoDbProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.GslbGeoDbProfile, error) {
	var plist []*models.GslbGeoDbProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing GslbGeoDbProfile by uuid
func (client *GslbGeoDbProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.GslbGeoDbProfile, error) {
	var obj *models.GslbGeoDbProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing GslbGeoDbProfile by name
func (client *GslbGeoDbProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.GslbGeoDbProfile, error) {
	var obj *models.GslbGeoDbProfile
	err := client.aviSession.GetObjectByName("gslbgeodbprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing GslbGeoDbProfile by filters like name, cloud, tenant
// Api creates GslbGeoDbProfile object with every call.
func (client *GslbGeoDbProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.GslbGeoDbProfile, error) {
	var obj *models.GslbGeoDbProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("gslbgeodbprofile", newOptions...)
	return obj, err
}

// Create a new GslbGeoDbProfile object
func (client *GslbGeoDbProfileClient) Create(obj *models.GslbGeoDbProfile, options ...session.ApiOptionsParams) (*models.GslbGeoDbProfile, error) {
	var robj *models.GslbGeoDbProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing GslbGeoDbProfile object
func (client *GslbGeoDbProfileClient) Update(obj *models.GslbGeoDbProfile, options ...session.ApiOptionsParams) (*models.GslbGeoDbProfile, error) {
	var robj *models.GslbGeoDbProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing GslbGeoDbProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.GslbGeoDbProfile
// or it should be json compatible of form map[string]interface{}
func (client *GslbGeoDbProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.GslbGeoDbProfile, error) {
	var robj *models.GslbGeoDbProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing GslbGeoDbProfile object with a given UUID
func (client *GslbGeoDbProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing GslbGeoDbProfile object with a given name
func (client *GslbGeoDbProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *GslbGeoDbProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
