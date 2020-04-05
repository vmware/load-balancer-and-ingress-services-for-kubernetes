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

// AlertSyslogConfigClient is a client for avi AlertSyslogConfig resource
type AlertSyslogConfigClient struct {
	aviSession *session.AviSession
}

// NewAlertSyslogConfigClient creates a new client for AlertSyslogConfig resource
func NewAlertSyslogConfigClient(aviSession *session.AviSession) *AlertSyslogConfigClient {
	return &AlertSyslogConfigClient{aviSession: aviSession}
}

func (client *AlertSyslogConfigClient) getAPIPath(uuid string) string {
	path := "api/alertsyslogconfig"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of AlertSyslogConfig objects
func (client *AlertSyslogConfigClient) GetAll(options ...session.ApiOptionsParams) ([]*models.AlertSyslogConfig, error) {
	var plist []*models.AlertSyslogConfig
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing AlertSyslogConfig by uuid
func (client *AlertSyslogConfigClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.AlertSyslogConfig, error) {
	var obj *models.AlertSyslogConfig
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing AlertSyslogConfig by name
func (client *AlertSyslogConfigClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.AlertSyslogConfig, error) {
	var obj *models.AlertSyslogConfig
	err := client.aviSession.GetObjectByName("alertsyslogconfig", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing AlertSyslogConfig by filters like name, cloud, tenant
// Api creates AlertSyslogConfig object with every call.
func (client *AlertSyslogConfigClient) GetObject(options ...session.ApiOptionsParams) (*models.AlertSyslogConfig, error) {
	var obj *models.AlertSyslogConfig
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("alertsyslogconfig", newOptions...)
	return obj, err
}

// Create a new AlertSyslogConfig object
func (client *AlertSyslogConfigClient) Create(obj *models.AlertSyslogConfig, options ...session.ApiOptionsParams) (*models.AlertSyslogConfig, error) {
	var robj *models.AlertSyslogConfig
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing AlertSyslogConfig object
func (client *AlertSyslogConfigClient) Update(obj *models.AlertSyslogConfig, options ...session.ApiOptionsParams) (*models.AlertSyslogConfig, error) {
	var robj *models.AlertSyslogConfig
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing AlertSyslogConfig object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.AlertSyslogConfig
// or it should be json compatible of form map[string]interface{}
func (client *AlertSyslogConfigClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.AlertSyslogConfig, error) {
	var robj *models.AlertSyslogConfig
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing AlertSyslogConfig object with a given UUID
func (client *AlertSyslogConfigClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing AlertSyslogConfig object with a given name
func (client *AlertSyslogConfigClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *AlertSyslogConfigClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
