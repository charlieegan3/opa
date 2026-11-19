// Copyright 2026 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/storage"
	inmem "github.com/open-policy-agent/opa/v1/storage/inmem/test"
	"github.com/open-policy-agent/opa/v1/test/cases"
)

var regoBuiltinModes = []struct {
	name    string
	enabled bool
}{
	{"go", false},
	{"rego", true},
}

// The YAML cases check results. This checks what they cannot: the trace of a
// Rego built-in must be the same as the trace of the Go one.
func TestRegoBuiltinsTrace(t *testing.T) {
	const policy = `package x

p if startswith("foobar", "foo")
`
	traces := map[bool][]string{}
	for _, mode := range regoBuiltinModes {
		c, err := compileWithRegoBuiltins(map[string]string{"x.rego": policy}, ast.ParserOptions{RegoVersion: ast.RegoV1}, mode.enabled)
		if err != nil {
			t.Fatal(err)
		}
		if got := c.RegoBuiltinRef("startswith") != nil; got != mode.enabled {
			t.Fatalf("%s: rego implementation in use: %v, want %v", mode.name, got, mode.enabled)
		}

		ctx := t.Context()
		store := inmem.New()
		txn := storage.NewTransactionOrDie(ctx, store)
		defer store.Abort(ctx, txn)

		buf := NewBufferTracer()
		rs, err := NewQuery(ast.MustParseBody("data.x.p = x")).
			WithStore(store).WithTransaction(txn).WithCompiler(c).WithQueryCompiler(c.QueryCompiler()).
			WithTracer(buf).Run(ctx)
		if err != nil || len(rs) != 1 {
			t.Fatalf("%s: unexpected result %v, error %v", mode.name, rs, err)
		}
		for _, ev := range *buf {
			traces[mode.enabled] = append(traces[mode.enabled], fmt.Sprintf("%v %v", ev.Op, ev.Node))
		}
	}
	if !slices.Equal(traces[false], traces[true]) {
		t.Fatalf("traces differ.\ngo:\n%s\nrego:\n%s", strings.Join(traces[false], "\n"), strings.Join(traces[true], "\n"))
	}
}

// compileWithRegoBuiltins compiles modules with the Rego built-ins mode set.
func compileWithRegoBuiltins(modules map[string]string, po ast.ParserOptions, enabled bool) (*ast.Compiler, error) {
	parsed := make(map[string]*ast.Module, len(modules))
	for name, src := range modules {
		mod, err := ast.ParseModuleWithOpts(name, src, po)
		if err != nil {
			return nil, err
		}
		parsed[name] = mod
	}
	c := ast.NewCompiler().
		WithDefaultRegoVersion(po.RegoVersion).
		WithCapabilities(po.Capabilities).
		WithRegoBuiltins(enabled)
	if c.Compile(parsed); c.Failed() {
		return nil, c.Errors
	}
	return c, nil
}

// regoBuiltinCases finds the YAML conformance cases that use a built-in with
// a Rego implementation, grouped by built-in. A case counts when the compiler
// adds a Rego implementation for it. So there is no list to keep up to date:
// a new file in ast/regobuiltins is picked up, and built-ins without one are
// left out.
func regoBuiltinCases(tb testing.TB) map[string][]cases.TestCase {
	tb.Helper()

	found := map[string][]cases.TestCase{}
	for _, tc := range cases.MustLoad("../test/cases/testdata/v1").Sorted().Cases {
		if len(tc.Env) > 0 || tc.ExperimentalKeywords {
			continue
		}
		po := ast.ParserOptions{RegoVersion: ast.RegoV1}
		c, err := compileWithRegoBuiltins(caseModules(tc), po, true)
		if err != nil {
			// A case that cannot compile at all is not ours. A case that only
			// fails with the mode on is a bug, so it must not be dropped.
			if _, goErr := compileWithRegoBuiltins(caseModules(tc), po, false); goErr == nil {
				tb.Errorf("%s: compiles with the Go built-ins but not the Rego ones: %v", tc.Note, err)
			}
			continue
		}
		for name := range c.Modules {
			if builtin, ok := strings.CutPrefix(name, "regobuiltins/"); ok {
				builtin = strings.TrimSuffix(builtin, ".rego")
				found[builtin] = append(found[builtin], tc)
			}
		}
	}
	return found
}

func caseModules(tc cases.TestCase) map[string]string {
	modules := make(map[string]string, len(tc.Modules))
	for i, module := range tc.Modules {
		modules[fmt.Sprintf("test-%d.rego", i)] = module
	}
	return modules
}

// TestRegoBuiltinsConformance runs the YAML cases that use a Rego built-in
// with the mode on, whatever the environment says.
func TestRegoBuiltinsConformance(t *testing.T) {
	t.Parallel()

	found := regoBuiltinCases(t)
	if len(found) == 0 {
		t.Fatal("no conformance cases use a Rego built-in")
	}
	for builtin, tcs := range found {
		for _, tc := range tcs {
			t.Run(builtin+"/"+tc.Note, func(t *testing.T) {
				t.Parallel()

				testRun(t, tc, regoVersion(ast.RegoV1), regoBuiltinsMode(true))
			})
		}
	}
}

// BenchmarkRegoBuiltins compares the Go and Rego implementations of each
// built-in, using the same YAML cases as the conformance tests. One op runs
// every case for the built-in once.
func BenchmarkRegoBuiltins(b *testing.B) {
	ctx := context.Background()

	for builtin, tcs := range regoBuiltinCases(b) {
		b.Run(builtin, func(b *testing.B) {
			for _, mode := range regoBuiltinModes {
				b.Run(mode.name, func(b *testing.B) {
					queries := make([]*Query, 0, len(tcs))
					for _, tc := range tcs {
						c, err := compileWithRegoBuiltins(caseModules(tc), ast.ParserOptions{RegoVersion: ast.RegoV1}, mode.enabled)
						if err != nil {
							b.Fatal(err)
						}
						query, err := c.QueryCompiler().Compile(ast.MustParseBody(tc.Query))
						if err != nil {
							b.Fatal(err)
						}
						store := inmem.New()
						if tc.Data != nil {
							store = inmem.NewFromObject(*tc.Data)
						}
						txn := storage.NewTransactionOrDie(ctx, store)
						defer store.Abort(ctx, txn)

						q := NewQuery(query).WithCompiler(c).WithStore(store).WithTransaction(txn)
						if tc.InputTerm != nil {
							q = q.WithInput(ast.MustParseTerm(*tc.InputTerm))
						} else if tc.Input != nil {
							q = q.WithInput(ast.NewTerm(ast.MustInterfaceToValue(*tc.Input)))
						}
						queries = append(queries, q)
					}

					b.ReportAllocs()
					for b.Loop() {
						for _, q := range queries {
							_, _ = q.Run(ctx) // cases that expect an error still cost the same in both modes
						}
					}
				})
			}
		})
	}
}
