// docker-unregstriy-untagger :- tests for string slice utils
// Copyright (c) 2017, Steffen Windoffer, Deutsche Telekom AG
// Contact: opensource@telekom.de
// This file is distributed under the conditions of the Apache2 license.
// For details see the files LICENSE at the toplevel.

package main

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	var tests = []struct {
		inSlice      []string
		inFilterFunc func(string) bool
		out          []string
	}{
		{
			[]string{"4444", "55555", "0"},
			func(s string) bool {
				if len(s) == 5 {
					return true
				}
				return false
			},
			[]string{"55555"},
		}, {
			[]string{"4444", "55555", "0"},
			func(s string) bool {
				if len(s) == 10 {
					return true
				}
				return false
			},
			[]string{},
		},
	}
	for i, tt := range tests {
		b := filter(tt.inSlice, tt.inFilterFunc)
		assert.Equal(t, tt.out, b, "test "+strconv.Itoa(i+1)+" they should be equal")
	}
}

func TestParallelFilter(t *testing.T) {
	var tests = []struct {
		inSlice      []string
		inFilterFunc func(string) bool
		out          []string
	}{
		{
			[]string{"4444", "55555", "0"},
			func(s string) bool {
				if len(s) == 5 {
					return true
				}
				return false
			},
			[]string{"55555"},
		}, {
			[]string{"4444", "55555", "0"},
			func(s string) bool {
				if len(s) == 10 {
					return true
				}
				return false
			},
			[]string{},
		},
	}
	for i, tt := range tests {
		b := parallelFilter(tt.inSlice, tt.inFilterFunc)
		assert.Equal(t, tt.out, b, "test "+strconv.Itoa(i+1)+" they should be equal")
	}
}

func TestContains(t *testing.T) {
	var tests = []struct {
		inSlice []string
		inItem  string
		out     bool
	}{
		{
			[]string{"9", "4444", "55555", "0"},
			"55555",
			true,
		}, {
			[]string{"4444", "55555", "0"},
			"9",
			false,
		},
	}
	for _, tt := range tests {
		b := contains(tt.inSlice, tt.inItem)
		if b != tt.out {
			t.Errorf("contains(%q, %q) => %t, want %t", tt.inSlice, tt.inItem, b, tt.out)
		}
	}
}

func TestNotIn(t *testing.T) {
	var tests = []struct {
		inA []string
		inB []string
		out []string
	}{
		{
			[]string{"9", "55555", "444334", "0"},
			[]string{"9", "4444", "55555", "0"},
			[]string{"444334"},
		}, {
			[]string{"4444", "55555", "0"},
			[]string{"9", "4444", "55555", "0"},
			[]string{},
		},
	}

	for i, tt := range tests {
		b := notIn(tt.inA, tt.inB)
		assert.Equal(t, tt.out, b, "TestNotIn "+strconv.Itoa(i+1)+" they should be equal")
	}
}
