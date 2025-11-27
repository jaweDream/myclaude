package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// Helper to reset test hooks
func resetTestHooks() {
	stdinReader = os.Stdin
	isTerminalFn = defaultIsTerminal
	codexCommand = "codex"
}

func TestParseArgs_NewMode(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *Config
		wantErr bool
	}{
		{
			name: "simple task",
			args: []string{"codex-wrapper", "analyze code"},
			want: &Config{
				Mode:          "new",
				Task:          "analyze code",
				WorkDir:       ".",
				ExplicitStdin: false,
			},
		},
		{
			name: "task with workdir",
			args: []string{"codex-wrapper", "analyze code", "/path/to/dir"},
			want: &Config{
				Mode:          "new",
				Task:          "analyze code",
				WorkDir:       "/path/to/dir",
				ExplicitStdin: false,
			},
		},
		{
			name: "explicit stdin mode",
			args: []string{"codex-wrapper", "-"},
			want: &Config{
				Mode:          "new",
				Task:          "-",
				WorkDir:       ".",
				ExplicitStdin: true,
			},
		},
		{
			name: "stdin with workdir",
			args: []string{"codex-wrapper", "-", "/some/dir"},
			want: &Config{
				Mode:          "new",
				Task:          "-",
				WorkDir:       "/some/dir",
				ExplicitStdin: true,
			},
		},
		{
			name:    "no args",
			args:    []string{"codex-wrapper"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			cfg, err := parseArgs()

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseArgs() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseArgs() unexpected error: %v", err)
				return
			}

			if cfg.Mode != tt.want.Mode {
				t.Errorf("Mode = %v, want %v", cfg.Mode, tt.want.Mode)
			}
			if cfg.Task != tt.want.Task {
				t.Errorf("Task = %v, want %v", cfg.Task, tt.want.Task)
			}
			if cfg.WorkDir != tt.want.WorkDir {
				t.Errorf("WorkDir = %v, want %v", cfg.WorkDir, tt.want.WorkDir)
			}
			if cfg.ExplicitStdin != tt.want.ExplicitStdin {
				t.Errorf("ExplicitStdin = %v, want %v", cfg.ExplicitStdin, tt.want.ExplicitStdin)
			}
		})
	}
}

func TestParseArgs_ResumeMode(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *Config
		wantErr bool
	}{
		{
			name: "resume with task",
			args: []string{"codex-wrapper", "resume", "session-123", "continue task"},
			want: &Config{
				Mode:          "resume",
				SessionID:     "session-123",
				Task:          "continue task",
				WorkDir:       ".",
				ExplicitStdin: false,
			},
		},
		{
			name: "resume with workdir",
			args: []string{"codex-wrapper", "resume", "session-456", "task", "/work"},
			want: &Config{
				Mode:          "resume",
				SessionID:     "session-456",
				Task:          "task",
				WorkDir:       "/work",
				ExplicitStdin: false,
			},
		},
		{
			name: "resume with stdin",
			args: []string{"codex-wrapper", "resume", "session-789", "-"},
			want: &Config{
				Mode:          "resume",
				SessionID:     "session-789",
				Task:          "-",
				WorkDir:       ".",
				ExplicitStdin: true,
			},
		},
		{
			name:    "resume missing session_id",
			args:    []string{"codex-wrapper", "resume"},
			wantErr: true,
		},
		{
			name:    "resume missing task",
			args:    []string{"codex-wrapper", "resume", "session-123"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			cfg, err := parseArgs()

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseArgs() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseArgs() unexpected error: %v", err)
				return
			}

			if cfg.Mode != tt.want.Mode {
				t.Errorf("Mode = %v, want %v", cfg.Mode, tt.want.Mode)
			}
			if cfg.SessionID != tt.want.SessionID {
				t.Errorf("SessionID = %v, want %v", cfg.SessionID, tt.want.SessionID)
			}
			if cfg.Task != tt.want.Task {
				t.Errorf("Task = %v, want %v", cfg.Task, tt.want.Task)
			}
			if cfg.WorkDir != tt.want.WorkDir {
				t.Errorf("WorkDir = %v, want %v", cfg.WorkDir, tt.want.WorkDir)
			}
			if cfg.ExplicitStdin != tt.want.ExplicitStdin {
				t.Errorf("ExplicitStdin = %v, want %v", cfg.ExplicitStdin, tt.want.ExplicitStdin)
			}
		})
	}
}

func TestShouldUseStdin(t *testing.T) {
	tests := []struct {
		name  string
		task  string
		piped bool
		want  bool
	}{
		{"simple task", "analyze code", false, false},
		{"piped input", "analyze code", true, true},
		{"contains newline", "line1\nline2", false, true},
		{"contains backslash", "path\\to\\file", false, true},
		{"long task", strings.Repeat("a", 801), false, true},
		{"exactly 800 chars", strings.Repeat("a", 800), false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldUseStdin(tt.task, tt.piped)
			if got != tt.want {
				t.Errorf("shouldUseStdin(%q, %v) = %v, want %v", truncate(tt.task, 20), tt.piped, got, tt.want)
			}
		})
	}
}

func TestBuildCodexArgs_NewMode(t *testing.T) {
	cfg := &Config{
		Mode:    "new",
		WorkDir: "/test/dir",
	}

	args := buildCodexArgs(cfg, "my task")

	expected := []string{
		"e",
		"--skip-git-repo-check",
		"-C", "/test/dir",
		"--json",
		"my task",
	}

	if len(args) != len(expected) {
		t.Errorf("buildCodexArgs() returned %d args, want %d", len(args), len(expected))
		return
	}

	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("buildCodexArgs()[%d] = %v, want %v", i, arg, expected[i])
		}
	}
}

func TestBuildCodexArgs_ResumeMode(t *testing.T) {
	cfg := &Config{
		Mode:      "resume",
		SessionID: "session-abc",
	}

	args := buildCodexArgs(cfg, "-")

	expected := []string{
		"e",
		"--skip-git-repo-check",
		"--json",
		"resume",
		"session-abc",
		"-",
	}

	if len(args) != len(expected) {
		t.Errorf("buildCodexArgs() returned %d args, want %d", len(args), len(expected))
		return
	}

	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("buildCodexArgs()[%d] = %v, want %v", i, arg, expected[i])
		}
	}
}

func TestResolveTimeout(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   int
	}{
		{"empty env", "", 7200},
		{"milliseconds", "7200000", 7200},
		{"seconds", "3600", 3600},
		{"invalid", "invalid", 7200},
		{"negative", "-100", 7200},
		{"zero", "0", 7200},
		{"small milliseconds", "5000", 5000},
		{"boundary", "10000", 10000},
		{"above boundary", "10001", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("CODEX_TIMEOUT", tt.envVal)
			defer os.Unsetenv("CODEX_TIMEOUT")

			got := resolveTimeout()
			if got != tt.want {
				t.Errorf("resolveTimeout() with env=%q = %v, want %v", tt.envVal, got, tt.want)
			}
		})
	}
}

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"string", "hello world", "hello world"},
		{"string array", []interface{}{"hello", " ", "world"}, "hello world"},
		{"empty array", []interface{}{}, ""},
		{"mixed array", []interface{}{"text", 123, "more"}, "textmore"},
		{"nil", nil, ""},
		{"number", 123, ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeText(tt.input)
			if got != tt.want {
				t.Errorf("normalizeText(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseJSONStream(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantMessage  string
		wantThreadID string
	}{
		{
			name: "thread started and agent message",
			input: `{"type":"thread.started","thread_id":"abc-123"}
{"type":"item.completed","item":{"type":"agent_message","text":"Hello world"}}`,
			wantMessage:  "Hello world",
			wantThreadID: "abc-123",
		},
		{
			name: "multiple agent messages (last wins)",
			input: `{"type":"item.completed","item":{"type":"agent_message","text":"First"}}
{"type":"item.completed","item":{"type":"agent_message","text":"Second"}}`,
			wantMessage:  "Second",
			wantThreadID: "",
		},
		{
			name:         "text as array",
			input:        `{"type":"item.completed","item":{"type":"agent_message","text":["Hello"," ","World"]}}`,
			wantMessage:  "Hello World",
			wantThreadID: "",
		},
		{
			name: "ignore other event types",
			input: `{"type":"other.event","data":"ignored"}
{"type":"item.completed","item":{"type":"other_type","text":"ignored"}}
{"type":"item.completed","item":{"type":"agent_message","text":"Valid"}}`,
			wantMessage:  "Valid",
			wantThreadID: "",
		},
		{
			name:         "empty input",
			input:        "",
			wantMessage:  "",
			wantThreadID: "",
		},
		{
			name:         "invalid JSON (skipped)",
			input:        "not valid json\n{\"type\":\"thread.started\",\"thread_id\":\"xyz\"}",
			wantMessage:  "",
			wantThreadID: "xyz",
		},
		{
			name:         "blank lines ignored",
			input:        "\n\n{\"type\":\"thread.started\",\"thread_id\":\"test\"}\n\n",
			wantMessage:  "",
			wantThreadID: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			gotMessage, gotThreadID := parseJSONStream(r)

			if gotMessage != tt.wantMessage {
				t.Errorf("parseJSONStream() message = %q, want %q", gotMessage, tt.wantMessage)
			}
			if gotThreadID != tt.wantThreadID {
				t.Errorf("parseJSONStream() threadID = %q, want %q", gotThreadID, tt.wantThreadID)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		defaultVal string
		envVal     string
		setEnv     bool
		want       string
	}{
		{"env set", "TEST_KEY", "default", "custom", true, "custom"},
		{"env not set", "TEST_KEY_MISSING", "default", "", false, "default"},
		{"env empty", "TEST_KEY_EMPTY", "default", "", true, "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)
			if tt.setEnv {
				os.Setenv(tt.key, tt.envVal)
				defer os.Unsetenv(tt.key)
			}

			got := getEnv(tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("getEnv(%q, %q) = %q, want %q", tt.key, tt.defaultVal, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate", "hello world", 5, "hello..."},
		{"empty", "", 5, ""},
		{"zero maxLen", "hello", 0, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, 0, -1},
		{0, -1, -1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := min(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestLogFunctions(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logInfo("info message")
	logWarn("warn message")
	logError("error message")

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "INFO: info message") {
		t.Errorf("logInfo output missing, got: %s", output)
	}
	if !strings.Contains(output, "WARN: warn message") {
		t.Errorf("logWarn output missing, got: %s", output)
	}
	if !strings.Contains(output, "ERROR: error message") {
		t.Errorf("logError output missing, got: %s", output)
	}
}

func TestPrintHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printHelp()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedPhrases := []string{
		"codex-wrapper",
		"Usage:",
		"resume",
		"CODEX_TIMEOUT",
		"Exit Codes:",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("printHelp() missing phrase %q", phrase)
		}
	}
}

// Tests for isTerminal with mock
func TestIsTerminal(t *testing.T) {
	defer resetTestHooks()

	tests := []struct {
		name   string
		mockFn func() bool
		want   bool
	}{
		{"is terminal", func() bool { return true }, true},
		{"is not terminal", func() bool { return false }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isTerminalFn = tt.mockFn
			got := isTerminal()
			if got != tt.want {
				t.Errorf("isTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Tests for readPipedTask with mock
func TestReadPipedTask(t *testing.T) {
	defer resetTestHooks()

	tests := []struct {
		name         string
		isTerminal   bool
		stdinContent string
		want         string
	}{
		{"terminal mode", true, "ignored", ""},
		{"piped with data", false, "task from pipe", "task from pipe"},
		{"piped empty", false, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isTerminalFn = func() bool { return tt.isTerminal }
			stdinReader = strings.NewReader(tt.stdinContent)

			got := readPipedTask()
			if got != tt.want {
				t.Errorf("readPipedTask() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Tests for runCodexProcess with mock command
func TestRunCodexProcess_CommandNotFound(t *testing.T) {
	defer resetTestHooks()

	codexCommand = "nonexistent-command-xyz"

	_, _, exitCode := runCodexProcess([]string{"arg1"}, "task", false, 10)

	if exitCode != 127 {
		t.Errorf("runCodexProcess() exitCode = %d, want 127 for command not found", exitCode)
	}
}

func TestRunCodexProcess_WithEcho(t *testing.T) {
	defer resetTestHooks()

	// Use echo to simulate codex output
	codexCommand = "echo"

	jsonOutput := `{"type":"thread.started","thread_id":"test-session"}
{"type":"item.completed","item":{"type":"agent_message","text":"Test output"}}`

	message, threadID, exitCode := runCodexProcess([]string{jsonOutput}, "", false, 10)

	if exitCode != 0 {
		t.Errorf("runCodexProcess() exitCode = %d, want 0", exitCode)
	}
	if message != "Test output" {
		t.Errorf("runCodexProcess() message = %q, want %q", message, "Test output")
	}
	if threadID != "test-session" {
		t.Errorf("runCodexProcess() threadID = %q, want %q", threadID, "test-session")
	}
}

func TestRunCodexProcess_NoMessage(t *testing.T) {
	defer resetTestHooks()

	codexCommand = "echo"

	// Output without agent_message
	jsonOutput := `{"type":"thread.started","thread_id":"test-session"}`

	_, _, exitCode := runCodexProcess([]string{jsonOutput}, "", false, 10)

	if exitCode != 1 {
		t.Errorf("runCodexProcess() exitCode = %d, want 1 for no message", exitCode)
	}
}

func TestRunCodexProcess_WithStdin(t *testing.T) {
	defer resetTestHooks()

	// Use cat to echo stdin back
	codexCommand = "cat"

	message, _, exitCode := runCodexProcess([]string{}, `{"type":"item.completed","item":{"type":"agent_message","text":"from stdin"}}`, true, 10)

	if exitCode != 0 {
		t.Errorf("runCodexProcess() exitCode = %d, want 0", exitCode)
	}
	if message != "from stdin" {
		t.Errorf("runCodexProcess() message = %q, want %q", message, "from stdin")
	}
}

func TestRunCodexProcess_ExitError(t *testing.T) {
	defer resetTestHooks()

	// Use false command which exits with code 1
	codexCommand = "false"

	_, _, exitCode := runCodexProcess([]string{}, "", false, 10)

	if exitCode == 0 {
		t.Errorf("runCodexProcess() exitCode = 0, want non-zero for failed command")
	}
}

func TestDefaultIsTerminal(t *testing.T) {
	// This test just ensures defaultIsTerminal doesn't panic
	// The actual result depends on the test environment
	_ = defaultIsTerminal()
}

// Tests for run() function
func TestRun_Version(t *testing.T) {
	defer resetTestHooks()

	os.Args = []string{"codex-wrapper", "--version"}
	exitCode := run()
	if exitCode != 0 {
		t.Errorf("run() with --version returned %d, want 0", exitCode)
	}
}

func TestRun_VersionShort(t *testing.T) {
	defer resetTestHooks()

	os.Args = []string{"codex-wrapper", "-v"}
	exitCode := run()
	if exitCode != 0 {
		t.Errorf("run() with -v returned %d, want 0", exitCode)
	}
}

func TestRun_Help(t *testing.T) {
	defer resetTestHooks()

	os.Args = []string{"codex-wrapper", "--help"}
	exitCode := run()
	if exitCode != 0 {
		t.Errorf("run() with --help returned %d, want 0", exitCode)
	}
}

func TestRun_HelpShort(t *testing.T) {
	defer resetTestHooks()

	os.Args = []string{"codex-wrapper", "-h"}
	exitCode := run()
	if exitCode != 0 {
		t.Errorf("run() with -h returned %d, want 0", exitCode)
	}
}

func TestRun_NoArgs(t *testing.T) {
	defer resetTestHooks()

	os.Args = []string{"codex-wrapper"}
	exitCode := run()
	if exitCode != 1 {
		t.Errorf("run() with no args returned %d, want 1", exitCode)
	}
}

func TestRun_ExplicitStdinEmpty(t *testing.T) {
	defer resetTestHooks()

	os.Args = []string{"codex-wrapper", "-"}
	stdinReader = strings.NewReader("")
	isTerminalFn = func() bool { return false }

	exitCode := run()
	if exitCode != 1 {
		t.Errorf("run() with empty stdin returned %d, want 1", exitCode)
	}
}

func TestRun_CommandFails(t *testing.T) {
	defer resetTestHooks()

	os.Args = []string{"codex-wrapper", "task"}
	stdinReader = strings.NewReader("")
	isTerminalFn = func() bool { return true }
	codexCommand = "false"

	exitCode := run()
	if exitCode == 0 {
		t.Errorf("run() with failing command returned 0, want non-zero")
	}
}
