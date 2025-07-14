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
