// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// ALBServicesFileDownloadClient is a client for avi ALBServicesFileDownload resource
type ALBServicesFileDownloadClient struct {
	aviSession *session.AviSession
}

// NewALBServicesFileDownloadClient creates a new client for ALBServicesFileDownload resource
func NewALBServicesFileDownloadClient(aviSession *session.AviSession) *ALBServicesFileDownloadClient {
	return &ALBServicesFileDownloadClient{aviSession: aviSession}
}

func (client *ALBServicesFileDownloadClient) getAPIPath(uuid string) string {
	path := "api/albservicesfiledownload"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ALBServicesFileDownload objects
func (client *ALBServicesFileDownloadClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ALBServicesFileDownload, error) {
	var plist []*models.ALBServicesFileDownload
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ALBServicesFileDownload by uuid
func (client *ALBServicesFileDownloadClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ALBServicesFileDownload, error) {
	var obj *models.ALBServicesFileDownload
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ALBServicesFileDownload by name
func (client *ALBServicesFileDownloadClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ALBServicesFileDownload, error) {
	var obj *models.ALBServicesFileDownload
	err := client.aviSession.GetObjectByName("albservicesfiledownload", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ALBServicesFileDownload by filters like name, cloud, tenant
// Api creates ALBServicesFileDownload object with every call.
func (client *ALBServicesFileDownloadClient) GetObject(options ...session.ApiOptionsParams) (*models.ALBServicesFileDownload, error) {
	var obj *models.ALBServicesFileDownload
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("albservicesfiledownload", newOptions...)
	return obj, err
}

// Create a new ALBServicesFileDownload object
func (client *ALBServicesFileDownloadClient) Create(obj *models.ALBServicesFileDownload, options ...session.ApiOptionsParams) (*models.ALBServicesFileDownload, error) {
	var robj *models.ALBServicesFileDownload
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ALBServicesFileDownload object
func (client *ALBServicesFileDownloadClient) Update(obj *models.ALBServicesFileDownload, options ...session.ApiOptionsParams) (*models.ALBServicesFileDownload, error) {
	var robj *models.ALBServicesFileDownload
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ALBServicesFileDownload object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ALBServicesFileDownload
// or it should be json compatible of form map[string]interface{}
func (client *ALBServicesFileDownloadClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ALBServicesFileDownload, error) {
	var robj *models.ALBServicesFileDownload
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ALBServicesFileDownload object with a given UUID
func (client *ALBServicesFileDownloadClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ALBServicesFileDownload object with a given name
func (client *ALBServicesFileDownloadClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ALBServicesFileDownloadClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
