// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// SystemReportClient is a client for avi SystemReport resource
type SystemReportClient struct {
	aviSession *session.AviSession
}

// NewSystemReportClient creates a new client for SystemReport resource
func NewSystemReportClient(aviSession *session.AviSession) *SystemReportClient {
	return &SystemReportClient{aviSession: aviSession}
}

func (client *SystemReportClient) getAPIPath(uuid string) string {
	path := "api/systemreport"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of SystemReport objects
func (client *SystemReportClient) GetAll(options ...session.ApiOptionsParams) ([]*models.SystemReport, error) {
	var plist []*models.SystemReport
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing SystemReport by uuid
func (client *SystemReportClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.SystemReport, error) {
	var obj *models.SystemReport
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing SystemReport by name
func (client *SystemReportClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.SystemReport, error) {
	var obj *models.SystemReport
	err := client.aviSession.GetObjectByName("systemreport", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing SystemReport by filters like name, cloud, tenant
// Api creates SystemReport object with every call.
func (client *SystemReportClient) GetObject(options ...session.ApiOptionsParams) (*models.SystemReport, error) {
	var obj *models.SystemReport
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("systemreport", newOptions...)
	return obj, err
}

// Create a new SystemReport object
func (client *SystemReportClient) Create(obj *models.SystemReport, options ...session.ApiOptionsParams) (*models.SystemReport, error) {
	var robj *models.SystemReport
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing SystemReport object
func (client *SystemReportClient) Update(obj *models.SystemReport, options ...session.ApiOptionsParams) (*models.SystemReport, error) {
	var robj *models.SystemReport
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing SystemReport object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.SystemReport
// or it should be json compatible of form map[string]interface{}
func (client *SystemReportClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.SystemReport, error) {
	var robj *models.SystemReport
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing SystemReport object with a given UUID
func (client *SystemReportClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing SystemReport object with a given name
func (client *SystemReportClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *SystemReportClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
