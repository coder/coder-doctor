package api

// PassResults returns a CheckResult when everything is OK.
func PassResult(name string, summary string) *CheckResult {
	return &CheckResult{
		Name:    name,
		State:   StatePassed,
		Summary: summary,
		Details: map[string]interface{}{},
	}
}
