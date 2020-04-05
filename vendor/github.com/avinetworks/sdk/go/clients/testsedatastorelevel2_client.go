/***************************************************************************
 *
 * AVI CONFIDENTIAL
 * __________________
 *
 * [2013] - [2018] Avi Networks Incorporated
 * All Rights Reserved.
 *
 * NOTICE: All information contained herein is, and remains the property
 * of Avi Networks Incorporated and its suppliers, if any. The intellectual
 * and technical concepts contained herein are proprietary to Avi Networks
 * Incorporated, and its suppliers and are covered by U.S. and Foreign
 * Patents, patents in process, and are protected by trade secret or
 * copyright law, and other laws. Dissemination of this information or
 * reproduction of this material is strictly forbidden unless prior written
 * permission is obtained from Avi Networks Incorporated.
 */

package clients

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

import (
	"github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
)

// TestSeDatastoreLevel2Client is a client for avi TestSeDatastoreLevel2 resource
type TestSeDatastoreLevel2Client struct {
	aviSession *session.AviSession
}

// NewTestSeDatastoreLevel2Client creates a new client for TestSeDatastoreLevel2 resource
func NewTestSeDatastoreLevel2Client(aviSession *session.AviSession) *TestSeDatastoreLevel2Client {
	return &TestSeDatastoreLevel2Client{aviSession: aviSession}
}

func (client *TestSeDatastoreLevel2Client) getAPIPath(uuid string) string {
	path := "api/testsedatastorelevel2"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of TestSeDatastoreLevel2 objects
func (client *TestSeDatastoreLevel2Client) GetAll(options ...session.ApiOptionsParams) ([]*models.TestSeDatastoreLevel2, error) {
	var plist []*models.TestSeDatastoreLevel2
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing TestSeDatastoreLevel2 by uuid
func (client *TestSeDatastoreLevel2Client) Get(uuid string, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel2, error) {
	var obj *models.TestSeDatastoreLevel2
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing TestSeDatastoreLevel2 by name
func (client *TestSeDatastoreLevel2Client) GetByName(name string, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel2, error) {
	var obj *models.TestSeDatastoreLevel2
	err := client.aviSession.GetObjectByName("testsedatastorelevel2", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing TestSeDatastoreLevel2 by filters like name, cloud, tenant
// Api creates TestSeDatastoreLevel2 object with every call.
func (client *TestSeDatastoreLevel2Client) GetObject(options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel2, error) {
	var obj *models.TestSeDatastoreLevel2
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("testsedatastorelevel2", newOptions...)
	return obj, err
}

// Create a new TestSeDatastoreLevel2 object
func (client *TestSeDatastoreLevel2Client) Create(obj *models.TestSeDatastoreLevel2, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel2, error) {
	var robj *models.TestSeDatastoreLevel2
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing TestSeDatastoreLevel2 object
func (client *TestSeDatastoreLevel2Client) Update(obj *models.TestSeDatastoreLevel2, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel2, error) {
	var robj *models.TestSeDatastoreLevel2
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing TestSeDatastoreLevel2 object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.TestSeDatastoreLevel2
// or it should be json compatible of form map[string]interface{}
func (client *TestSeDatastoreLevel2Client) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel2, error) {
	var robj *models.TestSeDatastoreLevel2
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing TestSeDatastoreLevel2 object with a given UUID
func (client *TestSeDatastoreLevel2Client) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing TestSeDatastoreLevel2 object with a given name
func (client *TestSeDatastoreLevel2Client) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *TestSeDatastoreLevel2Client) GetAviSession() *session.AviSession {
	return client.aviSession
}
