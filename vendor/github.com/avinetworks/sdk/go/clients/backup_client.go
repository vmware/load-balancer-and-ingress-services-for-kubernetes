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

// BackupClient is a client for avi Backup resource
type BackupClient struct {
	aviSession *session.AviSession
}

// NewBackupClient creates a new client for Backup resource
func NewBackupClient(aviSession *session.AviSession) *BackupClient {
	return &BackupClient{aviSession: aviSession}
}

func (client *BackupClient) getAPIPath(uuid string) string {
	path := "api/backup"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of Backup objects
func (client *BackupClient) GetAll(options ...session.ApiOptionsParams) ([]*models.Backup, error) {
	var plist []*models.Backup
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing Backup by uuid
func (client *BackupClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.Backup, error) {
	var obj *models.Backup
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing Backup by name
func (client *BackupClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.Backup, error) {
	var obj *models.Backup
	err := client.aviSession.GetObjectByName("backup", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing Backup by filters like name, cloud, tenant
// Api creates Backup object with every call.
func (client *BackupClient) GetObject(options ...session.ApiOptionsParams) (*models.Backup, error) {
	var obj *models.Backup
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("backup", newOptions...)
	return obj, err
}

// Create a new Backup object
func (client *BackupClient) Create(obj *models.Backup, options ...session.ApiOptionsParams) (*models.Backup, error) {
	var robj *models.Backup
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing Backup object
func (client *BackupClient) Update(obj *models.Backup, options ...session.ApiOptionsParams) (*models.Backup, error) {
	var robj *models.Backup
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing Backup object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.Backup
// or it should be json compatible of form map[string]interface{}
func (client *BackupClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.Backup, error) {
	var robj *models.Backup
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing Backup object with a given UUID
func (client *BackupClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing Backup object with a given name
func (client *BackupClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *BackupClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
