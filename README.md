![](assets/logo.png)

*Copyright (c) 2017 Steffen Windoffer, Deutsche Telekom AG*

# Docker Registry Untagger
Untag and remove old and unwanted container images!

# Introduction
Docker Registry Untagger was developed at DTAG for internal use and later open sourced.
With continuous integration a lot of maybe unused tags will be created and pushed to the registry. To clean up the registry and save space this projects sorts the tags by build number and removes tags that are old enough. This also works with multiple flavor build tags and it can be setup to ignore release or any other special kind of tag. This all happens with the regex magic!


# Getting started
Your Docker Registry needs to be started with `REGISTRY_STORAGE_DELETE_ENABLED=true`

* Install this Project and Setup the Config Files

* Run it, for safety reasons maybe do a `-dryRun` first. Tags will be removed but images will stay until the garbage-collector got executed

* To finally clean up all the unused images run `bin/registry garbage-collect [--dry-run] /path/to/config.yml`

# Installation
`go install github.com/telekom/docker-registry-untagger`

or download/compile a binary on your own and drop it into the bin folder

# Configs
## Example `config.yml`
```yml
host: http://localhost:5000
user: username
password: password
poolSize: 3
parallelDownloads: 100
```

## Description `config.yml`
* host: the full hostname with protocol and port
* user: username to connect with the registry
* password: the password to connect
* poolSize: how many repos should be scanned simultaneously
* parallelDownloads: number of concurrent api calls that should be exectued against the registry

## Example `rules.yml`
```yml
repositories:
  - testrepo
  - otherrepo

validTags:
  - [A-Za-z]+_release_[0-9]+
  - [A-Za-z]+_builds_[0-9]+

keepBuilds: 2
buildSortRegex: ([A-Za-z]+)_builds_([0-9]+)

minAgeBeforeDelete: 5
```

## Description `rules.yml`
* repositories: a list of repositories that should be cleand up
* validTags: a list of regexes that decribe which tags should be keeped
* keepBuilds: the number of desired build tags that should be kept
* buildSortRegex: a regex that is used to get the build tags and sort them with help of the build number. if you only have one pair of parentheses, they contain the build number. if you have multiply parentheses the first pair marks the flavor and the second marks the build number. this is usefull if a repo contains multiply versions e.g centos5, centos6, centos7. if you have multiply parentheses or the flavor isnt group 1 and buildnr isnt group 2 you can set a custom order with the help of group names. e.g `(?P<buildnr>[0-9]+)_(?P<flavor>[A-Za-z]+)`
* minAgeBeforeDelete: minimum number of age (in days) a container needs to have before it is considered for deletion regardless of marking for removal

## Commandline Args
```bash
docker-registry-untagger --help
  -config string
        the config file (default "config.yml")
  -dryRun
        dont remove images (default false)
  -insecure
        allowe insecure connection to the docker registry (default false)
  -rules string
        the rule file (default "rules.yml")
```

## Outlook
Current Tags are not first class. This means if 2 tags point to the same digest and the digest gets removed both tags are gone, because of that there is a safety check in this tool. If a tag is marked for deletion but another tag which points to the same tag is not marked, both tags will stay, since it is not possible to just delete a tag. As long as not all tags that point to one digest get marked for deletion all tags will stay. This *feature* can be removed if a tag will be first class (e.g https://github.com/docker/distribution/pull/2169, https://github.com/docker/distribution/pull/2170 and further get merged)

## License
All files are licensed under the Apache-2.0 license. (see [License file](LICENSE))
