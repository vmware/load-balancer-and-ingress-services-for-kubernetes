# Tenancy support in AKO

This feature allows AKO to map each kubernetes / OpenShift cluster uniquely to a tenant in Avi. 

## Tenant Context

AVI non admin tenants primarily operate in 2 modes, **provider context** and **tenant context**.

### Provider Context

Service Engine Groups are shared with `admin` tenant. All the other objects like Virtual Services and Pools are created within the tenant. Requires `config_settings.se_in_provider_context` flag to be set to `True` when creating tenant. 

### Tenant Context

Service Engines are isolated from `admin` tenant. A new `Default-Group` is created within the tenant. All the objects including Service Engines are created in tenant context. Requires `config_settings.se_in_provider_context` flag to be set to `False` when creating tenant. 

## Steps to enable Tenancy in AKO

* Create separate tenant for each cluster in AVI. For the below steps, lets assume `billing` tenant is created by the Avi controller admin.
![Alt text](images/tenant_path.png?raw=true)
* Click `create`
![Alt text](images/new_tenant.png?raw=true)
* Create the [`ako-admin`](roles/ako-admin.json) and [`ako-tenant`](roles/ako-tenant.json) roles which gives appropriate privileges to the ako user in `admin` and `billing` tenant.
![Alt text](images/role_list.png?raw=true)
* Create a new user for AKO in AVI under `Administration->Accounts->Tenants`
![Alt text](images/user_path.png?raw=true)
* Click `create`
![Alt text](images/new_user.png?raw=true)
* Assign [`ako-admin`](roles/ako-admin.json) and [`ako-tenant`](roles/ako-tenant.json) roles to admin and billing tenant respectively.
![Alt text](images/new_user_role.png?raw=true)
* In **AKO**, Set the `ControllerSettings.tenantName` to the tenant created in the earlier steps.
* In **AKO**, Set the `avicredentials.username` and `avicredentials.password` to the user credentials created above.

With the above settings AKO will map the `billing` cluster to the `billing` tenant and all the objects will be created in that tenant.

> **Note**: In `NodePort` mode of AKO (when `L7Settings.serviceType` is set to `NodePort`), VRFContext permissions are not required in `admin` tenant in AVI Controller.
