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

// WafProfileClient is a client for avi WafProfile resource
type WafProfileClient struct {
	aviSession *session.AviSession
}

// NewWafProfileClient creates a new client for WafProfile resource
func NewWafProfileClient(aviSession *session.AviSession) *WafProfileClient {
	return &WafProfileClient{aviSession: aviSession}
}

func (client *WafProfileClient) getAPIPath(uuid string) string {
	path := "api/wafprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of WafProfile objects
func (client *WafProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.WafProfile, error) {
	var plist []*models.WafProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing WafProfile by uuid
func (client *WafProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.WafProfile, error) {
	var obj *models.WafProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing WafProfile by name
func (client *WafProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.WafProfile, error) {
	var obj *models.WafProfile
	err := client.aviSession.GetObjectByName("wafprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing WafProfile by filters like name, cloud, tenant
// Api creates WafProfile object with every call.
func (client *WafProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.WafProfile, error) {
	var obj *models.WafProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("wafprofile", newOptions...)
	return obj, err
}

// Create a new WafProfile object
func (client *WafProfileClient) Create(obj *models.WafProfile, options ...session.ApiOptionsParams) (*models.WafProfile, error) {
	var robj *models.WafProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing WafProfile object
func (client *WafProfileClient) Update(obj *models.WafProfile, options ...session.ApiOptionsParams) (*models.WafProfile, error) {
	var robj *models.WafProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing WafProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.WafProfile
// or it should be json compatible of form map[string]interface{}
func (client *WafProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.WafProfile, error) {
	var robj *models.WafProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing WafProfile object with a given UUID
func (client *WafProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing WafProfile object with a given name
func (client *WafProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *WafProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
