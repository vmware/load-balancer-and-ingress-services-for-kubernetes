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

# -*- coding: utf-8 -*-
from collections import OrderedDict
import config_helper
import common_helper
import requests
import logging

#Github API Url
URL_GITHUB = "https://api.github.com/"

# Github API url for vmware/load-balancer-and-ingress-services-for-kubernetes GIT repository
URL_GITHUB_AVI_DEV = URL_GITHUB + "repos/vmware/load-balancer-and-ingress-services-for-kubernetes/"

#Github headers
HEADERS_GITHUB = {
            'Accept': 'application/vnd.github.v3+json'
        }

# Github Commit resource name
RESOURCE_NAME_GITHUB_COMMITS = 'commits'

# Rate limit resource name in Github
RESOURCE_NAME_GITHUB_RATELIMIT = "rate_limit"

MAX_API_THREADS = 8


def _is_dev_commit(file_paths):  
    """ Given a list of file paths, this function determines if the commit is a 
    developer commit based on location of the file paths.
    A commit is a developer commit if at least one file path associated with the 
    commit
    1. is in the GIT root
    2. is not present in any of the blacklisted folders
    :param file_paths: list of file paths
    :return True if the commit is a developer commit, False otherwise    
    """
    dev_commit = False
    for file_path in file_paths:
        if config_helper.is_gitroot_file(file_path) or not \
           config_helper.in_blacklist_folder(file_path):
               dev_commit = True
               break
    
    return dev_commit

def _extract_file_info(git_commit_files):
    """ Given the raw file paths payload from the raw  Git commit payload, this
    function produces a list of modified file paths object
    :param git_commit_files: raw file paths payload from the raw Git commit payload
    :return a list of processed file path dict objects
    """
    paths = []
    for commit_file in git_commit_files:
        if commit_file['status'] == 'added':
            paths.append({'editType': 'add', 'path': commit_file['filename']})
        elif commit_file['status'] == 'modified':
            paths.append({'editType': 'edit', 'path': commit_file['filename']})
        elif commit_file['status'] == 'renamed':
            paths.append({'editType': 'add', 'path': commit_file['filename']})
            paths.append({'editType': 'delete', 'path': commit_file['previous_filename']})
        elif commit_file['status'] == 'deleted':
            paths.append({'editType': 'delete', 'path': commit_file['filename']})
        else:
            paths.append({'editType': 'others', 'path': commit_file['filename']})
            
    return paths


def _extract_commit_info(git_commit_payload):
    """Given the raw Git Commit payload, this function produces modified dict
    object representing the commit.
    :param git_commit_payload: raw Git commit payload i.e. dict object
    :return modifed dict object representing the original commit payload
    """
    commit_info = {}
    
    sha = git_commit_payload['sha']
    commit = git_commit_payload['commit']
    author = commit['author']['name']
    authorEmail = commit['author']['email']
    authorTime = commit['author']['date']
    title = commit['message']
    committer = commit['committer']['name']
    committerEmail = commit['committer']['email']
    committerTime = commit['committer']['date']
    commit_html_url = git_commit_payload['html_url']
    paths = _extract_file_info(git_commit_payload['files'])
    
    commit_info['id'] = sha
    commit_info['author'] = author
    commit_info['authorEmail'] = authorEmail
    commit_info['authorTime'] = authorTime
    commit_info['committer'] = committer
    commit_info['committerEmail'] = committerEmail
    commit_info['committerTime'] = committerTime
    commit_info['paths'] = paths
    commit_info['title'] = title
    commit_info['dev_commit'] = _is_dev_commit([path['path'] for path in paths])
    commit_info['html_url'] = commit_html_url
    
    return commit_info
    
    
def _get_commitinfo_as_kvpair(commit_info):
    """ Converts the given dict object representing the processed commit 
    information into key-value dict object. The dict object returned by the 
    function can be used for console logging
    :param commit_info: dict object representing the processed commit information
    :return a dict object having printable key value pairs of the commit
    """
    ret_dict = OrderedDict()
    ret_dict['SHA'] = commit_info['id']
    title = commit_info['title']
    ret_dict['Title'] = title[:40] + "..." if len(title)>40 else title
    ret_dict['Url'] = commit_info['html_url']
    ret_dict['Author'] = "{} ({})".format(commit_info['author'], 
                                          commit_info['authorEmail'])
    ret_dict['Author-Time'] = commit_info['authorTime']
    ret_dict['Committer'] = "{} ({})".format(commit_info['committer'], 
                                          commit_info['committerEmail'])
    ret_dict['Committer-Time'] = commit_info['committerTime']
    ret_dict['Is-Dev-Commit'] = commit_info['dev_commit']
    
    return ret_dict

    
def get_git_commit(commit_sha, credentials_tuple):
    """ Retrieves the processed commit information as a dict object, given the 
    commit SHA and credentials to access the Github API
    :param commit_sha: str object representing the commit SHA whose information
                       is to be retrieved and processed
    :param credentials_tuple: a tuple of (git-user-name, git-password)
    :return a dict object representing the commit information if successful,
            None if the operation was unsuccessful
    """
    url = URL_GITHUB_AVI_DEV + RESOURCE_NAME_GITHUB_COMMITS + "/"
    url = url + commit_sha
    commit_info = None
    logging.info("Fetching commit {} info from url {}".format(commit_sha,
                                                               url))
    response = requests.get(url, 
                            auth=credentials_tuple,
                            headers=HEADERS_GITHUB)
    if response.ok:
        payload = response.json()
        commit_info = _extract_commit_info(payload)
    else:
        logging.error("Couldn't not fetch info for commit {} ({},{})".format(
                commit_sha, response.status_code, response.reason))
        
    return commit_info


def _get_all_paginated_commits(branch, since_timestamp, credentials_tuple):
    """Given the branch name, the 'since' timestamp, and the credentials tuple
    to access the Github API, this function retrieves the raw information on all
    the commits matching criteria (branch and since_timestamp)
    :param branch: name of the ako Git repository branch
    :param since_timestamp: str object representing the ISO 8601 timestamp. All
                            commits from this timestamp are retrieved
    :param credentials_tuple: A tuple representing (git-user-name, git-password)
    :return list of dict objects representing the raw commit information from
            Github
    """
    url = URL_GITHUB_AVI_DEV + RESOURCE_NAME_GITHUB_COMMITS + "?"
    url += "sha={}&since={}&per_page=100&page=1".format(branch, since_timestamp) 
    to_continue = True
    all_commits = []
    while to_continue:
        logging.info("Attempting to get paginated commits using url {}".format(
                url))
        response = requests.get(url, 
                                auth=credentials_tuple,
                                headers=HEADERS_GITHUB)
        if response.ok:
            all_commits.extend(response.json())
            links = getattr(response, 'links', None)
            if links and 'next' in links.keys() and 'url' in links['next']:
                url = links['next']['url']
            else:
                to_continue = False
        else:
            logging.error("{} GET failed ({},{})".format(url, 
                                                         response.status_code,
                                                         response.reason))
            to_continue = False
    
    return all_commits


def get_git_commits(branch, 
                    since_timestamp, 
                    credentials_tuple,
                    exclude_list = list(),
                    dev_commits_only = True):
    """ Retrieves the processed information of all the commits belonging to the
    specified branch and since the specified timestamp.
    :param branch: name of the ako Git repository branch
    :param since_timestamp: str object representing the ISO 8601 timestamp. All
                            commits from this timestamp are retrieved
    :param credentials_tuple: A tuple representing (git-user-name, git-password)
    :param exclude_list: A list of commit SHA str objects which are to be excluded
                         from the result
    :param dev_commits_only: Whether only developer commits must be included in
                             the result. 
    :return list of dict objects representing the processed information of all the
            commits matching the criteria
    """
    logging.info("Fetching all commits of {} branch since {}".format(branch, since_timestamp))
    all_commits = _get_all_paginated_commits(branch, since_timestamp, credentials_tuple)
    logging.info("Total commits count: {}".format(len(all_commits)))
    logging.info("Excluding the following commits\n {}".format('\n'.join(exclude_list)))
    filtered_commits = [commit for commit in all_commits if commit['sha'] not in exclude_list and commit['commit']['committer']['date'] != since_timestamp]
    logging.info("Filtered commits count: {}".format(len(filtered_commits)))
    
    commits_info = []
    thread_objs = []
    
    def _aggregate_fn():
        details = common_helper.dispatch_and_join(thread_objs)
        for detail in details:
            if detail:
                commits_info.append(detail)
            else:
                logging.error("A thread returned None instead of commit information")
                
                
    for f_commit in filtered_commits:
        sha = f_commit['sha']
        thread_objs.append(common_helper.thread_wrapper(sha, get_git_commit,
                                                        sha, credentials_tuple))
        
        if len(thread_objs) == MAX_API_THREADS:
            _aggregate_fn()
            thread_objs.clear()
            
    if thread_objs:
        _aggregate_fn()
        thread_objs.clear()
            
    filtered_commits_info = commits_info
    if dev_commits_only:        
        filtered_commits_info = [commit_info for commit_info in commits_info if
                                 commit_info['dev_commit']]
        logging.info("Developer commits count: {}".format(len(filtered_commits_info)))
        
    return filtered_commits_info
    
    
def get_git_ratelimit(credentials_tuple):
    """ Retrieves the raw Git rate limit information for the user specified 
    in the credentials tuple
    :param credentials_tuple: credentials representing (git-user-name, git-pass),
                              used for accessing the Github
    :return dict object representing the raw rate limit information from Github
    """
    url = URL_GITHUB + RESOURCE_NAME_GITHUB_RATELIMIT
    rate_limit = {}
    response = requests.get(url, 
                            headers=HEADERS_GITHUB,
                            auth=credentials_tuple)
    logging.debug("Attempting to get rate limit using url {}".format(url))    
    if response.ok:
        rate_limit = response.json()
    else:
        logging.error("Error retrieving rate limit information ({},{})".format(
                response.status_code, response.reason))
    
    return rate_limit
