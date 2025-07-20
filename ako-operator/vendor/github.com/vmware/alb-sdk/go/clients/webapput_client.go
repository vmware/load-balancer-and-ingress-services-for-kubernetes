// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// WebappUTClient is a client for avi WebappUT resource
type WebappUTClient struct {
	aviSession *session.AviSession
}

// NewWebappUTClient creates a new client for WebappUT resource
func NewWebappUTClient(aviSession *session.AviSession) *WebappUTClient {
	return &WebappUTClient{aviSession: aviSession}
}

func (client *WebappUTClient) getAPIPath(uuid string) string {
	path := "api/webapput"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of WebappUT objects
func (client *WebappUTClient) GetAll(options ...session.ApiOptionsParams) ([]*models.WebappUT, error) {
	var plist []*models.WebappUT
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing WebappUT by uuid
func (client *WebappUTClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.WebappUT, error) {
	var obj *models.WebappUT
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing WebappUT by name
func (client *WebappUTClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.WebappUT, error) {
	var obj *models.WebappUT
	err := client.aviSession.GetObjectByName("webapput", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing WebappUT by filters like name, cloud, tenant
// Api creates WebappUT object with every call.
func (client *WebappUTClient) GetObject(options ...session.ApiOptionsParams) (*models.WebappUT, error) {
	var obj *models.WebappUT
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("webapput", newOptions...)
	return obj, err
}

// Create a new WebappUT object
func (client *WebappUTClient) Create(obj *models.WebappUT, options ...session.ApiOptionsParams) (*models.WebappUT, error) {
	var robj *models.WebappUT
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing WebappUT object
func (client *WebappUTClient) Update(obj *models.WebappUT, options ...session.ApiOptionsParams) (*models.WebappUT, error) {
	var robj *models.WebappUT
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing WebappUT object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.WebappUT
// or it should be json compatible of form map[string]interface{}
func (client *WebappUTClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.WebappUT, error) {
	var robj *models.WebappUT
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing WebappUT object with a given UUID
func (client *WebappUTClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing WebappUT object with a given name
func (client *WebappUTClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *WebappUTClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
