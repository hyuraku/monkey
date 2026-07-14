package evaluator

import "monkey/object"

// builtins mirrors object.Builtins so that every builtin available to the
// compiler/VM path is also available to the tree-walking evaluator.
// Add new builtins in object/builtins.go only; both engines pick them up.
var builtins = map[string]*object.Builtin{}

func init() {
	for _, b := range object.Builtins {
		builtins[b.Name] = b.Builtin
	}
}
