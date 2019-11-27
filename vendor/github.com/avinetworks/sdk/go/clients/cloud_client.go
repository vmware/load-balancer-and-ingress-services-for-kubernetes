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

// CloudClient is a client for avi Cloud resource
type CloudClient struct {
	aviSession *session.AviSession
}

// NewCloudClient creates a new client for Cloud resource
func NewCloudClient(aviSession *session.AviSession) *CloudClient {
	return &CloudClient{aviSession: aviSession}
}

func (client *CloudClient) getAPIPath(uuid string) string {
	path := "api/cloud"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of Cloud objects
func (client *CloudClient) GetAll() ([]*models.Cloud, error) {
	var plist []*models.Cloud
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing Cloud by uuid
func (client *CloudClient) Get(uuid string) (*models.Cloud, error) {
	var obj *models.Cloud
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing Cloud by name
func (client *CloudClient) GetByName(name string) (*models.Cloud, error) {
	var obj *models.Cloud
	err := client.aviSession.GetObjectByName("cloud", name, &obj)
	return obj, err
}

// GetObject - Get an existing Cloud by filters like name, cloud, tenant
// Api creates Cloud object with every call.
func (client *CloudClient) GetObject(options ...session.ApiOptionsParams) (*models.Cloud, error) {
	var obj *models.Cloud
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("cloud", newOptions...)
	return obj, err
}

// Create a new Cloud object
func (client *CloudClient) Create(obj *models.Cloud) (*models.Cloud, error) {
	var robj *models.Cloud
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing Cloud object
func (client *CloudClient) Update(obj *models.Cloud) (*models.Cloud, error) {
	var robj *models.Cloud
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing Cloud object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.Cloud
// or it should be json compatible of form map[string]interface{}
func (client *CloudClient) Patch(uuid string, patch interface{}, patchOp string) (*models.Cloud, error) {
	var robj *models.Cloud
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing Cloud object with a given UUID
func (client *CloudClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing Cloud object with a given name
func (client *CloudClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *CloudClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
