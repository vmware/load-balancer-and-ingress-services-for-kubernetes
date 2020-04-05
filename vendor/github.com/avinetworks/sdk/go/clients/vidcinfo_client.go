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

// VIDCInfoClient is a client for avi VIDCInfo resource
type VIDCInfoClient struct {
	aviSession *session.AviSession
}

// NewVIDCInfoClient creates a new client for VIDCInfo resource
func NewVIDCInfoClient(aviSession *session.AviSession) *VIDCInfoClient {
	return &VIDCInfoClient{aviSession: aviSession}
}

func (client *VIDCInfoClient) getAPIPath(uuid string) string {
	path := "api/vidcinfo"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VIDCInfo objects
func (client *VIDCInfoClient) GetAll(options ...session.ApiOptionsParams) ([]*models.VIDCInfo, error) {
	var plist []*models.VIDCInfo
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing VIDCInfo by uuid
func (client *VIDCInfoClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.VIDCInfo, error) {
	var obj *models.VIDCInfo
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing VIDCInfo by name
func (client *VIDCInfoClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.VIDCInfo, error) {
	var obj *models.VIDCInfo
	err := client.aviSession.GetObjectByName("vidcinfo", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing VIDCInfo by filters like name, cloud, tenant
// Api creates VIDCInfo object with every call.
func (client *VIDCInfoClient) GetObject(options ...session.ApiOptionsParams) (*models.VIDCInfo, error) {
	var obj *models.VIDCInfo
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vidcinfo", newOptions...)
	return obj, err
}

// Create a new VIDCInfo object
func (client *VIDCInfoClient) Create(obj *models.VIDCInfo, options ...session.ApiOptionsParams) (*models.VIDCInfo, error) {
	var robj *models.VIDCInfo
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing VIDCInfo object
func (client *VIDCInfoClient) Update(obj *models.VIDCInfo, options ...session.ApiOptionsParams) (*models.VIDCInfo, error) {
	var robj *models.VIDCInfo
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing VIDCInfo object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VIDCInfo
// or it should be json compatible of form map[string]interface{}
func (client *VIDCInfoClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.VIDCInfo, error) {
	var robj *models.VIDCInfo
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing VIDCInfo object with a given UUID
func (client *VIDCInfoClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing VIDCInfo object with a given name
func (client *VIDCInfoClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *VIDCInfoClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
