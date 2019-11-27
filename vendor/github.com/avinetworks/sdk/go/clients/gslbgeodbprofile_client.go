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

// GslbGeoDbProfileClient is a client for avi GslbGeoDbProfile resource
type GslbGeoDbProfileClient struct {
	aviSession *session.AviSession
}

// NewGslbGeoDbProfileClient creates a new client for GslbGeoDbProfile resource
func NewGslbGeoDbProfileClient(aviSession *session.AviSession) *GslbGeoDbProfileClient {
	return &GslbGeoDbProfileClient{aviSession: aviSession}
}

func (client *GslbGeoDbProfileClient) getAPIPath(uuid string) string {
	path := "api/gslbgeodbprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of GslbGeoDbProfile objects
func (client *GslbGeoDbProfileClient) GetAll() ([]*models.GslbGeoDbProfile, error) {
	var plist []*models.GslbGeoDbProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing GslbGeoDbProfile by uuid
func (client *GslbGeoDbProfileClient) Get(uuid string) (*models.GslbGeoDbProfile, error) {
	var obj *models.GslbGeoDbProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing GslbGeoDbProfile by name
func (client *GslbGeoDbProfileClient) GetByName(name string) (*models.GslbGeoDbProfile, error) {
	var obj *models.GslbGeoDbProfile
	err := client.aviSession.GetObjectByName("gslbgeodbprofile", name, &obj)
	return obj, err
}

// GetObject - Get an existing GslbGeoDbProfile by filters like name, cloud, tenant
// Api creates GslbGeoDbProfile object with every call.
func (client *GslbGeoDbProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.GslbGeoDbProfile, error) {
	var obj *models.GslbGeoDbProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("gslbgeodbprofile", newOptions...)
	return obj, err
}

// Create a new GslbGeoDbProfile object
func (client *GslbGeoDbProfileClient) Create(obj *models.GslbGeoDbProfile) (*models.GslbGeoDbProfile, error) {
	var robj *models.GslbGeoDbProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing GslbGeoDbProfile object
func (client *GslbGeoDbProfileClient) Update(obj *models.GslbGeoDbProfile) (*models.GslbGeoDbProfile, error) {
	var robj *models.GslbGeoDbProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing GslbGeoDbProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.GslbGeoDbProfile
// or it should be json compatible of form map[string]interface{}
func (client *GslbGeoDbProfileClient) Patch(uuid string, patch interface{}, patchOp string) (*models.GslbGeoDbProfile, error) {
	var robj *models.GslbGeoDbProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing GslbGeoDbProfile object with a given UUID
func (client *GslbGeoDbProfileClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing GslbGeoDbProfile object with a given name
func (client *GslbGeoDbProfileClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *GslbGeoDbProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
