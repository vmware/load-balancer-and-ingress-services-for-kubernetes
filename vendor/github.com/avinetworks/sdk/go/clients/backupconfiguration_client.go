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

// BackupConfigurationClient is a client for avi BackupConfiguration resource
type BackupConfigurationClient struct {
	aviSession *session.AviSession
}

// NewBackupConfigurationClient creates a new client for BackupConfiguration resource
func NewBackupConfigurationClient(aviSession *session.AviSession) *BackupConfigurationClient {
	return &BackupConfigurationClient{aviSession: aviSession}
}

func (client *BackupConfigurationClient) getAPIPath(uuid string) string {
	path := "api/backupconfiguration"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of BackupConfiguration objects
func (client *BackupConfigurationClient) GetAll(options ...session.ApiOptionsParams) ([]*models.BackupConfiguration, error) {
	var plist []*models.BackupConfiguration
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing BackupConfiguration by uuid
func (client *BackupConfigurationClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.BackupConfiguration, error) {
	var obj *models.BackupConfiguration
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing BackupConfiguration by name
func (client *BackupConfigurationClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.BackupConfiguration, error) {
	var obj *models.BackupConfiguration
	err := client.aviSession.GetObjectByName("backupconfiguration", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing BackupConfiguration by filters like name, cloud, tenant
// Api creates BackupConfiguration object with every call.
func (client *BackupConfigurationClient) GetObject(options ...session.ApiOptionsParams) (*models.BackupConfiguration, error) {
	var obj *models.BackupConfiguration
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("backupconfiguration", newOptions...)
	return obj, err
}

// Create a new BackupConfiguration object
func (client *BackupConfigurationClient) Create(obj *models.BackupConfiguration, options ...session.ApiOptionsParams) (*models.BackupConfiguration, error) {
	var robj *models.BackupConfiguration
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing BackupConfiguration object
func (client *BackupConfigurationClient) Update(obj *models.BackupConfiguration, options ...session.ApiOptionsParams) (*models.BackupConfiguration, error) {
	var robj *models.BackupConfiguration
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing BackupConfiguration object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.BackupConfiguration
// or it should be json compatible of form map[string]interface{}
func (client *BackupConfigurationClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.BackupConfiguration, error) {
	var robj *models.BackupConfiguration
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing BackupConfiguration object with a given UUID
func (client *BackupConfigurationClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing BackupConfiguration object with a given name
func (client *BackupConfigurationClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *BackupConfigurationClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
