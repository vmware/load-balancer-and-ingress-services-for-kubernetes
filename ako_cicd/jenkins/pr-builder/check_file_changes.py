import os
import sys
from github import Github, Auth

def get_pr_files():
    # Get environment variables
    password = os.environ.get('GIT_PASSWORD')  # GitHub token for authentication
    pullid = os.environ.get('ghprbPullId')
    
    if not pullid:
        raise EnvironmentError("ghprbPullId must be set in environment variables")
    
    # Initialize GitHub client
    if password:
        # Use authentication if token is provided
        auth = Auth.Token(password)
        g = Github(auth=auth)
    else:
        # Use without authentication for public repositories (rate limited)
        g = Github()
    
    # Directly access the public repository
    repo = g.get_repo("vmware/load-balancer-and-ingress-services-for-kubernetes")
    
    # Get the pull request and its files
    pr = repo.get_pull(int(pullid))
    files = [f.filename for f in pr.get_files()]
    return files
    
def are_there_any_changes(folder_path):
    for fpath in get_pr_files():
        print(fpath)
        if folder_path in fpath:
            return True
    return False

if __name__ == '__main__':
    folder_path = sys.argv[1]
    if are_there_any_changes(folder_path):
        print(1)
    else:
        print(0)

