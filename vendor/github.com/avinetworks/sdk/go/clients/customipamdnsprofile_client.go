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

// CustomIPAMDNSProfileClient is a client for avi CustomIPAMDNSProfile resource
type CustomIPAMDNSProfileClient struct {
	aviSession *session.AviSession
}

// NewCustomIPAMDNSProfileClient creates a new client for CustomIPAMDNSProfile resource
func NewCustomIPAMDNSProfileClient(aviSession *session.AviSession) *CustomIPAMDNSProfileClient {
	return &CustomIPAMDNSProfileClient{aviSession: aviSession}
}

func (client *CustomIPAMDNSProfileClient) getAPIPath(uuid string) string {
	path := "api/customipamdnsprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of CustomIPAMDNSProfile objects
func (client *CustomIPAMDNSProfileClient) GetAll() ([]*models.CustomIPAMDNSProfile, error) {
	var plist []*models.CustomIPAMDNSProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing CustomIPAMDNSProfile by uuid
func (client *CustomIPAMDNSProfileClient) Get(uuid string) (*models.CustomIPAMDNSProfile, error) {
	var obj *models.CustomIPAMDNSProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing CustomIPAMDNSProfile by name
func (client *CustomIPAMDNSProfileClient) GetByName(name string) (*models.CustomIPAMDNSProfile, error) {
	var obj *models.CustomIPAMDNSProfile
	err := client.aviSession.GetObjectByName("customipamdnsprofile", name, &obj)
	return obj, err
}

// GetObject - Get an existing CustomIPAMDNSProfile by filters like name, cloud, tenant
// Api creates CustomIPAMDNSProfile object with every call.
func (client *CustomIPAMDNSProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.CustomIPAMDNSProfile, error) {
	var obj *models.CustomIPAMDNSProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("customipamdnsprofile", newOptions...)
	return obj, err
}

// Create a new CustomIPAMDNSProfile object
func (client *CustomIPAMDNSProfileClient) Create(obj *models.CustomIPAMDNSProfile) (*models.CustomIPAMDNSProfile, error) {
	var robj *models.CustomIPAMDNSProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing CustomIPAMDNSProfile object
func (client *CustomIPAMDNSProfileClient) Update(obj *models.CustomIPAMDNSProfile) (*models.CustomIPAMDNSProfile, error) {
	var robj *models.CustomIPAMDNSProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing CustomIPAMDNSProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.CustomIPAMDNSProfile
// or it should be json compatible of form map[string]interface{}
func (client *CustomIPAMDNSProfileClient) Patch(uuid string, patch interface{}, patchOp string) (*models.CustomIPAMDNSProfile, error) {
	var robj *models.CustomIPAMDNSProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing CustomIPAMDNSProfile object with a given UUID
func (client *CustomIPAMDNSProfileClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing CustomIPAMDNSProfile object with a given name
func (client *CustomIPAMDNSProfileClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *CustomIPAMDNSProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
