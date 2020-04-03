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

// WafPolicyPSMGroupClient is a client for avi WafPolicyPSMGroup resource
type WafPolicyPSMGroupClient struct {
	aviSession *session.AviSession
}

// NewWafPolicyPSMGroupClient creates a new client for WafPolicyPSMGroup resource
func NewWafPolicyPSMGroupClient(aviSession *session.AviSession) *WafPolicyPSMGroupClient {
	return &WafPolicyPSMGroupClient{aviSession: aviSession}
}

func (client *WafPolicyPSMGroupClient) getAPIPath(uuid string) string {
	path := "api/wafpolicypsmgroup"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of WafPolicyPSMGroup objects
func (client *WafPolicyPSMGroupClient) GetAll(options ...session.ApiOptionsParams) ([]*models.WafPolicyPSMGroup, error) {
	var plist []*models.WafPolicyPSMGroup
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing WafPolicyPSMGroup by uuid
func (client *WafPolicyPSMGroupClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.WafPolicyPSMGroup, error) {
	var obj *models.WafPolicyPSMGroup
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing WafPolicyPSMGroup by name
func (client *WafPolicyPSMGroupClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.WafPolicyPSMGroup, error) {
	var obj *models.WafPolicyPSMGroup
	err := client.aviSession.GetObjectByName("wafpolicypsmgroup", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing WafPolicyPSMGroup by filters like name, cloud, tenant
// Api creates WafPolicyPSMGroup object with every call.
func (client *WafPolicyPSMGroupClient) GetObject(options ...session.ApiOptionsParams) (*models.WafPolicyPSMGroup, error) {
	var obj *models.WafPolicyPSMGroup
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("wafpolicypsmgroup", newOptions...)
	return obj, err
}

// Create a new WafPolicyPSMGroup object
func (client *WafPolicyPSMGroupClient) Create(obj *models.WafPolicyPSMGroup, options ...session.ApiOptionsParams) (*models.WafPolicyPSMGroup, error) {
	var robj *models.WafPolicyPSMGroup
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing WafPolicyPSMGroup object
func (client *WafPolicyPSMGroupClient) Update(obj *models.WafPolicyPSMGroup, options ...session.ApiOptionsParams) (*models.WafPolicyPSMGroup, error) {
	var robj *models.WafPolicyPSMGroup
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing WafPolicyPSMGroup object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.WafPolicyPSMGroup
// or it should be json compatible of form map[string]interface{}
func (client *WafPolicyPSMGroupClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.WafPolicyPSMGroup, error) {
	var robj *models.WafPolicyPSMGroup
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing WafPolicyPSMGroup object with a given UUID
func (client *WafPolicyPSMGroupClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing WafPolicyPSMGroup object with a given name
func (client *WafPolicyPSMGroupClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *WafPolicyPSMGroupClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
