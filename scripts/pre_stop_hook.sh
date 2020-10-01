#!/bin/sh
set -e
function zip_old_files() {
	if [ $USE_PVC ]
	then
	cd $LOG_FILE_PATH
	old_log_tar_file=$LOG_FILE_PATH/old_logs.tar
	old_pods_gz_file=$LOG_FILE_PATH/old_logs.tar.gz
	if [ ! -f "$old_pods_gz_file" ]
	then
		tar -cf $old_log_tar_file $POD_NAME*.log*
 	else
		gunzip $old_pods_gz_file
		tar rf $old_log_tar_file $POD_NAME*.log*
        fi
	rm -rf $POD_NAME*.log*
	gzip $old_log_tar_file
    fi
}
zip_old_files
