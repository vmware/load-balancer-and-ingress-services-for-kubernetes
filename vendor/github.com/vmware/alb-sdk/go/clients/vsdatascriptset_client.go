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

// VSDataScriptSetClient is a client for avi VSDataScriptSet resource
type VSDataScriptSetClient struct {
	aviSession *session.AviSession
}

// NewVSDataScriptSetClient creates a new client for VSDataScriptSet resource
func NewVSDataScriptSetClient(aviSession *session.AviSession) *VSDataScriptSetClient {
	return &VSDataScriptSetClient{aviSession: aviSession}
}

func (client *VSDataScriptSetClient) getAPIPath(uuid string) string {
	path := "api/vsdatascriptset"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VSDataScriptSet objects
func (client *VSDataScriptSetClient) GetAll(options ...session.ApiOptionsParams) ([]*models.VSDataScriptSet, error) {
	var plist []*models.VSDataScriptSet
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing VSDataScriptSet by uuid
func (client *VSDataScriptSetClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.VSDataScriptSet, error) {
	var obj *models.VSDataScriptSet
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing VSDataScriptSet by name
func (client *VSDataScriptSetClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.VSDataScriptSet, error) {
	var obj *models.VSDataScriptSet
	err := client.aviSession.GetObjectByName("vsdatascriptset", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing VSDataScriptSet by filters like name, cloud, tenant
// Api creates VSDataScriptSet object with every call.
func (client *VSDataScriptSetClient) GetObject(options ...session.ApiOptionsParams) (*models.VSDataScriptSet, error) {
	var obj *models.VSDataScriptSet
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vsdatascriptset", newOptions...)
	return obj, err
}

// Create a new VSDataScriptSet object
func (client *VSDataScriptSetClient) Create(obj *models.VSDataScriptSet, options ...session.ApiOptionsParams) (*models.VSDataScriptSet, error) {
	var robj *models.VSDataScriptSet
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing VSDataScriptSet object
func (client *VSDataScriptSetClient) Update(obj *models.VSDataScriptSet, options ...session.ApiOptionsParams) (*models.VSDataScriptSet, error) {
	var robj *models.VSDataScriptSet
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing VSDataScriptSet object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VSDataScriptSet
// or it should be json compatible of form map[string]interface{}
func (client *VSDataScriptSetClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.VSDataScriptSet, error) {
	var robj *models.VSDataScriptSet
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing VSDataScriptSet object with a given UUID
func (client *VSDataScriptSetClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing VSDataScriptSet object with a given name
func (client *VSDataScriptSetClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *VSDataScriptSetClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
