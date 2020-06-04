# Copyright 2019-2020 VMware, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
import argparse
import subprocess
import re
import shutil
import yaml
import time
import warnings
from datetime import datetime
import logging
warnings.filterwarnings("ignore")
encoding = 'utf-8'

def getLogFolderName(args):
    return args.cluster + "-"+ args.helmchart + "-" + str(datetime.now().strftime("%Y-%m-%d-%H%M%S"))

def pvHelper(args) : 
    # Take details of PV and PVC from the ako pod helm chart
    helmResult = subprocess.check_output("helm get all %s -n %s" %(args.helmchart,args.namespace) , shell=True)
    logging.info("helm get all %s -n %s" %(args.helmchart,args.namespace))
    helmResult = helmResult.decode(encoding)
    listhelmResult = helmResult.split("---")
    for output in listhelmResult:
        dictOutput = yaml.safe_load(output)
        try:
            return dictOutput['spec']['template']['spec']
        except KeyError:
            continue

def findPVCName(templateSpec, args):
    #Return PV name of ako pod
    try: 
        pvcName = templateSpec['volumes'][0]['persistentVolumeClaim']['claimName']
        return pvcName
    except KeyError:
        logging.info("Persistent Volume for pod is not defined\nReading logs directly from the pod")
        folderName = getLogFolderName(args)
        logging.info("Creating directory %s" %folderName)
        output = subprocess.check_output("mkdir %s" %folderName, shell=True)
        logging.info("kubectl logs %s -n %s --since %s > %s/log-file" %(findPodName(args),args.namespace,args.since,folderName))
        output = subprocess.check_output("kubectl logs %s -n %s --since %s > %s/log-file" %(findPodName(args),args.namespace,args.since,folderName) , shell=True)
        getConfigMap(args,folderName)
        logging.info("Zipping directory %s" %folderName)
        shutil.make_archive(folderName, 'zip', folderName)
        logging.info("Clean up: rm -r %s" %folderName)
        output = subprocess.check_output("rm -r %s" %folderName, shell=True)
        print("\nSuccess, Logs zipped into %s.zip\n" %folderName)
        return "no pvc"
    
def findPVMount(templateSpec):
    #Return mount place of PV of ako pod
    try: 
        pvMount = templateSpec['containers'][0]['volumeMounts'][0]['mountPath']
        return pvMount
    except KeyError:
        print("Persistent Volume Mount for pod is not defined\nMounting the log files to /log path\n")
        return "/log"

def editDeploymentFile(pvcName,pvMount,args):
    deploymentDict = {'apiVersion': 'v1', 'kind':'Pod', 'metadata':{'name': 'custom-backup-pod', 'namespace': '' }, 'spec':{'containers':[{'image': 'avinetworks/server-os', 'name': 'myfrontend', 'volumeMounts':[{'mountPath': '', 'name': 'mypd'}]}], 'volumes':[{'name': 'mypd', 'persistentVolumeClaim':{'claimName': ''}}]}} 
    deploymentDict['spec']['containers'][0]['volumeMounts'][0]['mountPath'] = pvMount
    deploymentDict['spec']['volumes'][0]['persistentVolumeClaim']['claimName'] = pvcName
    deploymentDict['metadata']['namespace'] = args.namespace
    pod = open('pod.yaml','w+')
    yaml.dump(deploymentDict, pod)

def findPodName(args):
    logging.info("kubectl get pod -n %s -l app.kubernetes.io/instance=%s" %(args.namespace, args.helmchart))
    Pods = subprocess.check_output("kubectl get pod -n %s -l app.kubernetes.io/instance=%s" %(args.namespace, args.helmchart) , shell=True)
    Pods = Pods.decode(encoding)
    newLine = Pods.splitlines()[1]
    podName = newLine.split(' ')[0]
    return podName


def getConfigMap(args,folderName):
    logging.info("kubectl get cm -n %s -o yaml > %s/config-map.yaml" %(args.namespace,folderName))
    subprocess.check_output("kubectl get cm -n %s -o yaml > %s/config-map.yaml" %(args.namespace,folderName), shell=True)

def zipLogFile (args):
    podName = findPodName(args)
    try:
        #Find details of the Ako pod
        logging.info("kubectl describe pod %s -n %s" %(podName,args.namespace))
        statusOfAkoPod =  subprocess.check_output("kubectl describe pod %s -n %s" %(podName,args.namespace) , shell=True)
        statusOfAkoPod =  statusOfAkoPod.decode(encoding)
    except:
        #If details couldnt be fetched, the kubectl describe raises any exception, then return failure
        print(args.namespace,"\n")
        print("Error is describe of ako pod\n")
        return 0

    templateSpec = pvHelper(args)
    pvcName = findPVCName(templateSpec,args)
    if pvcName == "no pvc":
        return 1
    pvMount = findPVMount(templateSpec)
    folderName = getLogFolderName(args)

    #Check if the ako pod is up and running
    if (re.findall("Status: *Running", statusOfAkoPod)):
        #If ako pod is running, copy the log file to zip it
        try:
            logging.info("Creating directory %s" %folderName)
            output = subprocess.check_output("mkdir %s" %folderName, shell=True)
            logging.info("kubectl cp %s/%s:%s %s" %(args.namespace,podName,pvMount[1:],folderName))
            output = subprocess.check_output("kubectl cp %s/%s:%s %s" %(args.namespace,podName,pvMount[1:],folderName), shell=True)
            getConfigMap(args,folderName)
        except:
            print("Error is cp of ako pod\n")
            return 0
        logging.info("Zipping directory %s" %folderName)
        shutil.make_archive(folderName, 'zip', folderName)
        logging.info("Clean up: rm -r %s" %folderName)
        output = subprocess.check_output("rm -r %s" %folderName, shell=True)
        print("\nSuccess, Logs zipped into %s.zip\n" %folderName)
        return 1
    #If ako pod isnt running, then create backup pod named "mypod"
    else:
        #Creation of "mypod"
        logging.info("Creating backup pod as ako pod isn't running")
        editDeploymentFile(pvcName,pvMount,args)
        try:
            logging.info("kubectl apply -f pod.yaml")
            podCreated = subprocess.check_output("kubectl apply -f pod.yaml", shell=True)
        except:
            return 0
        timeout = time.time() + 10
        #Wait for "mypod" to start running
        while(1):
            try:
                logging.info("kubectl describe pod custom-backup-pod -n %s" %args.namespace)
                statusOfBackupPod =  subprocess.check_output("kubectl describe pod custom-backup-pod -n %s" %args.namespace , shell=True)
                statusOfBackupPod = statusOfBackupPod.decode(encoding)
            except: 
                return 0
            if (re.findall("Status: *Running", statusOfBackupPod)):
                #Once "mypod" is running, copy the log file to zip it
                print("\nBackup pod \'custom-backup-pod\' started\n")
                logging.info("Creating directory %s" %folderName)
                output = subprocess.check_output("mkdir %s" %folderName, shell=True)
                logging.info("kubectl cp %s/custom-backup-pod:%s %s" %(args.namespace,pvMount[1:],folderName))
                output = subprocess.check_output("kubectl cp %s/custom-backup-pod:%s %s" %(args.namespace,pvMount[1:],folderName),shell=True)
                getConfigMap(args,folderName)
                logging.info("Zipping directory %s" %folderName)
                shutil.make_archive(folderName, 'zip', folderName)
                #Clean up
                logging.info("Clean up: kubectl delete pod custom-backup-pod -n %s" %args.namespace)
                backupPodDeletion =  subprocess.check_output("kubectl delete pod custom-backup-pod -n %s" %args.namespace , shell=True)
                logging.info("Clean up: rm pod.yaml")
                backupPodDeletion =  subprocess.check_output("rm pod.yaml", shell= True)
                logging.info("Clean up: rm -r %s" %folderName)
                output = subprocess.check_output("rm -r %s" %folderName, shell=True)

                print("\nSuccess, Logs zipped into %s.zip\n" %folderName)
                return 1
            time.sleep(2)
            if time.time()>timeout:
                break
        print("Couldn't create backup pod\n")
    return 0

if __name__ == "__main__":
    #Parsing cli arguments
    parser = argparse.ArgumentParser(formatter_class=argparse.RawTextHelpFormatter)
    parser.add_argument('-n', '--namespace', help='Namespace in which the ako pod is present' )
    parser.add_argument('-H', '--helmchart', help='Helm Chart name' )
    parser.add_argument('-c', '--cluster', help='Cluster name' )
    parser.add_argument('-w', '--wait', default= 10, help='Number of seconds to wait for the backup pod to start running before exiting\nDefault is 10 seconds' )
    parser.add_argument('-s', '--since',default='24h', help='For pods not having persistent volume storage the logs since a given time duration can be fetched.\nExample : mention the time as 2s(for 2 seconds) or 4m(for 4 mins) or 24h(for 24 hours)\nDefault is taken to be 24h' )
    args = parser.parse_args()


    logging.basicConfig(format='%(asctime)s - %(message)s', level=logging.INFO)

    if (not args.helmchart or not args.namespace or not args.cluster):
        print("Scripts requires arguments\nTry \'python3 get_logs.py --help\' for more info\n\n")
        exit()

    if(zipLogFile(args)==0):
        print("Error getting log file\n\n")