import groovy.json.JsonBuilder
import java.util.zip.GZIPOutputStream

def zip(String s){
    def targetStream = new ByteArrayOutputStream()
    def zipStream = new GZIPOutputStream(targetStream)
    zipStream.write(s.getBytes('UTF-8'))
    zipStream.close()
    def zippedBytes = targetStream.toByteArray()
    targetStream.close()
    return zippedBytes.encodeBase64()
}

all_commits = []
def to_continue = true

def next_build = currentBuild

while(next_build) {
  for( def changeSet: next_build.changeSets){
    all_commits.addAll(changeSet.items)
  }
  next_build = next_build.previousBuild
  if(next_build && next_build.result.toString().equals('SUCCESS')){
    break
  }
}


def commits = []
for( def commit: all_commits){
  commit_detail = [:]
  commit_detail['id'] = commit.id.toString()
  commit_detail['title'] = commit.title.toString()
  commit_detail['author'] = commit.author.toString()
  commit_detail['authorEmail'] = commit.authorEmail.toString()
  commit_detail['authorTime'] = commit.authorTime.toString()
  commit_detail['committer'] = commit.committer.toString()
  commit_detail['committerEmail'] = commit.committerEmail.toString()
  commit_detail['committerTime'] = commit.committerTime.toString()
  commit_detail['paths'] = commit.paths.collect{ ['path': it.path, 'editType': it.editType.name] }
  commits.add(commit_detail)
}

// Initialize empty map for exporting environment variables as strings
def commits_map = [:]

// Initialize empty map for storing all desirable environment variables as 'raw' objects
def raw_commits_map = [:]

// Convert commit details array into JSON string
def builder = new JsonBuilder()
builder(commits)

// 'COMMITS' environment variable
commits_map['COMMITS'] = builder.toString()
raw_commits_map['COMMITS'] = commits
//builder = null

// Determine culprits i.e. list of all change authors & committers since the last successful build
// Store every culprit as a map item of 'name' and 'email'
def culprits = []
def authors = []
def committers = []
for(def commit: commits){

    culprits.add([name: commit['author'], 
                  email:commit['authorEmail']])

    culprits.add([name:commit['committer'], 
                  email:commit['committerEmail']])

                  
    authors.add([name: commit['author'], 
                  email:commit['authorEmail']])
    
    committers.add([name:commit['committer'], 
                  email:commit['committerEmail']])
                  
}


// Deduplicate the culprits and store them as JSON string in 'CULPRITS' environment variable
def unique_culprits = culprits.unique{ a, b -> a['email'] <=> b['email'] }
builder(unique_culprits)
commits_map['CULPRITS'] = builder.toString()
raw_commits_map['CULPRITS'] = unique_culprits

// Deduplicate the authors and store them as JSON string in 'AUTHORS' environment variable
def unique_authors = authors.unique{ a, b -> a['email'] <=> b['email'] }
builder(unique_authors)
commits_map['AUTHORS'] = builder.toString()
raw_commits_map['AUTHORS'] = unique_authors

// Deduplicate the committers and store them as JSON string in 'COMMITTERS' environment variable
def unique_committers = committers.unique{ a, b -> a['email'] <=> b['email'] }
builder(unique_committers)
commits_map['COMMITTERS'] = builder.toString()
raw_commits_map['COMMITTERS'] = unique_committers

// Find all culprits who don't have @avinetworks.com in their Email address
// Store them as JSON string in 'NO_AVI_CULPRITS' environment variable
def no_avi_culprits = unique_culprits.findAll{ !it['email'].contains('@avinetworks.com') }
builder(no_avi_culprits)
commits_map['NO_AVI_CULPRITS'] = builder.toString()
raw_commits_map['NO_AVI_CULPRITS'] = no_avi_culprits

// Store all culprits' Email addresses as a comma seaprated string in 'CULPRITS_EMAIL' environment variable
commits_map['CULPRITS_EMAIL'] = unique_culprits.collect{ it['email'] }.join(',')
raw_commits_map['CULPRITS_EMAIL'] = unique_culprits.collect{ it['email'] }.join(',')

// Store all 'no avi Email' culprits' Email addresses as a comma separated string in 'NO_AVI_CULPRITS_EMAIL' environment variable
commits_map['NO_AVI_CULPRITS_EMAIL'] = no_avi_culprits.collect{ it['email'] }.join(',')
raw_commits_map['NO_AVI_CULPRITS_EMAIL'] = no_avi_culprits.collect{ it['email'] }.join(',')

// Store all authors' Email addresses as a comma seaprated string in 'AUTHORS_EMAIL' environment variable
commits_map['AUTHORS_EMAIL'] = unique_authors.collect{ it['email'] }.join(',')
raw_commits_map['AUTHORS_EMAIL'] = unique_authors.collect{ it['email'] }.join(',')

// Store all committers' Email addresses as a comma seaprated string in 'COMMITTERS_EMAIL' environment variable
commits_map['COMMITTERS_EMAIL'] = unique_committers.collect{ it['email'] }.join(',')
raw_commits_map['COMMITTERS_EMAIL'] = unique_committers.collect{ it['email'] }.join(',')

builder = null

String origString = new JsonBuilder(raw_commits_map).toPrettyString()
String compressedEnvStr = zip(origString)

// No need to inject these as environment variables
// A minimum of emails of culprits, authors, and no-avi-culprits are sufficient for the Jenkins job
// The full details will be available in 'COMPRESSED_ENV_STR' env variable which basically contains
// all the environment variables as Map object --> strignfied --> compressed --> base64 encoded
// To retrieve the Map object, the following sequence of operations are necessary
// base64 decode --> uncompress -> bytes-to-string --> deserialize string to Map
commits_map.remove('COMMITS')
commits_map.remove('COMMITTERS')
commits_map.remove('CULPRITS')
commits_map.remove('AUTHORS')
commits_map.remove('NO_AVI_CULPRITS')

commits_map['COMPRESSED_ENV_STR'] = compressedEnvStr

return commits_map
