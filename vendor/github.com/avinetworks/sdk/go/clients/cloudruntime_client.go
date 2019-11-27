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

// CloudRuntimeClient is a client for avi CloudRuntime resource
type CloudRuntimeClient struct {
	aviSession *session.AviSession
}

// NewCloudRuntimeClient creates a new client for CloudRuntime resource
func NewCloudRuntimeClient(aviSession *session.AviSession) *CloudRuntimeClient {
	return &CloudRuntimeClient{aviSession: aviSession}
}

func (client *CloudRuntimeClient) getAPIPath(uuid string) string {
	path := "api/cloudruntime"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of CloudRuntime objects
func (client *CloudRuntimeClient) GetAll() ([]*models.CloudRuntime, error) {
	var plist []*models.CloudRuntime
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing CloudRuntime by uuid
func (client *CloudRuntimeClient) Get(uuid string) (*models.CloudRuntime, error) {
	var obj *models.CloudRuntime
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing CloudRuntime by name
func (client *CloudRuntimeClient) GetByName(name string) (*models.CloudRuntime, error) {
	var obj *models.CloudRuntime
	err := client.aviSession.GetObjectByName("cloudruntime", name, &obj)
	return obj, err
}

// GetObject - Get an existing CloudRuntime by filters like name, cloud, tenant
// Api creates CloudRuntime object with every call.
func (client *CloudRuntimeClient) GetObject(options ...session.ApiOptionsParams) (*models.CloudRuntime, error) {
	var obj *models.CloudRuntime
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("cloudruntime", newOptions...)
	return obj, err
}

// Create a new CloudRuntime object
func (client *CloudRuntimeClient) Create(obj *models.CloudRuntime) (*models.CloudRuntime, error) {
	var robj *models.CloudRuntime
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing CloudRuntime object
func (client *CloudRuntimeClient) Update(obj *models.CloudRuntime) (*models.CloudRuntime, error) {
	var robj *models.CloudRuntime
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing CloudRuntime object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.CloudRuntime
// or it should be json compatible of form map[string]interface{}
func (client *CloudRuntimeClient) Patch(uuid string, patch interface{}, patchOp string) (*models.CloudRuntime, error) {
	var robj *models.CloudRuntime
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing CloudRuntime object with a given UUID
func (client *CloudRuntimeClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing CloudRuntime object with a given name
func (client *CloudRuntimeClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *CloudRuntimeClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
