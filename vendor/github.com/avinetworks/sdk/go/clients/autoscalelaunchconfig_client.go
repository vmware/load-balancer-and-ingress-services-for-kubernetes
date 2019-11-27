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

// AutoScaleLaunchConfigClient is a client for avi AutoScaleLaunchConfig resource
type AutoScaleLaunchConfigClient struct {
	aviSession *session.AviSession
}

// NewAutoScaleLaunchConfigClient creates a new client for AutoScaleLaunchConfig resource
func NewAutoScaleLaunchConfigClient(aviSession *session.AviSession) *AutoScaleLaunchConfigClient {
	return &AutoScaleLaunchConfigClient{aviSession: aviSession}
}

func (client *AutoScaleLaunchConfigClient) getAPIPath(uuid string) string {
	path := "api/autoscalelaunchconfig"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of AutoScaleLaunchConfig objects
func (client *AutoScaleLaunchConfigClient) GetAll() ([]*models.AutoScaleLaunchConfig, error) {
	var plist []*models.AutoScaleLaunchConfig
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing AutoScaleLaunchConfig by uuid
func (client *AutoScaleLaunchConfigClient) Get(uuid string) (*models.AutoScaleLaunchConfig, error) {
	var obj *models.AutoScaleLaunchConfig
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing AutoScaleLaunchConfig by name
func (client *AutoScaleLaunchConfigClient) GetByName(name string) (*models.AutoScaleLaunchConfig, error) {
	var obj *models.AutoScaleLaunchConfig
	err := client.aviSession.GetObjectByName("autoscalelaunchconfig", name, &obj)
	return obj, err
}

// GetObject - Get an existing AutoScaleLaunchConfig by filters like name, cloud, tenant
// Api creates AutoScaleLaunchConfig object with every call.
func (client *AutoScaleLaunchConfigClient) GetObject(options ...session.ApiOptionsParams) (*models.AutoScaleLaunchConfig, error) {
	var obj *models.AutoScaleLaunchConfig
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("autoscalelaunchconfig", newOptions...)
	return obj, err
}

// Create a new AutoScaleLaunchConfig object
func (client *AutoScaleLaunchConfigClient) Create(obj *models.AutoScaleLaunchConfig) (*models.AutoScaleLaunchConfig, error) {
	var robj *models.AutoScaleLaunchConfig
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing AutoScaleLaunchConfig object
func (client *AutoScaleLaunchConfigClient) Update(obj *models.AutoScaleLaunchConfig) (*models.AutoScaleLaunchConfig, error) {
	var robj *models.AutoScaleLaunchConfig
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing AutoScaleLaunchConfig object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.AutoScaleLaunchConfig
// or it should be json compatible of form map[string]interface{}
func (client *AutoScaleLaunchConfigClient) Patch(uuid string, patch interface{}, patchOp string) (*models.AutoScaleLaunchConfig, error) {
	var robj *models.AutoScaleLaunchConfig
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing AutoScaleLaunchConfig object with a given UUID
func (client *AutoScaleLaunchConfigClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing AutoScaleLaunchConfig object with a given name
func (client *AutoScaleLaunchConfigClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *AutoScaleLaunchConfigClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
