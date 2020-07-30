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

// IcapProfileClient is a client for avi IcapProfile resource
type IcapProfileClient struct {
	aviSession *session.AviSession
}

// NewIcapProfileClient creates a new client for IcapProfile resource
func NewIcapProfileClient(aviSession *session.AviSession) *IcapProfileClient {
	return &IcapProfileClient{aviSession: aviSession}
}

func (client *IcapProfileClient) getAPIPath(uuid string) string {
	path := "api/icapprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of IcapProfile objects
func (client *IcapProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.IcapProfile, error) {
	var plist []*models.IcapProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing IcapProfile by uuid
func (client *IcapProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.IcapProfile, error) {
	var obj *models.IcapProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing IcapProfile by name
func (client *IcapProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.IcapProfile, error) {
	var obj *models.IcapProfile
	err := client.aviSession.GetObjectByName("icapprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing IcapProfile by filters like name, cloud, tenant
// Api creates IcapProfile object with every call.
func (client *IcapProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.IcapProfile, error) {
	var obj *models.IcapProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("icapprofile", newOptions...)
	return obj, err
}

// Create a new IcapProfile object
func (client *IcapProfileClient) Create(obj *models.IcapProfile, options ...session.ApiOptionsParams) (*models.IcapProfile, error) {
	var robj *models.IcapProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing IcapProfile object
func (client *IcapProfileClient) Update(obj *models.IcapProfile, options ...session.ApiOptionsParams) (*models.IcapProfile, error) {
	var robj *models.IcapProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing IcapProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.IcapProfile
// or it should be json compatible of form map[string]interface{}
func (client *IcapProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.IcapProfile, error) {
	var robj *models.IcapProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing IcapProfile object with a given UUID
func (client *IcapProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing IcapProfile object with a given name
func (client *IcapProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *IcapProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
