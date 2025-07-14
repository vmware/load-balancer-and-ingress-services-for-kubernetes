'''
*
* Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
* All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
* http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*
'''
import copy
from kubernetes import client, config
from kubernetes.client.exceptions import ApiException
import argparse
import os.path
from pprint import pprint

IP_ANNOTATION = 'ako.vmware.com/load-balancer-ip'
def update_annotation(v1, svc):    
    '''
    ************************************************************************
    Take dictionary of original service, copies Service.status.load_balancer.ingress[0].ip to Service.metadata.annotations and patches Service
    --------------
    Arguments:
    v1 :: CoreV1Api object
    svc :: original service as dictionary
    --------------
    Returns:
    None
    ************************************************************************
    '''
    
    if 'status' in svc.keys() and 'load_balancer' in svc['status'].keys() and 'ingress' in svc['status']['load_balancer'].keys():
        if svc['status']['load_balancer']['ingress'] is not None:
            if 'ip' in svc['status']['load_balancer']['ingress'][0].keys() and svc['status']['load_balancer']['ingress'][0]['ip'] != '':
                svc['metadata']['annotations'][IP_ANNOTATION] = svc['status']['load_balancer']['ingress'][0]['ip']
    service = client.V1Service()
    service.api_version = "v1"
    service.kind = "Service"
    service.metadata = svc['metadata']
    service.spec = svc['spec']
    service.status = svc['status']
    try:
        api_response = v1.patch_namespaced_service(name = svc["metadata"]["name"], namespace="avi-system", body=service)
        return True
    except ApiException as e:
        print(e)
        return False

def fetch_lbsvc_for_update(config_file):
    '''
    ************************************************************************
    Takes a kubeconfig file and retrieves all services found in all namespaces. If service if of type LoadBalancer, then calls update_annotation() to patch service
    --------------
    Arguements:
    config_file :: kubeconfig file path
    --------------
    Returns:
    svc_dict :: Dictionary storing values of all services as 
    {<SERVICE_NAME>:{
                    <SERVICE_NAME>, 
                    <SERVICE_NAMESPACE>, 
                    <SERVICE_CLUSTER_IP>, 
                    <SERVICE_TYPE>, 
                    <SERVICE_VS_IP> (if SERVICE_TYPE==LoadBalancer)
                    }
    }
    ************************************************************************
    '''
    config.load_kube_config(config_file=config_file)
    v1 = client.CoreV1Api()
    ret = v1.list_service_for_all_namespaces()
    svc_dict = {}
    updated_svc = []
    not_updated_svc = []
    for svc in ret.items:
        svc_dict[svc.metadata.name] = {'name':svc.metadata.name, 'namespace':svc.metadata.namespace, 'cluster_ip':svc.spec.cluster_ip, 'type':svc.spec.type}
        if svc.spec.type == 'LoadBalancer':
            if IP_ANNOTATION not in svc.metadata.annotations:
                new_lb = copy.deepcopy(svc.to_dict())
                error_msg = update_annotation(v1, new_lb)
                if not error_msg:
                    not_updated_svc.append(svc.metadata.name)
                else:
                    updated_svc.append(svc.metadata.name)
            else:
                not_updated_svc.append(svc.metadata.name)
    print("Services updated: ", updated_svc)
    print("Services not updated: ", not_updated_svc)
    return svc_dict

def main():
    parser = argparse.ArgumentParser(description="Script to read all namespaced services and add external IP to service annotation for LoadBalancer type services")
    parser.add_argument('kubeconfig_file', help="Path of kubeconfig file in local directory")
    args = parser.parse_args()
    config_file = args.kubeconfig_file
    while not os.path.isfile(config_file):
        config_file = input("Give valid kubeconfig filepath (Input \'End\' to exit program): ")
        if config_file == 'End' or config_file=='end':
            break
    if os.path.isfile(config_file):
        services = fetch_lbsvc_for_update(config_file)
        pprint(services)
    
if __name__ == '__main__':
    main()