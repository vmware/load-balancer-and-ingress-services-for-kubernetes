from collections import OrderedDict
from jinja2 import Environment, FileSystemLoader
from collections import namedtuple
from traceback import print_exc

import dateutil.parser as dateutil_parser
import dateutil.tz as dateutil_tz
import sys
import json
import os
import argparse
import pygal

# data type for injecting table data into Jinja2 template
TableDetail = namedtuple('TableDetail',
                         ['id', 
                          'heading', 
                          'description', 
                          'header_row', 
                          'data_rows',
                          'detail_data_rows_present'])

# data type for Table row
TableRow = namedtuple('TableRow', ['row', 'detail_rows'])

ChartDetail = namedtuple("ChartDetail", ['id','heading', 'description', 'chart'])

# Name of the Jinja2 template file
TEMPLATE_FILE = "gitChangeLogReportTemplate.html"

# Parent directory of this python script file
THIS_DIR = os.path.dirname(os.path.abspath(__file__))

# relative folder path to the Jinja2 templates
# relative from THIS_DIR
TEMPLATES_RELATIVE_DIR = os.path.join(THIS_DIR, "templates")

# Name of the html report file to be generated
REPORT_FILE_NAME = "GitChangeLogReport.html"

# Absolute path to the html report file to be generated
# This value is overridded by -r program paramter
REPORT_FILE_PATH = os.path.join(THIS_DIR, REPORT_FILE_NAME)

# Mapping of commit detail keys and the column display names
COMMITS_KEY_MAP = OrderedDict([
        ('id', 'ID'),
        ('title', "Title"),
        ('author', 'Author'),
        ('authorEmail', 'Author Email'),
        ('authorTime', 'Author Time'),
        ('committer', 'Committer'),
        ('committerEmail', 'Committer Email'),
        ('committerTime', 'Committer Time'),
    ])

# Mapping of culprit keys and the column display names
CULPRITS_KEY_MAP = OrderedDict([
        ('name', 'Name'),
        ('email', 'Email')
    ])

EDIT_TYPE_COLOR = OrderedDict([
                    ('add', 'green'),
                    ('edit', 'blue'),
                    ('delete', 'red'),
                    ('others', 'orange')
                ])

# Environment variable for commit details
ENV_NAME_COMMITS = 'COMMITS'

# Environement variable for culprits
ENV_NAME_CULPRITS = 'CULPRITS'

# Environment variable for no avi culprits
ENV_NAME_NO_AVI_CULPRITS = 'NO_AVI_CULPRITS'

# Environment variable for no avi culprits
ENV_NAME_AUTHORS = 'AUTHORS'

# Environment variable for no avi culprits
ENV_NAME_COMMITTERS = 'COMMITTERS'

# Branch variable name in the JINJA2 template file
JINJA_VAR_BRANCH = "branch"

# Build type variable name in the JINJA2 template file
JINJA_VAR_BUILD_TYPE = "build_type"

# Build number variable name in the JINJA2 template file
JINJA_VAR_BUILD_NUMBER = "build_number"

# Table details variable name in the JINJA2 template file
JINJA_VAR_TABLE_DETAILS = "table_details"

JINJA_VAR_CHART_DETAILS = "chart_details"

# program parameter for branch name
PARAM_BRANCH = ""

# program parameter for build type
PARAM_BUILD_TYPE = ""

# program parameter for build number
PARAM_BUILD_NUMBER = ""

# From local variable
FROM_LOCAL = False

# local json test file name
LOCAL_JSON_FILE_NAME = "changelog.json"

# local json file location
LOCAL_JSON_FILE_LOCATION = os.path.join(THIS_DIR, "test", LOCAL_JSON_FILE_NAME)

def get_env_value(key, default=None):
    """ Returns the environment variable's value
    :param key: name of the environment variable
    :default default: default value to be returned if environment variable is
    not defined

    :return environment variable's value if defined, default otherwise"""
    env_value = os.environ.get(key)
    return default if not env_value else env_value

# lambda to retrieve the commit details list object
get_commit_details = lambda: json.loads(get_env_value(ENV_NAME_COMMITS, "[]")) \
                                            if not FROM_LOCAL \
                                            else json.load(open(LOCAL_JSON_FILE_LOCATION))[ENV_NAME_COMMITS]

# lambda to retrieve the culprits details list object
get_culprits = lambda: json.loads(get_env_value(ENV_NAME_CULPRITS, "[]")) \
                        if not FROM_LOCAL \
                        else json.load(open(LOCAL_JSON_FILE_LOCATION))[ENV_NAME_CULPRITS]

# lambda to retrieve the 'no avi' culprits details list object
get_no_avi_culprits = lambda: json.loads(
                            get_env_value(ENV_NAME_NO_AVI_CULPRITS, "[]")) \
                        if not FROM_LOCAL \
                        else json.load(open(LOCAL_JSON_FILE_LOCATION))[ENV_NAME_NO_AVI_CULPRITS]

    
# lambda to retrieve the author details list object
get_author_details = lambda: json.loads(
                            get_env_value(ENV_NAME_AUTHORS, "[]")) \
                        if not FROM_LOCAL \
                        else json.load(open(LOCAL_JSON_FILE_LOCATION))[ENV_NAME_AUTHORS]

# lambda to retrieve the committer details list object
get_committer_details = lambda: json.loads(
                            get_env_value(ENV_NAME_COMMITTERS, "[]")) \
                        if not FROM_LOCAL \
                        else json.load(open(LOCAL_JSON_FILE_LOCATION))[ENV_NAME_COMMITTERS]

def get_header_row_cols(dict_obj):
    """ Returns the header row of the keys map dict object
    :param dict_obj: ordered dictionary object containing the mapping between
    the detail keys and column display names

    :return list of column display names
    """
    return list(dict_obj.values())

# lambda function to fetch the column display values for commit details
get_commit_details_header = lambda: get_header_row_cols(COMMITS_KEY_MAP)

# lamba function to fetch the column display values for culprits details
get_culprits_header = lambda: get_header_row_cols(CULPRITS_KEY_MAP)

# lambda function to fetch the column display values for 'no avi' culprits details
get_no_avi_culprits_header = lambda: get_header_row_cols(CULPRITS_KEY_MAP)

def _get_utc_iso8601(timestr):
    """ Converts a given ISO 8601 time string object of any time zone into an 
    UTC IS0 8601 time string object
    :param timestr: input ISO 8601 string, of any time zone
    return: ISO 8601 string of UTC timezone
    NOTE: In case an any error, the input string is returned after printing
    the error message on the console
    """
    returnValue = timestr
    try:
        datetimeobj = dateutil_parser.parse(timestr)
        newdatetimeobj = datetimeobj.astimezone(
                dateutil_tz.gettz('UTC')).replace(tzinfo=None)
        returnValue = newdatetimeobj.isoformat()
    except:
        print("Exception when trying to parse ISO 8601 time string {}".format(timestr))
        exc_info = sys.exc_info()
        print("Exception Details - {}: {}".format(exc_info[0], exc_info[1]))

    return returnValue

def print_commits_details():
    """ Prints the change list on the console """
    all_commits = get_commit_details()
    print("*" * 100 )
    print("GIT Change Log")
    print("*" * 100)

            
    if all_commits:
        for all_commit in all_commits:
            all_commit['id'] = "https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/commit/{}".format(all_commit['id'])
            all_commit['authorTime'] = _get_utc_iso8601(all_commit['authorTime'])
            all_commit['committerTime'] = _get_utc_iso8601(all_commit['committerTime'])
            print("=" * 80)
            print("\n".join(
                    "{:20}: {}".format(
                    COMMITS_KEY_MAP[k],
                    all_commit[k].encode('utf-8','ignore'))
                    for k in COMMITS_KEY_MAP.keys()))
            print("=" * 80)
    else:
        print('=' * 80)
        print("No Changes")
        print('=' * 80)


def get_data_row(detail_obj, keys):
    """ Retrieves the table data row of a given single detail object
    :param details_objs: a single detail dict object
    :param keys: a list of keys of the given dict object

    :return a single row i.e. list of table column values
    """
    return [detail_obj[k] for k in keys]
    
def get_data_rows(details_objs, keys):
    """ Retrieves the table data rows of a given details object
    :param details_objs: a list of details dict object
    :param keys: a list of keys of the given dict object

    :return a list of table rows
    """
    return [get_data_row(details_obj) for details_obj in details_objs]


def get_commit_table_details():
    """ Returns TableDetail object corresponding to commit information
    :param all_commits: list of all commit detail objects
    :return TableDetail object
    """
    all_commits = get_commit_details()
    for commit in all_commits:
        commit_id = commit['id']
        commit_url = "https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/commit/{}/".format(
                        commit_id)
        commit['id'] = "<a href=\"{}\" style=\"color:lightcoral\">{}</a>".format(
                commit_url, commit_id[:8])
        commit['title'] = commit['title'] if len(commit['title'])<=30 else commit['title'][:30] + "..."
        commit['authorTime'] = _get_utc_iso8601(commit['authorTime'])
        commit['committerTime'] = _get_utc_iso8601(commit['committerTime'])
        commit['detail_rows'] = []
        for commit_path_detail in commit['paths']:
            file_path = commit_path_detail['path']
            editType = commit_path_detail['editType']
            
            editTypeColor = EDIT_TYPE_COLOR.get(editType,
                                                EDIT_TYPE_COLOR['others'])
            
            blob_url = "https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/{}/{}".format(
                            commit_id, file_path)
            
            anchored_file_path = "<a href=\"{}\" style=\"color:{}\">{}</a>".format(
                    blob_url, editTypeColor, file_path)
            
            commit['detail_rows'].append([anchored_file_path, editType])

    return TableDetail(id="git-commits", heading="GIT Commits",
            description="List of all GIT commits picked by the build job",
            header_row=get_commit_details_header(),
            data_rows=[ TableRow(get_data_row(commit,COMMITS_KEY_MAP.keys()),
                        commit['detail_rows']) for commit in all_commits
                      ] if all_commits else [],
            detail_data_rows_present=True)
            
    
def get_all_table_details():
    """ Returns a list of all TableDetail objects, used for applying to the 
    Jinja2 template file
    :return a list of TableDetail objects
    """
    
    commits_details = get_commit_table_details()
    
    authors = get_author_details()
    authors_details = TableDetail(id="authors", heading="Authors",
                    description="List of Change authors",
                    header_row=get_culprits_header(),
                    data_rows=[ 
                        TableRow(get_data_row(author, CULPRITS_KEY_MAP.keys()),
                                 []) for author in authors
                            ] if authors else [],
                    detail_data_rows_present=False)

    committers = get_committer_details()
    committers_details = TableDetail(id="committers", heading="Committers",
                    description="List of all Change committers",
                    header_row=get_culprits_header(),
                    data_rows=[ 
                        TableRow(get_data_row(committer, CULPRITS_KEY_MAP.keys()),
                                 []) for committer in committers
                            ] if committers else [],
                    detail_data_rows_present=False)
                    
    culprits = get_culprits()
    culprits_details = TableDetail(id="all-contribs", heading="All Contributors",
                    description="List of all Change authors and change committers",
                    header_row=get_culprits_header(),
                    data_rows=[ 
                        TableRow(get_data_row(culprit, CULPRITS_KEY_MAP.keys()),
                                 []) for culprit in culprits
                            ] if culprits else [],
                    detail_data_rows_present=False)
                    
    no_avi_culprits = get_no_avi_culprits()
    
    no_avi_culprits_details = TableDetail(id="non-avi-contribs", 
                                          heading="Non-AVI Contributors",
                description="List of all Change authors/committers who don't have AVI Email address configured in their GIT configuration",
                header_row=get_culprits_header(),
                    data_rows=[ 
                        TableRow(get_data_row(no_avi_culprit, CULPRITS_KEY_MAP.keys()),
                                 []) for no_avi_culprit in no_avi_culprits
                            ] if no_avi_culprits else [],
                    detail_data_rows_present=False)
    
    all_table_details = [
                            commits_details,
                            authors_details,
                            committers_details,
                            culprits_details,
                            no_avi_culprits_details                
                        ]

    return all_table_details


def get_all_chart_details():
    """ Returns a list of all ChartDetail objects, used for applying to the 
    Jinja2 template file
    :return a list of ChartDetail objects
    """
    all_commits = get_commit_details()
    chart_details = []
    if all_commits:
        authors_dist = OrderedDict()
        f_dist = OrderedDict()
        get_edittype_count = lambda cobj, editType: len(
            list(filter(lambda item: item['editType'] == editType, cobj['paths'])))
        for commit in all_commits:
            author = commit['author']
            authors_dist[author] = authors_dist.get(author,0)+1
            f_dist.update(
                    {commit['id']: [get_edittype_count(commit, editType) for 
                     editType in ("add","edit","delete")]})
            f_dist[commit['id']].append(
                    len(commit['paths']) - sum(f_dist[commit['id']]))            
            
        
        pie_chart = pygal.Pie(print_values=True)
        pie_chart.title = "Commits Distribution By Authors"
        for key,value in authors_dist.items():
            pie_chart.add(key,value)
        author_chart = ChartDetail(id='authorChart',
                        heading='By Authors',
                        description='',
                        chart=pie_chart.render_data_uri() if authors_dist else None)
        
        chart_details.append(author_chart)
        
        bar_chart = pygal.StackedBar()
        
        bar_chart.title = "Commits Distribution By Files Count"
        
        bar_chart.x_labels = [k[:8] for k in f_dist.keys()]
        
        _get_commit_url = lambda commit_id: 'https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/commit/{}/'.format(commit_id)
        
        for index,item in enumerate(EDIT_TYPE_COLOR.keys()):
            bar_chart.add(item, 
                          [ {'value':v[index], 
                             'xlink': {'href':_get_commit_url(k), 
                                       'target':'_blank'} } 
                            for k,v in f_dist.items() ]) 
            
        files_chart = ChartDetail(id="filesChart", 
                        heading="By Files Count", 
                        description="", 
                        chart=bar_chart.render_data_uri() if f_dist else None)
        
        chart_details.append(files_chart)    
        
    return chart_details


def generateHtmlReport(filePath):
    """ Generates a HTML report report file containing the commits details,
    culprits list, and 'no avi' culprits list
    :param filePath: absolute path to the report file to be created

    :return None
    """
    fobj = open(filePath, 'w')

    j2_env = Environment(loader=FileSystemLoader(TEMPLATES_RELATIVE_DIR),
                         trim_blocks=True)
        
    table_details = get_all_table_details()            
    chart_details = get_all_chart_details()

    template = j2_env.get_template(TEMPLATE_FILE)

    html_str = template.render(
                    {
                        JINJA_VAR_BRANCH: PARAM_BRANCH,
                        JINJA_VAR_BUILD_TYPE: PARAM_BUILD_TYPE,
                        JINJA_VAR_BUILD_NUMBER: PARAM_BUILD_NUMBER,
                        JINJA_VAR_TABLE_DETAILS: table_details,
                        JINJA_VAR_CHART_DETAILS: chart_details
                    }
                )
    

    fobj.write(html_str)
    fobj.close()


def setup_args():
    """ Sets up the program argument """
    parser = argparse.ArgumentParser(
            description='Extracts GIT Change log and generates reports')
    parser.add_argument('branch', metavar='branch', type=str,
                        help=' AKO GIT branch name')
    parser.add_argument('build_type', metavar='build_type', type=str,
                        help='Build type (e.g. smoke, nightly')
    parser.add_argument('build_number', metavar='build_number', type=str,
                        help='Build number')

    parser.add_argument('-r',
                        '--report_file',
                        required=False,
                        action='store',
                        help='Absolute path to the report file to be generated',
                        default=REPORT_FILE_PATH)

    parser.add_argument('--file-mode',
                        required=False,
                        action='store_true',
                        help='Whether the script must fetch the inputs from {}/test/changelog.json file instead of environment variables. default=False'.format(THIS_DIR),
                        default=False)

    args = parser.parse_args()
    return args


def setup_globals(args):
    """Set up global variables with program arguments and override if necessary
    :param args: a namespace object of program arguments

    :return None
    """
    global PARAM_BRANCH
    global PARAM_BUILD_NUMBER
    global PARAM_BUILD_TYPE
    global REPORT_FILE_PATH
    global FROM_LOCAL

    PARAM_BRANCH = args.branch
    PARAM_BUILD_NUMBER = args.build_number
    PARAM_BUILD_TYPE = args.build_type
    FROM_LOCAL = args.file_mode

    if args.report_file:
        REPORT_FILE_PATH = args.report_file


def main():
    """ Main entry point function
    :return None
    """
    args = setup_args()
    setup_globals(args)
    print_commits_details()
    generateHtmlReport(REPORT_FILE_PATH)


if __name__ == "__main__":
    try:
        main()
    except SystemExit:
        pass
    except:
        info = sys.exc_info()
        print("Exception in the main function {} - {}".format(info[0], info[1]))
        print_exc(file=sys.stdout)

else:
    raise ImportError("Not an importable module")
