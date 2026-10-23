package main

import "fmt"

type Event struct{ Kind, Service, Detail string }

func Digest(events []Event) string {
	if len(events) == 0 {
		return "Developer tools digest\nNo events in this window."
	}
	out := fmt.Sprintf("Developer tools digest (%d events)\n", len(events))
	for _, e := range events {
		out += fmt.Sprintf("- %s: %s/%s\n", e.Kind, e.Service, e.Detail)
	}
	return out
}
