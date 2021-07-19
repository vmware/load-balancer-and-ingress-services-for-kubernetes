## Troubleshooting guide for Avi Kubernetes Operator

#### AKO POD is not running

#### Possible Reasons/Solutions

##### Check the reason why the POD didn't come up by doing the following

    kubectl get pods -n avi-system
    NAME                 READY   STATUS             RESTARTS   AGE
    ako-f776577b-5zpxh   0/1     ImagePullBackOff   0          15s

##### Solution

    Ensure that you have your docker registry configured properly or the image is configured locally.

#### AKO pod is restarting automatically and going to a state of crashloopbackoff after some time

##### Possible Reasons/Solutions

From AKO logs check if any input is invalid:

    Invalid input detected, AKO will be rebooted to retry

Check connectivity between AKO Pod and Avi controller.

#### AKO is not responding to my ingress object creations

#### Possible Reasons/Solutions

##### Look into the AKO container logs and see if you find a reason on why the sync is disabled like this

    2020-06-26T10:27:26.032+0530 INFO lib/lib.go:56 Setting AKOUser: ako-my-cluster for Avi Objects
    2020-06-26T10:27:26.337+0530 ERROR cache/controller_obj_cache.go:1814 Required param networkName not specified, syncing will be disabled.
    2020-06-26T10:27:26.337+0530 WARN cache/controller_obj_cache.go:1770 Invalid input detected, syncing will be disabled.

#### My Ingress object didn't sync in Avi

#### Possible Reasons/Solutions

    1. The ingress class is set as something other than "avi". defaultIngController is set to true. 
    2. For TLS ingress, the `Secret` object does not exist. Please ensure that the Secret object is pre-created.
    3. Check the connectivity between your AKO POD and the Avi Controller.

#### My virtualservice returns a CONNECTION REFUSED after sometime

#### Possible Reasons/Solutions

    Check if your virtualservice IP is in use somewhere else in your network.

#### My out-of-band virtualservice setting just got overwritten

#### Possible Reasons/Solutions

    You don't recommend changing properties of a shared virtualservice out-of-band.  If AKO has an ingress update 
    that related to this shared VS, then AKO would overwrite the configuration.

#### Static routes are populated, but my pools are down

#### Possible Reasons/Solutions

    Check if you have a dual nic kubernetes worker node setup. In case of a dual nic setup, AKO would populate the static
    routes using the default gateway network. However, the default gateway network might not be the port group network that
    you want to use as the data network. Hence service engines may not be able to reach the POD CIDRs using the default gateway
    network. 
    
    If it's impossible to make your data networks routable via the default gateway, disableStaticRoute sync in AKO and edit your
    static routes with the correct network.

## Log Collection

For every log collection, also collect the following information:

    1. What kubernetes distribution are you using? For example: RKE, PKS etc.
    2. What is the CNI you are using with versions? For example: Calico v3.15
    3. What is the Avi Controller version you are using? For example: 18.2.8

### How do I gather the AKO logs?

Get the script from [here](https://github.com/avinetworks/devops/tree/master/openshift/ako/log_collector.py)

The script is used to collect all relevant information for the AKO pod.

**About the script:**

1. Collects log file of AKO pod
2. Collects configmap  in a yaml file
3. Zips the folder and returns

_For logs collection, 3 cases are considered:_

Case 1 : A running AKO pod logging into a Persistent Volume Claim, in this case the logs are collected from the PVC that the pod uses.

Case 2 : A running AKO pod logging into console, in this case the logs are collected from the pod directly.

Case 3 : A dead AKO pod that uses a Persistent Volume Claim, in this case a backup pod is created with the same PVC attached to the AKO pod and the logs are collected from it.

**Configuring PVC for the AKO pod:**

We recommend using a Persistent Volume Claim for the ako pod. Refer this [link](https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/) to create a persistent volume(PV) and a Persistent Volume Claim(PVC).

Below is an example of hostpath persistent volume. We recommend you use the PV based on the storage class of your kubernetes environment.

    #persistent-volume.yaml
    apiVersion: v1
    kind: PersistentVolume
    metadata:
      name: ako-pv
      namespace : avi-system
      labels:
        type: local
    spec:
      storageClassName: manual
      capacity:
        storage: 10Gi
      accessModes:
        - ReadWriteOnce
      hostPath:
        path: <any-host-path-dir>  # make sure that the directory exists

A persistent volume claim can be created using the following file

    #persistent-volume-claim.yaml
    apiVersion: v1
    kind: PersistentVolumeClaim
    metadata:
      name: ako-pvc
      namespace : avi-system
    spec:
      storageClassName: manual
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 3Gi

Add PVC name into the ako/helm/ako/values.yaml before the creation of the ako pod like

    persistentVolumeClaim: ako-pvc
    mountPath: /log
    logFile: avi.log

**How to use the script for AKO**

Usage:

1. Case 1: With PVC, (Mandatory) --akoNamespace (-ako) : The namespace in which the AKO pod is present.

    `python3 log_collections.py -ako avi-system`

2. Case 2: Without PVC (Optional) --since (-s) : time duration from present time for logs.

    `python3 log_collections.py -ako avi-system -s 24h`

**Sample Run:**

At each stage of execution, the commands being executed are logged on the screen.
The results are stored in a zip file with the format below:

    ako-<helmchart name>-<current time>

Sample Output with PVC :

    2020-06-25 13:20:37,141 - ******************** AKO ********************
    2020-06-25 13:20:37,141 - For AKO : helm list -n avi-system
    2020-06-25 13:20:38,974 - kubectl get pod -n avi-system -l app.kubernetes.io/instance=my-ako-release
    2020-06-25 13:20:41,850 - kubectl describe pod ako-56887bd5b7-c2t6n -n avi-system
    2020-06-25 13:20:44,019 - helm get all my-ako-release -n avi-system
    2020-06-25 13:20:46,360 - PVC name is my-pvc
    2020-06-25 13:20:46,361 - PVC mount point found - /log
    2020-06-25 13:20:46,361 - Log file name is avi.log
    2020-06-25 13:20:46,362 - Creating directory ako-my-ako-release-2020-06-25-132046
    2020-06-25 13:20:46,373 - kubectl cp avi-system/ako-56887bd5b7-c2t6n:log/avi.log ako-my-ako-release-2020-06-25-132046/ako.log
    2020-06-25 13:21:02,098 - kubectl get cm -n avi-system -o yaml > ako-my-ako-release-2020-06-25-132046/config-map.yaml
    2020-06-25 13:21:03,495 - Zipping directory ako-my-ako-release-2020-06-25-132046
    2020-06-25 13:21:03,525 - Clean up: rm -r ako-my-ako-release-2020-06-25-132046

    Success, Logs zipped into ako-my-ako-release-2020-06-25-132046.zip

### How do I gather the controller tech support?

It's recommended we collect the controller tech support logs as well. Please follow this [link](https://avinetworks.com/docs/18.2/collecting-tech-support-logs/)  for the controller tech support.

## Troubleshooting for AKO CRDs

### Policy defined in the crd policy was not applied to the corresponding ingress/route objects

1. Make sure that the policy object being referred by the CRD is present in avi.
2. Ensure that connectivity between ako pod and avi controller is intact. For example if the avi controller is rebooting, connectivity may go down and we may face this problem.

## Troubleshooting for openshift route

### Route objects did not sync to avi

There can be different reasons behind this. Some common issues can be categorized as follows:

#### 1. The problem is for all routes

Some configuration parameter is missing. Check for logs like

    Invalid input detected, syncing will be disabled

Make the necessary changes in the configuration by checking the logs and restart AKO.

#### 2. Some routes are not getting handled in ako

Check if subdomain of the route is valid as per avi controller configuration

    Didn't find match for hostname :foo.abc.com Available sub-domains:avi.internal

#### 3. The problem is faced for One / very few routes

Check for status of route. If you see a message `MultipleBackendsWithSameServiceError`, then same service has been added multiple times in the backends. This is a wrong configuration and the route configuration has to be changed.

#### 4. The route which is not getting synced, is a secure route

Check the following conditions:

- Both key and cert are specified in the route spec.
- The default secret (router-certs-default) is present in avi-system Namespace.

 If both of these conditions are false, AKO can't process a secure route correctly. Either the default secret has to be created in avi-system Namesapce, or key and cert have to be specified in the route spec.
