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

// IPAMDNSProviderProfileClient is a client for avi IPAMDNSProviderProfile resource
type IPAMDNSProviderProfileClient struct {
	aviSession *session.AviSession
}

// NewIPAMDNSProviderProfileClient creates a new client for IPAMDNSProviderProfile resource
func NewIPAMDNSProviderProfileClient(aviSession *session.AviSession) *IPAMDNSProviderProfileClient {
	return &IPAMDNSProviderProfileClient{aviSession: aviSession}
}

func (client *IPAMDNSProviderProfileClient) getAPIPath(uuid string) string {
	path := "api/ipamdnsproviderprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of IPAMDNSProviderProfile objects
func (client *IPAMDNSProviderProfileClient) GetAll() ([]*models.IPAMDNSProviderProfile, error) {
	var plist []*models.IPAMDNSProviderProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing IPAMDNSProviderProfile by uuid
func (client *IPAMDNSProviderProfileClient) Get(uuid string) (*models.IPAMDNSProviderProfile, error) {
	var obj *models.IPAMDNSProviderProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing IPAMDNSProviderProfile by name
func (client *IPAMDNSProviderProfileClient) GetByName(name string) (*models.IPAMDNSProviderProfile, error) {
	var obj *models.IPAMDNSProviderProfile
	err := client.aviSession.GetObjectByName("ipamdnsproviderprofile", name, &obj)
	return obj, err
}

// GetObject - Get an existing IPAMDNSProviderProfile by filters like name, cloud, tenant
// Api creates IPAMDNSProviderProfile object with every call.
func (client *IPAMDNSProviderProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.IPAMDNSProviderProfile, error) {
	var obj *models.IPAMDNSProviderProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("ipamdnsproviderprofile", newOptions...)
	return obj, err
}

// Create a new IPAMDNSProviderProfile object
func (client *IPAMDNSProviderProfileClient) Create(obj *models.IPAMDNSProviderProfile) (*models.IPAMDNSProviderProfile, error) {
	var robj *models.IPAMDNSProviderProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing IPAMDNSProviderProfile object
func (client *IPAMDNSProviderProfileClient) Update(obj *models.IPAMDNSProviderProfile) (*models.IPAMDNSProviderProfile, error) {
	var robj *models.IPAMDNSProviderProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing IPAMDNSProviderProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.IPAMDNSProviderProfile
// or it should be json compatible of form map[string]interface{}
func (client *IPAMDNSProviderProfileClient) Patch(uuid string, patch interface{}, patchOp string) (*models.IPAMDNSProviderProfile, error) {
	var robj *models.IPAMDNSProviderProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing IPAMDNSProviderProfile object with a given UUID
func (client *IPAMDNSProviderProfileClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing IPAMDNSProviderProfile object with a given name
func (client *IPAMDNSProviderProfileClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *IPAMDNSProviderProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
