package api

// SkippedResult returns a CheckResult indicating the check was skipped.
func SkippedResult(name string, summary string, err error) *CheckResult {
	details := make(map[string]interface{})
	if err != nil {
		details["error"] = err
	}
	return &CheckResult{
		Name:    name,
		State:   StateSkipped,
		Summary: summary,
		Details: details,
	}
}
