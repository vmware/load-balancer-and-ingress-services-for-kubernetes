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

// NetworkSecurityPolicyClient is a client for avi NetworkSecurityPolicy resource
type NetworkSecurityPolicyClient struct {
	aviSession *session.AviSession
}

// NewNetworkSecurityPolicyClient creates a new client for NetworkSecurityPolicy resource
func NewNetworkSecurityPolicyClient(aviSession *session.AviSession) *NetworkSecurityPolicyClient {
	return &NetworkSecurityPolicyClient{aviSession: aviSession}
}

func (client *NetworkSecurityPolicyClient) getAPIPath(uuid string) string {
	path := "api/networksecuritypolicy"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of NetworkSecurityPolicy objects
func (client *NetworkSecurityPolicyClient) GetAll(options ...session.ApiOptionsParams) ([]*models.NetworkSecurityPolicy, error) {
	var plist []*models.NetworkSecurityPolicy
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing NetworkSecurityPolicy by uuid
func (client *NetworkSecurityPolicyClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.NetworkSecurityPolicy, error) {
	var obj *models.NetworkSecurityPolicy
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing NetworkSecurityPolicy by name
func (client *NetworkSecurityPolicyClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.NetworkSecurityPolicy, error) {
	var obj *models.NetworkSecurityPolicy
	err := client.aviSession.GetObjectByName("networksecuritypolicy", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing NetworkSecurityPolicy by filters like name, cloud, tenant
// Api creates NetworkSecurityPolicy object with every call.
func (client *NetworkSecurityPolicyClient) GetObject(options ...session.ApiOptionsParams) (*models.NetworkSecurityPolicy, error) {
	var obj *models.NetworkSecurityPolicy
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("networksecuritypolicy", newOptions...)
	return obj, err
}

// Create a new NetworkSecurityPolicy object
func (client *NetworkSecurityPolicyClient) Create(obj *models.NetworkSecurityPolicy, options ...session.ApiOptionsParams) (*models.NetworkSecurityPolicy, error) {
	var robj *models.NetworkSecurityPolicy
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing NetworkSecurityPolicy object
func (client *NetworkSecurityPolicyClient) Update(obj *models.NetworkSecurityPolicy, options ...session.ApiOptionsParams) (*models.NetworkSecurityPolicy, error) {
	var robj *models.NetworkSecurityPolicy
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing NetworkSecurityPolicy object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.NetworkSecurityPolicy
// or it should be json compatible of form map[string]interface{}
func (client *NetworkSecurityPolicyClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.NetworkSecurityPolicy, error) {
	var robj *models.NetworkSecurityPolicy
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing NetworkSecurityPolicy object with a given UUID
func (client *NetworkSecurityPolicyClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing NetworkSecurityPolicy object with a given name
func (client *NetworkSecurityPolicyClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *NetworkSecurityPolicyClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
