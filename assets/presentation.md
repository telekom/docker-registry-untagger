![](logo.png)

# Docker Registry Untagger

A tool to cleanup your docker registry

---

<!-- page_number: true -->
## Builds Builds Builds

---
## Tags Tags Tags

---
## Solution
A tool that removes old builds but keeps the important releases

---

## How?
Define multiply regex to define which tags should stay
`[A-Za-z]+_release_[0-9]+`
`[A-Za-z]+_prod_[0-9]+`
`[A-Za-z]+_builds_[0-9]+`

---

## How?
Define a regex to tell the tool how to sort build tags
`([A-Za-z]+)_builds_([0-9]+)`
`(?P<buildnr>[0-9]+)_(?P<flavor>[A-Za-z]+)`
`centos_([0-9]+)`


---

## Difficulties?
Currently tags aren't first class citizens.

---

## Example `rules.yml`

```yml
repositories:
  - transporter/sam3
  - transporter/yak
  - transporter/tomcat

validTags:
  - [A-Za-z]+_release_[0-9]+
  - [A-Za-z]+_builds_[0-9]+

buildSortRegex: ([A-Za-z]+)_build_([0-9]+)
keepBuilds: 20

minAgeBeforeDelete: 5
```

---


## Example `config.yml`

```yml
host: http://workbench:5000
user: username@telekom.de
password: geheim123!
poolSize: 2
parallelDownloads: 100
```

---

# Demo
