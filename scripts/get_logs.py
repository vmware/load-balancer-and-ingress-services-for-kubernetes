import os
import argparse
import subprocess
import re
import shutil
import yaml
import time
import warnings
warnings.filterwarnings("ignore")

def findPVCName(templateSpec):
    #Return PV name of ako pod
    try: 
        pvcName = templateSpec['volumes'][0]['persistentVolumeClaim']['claimName']
        return pvcName
    except KeyError:
        print "Persistent Volume for pod is not defined\nReading logs directly from the pod\n"
        subprocess.check_output("kubectl logs %s -n %s > log_files/log_file" %(podName,args.namespace) , shell=True)
        shutil.make_archive('LogsZip', 'zip', 'log_files')
    
def findPVMount(templateSpec):
    #Return mount place of PV of ako pod
    try: 
        pvMount = templateSpec['containers'][0]['volumeMounts'][0]['mountPath']
        return pvMount
    except KeyError:
        print "Persistent Volume Mount for pod is not defined\nMounting the log files to /log path\n"
        return "/log"
    
def findPVDetails(args) : 
    # Take details of PV and PVC from the ako pod helm chart
    helmResult = subprocess.check_output("helm get all %s -n %s" %(args.helmChart,args.namespace) , shell=True)
    listhelmResult = helmResult.split("---")
    for output in listhelmResult:
        dictOutput = yaml.safe_load(output)
        try:
            templateSpec = dictOutput['spec']['template']['spec']
            pvcName = findPVCName(templateSpec)
            pvMount = findPVMount(templateSpec)
            return pvcName, pvMount
        except KeyError:
            continue

def editDeploymentFile(pvcName,pvMount):
    stream = file('pod.yaml', 'r')
    deploymentDict = yaml.load(stream)   
    deploymentDict['spec']['containers'][0]['volumeMounts'][0]['mountPath'] = pvMount
    deploymentDict['spec']['volumes'][0]['persistentVolumeClaim']['claimName'] = pvcName
    with open('pod.yaml', 'w') as outfile:
        yaml.dump(deploymentDict, outfile, default_flow_style=False)

def findPodName(args):
    Pods = subprocess.check_output("kubectl get pod -n %s -l app.kubernetes.io/instance=%s" %(args.namespace, args.helmChart) , shell=True)
    newLine = Pods.find("\n")
    wordEnd = Pods.find(" ",newLine+1)
    podName = Pods[newLine+1:wordEnd]
    return podName

def zipLogFile (args):
    podName = findPodName(args)
    try:
        #Find details of the Ako pod
        statusOfAkoPod =  subprocess.check_output("kubectl describe pod %s -n %s" %(podName,args.namespace) , shell=True)
    except:
        #If details couldnt be, and the kubectl describe raises any exception, then return failure
        return 0
    pvcName, pvMount = findPVDetails(args)
    #Check if the ako pod is up and running
    if (re.findall("Status: *Running", statusOfAkoPod)):
        #If ako pod is running, copy the log file to zip it
        try:
            copyOutput = subprocess.check_output("kubectl cp %s/%s:%s/avi.log log_files/log_file" %(args.namespace,podName,pvMount[1:]), shell=True)
        except:
            return 0
        shutil.make_archive('LogsZip', 'zip', 'log_files')
        return 1
    #If ako pod isnt running, then create backup pod named "mypod"
    else:
        #Creation of "mypod"
        editDeploymentFile(pvcName,pvMount)
        try:
            podCreated = subprocess.check_output("kubectl apply -f pod.yaml", shell=True)
        except:
            return 0
        timeout = time.time() + 10
        #Wait for "mypod" to start running
        while(1):
            try:
                statusOfBackupPod =  subprocess.check_output("kubectl describe pod custom-backup-pod -n %s" %args.namespace , shell=True)
            except: 
                return 0
            if (re.findall("Status: *Running", statusOfBackupPod)):
                #Once "mypod" is running, copy the log file to zip it
                print "\nBACKUP POD RUNNING"
                copyOutput = subprocess.check_output("kubectl cp %s/custom-backup-pod:%s/avi.log log_files/log_file" %(args.namespace,pvMount[1:]),shell=True)
                shutil.make_archive('LogsZip', 'zip', 'log_files')
                backupPodDeletion =  subprocess.check_output("kubectl delete pod custom-backup-pod -n %s" %args.namespace , shell=True)
                print "\nBACKUP POD DELETED"
                return 1
            time.sleep(2)
            if time.time()>timeout:
                break
        print "\nCOULDN'T CREAT BACKUP POD\n"
    return 0

if __name__ == "__main__":
    #Parsing cli arguments
    parser = argparse.ArgumentParser(formatter_class=argparse.RawTextHelpFormatter)
    parser.add_argument('-n', '--namespace', help='namespace' )
    parser.add_argument('-helmChart', '--helmChart', help='Helm Chart' )
    args = parser.parse_args()

    if(zipLogFile(args)==0):
        print "\nError getting log file\n"
    else : 
        print "\nSuccess, Logs zipped into LogsZip.zip\n"