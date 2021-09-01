package api

// WarnResult returns a CheckResult when a warning occurs.
func WarnResult(name string, summary string) *CheckResult {
	return &CheckResult{
		Name:    name,
		State:   StateWarning,
		Summary: summary,
	}
}
