### Install AKO for OpenShift

Follow these steps to install AKO using helm in Openshift environment

> **Note**: Helm version 3.8 and above will be required to proceed with helm installation.

*Step-1*

Create the avi-system namespace.

```
oc new-project avi-system
```

*Step-2*

Search for available charts

```
helm show chart oci://projects.packages.broadcom.com/ako/helm-charts/ako --version 1.12.2

Pulled: projects.packages.broadcom.com/ako/helm-charts/ako:1.12.2
Digest: sha256:xyxyxxyxyx
apiVersion: v2
appVersion: 1.12.2
description: A helm chart for Avi Kubernetes Operator
name: ako
type: application
version: 1.12.2
```

*Step-3*

Edit the [values.yaml](../install/helm.md#parameters) file and update the details according to your environment.

```
helm show values oci://projects.packages.broadcom.com/ako/helm-charts/ako --version 1.12.2 > values.yaml

```

*Step-4*

Install AKO.

```
helm install --generate-name oci://projects.packages.broadcom.com/ako/helm-charts/ako --version 1.12.2 -f /path/to/values.yaml  --set ControllerSettings.controllerHost=<controller IP or Hostname> --set avicredentials.username=<avi-ctrl-username> --set avicredentials.password=<avi-ctrl-password> --namespace=avi-system
```


*Step-5*

Verify the installation

```
helm list -n avi-system

NAME          	NAMESPACE 	REVISION	UPDATED     STATUS  	CHART    	APP VERSION
ako-1691752136	avi-system	1       	2024-08-04	deployed	ako-1.12.2	1.12.2
```


