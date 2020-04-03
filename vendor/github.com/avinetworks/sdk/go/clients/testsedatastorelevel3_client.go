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

// TestSeDatastoreLevel3Client is a client for avi TestSeDatastoreLevel3 resource
type TestSeDatastoreLevel3Client struct {
	aviSession *session.AviSession
}

// NewTestSeDatastoreLevel3Client creates a new client for TestSeDatastoreLevel3 resource
func NewTestSeDatastoreLevel3Client(aviSession *session.AviSession) *TestSeDatastoreLevel3Client {
	return &TestSeDatastoreLevel3Client{aviSession: aviSession}
}

func (client *TestSeDatastoreLevel3Client) getAPIPath(uuid string) string {
	path := "api/testsedatastorelevel3"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of TestSeDatastoreLevel3 objects
func (client *TestSeDatastoreLevel3Client) GetAll(options ...session.ApiOptionsParams) ([]*models.TestSeDatastoreLevel3, error) {
	var plist []*models.TestSeDatastoreLevel3
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing TestSeDatastoreLevel3 by uuid
func (client *TestSeDatastoreLevel3Client) Get(uuid string, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel3, error) {
	var obj *models.TestSeDatastoreLevel3
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing TestSeDatastoreLevel3 by name
func (client *TestSeDatastoreLevel3Client) GetByName(name string, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel3, error) {
	var obj *models.TestSeDatastoreLevel3
	err := client.aviSession.GetObjectByName("testsedatastorelevel3", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing TestSeDatastoreLevel3 by filters like name, cloud, tenant
// Api creates TestSeDatastoreLevel3 object with every call.
func (client *TestSeDatastoreLevel3Client) GetObject(options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel3, error) {
	var obj *models.TestSeDatastoreLevel3
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("testsedatastorelevel3", newOptions...)
	return obj, err
}

// Create a new TestSeDatastoreLevel3 object
func (client *TestSeDatastoreLevel3Client) Create(obj *models.TestSeDatastoreLevel3, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel3, error) {
	var robj *models.TestSeDatastoreLevel3
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing TestSeDatastoreLevel3 object
func (client *TestSeDatastoreLevel3Client) Update(obj *models.TestSeDatastoreLevel3, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel3, error) {
	var robj *models.TestSeDatastoreLevel3
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing TestSeDatastoreLevel3 object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.TestSeDatastoreLevel3
// or it should be json compatible of form map[string]interface{}
func (client *TestSeDatastoreLevel3Client) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.TestSeDatastoreLevel3, error) {
	var robj *models.TestSeDatastoreLevel3
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing TestSeDatastoreLevel3 object with a given UUID
func (client *TestSeDatastoreLevel3Client) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing TestSeDatastoreLevel3 object with a given name
func (client *TestSeDatastoreLevel3Client) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *TestSeDatastoreLevel3Client) GetAviSession() *session.AviSession {
	return client.aviSession
}
