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

// LicenseLedgerDetailsClient is a client for avi LicenseLedgerDetails resource
type LicenseLedgerDetailsClient struct {
	aviSession *session.AviSession
}

// NewLicenseLedgerDetailsClient creates a new client for LicenseLedgerDetails resource
func NewLicenseLedgerDetailsClient(aviSession *session.AviSession) *LicenseLedgerDetailsClient {
	return &LicenseLedgerDetailsClient{aviSession: aviSession}
}

func (client *LicenseLedgerDetailsClient) getAPIPath(uuid string) string {
	path := "api/licenseledgerdetails"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of LicenseLedgerDetails objects
func (client *LicenseLedgerDetailsClient) GetAll(options ...session.ApiOptionsParams) ([]*models.LicenseLedgerDetails, error) {
	var plist []*models.LicenseLedgerDetails
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing LicenseLedgerDetails by uuid
func (client *LicenseLedgerDetailsClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.LicenseLedgerDetails, error) {
	var obj *models.LicenseLedgerDetails
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing LicenseLedgerDetails by name
func (client *LicenseLedgerDetailsClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.LicenseLedgerDetails, error) {
	var obj *models.LicenseLedgerDetails
	err := client.aviSession.GetObjectByName("licenseledgerdetails", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing LicenseLedgerDetails by filters like name, cloud, tenant
// Api creates LicenseLedgerDetails object with every call.
func (client *LicenseLedgerDetailsClient) GetObject(options ...session.ApiOptionsParams) (*models.LicenseLedgerDetails, error) {
	var obj *models.LicenseLedgerDetails
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("licenseledgerdetails", newOptions...)
	return obj, err
}

// Create a new LicenseLedgerDetails object
func (client *LicenseLedgerDetailsClient) Create(obj *models.LicenseLedgerDetails, options ...session.ApiOptionsParams) (*models.LicenseLedgerDetails, error) {
	var robj *models.LicenseLedgerDetails
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing LicenseLedgerDetails object
func (client *LicenseLedgerDetailsClient) Update(obj *models.LicenseLedgerDetails, options ...session.ApiOptionsParams) (*models.LicenseLedgerDetails, error) {
	var robj *models.LicenseLedgerDetails
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing LicenseLedgerDetails object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.LicenseLedgerDetails
// or it should be json compatible of form map[string]interface{}
func (client *LicenseLedgerDetailsClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.LicenseLedgerDetails, error) {
	var robj *models.LicenseLedgerDetails
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing LicenseLedgerDetails object with a given UUID
func (client *LicenseLedgerDetailsClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing LicenseLedgerDetails object with a given name
func (client *LicenseLedgerDetailsClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *LicenseLedgerDetailsClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
