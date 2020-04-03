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

// UpgradeStatusInfoClient is a client for avi UpgradeStatusInfo resource
type UpgradeStatusInfoClient struct {
	aviSession *session.AviSession
}

// NewUpgradeStatusInfoClient creates a new client for UpgradeStatusInfo resource
func NewUpgradeStatusInfoClient(aviSession *session.AviSession) *UpgradeStatusInfoClient {
	return &UpgradeStatusInfoClient{aviSession: aviSession}
}

func (client *UpgradeStatusInfoClient) getAPIPath(uuid string) string {
	path := "api/upgradestatusinfo"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of UpgradeStatusInfo objects
func (client *UpgradeStatusInfoClient) GetAll(options ...session.ApiOptionsParams) ([]*models.UpgradeStatusInfo, error) {
	var plist []*models.UpgradeStatusInfo
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing UpgradeStatusInfo by uuid
func (client *UpgradeStatusInfoClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.UpgradeStatusInfo, error) {
	var obj *models.UpgradeStatusInfo
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing UpgradeStatusInfo by name
func (client *UpgradeStatusInfoClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.UpgradeStatusInfo, error) {
	var obj *models.UpgradeStatusInfo
	err := client.aviSession.GetObjectByName("upgradestatusinfo", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing UpgradeStatusInfo by filters like name, cloud, tenant
// Api creates UpgradeStatusInfo object with every call.
func (client *UpgradeStatusInfoClient) GetObject(options ...session.ApiOptionsParams) (*models.UpgradeStatusInfo, error) {
	var obj *models.UpgradeStatusInfo
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("upgradestatusinfo", newOptions...)
	return obj, err
}

// Create a new UpgradeStatusInfo object
func (client *UpgradeStatusInfoClient) Create(obj *models.UpgradeStatusInfo, options ...session.ApiOptionsParams) (*models.UpgradeStatusInfo, error) {
	var robj *models.UpgradeStatusInfo
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing UpgradeStatusInfo object
func (client *UpgradeStatusInfoClient) Update(obj *models.UpgradeStatusInfo, options ...session.ApiOptionsParams) (*models.UpgradeStatusInfo, error) {
	var robj *models.UpgradeStatusInfo
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing UpgradeStatusInfo object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.UpgradeStatusInfo
// or it should be json compatible of form map[string]interface{}
func (client *UpgradeStatusInfoClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.UpgradeStatusInfo, error) {
	var robj *models.UpgradeStatusInfo
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing UpgradeStatusInfo object with a given UUID
func (client *UpgradeStatusInfoClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing UpgradeStatusInfo object with a given name
func (client *UpgradeStatusInfoClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *UpgradeStatusInfoClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
