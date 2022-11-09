# AKO Operator

AKO operator takes care of deploying, managing and removing AKO from openshift/kubernetes clusters. It takes the AKO installation/deployment configuration from a CRD called `AKOConfig`.

## Installing the operator

<!--

Commenting this out as helm release is not available currently

### 1. Install using Helm CLI

To install the Operator using Helm refer [here](../docs/install/operator.md)

-->

### 1. Install on Openshift cluster from OperatorHub using Web Console

<i>**Step 1**</i>: Login to the web console of your Openshift cluster.

<i>**Step 2**</i>: Navigate in the web console to the **Operators** → **OperatorHub** page.

<i>**Step 3**</i>: Find `AKO Operator` provided by VMware. 

<i>**Step 4**</i>: Click `install` and select the 1.8.1 version. The operator will be installed in `avi-system` namespace. The namespace will be created if it doesn't exist.

<i>**Step 5**</i>: Verify installation by checking the pods in `avi-system` namespace. 

> **Note**: Refer [akoconfig](#ako-config) to start the AKO controller

### 2. Manual Installation
### 2.1 Out of cluster execution:
<i>**Step 1**</i>: Clone the [AKO](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes) repo 

<i>**Step 2**</i>: Go to the operator directory
```
cd load-balancer-and-ingress-services-for-kubernetes/ako-operator
```

<i>**Step 3**</i>: To run the operator outside of a cluster, build the binary:
```
make ako-operator
```

<i>**Step 4**</i>: Execute the binary:
```
./bin/ako-operator
```

> **Note**: Refer [akoconfig](#ako-config) to start the AKO controller

### 2.2 In-cluster execution

<i>**Step 1**</i>: Clone the [AKO](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes) repo 

<i>**Step 2**</i>: Build the docker image:
```
cd load-balancer-and-ingress-services-for-kubernetes
make ako-operator-docker
```
<i>**Step 3**</i>: Go to the operator directory:
```
cd ako-operator
```

<i>**Step 4**</i>: Use the following to deploy it on the cluster.
```
make deploy
```

> **Note**: Refer [akoconfig](#ako-config) to start the AKO controller

<!-- 

Commenting this out as helm release is not available currently

Upgrading the operator using Helm CLI

<i>**Step 1**</i>: Run this command to update local AKO chart information from the chart repository:
```
helm repo update
```

<i>**Step 2**</i>: Helm does not upgrade the CRDs during a release upgrade. Before you upgrade a release, run the following command to upgrade the CRDs:
```
helm template ako/ako-operator --version 1.8.1 --include-crds --output-dir <output_dir>
```

<i>**Step 3**</i>: This will save the helm files to an output directory which will contain the CRDs corresponding to the Operator version. Install CRDs using:
```
kubectl apply -f <output_dir>/ako-operator/crds/
```

<i>**Step 4**</i>: List the release as shown below:
```
helm list -n avi-system
```

<i>**Step 5**</i>: Update the helm repo URL:
```
helm repo add --force-update ako https://projects.registry.vmware.com/chartrepo/ako

"ako" has been added to your repositories
```

<i>**Step 6**</i>: Get the values.yaml for the latest Operator version:
```
helm show values ako/ako-operator --version 1.8.1 > values.yaml
```
Edit the file according to your setup.

<i>**Step 7**</i>: Upgrade the helm chart:

```
helm upgrade <release-name> ako/ako-operator -f /path/to/values.yaml --version 1.8.1 --namespace=avi-system
```

--> 

## <a id="ako-config">AKOConfig Custom Resource

AKO Operator manages the AKO Controller. To deploy and manage the controller, it takes in a custom resource object called `AKOConfig`. Please go through the [description](../docs/akoconfig.md#AKOConfig-Custom-Resource) to understand the different fields of this object.

#### Create a secret with Avi Controller details 

Create a secret named `avi-secret` in the `avi-system` namespace. Edit [secret.yaml](config/secrets/secret.yaml) with the credentials of Avi Controller in base64 encoding. 
```
kubectl apply -f config/secrets/secret.yaml
```

#### Deploying the AKO Controller

If the AKO operator was installed using helm, a default `AKOConfig` object called `ako-config` is already added and hence, this step is not required for helm based installation.
**Note**: If the AKO operator was installed manually, then to install the AKO controller, add an `AKOConfig` object to the `avi-system` namespace.

A sample of akoconfig is present [here](config/samples/ako_v1alpha1_akoconfig.yaml). Edit this file according to your setup.

```
kubectl create -f config/samples/ako_v1alpha1_akoconfig.yaml
```

#### Tweaking/Manage the AKO Controller

If the user needs to change any properties of the AKO Controller, they can change the `AKOConfig` object and the changes will take effect once it is saved.

    kubectl edit akoconfig -n avi-system ako-config

**Note** that if the user edits the AKO controller's configmap/statefulset out-of-band, the changes will be overwritten by the AKO operator.

#### Removing the AKO Controller

To remove the AKO Controller, simply delete the `AKOConfig` object:

```
kubectl delete akoconfig -n avi-system ako-config
```

> **Troubleshooting**: If the Operator isn't running when akoconfig is deleted, the akoconfig will be stuck in terminating state. <br>
If this happens edit akoconfig using `kubectl edit akoconfig -n avi-system ako-config` and remove the `finalizers` part. 


### Versioning
| **Operator version** | **Supported AKO Version** |
| --------- | ----------- |
| 1.6.3 | 1.6.2 |
| 1.8.1 | 1.8.1 |