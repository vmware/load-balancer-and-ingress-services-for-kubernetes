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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/models"

	"google.golang.org/protobuf/proto"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type ClusterCredentials struct {
	Username string
	Password string
}

type ClusterRoles struct {
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
	{"PERMISSION_L4POLICYSET", "WRITE_ACCESS"},
}

var akoAllTenantsPermissions = []AKOPermission{
	{"PERMISSION_CONTROLLER", "READ_ACCESS"},
	{"PERMISSION_TENANT", "READ_ACCESS"},
}

// validateRolePermissions compares existing role permissions with expected AKO permissions
func validateRolePermissions(existingRole *models.Role, expectedPermissions []AKOPermission) error {
	if existingRole.Privileges == nil {
		return fmt.Errorf("existing role has no privileges defined")
	}

	expectedPermsMap := make(map[string]string)
	for _, perm := range expectedPermissions {
		expectedPermsMap[perm.Resource] = perm.Type
	}

	existingPermsMap := make(map[string]string)
	for _, privilege := range existingRole.Privileges {
		if privilege.Resource != nil && privilege.Type != nil {
			existingPermsMap[*privilege.Resource] = *privilege.Type
		}
	}

	var missingPerms []string
	var mismatchedPerms []string
	for resource, expectedType := range expectedPermsMap {
		if existingType, exists := existingPermsMap[resource]; !exists {
			missingPerms = append(missingPerms, fmt.Sprintf("%s:%s", resource, expectedType))
		} else if existingType != expectedType {
			mismatchedPerms = append(mismatchedPerms, fmt.Sprintf("%s: expected %s, got %s", resource, expectedType, existingType))
		}
	}

	var permissionErrors []string
	if len(missingPerms) > 0 {
		permissionErrors = append(permissionErrors, fmt.Sprintf("missing permissions: %v", missingPerms))
	}
	if len(mismatchedPerms) > 0 {
		permissionErrors = append(permissionErrors, fmt.Sprintf("permission mismatches: %v", mismatchedPerms))
	}

	if len(permissionErrors) > 0 {
		return fmt.Errorf("permission validation failed: %s", strings.Join(permissionErrors, "; "))
	}

	var extraPerms []string
	for resource := range existingPermsMap {
		if _, expected := expectedPermsMap[resource]; !expected {
			extraPerms = append(extraPerms, resource)
		}
	}

	if len(extraPerms) > 0 {
		utils.AviLog.Warnf("Role %s has unexpected permissions: %v", *existingRole.Name, extraPerms)
	}

	return nil
}

// createRole creates an Avi role from AKO permission definitions
// If the role already exists with correct permissions, it reuses the existing role
// If the role exists with outdated permissions, it deletes and recreates the role
func createRole(aviClient *clients.AviClient, permissions []AKOPermission,
	roleName, tenantName string, clusterFilter *models.RoleFilter) (*models.Role, error) {

	existingRole, err := aviClient.Role.GetByName(roleName)
	if err == nil && existingRole != nil {
		// Validate existing role has current permissions
		if validateErr := validateRolePermissions(existingRole, permissions); validateErr != nil {
			utils.AviLog.Infof("Role %s exists but has outdated permissions, updating: %v", roleName, validateErr)

			if deleteErr := aviClient.Role.Delete(*existingRole.UUID); deleteErr != nil {
				utils.AviLog.Errorf("Failed to delete outdated role %s: %v", roleName, deleteErr)
				return nil, fmt.Errorf("failed to delete outdated role %s: %v", roleName, deleteErr)
			}
			utils.AviLog.Infof("Deleted outdated role %s, will recreate with current permissions", roleName)
		} else {
			utils.AviLog.Infof("Role %s already exists with current permissions, reusing", roleName)
			return existingRole, nil
		}
	} else if err != nil {
		utils.AviLog.Infof("Role %s does not exist or failed to retrieve (will create new): %v", roleName, err)
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
		TenantRef:             proto.String(fmt.Sprintf("/api/tenant/?name=%s", tenantName)),
		AllowUnlabelledAccess: proto.Bool(true),
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

	utils.AviLog.Infof("Created role: %s (UUID: %s, Tenant: %s)",
		roleName, *createdRole.UUID, tenantName)

	return createdRole, nil
}

// CreateClusterRoles creates the three roles required for a cluster
// Admin and all-tenants roles are shared across clusters for efficiency
// Only tenant roles are cluster-specific due to cluster filtering requirements
func CreateClusterRoles(aviClient *clients.AviClient, clusterName, operationalTenant string) (*ClusterRoles, error) {
	if aviClient == nil {
		return nil, fmt.Errorf("avi Controller client not available - ensure AKO infra is properly initialized")
	}

	utils.AviLog.Infof("Creating cluster roles for %s (operational tenant: %s)", clusterName, operationalTenant)

	adminRole, err := createRole(aviClient, akoAdminPermissions, "vks-admin-role", "admin", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create/reuse shared admin role: %v", err)
	}

	allTenantsRole, err := createRole(aviClient, akoAllTenantsPermissions, "vks-all-tenants-role", "admin", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create/reuse shared all-tenants role: %v", err)
	}

	clusterFilter := &models.RoleFilter{
		MatchOperation: proto.String("ROLE_FILTER_EQUALS"),
		MatchLabel: &models.RoleFilterMatchLabel{
			Key:    proto.String("clustername"),
			Values: []string{clusterName},
		},
		Enabled: proto.Bool(true),
	}

	tenantRoleName := fmt.Sprintf("%s-tenant-role", clusterName)
	tenantRole, err := createRole(aviClient, akoTenantPermissions, tenantRoleName, operationalTenant, clusterFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster-specific tenant role: %v", err)
	}

	roles := &ClusterRoles{
		AdminRole:      adminRole,
		TenantRole:     tenantRole,
		AllTenantsRole: allTenantsRole,
	}

	utils.AviLog.Infof("Successfully created cluster roles for %s: admin=%s (shared), tenant=%s (cluster-specific), all-tenants=%s (shared)",
		clusterName, *adminRole.UUID, *tenantRole.UUID, *allTenantsRole.UUID)

	return roles, nil
}

// CreateClusterUserWithRoles creates a user account for a cluster with three-role access
// If the user already exists, it deletes and recreates it to ensure proper role assignment and fresh password
func CreateClusterUserWithRoles(aviClient *clients.AviClient, clusterName string, roles *ClusterRoles, operationalTenant string) (*models.User, string, error) {
	if aviClient == nil {
		return nil, "", fmt.Errorf("avi Controller client not available - ensure AKO infra is properly initialized")
	}

	userName := fmt.Sprintf("%s-user", clusterName)

	existingUUID, err := getUserByName(aviClient, userName)
	if err != nil {
		utils.AviLog.Warnf("Error checking for existing user %s: %v, will attempt creation", userName, err)
	} else if existingUUID != "" {
		utils.AviLog.Infof("User %s already exists, deleting to recreate with fresh credentials", userName)

		if err := deleteUserByUUID(aviClient, userName, existingUUID); err != nil {
			utils.AviLog.Warnf("Failed to delete existing user %s: %v, continuing", userName, err)
		} else {
			utils.AviLog.Infof("Successfully deleted existing user %s", userName)
		}
	}

	password, err := generateSecurePassword()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate secure password: %v", err)
	}

	userPayload := map[string]interface{}{
		"name":               userName,
		"username":           userName,
		"password":           password,
		"default_tenant_ref": fmt.Sprintf("/api/tenant/?name=%s", operationalTenant),
		"service_user":       true,
		"access": []map[string]interface{}{
			{
				"role_ref":   fmt.Sprintf("/api/role/?name=%s", *roles.AdminRole.Name),
				"tenant_ref": "/api/tenant/?name=admin",
			},
			{
				"role_ref":   fmt.Sprintf("/api/role/?name=%s", *roles.TenantRole.Name),
				"tenant_ref": fmt.Sprintf("/api/tenant/?name=%s", operationalTenant),
			},
			{
				"role_ref":    fmt.Sprintf("/api/role/?name=%s", *roles.AllTenantsRole.Name),
				"all_tenants": true,
				"tenant_ref":  "/api/tenant/?name=admin",
			},
		},
	}

	if err := createUser(aviClient, userPayload); err != nil {
		return nil, "", err
	}

	createdUUID, err := getUserByName(aviClient, userName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve created cluster service user %s: %v", userName, err)
	}

	utils.AviLog.Infof("Created cluster service user: %s (service_user: true, roles: admin + %s-tenant + all-tenants)",
		userName, operationalTenant)

	createdUser := &models.User{
		UUID:     &createdUUID,
		Name:     &userName,
		Username: &userName,
	}

	return createdUser, password, nil
}

// DeleteClusterRoles deletes cluster-specific roles only (not shared roles)
// Shared roles (vks-admin-role, vks-all-tenants-role) are kept for reuse by other clusters
func DeleteClusterRoles(aviClient *clients.AviClient, clusterName string) error {
	if aviClient == nil {
		utils.AviLog.Warnf("Avi Controller client not available for role cleanup of cluster %s", clusterName)
		return nil
	}

	tenantRoleName := fmt.Sprintf("%s-tenant-role", clusterName)

	utils.AviLog.Infof("Deleting cluster-specific roles for %s (preserving shared roles)", clusterName)

	role, err := aviClient.Role.GetByName(tenantRoleName)
	if err != nil {
		utils.AviLog.Warnf("Cluster-specific tenant role %s not found for deletion: %v", tenantRoleName, err)
		return nil
	}

	err = aviClient.Role.Delete(*role.UUID)
	if err != nil {
		return fmt.Errorf("failed to delete cluster-specific tenant role %s: %v", tenantRoleName, err)
	}

	utils.AviLog.Infof("Deleted cluster-specific tenant role: %s (shared admin and all-tenants roles preserved)", tenantRoleName)
	return nil
}

// DeleteClusterUser deletes the user account associated with a cluster
func DeleteClusterUser(aviClient *clients.AviClient, clusterName string) error {
	if aviClient == nil {
		utils.AviLog.Warnf("Avi Controller client not available for user cleanup of cluster %s", clusterName)
		return nil
	}

	userName := fmt.Sprintf("%s-user", clusterName)

	userUUID, err := getUserByName(aviClient, userName)
	if err != nil {
		utils.AviLog.Warnf("Error checking for cluster user %s: %v", userName, err)
		return nil
	}

	if userUUID == "" {
		utils.AviLog.Infof("Cluster user %s not found, already deleted", userName)
		return nil
	}

	return deleteUserByUUID(aviClient, userName, userUUID)
}

func CleanupSharedRoles(aviClient *clients.AviClient) error {
	if aviClient == nil {
		utils.AviLog.Warnf("Avi Controller client not available for shared role cleanup")
		return nil
	}

	sharedRoleNames := []string{
		"vks-admin-role",
		"vks-all-tenants-role",
	}

	utils.AviLog.Warnf("Cleaning up shared VKS roles - this will affect ALL VKS clusters")

	var errors []error
	for _, roleName := range sharedRoleNames {
		role, err := aviClient.Role.GetByName(roleName)
		if err != nil {
			utils.AviLog.Warnf("Shared role %s not found for deletion: %v", roleName, err)
			continue
		}

		err = aviClient.Role.Delete(*role.UUID)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to delete shared role %s: %v", roleName, err))
		} else {
			utils.AviLog.Infof("Deleted shared VKS role: %s", roleName)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to delete some shared roles: %v", errors)
	}

	utils.AviLog.Infof("Successfully cleaned up all shared VKS roles")
	return nil
}

func getUserByName(aviClient *clients.AviClient, userName string) (string, error) {
	uri := fmt.Sprintf("/api/user?name=%s&include_service_accounts=true&fields=uuid", userName)
	result, err := AviGetCollectionRaw(aviClient, uri)
	if err != nil {
		return "", fmt.Errorf("failed to query user %s: %v", userName, err)
	}

	if result.Count == 0 {
		return "", nil
	}

	var users []map[string]interface{}
	if err := json.Unmarshal(result.Results, &users); err != nil {
		return "", fmt.Errorf("failed to unmarshal user results: %v", err)
	}

	if len(users) == 0 {
		utils.AviLog.Warnf("User %s query returned count=%d but empty results after unmarshal", userName, result.Count)
		return "", nil
	}

	if uuid, ok := users[0]["uuid"].(string); ok {
		return uuid, nil
	}

	return "", fmt.Errorf("user %s found but UUID field missing", userName)
}

func deleteUserByUUID(aviClient *clients.AviClient, userName, uuid string) error {
	if uuid == "" {
		return fmt.Errorf("cannot delete user %s: empty UUID", userName)
	}

	deleteURI := fmt.Sprintf("/api/user/%s", uuid)
	err := aviClient.AviSession.Delete(deleteURI)
	if err != nil {
		return fmt.Errorf("failed to delete user %s (UUID: %s): %v", userName, uuid, err)
	}

	utils.AviLog.Infof("Deleted user: %s (UUID: %s)", userName, uuid)
	return nil
}

func createUser(aviClient *clients.AviClient, userPayload map[string]interface{}) error {
	var response interface{}
	err := aviClient.AviSession.Post("api/user", userPayload, &response)
	if err != nil {
		if userName, ok := userPayload["username"].(string); ok {
			return fmt.Errorf("failed to create user %s: %v", userName, err)
		}
		return fmt.Errorf("failed to create user: %v", err)
	}
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
