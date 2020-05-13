import os
import argparse
import subprocess
import re
import shutil
import yaml
import time
import warnings
from datetime import datetime
warnings.filterwarnings("ignore")
encoding = 'utf-8'

def getLogFolderName(args):
    return args.cluster + "-"+ args.helmChart + "-" + str(datetime.now().strftime("%Y-%m-%d-%H%M%S"))

def pvHelper(args) : 
    # Take details of PV and PVC from the ako pod helm chart
    helmResult = subprocess.check_output("helm get all %s -n %s" %(args.helmChart,args.namespace) , shell=True)
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
        print("Persistent Volume for pod is not defined\nReading logs directly from the pod\n")
        folderName = getLogFolderName(args)
        output = subprocess.check_output("mkdir %s" %folderName, shell=True)
        output = subprocess.check_output("kubectl logs %s -n %s --since %s > %s/log-file" %(findPodName(args),args.namespace,args.since,folderName) , shell=True)
        getConfigMap(args,folderName)
        shutil.make_archive(folderName, 'zip', folderName)
        print("Success, Logs zipped into %s.zip\n" %folderName)
        output = subprocess.check_output("rm -r %s" %folderName, shell=True)
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
    Pods = subprocess.check_output("kubectl get pod -n %s -l app.kubernetes.io/instance=%s" %(args.namespace, args.helmChart) , shell=True)
    Pods = Pods.decode(encoding)
    newLine = Pods.splitlines()[1]
    podName = newLine.split(' ')[0]
    return podName


def getConfigMap(args,folderName):
    subprocess.check_output("kubectl get cm -n %s -o yaml > %s/config-map.yaml" %(args.namespace,folderName), shell=True)

def zipLogFile (args):
    podName = findPodName(args)
    try:
        #Find details of the Ako pod
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
            output = subprocess.check_output("mkdir %s" %folderName, shell=True)
            output = subprocess.check_output("kubectl cp %s/%s:%s %s" %(args.namespace,podName,pvMount[1:],folderName), shell=True)
            getConfigMap(args,folderName)
        except:
            print("Error is cp of ako pod\n")
            return 0
        shutil.make_archive(folderName, 'zip', folderName)
        print("Success, Logs zipped into %s.zip\n" %folderName)
        output = subprocess.check_output("rm -r %s" %folderName, shell=True)
        return 1
    #If ako pod isnt running, then create backup pod named "mypod"
    else:
        #Creation of "mypod"
        editDeploymentFile(pvcName,pvMount,args)
        try:
            podCreated = subprocess.check_output("kubectl apply -f pod.yaml", shell=True)
        except:
            return 0
        timeout = time.time() + 10
        #Wait for "mypod" to start running
        while(1):
            try:
                statusOfBackupPod =  subprocess.check_output("kubectl describe pod custom-backup-pod -n %s" %args.namespace , shell=True)
                statusOfBackupPod = statusOfBackupPod.decode(encoding)
            except: 
                return 0
            if (re.findall("Status: *Running", statusOfBackupPod)):
                #Once "mypod" is running, copy the log file to zip it
                print("Backup pod \'custom-backup-pod\' started\n")
                output = subprocess.check_output("mkdir %s" %folderName, shell=True)
                output = subprocess.check_output("kubectl cp %s/custom-backup-pod:%s %s" %(args.namespace,pvMount[1:],folderName),shell=True)
                getConfigMap(args,folderName)
                shutil.make_archive(folderName, 'zip', folderName)
                output = subprocess.check_output("rm -r %s" %folderName, shell=True)
                print("Success, Logs zipped into %s.zip\n" %folderName)
                #Clean up
                print("Deleting backup pod and pod.yaml...\n")
                backupPodDeletion =  subprocess.check_output("kubectl delete pod custom-backup-pod -n %s" %args.namespace , shell=True)
                backupPodDeletion =  subprocess.check_output("rm pod.yaml", shell= True)
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
    parser.add_argument('-helmChart', '--helmChart', help='Helm Chart name' )
    parser.add_argument('-cluster', '--cluster', help='Cluster name' )
    parser.add_argument('-wait', '--wait', default= 10, help='Number of seconds to wait for the backup pod to start running before exiting\nDefault is 10 seconds' )
    parser.add_argument('-since', '--since',default='24h', help='For pods not having persistent volume storage the logs since a given time duration can be fetched.\nExample : mention the time as 2s(for 2 seconds) or 4m(for 4 mins) or 24h(for 24 hours)\nDefault is taken to be 24h' )
    args = parser.parse_args()

    if (not args.helmChart or not args.namespace or not args.cluster):
        print("Scripts requires arguments\nTry \'python3 get_logs.py --help\' for more info")
        exit()

    

    if(zipLogFile(args)==0):
        print("\nError getting log file\n")