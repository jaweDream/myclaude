#!/usr/bin/env node
import { spawn } from 'node:child_process';

const DEFAULT_MODEL = 'gpt-5-codex';
const DEFAULT_WORKDIR = '.';
const DEFAULT_TIMEOUT_MS = 7_200_000; // 2 hours
const FORCE_KILL_DELAY_MS = 5_000;

const args = process.argv.slice(2);
const [task, modelArg, workdirArg] = args;

const logError = (message) => {
  process.stderr.write(`ERROR: ${message}\n`);
};

const logWarn = (message) => {
  process.stderr.write(`WARN: ${message}\n`);
};

if (!task) {
  logError('Task required');
  process.exit(1);
}

const model = modelArg || DEFAULT_MODEL;
const workdir = workdirArg || DEFAULT_WORKDIR;

const resolveTimeout = () => {
  const raw = process.env.CODEX_TIMEOUT;
  if (raw == null || raw === '') {
    return DEFAULT_TIMEOUT_MS;
  }

  const parsed = Number(raw);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    logWarn(`Invalid CODEX_TIMEOUT "${raw}", falling back to ${DEFAULT_TIMEOUT_MS}ms`);
    return DEFAULT_TIMEOUT_MS;
  }

  return parsed;
};

const codexArgs = [
  'e',
  '-m',
  model,
  '--dangerously-bypass-approvals-and-sandbox',
  '--skip-git-repo-check',
  '-C',
  workdir,
  '--json',
  task,
];

const child = spawn('codex', codexArgs, {
  stdio: ['ignore', 'pipe', 'inherit'],
});

let timedOut = false;
let lastAgentMessage = null;
let stdoutBuffer = '';
let forceKillTimer = null;

const timeoutMs = resolveTimeout();

const forceTerminate = () => {
  if (!child.killed) {
    child.kill('SIGTERM');
    forceKillTimer = setTimeout(() => {
      if (!child.killed) {
        child.kill('SIGKILL');
      }
    }, FORCE_KILL_DELAY_MS);
  }
};

const timeoutHandle = setTimeout(() => {
  timedOut = true;
  logError('Codex execution timeout');
  forceTerminate();
}, timeoutMs);

const normalizeText = (text) => {
  if (typeof text === 'string') {
    return text;
  }
  if (Array.isArray(text)) {
    return text.join('');
  }
  return null;
};

const handleJsonLine = (line) => {
  const trimmed = line.trim();
  if (!trimmed) {
    return;
  }

  let event;
  try {
    event = JSON.parse(trimmed);
  } catch (err) {
    logWarn(`Failed to parse Codex output line: ${trimmed}`);
    return;
  }

  if (
    event &&
    event.type === 'item.completed' &&
    event.item &&
    event.item.type === 'agent_message'
  ) {
    const text = normalizeText(event.item.text);
    if (text != null) {
      lastAgentMessage = text;
    }
  }
};

child.stdout.on('data', (chunk) => {
  stdoutBuffer += chunk.toString('utf8');
  let newlineIndex = stdoutBuffer.indexOf('\n');

  while (newlineIndex !== -1) {
    const line = stdoutBuffer.slice(0, newlineIndex);
    stdoutBuffer = stdoutBuffer.slice(newlineIndex + 1);
    handleJsonLine(line);
    newlineIndex = stdoutBuffer.indexOf('\n');
  }
});

child.stdout.on('end', () => {
  if (stdoutBuffer) {
    handleJsonLine(stdoutBuffer);
    stdoutBuffer = '';
  }
});

child.on('error', (err) => {
  clearTimeout(timeoutHandle);
  if (forceKillTimer) {
    clearTimeout(forceKillTimer);
  }
  logError(`Failed to start Codex CLI: ${err.message}`);
  process.exit(1);
});

child.on('close', (code, signal) => {
  clearTimeout(timeoutHandle);
  if (forceKillTimer) {
    clearTimeout(forceKillTimer);
  }

  if (timedOut) {
    process.exit(124);
    return;
  }

  if (code === 0) {
    if (lastAgentMessage != null) {
      process.stdout.write(`${lastAgentMessage}\n`);
      process.exit(0);
    } else {
      logError('Codex completed without an agent_message output');
      process.exit(1);
    }
    return;
  }

  if (signal) {
    logError(`Codex terminated with signal ${signal}`);
    process.exit(code ?? 1);
    return;
  }

  logError(`Codex exited with status ${code}`);
  process.exit(code ?? 1);
});
