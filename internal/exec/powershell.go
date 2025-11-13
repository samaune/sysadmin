package exec
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// escapeForSingleQuotedPS escapes single quotes for single-quoted PowerShell strings
func escapeForSingleQuotedPS(s string) string {
	// in PowerShell single-quoted strings, single quote is escaped by doubling it
	return strings.ReplaceAll(s, "'", "''")
}

// runCmdElevatedCapture runs the given executable (with args) elevated (UAC prompt).
// It returns the combined stdout+stderr output produced by the elevated process.
func runCmdElevatedCapture(execPath string, args ...string) ([]byte, error) {
	// create temp dir
	tmpDir, err := os.MkdirTemp("", "elevate-run-*")
	if err != nil {
		return nil, fmt.Errorf("tempdir: %w", err)
	}
	// cleanup when done
	defer os.RemoveAll(tmpDir)

	outFile := filepath.Join(tmpDir, "out.txt")
	scriptFile := filepath.Join(tmpDir, "run.ps1")

	// Build the PowerShell command line that will run the target and redirect all streams
	// We will run: & 'execPath' arg1 arg2 ... *> 'outFile'
	// but use single-quoting for paths and escape single quotes inside them.
	var parts []string
	parts = append(parts, fmt.Sprintf("& '%s'", escapeForSingleQuotedPS(execPath)))
	for _, a := range args {
		// Quote each arg in single quotes (escaping any single quotes inside)
		parts = append(parts, fmt.Sprintf("'%s'", escapeForSingleQuotedPS(a)))
	}
	// `*>` redirects all streams to file in PowerShell (works in PS 3+). To be conservative, use "2>&1 | Out-File".
	// We'll redirect streams with `2>&1 | Out-File -FilePath '<outFile>' -Encoding utf8`
	// Combine to full command:
	cmdLine := strings.Join(parts, " ") + " 2>&1 | Out-File -FilePath '" + escapeForSingleQuotedPS(outFile) + "' -Encoding utf8"

	// Write the script
	if err := os.WriteFile(scriptFile, []byte(cmdLine), 0o600); err != nil {
		return nil, fmt.Errorf("write script: %w", err)
	}

	// Build the PowerShell command that will elevate and run the script, waiting for it to finish.
	// We call: powershell -NoProfile -Command "Start-Process -FilePath 'powershell' -ArgumentList '-NoProfile','-ExecutionPolicy','Bypass','-File','<scriptFile>' -Verb RunAs -Wait"
	// Use single quotes in the -Command and escape single quotes inside paths.
	psCmd := fmt.Sprintf(
		"Start-Process -FilePath 'powershell' -ArgumentList '-NoProfile','-ExecutionPolicy','Bypass','-File','%s' -Verb RunAs -WindowStyle Hidden -Wait",
		escapeForSingleQuotedPS(scriptFile),
	)

	// Execute the local (non-elevated) powershell which will invoke Start-Process (UAC prompt)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd)

	// We don't capture stdout/stderr here because Start-Process will spawn a separate elevated process.
	// But we may capture any immediate errors from Start-Process itself.
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		// Return combined error output for debugging
		return append(buf.Bytes(), []byte("\n"+err.Error())...), fmt.Errorf("start-elevated failed: %w", err)
	}

	// At this point the elevated process should have finished and written outFile. Read it.
	outBytes, err := os.ReadFile(outFile)
	if err != nil {
		// if the elevated process failed before creating output, return whatever we captured
		return append(buf.Bytes(), []byte("\n(no output file) "+err.Error())...), fmt.Errorf("read output failed: %w", err)
	}

	// Return the elevated process output
	return outBytes, nil
}

// Example wrapper for running a PowerShell command elevated and getting its output
func RunPwsh(command string) ([]byte, error) {
	// We will call pwsh.exe -Command "<command>"
	// If you want Windows PowerShell 5.1, use "powershell" instead of pwsh path.
	pwshPath := `C:\Program Files\PowerShell\7\pwsh.exe`
	return runCmdElevatedCapture(pwshPath, "-NoProfile", "-Command", command)
}