/***************************************************************************
 * ------------------------------------------------------------------------
 * Copyright 2020 VMware, Inc.  All rights reserved. VMware Confidential
 * ------------------------------------------------------------------------
 */

package clients

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// CertificateManagementProfileClient is a client for avi CertificateManagementProfile resource
type CertificateManagementProfileClient struct {
	aviSession *session.AviSession
}

// NewCertificateManagementProfileClient creates a new client for CertificateManagementProfile resource
func NewCertificateManagementProfileClient(aviSession *session.AviSession) *CertificateManagementProfileClient {
	return &CertificateManagementProfileClient{aviSession: aviSession}
}

func (client *CertificateManagementProfileClient) getAPIPath(uuid string) string {
	path := "api/certificatemanagementprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of CertificateManagementProfile objects
func (client *CertificateManagementProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.CertificateManagementProfile, error) {
	var plist []*models.CertificateManagementProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing CertificateManagementProfile by uuid
func (client *CertificateManagementProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.CertificateManagementProfile, error) {
	var obj *models.CertificateManagementProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing CertificateManagementProfile by name
func (client *CertificateManagementProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.CertificateManagementProfile, error) {
	var obj *models.CertificateManagementProfile
	err := client.aviSession.GetObjectByName("certificatemanagementprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing CertificateManagementProfile by filters like name, cloud, tenant
// Api creates CertificateManagementProfile object with every call.
func (client *CertificateManagementProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.CertificateManagementProfile, error) {
	var obj *models.CertificateManagementProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("certificatemanagementprofile", newOptions...)
	return obj, err
}

// Create a new CertificateManagementProfile object
func (client *CertificateManagementProfileClient) Create(obj *models.CertificateManagementProfile, options ...session.ApiOptionsParams) (*models.CertificateManagementProfile, error) {
	var robj *models.CertificateManagementProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing CertificateManagementProfile object
func (client *CertificateManagementProfileClient) Update(obj *models.CertificateManagementProfile, options ...session.ApiOptionsParams) (*models.CertificateManagementProfile, error) {
	var robj *models.CertificateManagementProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing CertificateManagementProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.CertificateManagementProfile
// or it should be json compatible of form map[string]interface{}
func (client *CertificateManagementProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.CertificateManagementProfile, error) {
	var robj *models.CertificateManagementProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing CertificateManagementProfile object with a given UUID
func (client *CertificateManagementProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing CertificateManagementProfile object with a given name
func (client *CertificateManagementProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *CertificateManagementProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
