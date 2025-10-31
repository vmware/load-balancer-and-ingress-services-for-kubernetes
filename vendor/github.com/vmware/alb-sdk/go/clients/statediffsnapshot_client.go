// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// StatediffSnapshotClient is a client for avi StatediffSnapshot resource
type StatediffSnapshotClient struct {
	aviSession *session.AviSession
}

// NewStatediffSnapshotClient creates a new client for StatediffSnapshot resource
func NewStatediffSnapshotClient(aviSession *session.AviSession) *StatediffSnapshotClient {
	return &StatediffSnapshotClient{aviSession: aviSession}
}

func (client *StatediffSnapshotClient) getAPIPath(uuid string) string {
	path := "api/statediffsnapshot"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of StatediffSnapshot objects
func (client *StatediffSnapshotClient) GetAll(options ...session.ApiOptionsParams) ([]*models.StatediffSnapshot, error) {
	var plist []*models.StatediffSnapshot
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing StatediffSnapshot by uuid
func (client *StatediffSnapshotClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.StatediffSnapshot, error) {
	var obj *models.StatediffSnapshot
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing StatediffSnapshot by name
func (client *StatediffSnapshotClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.StatediffSnapshot, error) {
	var obj *models.StatediffSnapshot
	err := client.aviSession.GetObjectByName("statediffsnapshot", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing StatediffSnapshot by filters like name, cloud, tenant
// Api creates StatediffSnapshot object with every call.
func (client *StatediffSnapshotClient) GetObject(options ...session.ApiOptionsParams) (*models.StatediffSnapshot, error) {
	var obj *models.StatediffSnapshot
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("statediffsnapshot", newOptions...)
	return obj, err
}

// Create a new StatediffSnapshot object
func (client *StatediffSnapshotClient) Create(obj *models.StatediffSnapshot, options ...session.ApiOptionsParams) (*models.StatediffSnapshot, error) {
	var robj *models.StatediffSnapshot
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing StatediffSnapshot object
func (client *StatediffSnapshotClient) Update(obj *models.StatediffSnapshot, options ...session.ApiOptionsParams) (*models.StatediffSnapshot, error) {
	var robj *models.StatediffSnapshot
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing StatediffSnapshot object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.StatediffSnapshot
// or it should be json compatible of form map[string]interface{}
func (client *StatediffSnapshotClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.StatediffSnapshot, error) {
	var robj *models.StatediffSnapshot
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing StatediffSnapshot object with a given UUID
func (client *StatediffSnapshotClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing StatediffSnapshot object with a given name
func (client *StatediffSnapshotClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *StatediffSnapshotClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
