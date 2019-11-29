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

// PoolGroupDeploymentPolicyClient is a client for avi PoolGroupDeploymentPolicy resource
type PoolGroupDeploymentPolicyClient struct {
	aviSession *session.AviSession
}

// NewPoolGroupDeploymentPolicyClient creates a new client for PoolGroupDeploymentPolicy resource
func NewPoolGroupDeploymentPolicyClient(aviSession *session.AviSession) *PoolGroupDeploymentPolicyClient {
	return &PoolGroupDeploymentPolicyClient{aviSession: aviSession}
}

func (client *PoolGroupDeploymentPolicyClient) getAPIPath(uuid string) string {
	path := "api/poolgroupdeploymentpolicy"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of PoolGroupDeploymentPolicy objects
func (client *PoolGroupDeploymentPolicyClient) GetAll() ([]*models.PoolGroupDeploymentPolicy, error) {
	var plist []*models.PoolGroupDeploymentPolicy
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing PoolGroupDeploymentPolicy by uuid
func (client *PoolGroupDeploymentPolicyClient) Get(uuid string) (*models.PoolGroupDeploymentPolicy, error) {
	var obj *models.PoolGroupDeploymentPolicy
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing PoolGroupDeploymentPolicy by name
func (client *PoolGroupDeploymentPolicyClient) GetByName(name string) (*models.PoolGroupDeploymentPolicy, error) {
	var obj *models.PoolGroupDeploymentPolicy
	err := client.aviSession.GetObjectByName("poolgroupdeploymentpolicy", name, &obj)
	return obj, err
}

// GetObject - Get an existing PoolGroupDeploymentPolicy by filters like name, cloud, tenant
// Api creates PoolGroupDeploymentPolicy object with every call.
func (client *PoolGroupDeploymentPolicyClient) GetObject(options ...session.ApiOptionsParams) (*models.PoolGroupDeploymentPolicy, error) {
	var obj *models.PoolGroupDeploymentPolicy
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("poolgroupdeploymentpolicy", newOptions...)
	return obj, err
}

// Create a new PoolGroupDeploymentPolicy object
func (client *PoolGroupDeploymentPolicyClient) Create(obj *models.PoolGroupDeploymentPolicy) (*models.PoolGroupDeploymentPolicy, error) {
	var robj *models.PoolGroupDeploymentPolicy
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing PoolGroupDeploymentPolicy object
func (client *PoolGroupDeploymentPolicyClient) Update(obj *models.PoolGroupDeploymentPolicy) (*models.PoolGroupDeploymentPolicy, error) {
	var robj *models.PoolGroupDeploymentPolicy
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing PoolGroupDeploymentPolicy object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.PoolGroupDeploymentPolicy
// or it should be json compatible of form map[string]interface{}
func (client *PoolGroupDeploymentPolicyClient) Patch(uuid string, patch interface{}, patchOp string) (*models.PoolGroupDeploymentPolicy, error) {
	var robj *models.PoolGroupDeploymentPolicy
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing PoolGroupDeploymentPolicy object with a given UUID
func (client *PoolGroupDeploymentPolicyClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing PoolGroupDeploymentPolicy object with a given name
func (client *PoolGroupDeploymentPolicyClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *PoolGroupDeploymentPolicyClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
