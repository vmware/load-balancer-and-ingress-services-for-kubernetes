// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// LabelGroupClient is a client for avi LabelGroup resource
type LabelGroupClient struct {
	aviSession *session.AviSession
}

// NewLabelGroupClient creates a new client for LabelGroup resource
func NewLabelGroupClient(aviSession *session.AviSession) *LabelGroupClient {
	return &LabelGroupClient{aviSession: aviSession}
}

func (client *LabelGroupClient) getAPIPath(uuid string) string {
	path := "api/labelgroup"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of LabelGroup objects
func (client *LabelGroupClient) GetAll(options ...session.ApiOptionsParams) ([]*models.LabelGroup, error) {
	var plist []*models.LabelGroup
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing LabelGroup by uuid
func (client *LabelGroupClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.LabelGroup, error) {
	var obj *models.LabelGroup
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing LabelGroup by name
func (client *LabelGroupClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.LabelGroup, error) {
	var obj *models.LabelGroup
	err := client.aviSession.GetObjectByName("labelgroup", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing LabelGroup by filters like name, cloud, tenant
// Api creates LabelGroup object with every call.
func (client *LabelGroupClient) GetObject(options ...session.ApiOptionsParams) (*models.LabelGroup, error) {
	var obj *models.LabelGroup
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("labelgroup", newOptions...)
	return obj, err
}

// Create a new LabelGroup object
func (client *LabelGroupClient) Create(obj *models.LabelGroup, options ...session.ApiOptionsParams) (*models.LabelGroup, error) {
	var robj *models.LabelGroup
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing LabelGroup object
func (client *LabelGroupClient) Update(obj *models.LabelGroup, options ...session.ApiOptionsParams) (*models.LabelGroup, error) {
	var robj *models.LabelGroup
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing LabelGroup object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.LabelGroup
// or it should be json compatible of form map[string]interface{}
func (client *LabelGroupClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.LabelGroup, error) {
	var robj *models.LabelGroup
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing LabelGroup object with a given UUID
func (client *LabelGroupClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing LabelGroup object with a given name
func (client *LabelGroupClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *LabelGroupClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
