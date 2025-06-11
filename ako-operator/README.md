# AKO Operator

AKO operator takes care of deploying, managing and removing AKO from OpenShift clusters. It takes the AKO installation/deployment configuration from a CRD called `AKOConfig`.

## Installing the operator

### Install on OpenShift cluster from OperatorHub using OpenShift Container Platform Web Console

<i>**Step 1**</i>: Login to the OpenShift Container Platform web console of your OpenShift cluster.

<i>**Step 2**</i>: Navigate in the web console to the **Operators** → **OperatorHub** page.

<i>**Step 3**</i>: Find `AKO Operator` provided by VMware.

<i>**Step 4**</i>: Click `install` and select the 1.12.3 version. The operator will be installed in `avi-system` namespace. The namespace will be created if it doesn't exist.

<i>**Step 5**</i>: Verify installation by checking the pods in `avi-system` namespace.

> **Note**: Refer [akoconfig](#ako-config) to start the AKO controller

## <a id="ako-config">AKOConfig Custom Resource

AKO Operator manages the AKO Controller. To deploy and manage the controller, it takes in a custom resource object called `AKOConfig`. Please go through the [description](../docs/akoconfig.md#AKOConfig-Custom-Resource) to understand the different fields of this object.

#### Create a secret with Avi Controller details 

Create a secret named `avi-secret` in the `avi-system` namespace. Edit [secret.yaml](config/secrets/secret.yaml) with the credentials of Avi Controller in base64 encoding. 
```
kubectl apply -f config/secrets/secret.yaml
```

Or, if using the OpenShift client, use
```
oc apply -f config/secrets/secret.yaml
```

#### Deploying the AKO Controller
If the AKO operator was installed on OpenShift cluster from OperatorHub, then to install the AKO controller, add an `AKOConfig` object to the `avi-system` namespace.

A sample of akoconfig is present [here](config/samples/ako_v1alpha1_akoconfig.yaml). Edit this file according to your setup.

```
kubectl create -f config/samples/ako_v1alpha1_akoconfig.yaml
```

Or, if using the OpenShift client, use

```
oc create -f config/samples/ako_v1alpha1_akoconfig.yaml
```

AKO Controller can also be deployed on OpenShift cluster, with AKOConfig custom resource using OpenShift Container Platform Web Console.
#### Prerequisite ####
AKO Operator should already be installed on OpenShift cluster. Once this prerequisite is met, following steps need to be followed.

<i>**Step 1**</i>: Login to the OpenShift Container Platform web console of your OpenShift cluster.

<i>**Step 2**</i>: Navigate in the web console to the **Operators** → **Installed Operators** page. AKO Operator, if already installed, should be listed.

<i>**Step 3**</i>: In the **Provided APIs** section click on `AKOConfig`, and then click on `Create AKOConfig` button.

<i>**Step 4**</i>: You will be provided two configuration options, **Form view** and **YAML view**. Please select the preferred option and populate the fields as required. The AKOConfig custom resource description and sample yaml manifest file can be referred for assistance.

<i>**Step 5**</i>: Once the fields are populated, click on `Create` button.

<i>**Step 6**</i>: Verify installation by checking the pods in `avi-system` namespace.

#### Tweaking/Manage the AKO Controller

If the user needs to change any properties of the AKO Controller, they can change the `AKOConfig` object and the changes will take effect once it is saved.

    kubectl edit akoconfig -n avi-system ako-config

Or, if using the OpenShift client, use

    oc edit akoconfig -n avi-system ako-config

**Note** that if the user edits the AKO controller's configmap/statefulset out-of-band, the changes will be overwritten by the AKO operator.

#### Removing the AKO Controller

To remove the AKO Controller, simply delete the `AKOConfig` object:

```
kubectl delete akoconfig -n avi-system ako-config
```

Or, if using the OpenShift client, use

```
oc delete akoconfig -n avi-system ako-config
```

> **Troubleshooting**: If the Operator isn't running when akoconfig is deleted, the akoconfig will be stuck in terminating state. <br>
If this happens edit akoconfig using `kubectl edit akoconfig -n avi-system ako-config` and remove the `finalizers` section. 


### Versioning
| **Operator version** | **Supported AKO Version** |
| --------- | ----------- |
| 1.11.1 | 1.11.1 |
| 1.12.3 | 1.12.3 |