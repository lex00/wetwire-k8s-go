package domain

import (
	coredomain "github.com/lex00/wetwire-core-go/domain"
)

// Compile-time interface checks
var (
	_ coredomain.Domain        = (*K8sDomain)(nil)
	_ coredomain.ListerDomain  = (*K8sDomain)(nil)
	_ coredomain.GrapherDomain = (*K8sDomain)(nil)
)
