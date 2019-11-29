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

// SecurityPolicyClient is a client for avi SecurityPolicy resource
type SecurityPolicyClient struct {
	aviSession *session.AviSession
}

// NewSecurityPolicyClient creates a new client for SecurityPolicy resource
func NewSecurityPolicyClient(aviSession *session.AviSession) *SecurityPolicyClient {
	return &SecurityPolicyClient{aviSession: aviSession}
}

func (client *SecurityPolicyClient) getAPIPath(uuid string) string {
	path := "api/securitypolicy"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of SecurityPolicy objects
func (client *SecurityPolicyClient) GetAll() ([]*models.SecurityPolicy, error) {
	var plist []*models.SecurityPolicy
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing SecurityPolicy by uuid
func (client *SecurityPolicyClient) Get(uuid string) (*models.SecurityPolicy, error) {
	var obj *models.SecurityPolicy
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing SecurityPolicy by name
func (client *SecurityPolicyClient) GetByName(name string) (*models.SecurityPolicy, error) {
	var obj *models.SecurityPolicy
	err := client.aviSession.GetObjectByName("securitypolicy", name, &obj)
	return obj, err
}

// GetObject - Get an existing SecurityPolicy by filters like name, cloud, tenant
// Api creates SecurityPolicy object with every call.
func (client *SecurityPolicyClient) GetObject(options ...session.ApiOptionsParams) (*models.SecurityPolicy, error) {
	var obj *models.SecurityPolicy
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("securitypolicy", newOptions...)
	return obj, err
}

// Create a new SecurityPolicy object
func (client *SecurityPolicyClient) Create(obj *models.SecurityPolicy) (*models.SecurityPolicy, error) {
	var robj *models.SecurityPolicy
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing SecurityPolicy object
func (client *SecurityPolicyClient) Update(obj *models.SecurityPolicy) (*models.SecurityPolicy, error) {
	var robj *models.SecurityPolicy
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing SecurityPolicy object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.SecurityPolicy
// or it should be json compatible of form map[string]interface{}
func (client *SecurityPolicyClient) Patch(uuid string, patch interface{}, patchOp string) (*models.SecurityPolicy, error) {
	var robj *models.SecurityPolicy
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing SecurityPolicy object with a given UUID
func (client *SecurityPolicyClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing SecurityPolicy object with a given name
func (client *SecurityPolicyClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *SecurityPolicyClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
