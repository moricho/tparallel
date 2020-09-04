# tparallel
[![tparallel](https://github.com/moricho/tparallel/workflows/tparallel/badge.svg?branch=master)](https://github.com/moricho/tparallel/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/moricho/tparallel)](https://goreportcard.com/report/github.com/moricho/tparallel)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

`tparallel` finds inappropriate usage of `t.Parallel()` method in your Go test codes.  
It detects the following:  
- `t.Parallel()` is called for either a top-level test function or a sub-test function
- Although `t.Parallel()` is called in the sub-test function, it is post-processed by `defer` instead of `t.Cleanup()`
  
This tool was inspired by this blog: [Go言語でのテストの並列化 〜t.Parallel()メソッドを理解する〜](https://engineering.mercari.com/blog/entry/how_to_use_t_parallel/)

## Installation

```sh
$ go get github.com/moricho/tparallel/cmd/tparallel
```

## Usage

```sh
$ go vet -vettool=`which tparallel` <pkgname>
```

## Example

```go
package sample

import (
	"testing"
)

func Test_Table1(t *testing.T) {
	teardown := setup("Test_Table1")
	defer teardown()

	tests := []struct {
		name string
	}{
		{
			name: "Table1_Sub1",
		},
		{
			name: "Table1_Sub2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			call(tt.name)
		})
	}
}

func Test_Table2(t *testing.T) {
	teardown := setup("Test_Table2")
	t.Cleanup(teardown)
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "Table2_Sub1",
		},
		{
			name: "Table2_Sub2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call(tt.name)
		})
	}
}
```

```console
# github.com/moricho/tparallel/testdata/src/sample
testdata/src/sample/table_test.go:7:6: Test_Table1 should use t.Cleanup
testdata/src/sample/table_test.go:7:6: Test_Table1 should call t.Parallel on the top level
testdata/src/sample/table_test.go:30:6: Test_Table2's sub tests should call t.Parallel
```
