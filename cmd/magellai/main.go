// ABOUTME: Main entry point for the Magellai CLI
// ABOUTME: Handles command parsing and execution using Kong

package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/willabides/kongplete"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/command/core"
	"github.com/lexlapax/magellai/pkg/config"
)

// Version information (set during build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// CLI represents the command line interface structure
type CLI struct {
	// Global flags
	Verbosity   int    `short:"v" type:"counter" help:"Increase verbosity level"`
	Output      string `short:"o" enum:"text,json,markdown" default:"text" help:"Output format"`
	ConfigFile  string `short:"c" type:"path" help:"Config file to use"`
	Profile     string `help:"Configuration profile to use"`
	NoColor     bool   `help:"Disable color output"`
	ShowVersion bool   `name:"version" help:"Show version information"`

	// Subcommands
	Version VersionCmd `cmd:"" help:"Show version information"`
	Ask     AskCmd     `cmd:"" help:"Send a one-shot query to the LLM"`
	Chat    ChatCmd    `cmd:"" help:"Start an interactive chat session"`
	Config  ConfigCmd  `cmd:"" help:"Manage configuration"`

	// Hidden completion command
	InstallCompletions kongplete.InstallCompletions `cmd:"" help:"Install shell completions"`
}

// VersionCmd handles the version command
type VersionCmd struct{}

// Run executes the version command
func (v *VersionCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
		Flags:   make(map[string]interface{}),
		Data:    make(map[string]interface{}),
	}

	// Pass the output format if specified globally
	switch v := ctx.Model.Target.Interface().(type) {
	case CLI:
		if v.Output != "text" {
			exec.Flags["format"] = v.Output
		}
	case *CLI:
		if v.Output != "text" {
			exec.Flags["format"] = v.Output
		}
	}

	err := ctx.Registry.GetExecutor().Execute(ctx.Ctx, "version", exec)
	if err != nil {
		return err
	}

	// Print the output
	if output, ok := exec.Data["output"].(string); ok {
		fmt.Fprintln(ctx.Stdout, output)
	}

	return nil
}

// AskCmd handles the ask command
type AskCmd struct {
	Prompt      string   `arg:"" required:"" help:"The prompt to send to the LLM"`
	Model       string   `short:"m" help:"Model to use (provider/model format)"`
	Attach      []string `short:"a" help:"Files to attach to the prompt"`
	Stream      bool     `short:"s" help:"Enable streaming response"`
	Temperature float64  `short:"t" help:"Temperature for the model"`
}

// Run executes the ask command
func (a *AskCmd) Run(ctx *Context) error {
	// Convert Kong command to our command system
	exec := &command.ExecutionContext{
		Args:    []string{a.Prompt},
		Flags:   make(map[string]interface{}),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}

	// Map flags
	if a.Model != "" {
		exec.Flags["model"] = a.Model
	}
	if len(a.Attach) > 0 {
		exec.Flags["attach"] = a.Attach
	}
	if a.Stream {
		exec.Flags["stream"] = a.Stream
	}
	if a.Temperature != 0 {
		exec.Flags["temperature"] = a.Temperature
	}

	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "ask", exec)
}

// ChatCmd handles the chat command
type ChatCmd struct {
	Resume string   `short:"r" help:"Resume a previous session by ID"`
	Model  string   `short:"m" help:"Model to use (provider/model format)"`
	Attach []string `short:"a" help:"Initial files to attach"`
}

// Run executes the chat command
func (c *ChatCmd) Run(ctx *Context) error {
	// Convert Kong command to our command system
	exec := &command.ExecutionContext{
		Args:    []string{},
		Flags:   make(map[string]interface{}),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}

	// Map flags
	if c.Resume != "" {
		exec.Flags["resume"] = c.Resume
	}
	if c.Model != "" {
		exec.Flags["model"] = c.Model
	}
	if len(c.Attach) > 0 {
		exec.Flags["attach"] = c.Attach
	}

	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "chat", exec)
}

// ConfigCmd handles the config command
type ConfigCmd struct {
	List     ConfigListCmd     `cmd:"" help:"List configuration settings"`
	Get      ConfigGetCmd      `cmd:"" help:"Get a configuration value"`
	Set      ConfigSetCmd      `cmd:"" help:"Set a configuration value"`
	Validate ConfigValidateCmd `cmd:"" help:"Validate configuration"`
}

// ConfigListCmd handles config list
type ConfigListCmd struct{}

func (c *ConfigListCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"list"},
		Flags:   make(map[string]interface{}),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "config", exec)
}

// ConfigGetCmd handles config get
type ConfigGetCmd struct {
	Key string `arg:"" required:"" help:"Configuration key to get"`
}

func (c *ConfigGetCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"get", c.Key},
		Flags:   make(map[string]interface{}),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "config", exec)
}

// ConfigSetCmd handles config set
type ConfigSetCmd struct {
	Key   string `arg:"" required:"" help:"Configuration key to set"`
	Value string `arg:"" required:"" help:"Value to set"`
}

func (c *ConfigSetCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"set", c.Key, c.Value},
		Flags:   make(map[string]interface{}),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "config", exec)
}

// ConfigValidateCmd handles config validate
type ConfigValidateCmd struct{}

func (c *ConfigValidateCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"validate"},
		Flags:   make(map[string]interface{}),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "config", exec)
}

// Context provides runtime context for commands
type Context struct {
	*kong.Context
	Registry *command.Registry
	Config   *config.Config
	Logger   *logging.Logger
	Stdout   io.Writer
	Stderr   io.Writer
	Ctx      context.Context
}

func main() {
	// Create the CLI parser
	parser := kong.Must(&CLI{},
		kong.Name("magellai"),
		kong.Description("A command-line interface for interacting with Large Language Models"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	// Check for version flag early
	for _, arg := range os.Args[1:] {
		if arg == "--version" {
			fmt.Printf("magellai version %s\n", version)
			os.Exit(0)
		}
	}

	// Parse arguments
	kongCtx, err := parser.Parse(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	// Initialize logger
	logConfig := logging.LogConfig{
		Level:      "info",
		Format:     "text",
		OutputPath: "stderr",
		AddSource:  false,
	}
	if err := logging.Initialize(logConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	logger := logging.GetLogger()

	// Initialize configuration
	if err := config.Init(); err != nil {
		logger.Error("Failed to initialize configuration", "error", err)
		os.Exit(1)
	}
	cfg := config.Manager

	// Apply global flags to configuration
	var cli CLI
	switch v := kongCtx.Model.Target.Interface().(type) {
	case CLI:
		cli = v
	case *CLI:
		cli = *v
	default:
		logger.Error("Failed to get CLI interface")
		os.Exit(1)
	}
	if cli.ConfigFile != "" {
		// Load specific config file
		if err := cfg.LoadFile(cli.ConfigFile); err != nil {
			logger.Error("Failed to load config file", "file", cli.ConfigFile, "error", err)
			os.Exit(1)
		}
	}
	if cli.Profile != "" {
		if err := cfg.SetProfile(cli.Profile); err != nil {
			logger.Error("Failed to set profile", "profile", cli.Profile, "error", err)
			os.Exit(1)
		}
	}

	// Set verbosity
	if cli.Verbosity > 0 {
		switch cli.Verbosity {
		case 1:
			logger.SetLevel(slog.LevelDebug)
		default:
			logger.SetLevel(slog.LevelDebug) // slog doesn't have trace level, use debug
		}
	}

	// Initialize command registry
	registry := command.NewRegistry()

	// Register core commands
	configCmd := core.NewConfigCommand(cfg)
	if err := registry.Register(configCmd); err != nil {
		logger.Error("failed to register config command", "error", err)
		os.Exit(1)
	}

	profileCmd := core.NewProfileCommand(cfg)
	if err := registry.Register(profileCmd); err != nil {
		logger.Error("failed to register profile command", "error", err)
		os.Exit(1)
	}

	modelCmd := core.NewModelCommand(cfg)
	if err := registry.Register(modelCmd); err != nil {
		logger.Error("failed to register model command", "error", err)
		os.Exit(1)
	}

	aliasCmd := core.NewAliasCommand(cfg)
	if err := registry.Register(aliasCmd); err != nil {
		logger.Error("failed to register alias command", "error", err)
		os.Exit(1)
	}

	helpCmd := core.NewHelpCommand(registry, cfg)
	if err := registry.Register(helpCmd); err != nil {
		logger.Error("failed to register help command", "error", err)
		os.Exit(1)
	}

	versionCmd := core.NewVersionCommand(version, commit, date)
	if err := registry.Register(versionCmd); err != nil {
		logger.Error("failed to register version command", "error", err)
		os.Exit(1)
	}

	// Register stub commands (temporary implementations)
	if err := RegisterStubCommands(registry); err != nil {
		logger.Error("failed to register stub commands", "error", err)
		os.Exit(1)
	}

	// Create context
	ctx := &Context{
		Context:  kongCtx,
		Registry: registry,
		Config:   cfg,
		Logger:   logger,
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		Ctx:      context.Background(),
	}

	// Handle completions
	kongplete.Complete(parser)

	// Run the command
	err = kongCtx.Run(ctx)
	if err != nil {
		logger.Error("Command failed", "error", err)
		os.Exit(1)
	}
}
