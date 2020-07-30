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

// WafApplicationSignatureProviderClient is a client for avi WafApplicationSignatureProvider resource
type WafApplicationSignatureProviderClient struct {
	aviSession *session.AviSession
}

// NewWafApplicationSignatureProviderClient creates a new client for WafApplicationSignatureProvider resource
func NewWafApplicationSignatureProviderClient(aviSession *session.AviSession) *WafApplicationSignatureProviderClient {
	return &WafApplicationSignatureProviderClient{aviSession: aviSession}
}

func (client *WafApplicationSignatureProviderClient) getAPIPath(uuid string) string {
	path := "api/wafapplicationsignatureprovider"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of WafApplicationSignatureProvider objects
func (client *WafApplicationSignatureProviderClient) GetAll(options ...session.ApiOptionsParams) ([]*models.WafApplicationSignatureProvider, error) {
	var plist []*models.WafApplicationSignatureProvider
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing WafApplicationSignatureProvider by uuid
func (client *WafApplicationSignatureProviderClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.WafApplicationSignatureProvider, error) {
	var obj *models.WafApplicationSignatureProvider
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing WafApplicationSignatureProvider by name
func (client *WafApplicationSignatureProviderClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.WafApplicationSignatureProvider, error) {
	var obj *models.WafApplicationSignatureProvider
	err := client.aviSession.GetObjectByName("wafapplicationsignatureprovider", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing WafApplicationSignatureProvider by filters like name, cloud, tenant
// Api creates WafApplicationSignatureProvider object with every call.
func (client *WafApplicationSignatureProviderClient) GetObject(options ...session.ApiOptionsParams) (*models.WafApplicationSignatureProvider, error) {
	var obj *models.WafApplicationSignatureProvider
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("wafapplicationsignatureprovider", newOptions...)
	return obj, err
}

// Create a new WafApplicationSignatureProvider object
func (client *WafApplicationSignatureProviderClient) Create(obj *models.WafApplicationSignatureProvider, options ...session.ApiOptionsParams) (*models.WafApplicationSignatureProvider, error) {
	var robj *models.WafApplicationSignatureProvider
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing WafApplicationSignatureProvider object
func (client *WafApplicationSignatureProviderClient) Update(obj *models.WafApplicationSignatureProvider, options ...session.ApiOptionsParams) (*models.WafApplicationSignatureProvider, error) {
	var robj *models.WafApplicationSignatureProvider
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing WafApplicationSignatureProvider object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.WafApplicationSignatureProvider
// or it should be json compatible of form map[string]interface{}
func (client *WafApplicationSignatureProviderClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.WafApplicationSignatureProvider, error) {
	var robj *models.WafApplicationSignatureProvider
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing WafApplicationSignatureProvider object with a given UUID
func (client *WafApplicationSignatureProviderClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing WafApplicationSignatureProvider object with a given name
func (client *WafApplicationSignatureProviderClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *WafApplicationSignatureProviderClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
