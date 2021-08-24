package api

// SkippedResult returns a CheckResult indicating the check was skipped.
func SkippedResult(name string, summary string) *CheckResult {
	return &CheckResult{
		Name:    name,
		State:   StateSkipped,
		Summary: summary,
		Details: map[string]interface{}{},
	}
}
