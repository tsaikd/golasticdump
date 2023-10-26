package main

import "github.com/tsaikd/KDGoLib/version"

// Development version of gogstash
const devVersion = "0.7.0"

func init() {
	if version.VERSION == "0.0.0" {
		version.VERSION = devVersion + "-dev"
	}
}
