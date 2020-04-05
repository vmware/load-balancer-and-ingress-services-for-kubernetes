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

// VsVipClient is a client for avi VsVip resource
type VsVipClient struct {
	aviSession *session.AviSession
}

// NewVsVipClient creates a new client for VsVip resource
func NewVsVipClient(aviSession *session.AviSession) *VsVipClient {
	return &VsVipClient{aviSession: aviSession}
}

func (client *VsVipClient) getAPIPath(uuid string) string {
	path := "api/vsvip"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VsVip objects
func (client *VsVipClient) GetAll(options ...session.ApiOptionsParams) ([]*models.VsVip, error) {
	var plist []*models.VsVip
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing VsVip by uuid
func (client *VsVipClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.VsVip, error) {
	var obj *models.VsVip
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing VsVip by name
func (client *VsVipClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.VsVip, error) {
	var obj *models.VsVip
	err := client.aviSession.GetObjectByName("vsvip", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing VsVip by filters like name, cloud, tenant
// Api creates VsVip object with every call.
func (client *VsVipClient) GetObject(options ...session.ApiOptionsParams) (*models.VsVip, error) {
	var obj *models.VsVip
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vsvip", newOptions...)
	return obj, err
}

// Create a new VsVip object
func (client *VsVipClient) Create(obj *models.VsVip, options ...session.ApiOptionsParams) (*models.VsVip, error) {
	var robj *models.VsVip
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing VsVip object
func (client *VsVipClient) Update(obj *models.VsVip, options ...session.ApiOptionsParams) (*models.VsVip, error) {
	var robj *models.VsVip
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing VsVip object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VsVip
// or it should be json compatible of form map[string]interface{}
func (client *VsVipClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.VsVip, error) {
	var robj *models.VsVip
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing VsVip object with a given UUID
func (client *VsVipClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing VsVip object with a given name
func (client *VsVipClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *VsVipClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
