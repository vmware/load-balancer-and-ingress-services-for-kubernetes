# Tenancy support in AKO

This feature allows AKO to map each kubernetes / OpenShift cluster uniquely to a tenant in Avi. `ControllerSettings.tenantsPerCluster` needs to be set to `true` to enable this feature.

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
* In **AKO**, Set the `ControllerSettings.tenantsPerCluster` to `true` and `ControllerSettings.tenantName` to the tenant created in the earlier steps.
* In **AKO**, Set the `avicredentials.username` and `avicredentials.password` to the user credentials created above.

With the above settings AKO will map the `billing` cluster to the `billing` tenant and all the objects will be created in that tenant.

Note:

* In `NodePort` mode of AKO (when `L7Settings.serviceType` is set to `NodePort`), VRFContext permissions are not required in `admin` tenant in AVI Controller.
