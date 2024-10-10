#!/bin/bash

set -xe

cd $WORKSPACE/hack/jenkins
sudo rm -rf /tmp/dummy || true 
version_tag=`/bin/bash get_build_version.sh  "dummy" $build_num`
build_tag=ci-build-${version_tag};
build_src=/mnt/builds/ako_OS/${branch};
src_build_path=$build_src/$build_tag;

tgt_build_path=$GCP_BUCKET_NAME/ako_OS/$branch/$build_tag

sudo rm -rf $WORKSPACE/venv
sudo rm -rf ~/.boto
sudo rm -rf ~/master-cloud-249007-mnt-read-write-sa.json

sudo cp /mnt/builds/build_j2m_backend/sync_ci_build/.boto ~/.boto
sudo cp /mnt/builds/build_j2m_backend/sync_ci_build/master-cloud-249007-mnt-read-write-sa.json ~/

sudo chmod 700 ~/master-cloud-249007-mnt-read-write-sa.json
sudo chmod 600 ~/.boto

sudo chown $USER:$USER ~/master-cloud-249007-mnt-read-write-sa.json ~/.boto

sudo apt-get -y install gcc python-dev
which virtualenv
whereis virtualenv
virtualenv -p python3 venv
source venv/bin/activate
pip install -U google-api-python-client
pip install -U gsutil setuptools
pip install requests==2.22.0
pip install -U crcmod

set +e
gsutil -m rsync -C -r $src_build_path $tgt_build_path
rc=$?;
set -e;

echo "INFO: Verifying if rclone installation is needed or not"
if ! [ -x "$(command -v rclone)" ] ; then
	wget https://downloads.rclone.org/v1.50.2/rclone-v1.50.2-linux-amd64.zip
	unzip *rclone*.zip
	cd rclone-*-linux-amd64

	sudo cp rclone /usr/bin/
	sudo chown root:root /usr/bin/rclone
	sudo chmod 755 /usr/bin/rclone

	sudo mkdir -p /usr/local/share/man/man1
	sudo cp rclone.1 /usr/local/share/man/man1/
	sudo mandb 

	cd ..
	sudo rm -rf *rclone* 
fi

if [[ $rc != 0 ]];then
	echo "INFO: Using Rclone.."
	sudo rclone sync --config /mnt/builds/build_j2m_backend/avi_rclone/avi_rclone.conf  -P -l $src_build_path $tgt_build_path
fi
