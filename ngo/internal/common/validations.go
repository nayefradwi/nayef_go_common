package common

import "regexp"

var NameRegex = regexp.MustCompile(`^[a-z][a-z0-9]*$`)
