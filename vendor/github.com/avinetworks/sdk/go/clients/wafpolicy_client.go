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

// WafPolicyClient is a client for avi WafPolicy resource
type WafPolicyClient struct {
	aviSession *session.AviSession
}

// NewWafPolicyClient creates a new client for WafPolicy resource
func NewWafPolicyClient(aviSession *session.AviSession) *WafPolicyClient {
	return &WafPolicyClient{aviSession: aviSession}
}

func (client *WafPolicyClient) getAPIPath(uuid string) string {
	path := "api/wafpolicy"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of WafPolicy objects
func (client *WafPolicyClient) GetAll(options ...session.ApiOptionsParams) ([]*models.WafPolicy, error) {
	var plist []*models.WafPolicy
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing WafPolicy by uuid
func (client *WafPolicyClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.WafPolicy, error) {
	var obj *models.WafPolicy
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing WafPolicy by name
func (client *WafPolicyClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.WafPolicy, error) {
	var obj *models.WafPolicy
	err := client.aviSession.GetObjectByName("wafpolicy", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing WafPolicy by filters like name, cloud, tenant
// Api creates WafPolicy object with every call.
func (client *WafPolicyClient) GetObject(options ...session.ApiOptionsParams) (*models.WafPolicy, error) {
	var obj *models.WafPolicy
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("wafpolicy", newOptions...)
	return obj, err
}

// Create a new WafPolicy object
func (client *WafPolicyClient) Create(obj *models.WafPolicy, options ...session.ApiOptionsParams) (*models.WafPolicy, error) {
	var robj *models.WafPolicy
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing WafPolicy object
func (client *WafPolicyClient) Update(obj *models.WafPolicy, options ...session.ApiOptionsParams) (*models.WafPolicy, error) {
	var robj *models.WafPolicy
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing WafPolicy object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.WafPolicy
// or it should be json compatible of form map[string]interface{}
func (client *WafPolicyClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.WafPolicy, error) {
	var robj *models.WafPolicy
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing WafPolicy object with a given UUID
func (client *WafPolicyClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing WafPolicy object with a given name
func (client *WafPolicyClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *WafPolicyClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
