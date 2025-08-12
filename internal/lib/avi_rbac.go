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

/*
Package lib provides VKS (VMware Kubernetes Services) RBAC management for Avi Controller.

This implementation follows the AKO tenancy model with three-role structure:

1. ADMIN TENANT ROLE: Minimal admin tenant access for infrastructure operations
2. OPERATIONAL TENANT ROLE: Full operational permissions with cluster-specific filtering where supported
3. ALL-TENANTS ROLE: Controller read access across all tenants

Permissions are based on existing AKO role definitions (ako-admin.json, ako-tenant.json,
ako-all-tenants-permission-controller.json) but implemented as Go constants.

Cluster isolation is achieved through Avi's Extended Granular RBAC with markers:
- Objects supporting markers get cluster-specific filtering (clustername=cluster-name)
- Objects not supporting markers are shared within tenant
- This aligns with AKO's existing cluster marker system

References:
- docs/ako_tenancy.md
- docs/roles/*.json
- https://techdocs.broadcom.com/...extended-granular-rbac/using-markers.html
*/

package lib

import (
	"crypto/rand"
	"fmt"

	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/models"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type ClusterCredentials struct {
	Username string
	Password string
}

type VKSClusterRoles struct {
	AdminRole      *models.Role
	TenantRole     *models.Role
	AllTenantsRole *models.Role
}

type AKOPermission struct {
	Resource string
	Type     string
}

var akoAdminPermissions = []AKOPermission{
	{"PERMISSION_STRINGGROUP", "WRITE_ACCESS"},
	{"PERMISSION_CLOUD", "READ_ACCESS"},
	{"PERMISSION_SERVICEENGINEGROUP", "WRITE_ACCESS"},
	{"PERMISSION_NETWORK", "READ_ACCESS"},
	{"PERMISSION_VRFCONTEXT", "WRITE_ACCESS"},
	{"PERMISSION_TENANT", "READ_ACCESS"},
}

var akoTenantPermissions = []AKOPermission{
	{"PERMISSION_VIRTUALSERVICE", "WRITE_ACCESS"},
	{"PERMISSION_POOL", "WRITE_ACCESS"},
	{"PERMISSION_POOLGROUP", "WRITE_ACCESS"},
	{"PERMISSION_HTTPPOLICYSET", "WRITE_ACCESS"},
	{"PERMISSION_NETWORKSECURITYPOLICY", "WRITE_ACCESS"},
	{"PERMISSION_AUTOSCALE", "WRITE_ACCESS"},
	{"PERMISSION_DNSPOLICY", "WRITE_ACCESS"},
	{"PERMISSION_NETWORKPROFILE", "WRITE_ACCESS"},
	{"PERMISSION_APPLICATIONPROFILE", "WRITE_ACCESS"},
	{"PERMISSION_APPLICATIONPERSISTENCEPROFILE", "WRITE_ACCESS"},
	{"PERMISSION_HEALTHMONITOR", "WRITE_ACCESS"},
	{"PERMISSION_ANALYTICSPROFILE", "WRITE_ACCESS"},
	{"PERMISSION_IPAMDNSPROVIDERPROFILE", "WRITE_ACCESS"},
	{"PERMISSION_CUSTOMIPAMDNSPROFILE", "WRITE_ACCESS"},
	{"PERMISSION_TRAFFICCLONEPROFILE", "WRITE_ACCESS"},
	{"PERMISSION_IPADDRGROUP", "READ_ACCESS"},
	{"PERMISSION_STRINGGROUP", "WRITE_ACCESS"},
	{"PERMISSION_VSDATASCRIPTSET", "WRITE_ACCESS"},
	{"PERMISSION_PROTOCOLPARSER", "READ_ACCESS"},
	{"PERMISSION_SSLPROFILE", "READ_ACCESS"},
	{"PERMISSION_AUTHPROFILE", "READ_ACCESS"},
	{"PERMISSION_PINGACCESSAGENT", "READ_ACCESS"},
	{"PERMISSION_PKIPROFILE", "WRITE_ACCESS"},
	{"PERMISSION_SSLKEYANDCERTIFICATE", "WRITE_ACCESS"},
	{"PERMISSION_CERTIFICATEMANAGEMENTPROFILE", "READ_ACCESS"},
	{"PERMISSION_HARDWARESECURITYMODULEGROUP", "READ_ACCESS"},
	{"PERMISSION_SSOPOLICY", "READ_ACCESS"},
	{"PERMISSION_WAFPROFILE", "READ_ACCESS"},
	{"PERMISSION_WAFPOLICY", "READ_ACCESS"},
	{"PERMISSION_CLOUD", "READ_ACCESS"},
	{"PERMISSION_SERVICEENGINEGROUP", "WRITE_ACCESS"},
	{"PERMISSION_NETWORK", "WRITE_ACCESS"},
	{"PERMISSION_VRFCONTEXT", "WRITE_ACCESS"},
	{"PERMISSION_SYSTEMCONFIGURATION", "READ_ACCESS"},
	{"PERMISSION_TENANT", "READ_ACCESS"},
	{"PERMISSION_L4POLICYSET", "WRITE_ACCESS"},
}

var akoAllTenantsPermissions = []AKOPermission{
	{"PERMISSION_CONTROLLER", "READ_ACCESS"},
}

// createRoleFromPermissions creates an Avi role from AKO permission definitions
// If the role already exists, it returns the existing role
func createRoleFromPermissions(aviClient *clients.AviClient, permissions []AKOPermission,
	roleName, tenantName string, clusterFilter *models.RoleFilter) (*models.Role, error) {

	existingRole, err := aviClient.Role.GetByName(roleName)
	if err == nil && existingRole != nil {
		utils.AviLog.Infof("VKS role %s already exists (UUID: %s), reusing existing role",
			roleName, *existingRole.UUID)
		return existingRole, nil
	}

	var privileges []*models.Permission
	for _, perm := range permissions {
		privileges = append(privileges, &models.Permission{
			Type:     &perm.Type,
			Resource: &perm.Resource,
		})
	}

	role := &models.Role{
		Name:                  &roleName,
		Privileges:            privileges,
		TenantRef:             func() *string { s := fmt.Sprintf("/api/tenant/?name=%s", tenantName); return &s }(),
		AllowUnlabelledAccess: func() *bool { b := true; return &b }(),
	}

	if clusterFilter != nil {
		role.Filters = []*models.RoleFilter{clusterFilter}
		utils.AviLog.Infof("Creating role %s with cluster filter: clustername=%s",
			roleName, clusterFilter.MatchLabel.Values[0])
	}

	_, err = aviClient.Role.Create(role)
	if err != nil {
		return nil, fmt.Errorf("failed to create role %s: %v", roleName, err)
	}

	createdRole, err := aviClient.Role.GetByName(roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created role %s: %v", roleName, err)
	}

	utils.AviLog.Infof("Created VKS role: %s (UUID: %s, Tenant: %s)",
		roleName, *createdRole.UUID, tenantName)

	return createdRole, nil
}

// CreateVKSClusterRoles creates the three roles required for a VKS cluster
func CreateVKSClusterRoles(aviClient *clients.AviClient, clusterName, operationalTenant string) (*VKSClusterRoles, error) {
	if aviClient == nil {
		return nil, fmt.Errorf("avi Controller client not available - ensure AKO infra is properly initialized")
	}

	utils.AviLog.Infof("Creating VKS cluster roles for %s (operational tenant: %s)", clusterName, operationalTenant)

	// Create admin tenant role
	adminRoleName := fmt.Sprintf("vks-cluster-%s-admin-role", clusterName)
	adminRole, err := createRoleFromPermissions(aviClient, akoAdminPermissions, adminRoleName, "admin", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin tenant role: %v", err)
	}

	// Create operational tenant role with cluster filtering
	clusterFilter := &models.RoleFilter{
		MatchOperation: func() *string { s := "ROLE_FILTER_EQUALS"; return &s }(),
		MatchLabel: &models.RoleFilterMatchLabel{
			Key:    func() *string { s := "clustername"; return &s }(),
			Values: []string{clusterName},
		},
		Enabled: func() *bool { b := true; return &b }(),
	}

	tenantRoleName := fmt.Sprintf("vks-cluster-%s-tenant-role", clusterName)
	tenantRole, err := createRoleFromPermissions(aviClient, akoTenantPermissions, tenantRoleName, operationalTenant, clusterFilter)
	if err != nil {
		aviClient.Role.Delete(*adminRole.UUID)
		return nil, fmt.Errorf("failed to create operational tenant role: %v", err)
	}

	// Create all-tenants role
	allTenantsRoleName := fmt.Sprintf("vks-cluster-%s-all-tenants-role", clusterName)
	allTenantsRole, err := createRoleFromPermissions(aviClient, akoAllTenantsPermissions, allTenantsRoleName, "admin", nil)
	if err != nil {
		aviClient.Role.Delete(*adminRole.UUID)
		aviClient.Role.Delete(*tenantRole.UUID)
		return nil, fmt.Errorf("failed to create all-tenants role: %v", err)
	}

	roles := &VKSClusterRoles{
		AdminRole:      adminRole,
		TenantRole:     tenantRole,
		AllTenantsRole: allTenantsRole,
	}

	utils.AviLog.Infof("Successfully created VKS cluster roles for %s: admin=%s, tenant=%s, all-tenants=%s",
		clusterName, *adminRole.UUID, *tenantRole.UUID, *allTenantsRole.UUID)

	return roles, nil
}

// CreateVKSClusterUserWithRoles creates a user account for a VKS cluster with three-role access
// If the user already exists, it deletes and recreates it to ensure proper role assignment and fresh password
func CreateVKSClusterUserWithRoles(aviClient *clients.AviClient, clusterName string, roles *VKSClusterRoles, operationalTenant string) (*models.User, string, error) {
	if aviClient == nil {
		return nil, "", fmt.Errorf("avi Controller client not available - ensure AKO infra is properly initialized")
	}

	userName := fmt.Sprintf("vks-cluster-%s-user", clusterName)

	// Check if user already exists and delete it
	// We recreate users to ensure fresh passwords and correct role assignments
	existingUser, err := aviClient.User.GetByName(userName)
	if err == nil && existingUser != nil {
		utils.AviLog.Infof("VKS user %s already exists (UUID: %s), deleting to recreate with fresh credentials",
			userName, *existingUser.UUID)

		err = aviClient.User.Delete(*existingUser.UUID)
		if err != nil {
			utils.AviLog.Warnf("Failed to delete existing VKS user %s: %v, attempting to continue", userName, err)
		} else {
			utils.AviLog.Infof("Successfully deleted existing VKS user %s", userName)
		}
	}

	password, err := generateSecurePassword()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate secure password: %v", err)
	}

	userAccess := []*models.UserRole{
		{
			RoleRef:   roles.AdminRole.UUID,
			TenantRef: func() *string { s := "/api/tenant/?name=admin"; return &s }(),
		},
		{
			RoleRef:   roles.TenantRole.UUID,
			TenantRef: func() *string { s := fmt.Sprintf("/api/tenant/?name=%s", operationalTenant); return &s }(),
		},
		{
			RoleRef:   roles.AllTenantsRole.UUID,
			TenantRef: func() *string { s := "*"; return &s }(),
		},
	}

	user := &models.User{
		Name:             &userName,
		Username:         &userName,
		Password:         &password,
		DefaultTenantRef: func() *string { s := fmt.Sprintf("/api/tenant/?name=%s", operationalTenant); return &s }(),
		Access:           userAccess,
	}

	utils.AviLog.Infof("Creating VKS cluster user: %s (operational tenant: %s)", userName, operationalTenant)

	_, err = aviClient.User.Create(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create VKS cluster user %s: %v", userName, err)
	}

	createdUser, err := aviClient.User.GetByName(userName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve created VKS cluster user %s: %v", userName, err)
	}

	utils.AviLog.Infof("Created VKS cluster user with three-role access: %s (UUID: %s)", userName, *createdUser.UUID)
	utils.AviLog.Infof("User access: admin tenant, operational tenant (%s), all-tenants controller read", operationalTenant)

	return createdUser, password, nil
}

// DeleteVKSClusterRoles deletes all roles associated with a VKS cluster
func DeleteVKSClusterRoles(aviClient *clients.AviClient, clusterName string) error {
	if aviClient == nil {
		utils.AviLog.Warnf("Avi Controller client not available for role cleanup of cluster %s", clusterName)
		return nil
	}

	roleNames := []string{
		fmt.Sprintf("vks-cluster-%s-admin-role", clusterName),
		fmt.Sprintf("vks-cluster-%s-tenant-role", clusterName),
		fmt.Sprintf("vks-cluster-%s-all-tenants-role", clusterName),
	}

	var errors []error
	for _, roleName := range roleNames {
		role, err := aviClient.Role.GetByName(roleName)
		if err != nil {
			utils.AviLog.Warnf("VKS cluster role %s not found for deletion: %v", roleName, err)
			continue
		}

		err = aviClient.Role.Delete(*role.UUID)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to delete VKS cluster role %s: %v", roleName, err))
		} else {
			utils.AviLog.Infof("Deleted VKS cluster role: %s", roleName)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to delete some VKS cluster roles: %v", errors)
	}

	return nil
}

// DeleteVKSClusterUser deletes the user account associated with a VKS cluster
func DeleteVKSClusterUser(aviClient *clients.AviClient, clusterName string) error {
	if aviClient == nil {
		utils.AviLog.Warnf("Avi Controller client not available for user cleanup of cluster %s", clusterName)
		return nil
	}

	userName := fmt.Sprintf("vks-cluster-%s-user", clusterName)

	user, err := aviClient.User.GetByName(userName)
	if err != nil {
		utils.AviLog.Warnf("VKS cluster user %s not found for deletion: %v", userName, err)
		return nil
	}

	err = aviClient.User.Delete(*user.UUID)
	if err != nil {
		return fmt.Errorf("failed to delete VKS cluster user %s: %v", userName, err)
	}

	utils.AviLog.Infof("Deleted VKS cluster user: %s", userName)
	return nil
}

// generateSecurePassword generates a cryptographically secure random password
func generateSecurePassword() (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		utils.AviLog.Errorf("Failed to generate secure password: %v", err)
		return "", fmt.Errorf("failed to generate secure password: %v", err)
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	password := make([]byte, 16)
	for i, b := range randomBytes {
		password[i] = charset[int(b)%len(charset)]
	}

	return string(password), nil
}
