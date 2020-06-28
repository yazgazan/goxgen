// +build gox

package views

// Definitions for functions goxgen's typechecker cannot resolve.
// The function bodies can be omitted since the file won't be compiled.

import (
	gox "github.com/yazgazan/goxgen/src"
)

func T(body gox.ComponentOrHTML) gox.ComponentOrHTML
