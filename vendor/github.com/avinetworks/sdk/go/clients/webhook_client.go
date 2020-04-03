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

// WebhookClient is a client for avi Webhook resource
type WebhookClient struct {
	aviSession *session.AviSession
}

// NewWebhookClient creates a new client for Webhook resource
func NewWebhookClient(aviSession *session.AviSession) *WebhookClient {
	return &WebhookClient{aviSession: aviSession}
}

func (client *WebhookClient) getAPIPath(uuid string) string {
	path := "api/webhook"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of Webhook objects
func (client *WebhookClient) GetAll(options ...session.ApiOptionsParams) ([]*models.Webhook, error) {
	var plist []*models.Webhook
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing Webhook by uuid
func (client *WebhookClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.Webhook, error) {
	var obj *models.Webhook
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing Webhook by name
func (client *WebhookClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.Webhook, error) {
	var obj *models.Webhook
	err := client.aviSession.GetObjectByName("webhook", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing Webhook by filters like name, cloud, tenant
// Api creates Webhook object with every call.
func (client *WebhookClient) GetObject(options ...session.ApiOptionsParams) (*models.Webhook, error) {
	var obj *models.Webhook
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("webhook", newOptions...)
	return obj, err
}

// Create a new Webhook object
func (client *WebhookClient) Create(obj *models.Webhook, options ...session.ApiOptionsParams) (*models.Webhook, error) {
	var robj *models.Webhook
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing Webhook object
func (client *WebhookClient) Update(obj *models.Webhook, options ...session.ApiOptionsParams) (*models.Webhook, error) {
	var robj *models.Webhook
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing Webhook object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.Webhook
// or it should be json compatible of form map[string]interface{}
func (client *WebhookClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.Webhook, error) {
	var robj *models.Webhook
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing Webhook object with a given UUID
func (client *WebhookClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing Webhook object with a given name
func (client *WebhookClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *WebhookClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
