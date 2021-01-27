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

// DynamicDNSRecordClient is a client for avi DynamicDNSRecord resource
type DynamicDNSRecordClient struct {
	aviSession *session.AviSession
}

// NewDynamicDNSRecordClient creates a new client for DynamicDNSRecord resource
func NewDynamicDNSRecordClient(aviSession *session.AviSession) *DynamicDNSRecordClient {
	return &DynamicDNSRecordClient{aviSession: aviSession}
}

func (client *DynamicDNSRecordClient) getAPIPath(uuid string) string {
	path := "api/dynamicdnsrecord"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of DynamicDNSRecord objects
func (client *DynamicDNSRecordClient) GetAll(options ...session.ApiOptionsParams) ([]*models.DynamicDNSRecord, error) {
	var plist []*models.DynamicDNSRecord
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing DynamicDNSRecord by uuid
func (client *DynamicDNSRecordClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.DynamicDNSRecord, error) {
	var obj *models.DynamicDNSRecord
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing DynamicDNSRecord by name
func (client *DynamicDNSRecordClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.DynamicDNSRecord, error) {
	var obj *models.DynamicDNSRecord
	err := client.aviSession.GetObjectByName("dynamicdnsrecord", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing DynamicDNSRecord by filters like name, cloud, tenant
// Api creates DynamicDNSRecord object with every call.
func (client *DynamicDNSRecordClient) GetObject(options ...session.ApiOptionsParams) (*models.DynamicDNSRecord, error) {
	var obj *models.DynamicDNSRecord
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("dynamicdnsrecord", newOptions...)
	return obj, err
}

// Create a new DynamicDNSRecord object
func (client *DynamicDNSRecordClient) Create(obj *models.DynamicDNSRecord, options ...session.ApiOptionsParams) (*models.DynamicDNSRecord, error) {
	var robj *models.DynamicDNSRecord
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing DynamicDNSRecord object
func (client *DynamicDNSRecordClient) Update(obj *models.DynamicDNSRecord, options ...session.ApiOptionsParams) (*models.DynamicDNSRecord, error) {
	var robj *models.DynamicDNSRecord
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing DynamicDNSRecord object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.DynamicDNSRecord
// or it should be json compatible of form map[string]interface{}
func (client *DynamicDNSRecordClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.DynamicDNSRecord, error) {
	var robj *models.DynamicDNSRecord
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing DynamicDNSRecord object with a given UUID
func (client *DynamicDNSRecordClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing DynamicDNSRecord object with a given name
func (client *DynamicDNSRecordClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *DynamicDNSRecordClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
