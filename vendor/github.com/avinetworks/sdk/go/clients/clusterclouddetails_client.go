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

// ClusterCloudDetailsClient is a client for avi ClusterCloudDetails resource
type ClusterCloudDetailsClient struct {
	aviSession *session.AviSession
}

// NewClusterCloudDetailsClient creates a new client for ClusterCloudDetails resource
func NewClusterCloudDetailsClient(aviSession *session.AviSession) *ClusterCloudDetailsClient {
	return &ClusterCloudDetailsClient{aviSession: aviSession}
}

func (client *ClusterCloudDetailsClient) getAPIPath(uuid string) string {
	path := "api/clusterclouddetails"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ClusterCloudDetails objects
func (client *ClusterCloudDetailsClient) GetAll() ([]*models.ClusterCloudDetails, error) {
	var plist []*models.ClusterCloudDetails
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing ClusterCloudDetails by uuid
func (client *ClusterCloudDetailsClient) Get(uuid string) (*models.ClusterCloudDetails, error) {
	var obj *models.ClusterCloudDetails
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing ClusterCloudDetails by name
func (client *ClusterCloudDetailsClient) GetByName(name string) (*models.ClusterCloudDetails, error) {
	var obj *models.ClusterCloudDetails
	err := client.aviSession.GetObjectByName("clusterclouddetails", name, &obj)
	return obj, err
}

// GetObject - Get an existing ClusterCloudDetails by filters like name, cloud, tenant
// Api creates ClusterCloudDetails object with every call.
func (client *ClusterCloudDetailsClient) GetObject(options ...session.ApiOptionsParams) (*models.ClusterCloudDetails, error) {
	var obj *models.ClusterCloudDetails
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("clusterclouddetails", newOptions...)
	return obj, err
}

// Create a new ClusterCloudDetails object
func (client *ClusterCloudDetailsClient) Create(obj *models.ClusterCloudDetails) (*models.ClusterCloudDetails, error) {
	var robj *models.ClusterCloudDetails
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing ClusterCloudDetails object
func (client *ClusterCloudDetailsClient) Update(obj *models.ClusterCloudDetails) (*models.ClusterCloudDetails, error) {
	var robj *models.ClusterCloudDetails
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing ClusterCloudDetails object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ClusterCloudDetails
// or it should be json compatible of form map[string]interface{}
func (client *ClusterCloudDetailsClient) Patch(uuid string, patch interface{}, patchOp string) (*models.ClusterCloudDetails, error) {
	var robj *models.ClusterCloudDetails
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing ClusterCloudDetails object with a given UUID
func (client *ClusterCloudDetailsClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing ClusterCloudDetails object with a given name
func (client *ClusterCloudDetailsClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *ClusterCloudDetailsClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
