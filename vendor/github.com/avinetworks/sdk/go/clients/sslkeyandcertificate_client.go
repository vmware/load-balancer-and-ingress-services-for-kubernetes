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

// SSLKeyAndCertificateClient is a client for avi SSLKeyAndCertificate resource
type SSLKeyAndCertificateClient struct {
	aviSession *session.AviSession
}

// NewSSLKeyAndCertificateClient creates a new client for SSLKeyAndCertificate resource
func NewSSLKeyAndCertificateClient(aviSession *session.AviSession) *SSLKeyAndCertificateClient {
	return &SSLKeyAndCertificateClient{aviSession: aviSession}
}

func (client *SSLKeyAndCertificateClient) getAPIPath(uuid string) string {
	path := "api/sslkeyandcertificate"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of SSLKeyAndCertificate objects
func (client *SSLKeyAndCertificateClient) GetAll(options ...session.ApiOptionsParams) ([]*models.SSLKeyAndCertificate, error) {
	var plist []*models.SSLKeyAndCertificate
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing SSLKeyAndCertificate by uuid
func (client *SSLKeyAndCertificateClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.SSLKeyAndCertificate, error) {
	var obj *models.SSLKeyAndCertificate
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing SSLKeyAndCertificate by name
func (client *SSLKeyAndCertificateClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.SSLKeyAndCertificate, error) {
	var obj *models.SSLKeyAndCertificate
	err := client.aviSession.GetObjectByName("sslkeyandcertificate", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing SSLKeyAndCertificate by filters like name, cloud, tenant
// Api creates SSLKeyAndCertificate object with every call.
func (client *SSLKeyAndCertificateClient) GetObject(options ...session.ApiOptionsParams) (*models.SSLKeyAndCertificate, error) {
	var obj *models.SSLKeyAndCertificate
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("sslkeyandcertificate", newOptions...)
	return obj, err
}

// Create a new SSLKeyAndCertificate object
func (client *SSLKeyAndCertificateClient) Create(obj *models.SSLKeyAndCertificate, options ...session.ApiOptionsParams) (*models.SSLKeyAndCertificate, error) {
	var robj *models.SSLKeyAndCertificate
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing SSLKeyAndCertificate object
func (client *SSLKeyAndCertificateClient) Update(obj *models.SSLKeyAndCertificate, options ...session.ApiOptionsParams) (*models.SSLKeyAndCertificate, error) {
	var robj *models.SSLKeyAndCertificate
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing SSLKeyAndCertificate object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.SSLKeyAndCertificate
// or it should be json compatible of form map[string]interface{}
func (client *SSLKeyAndCertificateClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.SSLKeyAndCertificate, error) {
	var robj *models.SSLKeyAndCertificate
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing SSLKeyAndCertificate object with a given UUID
func (client *SSLKeyAndCertificateClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing SSLKeyAndCertificate object with a given name
func (client *SSLKeyAndCertificateClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *SSLKeyAndCertificateClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
