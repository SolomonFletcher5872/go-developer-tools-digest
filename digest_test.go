package main

import "testing"

func TestDigestDecision(t *testing.T) {
	tests := []struct {
		name string
		in   []Event
		want string
	}{
		{"empty", nil, "Developer tools digest\nNo events in this window."},
		{"release and diagnostic", []Event{{"release", "api", "v2.1"}, {"diagnostic", "worker", "latency high"}}, "Developer tools digest (2 events)\n- release: api/v2.1\n- diagnostic: worker/latency high\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Digest(tt.in); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
