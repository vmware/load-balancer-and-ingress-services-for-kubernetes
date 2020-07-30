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

// UpgradeStatusSummaryClient is a client for avi UpgradeStatusSummary resource
type UpgradeStatusSummaryClient struct {
	aviSession *session.AviSession
}

// NewUpgradeStatusSummaryClient creates a new client for UpgradeStatusSummary resource
func NewUpgradeStatusSummaryClient(aviSession *session.AviSession) *UpgradeStatusSummaryClient {
	return &UpgradeStatusSummaryClient{aviSession: aviSession}
}

func (client *UpgradeStatusSummaryClient) getAPIPath(uuid string) string {
	path := "api/upgradestatussummary"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of UpgradeStatusSummary objects
func (client *UpgradeStatusSummaryClient) GetAll(options ...session.ApiOptionsParams) ([]*models.UpgradeStatusSummary, error) {
	var plist []*models.UpgradeStatusSummary
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing UpgradeStatusSummary by uuid
func (client *UpgradeStatusSummaryClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.UpgradeStatusSummary, error) {
	var obj *models.UpgradeStatusSummary
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing UpgradeStatusSummary by name
func (client *UpgradeStatusSummaryClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.UpgradeStatusSummary, error) {
	var obj *models.UpgradeStatusSummary
	err := client.aviSession.GetObjectByName("upgradestatussummary", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing UpgradeStatusSummary by filters like name, cloud, tenant
// Api creates UpgradeStatusSummary object with every call.
func (client *UpgradeStatusSummaryClient) GetObject(options ...session.ApiOptionsParams) (*models.UpgradeStatusSummary, error) {
	var obj *models.UpgradeStatusSummary
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("upgradestatussummary", newOptions...)
	return obj, err
}

// Create a new UpgradeStatusSummary object
func (client *UpgradeStatusSummaryClient) Create(obj *models.UpgradeStatusSummary, options ...session.ApiOptionsParams) (*models.UpgradeStatusSummary, error) {
	var robj *models.UpgradeStatusSummary
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing UpgradeStatusSummary object
func (client *UpgradeStatusSummaryClient) Update(obj *models.UpgradeStatusSummary, options ...session.ApiOptionsParams) (*models.UpgradeStatusSummary, error) {
	var robj *models.UpgradeStatusSummary
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing UpgradeStatusSummary object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.UpgradeStatusSummary
// or it should be json compatible of form map[string]interface{}
func (client *UpgradeStatusSummaryClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.UpgradeStatusSummary, error) {
	var robj *models.UpgradeStatusSummary
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing UpgradeStatusSummary object with a given UUID
func (client *UpgradeStatusSummaryClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing UpgradeStatusSummary object with a given name
func (client *UpgradeStatusSummaryClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *UpgradeStatusSummaryClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
