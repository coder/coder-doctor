package api

// ErrorResult returns a CheckResult when an error occurs.
func ErrorResult(name string, summary string, err error) *CheckResult {
	return &CheckResult{
		Name:    name,
		State:   StateFailed,
		Summary: summary,
		Details: map[string]interface{}{
			"error": err,
		},
	}
}
