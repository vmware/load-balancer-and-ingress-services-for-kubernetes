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

// PKIprofileClient is a client for avi PKIprofile resource
type PKIprofileClient struct {
	aviSession *session.AviSession
}

// NewPKIprofileClient creates a new client for PKIprofile resource
func NewPKIprofileClient(aviSession *session.AviSession) *PKIprofileClient {
	return &PKIprofileClient{aviSession: aviSession}
}

func (client *PKIprofileClient) getAPIPath(uuid string) string {
	path := "api/pkiprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of PKIprofile objects
func (client *PKIprofileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.PKIprofile, error) {
	var plist []*models.PKIprofile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing PKIprofile by uuid
func (client *PKIprofileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.PKIprofile, error) {
	var obj *models.PKIprofile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing PKIprofile by name
func (client *PKIprofileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.PKIprofile, error) {
	var obj *models.PKIprofile
	err := client.aviSession.GetObjectByName("pkiprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing PKIprofile by filters like name, cloud, tenant
// Api creates PKIprofile object with every call.
func (client *PKIprofileClient) GetObject(options ...session.ApiOptionsParams) (*models.PKIprofile, error) {
	var obj *models.PKIprofile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("pkiprofile", newOptions...)
	return obj, err
}

// Create a new PKIprofile object
func (client *PKIprofileClient) Create(obj *models.PKIprofile, options ...session.ApiOptionsParams) (*models.PKIprofile, error) {
	var robj *models.PKIprofile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing PKIprofile object
func (client *PKIprofileClient) Update(obj *models.PKIprofile, options ...session.ApiOptionsParams) (*models.PKIprofile, error) {
	var robj *models.PKIprofile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing PKIprofile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.PKIprofile
// or it should be json compatible of form map[string]interface{}
func (client *PKIprofileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.PKIprofile, error) {
	var robj *models.PKIprofile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing PKIprofile object with a given UUID
func (client *PKIprofileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing PKIprofile object with a given name
func (client *PKIprofileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *PKIprofileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
