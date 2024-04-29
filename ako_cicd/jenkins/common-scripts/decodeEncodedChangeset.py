#!/usr/bin/env python


"""
This snippet of Python code extracts all the injected variables by the post-SCM env injector step in <branch>-ci-build
The full details of the injected variables is expected in the environment variable COMPRESSED_ENV_STR.
Also, the absolute location of the git workspace is expected in the environment variable WORKSPACE.
But, the json dict object containing the injected variables is a string obtained by the following operation
dict object --> serialized --> compressed --> base64 encoded string

So, to extract the original json dict, the opposite sequence of operations must be performed.
base64 decoding --> decompression --> deserialization --> json object

This script must be added as an in-line execute shell build step immediately following the post-SCM env injector step
"""

import os
import json
import zlib
import base64

ws = os.environ.get('WORKSPACE')

# Retrieve the string which contains all the injected environment variables
# This string is compressed followed by base64 encoding
encoded_compressed_string = os.environ.get('COMPRESSED_ENV_STR')

decoded_compressed_bytes = base64.b64decode(encoded_compressed_string)


decoded_uncompressed_string = zlib.decompress(decoded_compressed_bytes, zlib.MAX_WBITS|32).decode('utf-8')


json_obj = json.loads(decoded_uncompressed_string)


file_path = os.path.join(ws, "ako_cicd/jenkins/git-changelog/test/changelog.json")


json.dump(json_obj, open(file_path,'w'), indent=2)

