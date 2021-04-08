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

// ProtocolParserClient is a client for avi ProtocolParser resource
type ProtocolParserClient struct {
	aviSession *session.AviSession
}

// NewProtocolParserClient creates a new client for ProtocolParser resource
func NewProtocolParserClient(aviSession *session.AviSession) *ProtocolParserClient {
	return &ProtocolParserClient{aviSession: aviSession}
}

func (client *ProtocolParserClient) getAPIPath(uuid string) string {
	path := "api/protocolparser"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ProtocolParser objects
func (client *ProtocolParserClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ProtocolParser, error) {
	var plist []*models.ProtocolParser
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ProtocolParser by uuid
func (client *ProtocolParserClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ProtocolParser, error) {
	var obj *models.ProtocolParser
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ProtocolParser by name
func (client *ProtocolParserClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ProtocolParser, error) {
	var obj *models.ProtocolParser
	err := client.aviSession.GetObjectByName("protocolparser", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ProtocolParser by filters like name, cloud, tenant
// Api creates ProtocolParser object with every call.
func (client *ProtocolParserClient) GetObject(options ...session.ApiOptionsParams) (*models.ProtocolParser, error) {
	var obj *models.ProtocolParser
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("protocolparser", newOptions...)
	return obj, err
}

// Create a new ProtocolParser object
func (client *ProtocolParserClient) Create(obj *models.ProtocolParser, options ...session.ApiOptionsParams) (*models.ProtocolParser, error) {
	var robj *models.ProtocolParser
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ProtocolParser object
func (client *ProtocolParserClient) Update(obj *models.ProtocolParser, options ...session.ApiOptionsParams) (*models.ProtocolParser, error) {
	var robj *models.ProtocolParser
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ProtocolParser object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ProtocolParser
// or it should be json compatible of form map[string]interface{}
func (client *ProtocolParserClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ProtocolParser, error) {
	var robj *models.ProtocolParser
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ProtocolParser object with a given UUID
func (client *ProtocolParserClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ProtocolParser object with a given name
func (client *ProtocolParserClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ProtocolParserClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
