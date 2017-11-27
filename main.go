// docker-unregstriy-untagger :- main module
// Copyright (c) 2017, Steffen Windoffer, Deutsche Telekom AG
// Contact: opensource@telekom.de
// This file is distributed under the conditions of the Apache2 license.
// For details see the files LICENSE at the toplevel.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/opencontainers/go-digest"
	"github.com/wind0r/docker-registry-client/registry"

	"gopkg.in/yaml.v1"
)

type config struct {
	Host              string `yaml:"host"`
	User              string `yaml:"user"`
	Password          string `yaml:"password"`
	PoolSize          int    `yaml:"poolSize"`
	ParallelDownloads int    `yaml:"parallelDownloads"`
}

type rule struct {
	Repositories   []string `yaml:"repositories"`
	ValidTags      []string `yaml:"validTags"`
	ValidTagsRegex []*regexp.Regexp

	SortAndFilter      string `yaml:"buildSortRegex"`
	SortAndFilterRegex *regexp.Regexp
	KeepNewestBySort   int `yaml:"keepBuilds"`

	MinAge int `yaml:"minAgeBeforeDelete"`
}

type tagFlavor struct {
	name   string
	number int
}

type tagFlavors []tagFlavor

func (t tagFlavors) Len() int {
	return len(t)
}

func (t tagFlavors) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t tagFlavors) Less(i, j int) bool {
	return t[i].number > t[j].number
}

type layer struct {
	Created time.Time `json:"created"`
}

var (
	cfg       config
	rules     rule
	pool      chan bool
	downloads chan bool

	dryRun   *bool
	insecure *bool
	hub      *registry.Registry
)

func init() {
	dryRun = flag.Bool("dryRun", false, "dont remove images (default false)")
	insecure = flag.Bool("insecure", false, "allowe insecure connection to the docker registry (default false)")
	configFileName := flag.String("config", "config.yml", "the config file")
	rulesFileName := flag.String("rules", "rules.yml", "the rule file")
	flag.Parse()

	configFile, err := ioutil.ReadFile(*configFileName)
	if err != nil {
		log.Fatal("Config file is missing: config.yml\n", err)
	}

	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		log.Fatal("credentials file is malformed\n", err)
	}

	rulesFile, err := ioutil.ReadFile(*rulesFileName)
	if err != nil {
		log.Fatal("Config file is missing: rules.yml\n", err)
	}

	if err := yaml.Unmarshal(rulesFile, &rules); err != nil {
		log.Fatal("rules file is malformed\n", err)
	}

	sort.Strings(rules.Repositories)

	verifyRules()

	pool = make(chan bool, cfg.PoolSize)
	downloads = make(chan bool, cfg.ParallelDownloads)
}

func verifyRules() {
	if len(rules.Repositories) == 0 {
		log.Fatalf("atleast one repositories needes to be added")
	}

	if len(rules.ValidTags) != 0 {
		for _, repo := range rules.ValidTags {
			regex, err := regexp.Compile(repo)
			if err != nil {
				log.Fatalf("some tag regexp isnt valid (%s)", err)
			}
			rules.ValidTagsRegex = append(rules.ValidTagsRegex, regex)
		}
	} else {
		log.Fatalf("atleast one tag regex needs to be added")
	}

	regex, err := regexp.Compile(rules.SortAndFilter)
	if err != nil {
		log.Fatalf("sort release regex isnt valid (%s)", err)
	}
	rules.SortAndFilterRegex = regex
}

func removeImage(repo string, digest digest.Digest) {
	err := hub.DeleteManifest(repo, digest)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
}

// filterOlderTagsn returns all tags that are older then age
func oldTags(age int, repo string) func(string) bool {
	return func(tag string) bool {
		if age < 0 {
			return false
		}

		if age == 0 {
			return true
		}

		downloads <- true
		defer func() { <-downloads }()

		mani, err := hub.Manifest(repo, tag)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}

		reader, err := hub.DownloadLayer(repo, mani.References()[0].Digest)
		defer reader.Close()
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}

		b, err := ioutil.ReadAll(reader)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}

		l := layer{}
		err = json.Unmarshal(b, &l)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}

		if time.Now().Sub(l.Created) >= time.Duration(age)*24*time.Hour {
			return true
		}

		return false
	}
}

func getFlavor(keepRegex *regexp.Regexp, tags []string) map[string]tagFlavors {
	flavor := make(map[string]tagFlavors)

	if len(keepRegex.SubexpNames()) == 2 {
		for _, tag := range tags {
			sub := keepRegex.FindStringSubmatch(tag)
			if len(sub) != 2 {
				continue
			}

			number, err := strconv.Atoi(sub[1])
			if err != nil {
				continue
			}

			flavor["default"] = append(flavor["default"], tagFlavor{name: tag, number: number})
		}
	} else {
		// default layout group 1 flavor group 2 buildnr
		flavorID := 1
		buildNrID := 2

		// if group names are set use them else use default layout
		for i, name := range keepRegex.SubexpNames() {
			if name == "flavor" {
				flavorID = i
			} else if name == "buildnr" {
				buildNrID = i
			}
		}

		for _, tag := range tags {
			sub := keepRegex.FindStringSubmatch(tag)
			if len(sub) <= flavorID && len(sub) <= buildNrID {
				continue
			}

			number, err := strconv.Atoi(sub[buildNrID])
			if err != nil {
				continue
			}

			flavor[sub[flavorID]] = append(flavor[sub[flavorID]], tagFlavor{name: tag, number: number})
		}
	}

	for i := range flavor {
		sort.Sort(flavor[i])
	}

	return flavor
}

func getExpiredBuildTags(number int, keepRegex *regexp.Regexp, tags tagFlavors) []string {
	ret := make([]string, 0)

	if number < 0 || len(tags) < number {
		return ret
	}

	for _, tag := range tags[number:] {
		ret = append(ret, tag.name)
	}
	return ret
}

func getDigestForTags(repo string, tags []string) []string {
	digestMap := make([]string, 0)
	for _, tag := range tags {
		digest, err := hub.ManifestDigest(repo, tag)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		digestMap = append(digestMap, digest.String())
	}
	return digestMap
}

func getSaveTagsToRemove(repo string, candidatesToRemove, digestToSave []string) ([]string, []digest.Digest) {
	tagsToRemove := make([]string, 0)
	digestToRemove := make([]digest.Digest, 0)

	if !sort.StringsAreSorted(digestToSave) {
		sort.Strings(digestToSave)
	}

	var wg sync.WaitGroup
	var mutex = &sync.Mutex{}

	for _, tag := range candidatesToRemove {
		wg.Add(1)
		downloads <- true
		go func(repo, tag string) {
			digest, err := hub.ManifestDigest(repo, tag)
			<-downloads
			if err != nil {
				log.Fatalf("ERROR: %s", err)
			}
			if !contains(digestToSave, digest.String()) {
				mutex.Lock()
				tagsToRemove = append(tagsToRemove, tag)
				digestToRemove = append(digestToRemove, digest)
				mutex.Unlock()
			}
			wg.Done()
		}(repo, tag)
	}
	wg.Wait()

	return tagsToRemove, digestToRemove
}

func getInvalidTags(valid []*regexp.Regexp, tags []string) []string {
	invalidTags := make([]string, 0)
	for _, tag := range tags {
		if !validTag(valid, tag) {
			invalidTags = append(invalidTags, tag)
		}
	}
	return invalidTags
}

func validTag(valid []*regexp.Regexp, tag string) bool {
	for _, regex := range valid {
		if regex.FindString(tag) != "" {
			return true
		}
	}
	return false
}

func main() {
	var err error

	if !*insecure {
		hub, err = registry.New(cfg.Host, cfg.User, cfg.Password, registry.Quiet)
	} else {
		hub, err = registry.NewInsecure(cfg.Host, cfg.User, cfg.Password, registry.Quiet)
	}

	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	var wg sync.WaitGroup

	for _, repo := range rules.Repositories {
		pool <- true
		wg.Add(1)
		go work(repo, &wg, pool)
	}

	wg.Wait()
}

func work(repo string, wg *sync.WaitGroup, pool chan bool) {
	defer wg.Done()
	defer func() { <-pool }()

	tags, err := hub.Tags(repo)
	if err != nil || len(tags) == 0 {
		fmt.Println(err)
		return
	}

	invalidTags := getInvalidTags(rules.ValidTagsRegex, tags)

	flavorTags := getFlavor(rules.SortAndFilterRegex, tags)
	expiredBuildTags := make([]string, 0)
	for _, ftags := range flavorTags {
		expiredBuildTags = append(expiredBuildTags, getExpiredBuildTags(rules.KeepNewestBySort, rules.SortAndFilterRegex, ftags)...)
	}
	removeCandidate := append(invalidTags, expiredBuildTags...)

	tagsToRemove := parallelFilter(removeCandidate, oldTags(rules.MinAge, repo))
	digestToSave := getDigestForTags(repo, notIn(tags, tagsToRemove))

	tagsSaveToRemove, digestSaveToRemove := getSaveTagsToRemove(repo, tagsToRemove, digestToSave)

	fmt.Println(repo, "Tags that will be removed: ", tagsSaveToRemove)
	if !*dryRun {
		for i := range digestSaveToRemove {
			removeImage(repo, digestSaveToRemove[i])
		}
	}
}
