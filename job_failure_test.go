package main

import "testing"

func TestRunScheduledJobReportsTheExampleFailure(t *testing.T) {
	if err := runScheduledJob(); err == nil {
		t.Fatal("expected the sample job to produce an error")
	}
}
