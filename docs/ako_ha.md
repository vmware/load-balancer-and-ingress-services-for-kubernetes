# High Availability support in AKO

This feature allows the user to run two instances of AKO in a Kubernetes/OpenShift cluster - one in active mode and the other in passive mode.

AKO's high availability architecture is as described below:
![Alt text](images/ako_ha_arch.png?raw=true)
<div align="center">
Fig: AKO High Availability Architecture
<br/>
<br/>
</div>

Active and passive modes are assigned automatically by performing a leadership election among the AKOs. A lease lock(Kubernetes object) named `ako-lease-lock` in `avi-system` has been used to keep track of the current active AKO. The lease lock object has the identity of the current active AKO and a field named `renewTime` which active AKO periodically refreshes. The passive AKO periodically polls the lease lock object and updates its identity in the lease lock object when the `renewTime` goes beyond the deadline.

Leader election between AKOs occurs as described below:
![Alt text](images/ako_ha_election.png?raw=true)
<div align="center">
Fig: Diagram showing leader election among AKOs
<br/>
<br/>
</div>

Active AKO does the following:
* Creates the AVI objects in the AVI controller.
* Updates the status of the Ingress/Routes/Service of type LB.
* Cleans up the stale AVI objects from the AVI controller.
* Cleans up the AVI objects created by AKO from the controller when `deleteConfig` has been set.
* Creates the lease object in the `avi-system` namespace and periodically renews the `renewTime` of the lease object.

Passive AKO does the following:
* Polls the lease object in the `avi-system` namespace.
* Reads the objects in Kubernetes/OpenShift cluster and populates the cache.
* Reads the AVI objects configured by Active AKO and builds the cache.

## Steps to run AKO in High Availability

* Change the `replicaCount` in `values.yaml` to two.
* Execute the helm upgrade command and provide the updated `values.yaml` file.

helm upgrade ako-1593523840 oci://projects.registry.vmware.com/ako/helm-charts/ako -f /path/to/values.yaml --version 1.11.5 --set ControllerSettings.controllerHost=<IP or Hostname> --set avicredentials.password=<username> --set avicredentials.username=<username> --namespace=avi-system

**Note:**
1. Currently, more than two replicas are not supported.
2. Both instances of AKO should be on the same version.
