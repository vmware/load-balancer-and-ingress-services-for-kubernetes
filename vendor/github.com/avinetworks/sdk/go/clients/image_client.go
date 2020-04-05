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

// ImageClient is a client for avi Image resource
type ImageClient struct {
	aviSession *session.AviSession
}

// NewImageClient creates a new client for Image resource
func NewImageClient(aviSession *session.AviSession) *ImageClient {
	return &ImageClient{aviSession: aviSession}
}

func (client *ImageClient) getAPIPath(uuid string) string {
	path := "api/image"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of Image objects
func (client *ImageClient) GetAll(options ...session.ApiOptionsParams) ([]*models.Image, error) {
	var plist []*models.Image
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing Image by uuid
func (client *ImageClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.Image, error) {
	var obj *models.Image
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing Image by name
func (client *ImageClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.Image, error) {
	var obj *models.Image
	err := client.aviSession.GetObjectByName("image", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing Image by filters like name, cloud, tenant
// Api creates Image object with every call.
func (client *ImageClient) GetObject(options ...session.ApiOptionsParams) (*models.Image, error) {
	var obj *models.Image
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("image", newOptions...)
	return obj, err
}

// Create a new Image object
func (client *ImageClient) Create(obj *models.Image, options ...session.ApiOptionsParams) (*models.Image, error) {
	var robj *models.Image
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing Image object
func (client *ImageClient) Update(obj *models.Image, options ...session.ApiOptionsParams) (*models.Image, error) {
	var robj *models.Image
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing Image object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.Image
// or it should be json compatible of form map[string]interface{}
func (client *ImageClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.Image, error) {
	var robj *models.Image
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing Image object with a given UUID
func (client *ImageClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing Image object with a given name
func (client *ImageClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ImageClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
