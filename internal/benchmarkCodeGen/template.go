package main

const (
	benchmarksFileTemplate string = `
{{ define "header" }}
// This file had been automatically generated by utility "git.dx.center/trafficstars/testJob0/internal/benchmarkCodeGen"

package {{ .PackageName }}

import (
	"testing"

	benchmark "git.dx.center/trafficstars/testJob0/internal/benchmarkRoutines"
)
{{ end }}

{{ define "benchmarkFunction" }}
func Benchmark{{ .Action }}{{ if .KeyIsString }}String{{ else }}Int{{ end }}_blockSize{{ .BlockSize }}_keyAmount{{ .KeyAmount }}(b *testing.B) {
	benchmark.DoBenchmarkOf{{ .Action }}(b, NewHashMap, {{ .BlockSize }}, {{ .KeyAmount }}, {{ .KeyIsString }})
}
{{ end }}
{{ define "testFunction" }}
func TestMap(t *testing.T) {
	benchmark.DoTest(t, NewHashMap)
}
{{ end }}
`
)
