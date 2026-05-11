package app

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var errVersion = fmt.Errorf("version printed")

// CLI is the entry point. Returns an exit code.
func CLI(args []string, version, commit string) int {
	var app appEnv
	if err := app.fromArgs(args, version, commit); err != nil {
		if err == errVersion {
			return 0
		}
		return 1
	}
	app.registerDefaultTools()
	if err := app.run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}
	return 0
}

type appEnv struct {
	hc         http.Client
	apiKey     string
	model      string
	baseURL    string
	systemMsg  string
	version    string
	commit     string
	tools      []ToolDefinition
}

func (app *appEnv) fromArgs(args []string, version, commit string) error {
	app.version = version
	app.commit = commit
	app.hc = *http.DefaultClient

	fl := flag.NewFlagSet("code-agent", flag.ContinueOnError)
	fl.StringVar(&app.apiKey, "key", "", "API key (or set CODE_AGENT_API_KEY env var)")
	fl.StringVar(&app.model, "model", "qwen/qwn3.5-9b", "model to use")
	fl.StringVar(&app.baseURL, "url", "https://openrouter.ai/api/v1", "OpenAI-compatible API base URL")
	fl.StringVar(&app.systemMsg, "system", "You are a helpful coding assistant. Reply concisely.", "system prompt")
	fl.DurationVar(&app.hc.Timeout, "t", 120*time.Second, "request timeout")

	showVersion := fl.Bool("version", false, "print version and exit")

	if err := fl.Parse(args); err != nil {
		return err
	}

	if *showVersion {
		fmt.Printf("code-agent %s (%s)\n", app.version, app.commit)
		return errVersion
	}

	app.baseURL = strings.TrimRight(strings.TrimSpace(app.baseURL), "/")

	// Resolve API key: flag > env var
	if app.apiKey == "" {
		app.apiKey = os.Getenv("CODE_AGENT_API_KEY")
	}
	if app.apiKey == "" {
		fmt.Fprintf(os.Stderr, "Error: API key required. Use -key flag or CODE_AGENT_API_KEY env var.\n")
		fl.Usage()
		return flag.ErrHelp
	}

	return nil
}

func (app *appEnv) RegisterTool(name, description string, inputSchema json.RawMessage, fn func(input json.RawMessage) (string, error)) {
	app.tools = append(app.tools, ToolDefinition{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		Function:    fn,
	})
}

func (app *appEnv) registerDefaultTools() {
	app.RegisterTool(ReadFileDefinition.Name, ReadFileDefinition.Description, ReadFileDefinition.InputSchema, ReadFileDefinition.Function)
	app.RegisterTool(ListFilesDefinition.Name, ListFilesDefinition.Description, ListFilesDefinition.InputSchema, ListFilesDefinition.Function)
	app.RegisterTool(EditFileDefinition.Name, EditFileDefinition.Description, EditFileDefinition.InputSchema, EditFileDefinition.Function)
}

func (app *appEnv) run() error {
	// Build initial conversation with system prompt
	messages := []Message{
		{Role: "system", Content: app.systemMsg},
	}

	fmt.Println("Code Agent (Ctrl+C to exit)")
	fmt.Println("─────────────────────────────")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			break
		}

		messages = append(messages, Message{Role: "user", Content: input})

		response, err := app.sendCompletion(messages)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
			// Remove the failed user message
			messages = messages[:len(messages)-1]
			continue
		}

		messages = append(messages, Message{Role: "assistant", Content: response})
		fmt.Printf("\n%s\n", response)
	}

	return nil
}