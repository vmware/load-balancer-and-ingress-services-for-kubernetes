/*
 * Copyright 2024 VMware, Inc.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package lib

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

// VKS RBAC Types
type AviRole struct {
	UUID                  string          `json:"uuid,omitempty"`
	Name                  string          `json:"name"`
	TenantRef             string          `json:"tenant_ref"`
	Privileges            []AviPermission `json:"privileges"`
	Filters               []AviRoleFilter `json:"filters,omitempty"`
	AllowUnlabelledAccess bool            `json:"allow_unlabelled_access"`
	Markers               []AviMarker     `json:"markers,omitempty"`
}

type AviPermission struct {
	Type     string `json:"type"`     // READ_ACCESS, WRITE_ACCESS
	Resource string `json:"resource"` // PERMISSION_VIRTUALSERVICE, etc.
}

type AviRoleFilter struct {
	Name           string        `json:"name"`
	MatchOperation string        `json:"match_operation"` // EQUALS, NOT_EQUALS
	MatchLabel     AviMatchLabel `json:"match_label"`
	Enabled        bool          `json:"enabled"`
}

type AviMatchLabel struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

type AviMarker struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

type AviUser struct {
	UUID             string        `json:"uuid,omitempty"`
	Name             string        `json:"name"`
	Username         string        `json:"username"`
	Password         string        `json:"password"`
	Email            string        `json:"email,omitempty"`
	FullName         string        `json:"full_name,omitempty"`
	IsActive         bool          `json:"is_active"`
	IsSuperuser      bool          `json:"is_superuser"`
	Access           []AviUserRole `json:"access"`
	DefaultTenantRef string        `json:"default_tenant_ref"`
}

type AviUserRole struct {
	RoleRef    string `json:"role_ref"`
	TenantRef  string `json:"tenant_ref"`
	AllTenants bool   `json:"all_tenants"`
}

type ClusterCredentials struct {
	Username string
	Password string
	UserUUID string
	RoleUUID string
}

// CreateVKSClusterRole creates a cluster-specific role with proper global/cluster-scoped permissions
func CreateVKSClusterRole(client *clients.AviClient, clusterName, tenant string) (*AviRole, error) {
	roleName := fmt.Sprintf("%s-role", clusterName)

	// Define permissions for GLOBAL objects (no cluster filtering required)
	// These objects are inherently global and cannot be scoped to individual clusters
	globalPermissions := []AviPermission{
		// System and Controller Configuration (Global)
		{Type: "READ_ACCESS", Resource: "PERMISSION_SYSTEMCONFIGURATION"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_CONTROLLERPROPERTIES"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_TENANT"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_CLOUD"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_SERVICEENGINEGROUP"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_VRFCONTEXT"},

		// Global Configuration Templates (Read-only)
		{Type: "READ_ACCESS", Resource: "PERMISSION_NETWORKPROFILE"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_APPLICATIONPROFILE"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_HEALTHMONITOR"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_ANALYTICSPROFILE"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_SSLPROFILE"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_DNSPOLICY"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_NETWORKSECURITYPOLICY"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_APPLICATIONPERSISTENCEPROFILE"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_TRAFFICCLONEPROFILE"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_IPADDRGROUP"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_STRINGGROUP"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_AUTHPROFILE"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_PINGACCESSAGENT"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_CERTIFICATEMANAGEMENTPROFILE"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_HARDWARESECURITYMODULEGROUP"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_SSOPOLICY"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_WAFPROFILE"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_WAFPOLICY"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_ERRORPAGEPROFILE"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_PROTOCOLPARSER"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_NETWORK"},
		{Type: "READ_ACCESS", Resource: "PERMISSION_IPAMDNSPROVIDERPROFILE"},

		// User and Role management (for self-management)
		{Type: "READ_ACCESS", Resource: "PERMISSION_USER"},
		{Type: "WRITE_ACCESS", Resource: "PERMISSION_ROLE"},
	}

	// Define permissions for CLUSTER-SCOPED objects (filtered by clustername label)
	// These objects are created by AKO and should be isolated per cluster
	clusterScopedPermissions := []AviPermission{
		// Objects Created by AKO (WRITE_ACCESS with cluster filtering)
		{Type: "WRITE_ACCESS", Resource: "PERMISSION_VIRTUALSERVICE"},
		{Type: "WRITE_ACCESS", Resource: "PERMISSION_POOL"},
		{Type: "WRITE_ACCESS", Resource: "PERMISSION_POOLGROUP"},
		{Type: "WRITE_ACCESS", Resource: "PERMISSION_HTTPPOLICYSET"},
		{Type: "WRITE_ACCESS", Resource: "PERMISSION_L4POLICYSET"},
		{Type: "WRITE_ACCESS", Resource: "PERMISSION_SSLKEYANDCERTIFICATE"},
		{Type: "WRITE_ACCESS", Resource: "PERMISSION_PKIPROFILE"},
		{Type: "WRITE_ACCESS", Resource: "PERMISSION_VSDATASCRIPTSET"},
	}

	// Create role filter for cluster-specific access (only applies to cluster-scoped objects)
	filters := []AviRoleFilter{
		{
			Name:           fmt.Sprintf("%s-cluster-filter", clusterName),
			MatchOperation: "ROLE_FILTER_EQUALS",
			MatchLabel: AviMatchLabel{
				Key:    "clustername",
				Values: []string{clusterName},
			},
			Enabled: true,
		},
	}

	// Combine global and cluster-scoped permissions
	allPermissions := append(globalPermissions, clusterScopedPermissions...)

	role := &AviRole{
		Name:                  roleName,
		TenantRef:             fmt.Sprintf("/api/tenant/?name=%s", tenant),
		Privileges:            allPermissions,
		Filters:               filters, // Only applies to cluster-scoped objects
		AllowUnlabelledAccess: true,    // Allow access to global objects without labels
		Markers: []AviMarker{
			{
				Key:    "ako-cluster",
				Values: []string{clusterName},
			},
			{
				Key:    "managed-by",
				Values: []string{"vks-dependency-manager"},
			},
		},
	}

	// Create role using AKO's established API patterns
	var result AviRole
	uri := "api/role"
	err := client.AviSession.Post(uri, role, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create VKS cluster role %s: %v", roleName, err)
	}

	utils.AviLog.Infof("Created VKS cluster role with global + cluster-scoped permissions: %s (UUID: %s)", result.Name, result.UUID)
	return &result, nil
}

// CreateVKSClusterUser creates a cluster-specific user with the cluster role
func CreateVKSClusterUser(client *clients.AviClient, clusterName, roleUUID, tenant string) (*AviUser, string, error) {
	username := fmt.Sprintf("%s-user", clusterName)
	password := generateVKSSecurePassword(clusterName)

	user := &AviUser{
		Name:             username,
		Username:         username,
		Password:         password,
		Email:            fmt.Sprintf("%s@cluster.local", username),
		FullName:         fmt.Sprintf("AKO User for Cluster %s", clusterName),
		IsActive:         true,
		IsSuperuser:      false,
		DefaultTenantRef: fmt.Sprintf("/api/tenant/?name=%s", tenant),
		Access: []AviUserRole{
			{
				RoleRef:    fmt.Sprintf("/api/role/%s", roleUUID),
				TenantRef:  fmt.Sprintf("/api/tenant/?name=%s", tenant),
				AllTenants: false,
			},
		},
	}

	// Create user using AKO's established API patterns
	var result AviUser
	uri := "api/user"
	err := client.AviSession.Post(uri, user, &result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create VKS cluster user %s: %v", username, err)
	}

	utils.AviLog.Infof("Created VKS cluster user: %s (UUID: %s)", result.Username, result.UUID)
	return &result, password, nil
}

// DeleteVKSClusterRole deletes a cluster-specific role
func DeleteVKSClusterRole(client *clients.AviClient, clusterName string) error {
	roleName := fmt.Sprintf("%s-role", clusterName)

	// Find role by name using AKO patterns
	var role AviRole
	uri := fmt.Sprintf("api/role?name=%s", roleName)
	err := client.AviSession.Get(uri, &role)
	if err != nil {
		utils.AviLog.Infof("VKS role %s not found, may already be deleted: %v", roleName, err)
		return nil
	}

	// Delete role using AKO patterns
	deleteURI := fmt.Sprintf("api/role/%s", role.UUID)
	err = client.AviSession.Delete(deleteURI)
	if err != nil {
		return fmt.Errorf("failed to delete VKS cluster role %s: %v", roleName, err)
	}

	utils.AviLog.Infof("Deleted VKS cluster role: %s", roleName)
	return nil
}

// DeleteVKSClusterUser deletes a cluster-specific user
func DeleteVKSClusterUser(client *clients.AviClient, clusterName string) error {
	username := fmt.Sprintf("%s-user", clusterName)

	// Find user by username using AKO patterns
	var user AviUser
	uri := fmt.Sprintf("api/user?username=%s", username)
	err := client.AviSession.Get(uri, &user)
	if err != nil {
		utils.AviLog.Infof("VKS user %s not found, may already be deleted: %v", username, err)
		return nil
	}

	// Delete user using AKO patterns
	deleteURI := fmt.Sprintf("api/user/%s", user.UUID)
	err = client.AviSession.Delete(deleteURI)
	if err != nil {
		return fmt.Errorf("failed to delete VKS cluster user %s: %v", username, err)
	}

	utils.AviLog.Infof("Deleted VKS cluster user: %s", username)
	return nil
}

// generateVKSSecurePassword generates a cryptographically secure password for VKS cluster users
func generateVKSSecurePassword(clusterName string) string {
	// Generate random bytes
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Fallback to deterministic but secure generation
		hash := sha256.Sum256([]byte(fmt.Sprintf("vks-%s-%d", clusterName, time.Now().Unix())))
		randomBytes = hash[:16]
	}

	// Create base64 encoded password with VKS prefix
	password := fmt.Sprintf("vks-%s-%s", clusterName[:min(8, len(clusterName))], base64.URLEncoding.EncodeToString(randomBytes)[:16])
	return password
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
