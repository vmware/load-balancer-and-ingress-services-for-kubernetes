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

// VCenterServerClient is a client for avi VCenterServer resource
type VCenterServerClient struct {
	aviSession *session.AviSession
}

// NewVCenterServerClient creates a new client for VCenterServer resource
func NewVCenterServerClient(aviSession *session.AviSession) *VCenterServerClient {
	return &VCenterServerClient{aviSession: aviSession}
}

func (client *VCenterServerClient) getAPIPath(uuid string) string {
	path := "api/vcenterserver"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VCenterServer objects
func (client *VCenterServerClient) GetAll(options ...session.ApiOptionsParams) ([]*models.VCenterServer, error) {
	var plist []*models.VCenterServer
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing VCenterServer by uuid
func (client *VCenterServerClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.VCenterServer, error) {
	var obj *models.VCenterServer
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing VCenterServer by name
func (client *VCenterServerClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.VCenterServer, error) {
	var obj *models.VCenterServer
	err := client.aviSession.GetObjectByName("vcenterserver", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing VCenterServer by filters like name, cloud, tenant
// Api creates VCenterServer object with every call.
func (client *VCenterServerClient) GetObject(options ...session.ApiOptionsParams) (*models.VCenterServer, error) {
	var obj *models.VCenterServer
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vcenterserver", newOptions...)
	return obj, err
}

// Create a new VCenterServer object
func (client *VCenterServerClient) Create(obj *models.VCenterServer, options ...session.ApiOptionsParams) (*models.VCenterServer, error) {
	var robj *models.VCenterServer
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing VCenterServer object
func (client *VCenterServerClient) Update(obj *models.VCenterServer, options ...session.ApiOptionsParams) (*models.VCenterServer, error) {
	var robj *models.VCenterServer
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing VCenterServer object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VCenterServer
// or it should be json compatible of form map[string]interface{}
func (client *VCenterServerClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.VCenterServer, error) {
	var robj *models.VCenterServer
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing VCenterServer object with a given UUID
func (client *VCenterServerClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing VCenterServer object with a given name
func (client *VCenterServerClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *VCenterServerClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
