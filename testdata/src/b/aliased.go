package b

import (
	cli "versioned/cli/v3"
)

// Aliased spells the same binding explicitly.
func Aliased() string { return cli.Run }
