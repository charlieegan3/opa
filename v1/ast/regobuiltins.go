// Copyright 2026 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"embed"
	"os"
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/v1/util"
)

// Rego built-ins are built-in functions that are written in Rego.
//
// Each file in regobuiltins/ is named after a built-in, for example
// "trim_prefix.rego". It has the package "system.regobuiltins.<name>" and one
// function called "impl". The built-in keeps its Go declaration and its Go
// implementation. When the experimental mode is on, the compiler adds the
// Rego module to the policy, and the evaluator calls "impl" in place of the Go
// function. Turning the mode off goes back to the Go function.
//
// The mode is off unless the OPA_EXPERIMENTAL_REGO_BUILTINS environment
// variable is set, or Compiler.WithRegoBuiltins is used.
//
//go:embed regobuiltins/*.rego
var regoBuiltinFS embed.FS

const regoBuiltinDir = "regobuiltins"

var regoBuiltinsEnvEnabled = sync.OnceValue(func() bool {
	return os.Getenv("OPA_EXPERIMENTAL_REGO_BUILTINS") != ""
})

// regoBuiltinModules parses the embedded modules once. The modules are only
// read after this, the compiler copies them before use.
var regoBuiltinModules = sync.OnceValue(func() map[string]*Module {
	entries, err := regoBuiltinFS.ReadDir(regoBuiltinDir)
	if err != nil {
		panic(err)
	}
	mods := make(map[string]*Module, len(entries))
	for _, entry := range entries {
		path := regoBuiltinDir + "/" + entry.Name()
		src, err := regoBuiltinFS.ReadFile(path)
		if err != nil {
			panic(err)
		}
		mods[strings.TrimSuffix(entry.Name(), ".rego")] = MustParseModuleWithOpts(string(src), ParserOptions{RegoVersion: RegoV1})
	}
	return mods
})

func regoBuiltinRef(name string) Ref {
	ref := DefaultRootRef.Append(NewTerm(SystemDocumentKey)).Append(StringTerm(regoBuiltinDir))
	for part := range strings.SplitSeq(name, ".") {
		ref = ref.Append(StringTerm(part))
	}
	return ref.Append(StringTerm("impl"))
}

// WithRegoBuiltins turns the experimental Rego built-ins on or off for this
// compiler. When on, built-ins that have a Rego implementation are evaluated
// using it, instead of the Go implementation.
func (c *Compiler) WithRegoBuiltins(enabled bool) *Compiler {
	c.regoBuiltins = enabled
	return c
}

// RegoBuiltinRef returns the reference of the Rego function that implements
// the built-in, or nil if the built-in is not implemented in Rego for this
// compiler. The evaluator uses this to route calls.
func (c *Compiler) RegoBuiltinRef(name string) Ref {
	return c.regoBuiltinRefs[name]
}

// injectRegoBuiltins adds the Rego implementation of every built-in the
// policy calls. Only the used ones are added, so policies that call none
// pay nothing.
func (c *Compiler) injectRegoBuiltins() {
	if !c.regoBuiltins {
		return
	}

	available := regoBuiltinModules()
	c.regoBuiltinRefs = map[string]Ref{}

	// An implementation can call other built-ins that are also in Rego, so
	// scan what was added until nothing new is found.
	pending := make([]*Module, 0, len(c.Modules))
	for _, mod := range c.Modules {
		pending = append(pending, mod)
	}
	for len(pending) > 0 {
		var added []*Module
		for _, mod := range pending {
			WalkRefs(mod, func(ref Ref) bool {
				if len(ref) == 0 {
					return false
				}
				name := ref.String()
				src, ok := available[name]
				if !ok || c.regoBuiltinRefs[name] != nil {
					return false
				}
				if _, ok := c.builtins[name]; !ok {
					return false // not allowed by the capabilities
				}
				cpy := src.Copy()
				c.Modules[regoBuiltinDir+"/"+name+".rego"] = cpy
				c.regoBuiltinRefs[name] = regoBuiltinRef(name)
				added = append(added, cpy)
				return false
			})
		}
		pending = added
	}
	c.sorted = util.KeysSorted(c.Modules)
}
