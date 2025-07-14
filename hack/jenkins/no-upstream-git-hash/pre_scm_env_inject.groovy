/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/


import groovy.io.FileType
    
def get_essential_mapping(String branch_name, 
                String up_job, 
                String default_up_job,
                String build_number, 
                String no_upstream_git_hash,
                String branch_build_folder){

    // If upstream job is not defined, the same must be auto computed from ako_branch name
    if(up_job == null || up_job.isEmpty()){
        up_job =  default_up_job
    }
    
    boolean fetch_upstream_hash =  !(no_upstream_git_hash.toBoolean())
    String head_commit = branch_name

    if(fetch_upstream_hash) {
        // If no_upstream_git_hash is false, then fetch the head commit of the build-
        // from the build folder's HEAD_COMMIT file
        if(branch_name.isEmpty())
        {
            throw new Exception("branch parameter not specified or is empty")
        }
        if(build_number.isEmpty())
        {
            throw new Exception("'build number parameter not specified or is empty")
        }
        def dir = new File("${branch_build_folder}")
      
        // Compute CI Build folder pattern to be regex'd 
        
        int[] base_build_nums = [5000, 9000] // base build number is usually 5000 (for eng/master) or 9000 (for other MR branches)
        int build_number_int = "${build_number}".toInteger()
        String build_nums = base_build_nums.collect{(it+build_number_int).toString()}.join('|')
        build_nums = "(?:" + build_nums + ")"
        String ci_build_pattern = ".*ci-build-.*${build_nums}"
        println("Searching for pattern/s:     ${ci_build_pattern}")
      
        // Now regex the required CI build folder and store the same in ci_build_folder
        String ci_build_folder = ""
        dir.eachDir() { file ->
          println "Examining: " + file.toString()
          if(file.toString().matches(ci_build_pattern)){
            ci_build_folder = file.toString()
            println "CI Build Folder found: " + ci_build_folder
            return
          }
        }
        if(!ci_build_folder.isEmpty()){
          head_commit = new File("${ci_build_folder}/HEAD_COMMIT").getText('UTF-8').trim()
          if(head_commit.isEmpty()){
            throw new Exception("Could not determine for head commit for branch: ${branch_name} and build number: ${build_number}")
          }
          println("Setting head commit to ${head_commit}")
        }
        else{
          throw new Exception("CI Build folder for branch: ${branch_name} and build number: ${build_number} NOT FOUND in ${branch_build_folder}")
        }    
    }
    return ['head_commit': head_commit, 'up_job': up_job]
}

println("Computing and fetching head commit for avi-dev\n")
avi_dev_map = get_essential_mapping("${branch}", 
                                    "${upstream_job}", 
                                    "${branch}-ci-build", 
                                    "${build_num}",
                                    "${NO_UPSTREAM_GIT_HASH}",
                                    "/mnt/builds/${branch}")
                                    

println("\n\nComputing and fetching head commit for ako\n")
ako_map = get_essential_mapping("${ako_branch}", 
                                "${ako_upstream_job}", 
                                "ako-${ako_branch}-ci-build", 
                                "${ako_build_num}",
                                "${AKO_NO_UPSTREAM_GIT_HASH}",
                                "/mnt/builds/ako_OS/${ako_branch}")



// Inject HEAD_COMMIT environment variable                                
def map = [:]
map['HEAD_COMMIT'] = avi_dev_map['head_commit']
map['upstream_job'] = avi_dev_map['up_job']
map['AKO_HEAD_COMMIT'] = ako_map['head_commit']
map['ako_upstream_job'] = ako_map['up_job']

return map
