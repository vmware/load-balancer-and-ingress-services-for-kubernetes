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

// PriorityLabelsClient is a client for avi PriorityLabels resource
type PriorityLabelsClient struct {
	aviSession *session.AviSession
}

// NewPriorityLabelsClient creates a new client for PriorityLabels resource
func NewPriorityLabelsClient(aviSession *session.AviSession) *PriorityLabelsClient {
	return &PriorityLabelsClient{aviSession: aviSession}
}

func (client *PriorityLabelsClient) getAPIPath(uuid string) string {
	path := "api/prioritylabels"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of PriorityLabels objects
func (client *PriorityLabelsClient) GetAll(options ...session.ApiOptionsParams) ([]*models.PriorityLabels, error) {
	var plist []*models.PriorityLabels
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing PriorityLabels by uuid
func (client *PriorityLabelsClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.PriorityLabels, error) {
	var obj *models.PriorityLabels
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing PriorityLabels by name
func (client *PriorityLabelsClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.PriorityLabels, error) {
	var obj *models.PriorityLabels
	err := client.aviSession.GetObjectByName("prioritylabels", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing PriorityLabels by filters like name, cloud, tenant
// Api creates PriorityLabels object with every call.
func (client *PriorityLabelsClient) GetObject(options ...session.ApiOptionsParams) (*models.PriorityLabels, error) {
	var obj *models.PriorityLabels
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("prioritylabels", newOptions...)
	return obj, err
}

// Create a new PriorityLabels object
func (client *PriorityLabelsClient) Create(obj *models.PriorityLabels, options ...session.ApiOptionsParams) (*models.PriorityLabels, error) {
	var robj *models.PriorityLabels
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing PriorityLabels object
func (client *PriorityLabelsClient) Update(obj *models.PriorityLabels, options ...session.ApiOptionsParams) (*models.PriorityLabels, error) {
	var robj *models.PriorityLabels
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing PriorityLabels object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.PriorityLabels
// or it should be json compatible of form map[string]interface{}
func (client *PriorityLabelsClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.PriorityLabels, error) {
	var robj *models.PriorityLabels
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing PriorityLabels object with a given UUID
func (client *PriorityLabelsClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing PriorityLabels object with a given name
func (client *PriorityLabelsClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *PriorityLabelsClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
