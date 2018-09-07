package main

import "github.com/davecgh/go-spew/spew"

func spewInit() {
	spew.Config.DisablePointerAddresses = true
	spew.Config.DisablePointerMethods = true
	spew.Config.DisableCapacities = true
	spew.Config.DisableMethods = true
}
