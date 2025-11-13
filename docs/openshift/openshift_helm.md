### Install AKO for OpenShift

Follow these steps to install AKO using helm in OpenShift environment.

> **Note**: Helm version 3.8 or above is required to proceed with Helm installation.

*Step 1*

Create the avi-system namespace:

```
oc new-project avi-system
```

*Step 2*

Search for available charts:

```
helm show chart oci://projects.packages.broadcom.com/ako/helm-charts/ako --version 2.1.1
Pulled: projects.packages.broadcom.com/ako/helm-charts/ako:2.1.1
Digest: sha256:XXXX
apiVersion: v2
appVersion: 2.1.1
dependencies:
- condition: ako-crd-operator.enabled
  name: ako-crd-operator
  repository: oci://projects.packages.broadcom.com/ako/helm-charts
  version: 2.1.1
description: A helm chart for Avi Kubernetes Operator
name: ako
type: application
version: 2.1.1
```

*Step 3*

Update the values.yaml file according to your environment. Values and their corresponding index can be found [here](../install/helm.md#parameters).
```
helm show values oci://projects.packages.broadcom.com/ako/helm-charts/ako --version 2.1.1 > values.yaml

```
> **Note**: Starting from AKO-2.1.1, the AKO Helm chart has a dependency chart `ako-crd-operator`. The installation can be enabled/disabled by setting `ako-crd-operator.enabled` in the AKO values.yaml.

*Step 4*

Install AKO:

```
helm install --generate-name oci://projects.packages.broadcom.com/ako/helm-charts/ako --version 2.1.1 -f /path/to/values.yaml  --set ControllerSettings.controllerHost=<controller IP or Hostname> --set avicredentials.username=<avi-ctrl-username> --set avicredentials.password=<avi-ctrl-password> --namespace=avi-system
```

*Step 5*

Verify the installation:

```
helm list -n avi-system

NAME          	NAMESPACE 	REVISION	UPDATED     STATUS  	CHART    	APP VERSION
ako-1691752136	avi-system	1       	2025-09-28	deployed	ako-2.1.1	2.1.1
```
