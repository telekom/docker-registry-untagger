// docker-unregstriy-untagger :- tests for main module
// Copyright (c) 2017, Steffen Windoffer, Deutsche Telekom AG
// Contact: opensource@telekom.de
// This file is distributed under the conditions of the Apache2 license.
// For details see the files LICENSE at the toplevel.

package main

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFlavor(t *testing.T) {
	var tests = []struct {
		inRegex *regexp.Regexp
		inTags  []string
		out     map[string]tagFlavors
	}{
		{
			regexp.MustCompile("([A-Za-z]+)_([0-9]+)"),
			[]string{"bird_5", "blue_5", "bird_7", "blue_6", "bird_6", "release_5", "blue_7"},
			map[string]tagFlavors{
				"bird": tagFlavors{
					{name: "bird_7", number: 7},
					{name: "bird_6", number: 6},
					{name: "bird_5", number: 5},
				},
				"blue": tagFlavors{
					{name: "blue_7", number: 7},
					{name: "blue_6", number: 6},
					{name: "blue_5", number: 5},
				},
				"release": tagFlavors{
					{name: "release_5", number: 5},
				},
			},
		}, {
			regexp.MustCompile("([?P<flavor>A-Za-z]+)_([?P<buildnr>0-9]+)"),
			[]string{"bird_5", "blue_5", "bird_7", "blue_6", "bird_6", "release_5", "blue_7"},
			map[string]tagFlavors{
				"bird": tagFlavors{
					{name: "bird_7", number: 7},
					{name: "bird_6", number: 6},
					{name: "bird_5", number: 5},
				},
				"blue": tagFlavors{
					{name: "blue_7", number: 7},
					{name: "blue_6", number: 6},
					{name: "blue_5", number: 5},
				},
				"release": tagFlavors{
					{name: "release_5", number: 5},
				},
			},
		}, {
			regexp.MustCompile("(?P<buildnr>[0-9]+)_(?P<flavor>[A-Za-z]+)"),
			[]string{"5_bird", "5_blue", "7_bird", "6_blue", "6_bird", "5_release", "7_blue"},
			map[string]tagFlavors{
				"bird": tagFlavors{
					{name: "7_bird", number: 7},
					{name: "6_bird", number: 6},
					{name: "5_bird", number: 5},
				},
				"blue": tagFlavors{
					{name: "7_blue", number: 7},
					{name: "6_blue", number: 6},
					{name: "5_blue", number: 5},
				},
				"release": tagFlavors{
					{name: "5_release", number: 5},
				},
			},
		}, {
			regexp.MustCompile("([A-Za-z]+)_([0-9]+)"),
			[]string{"bird_a", "blue_5", "bird_7", "blue_6", "bird_6", "release_5", "blue_7"},
			map[string]tagFlavors{
				"bird": tagFlavors{
					{name: "bird_7", number: 7},
					{name: "bird_6", number: 6},
				},
				"blue": tagFlavors{
					{name: "blue_7", number: 7},
					{name: "blue_6", number: 6},
					{name: "blue_5", number: 5},
				},
				"release": tagFlavors{
					{name: "release_5", number: 5},
				},
			},
		}, {
			regexp.MustCompile("([A-Za-z]+)_([a])"),
			[]string{"bird_a"},
			map[string]tagFlavors{},
		}, {
			regexp.MustCompile("(release)_([0-9+])"),
			[]string{"release_3", "release_2"},
			map[string]tagFlavors{
				"release": tagFlavors{
					{name: "release_3", number: 3},
					{name: "release_2", number: 2},
				},
			},
		}, {
			regexp.MustCompile("[A-Za-z]+_([0-9a]+)"),
			[]string{"blue_5", "blue_6", "blue_7", "blue_a", "blue_b"},
			map[string]tagFlavors{
				"default": tagFlavors{
					{name: "blue_7", number: 7},
					{name: "blue_6", number: 6},
					{name: "blue_5", number: 5},
				},
			},
		},
	}

	for i, tt := range tests {
		b := getFlavor(tt.inRegex, tt.inTags)
		assert.Equal(t, tt.out, b, "TestGetFlavor "+strconv.Itoa(i+1)+" values should be equal")
	}
}

func TestGetExpiredBuildTags(t *testing.T) {
	var tests = []struct {
		inNumber int
		inRegex  *regexp.Regexp
		inTag    tagFlavors
		out      []string
	}{
		{
			-1,
			regexp.MustCompile("release_[0-9]+"),
			tagFlavors{{name: "release_5", number: 5}},
			[]string{},
		}, {
			0,
			regexp.MustCompile("release_[0-9]+"),
			tagFlavors{{name: "release_7", number: 7}, {name: "release_6", number: 6}, {name: "release_5", number: 5}},
			[]string{"release_7", "release_6", "release_5"},
		}, {
			2,
			regexp.MustCompile("release_[0-9]+"),
			tagFlavors{{name: "release_7", number: 7}, {name: "release_6", number: 6}, {name: "release_5", number: 5}},
			[]string{"release_5"},
		}, {
			100,
			regexp.MustCompile("release_[0-9]+"),
			tagFlavors{{name: "release_7", number: 7}, {name: "release_6", number: 6}, {name: "release_5", number: 5}},
			[]string{},
		},
	}

	for i, tt := range tests {
		b := getExpiredBuildTags(tt.inNumber, tt.inRegex, tt.inTag)
		assert.Equal(t, tt.out, b, "TestGetExpiredBuildTags "+strconv.Itoa(i+1)+" values should be equal")
	}
}

func TestGetInvalidTags(t *testing.T) {
	var tests = []struct {
		inRegex []*regexp.Regexp
		inTag   []string
		out     []string
	}{
		{[]*regexp.Regexp{regexp.MustCompile("release_[0-9]+")}, []string{"release_5"}, []string{}},
		{[]*regexp.Regexp{regexp.MustCompile("release_[0-9]+")}, []string{"releae_5"}, []string{"releae_5"}},
		{[]*regexp.Regexp{regexp.MustCompile("release_[0-9]+")}, []string{"release_102"}, []string{}},
	}
	for i, tt := range tests {
		b := getInvalidTags(tt.inRegex, tt.inTag)
		assert.Equal(t, tt.out, b, "TestGetInvalidTags "+strconv.Itoa(i+1)+" they should be equal")
	}
}

func TestValidTag(t *testing.T) {
	var tests = []struct {
		inRegex []*regexp.Regexp
		inTag   string
		out     bool
	}{
		{[]*regexp.Regexp{regexp.MustCompile("release_[0-9]+")}, "release_5", true},
		{[]*regexp.Regexp{regexp.MustCompile("release_[0-9]+")}, "releae_5", false},
		{[]*regexp.Regexp{regexp.MustCompile("release_[0-9]+")}, "release_102", true},
	}
	for _, tt := range tests {
		b := validTag(tt.inRegex, tt.inTag)
		if b != tt.out {
			t.Errorf("validTag(%q, %q) => %t, want %t", tt.inRegex, tt.inTag, b, tt.out)
		}
	}
}

func TestLen(t *testing.T) {
	var tests = []struct {
		tf  tagFlavors
		out int
	}{
		{
			tagFlavors{{name: "release_7", number: 7}, {name: "release_6", number: 6}, {name: "release_5", number: 5}},
			3,
		},
		{
			tagFlavors{{name: "release_6", number: 6}, {name: "release_5", number: 5}},
			2,
		},
		{
			tagFlavors{},
			0,
		},
	}
	for _, tt := range tests {
		b := tt.tf.Len()
		if b != tt.out {
			t.Errorf("%q.Len() => %q, want %q", tt.tf, b, tt.out)
		}
	}
}

func TestSwap(t *testing.T) {
	var tests = []struct {
		tf    tagFlavors
		i, j  int
		tfout tagFlavors
	}{
		{
			tagFlavors{{name: "release_7", number: 7}, {name: "release_6", number: 6}, {name: "release_5", number: 5}},
			1, 2,
			tagFlavors{{name: "release_7", number: 7}, {name: "release_5", number: 5}, {name: "release_6", number: 6}},
		},
		{
			tagFlavors{{name: "release_6", number: 6}, {name: "release_5", number: 5}},
			0, 1,
			tagFlavors{{name: "release_5", number: 5}, {name: "release_6", number: 6}},
		},
	}
	for i, tt := range tests {
		tt.tf.Swap(tt.i, tt.j)
		assert.Equal(t, tt.tfout, tt.tf, "TestSwap "+strconv.Itoa(i+1)+" they should be equal")
	}
}
