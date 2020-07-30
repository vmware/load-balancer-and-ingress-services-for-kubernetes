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

// AvailabilityZoneClient is a client for avi AvailabilityZone resource
type AvailabilityZoneClient struct {
	aviSession *session.AviSession
}

// NewAvailabilityZoneClient creates a new client for AvailabilityZone resource
func NewAvailabilityZoneClient(aviSession *session.AviSession) *AvailabilityZoneClient {
	return &AvailabilityZoneClient{aviSession: aviSession}
}

func (client *AvailabilityZoneClient) getAPIPath(uuid string) string {
	path := "api/availabilityzone"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of AvailabilityZone objects
func (client *AvailabilityZoneClient) GetAll(options ...session.ApiOptionsParams) ([]*models.AvailabilityZone, error) {
	var plist []*models.AvailabilityZone
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing AvailabilityZone by uuid
func (client *AvailabilityZoneClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.AvailabilityZone, error) {
	var obj *models.AvailabilityZone
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing AvailabilityZone by name
func (client *AvailabilityZoneClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.AvailabilityZone, error) {
	var obj *models.AvailabilityZone
	err := client.aviSession.GetObjectByName("availabilityzone", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing AvailabilityZone by filters like name, cloud, tenant
// Api creates AvailabilityZone object with every call.
func (client *AvailabilityZoneClient) GetObject(options ...session.ApiOptionsParams) (*models.AvailabilityZone, error) {
	var obj *models.AvailabilityZone
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("availabilityzone", newOptions...)
	return obj, err
}

// Create a new AvailabilityZone object
func (client *AvailabilityZoneClient) Create(obj *models.AvailabilityZone, options ...session.ApiOptionsParams) (*models.AvailabilityZone, error) {
	var robj *models.AvailabilityZone
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing AvailabilityZone object
func (client *AvailabilityZoneClient) Update(obj *models.AvailabilityZone, options ...session.ApiOptionsParams) (*models.AvailabilityZone, error) {
	var robj *models.AvailabilityZone
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing AvailabilityZone object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.AvailabilityZone
// or it should be json compatible of form map[string]interface{}
func (client *AvailabilityZoneClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.AvailabilityZone, error) {
	var robj *models.AvailabilityZone
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing AvailabilityZone object with a given UUID
func (client *AvailabilityZoneClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing AvailabilityZone object with a given name
func (client *AvailabilityZoneClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *AvailabilityZoneClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
