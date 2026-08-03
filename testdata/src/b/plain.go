// Package b imports a version-suffixed path plainly in one file and with an
// explicit alias equal to the package's own name in another. Both bind `cli`,
// so they agree and nothing is reported — reading the name from the path's last
// element instead of the type checker would call this a disagreement.
package b

import (
	"versioned/cli/v3"
)

// Plain reads the plainly-imported package.
func Plain() string { return cli.Run }
