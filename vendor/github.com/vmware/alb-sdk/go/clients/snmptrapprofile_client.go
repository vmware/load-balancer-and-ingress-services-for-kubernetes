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

// SnmpTrapProfileClient is a client for avi SnmpTrapProfile resource
type SnmpTrapProfileClient struct {
	aviSession *session.AviSession
}

// NewSnmpTrapProfileClient creates a new client for SnmpTrapProfile resource
func NewSnmpTrapProfileClient(aviSession *session.AviSession) *SnmpTrapProfileClient {
	return &SnmpTrapProfileClient{aviSession: aviSession}
}

func (client *SnmpTrapProfileClient) getAPIPath(uuid string) string {
	path := "api/snmptrapprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of SnmpTrapProfile objects
func (client *SnmpTrapProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.SnmpTrapProfile, error) {
	var plist []*models.SnmpTrapProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing SnmpTrapProfile by uuid
func (client *SnmpTrapProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.SnmpTrapProfile, error) {
	var obj *models.SnmpTrapProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing SnmpTrapProfile by name
func (client *SnmpTrapProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.SnmpTrapProfile, error) {
	var obj *models.SnmpTrapProfile
	err := client.aviSession.GetObjectByName("snmptrapprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing SnmpTrapProfile by filters like name, cloud, tenant
// Api creates SnmpTrapProfile object with every call.
func (client *SnmpTrapProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.SnmpTrapProfile, error) {
	var obj *models.SnmpTrapProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("snmptrapprofile", newOptions...)
	return obj, err
}

// Create a new SnmpTrapProfile object
func (client *SnmpTrapProfileClient) Create(obj *models.SnmpTrapProfile, options ...session.ApiOptionsParams) (*models.SnmpTrapProfile, error) {
	var robj *models.SnmpTrapProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing SnmpTrapProfile object
func (client *SnmpTrapProfileClient) Update(obj *models.SnmpTrapProfile, options ...session.ApiOptionsParams) (*models.SnmpTrapProfile, error) {
	var robj *models.SnmpTrapProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing SnmpTrapProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.SnmpTrapProfile
// or it should be json compatible of form map[string]interface{}
func (client *SnmpTrapProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.SnmpTrapProfile, error) {
	var robj *models.SnmpTrapProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing SnmpTrapProfile object with a given UUID
func (client *SnmpTrapProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing SnmpTrapProfile object with a given name
func (client *SnmpTrapProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *SnmpTrapProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
