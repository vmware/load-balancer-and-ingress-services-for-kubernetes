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

// IPAddrGroupClient is a client for avi IPAddrGroup resource
type IPAddrGroupClient struct {
	aviSession *session.AviSession
}

// NewIPAddrGroupClient creates a new client for IPAddrGroup resource
func NewIPAddrGroupClient(aviSession *session.AviSession) *IPAddrGroupClient {
	return &IPAddrGroupClient{aviSession: aviSession}
}

func (client *IPAddrGroupClient) getAPIPath(uuid string) string {
	path := "api/ipaddrgroup"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of IPAddrGroup objects
func (client *IPAddrGroupClient) GetAll(options ...session.ApiOptionsParams) ([]*models.IPAddrGroup, error) {
	var plist []*models.IPAddrGroup
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing IPAddrGroup by uuid
func (client *IPAddrGroupClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.IPAddrGroup, error) {
	var obj *models.IPAddrGroup
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing IPAddrGroup by name
func (client *IPAddrGroupClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.IPAddrGroup, error) {
	var obj *models.IPAddrGroup
	err := client.aviSession.GetObjectByName("ipaddrgroup", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing IPAddrGroup by filters like name, cloud, tenant
// Api creates IPAddrGroup object with every call.
func (client *IPAddrGroupClient) GetObject(options ...session.ApiOptionsParams) (*models.IPAddrGroup, error) {
	var obj *models.IPAddrGroup
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("ipaddrgroup", newOptions...)
	return obj, err
}

// Create a new IPAddrGroup object
func (client *IPAddrGroupClient) Create(obj *models.IPAddrGroup, options ...session.ApiOptionsParams) (*models.IPAddrGroup, error) {
	var robj *models.IPAddrGroup
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing IPAddrGroup object
func (client *IPAddrGroupClient) Update(obj *models.IPAddrGroup, options ...session.ApiOptionsParams) (*models.IPAddrGroup, error) {
	var robj *models.IPAddrGroup
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing IPAddrGroup object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.IPAddrGroup
// or it should be json compatible of form map[string]interface{}
func (client *IPAddrGroupClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.IPAddrGroup, error) {
	var robj *models.IPAddrGroup
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing IPAddrGroup object with a given UUID
func (client *IPAddrGroupClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing IPAddrGroup object with a given name
func (client *IPAddrGroupClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *IPAddrGroupClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
