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

// SiteVersionClient is a client for avi SiteVersion resource
type SiteVersionClient struct {
	aviSession *session.AviSession
}

// NewSiteVersionClient creates a new client for SiteVersion resource
func NewSiteVersionClient(aviSession *session.AviSession) *SiteVersionClient {
	return &SiteVersionClient{aviSession: aviSession}
}

func (client *SiteVersionClient) getAPIPath(uuid string) string {
	path := "api/siteversion"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of SiteVersion objects
func (client *SiteVersionClient) GetAll(options ...session.ApiOptionsParams) ([]*models.SiteVersion, error) {
	var plist []*models.SiteVersion
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing SiteVersion by uuid
func (client *SiteVersionClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.SiteVersion, error) {
	var obj *models.SiteVersion
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing SiteVersion by name
func (client *SiteVersionClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.SiteVersion, error) {
	var obj *models.SiteVersion
	err := client.aviSession.GetObjectByName("siteversion", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing SiteVersion by filters like name, cloud, tenant
// Api creates SiteVersion object with every call.
func (client *SiteVersionClient) GetObject(options ...session.ApiOptionsParams) (*models.SiteVersion, error) {
	var obj *models.SiteVersion
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("siteversion", newOptions...)
	return obj, err
}

// Create a new SiteVersion object
func (client *SiteVersionClient) Create(obj *models.SiteVersion, options ...session.ApiOptionsParams) (*models.SiteVersion, error) {
	var robj *models.SiteVersion
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing SiteVersion object
func (client *SiteVersionClient) Update(obj *models.SiteVersion, options ...session.ApiOptionsParams) (*models.SiteVersion, error) {
	var robj *models.SiteVersion
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing SiteVersion object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.SiteVersion
// or it should be json compatible of form map[string]interface{}
func (client *SiteVersionClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.SiteVersion, error) {
	var robj *models.SiteVersion
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing SiteVersion object with a given UUID
func (client *SiteVersionClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing SiteVersion object with a given name
func (client *SiteVersionClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *SiteVersionClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
