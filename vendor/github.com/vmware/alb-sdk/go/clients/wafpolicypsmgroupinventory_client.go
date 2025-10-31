// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// WafPolicyPSMGroupInventoryClient is a client for avi WafPolicyPSMGroupInventory resource
type WafPolicyPSMGroupInventoryClient struct {
	aviSession *session.AviSession
}

// NewWafPolicyPSMGroupInventoryClient creates a new client for WafPolicyPSMGroupInventory resource
func NewWafPolicyPSMGroupInventoryClient(aviSession *session.AviSession) *WafPolicyPSMGroupInventoryClient {
	return &WafPolicyPSMGroupInventoryClient{aviSession: aviSession}
}

func (client *WafPolicyPSMGroupInventoryClient) getAPIPath(uuid string) string {
	path := "api/wafpolicypsmgroupinventory"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of WafPolicyPSMGroupInventory objects
func (client *WafPolicyPSMGroupInventoryClient) GetAll(options ...session.ApiOptionsParams) ([]*models.WafPolicyPSMGroupInventory, error) {
	var plist []*models.WafPolicyPSMGroupInventory
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing WafPolicyPSMGroupInventory by uuid
func (client *WafPolicyPSMGroupInventoryClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.WafPolicyPSMGroupInventory, error) {
	var obj *models.WafPolicyPSMGroupInventory
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing WafPolicyPSMGroupInventory by name
func (client *WafPolicyPSMGroupInventoryClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.WafPolicyPSMGroupInventory, error) {
	var obj *models.WafPolicyPSMGroupInventory
	err := client.aviSession.GetObjectByName("wafpolicypsmgroupinventory", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing WafPolicyPSMGroupInventory by filters like name, cloud, tenant
// Api creates WafPolicyPSMGroupInventory object with every call.
func (client *WafPolicyPSMGroupInventoryClient) GetObject(options ...session.ApiOptionsParams) (*models.WafPolicyPSMGroupInventory, error) {
	var obj *models.WafPolicyPSMGroupInventory
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("wafpolicypsmgroupinventory", newOptions...)
	return obj, err
}

// Create a new WafPolicyPSMGroupInventory object
func (client *WafPolicyPSMGroupInventoryClient) Create(obj *models.WafPolicyPSMGroupInventory, options ...session.ApiOptionsParams) (*models.WafPolicyPSMGroupInventory, error) {
	var robj *models.WafPolicyPSMGroupInventory
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing WafPolicyPSMGroupInventory object
func (client *WafPolicyPSMGroupInventoryClient) Update(obj *models.WafPolicyPSMGroupInventory, options ...session.ApiOptionsParams) (*models.WafPolicyPSMGroupInventory, error) {
	var robj *models.WafPolicyPSMGroupInventory
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing WafPolicyPSMGroupInventory object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.WafPolicyPSMGroupInventory
// or it should be json compatible of form map[string]interface{}
func (client *WafPolicyPSMGroupInventoryClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.WafPolicyPSMGroupInventory, error) {
	var robj *models.WafPolicyPSMGroupInventory
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing WafPolicyPSMGroupInventory object with a given UUID
func (client *WafPolicyPSMGroupInventoryClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing WafPolicyPSMGroupInventory object with a given name
func (client *WafPolicyPSMGroupInventoryClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *WafPolicyPSMGroupInventoryClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
