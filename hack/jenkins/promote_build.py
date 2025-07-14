# -*- coding: utf-8 -*-
# Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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

from argparse import ArgumentParser
from datetime import datetime
import os
import subprocess
import sys
import logging
import traceback
import shutil
import random
import string
import re
import json
import getpass

# All branch build folders pointed by the following symbolic links must be retained
ESSENTIAL_SYMLINKS = [
                        'last-built',
                        'last-last-built',
                        'last-good-smoke',
                        'last-last-good-smoke',
                        'last-candidate-nightly',
                        'last-good-nightly',
                        'last-last-good-nightly'
                     ]


# Different build promotion modes supported by the script
BUILD_PROMOTION_MODES = ( 'CI', 'SMOKE', 'NIGHTLY', 'NIGHTLY-CANDIDATE')

# Symlink pattern corresponding to different build promotion modes
# The actual symlink names can be constructoedby prefixing last- and last-last-
# to these patterns
SYMLINK_PATTERN = {
                    'CI': 'built',
                    'SMOKE': 'good-smoke',
                    'NIGHTLY': 'good-nightly',
                    'NIGHTLY-CANDIDATE': 'candidate-nightly'
                  }


# File location of base_build_num file, relative to the GIT workspace
FILE_LOCATION_BASE_BUILD_NUMBER = 'base_build_num'

# File location of version file, relative to the GIT workspace
FILE_LOCATION_BUILD_VERSION = 'hack/jenkins/get_build_version.sh'

# File location of version file, relative to the GIT workspace
FILE_LOCATION_BRANCH_VERSION = 'hack/jenkins/get_branch_version.sh'

# File location of the builds_details.json file, relative to the branch's
# builds archival location
FILE_LOCATION_BUILDS_DETAILS = "builds_details.json"

# File containing the head commit of the build, relative to the build folder
FILE_LOCATION_HEAD_COMMIT = "HEAD_COMMIT"

# Root folder of the builds archival location
ROOT_FOLDER_BUILDS_ARCHIVAL = '/mnt/builds/ako_OS'

# Common tag which appears at the beginning of every builds folder
COMMON_BUILDS_TAG = 'ci-build-'

# Common tag which appears at the beginning of every essential symlink
COMMON_SYMLINK_TAG = "last-"

# Maximum rentention hours for orphan folders
# Currently, set to 72 hours
MAXIMUM_RETENTION_HOURS = 72

# Maximum number of orphan build folders to be retained within MAXIMUM_RETENTION_HOURS
# Currently, set to 3 orphan builds
MAXIMUM_RETENTION_BUILDS = 3

# Default Log file location, relative to script execution folder
# Can be overridden by CLI argument -i / --logging_file
LOGGING_FILE_NAME = "./info.log"

# Default logging level
# Can be overridden by CLI argument -l / --logging_level
LOGGING_LEVEL = logging.INFO

# Environment variable name for Git user
ENV_VAR_GIT_USER = 'GIT_USER'

# Environment variable name for Git password
ENV_VAR_GIT_PASSWORD = 'GIT_PASS'

# Git username - to be extracted from ENV_VAR_GIT_USER environment variable or
# through user input
GIT_USER = 'nobody'

# Git password - to be extracted from ENV_VAR_GIT_PASS environment variable or
# through user input
GIT_PASS = 'my-very-private-password'

# Partial Github URL to ako GIT repository
URL_GITHUB_AVI_DEV = 'github.com/vmware/load-balancer-and-ingress-services-for-kubernetes.git'


def _setup_logging():
    """
    Sets up global logging
    """
    level = LOGGING_LEVEL
    logging.basicConfig(level=level,
                        format='%(asctime)s %(levelname)s: %(message)s',
                        datefmt='%Y-%m-%d %I:%M:%S %p')



def _execute_shell_command(cmd_str_list, cwd=None, env=None, quiet=True, use_bash=True):
    """ Executes the shell command
    :param cmd_str_list: list of ansible-playbook command string, including extra options
    :param cwd: current working directory from where the playbook is to be executed
    :param env: environment
    :param quiet: quite mode True/False
    :return: None
    """
    kargs = {}
    kargs['stdout'] = subprocess.PIPE if quiet else None
    kargs['stderr'] = subprocess.PIPE if quiet else None
    if cwd:
        kargs['cwd'] = cwd
    if env:
        kargs['env'] = env


    cmd_list = cmd_str_list
    if use_bash:
        cmd_list = ['bash'] + cmd_list

    logging.debug("Executing shell - {}".format(' '.join(cmd_list)))
    p = subprocess.Popen(cmd_list, **kargs)
    out, err = p.communicate()
    if out is not None:
        out = out.decode()
        
    if err is not None:
        err = err.decode()
        
    logging.debug("Shell out: {}; Shell err: {}".format(out,err))
    return out, err


def _get_git_ws():
    """ Gets the GIT checkout folder using git rev-parse command
    :return: str object. folder path of git checkout
    """
    ws = ''
    try:
        cmd = 'git rev-parse --show-toplevel'
        logging.debug("Executing command '{}'".format(cmd))
        ws = subprocess.check_output(cmd, shell=True)
        if ws is not None:
            ws = ws.decode()
    except:
        logging.error("Exception: {}".format(sys.exc_info()[0]))
        print('-'*60)
        traceback.print_exc(file=sys.stdout)
        print('-'*60)
    finally:
        ws = ws.strip()

    logging.info("GIT Workspace: {}".format(ws))
    return ws


def _get_build_version(build_number):
    """ Given the build number, retrieves the build version string for the currently
    checked out ako repository branch
    :param build_number: build number of the branch build for which build version
                         string is to be found
    :return: str object representing the version string for the given branch
             and build number, if successful. Raises exception otherwise
    """
    ws = _get_git_ws()
    if (not ws) or (not os.path.isdir(ws)):
        msg = "Error retrieving GIT workspace location"
        logging.error(msg)
        raise Exception(msg)
    else:
        script_path = os.path.join(ws, FILE_LOCATION_BUILD_VERSION)
        cmd_string = "{} {} {}".format(script_path, "dummy", build_number)
        out, err = _execute_shell_command(cmd_string.split())
        if err or not out:
            msg = "Could not retrieve build version string"
            logging.error(msg)
            raise Exception(msg)        
            
        return out.strip("\n")
    
    
def _get_branch_version():
    """ Retrieves the branch version string for the currently checked out ako
    repository branch
    :return: str object representing the branch version string, if successful. 
             Raises exception otherwise
    """
    ws = _get_git_ws()
    if (not ws) or (not os.path.isdir(ws)):
        msg = "Error retrieving GIT workspace location"
        logging.error(msg)
        raise Exception(msg)
    else:
        script_path = os.path.join(ws, FILE_LOCATION_BRANCH_VERSION)
        cmd_string = "{}".format(script_path)
        out, err = _execute_shell_command(cmd_string.split())
        if err or not out:
            msg = "Could not retrieve branch version string"
            logging.error(msg)
            raise Exception(msg)        
            
        return out.strip("\n") 


def _get_buildnumber_from_buildversion(build_version):
    """ Extracts and retrieves the branch build number from the build version
    string. This is the inverse of _get_build_version() function
    :param build_version: build version string
    :return integer representing the branch build number
    """
    ws = _get_git_ws()
    if (not ws) or (not os.path.isdir(ws)):
        msg = "Error retrieving GIT workspace location"
        logging.error(msg)
        raise Exception(msg)

    build_num_file = os.path.join(ws, FILE_LOCATION_BASE_BUILD_NUMBER)

    logging.debug("Fetching base build number from file {}".format(build_num_file))
    base_build_num = eval(open(build_num_file, 'r').read().strip())

    branch_version = _get_branch_version()
    pattern = r'^{}-(\S.*?)$'.format(branch_version)
    number = int(re.findall(pattern, build_version)[0])
    build_number = number - base_build_num
    logging.info("Extracted build number {} from build version {}".format(
            build_number, build_version))
    return build_number


def _get_buildversion_from_folddername(branch_build_location):
    """ Extacts and retrieves the build version string from the branch build
    folder name.
    :param branch_build_location: absolute or relative path to the branch build
                                  folder
    :return string object representing the build version of the build folder
    """
    basename = os.path.basename(branch_build_location)
    pattern = r'^{}(\S.*?)$'.format(COMMON_BUILDS_TAG)
    build_version = re.findall(pattern, basename)[0]
    logging.info("Extracted build version {} from build folder {}".format(
            build_version, branch_build_location))
    return build_version


def _get_buildnumber_from_foldername(branch_build_location):
    """ Extracts the build number i.e branch build number from the build folder's
    name.
    :param branch_build_location: absolute or relative path to the build folder
    :return integer value representing the branch build number
    """
    build_version = _get_buildversion_from_folddername(branch_build_location)
    return _get_buildnumber_from_buildversion(build_version)


def _get_root_archival_location():
    """ Retrieves the root folder archival location of all the branch builds
    The folder location, which is usually mounted NFS cache server, is returned
    if and only if it is not stale
    """
    cmd_string = "timeout 10s df -h {}".format(ROOT_FOLDER_BUILDS_ARCHIVAL)
    out, err = _execute_shell_command(cmd_string.split(), use_bash=False)
    if err or not out:
        msg = "Could not stat Root archival folder {}".format(
                ROOT_FOLDER_BUILDS_ARCHIVAL)
        logging.error(msg)
        raise Exception(msg)

    return ROOT_FOLDER_BUILDS_ARCHIVAL


def _get_branch_builds_location(branch_name):
    """ Retrieves the builds archival location for the given branch
    :param branch_name: Name of the ako GIT repository branch
    """
    branch_builds_location = os.path.join(_get_root_archival_location(),
                                          branch_name)

    # Check if path exists
    if os.path.isdir(branch_builds_location):
        return branch_builds_location
    else:
        msg = "Branch builds folder {} does not exist".format(branch_builds_location)
        logging.error(msg)
        raise Exception(msg)


def _get_branch_build_location(branch_name, build_number):
    """ Retrieves the absolute builds archival folder location of the given
    branch and build number
    :param branch_name: name of the ako GIT branch
    :param build_number: build number of the branch build
    :return str object representing the absolute path to the builds archival
            location
    """
    branch_build_location = os.path.join(
            _get_branch_builds_location(branch_name),
            COMMON_BUILDS_TAG + _get_build_version(build_number))

    # Check if path exists
    if os.path.isdir(branch_build_location):
        return branch_build_location

    else:
        msg = "Branch build folder {} for branch {} and build number {} doesn't exist".format(
                    branch_build_location, branch_name, build_number)
        logging.error(msg)
        raise Exception(msg)


def _get_all_branch_builds(branch_name):
    """ Retrieves all the build folders within the branch's build archival
    location
    :param branch_name: Name of the ako GIT repository branch
    :return list of absolute paths of build folders for this branch
    """
    branch_builds_location = _get_branch_builds_location(branch_name)

    return [os.path.join(branch_builds_location, folder) for
                folder in os.listdir(branch_builds_location) if
                os.path.isdir(os.path.join(branch_builds_location,folder)) and
                os.path.basename(folder).startswith(COMMON_BUILDS_TAG)]


def _get_all_branch_symlinks(branch_name):
    """ Retrieves all the build symlinks within the branch's build archival
    location
    :param branch_name: Name of the ako GIT repository branch
    :return list of absolute paths of build symlinks for this branch
    """
    branch_builds_location = _get_branch_builds_location(branch_name)

    return [os.path.join(branch_builds_location, link) for
            link in os.listdir(branch_builds_location) if
            os.path.islink(os.path.join(branch_builds_location,link)) and
            os.path.basename(link).startswith(COMMON_SYMLINK_TAG)]


def _delete_invalid_symlinks(symlinks):
    """ Filters out all invalid build symlinks and return the valid ones
    Filtering is done by finding and deleting all such essential build symlinks
    which either don't point to any target or are not having a directory as a
    target
    :param symlinks: list of absolute path to the symlinks
    :return list of valid symlinks
    """
    valid_symlinks = []
    for symlink in symlinks:
        logging.debug("Examining symlink {}".format(symlink))
        target = ""
        try:
            target = os.readlink(symlink)
        except:
            logging.error("Exception encountered when finding the target build folder for {}".format(
                    symlink))
            target = ""

        # Delete the symlink, if it is one of the essential build symlinks and
        # any one of the following
        # a. not pointing to any target
        # b. not pointing to any valid directory
        if (os.path.basename(symlink) in ESSENTIAL_SYMLINKS) and \
           ( not target or \
             not os.path.isdir(target) ):
               logging.debug("Symlink {} target {}".format(symlink, target))
               logging.debug("Symlink {} invalid. Deleting".format(symlink))
               os.unlink(symlink)
        else:
            logging.debug("Symlink {} valid".format(symlink))
            valid_symlinks.append(symlink)

    return valid_symlinks


def _get_temporary_symlink_name(branch_builds_location, length=10):
    """ Returns a temporary and random symbolic link name which starts with 'last-'
    :param branch_builds_location: - the base path or the parent folder to the symlink file
    :param length: length of the random string in the filename
    :return temporary filename
    """
    return os.path.join(branch_builds_location,
                        'last-' + ''.join(random.choice(string.ascii_lowercase)
                                           for i in range(length)))


def _cleanup_build_folders(branch_name):
    """ Cleans up the build folders in the branch's build archival location as
    per the retention policy
    :param branch_name: name of the ako GIT repository branch
    :return None
    """
    logging.debug("Cleaning up Build folders as per retention policy")
    all_build_folders = _get_all_branch_builds(branch_name)

    logging.debug("Listing all build folders:-\n" + "\n".join(all_build_folders))
    branch_builds_location = _get_branch_builds_location(branch_name)

    cleanup_folders_1 = []
    now = datetime.now()
    # Gather all build folders which are not targets of essential build symlinks
    for folder in all_build_folders:
        logging.debug("Examining build folder {}".format(folder))
        cleanup = True

        for essential_symlink in ESSENTIAL_SYMLINKS:
            symlink = os.path.join(branch_builds_location, essential_symlink)
            if os.path.islink(symlink):
                target = ""
                try:
                    target = os.readlink(symlink)
                except:
                    logging.error("Symlink {} points to invalid target".format(
                            essential_symlink))
                    target = ""

                if target and target==folder:
                    logging.debug("Symlink {} points to {}. Not marking for cleanup".format(
                                    essential_symlink, folder))
                    cleanup = False
                    break

        if cleanup:
            logging.debug("Marking Build folder {} for cleanup".format(folder))
            cleanup_folders_1.append(folder)

    logging.debug("Cleanup folders after symlink based filtering:- \n" + \
                  "\n".join(cleanup_folders_1))

    # Filter out build folders which are lying around for more than MAXIMUM_RETENTION_HOURS
    to_be_retained_folders = []
    cleanup_folders = []
    for folder in cleanup_folders_1:
        folder_modified_ts = datetime.fromtimestamp(os.path.getmtime(folder))
        delta = now - folder_modified_ts
        delta_hrs = delta.total_seconds() // 3600
        logging.debug("Build folder {} present since {}+ hours".format(
                folder, delta_hrs))
        if delta_hrs > MAXIMUM_RETENTION_HOURS:
            logging.debug("Marking build folder {} for cleanup".format(folder))
            cleanup_folders.append(folder)
        else:
            to_be_retained_folders.append(folder)

    logging.debug("Cleanup folders after MAXIMUM_RETENTION_HOURS based filtering:- \n" + \
                  "\n".join(cleanup_folders))

    # Now, from to_be_retained_folders, retain only MAXIMUM_RETENTION_BUILDS
    # Only the most recent folders are retained.
    cleanup_folders_2 = sorted(to_be_retained_folders,
                               key=lambda item: os.path.getmtime(item),
                               reverse=True)[MAXIMUM_RETENTION_BUILDS:]
    cleanup_folders.extend(cleanup_folders_2)

    logging.debug("To be deleted folders after all filtering: -\n" + \
                  "\n".join(cleanup_folders))

    for folder in cleanup_folders:
        shutil.rmtree(folder)


def _move_symlinks(branch_name,
                   build_number,
                   promotion_mode,
                   move_last_last=True,
                   strict=True):
    """ Moves the appropriate symlinks reliably for different build promotion
    modes. This function's responsibility is
    1. Delete invalid essential build symlinks
    2. to move last- and last-last- symlinks reliably
    3. update the json file
    :param branch_name: name of the ako GIT repository branch
    :param build_number: build number of the branch build
    :param promotion_mode: one of BUILD_PROMOTION_MODES values
    :param move_last_last: Whether last-last- symlink also needs to be moved.
                           Needs to be True if yes, False otherwise
    :param strict: Whether build promotion must go through if the build_number
                   is less recent than the already promoted build. For example,
                   In CI mode, if build 10 is being promoted when build 11 is
                   already promoted. If strict=True, such a promotion will fail.
                   If strict=False, such a promotion will go through.
    :return a dictionary object of created/moved symlinks and their new targets
    """
    pattern = SYMLINK_PATTERN.get(promotion_mode, None)
    if not pattern:
        msg = "Invalid promotion mode {}".format(promotion_mode)
        logging.error(msg)
        raise ValueError(msg)

    # Fetch the branch's builds archival location
    branch_builds_location = _get_branch_builds_location(branch_name)

    # Fetch the branch's specific build folder
    branch_build_location = _get_branch_build_location(branch_name, build_number)

    # Get all valid build symlinks
    # This also cleans up invalid symlinks in the branch's archival location
    build_symlinks = _delete_invalid_symlinks(
                            _get_all_branch_symlinks(branch_name))

    last_link = os.path.join(branch_builds_location, "last-" + pattern)
    last_last_link = os.path.join(branch_builds_location, "last-last-" + pattern)

    logging.debug("New {} target: {}".format(last_link, branch_build_location))

    moved_symlinks_info = {}

    # previous last symlink exists
    if last_link in build_symlinks:
        old_last_build = os.readlink(last_link)
        logging.debug("Previous {} symlink exists".format(last_link))
        logging.debug("Old {} target: {}".format(last_link, old_last_build))

        to_continue = True

        # if old_last_build is same as branch_build_location
        if old_last_build == branch_build_location:
            logging.warning("Old and New targets of {} are the same. Not continuing further".format(last_link))
            to_continue = False

        if to_continue:
            # If old_last_build is more recent than branch_build_location
            if(os.path.basename(old_last_build) > os.path.basename(branch_build_location)):
                msg = "The previous {} target i.e. {} is more recent than {}".format(
                        last_link, old_last_build, branch_build_location)
                if strict:
                    logging.error(msg)
                    raise ValueError(msg)
                else:
                    logging.warning(msg)

            if move_last_last:
                # Update last-last symlink to point to old_last_build
                logging.debug("Creating/Moving {} symlink to {}".format(
                        last_last_link, old_last_build))
                tmp_symlink = _get_temporary_symlink_name(branch_builds_location)
                os.symlink(old_last_build, tmp_symlink)
                os.rename(tmp_symlink, last_last_link)
                moved_symlinks_info[os.path.basename(last_last_link)] = old_last_build

            # Update last-built to branch_build_location
            logging.debug("Creating/Moving {} symlink to {}".format(
                    last_link, branch_build_location))
            tmp_symlink = _get_temporary_symlink_name(branch_builds_location)
            os.symlink(branch_build_location, tmp_symlink)
            os.rename(tmp_symlink, last_link)
            moved_symlinks_info[os.path.basename(last_link)] = branch_build_location

    # previous last symlink doesn't exist
    else:
        logging.debug("Previous {} symlink didn't exist. Creating one".format(last_link))
        os.symlink(branch_build_location, last_link)
        moved_symlinks_info[os.path.basename(last_link)] = branch_build_location

    return moved_symlinks_info

def _extract_build_head_commit(build_folder):
    """ Extracts and returns the builds' head commit information from a specific
    file (FILE_LOCATION_HEAD_COMMIT) within the build folder if the file exists.
    Else, returns empty string
    :param build_folder: Absolute location of the branch build folder
    :return string object representing the builds' head commit if successful,
            else an empty string object
    """
    head_commit_file_location = os.path.join(build_folder, FILE_LOCATION_HEAD_COMMIT)

    head_commit = ""
    if os.path.isfile(head_commit_file_location):
        logging.debug("Reading HEAD_COMMIT file {}".format(
                                    head_commit_file_location))
        try:
            with open(head_commit_file_location) as obj:
                head_commit = obj.read().strip()
        except:
            logging.error("Exception while reading HEAD_COMMIT file {}".format(
                            head_commit_file_location))
            head_commit = ""
    else:
        logging.warning("HEAD_COMMIT file {} not found".format(
                head_commit_file_location))

    return head_commit

def _update_builds_json(branch_name, moved_symlinks_info):
    """ Updates the builds mapping file in the branch's builds archival location
    with the latest info on the essential build symlinks and their respective
    target build folders
    :param branch_name: Name of the ako GIT repository branch
    :param moved_symlinks_info: a dictionary object with symlink name as the key
                                and the absolute path to the target build folder
    :return the map object between the newly moved/created symlinks and the info
            specific to their target build i.e. build version string, build
            folder path, build's head commit, etc
    """
    if not moved_symlinks_info:
        logging.info("No new essential build symlink created/moved. Nothing to update in JSON file")
        return dict()

    builds_json_file = os.path.join(_get_branch_builds_location(branch_name),
                                    FILE_LOCATION_BUILDS_DETAILS)


    mapping_info = {}
    if os.path.isfile(builds_json_file):
        logging.debug("{} exists. Reading from the same".format(
                builds_json_file))
        mapping_info = json.load(open(builds_json_file, 'r'))

    else:
        logging.warning("{} doesn't exist. Will be created shortly".format(
                                            builds_json_file))

    for symlink, build_folder in moved_symlinks_info.items():
        mapping_info[symlink] = {}
        mapping_info[symlink]['build_folder'] = build_folder
        build_version = _get_buildversion_from_folddername(build_folder)
        build_number = _get_buildnumber_from_buildversion(build_version)
        mapping_info[symlink]['build_version_string'] = build_version
        mapping_info[symlink]['jenkins_build_number'] = build_number
        mapping_info[symlink]['head_commit'] = _extract_build_head_commit(
                                                            build_folder)

    json.dump(mapping_info, open(builds_json_file,'w'), indent=2)

    # Retrieve mapping info for only changed/moved symlinks and return this
    # dictionary object
    changed_symlinks_mapping_info = {symlink:mapping_info[symlink] for symlink
                                     in moved_symlinks_info.keys()}

    return changed_symlinks_mapping_info


def _delete_tag(tag_name, ignore_errors=True):
    """ Deletes a tag from in both local and remote repository
    :param tag_name: Name of the tag to be deleted
    :return None
    """
    tag_local_delete = "git tag --delete {}".format(tag_name)
    out_local, err_local = _execute_shell_command(tag_local_delete.split(), use_bash=False)
    if not err_local:
        logging.info("Deleted tag {} from local successfully".format(tag_name))
        tag_remote_delete = 'git push -f https://{}:{}@{} :refs/tags/{}'.format(
                GIT_USER, GIT_PASS, URL_GITHUB_AVI_DEV, tag_name)
        out_remote, err_remote = _execute_shell_command(tag_remote_delete.split())

        if not err_remote:
            logging.warning("Unable to push the locally deleted tag {} to remote".format(
                    tag_name))
        else:
            logging.info("Deleted tag {} pushed to remote successfully".format(
                    tag_name))

        # Handle error if ignore_errors==False
        if not ignore_errors:
            pass
    # local tag delete operation has failed
    # The tag may not even exist if it is being created for the first time
    else:
        logging.warning("Deleting tag {} from local failed".format(tag_name))
        # Handle error
        if not ignore_errors:
            pass



def _create_push_tag(tag_name, commit, message):
    """ Creates an annotated tag for the given commit and with the given message,
    and also pushes the same to the remote
    :param tag_name: name of the annotated tag
    :param commit: commit sha to be associated with the tag
    :param message: message for the annotated tag
    :return None
    """
    if not tag_name:
        raise Exception("To create annotated tags, tag_name param cannot be empty or None")

    if not commit:
        raise Exception("To create annotated tags, commit param cannot be empty or None")

    if not message:
        raise Exception("To create annotated tags, message param cannot be empty or None")

    tag_local_create = 'git tag -a -f {} {} -m "{}"'.format(tag_name, commit, message)
    out_local, err_local = _execute_shell_command(tag_local_create.split(), use_bash=False)

    if not err_local:
        logging.debug("Tag {} against commit {} was successfully created locally".format(
                            tag_name, commit))
        tag_remote_create = 'git push -f https://{}:{}@{} tag {}'.format(
                        GIT_USER, GIT_PASS, URL_GITHUB_AVI_DEV, tag_name)

        # Not able to use _execute_shell_command here because subproces.Popen
        # is returning the output in stderr although the operation is successful
        # Therefore, as a workaround, using subprocess.check_output
        try:
            out_remote = subprocess.check_output(tag_remote_create.split())
            logging.debug("Successfully pushed annotated tag {} to remote".format(
                    tag_name))
            logging.debug("git push tag output: \n{}".format(out_remote))
        except subprocess.CalledProcessError as e:
            msg = "Failed to push new annotated tag {} to remote".format(tag_name)
            logging.error(msg)
            logging.exception(e)
            raise Exception(msg)

#        out_remote, err_remote = _execute_shell_command(tag_remote_create.split())
#        if err_remote:
#            msg = "Failed to push new annotated tag {} to remote"
#            logging.error(msg)
#            raise Exception(msg)
#        else:
#            logging.info("New annotated tag {} pushed to remote successfully".format(
#                    tag_name))
    else:
        logging.error("Local tag creation failed with the error:\n{}".format(err_local))
        raise Exception("Unable to create tag {} against commit {} locally".format(
                tag_name, commit))


def _create_push_build_tags(branch_name, mapping_info):
    """ Creates and pushes one or more annotated tags against the 'last-built' build's commit.
    :param branch_name: Name of the ako GIT repository branch
    :param mapping_info: A dict object containing information for different types
                         of essential builds i.e. last-* builds. 
                         This dict is expected to contain the info for last-built
                         build
    :return None
    """
    if not mapping_info:
        raise Exception("Mapping information object can't be empty or None")

    if 'last-built' not in mapping_info:
        raise Exception("'last-built' build information is missing in mapping_info object")

    commit = mapping_info['last-built']['head_commit']
    message = '''ako-{}-ci-build#{}'''.format(
            branch_name,
            mapping_info['last-built']['jenkins_build_number'])
    tag_name = mapping_info['last-built']['build_version_string']
    logging.debug("Attempting to create/push build tag {} for commit {}".format(
            tag_name, commit))
    _create_push_tag(tag_name, commit, message)


def _create_push_tags(branch_name, mapping_info):
    """ Creates and pushes annotated tags corresponding to the symlinks which
    have been moved/created. The moved/created symlink information is expected
    in mapping_info dictionary object.
    :param branch_name: Name of the ako GIT repository branch
    :param mapping_info: A dictionary object with the moved/created symlinks as
                         keys and the symlink's information as a dictionary object
    :return None
    """
    for symlink, info in mapping_info.items():
        tag_name = "{}-{}".format(branch_name, symlink)
        message = info['build_version_string']
        commit = info['head_commit']
        logging.debug("Attempting to create/push tag {} for commit {}".format(
                tag_name, commit))
        _delete_tag(tag_name)
        _create_push_tag(tag_name, commit, message)


def ci_mode(branch_name, build_number, strict=True):
    """ Execution function for CI script mode
    This function's responsibility is
    1. Delete invalid essential build symlinks
    2. to move last-built and last-last-built symlinks reliably
    3. update the json file
    4. Cleanup build folders as per the retention policy
    :param branch_name: name of the ako GIT repository branch
    :param build_number: build number of the branch build
    :param strict: Whether strict mode is enabled for Build promotion.
                   default=True
    :return a dictionary object containing the mapping between the symlink
            and target information associated with the symlink
    """
    symlinks_info = _move_symlinks(branch_name,
                   build_number,
                   'CI',
                   strict=strict)

    # Update builds_json file
    mapping_info = _update_builds_json(branch_name, symlinks_info)

    # clean up build folders as per retention policy
    _cleanup_build_folders(branch_name)

    return mapping_info


def smoke_mode(branch_name, build_number, strict=True):
    """ Execution function for SMOKE script mode
    This function's responsibility is
    1. Delete invalid essential build symlinks
    2. to move last-good-smoke and last-last-good-smoke symlinks reliably
    3. update the json file
    4. Cleanup build folders as per the retention policy
    :param branch_name: name of the ako GIT repository branch
    :param build_number: build number of the branch build
    :param strict: Whether strict mode is enabled for Build promotion.
                   default=True
    :return a dictionary object containing the mapping between the symlink
            and target information associated with the symlink
    """
    symlinks_info = _move_symlinks(branch_name,
                   build_number,
                   'SMOKE',
                   strict=strict)

    # Update builds_json file
    mapping_info = _update_builds_json(branch_name, symlinks_info)

    # clean up build folders as per retention policy
    _cleanup_build_folders(branch_name)

    return mapping_info


def nightly_mode(branch_name, build_number, strict=True):
    """ Execution function for NIGHTLY script mode
    This function's responsibility is
    1. Delete invalid essential build symlinks
    2. to move last-good-nightly and last-last-good-nightly symlinks reliably
    3. update the json file
    4. Cleanup build folders as per the retention policy
    :param branch_name: name of the ako GIT repository branch
    :param build_number: build number of the branch build
    :param strict: Whether strict mode is enabled for Build promotion.
                   default=True
    :return a dictionary object containing the mapping between the symlink
            and target information associated with the symlink
    """
    symlinks_info = _move_symlinks(branch_name,
                   build_number,
                   'NIGHTLY',
                   strict=strict)

    # Update builds_json file
    mapping_info = _update_builds_json(branch_name, symlinks_info)

    # clean up build folders as per retention policy
    _cleanup_build_folders(branch_name)

    return mapping_info


def nightly_candidate_mode(branch_name, build_number, strict=True):
    """ Execution function for NIGHTLY-CANDIDATE script mode
    This function's responsibility is
    1. Delete invalid essential build symlinks
    2. to move last-candidate-nightly symlink reliably
    3. update the json file
    4. Cleanup build folders as per the retention policy
    :param branch_name: name of the ako GIT repository branch
    :param build_number: build number of the branch build
    :param strict: Whether strict mode is enabled for Build promotion.
                   default=True
    :return a dictionary object containing the mapping between the symlink
            and target information associated with the symlink
    """
    symlinks_info = _move_symlinks(branch_name,
                   build_number,
                   'NIGHTLY-CANDIDATE',
                   move_last_last=False,
                   strict=strict)

    # Update builds_json file
    mapping_info = _update_builds_json(branch_name, symlinks_info)

    # clean up build folders as per retention policy
    _cleanup_build_folders(branch_name)

    return mapping_info

def _get_args():
    """
    Get CLI arguments
    """
    global LOGGING_LEVEL

    parser = ArgumentParser(description='''
This script performs different types of build promotion.
Build promotion mode, branch name, and build number are mandatory script arguments.
--verbose mode enables verbose logging.
--no-strict mode disables strict build promotion, which means that it is
possible to promote a less recent build. For example, if build 11 is already
promoted, --no-strict mode enables build 10 to be promoted. Without --no-strict,
such a build promotion leads to error.
--push-tags mode also creates/pushes annotated tags against the builds' commits
--prompt-credentials mode will enable user to enter GIT credentials, intead of
setting environment variables for the same
'''
)

    parser.add_argument('-m', '--mode',
                        required=True,
                        action='store',
                        choices = BUILD_PROMOTION_MODES,
                        help="Mode of build promotion")

    parser.add_argument('-b', '--branch-name',
                        required=True,
                        action='store',
                        help="Name of the ako GIT repository branch")

    parser.add_argument('-n', '--build-number',
                        required=True,
                        action='store',
                        help="Build number of the Jenkins branch build job")

    parser.add_argument('-c', '--commit',
                        help='GIT Commit SHA of the branch build',
                        required=False,
                        action='store',
                        default=None)

    parser.add_argument('-v', '--verbose',
                       help="Enables verbose logging",
                       action='store_true',
                       )

    parser.add_argument('--no-strict',
                        action='store_true',
                        default=False,
                        help="Disables strict mode build promotion"
                        )

    parser.add_argument('--push-tags',
                        action='store_true',
                        default=False,
                        help="""Creates annotated tags against the head commit/s and pushes them to remote.
To push tags, GIT_USER and GIT_PASS environment variables to be set before the script invocation, or they must \
be supplied by the user during script execution if executing in --prompt-credentials mode"""
                       )

    parser.add_argument('--prompt-credentials',
                        action='store_true',
                        default=False,
                        help='''Executing script in this mode enables the user to supply GIT credentials through \
user input rather than through environment variables'''
                    )


    args = parser.parse_args()

    # Set DEBUG logging level if verbose mode is enabled
    if args.verbose:
        LOGGING_LEVEL = logging.DEBUG

    return args


def _setup_globals(args):
    """ Sets up global variables
    :param args: command line parameters from argparse
    :return None
    """
    global GIT_USER
    global GIT_PASS

    if args.push_tags:
        logging.info("--push-tags mode enabled")
        if args.prompt_credentials:
            GIT_USER = input('Enter GIT username: ')
            GIT_PASS = getpass.getpass('Enter GIT password: ')
        else:
            logging.info("Looking for {} and {} environment variables".format(
                        ENV_VAR_GIT_USER, ENV_VAR_GIT_PASSWORD))
            GIT_USER = os.environ.get(ENV_VAR_GIT_USER, None)
            GIT_PASS = os.environ.get(ENV_VAR_GIT_PASSWORD, None)
            if not GIT_USER or not GIT_PASS:
                logging.error("GIT credentials unspecified or incomplete")
                raise ValueError('GIT credentials unspecified or incomplete')
    else:
        logging.info("--push-tags mode not enabled")

def main():
    args = _get_args()
    _setup_logging()
    _setup_globals(args)
    logging.debug("Executing in promotion mode {} for branch {} and build number {}".format(
            args.mode, args.branch_name, args.build_number))

    print(args)

    mapping_info = {}
    if args.mode == 'CI':
        mapping_info = ci_mode(args.branch_name,
                args.build_number,
                strict=not args.no_strict)

    elif args.mode == 'SMOKE':
        mapping_info = smoke_mode(args.branch_name,
                   args.build_number,
                   strict=not args.no_strict)

    elif args.mode == 'NIGHTLY':
        mapping_info = nightly_mode(args.branch_name,
                     args.build_number,
                     strict=not args.no_strict)

    elif args.mode == 'NIGHTLY-CANDIDATE':
        mapping_info = nightly_candidate_mode(args.branch_name,
                               args.build_number,
                               strict=not args.no_strict)

    else:
        raise ValueError("Invalid build promotion mode {}".format(args.mode))

    if args.push_tags and mapping_info:
        _create_push_tags(args.branch_name, mapping_info)
        if args.mode == 'CI':
            _create_push_build_tags(branch_name=args.branch_name,
                                    mapping_info=mapping_info)

if __name__ == '__main__':
    main()
