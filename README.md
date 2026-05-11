# code-agent

**This is an AI-coded clanker harness that follows the [guide on how to build an agent by Thorsten Ball](https://ampcode.com/notes/how-to-build-an-agent).**

A simple interactive coding agent that _potentially_ talks to any OpenAI-compatible API (OpenRouter, Ollama, OpenAI, etc.).

No use of the [OpenAI Go SDK](https://github.com/openai/openai-go) as I couldn't get it to compile on my VM. The repo shows a plain HTTP-only approach and is thus very limited.

I tested it with OpenRouter:"qwen/qwen3.5-9b". Models like "meta-llama/llama-3.1-8b-instruct" don't seem to work. 

## Build

```bash
go build -o code-agent .
```

With version info:

```bash
go build -ldflags "-X main.version=0.1.0 -X main.commit=$(git rev-parse --short HEAD)" -o code-agent .
```

## Usage

```bash
./code-agent [flags]
```

The agent starts an interactive chat session. Type a message and press Enter. Type `exit` or `quit` to leave.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-key` | | API key (or set `CODE_AGENT_API_KEY` env var) |
| `-model` | `anthropic/claude-3.5-sonnet` | Model identifier |
| `-url` | `https://openrouter.ai/api/v1` | OpenAI-compatible API base URL |
| `-system` | `You are a helpful coding assistant. Reply concisely.` | System prompt |
| `-t` | `120s` | Request timeout |
| `-version` | | Print version and exit |

## Examples

### OpenRouter (default)

```bash
export OPENROUTER_API_KEY=sk-or-...
./code-agent
```

### OpenAI

```bash
./code-agent -key $OPENAI_API_KEY -model gpt-4 -url https://api.openai.com/v1
```

### Ollama (local)

```bash
./code-agent -model llama3 -url http://localhost:11434/v1 -key ollama
```

### Custom system prompt

```bash
./code-agent -system "You are a Python expert. Always include type hints."
```

### Pipe a single question (non-interactive)

```bash
echo "Write a Go function to reverse a string" | ./code-agent
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | User error (bad flags, missing API key) |
| 2 | Runtime error (network, API failure) |

## Tools

| Tool | Description |
|------|-------------|
| `list_files` | List files and directories at a given path. If no path is provided, lists files in the current directory. |
| `edit_file` | Make edits to a text file. Replaces 'old_str' with 'new_str' in the given file. 'old_str' and 'new_str' MUST be different from each other. If the file specified with path doesn't exist, it will be created. |
| `read_file` | Read the content of a file. |

## Project Structure

```
main.go          # Entry point: os.Exit(app.CLI(...))
app/
  app.go         # CLI entry, flag parsing, conversation loop
  api.go         # OpenAI-compatible types, HTTP completion helper
  edit_file.go   # Tool for editing files
  edit_file_test.go # Tests for the edit_file tool
  list_files.go  # Tool for listing files in a directory
  read_file.go   # Tool for reading file content
go.mod
```
