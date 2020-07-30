/***************************************************************************
 * ------------------------------------------------------------------------
 * Copyright 2020 VMware, Inc.  All rights reserved. VMware Confidential
 * ------------------------------------------------------------------------
 */

package clients

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

import (
	"github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
)

// NsxtSegmentRuntimeClient is a client for avi NsxtSegmentRuntime resource
type NsxtSegmentRuntimeClient struct {
	aviSession *session.AviSession
}

// NewNsxtSegmentRuntimeClient creates a new client for NsxtSegmentRuntime resource
func NewNsxtSegmentRuntimeClient(aviSession *session.AviSession) *NsxtSegmentRuntimeClient {
	return &NsxtSegmentRuntimeClient{aviSession: aviSession}
}

func (client *NsxtSegmentRuntimeClient) getAPIPath(uuid string) string {
	path := "api/nsxtsegmentruntime"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of NsxtSegmentRuntime objects
func (client *NsxtSegmentRuntimeClient) GetAll(options ...session.ApiOptionsParams) ([]*models.NsxtSegmentRuntime, error) {
	var plist []*models.NsxtSegmentRuntime
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing NsxtSegmentRuntime by uuid
func (client *NsxtSegmentRuntimeClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.NsxtSegmentRuntime, error) {
	var obj *models.NsxtSegmentRuntime
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing NsxtSegmentRuntime by name
func (client *NsxtSegmentRuntimeClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.NsxtSegmentRuntime, error) {
	var obj *models.NsxtSegmentRuntime
	err := client.aviSession.GetObjectByName("nsxtsegmentruntime", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing NsxtSegmentRuntime by filters like name, cloud, tenant
// Api creates NsxtSegmentRuntime object with every call.
func (client *NsxtSegmentRuntimeClient) GetObject(options ...session.ApiOptionsParams) (*models.NsxtSegmentRuntime, error) {
	var obj *models.NsxtSegmentRuntime
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("nsxtsegmentruntime", newOptions...)
	return obj, err
}

// Create a new NsxtSegmentRuntime object
func (client *NsxtSegmentRuntimeClient) Create(obj *models.NsxtSegmentRuntime, options ...session.ApiOptionsParams) (*models.NsxtSegmentRuntime, error) {
	var robj *models.NsxtSegmentRuntime
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing NsxtSegmentRuntime object
func (client *NsxtSegmentRuntimeClient) Update(obj *models.NsxtSegmentRuntime, options ...session.ApiOptionsParams) (*models.NsxtSegmentRuntime, error) {
	var robj *models.NsxtSegmentRuntime
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing NsxtSegmentRuntime object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.NsxtSegmentRuntime
// or it should be json compatible of form map[string]interface{}
func (client *NsxtSegmentRuntimeClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.NsxtSegmentRuntime, error) {
	var robj *models.NsxtSegmentRuntime
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing NsxtSegmentRuntime object with a given UUID
func (client *NsxtSegmentRuntimeClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing NsxtSegmentRuntime object with a given name
func (client *NsxtSegmentRuntimeClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *NsxtSegmentRuntimeClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
