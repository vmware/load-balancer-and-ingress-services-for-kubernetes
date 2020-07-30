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

// FederationCheckpointClient is a client for avi FederationCheckpoint resource
type FederationCheckpointClient struct {
	aviSession *session.AviSession
}

// NewFederationCheckpointClient creates a new client for FederationCheckpoint resource
func NewFederationCheckpointClient(aviSession *session.AviSession) *FederationCheckpointClient {
	return &FederationCheckpointClient{aviSession: aviSession}
}

func (client *FederationCheckpointClient) getAPIPath(uuid string) string {
	path := "api/federationcheckpoint"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of FederationCheckpoint objects
func (client *FederationCheckpointClient) GetAll(options ...session.ApiOptionsParams) ([]*models.FederationCheckpoint, error) {
	var plist []*models.FederationCheckpoint
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing FederationCheckpoint by uuid
func (client *FederationCheckpointClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.FederationCheckpoint, error) {
	var obj *models.FederationCheckpoint
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing FederationCheckpoint by name
func (client *FederationCheckpointClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.FederationCheckpoint, error) {
	var obj *models.FederationCheckpoint
	err := client.aviSession.GetObjectByName("federationcheckpoint", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing FederationCheckpoint by filters like name, cloud, tenant
// Api creates FederationCheckpoint object with every call.
func (client *FederationCheckpointClient) GetObject(options ...session.ApiOptionsParams) (*models.FederationCheckpoint, error) {
	var obj *models.FederationCheckpoint
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("federationcheckpoint", newOptions...)
	return obj, err
}

// Create a new FederationCheckpoint object
func (client *FederationCheckpointClient) Create(obj *models.FederationCheckpoint, options ...session.ApiOptionsParams) (*models.FederationCheckpoint, error) {
	var robj *models.FederationCheckpoint
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing FederationCheckpoint object
func (client *FederationCheckpointClient) Update(obj *models.FederationCheckpoint, options ...session.ApiOptionsParams) (*models.FederationCheckpoint, error) {
	var robj *models.FederationCheckpoint
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing FederationCheckpoint object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.FederationCheckpoint
// or it should be json compatible of form map[string]interface{}
func (client *FederationCheckpointClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.FederationCheckpoint, error) {
	var robj *models.FederationCheckpoint
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing FederationCheckpoint object with a given UUID
func (client *FederationCheckpointClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing FederationCheckpoint object with a given name
func (client *FederationCheckpointClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *FederationCheckpointClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
