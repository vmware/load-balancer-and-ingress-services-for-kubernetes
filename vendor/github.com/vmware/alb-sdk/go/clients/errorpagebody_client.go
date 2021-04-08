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

// ErrorPageBodyClient is a client for avi ErrorPageBody resource
type ErrorPageBodyClient struct {
	aviSession *session.AviSession
}

// NewErrorPageBodyClient creates a new client for ErrorPageBody resource
func NewErrorPageBodyClient(aviSession *session.AviSession) *ErrorPageBodyClient {
	return &ErrorPageBodyClient{aviSession: aviSession}
}

func (client *ErrorPageBodyClient) getAPIPath(uuid string) string {
	path := "api/errorpagebody"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ErrorPageBody objects
func (client *ErrorPageBodyClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ErrorPageBody, error) {
	var plist []*models.ErrorPageBody
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ErrorPageBody by uuid
func (client *ErrorPageBodyClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ErrorPageBody, error) {
	var obj *models.ErrorPageBody
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ErrorPageBody by name
func (client *ErrorPageBodyClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ErrorPageBody, error) {
	var obj *models.ErrorPageBody
	err := client.aviSession.GetObjectByName("errorpagebody", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ErrorPageBody by filters like name, cloud, tenant
// Api creates ErrorPageBody object with every call.
func (client *ErrorPageBodyClient) GetObject(options ...session.ApiOptionsParams) (*models.ErrorPageBody, error) {
	var obj *models.ErrorPageBody
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("errorpagebody", newOptions...)
	return obj, err
}

// Create a new ErrorPageBody object
func (client *ErrorPageBodyClient) Create(obj *models.ErrorPageBody, options ...session.ApiOptionsParams) (*models.ErrorPageBody, error) {
	var robj *models.ErrorPageBody
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ErrorPageBody object
func (client *ErrorPageBodyClient) Update(obj *models.ErrorPageBody, options ...session.ApiOptionsParams) (*models.ErrorPageBody, error) {
	var robj *models.ErrorPageBody
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ErrorPageBody object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ErrorPageBody
// or it should be json compatible of form map[string]interface{}
func (client *ErrorPageBodyClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ErrorPageBody, error) {
	var robj *models.ErrorPageBody
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ErrorPageBody object with a given UUID
func (client *ErrorPageBodyClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ErrorPageBody object with a given name
func (client *ErrorPageBodyClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ErrorPageBodyClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
