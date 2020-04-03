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

// VIPGNameInfoClient is a client for avi VIPGNameInfo resource
type VIPGNameInfoClient struct {
	aviSession *session.AviSession
}

// NewVIPGNameInfoClient creates a new client for VIPGNameInfo resource
func NewVIPGNameInfoClient(aviSession *session.AviSession) *VIPGNameInfoClient {
	return &VIPGNameInfoClient{aviSession: aviSession}
}

func (client *VIPGNameInfoClient) getAPIPath(uuid string) string {
	path := "api/vipgnameinfo"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VIPGNameInfo objects
func (client *VIPGNameInfoClient) GetAll(options ...session.ApiOptionsParams) ([]*models.VIPGNameInfo, error) {
	var plist []*models.VIPGNameInfo
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing VIPGNameInfo by uuid
func (client *VIPGNameInfoClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.VIPGNameInfo, error) {
	var obj *models.VIPGNameInfo
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing VIPGNameInfo by name
func (client *VIPGNameInfoClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.VIPGNameInfo, error) {
	var obj *models.VIPGNameInfo
	err := client.aviSession.GetObjectByName("vipgnameinfo", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing VIPGNameInfo by filters like name, cloud, tenant
// Api creates VIPGNameInfo object with every call.
func (client *VIPGNameInfoClient) GetObject(options ...session.ApiOptionsParams) (*models.VIPGNameInfo, error) {
	var obj *models.VIPGNameInfo
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vipgnameinfo", newOptions...)
	return obj, err
}

// Create a new VIPGNameInfo object
func (client *VIPGNameInfoClient) Create(obj *models.VIPGNameInfo, options ...session.ApiOptionsParams) (*models.VIPGNameInfo, error) {
	var robj *models.VIPGNameInfo
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing VIPGNameInfo object
func (client *VIPGNameInfoClient) Update(obj *models.VIPGNameInfo, options ...session.ApiOptionsParams) (*models.VIPGNameInfo, error) {
	var robj *models.VIPGNameInfo
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing VIPGNameInfo object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VIPGNameInfo
// or it should be json compatible of form map[string]interface{}
func (client *VIPGNameInfoClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.VIPGNameInfo, error) {
	var robj *models.VIPGNameInfo
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing VIPGNameInfo object with a given UUID
func (client *VIPGNameInfoClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing VIPGNameInfo object with a given name
func (client *VIPGNameInfoClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *VIPGNameInfoClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
