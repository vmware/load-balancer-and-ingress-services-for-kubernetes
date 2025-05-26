# Tenancy support in AKO

This feature allows AKO to map each kubernetes / OpenShift cluster uniquely to a tenant in Avi or to map each namespace in single kubernetes / OpenShift cluster uniquely to a tenant in Avi. 

## Tenant Context

AVI non admin tenants primarily operate in 2 modes, **provider context** and **tenant context**.

### Provider Context

Service Engine Groups are shared with `admin` tenant. All the other objects like Virtual Services and Pools are created within the tenant. Requires `config_settings.se_in_provider_context` flag to be set to `True` when creating tenant. 

### Tenant Context

Service Engines are isolated from `admin` tenant. A new `Default-Group` is created within the tenant. All the objects including Service Engines are created in tenant context. Requires `config_settings.se_in_provider_context` flag to be set to `False` when creating tenant. 

## Steps to enable Tenancy in AKO to map each kubernetes / OpenShift cluster uniquely to a tenant in Avi

* Create separate tenant for each cluster in AVI. For the below steps, lets assume `billing` tenant is created by the Avi controller admin.
![Alt text](images/tenant_path.png?raw=true)
* Click `create`
![Alt text](images/new_tenant.png?raw=true)
* Create the [`ako-admin`](roles/ako-admin.json) and [`ako-tenant`](roles/ako-tenant.json) roles which gives appropriate privileges to the ako user in `admin` and `billing` tenant.
![Alt text](images/role_list.png?raw=true)
* Create the [`ako-all-tenants-permission-controller`](roles/ako-all-tenants-permission-controller.json) role, which grants the AKO user read privilege for the controller for all tenants.
![Alt text](images/ako-all-tenants-permission-controller.png?raw=true)
* Create a new user for AKO in AVI under `Administration->Accounts->Tenants`
![Alt text](images/user_path.png?raw=true)
* Click `create`
![Alt text](images/new_user.png?raw=true)
* Assign [`ako-admin`](roles/ako-admin.json) and [`ako-tenant`](roles/ako-tenant.json) roles to admin and billing tenant respectively.
![Alt text](images/new_user_role.png?raw=true)
* Assign the [`ako-all-tenants-permission-controller`](roles/ako-all-tenants-permission-controller.json) role to all tenants.
![Alt text](images/all-tenants-role.png?raw=true)
* In **AKO**, Set the `ControllerSettings.tenantName` to the tenant created in the earlier steps.
* In **AKO**, Set the `avicredentials.username` and `avicredentials.password` to the user credentials created above.

With the above settings AKO will map the `billing` cluster to the `billing` tenant and all the objects will be created in that tenant.

> **Note**: In `NodePort` mode of AKO (when `L7Settings.serviceType` is set to `NodePort`), VRFContext permissions are not required in `admin` tenant in AVI Controller.

## Steps to enable Tenancy in AKO at namespace level of single cluster

* At the namespace level, it will be done by annotating the namespace with the corresponding tenant name in Avi. This will enable all the resources in the namespace to use the annotated tenant. This way namespace can have a relationship with tenant in Avi.

* AKO will determine the tenant to create AVI objects from `ako.vmware.com/tenant-name` annotation value specified in the namespace of Kubernetes/openshift objects.

* If `ako.vmware.com/tenant-name` annotation is empty or missing AKO will determine tenant from [`tenantName`](values.md#controllersettingstenantname) field.

* All references to AVI objects in AKO CRD's should be accessible in the tenant annotated in the namespace to the AKO User.If they are not accesible CRD would transition to error status and won't be applied to VS. 

* AKO user should have [`ako-tenant`](roles/ako-tenant.json) role assigned in all tenants used so that avi objects could be created/updated.  

**Notes**: 
* In case of tenant update on namespace AKO Restart will be required to delete stale avi objects and update status correctly on kubernetes/OpenShift objects.
* This feature is not supported for services of type LoadBalancer using shared [VIP](./shared_vip.md).  


### Example of an annotated tenant in AKO

In this example AKO will create virtual Services in `billing` tenant for Kubernetes/openshift objects in `n1` namespace. For other namespaces which are missing the annotation virtual service will be created in the tenant where AKO is installed .

```
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    ako.vmware.com/tenant-name: billing
  name: n1
```

With the above settings AKO will map the `n1` namespace to the `billing` tenant and all the objects will be created in that tenant .

## Troubleshooting

**Q: I've applied the `ako-tenant.json` role, but the controller UUID is not being populated. What should I do?**

**A:** You need to assign the `ako-all-tenants-permission-controller.json` role to the "All Tenants" section in the Avi UI. This grants the AKO user the necessary read privileges for the controller across all tenants.

To do this, navigate to `Administration -> Accounts -> Users` in the Avi UI. Select the relevant user (e.g., the billing_ako_user in this case). Then, in the **Roles for all Tenants** section for that user, assign the `ako-all-tenants-permission-controller` role to "All Tenants" as shown below:
![Assign ako-all-tenants-permission-controller to All Tenants](images/all-tenants-role.png?raw=true)
