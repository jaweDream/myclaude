package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	version           = "1.0.0"
	defaultWorkdir    = "."
	defaultTimeout    = 7200 // seconds
	forceKillDelay    = 5    // seconds
	stdinSpecialChars = "\n\\\"'`$"
)

// Test hooks for dependency injection
var (
	stdinReader      io.Reader = os.Stdin
	isTerminalFn               = defaultIsTerminal
	codexCommand               = "codex"
	buildCodexArgsFn           = buildCodexArgs
	commandContext             = exec.CommandContext
	jsonMarshal                = json.Marshal
)

// Config holds CLI configuration
type Config struct {
	Mode          string // "new" or "resume"
	Task          string
	SessionID     string
	WorkDir       string
	ExplicitStdin bool
	Timeout       int
}

// ParallelConfig defines the JSON schema for parallel execution
type ParallelConfig struct {
	Tasks []TaskSpec `json:"tasks"`
}

// TaskSpec describes an individual task entry in the parallel config
type TaskSpec struct {
	ID           string   `json:"id"`
	Task         string   `json:"task"`
	WorkDir      string   `json:"workdir,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	SessionID    string   `json:"session_id,omitempty"`
	Mode         string   `json:"-"`
	UseStdin     bool     `json:"-"`
}

// TaskResult captures the execution outcome of a task
type TaskResult struct {
	TaskID    string `json:"task_id"`
	ExitCode  int    `json:"exit_code"`
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
	Error     string `json:"error"`
}

func parseParallelConfig(data []byte) (*ParallelConfig, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("parallel config is empty")
	}

	tasks := strings.Split(string(trimmed), "---TASK---")
	var cfg ParallelConfig
	seen := make(map[string]struct{})

	for _, taskBlock := range tasks {
		taskBlock = strings.TrimSpace(taskBlock)
		if taskBlock == "" {
			continue
		}

		parts := strings.SplitN(taskBlock, "---CONTENT---", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("task block missing ---CONTENT--- separator")
		}

		meta := strings.TrimSpace(parts[0])
		content := strings.TrimSpace(parts[1])

		task := TaskSpec{WorkDir: defaultWorkdir}
		for _, line := range strings.Split(meta, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			kv := strings.SplitN(line, ":", 2)
			if len(kv) != 2 {
				continue
			}
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])

			switch key {
			case "id":
				task.ID = value
			case "workdir":
				task.WorkDir = value
			case "session_id":
				task.SessionID = value
				task.Mode = "resume"
			case "dependencies":
				for _, dep := range strings.Split(value, ",") {
					dep = strings.TrimSpace(dep)
					if dep != "" {
						task.Dependencies = append(task.Dependencies, dep)
					}
				}
			}
		}

		if task.ID == "" {
			return nil, fmt.Errorf("task missing id field")
		}
		if content == "" {
			return nil, fmt.Errorf("task %q missing content", task.ID)
		}
		if _, exists := seen[task.ID]; exists {
			return nil, fmt.Errorf("duplicate task id: %s", task.ID)
		}

		task.Task = content
		cfg.Tasks = append(cfg.Tasks, task)
		seen[task.ID] = struct{}{}
	}

	if len(cfg.Tasks) == 0 {
		return nil, fmt.Errorf("no tasks found")
	}

	return &cfg, nil
}

func topologicalSort(tasks []TaskSpec) ([][]TaskSpec, error) {
	idToTask := make(map[string]TaskSpec, len(tasks))
	indegree := make(map[string]int, len(tasks))
	adj := make(map[string][]string, len(tasks))

	for _, task := range tasks {
		idToTask[task.ID] = task
		indegree[task.ID] = 0
	}

	for _, task := range tasks {
		for _, dep := range task.Dependencies {
			if _, ok := idToTask[dep]; !ok {
				return nil, fmt.Errorf("dependency %q not found for task %q", dep, task.ID)
			}
			indegree[task.ID]++
			adj[dep] = append(adj[dep], task.ID)
		}
	}

	queue := make([]string, 0, len(tasks))
	for _, task := range tasks {
		if indegree[task.ID] == 0 {
			queue = append(queue, task.ID)
		}
	}

	layers := make([][]TaskSpec, 0)
	processed := 0

	for len(queue) > 0 {
		current := queue
		queue = nil
		layer := make([]TaskSpec, len(current))
		for i, id := range current {
			layer[i] = idToTask[id]
			processed++
		}
		layers = append(layers, layer)

		next := make([]string, 0)
		for _, id := range current {
			for _, neighbor := range adj[id] {
				indegree[neighbor]--
				if indegree[neighbor] == 0 {
					next = append(next, neighbor)
				}
			}
		}
		queue = append(queue, next...)
	}

	if processed != len(tasks) {
		cycleIDs := make([]string, 0)
		for id, deg := range indegree {
			if deg > 0 {
				cycleIDs = append(cycleIDs, id)
			}
		}
		sort.Strings(cycleIDs)
		return nil, fmt.Errorf("cycle detected involving tasks: %s", strings.Join(cycleIDs, ","))
	}

	return layers, nil
}

var runCodexTaskFn = func(task TaskSpec, timeout int) TaskResult {
	if task.WorkDir == "" {
		task.WorkDir = defaultWorkdir
	}
	if task.Mode == "" {
		task.Mode = "new"
	}
	if task.UseStdin || shouldUseStdin(task.Task, false) {
		task.UseStdin = true
	}

	return runCodexTask(task, true, timeout)
}

func executeConcurrent(layers [][]TaskSpec, timeout int) []TaskResult {
	totalTasks := 0
	for _, layer := range layers {
		totalTasks += len(layer)
	}

	results := make([]TaskResult, 0, totalTasks)
	failed := make(map[string]TaskResult, totalTasks)
	resultsCh := make(chan TaskResult, totalTasks)

	for _, layer := range layers {
		var wg sync.WaitGroup
		executed := 0

		for _, task := range layer {
			if skip, reason := shouldSkipTask(task, failed); skip {
				res := TaskResult{TaskID: task.ID, ExitCode: 1, Error: reason}
				results = append(results, res)
				failed[task.ID] = res
				continue
			}

			executed++
			wg.Add(1)
			go func(ts TaskSpec) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						resultsCh <- TaskResult{TaskID: ts.ID, ExitCode: 1, Error: fmt.Sprintf("panic: %v", r)}
					}
				}()
				resultsCh <- runCodexTaskFn(ts, timeout)
			}(task)
		}

		wg.Wait()

		for i := 0; i < executed; i++ {
			res := <-resultsCh
			results = append(results, res)
			if res.ExitCode != 0 || res.Error != "" {
				failed[res.TaskID] = res
			}
		}
	}

	return results
}

func shouldSkipTask(task TaskSpec, failed map[string]TaskResult) (bool, string) {
	if len(task.Dependencies) == 0 {
		return false, ""
	}

	var blocked []string
	for _, dep := range task.Dependencies {
		if _, ok := failed[dep]; ok {
			blocked = append(blocked, dep)
		}
	}

	if len(blocked) == 0 {
		return false, ""
	}

	return true, fmt.Sprintf("skipped due to failed dependencies: %s", strings.Join(blocked, ","))
}

func generateFinalOutput(results []TaskResult) string {
	var sb strings.Builder

	success := 0
	failed := 0
	for _, res := range results {
		if res.ExitCode == 0 && res.Error == "" {
			success++
		} else {
			failed++
		}
	}

	sb.WriteString(fmt.Sprintf("=== Parallel Execution Summary ===\n"))
	sb.WriteString(fmt.Sprintf("Total: %d | Success: %d | Failed: %d\n\n", len(results), success, failed))

	for _, res := range results {
		sb.WriteString(fmt.Sprintf("--- Task: %s ---\n", res.TaskID))
		if res.Error != "" {
			sb.WriteString(fmt.Sprintf("Status: FAILED (exit code %d)\nError: %s\n", res.ExitCode, res.Error))
		} else if res.ExitCode != 0 {
			sb.WriteString(fmt.Sprintf("Status: FAILED (exit code %d)\n", res.ExitCode))
		} else {
			sb.WriteString("Status: SUCCESS\n")
		}
		if res.SessionID != "" {
			sb.WriteString(fmt.Sprintf("Session: %s\n", res.SessionID))
		}
		if res.Message != "" {
			sb.WriteString(fmt.Sprintf("\n%s\n", res.Message))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// JSONEvent represents a Codex JSON output event
type JSONEvent struct {
	Type     string     `json:"type"`
	ThreadID string     `json:"thread_id,omitempty"`
	Item     *EventItem `json:"item,omitempty"`
}

// EventItem represents the item field in a JSON event
type EventItem struct {
	Type string      `json:"type"`
	Text interface{} `json:"text"`
}

func main() {
	exitCode := run()
	os.Exit(exitCode)
}

// run is the main logic, returns exit code for testability
func run() int {
	// Handle --version and --help first
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("codex-wrapper version %s\n", version)
			return 0
		case "--help", "-h":
			printHelp()
			return 0
		case "--parallel":
			// Parallel mode: read task config from stdin
			data, err := io.ReadAll(stdinReader)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: failed to read stdin: %v\n", err)
				return 1
			}

			cfg, err := parseParallelConfig(data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
				return 1
			}

			timeoutSec := resolveTimeout()
			layers, err := topologicalSort(cfg.Tasks)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
				return 1
			}

			results := executeConcurrent(layers, timeoutSec)
			fmt.Println(generateFinalOutput(results))

			exitCode := 0
			for _, res := range results {
				if res.ExitCode != 0 {
					exitCode = res.ExitCode
				}
			}

			return exitCode
		}
	}

	logInfo("Script started")

	cfg, err := parseArgs()
	if err != nil {
		logError(err.Error())
		return 1
	}
	logInfo(fmt.Sprintf("Parsed args: mode=%s, task_len=%d", cfg.Mode, len(cfg.Task)))

	timeoutSec := resolveTimeout()
	logInfo(fmt.Sprintf("Timeout: %ds", timeoutSec))
	cfg.Timeout = timeoutSec

	// Determine task text and stdin mode
	var taskText string
	var piped bool

	if cfg.ExplicitStdin {
		logInfo("Explicit stdin mode: reading task from stdin")
		data, err := io.ReadAll(stdinReader)
		if err != nil {
			logError("Failed to read stdin: " + err.Error())
			return 1
		}
		taskText = string(data)
		if taskText == "" {
			logError("Explicit stdin mode requires task input from stdin")
			return 1
		}
		piped = !isTerminal()
	} else {
		pipedTask := readPipedTask()
		piped = pipedTask != ""
		if piped {
			taskText = pipedTask
		} else {
			taskText = cfg.Task
		}
	}

	useStdin := cfg.ExplicitStdin || shouldUseStdin(taskText, piped)

	if useStdin {
		var reasons []string
		if piped {
			reasons = append(reasons, "piped input")
		}
		if cfg.ExplicitStdin {
			reasons = append(reasons, "explicit \"-\"")
		}
		if strings.Contains(taskText, "\n") {
			reasons = append(reasons, "newline")
		}
		if strings.Contains(taskText, "\\") {
			reasons = append(reasons, "backslash")
		}
		if strings.Contains(taskText, "\"") {
			reasons = append(reasons, "double-quote")
		}
		if strings.Contains(taskText, "'") {
			reasons = append(reasons, "single-quote")
		}
		if strings.Contains(taskText, "`") {
			reasons = append(reasons, "backtick")
		}
		if strings.Contains(taskText, "$") {
			reasons = append(reasons, "dollar")
		}
		if len(taskText) > 800 {
			reasons = append(reasons, "length>800")
		}
		if len(reasons) > 0 {
			logWarn(fmt.Sprintf("Using stdin mode for task due to: %s", strings.Join(reasons, ", ")))
		}
	}

	logInfo("codex running...")

	taskSpec := TaskSpec{
		Task:      taskText,
		WorkDir:   cfg.WorkDir,
		Mode:      cfg.Mode,
		SessionID: cfg.SessionID,
		UseStdin:  useStdin,
	}

	result := runCodexTask(taskSpec, false, cfg.Timeout)

	if result.ExitCode != 0 {
		return result.ExitCode
	}

	// Output agent_message
	fmt.Println(result.Message)

	// Output session_id if present
	if result.SessionID != "" {
		fmt.Printf("\n---\nSESSION_ID: %s\n", result.SessionID)
	}

	return 0
}

func parseArgs() (*Config, error) {
	args := os.Args[1:]
	if len(args) == 0 {
		return nil, fmt.Errorf("task required")
	}

	cfg := &Config{
		WorkDir: defaultWorkdir,
	}

	// Check for resume mode
	if args[0] == "resume" {
		if len(args) < 3 {
			return nil, fmt.Errorf("resume mode requires: resume <session_id> <task>")
		}
		cfg.Mode = "resume"
		cfg.SessionID = args[1]
		cfg.Task = args[2]
		cfg.ExplicitStdin = (args[2] == "-")
		if len(args) > 3 {
			cfg.WorkDir = args[3]
		}
	} else {
		cfg.Mode = "new"
		cfg.Task = args[0]
		cfg.ExplicitStdin = (args[0] == "-")
		if len(args) > 1 {
			cfg.WorkDir = args[1]
		}
	}

	return cfg, nil
}

func readPipedTask() string {
	if isTerminal() {
		logInfo("Stdin is tty, skipping pipe read")
		return ""
	}
	logInfo("Reading from stdin pipe...")
	data, err := io.ReadAll(stdinReader)
	if err != nil || len(data) == 0 {
		logInfo("Stdin pipe returned empty data")
		return ""
	}
	logInfo(fmt.Sprintf("Read %d bytes from stdin pipe", len(data)))
	return string(data)
}

func shouldUseStdin(taskText string, piped bool) bool {
	if piped {
		return true
	}
	if len(taskText) > 800 {
		return true
	}
	return strings.IndexAny(taskText, stdinSpecialChars) >= 0
}

func buildCodexArgs(cfg *Config, targetArg string) []string {
	if cfg.Mode == "resume" {
		return []string{
			"e",
			"--skip-git-repo-check",
			"--json",
			"resume",
			cfg.SessionID,
			targetArg,
		}
	}
	return []string{
		"e",
		"--skip-git-repo-check",
		"-C", cfg.WorkDir,
		"--json",
		targetArg,
	}
}

func runCodexTask(taskSpec TaskSpec, silent bool, timeoutSec int) TaskResult {
	result := TaskResult{
		TaskID: taskSpec.ID,
	}

	cfg := &Config{
		Mode:      taskSpec.Mode,
		Task:      taskSpec.Task,
		SessionID: taskSpec.SessionID,
		WorkDir:   taskSpec.WorkDir,
	}
	if cfg.Mode == "" {
		cfg.Mode = "new"
	}
	if cfg.WorkDir == "" {
		cfg.WorkDir = defaultWorkdir
	}

	useStdin := taskSpec.UseStdin
	targetArg := taskSpec.Task
	if useStdin {
		targetArg = "-"
	}

	codexArgs := buildCodexArgsFn(cfg, targetArg)

	logInfoFn := logInfo
	logWarnFn := logWarn
	logErrorFn := logError
	stderrWriter := io.Writer(os.Stderr)
	if silent {
		logInfoFn = func(string) {}
		logWarnFn = func(string) {}
		logErrorFn = func(string) {}
		stderrWriter = io.Discard
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := commandContext(ctx, codexCommand, codexArgs...)
	cmd.Stderr = stderrWriter

	// Setup stdin if needed
	var stdinPipe io.WriteCloser
	var err error
	if useStdin {
		stdinPipe, err = cmd.StdinPipe()
		if err != nil {
			logErrorFn("Failed to create stdin pipe: " + err.Error())
			result.ExitCode = 1
			result.Error = "failed to create stdin pipe: " + err.Error()
			return result
		}
	}

	// Setup stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logErrorFn("Failed to create stdout pipe: " + err.Error())
		result.ExitCode = 1
		result.Error = "failed to create stdout pipe: " + err.Error()
		return result
	}

	logInfoFn(fmt.Sprintf("Starting codex with args: codex %s...", strings.Join(codexArgs[:min(5, len(codexArgs))], " ")))

	// Start process
	if err := cmd.Start(); err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			logErrorFn("codex command not found in PATH")
			result.ExitCode = 127
			result.Error = "codex command not found in PATH"
			return result
		}
		logErrorFn("Failed to start codex: " + err.Error())
		result.ExitCode = 1
		result.Error = "failed to start codex: " + err.Error()
		return result
	}
	logInfoFn(fmt.Sprintf("Process started with PID: %d", cmd.Process.Pid))

	// Write to stdin if needed
	if useStdin && stdinPipe != nil {
		logInfoFn(fmt.Sprintf("Writing %d chars to stdin...", len(taskSpec.Task)))
		go func() {
			defer stdinPipe.Close()
			io.WriteString(stdinPipe, taskSpec.Task)
		}()
		logInfoFn("Stdin closed")
	}

	forwardSignals(ctx, cmd, logErrorFn)

	logInfoFn("Reading stdout...")

	// Parse JSON stream
	message, threadID := parseJSONStreamWithWarn(stdout, logWarnFn)

	// Wait for process to complete
	err = cmd.Wait()

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		logErrorFn("Codex execution timeout")
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		result.ExitCode = 124
		result.Error = "codex execution timeout"
		return result
	}

	// Check exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			logErrorFn(fmt.Sprintf("Codex exited with status %d", code))
			result.ExitCode = code
			result.Error = fmt.Sprintf("codex exited with status %d", code)
			return result
		}
		logErrorFn("Codex error: " + err.Error())
		result.ExitCode = 1
		result.Error = "codex error: " + err.Error()
		return result
	}

	if message == "" {
		logErrorFn("Codex completed without agent_message output")
		result.ExitCode = 1
		result.Error = "codex completed without agent_message output"
		return result
	}

	result.ExitCode = 0
	result.Message = message
	result.SessionID = threadID

	return result
}

func forwardSignals(ctx context.Context, cmd *exec.Cmd, logErrorFn func(string)) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer signal.Stop(sigCh)
		select {
		case sig := <-sigCh:
			logErrorFn(fmt.Sprintf("Received signal: %v", sig))
			if cmd.Process != nil {
				cmd.Process.Signal(syscall.SIGTERM)
				time.AfterFunc(time.Duration(forceKillDelay)*time.Second, func() {
					if cmd.Process != nil {
						cmd.Process.Kill()
					}
				})
			}
		case <-ctx.Done():
		}
	}()
}

func parseJSONStream(r io.Reader) (message, threadID string) {
	return parseJSONStreamWithWarn(r, logWarn)
}

func parseJSONStreamWithWarn(r io.Reader, warnFn func(string)) (message, threadID string) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)

	if warnFn == nil {
		warnFn = func(string) {}
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var event JSONEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			warnFn(fmt.Sprintf("Failed to parse line: %s", truncate(line, 100)))
			continue
		}

		// Capture thread_id
		if event.Type == "thread.started" {
			threadID = event.ThreadID
		}

		// Capture agent_message
		if event.Type == "item.completed" && event.Item != nil && event.Item.Type == "agent_message" {
			if text := normalizeText(event.Item.Text); text != "" {
				message = text
			}
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		warnFn("Read stdout error: " + err.Error())
	}

	return message, threadID
}

func normalizeText(text interface{}) string {
	switch v := text.(type) {
	case string:
		return v
	case []interface{}:
		var sb strings.Builder
		for _, item := range v {
			if s, ok := item.(string); ok {
				sb.WriteString(s)
			}
		}
		return sb.String()
	default:
		return ""
	}
}

func resolveTimeout() int {
	raw := os.Getenv("CODEX_TIMEOUT")
	if raw == "" {
		return defaultTimeout
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		logWarn(fmt.Sprintf("Invalid CODEX_TIMEOUT '%s', falling back to %ds", raw, defaultTimeout))
		return defaultTimeout
	}

	// Environment variable is in milliseconds if > 10000, convert to seconds
	if parsed > 10000 {
		return parsed / 1000
	}
	return parsed
}

func defaultIsTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return true
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func isTerminal() bool {
	return isTerminalFn()
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func hello() string {
	return "hello world"
}

func greet(name string) string {
	return "hello " + name
}

func farewell(name string) string {
	return "goodbye " + name
}

func logInfo(msg string) {
	fmt.Fprintf(os.Stderr, "INFO: %s\n", msg)
}

func logWarn(msg string) {
	fmt.Fprintf(os.Stderr, "WARN: %s\n", msg)
}

func logError(msg string) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", msg)
}

func printHelp() {
	help := `codex-wrapper - Go wrapper for Codex CLI

Usage:
    codex-wrapper "task" [workdir]
    codex-wrapper - [workdir]              Read task from stdin
    codex-wrapper resume <session_id> "task" [workdir]
    codex-wrapper resume <session_id> - [workdir]
    codex-wrapper --version
    codex-wrapper --help

Environment Variables:
    CODEX_TIMEOUT  Timeout in milliseconds (default: 7200000)

Exit Codes:
    0    Success
    1    General error (missing args, no output)
    124  Timeout
    127  codex command not found
    130  Interrupted (Ctrl+C)
    *    Passthrough from codex process`
	fmt.Println(help)
}
