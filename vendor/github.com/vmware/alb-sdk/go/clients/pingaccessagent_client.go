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

// PingAccessAgentClient is a client for avi PingAccessAgent resource
type PingAccessAgentClient struct {
	aviSession *session.AviSession
}

// NewPingAccessAgentClient creates a new client for PingAccessAgent resource
func NewPingAccessAgentClient(aviSession *session.AviSession) *PingAccessAgentClient {
	return &PingAccessAgentClient{aviSession: aviSession}
}

func (client *PingAccessAgentClient) getAPIPath(uuid string) string {
	path := "api/pingaccessagent"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of PingAccessAgent objects
func (client *PingAccessAgentClient) GetAll(options ...session.ApiOptionsParams) ([]*models.PingAccessAgent, error) {
	var plist []*models.PingAccessAgent
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing PingAccessAgent by uuid
func (client *PingAccessAgentClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.PingAccessAgent, error) {
	var obj *models.PingAccessAgent
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing PingAccessAgent by name
func (client *PingAccessAgentClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.PingAccessAgent, error) {
	var obj *models.PingAccessAgent
	err := client.aviSession.GetObjectByName("pingaccessagent", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing PingAccessAgent by filters like name, cloud, tenant
// Api creates PingAccessAgent object with every call.
func (client *PingAccessAgentClient) GetObject(options ...session.ApiOptionsParams) (*models.PingAccessAgent, error) {
	var obj *models.PingAccessAgent
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("pingaccessagent", newOptions...)
	return obj, err
}

// Create a new PingAccessAgent object
func (client *PingAccessAgentClient) Create(obj *models.PingAccessAgent, options ...session.ApiOptionsParams) (*models.PingAccessAgent, error) {
	var robj *models.PingAccessAgent
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing PingAccessAgent object
func (client *PingAccessAgentClient) Update(obj *models.PingAccessAgent, options ...session.ApiOptionsParams) (*models.PingAccessAgent, error) {
	var robj *models.PingAccessAgent
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing PingAccessAgent object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.PingAccessAgent
// or it should be json compatible of form map[string]interface{}
func (client *PingAccessAgentClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.PingAccessAgent, error) {
	var robj *models.PingAccessAgent
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing PingAccessAgent object with a given UUID
func (client *PingAccessAgentClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing PingAccessAgent object with a given name
func (client *PingAccessAgentClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *PingAccessAgentClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
