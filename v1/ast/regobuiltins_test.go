// Copyright 2026 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"strings"
	"testing"
)

func TestRegoBuiltinModules(t *testing.T) {
	for name, mod := range regoBuiltinModules() {
		if _, ok := BuiltinMap[name]; !ok {
			t.Errorf("%s: no Go declaration for this built-in", name)
		}
		if want := "data.system.regobuiltins." + name; mod.Package.Path.String() != want {
			t.Errorf("%s: package is %v, want %v", name, mod.Package.Path, want)
		}
		if len(mod.Rules) == 0 || !strings.HasSuffix(mod.Rules[0].Ref().String(), "impl") {
			t.Errorf("%s: expected a function called impl", name)
		}
	}
}

func TestCompilerOnlyInjectsUsedRegoBuiltins(t *testing.T) {
	c := NewCompiler().WithRegoBuiltins(true)
	c.Compile(map[string]*Module{"x.rego": MustParseModuleWithOpts("package x\np := trim_prefix(\"ab\", \"a\")\n", ParserOptions{RegoVersion: RegoV1})})
	if c.Failed() {
		t.Fatal(c.Errors)
	}
	if c.RegoBuiltinRef("trim_prefix") == nil {
		t.Fatal("trim_prefix should be injected")
	}
	if c.RegoBuiltinRef("trim_suffix") != nil {
		t.Fatal("trim_suffix is not used and should not be injected")
	}

	c = NewCompiler().WithRegoBuiltins(false)
	c.Compile(map[string]*Module{"x.rego": MustParseModuleWithOpts("package x\np := trim_prefix(\"ab\", \"a\")\n", ParserOptions{RegoVersion: RegoV1})})
	if c.RegoBuiltinRef("trim_prefix") != nil || len(c.Modules) != 1 {
		t.Fatal("nothing should be injected when the mode is off")
	}
}
