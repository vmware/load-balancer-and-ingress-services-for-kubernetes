#!/bin/bash -xe 

REPORT_FILE=${REPORT_FILE:-"GitChangeLogReport.html"}

# Delete old report file/s
rm -rf $WORKSPACE/ako_cicd/jenkins/git-changelog/$REPORT_FILE
rm -rf $WORKSPACE/$REPORT_FILE

cd $WORKSPACE/ako_cicd/jenkins/git-changelog;

python -m pip install jinja2
python -m pip install pygal==2.4.0 --yes
python -m pip install python-dateutil

# Extract, print, and generate report file for GIT change logs
python gitChangeLogExtractor.py CI smoke $BUILD_NUMBER -r $REPORT_FILE --file-mode

# If previous script failed, at least touch an empty report file to prevent post-build action failure
if [ ! -f ./$REPORT_FILE ]; 
then
    echo "Git ChangeLog Report file not found. Creating an empty report file"
    touch ./$REPORT_FILE
fi

cp ./$REPORT_FILE $WORKSPACE/;

