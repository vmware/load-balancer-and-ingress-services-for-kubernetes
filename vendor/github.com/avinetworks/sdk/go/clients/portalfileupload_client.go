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

// PortalFileUploadClient is a client for avi PortalFileUpload resource
type PortalFileUploadClient struct {
	aviSession *session.AviSession
}

// NewPortalFileUploadClient creates a new client for PortalFileUpload resource
func NewPortalFileUploadClient(aviSession *session.AviSession) *PortalFileUploadClient {
	return &PortalFileUploadClient{aviSession: aviSession}
}

func (client *PortalFileUploadClient) getAPIPath(uuid string) string {
	path := "api/portalfileupload"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of PortalFileUpload objects
func (client *PortalFileUploadClient) GetAll(options ...session.ApiOptionsParams) ([]*models.PortalFileUpload, error) {
	var plist []*models.PortalFileUpload
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing PortalFileUpload by uuid
func (client *PortalFileUploadClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.PortalFileUpload, error) {
	var obj *models.PortalFileUpload
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing PortalFileUpload by name
func (client *PortalFileUploadClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.PortalFileUpload, error) {
	var obj *models.PortalFileUpload
	err := client.aviSession.GetObjectByName("portalfileupload", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing PortalFileUpload by filters like name, cloud, tenant
// Api creates PortalFileUpload object with every call.
func (client *PortalFileUploadClient) GetObject(options ...session.ApiOptionsParams) (*models.PortalFileUpload, error) {
	var obj *models.PortalFileUpload
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("portalfileupload", newOptions...)
	return obj, err
}

// Create a new PortalFileUpload object
func (client *PortalFileUploadClient) Create(obj *models.PortalFileUpload, options ...session.ApiOptionsParams) (*models.PortalFileUpload, error) {
	var robj *models.PortalFileUpload
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing PortalFileUpload object
func (client *PortalFileUploadClient) Update(obj *models.PortalFileUpload, options ...session.ApiOptionsParams) (*models.PortalFileUpload, error) {
	var robj *models.PortalFileUpload
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing PortalFileUpload object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.PortalFileUpload
// or it should be json compatible of form map[string]interface{}
func (client *PortalFileUploadClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.PortalFileUpload, error) {
	var robj *models.PortalFileUpload
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing PortalFileUpload object with a given UUID
func (client *PortalFileUploadClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing PortalFileUpload object with a given name
func (client *PortalFileUploadClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *PortalFileUploadClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
