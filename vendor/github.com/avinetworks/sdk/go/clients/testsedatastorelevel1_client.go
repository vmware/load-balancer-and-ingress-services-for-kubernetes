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

// TestSeDatastoreLevel1Client is a client for avi TestSeDatastoreLevel1 resource
type TestSeDatastoreLevel1Client struct {
	aviSession *session.AviSession
}

// NewTestSeDatastoreLevel1Client creates a new client for TestSeDatastoreLevel1 resource
func NewTestSeDatastoreLevel1Client(aviSession *session.AviSession) *TestSeDatastoreLevel1Client {
	return &TestSeDatastoreLevel1Client{aviSession: aviSession}
}

func (client *TestSeDatastoreLevel1Client) getAPIPath(uuid string) string {
	path := "api/testsedatastorelevel1"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of TestSeDatastoreLevel1 objects
func (client *TestSeDatastoreLevel1Client) GetAll(options ...session.ApiOptionsParams) ([]*models.TestSeDatastoreLevel1, error) {
	var plist []*models.TestSeDatastoreLevel1
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing TestSeDatastoreLevel1 by uuid
func (client *TestSeDatastoreLevel1Client) Get(uuid string, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel1, error) {
	var obj *models.TestSeDatastoreLevel1
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing TestSeDatastoreLevel1 by name
func (client *TestSeDatastoreLevel1Client) GetByName(name string, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel1, error) {
	var obj *models.TestSeDatastoreLevel1
	err := client.aviSession.GetObjectByName("testsedatastorelevel1", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing TestSeDatastoreLevel1 by filters like name, cloud, tenant
// Api creates TestSeDatastoreLevel1 object with every call.
func (client *TestSeDatastoreLevel1Client) GetObject(options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel1, error) {
	var obj *models.TestSeDatastoreLevel1
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("testsedatastorelevel1", newOptions...)
	return obj, err
}

// Create a new TestSeDatastoreLevel1 object
func (client *TestSeDatastoreLevel1Client) Create(obj *models.TestSeDatastoreLevel1, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel1, error) {
	var robj *models.TestSeDatastoreLevel1
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing TestSeDatastoreLevel1 object
func (client *TestSeDatastoreLevel1Client) Update(obj *models.TestSeDatastoreLevel1, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel1, error) {
	var robj *models.TestSeDatastoreLevel1
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing TestSeDatastoreLevel1 object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.TestSeDatastoreLevel1
// or it should be json compatible of form map[string]interface{}
func (client *TestSeDatastoreLevel1Client) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel1, error) {
	var robj *models.TestSeDatastoreLevel1
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing TestSeDatastoreLevel1 object with a given UUID
func (client *TestSeDatastoreLevel1Client) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing TestSeDatastoreLevel1 object with a given name
func (client *TestSeDatastoreLevel1Client) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *TestSeDatastoreLevel1Client) GetAviSession() *session.AviSession {
	return client.aviSession
}
