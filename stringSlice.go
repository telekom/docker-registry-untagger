// docker-unregstriy-untagger :- string slice utils
// Copyright (c) 2017, Steffen Windoffer, Deutsche Telekom AG
// Contact: opensource@telekom.de
// This file is distributed under the conditions of the Apache2 license.
// For details see the files LICENSE at the toplevel.

package main

import (
	"sort"
	"sync"
)

func filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func parallelFilter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	var wg sync.WaitGroup
	var mutex = &sync.Mutex{}
	for _, v := range vs {
		wg.Add(1)
		go func(v string) {
			if f(v) {
				mutex.Lock()
				vsf = append(vsf, v)
				mutex.Unlock()
			}
			wg.Done()
		}(v)
	}
	wg.Wait()
	return vsf
}

// contains return true if item is in slice. slice needs to be sorted
func contains(slice []string, item string) bool {
	spot := sort.SearchStrings(slice, item)
	if len(slice) == spot || slice[spot] != item {
		return false
	}
	return true
}

// notIn returns all elements from a that are not in b
func notIn(a, b []string) []string {
	ret := make([]string, 0)
	if !sort.StringsAreSorted(b) {
		sort.Strings(b)
	}

	for _, ele := range a {
		if !contains(b, ele) {
			ret = append(ret, ele)
		}
	}
	return ret
}
