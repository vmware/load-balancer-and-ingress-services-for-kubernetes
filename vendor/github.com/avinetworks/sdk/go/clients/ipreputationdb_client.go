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

// IPReputationDBClient is a client for avi IPReputationDB resource
type IPReputationDBClient struct {
	aviSession *session.AviSession
}

// NewIPReputationDBClient creates a new client for IPReputationDB resource
func NewIPReputationDBClient(aviSession *session.AviSession) *IPReputationDBClient {
	return &IPReputationDBClient{aviSession: aviSession}
}

func (client *IPReputationDBClient) getAPIPath(uuid string) string {
	path := "api/ipreputationdb"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of IPReputationDB objects
func (client *IPReputationDBClient) GetAll(options ...session.ApiOptionsParams) ([]*models.IPReputationDB, error) {
	var plist []*models.IPReputationDB
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing IPReputationDB by uuid
func (client *IPReputationDBClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.IPReputationDB, error) {
	var obj *models.IPReputationDB
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing IPReputationDB by name
func (client *IPReputationDBClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.IPReputationDB, error) {
	var obj *models.IPReputationDB
	err := client.aviSession.GetObjectByName("ipreputationdb", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing IPReputationDB by filters like name, cloud, tenant
// Api creates IPReputationDB object with every call.
func (client *IPReputationDBClient) GetObject(options ...session.ApiOptionsParams) (*models.IPReputationDB, error) {
	var obj *models.IPReputationDB
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("ipreputationdb", newOptions...)
	return obj, err
}

// Create a new IPReputationDB object
func (client *IPReputationDBClient) Create(obj *models.IPReputationDB, options ...session.ApiOptionsParams) (*models.IPReputationDB, error) {
	var robj *models.IPReputationDB
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing IPReputationDB object
func (client *IPReputationDBClient) Update(obj *models.IPReputationDB, options ...session.ApiOptionsParams) (*models.IPReputationDB, error) {
	var robj *models.IPReputationDB
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing IPReputationDB object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.IPReputationDB
// or it should be json compatible of form map[string]interface{}
func (client *IPReputationDBClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.IPReputationDB, error) {
	var robj *models.IPReputationDB
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing IPReputationDB object with a given UUID
func (client *IPReputationDBClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing IPReputationDB object with a given name
func (client *IPReputationDBClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *IPReputationDBClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
