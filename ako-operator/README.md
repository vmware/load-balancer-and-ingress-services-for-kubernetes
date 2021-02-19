## An operator for AKO

This operator takes care of deploying, managing and removing AKO from openshift/kubernetes clusters. It takes the AKO installation/deployment configuration from a CRD called `AKOConfig`.

### Pre-reqs before deploying the operator
- CRD for `AKOConfig` must be installed.
- A secret called `avi-secret` must be installed. This contains the username
  and password for the Avi controller in base64 encoding (edit [secret.yaml](config/secrets/secret.yaml)).
- And, CRD definitions for `HostRule` and `HttpRule` must be installed
  (Currently, not enforced).

Run `make install` to install the above definitions.

### Installing the operator
#### Out of cluster execution
To run the operator outside of a cluster, first build the binary:
```
cd ako-operator
make ako-operator
```
And then from `ako-operator` directory, run it using:
```
bin/ako-operator
```
#### In-cluster execution
First build the docker image:
```
make docker-build
```
And then, use the following to deploy it on a k8s cluster.
```
make deploy
```

### Configuration values
Create an `AKOConfig` on the cluster. A sample file is present in [samples](config/samples/ako_v1alpha1_akoconfig.yaml).

```
kubectl create -f config/samples/ako_v1alpha1_akoconfig.yaml
```
The operator has now started syncing the AKO controller.