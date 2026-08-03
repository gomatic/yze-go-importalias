// Package lib collides by name with the other lib, so a file importing both
// must alias one of them.
package lib

// Other is a value the importers refer to.
const Other = "other"
