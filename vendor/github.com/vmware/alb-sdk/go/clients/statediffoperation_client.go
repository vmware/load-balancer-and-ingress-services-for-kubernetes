// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// StatediffOperationClient is a client for avi StatediffOperation resource
type StatediffOperationClient struct {
	aviSession *session.AviSession
}

// NewStatediffOperationClient creates a new client for StatediffOperation resource
func NewStatediffOperationClient(aviSession *session.AviSession) *StatediffOperationClient {
	return &StatediffOperationClient{aviSession: aviSession}
}

func (client *StatediffOperationClient) getAPIPath(uuid string) string {
	path := "api/statediffoperation"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of StatediffOperation objects
func (client *StatediffOperationClient) GetAll(options ...session.ApiOptionsParams) ([]*models.StatediffOperation, error) {
	var plist []*models.StatediffOperation
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing StatediffOperation by uuid
func (client *StatediffOperationClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.StatediffOperation, error) {
	var obj *models.StatediffOperation
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing StatediffOperation by name
func (client *StatediffOperationClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.StatediffOperation, error) {
	var obj *models.StatediffOperation
	err := client.aviSession.GetObjectByName("statediffoperation", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing StatediffOperation by filters like name, cloud, tenant
// Api creates StatediffOperation object with every call.
func (client *StatediffOperationClient) GetObject(options ...session.ApiOptionsParams) (*models.StatediffOperation, error) {
	var obj *models.StatediffOperation
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("statediffoperation", newOptions...)
	return obj, err
}

// Create a new StatediffOperation object
func (client *StatediffOperationClient) Create(obj *models.StatediffOperation, options ...session.ApiOptionsParams) (*models.StatediffOperation, error) {
	var robj *models.StatediffOperation
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing StatediffOperation object
func (client *StatediffOperationClient) Update(obj *models.StatediffOperation, options ...session.ApiOptionsParams) (*models.StatediffOperation, error) {
	var robj *models.StatediffOperation
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing StatediffOperation object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.StatediffOperation
// or it should be json compatible of form map[string]interface{}
func (client *StatediffOperationClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.StatediffOperation, error) {
	var robj *models.StatediffOperation
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing StatediffOperation object with a given UUID
func (client *StatediffOperationClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing StatediffOperation object with a given name
func (client *StatediffOperationClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *StatediffOperationClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
