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

// SePropertiesClient is a client for avi SeProperties resource
type SePropertiesClient struct {
	aviSession *session.AviSession
}

// NewSePropertiesClient creates a new client for SeProperties resource
func NewSePropertiesClient(aviSession *session.AviSession) *SePropertiesClient {
	return &SePropertiesClient{aviSession: aviSession}
}

func (client *SePropertiesClient) getAPIPath(uuid string) string {
	path := "api/seproperties"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of SeProperties objects
func (client *SePropertiesClient) GetAll(options ...session.ApiOptionsParams) ([]*models.SeProperties, error) {
	var plist []*models.SeProperties
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing SeProperties by uuid
func (client *SePropertiesClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.SeProperties, error) {
	var obj *models.SeProperties
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing SeProperties by name
func (client *SePropertiesClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.SeProperties, error) {
	var obj *models.SeProperties
	err := client.aviSession.GetObjectByName("seproperties", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing SeProperties by filters like name, cloud, tenant
// Api creates SeProperties object with every call.
func (client *SePropertiesClient) GetObject(options ...session.ApiOptionsParams) (*models.SeProperties, error) {
	var obj *models.SeProperties
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("seproperties", newOptions...)
	return obj, err
}

// Create a new SeProperties object
func (client *SePropertiesClient) Create(obj *models.SeProperties, options ...session.ApiOptionsParams) (*models.SeProperties, error) {
	var robj *models.SeProperties
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing SeProperties object
func (client *SePropertiesClient) Update(obj *models.SeProperties, options ...session.ApiOptionsParams) (*models.SeProperties, error) {
	var robj *models.SeProperties
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing SeProperties object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.SeProperties
// or it should be json compatible of form map[string]interface{}
func (client *SePropertiesClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.SeProperties, error) {
	var robj *models.SeProperties
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing SeProperties object with a given UUID
func (client *SePropertiesClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing SeProperties object with a given name
func (client *SePropertiesClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *SePropertiesClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
