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

// ServiceEngineGroupClient is a client for avi ServiceEngineGroup resource
type ServiceEngineGroupClient struct {
	aviSession *session.AviSession
}

// NewServiceEngineGroupClient creates a new client for ServiceEngineGroup resource
func NewServiceEngineGroupClient(aviSession *session.AviSession) *ServiceEngineGroupClient {
	return &ServiceEngineGroupClient{aviSession: aviSession}
}

func (client *ServiceEngineGroupClient) getAPIPath(uuid string) string {
	path := "api/serviceenginegroup"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ServiceEngineGroup objects
func (client *ServiceEngineGroupClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ServiceEngineGroup, error) {
	var plist []*models.ServiceEngineGroup
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ServiceEngineGroup by uuid
func (client *ServiceEngineGroupClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ServiceEngineGroup, error) {
	var obj *models.ServiceEngineGroup
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ServiceEngineGroup by name
func (client *ServiceEngineGroupClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ServiceEngineGroup, error) {
	var obj *models.ServiceEngineGroup
	err := client.aviSession.GetObjectByName("serviceenginegroup", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ServiceEngineGroup by filters like name, cloud, tenant
// Api creates ServiceEngineGroup object with every call.
func (client *ServiceEngineGroupClient) GetObject(options ...session.ApiOptionsParams) (*models.ServiceEngineGroup, error) {
	var obj *models.ServiceEngineGroup
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("serviceenginegroup", newOptions...)
	return obj, err
}

// Create a new ServiceEngineGroup object
func (client *ServiceEngineGroupClient) Create(obj *models.ServiceEngineGroup, options ...session.ApiOptionsParams) (*models.ServiceEngineGroup, error) {
	var robj *models.ServiceEngineGroup
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ServiceEngineGroup object
func (client *ServiceEngineGroupClient) Update(obj *models.ServiceEngineGroup, options ...session.ApiOptionsParams) (*models.ServiceEngineGroup, error) {
	var robj *models.ServiceEngineGroup
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ServiceEngineGroup object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ServiceEngineGroup
// or it should be json compatible of form map[string]interface{}
func (client *ServiceEngineGroupClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ServiceEngineGroup, error) {
	var robj *models.ServiceEngineGroup
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ServiceEngineGroup object with a given UUID
func (client *ServiceEngineGroupClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ServiceEngineGroup object with a given name
func (client *ServiceEngineGroupClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ServiceEngineGroupClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
