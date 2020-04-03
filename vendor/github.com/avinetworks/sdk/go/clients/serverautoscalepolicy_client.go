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

// ServerAutoScalePolicyClient is a client for avi ServerAutoScalePolicy resource
type ServerAutoScalePolicyClient struct {
	aviSession *session.AviSession
}

// NewServerAutoScalePolicyClient creates a new client for ServerAutoScalePolicy resource
func NewServerAutoScalePolicyClient(aviSession *session.AviSession) *ServerAutoScalePolicyClient {
	return &ServerAutoScalePolicyClient{aviSession: aviSession}
}

func (client *ServerAutoScalePolicyClient) getAPIPath(uuid string) string {
	path := "api/serverautoscalepolicy"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ServerAutoScalePolicy objects
func (client *ServerAutoScalePolicyClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ServerAutoScalePolicy, error) {
	var plist []*models.ServerAutoScalePolicy
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ServerAutoScalePolicy by uuid
func (client *ServerAutoScalePolicyClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ServerAutoScalePolicy, error) {
	var obj *models.ServerAutoScalePolicy
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ServerAutoScalePolicy by name
func (client *ServerAutoScalePolicyClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ServerAutoScalePolicy, error) {
	var obj *models.ServerAutoScalePolicy
	err := client.aviSession.GetObjectByName("serverautoscalepolicy", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ServerAutoScalePolicy by filters like name, cloud, tenant
// Api creates ServerAutoScalePolicy object with every call.
func (client *ServerAutoScalePolicyClient) GetObject(options ...session.ApiOptionsParams) (*models.ServerAutoScalePolicy, error) {
	var obj *models.ServerAutoScalePolicy
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("serverautoscalepolicy", newOptions...)
	return obj, err
}

// Create a new ServerAutoScalePolicy object
func (client *ServerAutoScalePolicyClient) Create(obj *models.ServerAutoScalePolicy, options ...session.ApiOptionsParams) (*models.ServerAutoScalePolicy, error) {
	var robj *models.ServerAutoScalePolicy
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ServerAutoScalePolicy object
func (client *ServerAutoScalePolicyClient) Update(obj *models.ServerAutoScalePolicy, options ...session.ApiOptionsParams) (*models.ServerAutoScalePolicy, error) {
	var robj *models.ServerAutoScalePolicy
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ServerAutoScalePolicy object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ServerAutoScalePolicy
// or it should be json compatible of form map[string]interface{}
func (client *ServerAutoScalePolicyClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ServerAutoScalePolicy, error) {
	var robj *models.ServerAutoScalePolicy
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ServerAutoScalePolicy object with a given UUID
func (client *ServerAutoScalePolicyClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ServerAutoScalePolicy object with a given name
func (client *ServerAutoScalePolicyClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ServerAutoScalePolicyClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
