// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// GeoDBClient is a client for avi GeoDB resource
type GeoDBClient struct {
	aviSession *session.AviSession
}

// NewGeoDBClient creates a new client for GeoDB resource
func NewGeoDBClient(aviSession *session.AviSession) *GeoDBClient {
	return &GeoDBClient{aviSession: aviSession}
}

func (client *GeoDBClient) getAPIPath(uuid string) string {
	path := "api/geodb"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of GeoDB objects
func (client *GeoDBClient) GetAll(options ...session.ApiOptionsParams) ([]*models.GeoDB, error) {
	var plist []*models.GeoDB
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing GeoDB by uuid
func (client *GeoDBClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.GeoDB, error) {
	var obj *models.GeoDB
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing GeoDB by name
func (client *GeoDBClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.GeoDB, error) {
	var obj *models.GeoDB
	err := client.aviSession.GetObjectByName("geodb", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing GeoDB by filters like name, cloud, tenant
// Api creates GeoDB object with every call.
func (client *GeoDBClient) GetObject(options ...session.ApiOptionsParams) (*models.GeoDB, error) {
	var obj *models.GeoDB
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("geodb", newOptions...)
	return obj, err
}

// Create a new GeoDB object
func (client *GeoDBClient) Create(obj *models.GeoDB, options ...session.ApiOptionsParams) (*models.GeoDB, error) {
	var robj *models.GeoDB
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing GeoDB object
func (client *GeoDBClient) Update(obj *models.GeoDB, options ...session.ApiOptionsParams) (*models.GeoDB, error) {
	var robj *models.GeoDB
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing GeoDB object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.GeoDB
// or it should be json compatible of form map[string]interface{}
func (client *GeoDBClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.GeoDB, error) {
	var robj *models.GeoDB
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing GeoDB object with a given UUID
func (client *GeoDBClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing GeoDB object with a given name
func (client *GeoDBClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *GeoDBClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
