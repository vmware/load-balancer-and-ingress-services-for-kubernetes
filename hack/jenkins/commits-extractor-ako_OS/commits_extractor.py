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
import commit_helper
import logging
import sys
import os
import getpass
import argparse
import subprocess
import json

MESSAGE_LOGGING_LEVEL = 25
MESSAGE_LOGGING_NAME = 'MESSAGE'
        
logging.addLevelName(MESSAGE_LOGGING_LEVEL, MESSAGE_LOGGING_NAME)

# Environment variable name for GIT user
ENV_NAME_GITHUB_USER = 'GIT_USER'

# Environment variable name for GIT password
ENV_NAME_GITHUB_PASSWORD = 'GIT_PASS'

# Absolute path to the logging file
LOGGING_FILE_NAME = "git_commits_extractor.log"

# logging level
LOGGING_LEVEL = MESSAGE_LOGGING_LEVEL

# Location of the output file into which the developer commit information will be logged
FILE_LOCATION_COMMIT_INFO = "/tmp/commit-stats.json"

# Exclude commits list
EXCLUDE_COMMITS_LIST = []

# Name of the vmware/load-balancer-and-ingress-services-for-kubernetes GIT branch for which the dev commits must be fetched
BRANCH_NAME = 'dummy-nonexistent-branch'

# Since criterion value, timestamp for 'since-time' mode, commit SHA for 'since-commit' mode
SINCE_VALUE = ''

# Developer Commits only
IS_DEV_COMMITS_ONLY = True

# Minimum Rate Limit Threshold required to continue with script execution
# This represents the minumum number of Github API calls we need for a successful
# script execution
MINIMUM_GITHUB_RATELIMIT = 200


def _get_all_committers(commits_info):
    """ Returns all committers' information 
    The information includes name and Email
    """
    committers_unduplicated = [{'name': commit_info['committer'], 'email': commit_info['committerEmail']} 
                for commit_info in commits_info]
    
    all_names = []
    committers = []
    for committer in committers_unduplicated:
        if committer['name'] not in all_names:
            committers.append(committer)
            all_names.append(committer['name'])
    
    return committers


def _get_all_authors(commits_info):
    """ Returns all authors' information 
    The information includes name and Email
    """
    authors_unduplicated = [{'name': commit_info['author'], 'email': commit_info['authorEmail']} 
                for commit_info in commits_info]
    
    all_names = []
    authors = []
    for author in authors_unduplicated:
        if author['name'] not in all_names:
            authors.append(author)
            all_names.append(author['name'])
    
    return authors
    

def _get_all_culprits(commits_info):
    """ Returns all cuplrits' information
    culprits = unique( authors + committers )
    The information includes name and Email
    """
    authors = _get_all_authors(commits_info)
    committers = _get_all_committers(commits_info)
    
    culprits_unduplicated = authors + committers
    all_names = []
    culprits = []
    for culprit in culprits_unduplicated:
        if culprit['name'] not in all_names:
            culprits.append(culprit)
            all_names.append(culprit['name'])
            
    return culprits
    
def _get_non_avi_culprits(commits_info):
    """ Returns all non avi culprits' information
    culprits = unique( authors + committers )
    non_avi_culprits are culprits with non Avi Email address
    """
    culprits = _get_all_culprits(commits_info)
    
    non_avi_culprits = [culprit for culprit in culprits if '@avinetworks.com' not in culprit['email']]
    
    return non_avi_culprits


def _delete_output_file(file_location):
    """ Deletes the specified output log file """
    ret_status = True
    cmd = "rm -rf {}".format(file_location)
    logging.log(MESSAGE_LOGGING_LEVEL, "Deleting the previous {}".format(file_location))
    try:
        subprocess.check_output(cmd.split())
    except subprocess.CalledProcessError:
        ret_status = False
        info = sys.exc_info()
        logging.error("Error while deleting the file {}. ({},{})".format(
                file_location, info[0], info[1]))
    
    return ret_status
        

def _dump_to_output_file(json_dict, file_location):
    """ Writes the given commits information into the specified json file """
    with open(file_location,'w') as fobj:
        json.dump(json_dict, fobj, indent=2)
        
        
def _check_rate_limit(credentials_tuple):
    """ Determines if the user has the requisite number of Github API calls for
    the script to continue to execute
    :param credentials_tuple: credentials tuple representing (git-user-name, git-pass)
    :return True if rate limiting is not reached, False otherwise
    """
    return_value = False
    rate_obj = commit_helper.get_git_ratelimit(credentials_tuple=credentials_tuple)
    if not rate_obj:
        logging.error("Error retreiving the rate limit object")
    else:
        logging.log(MESSAGE_LOGGING_LEVEL, "Remaining API Calls: {}, Minimum Required: {}".format(
                rate_obj['rate']['remaining'], MINIMUM_GITHUB_RATELIMIT))
        if rate_obj['rate']['remaining'] >= MINIMUM_GITHUB_RATELIMIT:
            return_value = True
    
    return return_value

def _get_commits_since_time(credentials_tuple, 
                            branch,
                            since_timestamp,
                            exclude_list,
                            dev_commits_only):
    """Retrieves the processed information of all the commits since the specified
    time stamp.
    :param credentials_tuple: A tuple representing (git-user-name, git-password)
    :param branch: name of the ako Git repository branch
    :param since_timestamp: str object representing the ISO 8601 timestamp. All
                            commits from this timestamp are retrieved
    :param exclude_list: A list of commit SHA str objects which are to be excluded
                         from the result
    :param dev_commits_only: Whether only developer commits must be included in
                             the result. 
    :return list of dict objects representing the processed information of all
            the commits matching the find criterion
    """
    return commit_helper.get_git_commits(branch=branch,
                                         since_timestamp=since_timestamp,
                                         credentials_tuple=credentials_tuple,
                                         exclude_list=exclude_list,
                                         dev_commits_only=dev_commits_only)

def _get_commits_since_commit(credentials_tuple, 
                            branch,
                            since_commit,
                            exclude_list,
                            dev_commits_only):
    """Retrieves the processed information of all the commits since the specified
    commit SHA.
    :param credentials_tuple: A tuple representing (git-user-name, git-password)
    :param branch: name of the ako Git repository branch
    :param since_sha: str object representing the commit SHA.
                            The information for all the commits since this commit
                            is fetched. This commit is excluded from the result
    :param exclude_list: A list of commit SHA str objects which are to be excluded
                         from the result
    :param dev_commits_only: Whether only developer commits must be included in
                             the result. 
    :return list of dict objects representing the processed information of all
            the commits matching the find criterion
    """
    commits_info = []
    commit_info = commit_helper.get_git_commit(commit_sha=since_commit, 
                                               credentials_tuple=credentials_tuple)
    if not commit_info:
        logging.error("Error retrieving info for commit {}. Cannot proceed any further".format(
                since_commit))
    
    else:
        since_timestamp = commit_info['committerTime']
        exclude_list = [since_commit]
        logging.log(MESSAGE_LOGGING_LEVEL, "Examining all commits of {} branch since {}, excluding commit {}".format(
                branch, since_timestamp, since_commit))
        return commit_helper.get_git_commits(branch=branch,
                                             since_timestamp=since_timestamp,
                                             credentials_tuple=credentials_tuple,
                                             exclude_list=exclude_list,
                                             dev_commits_only=dev_commits_only)
    
    return commits_info

def _get_git_credentials(prompt=False):
    """ Retrieves the GIT credentials from either environment variables or from
    the user.
    :param prompt: Whether the GIT credentials must be retrieved through user input
    :return a tuple representing the credentials (git-user-name, git-pass)
    """
    git_user = git_pass = 'dummy-value'
    if not prompt:
        logging.log(MESSAGE_LOGGING_LEVEL, "Getting GIT credentials from environment variables")
        git_user = os.environ.get(ENV_NAME_GITHUB_USER, 'dummy-user')
        git_pass = os.environ.get(ENV_NAME_GITHUB_PASSWORD, 'dummy-pass')
        
    else:
        logging.log(MESSAGE_LOGGING_LEVEL, "Getting GIT credentials from user")
        git_user = input("Enter GIT username: ")
        git_pass = getpass.getpass("Enter GIT password: ")
    
    return git_user, git_pass

    
def _setup_logging():
    """ 
    Sets up global logging
    """
    level = LOGGING_LEVEL
    logging.basicConfig(level=level,
                        filename=LOGGING_FILE_NAME,
                        format='%(asctime)s %(levelname)s: %(message)s',
                        datefmt='%Y-%m-%d %I:%M:%S %p')
    

def _setup_globals(args):
    """ Initializes the global variables based on program arguments """
    global FILE_LOCATION_COMMIT_INFO
    global EXCLUDE_COMMITS_LIST
    global BRANCH_NAME
    global SINCE_VALUE
    global IS_DEV_COMMITS_ONLY
    
    # Add console logger if enabled    
    if args.enable_console_logging:
        logging.log(MESSAGE_LOGGING_LEVEL, "Enabling console logging")
        console_log_formatter = logging.Formatter('%(asctime)s %(levelname)s: %(message)s')
        console_log_handler = logging.StreamHandler()
        console_log_handler.setFormatter(console_log_formatter)
        
        root_logger = logging.getLogger()
        root_logger.addHandler(console_log_handler)
        
    # Enable verbose mode logging if required
    if args.verbose:
        logging.log(MESSAGE_LOGGING_LEVEL, "Enabling verbose logging")
        root_logger = logging.getLogger()
        root_logger.setLevel(logging.DEBUG)
    
    IS_DEV_COMMITS_ONLY = not args.all_commits
    
    BRANCH_NAME = args.branch
    SINCE_VALUE = args.since
    EXCLUDE_COMMITS_LIST = [] if not args.exclude_commits else [commit.strip() for commit in args.exclude_commits.split(',')]
    if args.extract_to:
        FILE_LOCATION_COMMIT_INFO = args.extract_to
                
    logging.log(MESSAGE_LOGGING_LEVEL, "Branch Name: {}".format(BRANCH_NAME))
    logging.log(MESSAGE_LOGGING_LEVEL, "Since Mode: {}".format(args.cmd))
    logging.log(MESSAGE_LOGGING_LEVEL, "Dev Commits Only: {}".format(IS_DEV_COMMITS_ONLY))
    logging.log(MESSAGE_LOGGING_LEVEL, "Exclude Commits List: {}".format(';'.join(EXCLUDE_COMMITS_LIST)))
    logging.log(MESSAGE_LOGGING_LEVEL, "Since Value: {}".format(SINCE_VALUE))
    logging.log(MESSAGE_LOGGING_LEVEL, "Output file location: {}".format(FILE_LOCATION_COMMIT_INFO))
    
def _setup_args():
    """ Sets up the program argument """
    parser = argparse.ArgumentParser(
            description='Extracts all commits of the given ako branch since the specified timestamp or since the specified commit SHA')

    sub_parsers = parser.add_subparsers()
    
    time_parser = sub_parsers.add_parser("since-time", help="This mode supports extraction of commits since the specified timestamp")
    commit_parser = sub_parsers.add_parser("since-commit", help="This mode supports extraction of developer commits since the specified commit SHA")

    time_parser.set_defaults(cmd='since-time')
    commit_parser.set_defaults(cmd='since-commit')
    
    for p in [time_parser, commit_parser]:
        p.add_argument('-b',
                       '--branch',
                       required=True,
                       action='store',
                       type=str,
                       help='The name of the vmware/load-balancer-and-ingress-services-for-kubernetes.git GIT repository branch for which the developer commits must be fetched')
    
        p.add_argument('-s',
                       '--since',
                       required=True,
                       action='store',
                       type=str,
                       help="Since criterion. {}".format(
                       "ISO 8601 timestamp" if id(p)==id(time_parser) else "Full commit SHA"))

        p.add_argument('-a',
                       '--all-commits',
                       required=False,
                       action='store_true',
                       default=False,
                       help='Whether all commits must be extracted. By default, only developer commits are extracted. default=False')

            
        p.add_argument('-x',
                       '--exclude-commits',
                       required=False,
                       action='store',
                       type=str,
                       help='The comma separated list of GIT commit SHAs to be disregarded while examining the dev commits')
    
        p.add_argument('-o',
                       '--extract-to',
                       required=False,
                       action='store',
                       type=str,
                       help="Full path to the output JSON file into which the info on developer commits must be logged. Script must have write/delete permission to this file. default='/tmp/commit-stats.json")
        
        p.add_argument('-v',
                       '--verbose',
                       required=False,
                       action='store_true',
                       default=False,
                       help='Verbose mode logging. If specified, the logging level is set to DEBUG. default=False')
            
        p.add_argument('-p',
                       '--prompt-credentials',
                       required=False,
                       action='store_true',
                       default=False,
                       help='Whether user must be promted to enter GIT credentials for making API calls. Otherwise, Git credentials are expected in {} and {} environment variables'.format(ENV_NAME_GITHUB_USER, ENV_NAME_GITHUB_PASSWORD))
        
        p.add_argument('--enable-console-logging',
                       required=False,
                       action='store_true',
                       default=False,
                       help='Enables console logging apart from logging into info.log file. default=False')
    
    
    args = parser.parse_args()
            
    return args


def main():
    """Main entry point function for program """
    _setup_logging()
    args = _setup_args()
    _setup_globals(args)
    
    logging.log(MESSAGE_LOGGING_LEVEL, "Deleting any output file if it exsists")
    if not _delete_output_file(file_location=FILE_LOCATION_COMMIT_INFO):
        sys.exit(1)
        
    logging.log(MESSAGE_LOGGING_LEVEL, "Retrieving GIT credentials")
    git_credentials = _get_git_credentials(args.prompt_credentials)
    
    logging.log(MESSAGE_LOGGING_LEVEL, "Validating Rate Limit consideration")
    if not _check_rate_limit(credentials_tuple=git_credentials):
        logging.error("Github Rate limiting reached. Cannot continue script execution")
        sys.exit(1)
    
    commits_info = None
    if args.cmd == 'since-time':                
        commits_info = _get_commits_since_time(
                            credentials_tuple=git_credentials,
                            branch=BRANCH_NAME,
                            since_timestamp=SINCE_VALUE,
                            exclude_list=EXCLUDE_COMMITS_LIST,
                            dev_commits_only=IS_DEV_COMMITS_ONLY)        
        
    elif args.cmd == "since-commit":
        commits_info = _get_commits_since_commit(
                            credentials_tuple=git_credentials,
                            branch=BRANCH_NAME,
                            since_commit=SINCE_VALUE,
                            exclude_list=EXCLUDE_COMMITS_LIST,
                            dev_commits_only=IS_DEV_COMMITS_ONLY)                
    
    if not commits_info:
        logging.warning("Could not retrieve any commits matching the find criteria. Not writing anything to the output file")

    else:
        full_commits_stats = {}
        full_commits_stats['AUTHORS'] = _get_all_authors(commits_info)
        full_commits_stats['COMMITTERS'] = _get_all_committers(commits_info)
        full_commits_stats['CULPRITS'] = _get_all_culprits(commits_info)
        full_commits_stats['NO_AVI_CULPRITS'] = _get_non_avi_culprits(commits_info)
        full_commits_stats['COMMITS'] = commits_info
        logging.log(MESSAGE_LOGGING_LEVEL, "-"*80)
        logging.log(MESSAGE_LOGGING_LEVEL, "Commits Details (Total: {})\n".format(
                len(commits_info)))
        for commit_info in commits_info:
            logging.log(MESSAGE_LOGGING_LEVEL, '#'*50)
            kv_pairs = commit_helper._get_commitinfo_as_kvpair(commit_info)
            for k,v in kv_pairs.items():
                logging.log(MESSAGE_LOGGING_LEVEL, "{}: {}".format(k,v))
            logging.log(MESSAGE_LOGGING_LEVEL, '#'*50 + "\n")
        logging.log(MESSAGE_LOGGING_LEVEL, "-"*80 + "\n")
        _dump_to_output_file(full_commits_stats, file_location=FILE_LOCATION_COMMIT_INFO)
    
    

if __name__ == '__main__':
    main()
