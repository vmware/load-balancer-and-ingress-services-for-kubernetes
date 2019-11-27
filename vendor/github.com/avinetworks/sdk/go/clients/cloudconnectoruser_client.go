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

// CloudConnectorUserClient is a client for avi CloudConnectorUser resource
type CloudConnectorUserClient struct {
	aviSession *session.AviSession
}

// NewCloudConnectorUserClient creates a new client for CloudConnectorUser resource
func NewCloudConnectorUserClient(aviSession *session.AviSession) *CloudConnectorUserClient {
	return &CloudConnectorUserClient{aviSession: aviSession}
}

func (client *CloudConnectorUserClient) getAPIPath(uuid string) string {
	path := "api/cloudconnectoruser"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of CloudConnectorUser objects
func (client *CloudConnectorUserClient) GetAll() ([]*models.CloudConnectorUser, error) {
	var plist []*models.CloudConnectorUser
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing CloudConnectorUser by uuid
func (client *CloudConnectorUserClient) Get(uuid string) (*models.CloudConnectorUser, error) {
	var obj *models.CloudConnectorUser
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing CloudConnectorUser by name
func (client *CloudConnectorUserClient) GetByName(name string) (*models.CloudConnectorUser, error) {
	var obj *models.CloudConnectorUser
	err := client.aviSession.GetObjectByName("cloudconnectoruser", name, &obj)
	return obj, err
}

// GetObject - Get an existing CloudConnectorUser by filters like name, cloud, tenant
// Api creates CloudConnectorUser object with every call.
func (client *CloudConnectorUserClient) GetObject(options ...session.ApiOptionsParams) (*models.CloudConnectorUser, error) {
	var obj *models.CloudConnectorUser
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("cloudconnectoruser", newOptions...)
	return obj, err
}

// Create a new CloudConnectorUser object
func (client *CloudConnectorUserClient) Create(obj *models.CloudConnectorUser) (*models.CloudConnectorUser, error) {
	var robj *models.CloudConnectorUser
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing CloudConnectorUser object
func (client *CloudConnectorUserClient) Update(obj *models.CloudConnectorUser) (*models.CloudConnectorUser, error) {
	var robj *models.CloudConnectorUser
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing CloudConnectorUser object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.CloudConnectorUser
// or it should be json compatible of form map[string]interface{}
func (client *CloudConnectorUserClient) Patch(uuid string, patch interface{}, patchOp string) (*models.CloudConnectorUser, error) {
	var robj *models.CloudConnectorUser
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing CloudConnectorUser object with a given UUID
func (client *CloudConnectorUserClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing CloudConnectorUser object with a given name
func (client *CloudConnectorUserClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *CloudConnectorUserClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
