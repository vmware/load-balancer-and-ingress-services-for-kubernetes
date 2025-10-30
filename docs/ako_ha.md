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

**Note:** Lease lock object won't be created when AKO is running with single replica.

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

Step 1: Search the available charts for AKO

```
helm show chart oci://projects.packages.broadcom.com/ako/helm-charts/ako --version 2.2.1

Pulled: projects.packages.broadcom.com/ako/helm-charts/ako:2.2.1
Digest: sha256:xxxxxxxx
apiVersion: v2
appVersion: 2.2.1
dependencies:
- condition: ako-crd-operator.enabled
  name: ako-crd-operator
  repository: oci://projects.packages.broadcom.com/ako/helm-charts
  version: 2.2.1
description: A helm chart for Avi Kubernetes Operator
name: ako
type: application
version: 2.2.1
```

Step 2: Pull AKO helm chart
```
helm pull oci://projects.packages.broadcom.com/ako/helm-charts/ako --version 2.2.1 --untar
```

Step 3: Update helm dependency after going into ako directory
```
cd ako
helm dependency build
```

### Transitioning from Single Replica to High Availability

To transition from a single AKO replica to high availability mode, you must follow this specific sequence to ensure proper leader election:

1. **Scale down to zero replicas**: Change the `replicaCount` in `values.yaml` to 0 and execute the helm upgrade command.
2. **Scale up to two replicas**: Change the `replicaCount` in `values.yaml` to 2 and execute the helm upgrade command.

```bash
# upgrade command
helm upgrade ako-1593523840 . --set ControllerSettings.controllerHost=<IP or Hostname> --set avicredentials.password=<username> --set avicredentials.username=<username> --set ako-crd-operator.enabled=false --namespace=avi-system
```

**Note**: Set ako-crd-operator.enabled to true to install ako-crd-operator as part of upgrade.

**Important:** This two-step process is required because:
- When scaling directly from 1 to 2 replicas, the first replica skips leader election (thinking it's still single replica)
- The second replica starts leader election, but the first replica doesn't participate
- This can lead to both replicas becoming active simultaneously
- By scaling to 0 first, both replicas start fresh and properly participate in leader election

**Note:**
1. Currently, more than two replicas are not supported.
2. Both instances of AKO should be on the same version.
