package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// stdinIsInteractive reports whether stdin is a terminal. When skills-check is
// run with no subcommand in a real shell we launch the guided wizard; when
// stdin is piped or redirected (CI, scripts, `skills-check | cat`) we must NOT
// block on a prompt, so the caller falls back to printing help instead.
func stdinIsInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// runWizard renders the interactive "what do you want to do?" menu and prints a
// tailored next step. It is deliberately advisory (it hands you the exact
// command) except for "scan now", which runs the gate immediately. Split out
// from the cobra wiring and parameterised on in/out so it is unit-testable.
// The returned showHelp is true when the user asked to see every command, so
// the caller can render full cobra help.
func runWizard(in io.Reader, out io.Writer) (showHelp bool, err error) {
	r := bufio.NewReader(in)

	fmt.Fprintln(out, "🛡  SecureVibe — let's get you set up.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "What do you want to do?")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  1) Add the 29 security skills to my AI assistant  (catch issues as code is written)")
	fmt.Fprintln(out, "  2) Scan a file or folder right now                (secrets, typosquats, vulnerable deps)")
	fmt.Fprintln(out, "  3) Set up the gate in CI / pre-commit             (fail the build on findings)")
	fmt.Fprintln(out, "  4) See every command                              (full help)")
	fmt.Fprintln(out)
	fmt.Fprint(out, "Enter 1-4 (or q to quit): ")

	line, _ := r.ReadString('\n')
	choice := strings.TrimSpace(line)

	fmt.Fprintln(out)
	switch choice {
	case "1":
		fmt.Fprintln(out, "Add the skills to your AI assistant (no install needed):")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "  Claude Code:  npx @shieldnet360/secure-code-skill init")
		fmt.Fprintln(out, "  Live scanning over MCP (any agent):")
		fmt.Fprintln(out, "                claude mcp add SecureVibe -- npx -y @shieldnet360/secure-code-mcp")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "  Cursor / Copilot / Codex / Windsurf / Cline / Devin: copy one file from dist/")
		fmt.Fprintln(out, "  → see 'Embed in your IDE' in the README.")
		return false, nil
	case "2":
		fmt.Fprint(out, "Path to scan (file or folder) [.]: ")
		pLine, _ := r.ReadString('\n')
		path := strings.TrimSpace(pLine)
		if path == "" {
			path = "."
		}
		fmt.Fprintf(out, "\n→ skills-check gate %s --severity-floor high\n\n", path)
		gc := policyCheckCmd()
		gc.SetOut(out)
		gc.SetErr(out)
		gc.SetArgs([]string{path, "--severity-floor", "high"})
		// A failing gate returns a non-nil error (exit 1) — that's a
		// successful demonstration, not a wizard failure, so swallow it
		// after the findings have been printed.
		if err := gc.Execute(); err != nil {
			fmt.Fprintf(out, "\n%s\n", err)
		}
		return false, nil
	case "3":
		fmt.Fprintln(out, "Fail the build on secrets / bad deps. Two ways:")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "  pre-commit  → add the 'secure-code-gate' hook to .pre-commit-config.yaml")
		fmt.Fprintln(out, "  GitHub CI   → copy examples/ci/securevibe.yml to .github/workflows/")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "  Try it offline first:  make demo")
		fmt.Fprintln(out, "  → see 'Gate in pre-commit and CI' in the README.")
		return false, nil
	case "4":
		return true, nil // caller prints full help
	case "q", "Q", "":
		fmt.Fprintln(out, "Nothing to do — run 'skills-check --help' any time.")
		return false, nil
	default:
		fmt.Fprintf(out, "Didn't recognise %q. Run 'skills-check --help' to see every command.\n", choice)
		return false, nil
	}
}

// maybeRunWizard is the root command's RunE. With no subcommand it launches the
// guided wizard on an interactive terminal, and otherwise prints help (so CI
// and piped invocations behave exactly as before).
func maybeRunWizard(c *cobra.Command, _ []string) error {
	if !stdinIsInteractive() {
		return c.Help()
	}
	showHelp, err := runWizard(os.Stdin, c.OutOrStdout())
	if err != nil {
		return err
	}
	if showHelp {
		return c.Help()
	}
	return nil
}
