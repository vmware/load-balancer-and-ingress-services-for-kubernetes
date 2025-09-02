# Copyright 2019-2025 VMware, Inc.
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
		old_pods_gz_file=$LOG_FILE_PATH/old_logs.tar.gz
		if [ ! -f "$old_pods_gz_file" ]
		then
			tar -czf $old_pods_gz_file $POD_NAME*.log*
		else
			mkdir temp_dir
			tar -xzf $old_pods_gz_file -C temp_dir
			cp $POD_NAME*.log* temp_dir
			tar -czf $old_pods_gz_file -C temp_dir .
			rm -rf temp_dir
		fi		
	fi
	rm -rf $POD_NAME*.log*
}
zip_old_files
