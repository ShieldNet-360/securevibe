package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunWizardSkillsChoice(t *testing.T) {
	var out bytes.Buffer
	showHelp, err := runWizard(strings.NewReader("1\n"), &out)
	if err != nil {
		t.Fatalf("runWizard: %v", err)
	}
	if showHelp {
		t.Fatalf("choice 1 should not request help")
	}
	if !strings.Contains(out.String(), "secure-code-skill init") {
		t.Fatalf("expected skills install guidance, got:\n%s", out.String())
	}
}

func TestRunWizardSeeEveryCommandRequestsHelp(t *testing.T) {
	var out bytes.Buffer
	showHelp, err := runWizard(strings.NewReader("4\n"), &out)
	if err != nil {
		t.Fatalf("runWizard: %v", err)
	}
	if !showHelp {
		t.Fatalf("choice 4 should request full help")
	}
}

func TestRunWizardQuit(t *testing.T) {
	var out bytes.Buffer
	showHelp, err := runWizard(strings.NewReader("q\n"), &out)
	if err != nil {
		t.Fatalf("runWizard: %v", err)
	}
	if showHelp {
		t.Fatalf("quit should not request help")
	}
	if !strings.Contains(out.String(), "--help") {
		t.Fatalf("expected a help hint on quit, got:\n%s", out.String())
	}
}

func TestRunWizardUnrecognised(t *testing.T) {
	var out bytes.Buffer
	if _, err := runWizard(strings.NewReader("banana\n"), &out); err != nil {
		t.Fatalf("runWizard: %v", err)
	}
	if !strings.Contains(out.String(), "Didn't recognise") {
		t.Fatalf("expected unrecognised-input message, got:\n%s", out.String())
	}
}
