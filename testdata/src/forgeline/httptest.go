package forgeline

import (
	shim "lib" // want `import "lib" is spelled shim here but lib elsewhere`
)

// Helper sits at the matcher's LEFT edge — the underscore that separates a base
// name from the "_test.go" suffix. Its file's name ends in "test.go" and not in
// "_test.go", so it is ordinary source, bound by the agreement like any other
// and counted toward it like any other.
//
// This edge is not hypothetical and not latent: `find ~/src/github.com -name
// '*test.go' -not -name '*_test.go'` returns 39 files, among them
// net/http/httptest/httptest.go and gomatic/go-wofl/internal/pgtest/pgtest.go.
// A matcher that dropped the underscore would excuse every one of them from the
// agreement AND withdraw their votes from it.
func Helper() string { return shim.Name }
