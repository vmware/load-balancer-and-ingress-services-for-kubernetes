// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// LicenseStatusClient is a client for avi LicenseStatus resource
type LicenseStatusClient struct {
	aviSession *session.AviSession
}

// NewLicenseStatusClient creates a new client for LicenseStatus resource
func NewLicenseStatusClient(aviSession *session.AviSession) *LicenseStatusClient {
	return &LicenseStatusClient{aviSession: aviSession}
}

func (client *LicenseStatusClient) getAPIPath(uuid string) string {
	path := "api/licensestatus"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of LicenseStatus objects
func (client *LicenseStatusClient) GetAll(options ...session.ApiOptionsParams) ([]*models.LicenseStatus, error) {
	var plist []*models.LicenseStatus
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing LicenseStatus by uuid
func (client *LicenseStatusClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.LicenseStatus, error) {
	var obj *models.LicenseStatus
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing LicenseStatus by name
func (client *LicenseStatusClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.LicenseStatus, error) {
	var obj *models.LicenseStatus
	err := client.aviSession.GetObjectByName("licensestatus", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing LicenseStatus by filters like name, cloud, tenant
// Api creates LicenseStatus object with every call.
func (client *LicenseStatusClient) GetObject(options ...session.ApiOptionsParams) (*models.LicenseStatus, error) {
	var obj *models.LicenseStatus
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("licensestatus", newOptions...)
	return obj, err
}

// Create a new LicenseStatus object
func (client *LicenseStatusClient) Create(obj *models.LicenseStatus, options ...session.ApiOptionsParams) (*models.LicenseStatus, error) {
	var robj *models.LicenseStatus
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing LicenseStatus object
func (client *LicenseStatusClient) Update(obj *models.LicenseStatus, options ...session.ApiOptionsParams) (*models.LicenseStatus, error) {
	var robj *models.LicenseStatus
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing LicenseStatus object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.LicenseStatus
// or it should be json compatible of form map[string]interface{}
func (client *LicenseStatusClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.LicenseStatus, error) {
	var robj *models.LicenseStatus
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing LicenseStatus object with a given UUID
func (client *LicenseStatusClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing LicenseStatus object with a given name
func (client *LicenseStatusClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *LicenseStatusClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
