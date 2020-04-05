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

// AlertEmailConfigClient is a client for avi AlertEmailConfig resource
type AlertEmailConfigClient struct {
	aviSession *session.AviSession
}

// NewAlertEmailConfigClient creates a new client for AlertEmailConfig resource
func NewAlertEmailConfigClient(aviSession *session.AviSession) *AlertEmailConfigClient {
	return &AlertEmailConfigClient{aviSession: aviSession}
}

func (client *AlertEmailConfigClient) getAPIPath(uuid string) string {
	path := "api/alertemailconfig"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of AlertEmailConfig objects
func (client *AlertEmailConfigClient) GetAll(options ...session.ApiOptionsParams) ([]*models.AlertEmailConfig, error) {
	var plist []*models.AlertEmailConfig
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing AlertEmailConfig by uuid
func (client *AlertEmailConfigClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.AlertEmailConfig, error) {
	var obj *models.AlertEmailConfig
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing AlertEmailConfig by name
func (client *AlertEmailConfigClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.AlertEmailConfig, error) {
	var obj *models.AlertEmailConfig
	err := client.aviSession.GetObjectByName("alertemailconfig", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing AlertEmailConfig by filters like name, cloud, tenant
// Api creates AlertEmailConfig object with every call.
func (client *AlertEmailConfigClient) GetObject(options ...session.ApiOptionsParams) (*models.AlertEmailConfig, error) {
	var obj *models.AlertEmailConfig
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("alertemailconfig", newOptions...)
	return obj, err
}

// Create a new AlertEmailConfig object
func (client *AlertEmailConfigClient) Create(obj *models.AlertEmailConfig, options ...session.ApiOptionsParams) (*models.AlertEmailConfig, error) {
	var robj *models.AlertEmailConfig
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing AlertEmailConfig object
func (client *AlertEmailConfigClient) Update(obj *models.AlertEmailConfig, options ...session.ApiOptionsParams) (*models.AlertEmailConfig, error) {
	var robj *models.AlertEmailConfig
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing AlertEmailConfig object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.AlertEmailConfig
// or it should be json compatible of form map[string]interface{}
func (client *AlertEmailConfigClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.AlertEmailConfig, error) {
	var robj *models.AlertEmailConfig
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing AlertEmailConfig object with a given UUID
func (client *AlertEmailConfigClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing AlertEmailConfig object with a given name
func (client *AlertEmailConfigClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *AlertEmailConfigClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
