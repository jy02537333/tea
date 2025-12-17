package service

import "sync"

var (
	depsOnce   sync.Once
	globalDeps SummaryDeps
)

// SetSummaryDeps initializes global dependencies once.
func SetSummaryDeps(d SummaryDeps) {
	depsOnce.Do(func() { globalDeps = d })
}

// NewUserSummaryServiceFromGlobal returns a service using globally set deps.
func NewUserSummaryServiceFromGlobal() UserSummaryService {
	return NewUserSummaryServiceWithDeps(globalDeps)
}
