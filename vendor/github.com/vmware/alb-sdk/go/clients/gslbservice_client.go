/***************************************************************************
 * ------------------------------------------------------------------------
 * Copyright 2020 VMware, Inc.  All rights reserved. VMware Confidential
 * ------------------------------------------------------------------------
 */

package clients

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// GslbServiceClient is a client for avi GslbService resource
type GslbServiceClient struct {
	aviSession *session.AviSession
}

// NewGslbServiceClient creates a new client for GslbService resource
func NewGslbServiceClient(aviSession *session.AviSession) *GslbServiceClient {
	return &GslbServiceClient{aviSession: aviSession}
}

func (client *GslbServiceClient) getAPIPath(uuid string) string {
	path := "api/gslbservice"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of GslbService objects
func (client *GslbServiceClient) GetAll(options ...session.ApiOptionsParams) ([]*models.GslbService, error) {
	var plist []*models.GslbService
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing GslbService by uuid
func (client *GslbServiceClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.GslbService, error) {
	var obj *models.GslbService
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing GslbService by name
func (client *GslbServiceClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.GslbService, error) {
	var obj *models.GslbService
	err := client.aviSession.GetObjectByName("gslbservice", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing GslbService by filters like name, cloud, tenant
// Api creates GslbService object with every call.
func (client *GslbServiceClient) GetObject(options ...session.ApiOptionsParams) (*models.GslbService, error) {
	var obj *models.GslbService
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("gslbservice", newOptions...)
	return obj, err
}

// Create a new GslbService object
func (client *GslbServiceClient) Create(obj *models.GslbService, options ...session.ApiOptionsParams) (*models.GslbService, error) {
	var robj *models.GslbService
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing GslbService object
func (client *GslbServiceClient) Update(obj *models.GslbService, options ...session.ApiOptionsParams) (*models.GslbService, error) {
	var robj *models.GslbService
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing GslbService object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.GslbService
// or it should be json compatible of form map[string]interface{}
func (client *GslbServiceClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.GslbService, error) {
	var robj *models.GslbService
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing GslbService object with a given UUID
func (client *GslbServiceClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing GslbService object with a given name
func (client *GslbServiceClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *GslbServiceClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
