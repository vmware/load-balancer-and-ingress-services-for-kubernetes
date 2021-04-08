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

// PoolGroupDeploymentPolicyClient is a client for avi PoolGroupDeploymentPolicy resource
type PoolGroupDeploymentPolicyClient struct {
	aviSession *session.AviSession
}

// NewPoolGroupDeploymentPolicyClient creates a new client for PoolGroupDeploymentPolicy resource
func NewPoolGroupDeploymentPolicyClient(aviSession *session.AviSession) *PoolGroupDeploymentPolicyClient {
	return &PoolGroupDeploymentPolicyClient{aviSession: aviSession}
}

func (client *PoolGroupDeploymentPolicyClient) getAPIPath(uuid string) string {
	path := "api/poolgroupdeploymentpolicy"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of PoolGroupDeploymentPolicy objects
func (client *PoolGroupDeploymentPolicyClient) GetAll(options ...session.ApiOptionsParams) ([]*models.PoolGroupDeploymentPolicy, error) {
	var plist []*models.PoolGroupDeploymentPolicy
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing PoolGroupDeploymentPolicy by uuid
func (client *PoolGroupDeploymentPolicyClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.PoolGroupDeploymentPolicy, error) {
	var obj *models.PoolGroupDeploymentPolicy
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing PoolGroupDeploymentPolicy by name
func (client *PoolGroupDeploymentPolicyClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.PoolGroupDeploymentPolicy, error) {
	var obj *models.PoolGroupDeploymentPolicy
	err := client.aviSession.GetObjectByName("poolgroupdeploymentpolicy", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing PoolGroupDeploymentPolicy by filters like name, cloud, tenant
// Api creates PoolGroupDeploymentPolicy object with every call.
func (client *PoolGroupDeploymentPolicyClient) GetObject(options ...session.ApiOptionsParams) (*models.PoolGroupDeploymentPolicy, error) {
	var obj *models.PoolGroupDeploymentPolicy
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("poolgroupdeploymentpolicy", newOptions...)
	return obj, err
}

// Create a new PoolGroupDeploymentPolicy object
func (client *PoolGroupDeploymentPolicyClient) Create(obj *models.PoolGroupDeploymentPolicy, options ...session.ApiOptionsParams) (*models.PoolGroupDeploymentPolicy, error) {
	var robj *models.PoolGroupDeploymentPolicy
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing PoolGroupDeploymentPolicy object
func (client *PoolGroupDeploymentPolicyClient) Update(obj *models.PoolGroupDeploymentPolicy, options ...session.ApiOptionsParams) (*models.PoolGroupDeploymentPolicy, error) {
	var robj *models.PoolGroupDeploymentPolicy
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing PoolGroupDeploymentPolicy object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.PoolGroupDeploymentPolicy
// or it should be json compatible of form map[string]interface{}
func (client *PoolGroupDeploymentPolicyClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.PoolGroupDeploymentPolicy, error) {
	var robj *models.PoolGroupDeploymentPolicy
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing PoolGroupDeploymentPolicy object with a given UUID
func (client *PoolGroupDeploymentPolicyClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing PoolGroupDeploymentPolicy object with a given name
func (client *PoolGroupDeploymentPolicyClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *PoolGroupDeploymentPolicyClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
