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

// AlertConfigClient is a client for avi AlertConfig resource
type AlertConfigClient struct {
	aviSession *session.AviSession
}

// NewAlertConfigClient creates a new client for AlertConfig resource
func NewAlertConfigClient(aviSession *session.AviSession) *AlertConfigClient {
	return &AlertConfigClient{aviSession: aviSession}
}

func (client *AlertConfigClient) getAPIPath(uuid string) string {
	path := "api/alertconfig"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of AlertConfig objects
func (client *AlertConfigClient) GetAll(options ...session.ApiOptionsParams) ([]*models.AlertConfig, error) {
	var plist []*models.AlertConfig
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing AlertConfig by uuid
func (client *AlertConfigClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.AlertConfig, error) {
	var obj *models.AlertConfig
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing AlertConfig by name
func (client *AlertConfigClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.AlertConfig, error) {
	var obj *models.AlertConfig
	err := client.aviSession.GetObjectByName("alertconfig", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing AlertConfig by filters like name, cloud, tenant
// Api creates AlertConfig object with every call.
func (client *AlertConfigClient) GetObject(options ...session.ApiOptionsParams) (*models.AlertConfig, error) {
	var obj *models.AlertConfig
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("alertconfig", newOptions...)
	return obj, err
}

// Create a new AlertConfig object
func (client *AlertConfigClient) Create(obj *models.AlertConfig, options ...session.ApiOptionsParams) (*models.AlertConfig, error) {
	var robj *models.AlertConfig
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing AlertConfig object
func (client *AlertConfigClient) Update(obj *models.AlertConfig, options ...session.ApiOptionsParams) (*models.AlertConfig, error) {
	var robj *models.AlertConfig
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing AlertConfig object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.AlertConfig
// or it should be json compatible of form map[string]interface{}
func (client *AlertConfigClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.AlertConfig, error) {
	var robj *models.AlertConfig
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing AlertConfig object with a given UUID
func (client *AlertConfigClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing AlertConfig object with a given name
func (client *AlertConfigClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *AlertConfigClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
