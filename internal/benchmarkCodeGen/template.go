package main

const (
	benchmarksFileTemplate string = `{{ define "header" }}// This file had been automatically generated by utility "github.com/xaionaro-go/atomicmap/internal/benchmarkCodeGen"

package {{ .PackageName }}

import (
	"testing"

	benchmark "github.com/xaionaro-go/atomicmap/internal/benchmarkRoutines"
)
{{ end }}

{{ define "benchmarkFunction" }}
func Benchmark_{{ .PackageName }}_{{ .Action }}_{{ .KeyType }}KeyType_blockSize{{ .BlockSize }}_keyAmount{{ .KeyAmount }}_{{ .ThreadSafety }}ThreadSafety(b *testing.B) {
	{{if and (not .ThreadSafety) (eq .PackageName "openAddressGrowingMap")}}threadSafe = false; {{end}}benchmark.DoBenchmarkOf{{ .Action }}(b, newWithArgsIface, {{ .BlockSize }}, {{ .KeyAmount }}, "{{ .KeyType }}"){{if and (not .ThreadSafety) (eq .PackageName "openAddressGrowingMap")}}; threadSafe = true{{end}}
}
{{ if .ThreadSafety }}
{{ if ne .Action "Unset" }}
func BenchmarkParallel_{{ .PackageName }}_{{ .Action }}_{{ .KeyType }}KeyType_blockSize{{ .BlockSize }}_keyAmount{{ .KeyAmount }}_{{ .ThreadSafety }}ThreadSafety(b *testing.B) {
	{{if and (not .ThreadSafety) (eq .PackageName "openAddressGrowingMap")}}threadSafe = false; {{end}}benchmark.DoParallelBenchmarkOf{{ .Action }}(b, newWithArgsIface, {{ .BlockSize }}, {{ .KeyAmount }}, "{{ .KeyType }}"){{if and (not .ThreadSafety) (eq .PackageName "openAddressGrowingMap")}}; threadSafe = true{{end}}
}
{{ end }}
{{ end }}
{{ end }}
{{ define "testFunction" }}
func TestMap(t *testing.T) {
	benchmark.DoTest(t, newWithArgsIface)
}
{{ end }}
{{ define "testCollisionsFunction" }}
func TestMapCollisions(t *testing.T) {
	benchmark.DoTestCollisions(t, newWithArgsIface)
}
{{ end }}
{{ define "testConcurrencyFunction" }}
func TestMapConcurrency(t *testing.T) {
	benchmark.DoTestConcurrency(t, newWithArgsIface)
}
{{ end }}
`
)
