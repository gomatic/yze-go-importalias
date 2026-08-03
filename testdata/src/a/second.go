package a

import (
	alias "lib" // want `import "lib" is spelled alias here but lib elsewhere`
)

// Second reads the same package under a different name — the divergence.
func Second() string { return alias.Name }
