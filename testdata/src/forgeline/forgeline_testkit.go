package forgeline

import (
	shard "lib" // want `import "lib" is spelled shard here but lib elsewhere`
)

// Kit sits at the matcher's boundary. Its file's name CONTAINS "_test" and does
// not END in "_test.go", so it is ordinary source, bound by the agreement like
// any other. A matcher widened from a suffix to a substring would excuse it.
func Kit() string { return shard.Name }
