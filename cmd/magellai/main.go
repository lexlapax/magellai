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
	ProfileName string `name:"profile" help:"Configuration profile to use"`
	NoColor     bool   `help:"Disable color output"`
	ShowVersion bool   `name:"version" help:"Show version information"`

	// Subcommands
	Ask  AskCmd  `cmd:"" help:"Send a one-shot query to the LLM" group:"core"`
	Chat ChatCmd `cmd:"" help:"Start an interactive chat session" group:"core"`

	// Help command
	Version VersionCmd `cmd:"" help:"Show version information" group:"info"`

	// Configuration commands
	Config  ConfigCmd  `cmd:"" help:"Manage configuration" group:"config"`
	Model   ModelCmd   `cmd:"" help:"Manage LLM models" group:"config"`
	Profile ProfileCmd `cmd:"" help:"Manage configuration profiles" group:"config"`
	Alias   AliasCmd   `cmd:"" help:"Manage command aliases" group:"config"`

	// Session management commands
	History HistoryCmd `cmd:"" help:"Manage REPL session history" group:"session"`

	// Shell completion command
	InstallCompletions kongplete.InstallCompletions `cmd:"" help:"Install shell completions" group:"config"`
}

// VersionCmd handles the version command
type VersionCmd struct{}

// Run executes the version command
func (v *VersionCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
		Flags:   command.NewFlags(nil),
		Data:    make(map[string]interface{}),
	}

	// Pass the output format if specified globally
	switch v := ctx.Model.Target.Interface().(type) {
	case CLI:
		if v.Output != "text" {
			exec.Flags.Set("format", v.Output)
		}
	case *CLI:
		if v.Output != "text" {
			exec.Flags.Set("format", v.Output)
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
	Prompt         string   `arg:"" optional:"" help:"The prompt to send to the LLM (reads from stdin if not provided)"`
	Model          string   `short:"m" help:"Model to use (provider/model format)"`
	Attach         []string `short:"a" help:"Files to attach to the prompt"`
	Stream         bool     `help:"Enable streaming response"`
	Temperature    float64  `short:"t" help:"Temperature for the model"`
	MaxTokens      int      `name:"max-tokens" help:"Maximum tokens in response"`
	System         string   `short:"s" help:"System prompt"`
	ResponseFormat string   `name:"format" help:"Response format (text, json, markdown)"`
}

// Run executes the ask command
func (a *AskCmd) Run(ctx *Context) error {
	var prompt string
	var stdinData string

	// Check if stdin has data (not a terminal)
	if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		stdinData = string(data)
	}

	// Handle prompt logic
	if a.Prompt != "" && stdinData != "" {
		// Both provided - combine them
		prompt = stdinData + "\n\n" + a.Prompt
	} else if a.Prompt != "" {
		// Only command line prompt
		prompt = a.Prompt
	} else if stdinData != "" {
		// Only stdin data
		prompt = stdinData
	} else {
		// Neither provided
		return fmt.Errorf("no prompt provided (use argument or pipe data to stdin)")
	}

	// Convert Kong command to our command system
	exec := &command.ExecutionContext{
		Args:    []string{prompt},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}

	// Map flags
	if a.Model != "" {
		exec.Flags.Set("model", a.Model)
	}
	if len(a.Attach) > 0 {
		exec.Flags.Set("attach", a.Attach)
	}
	if a.Stream {
		exec.Flags.Set("stream", a.Stream)
	}
	if a.Temperature != 0 {
		exec.Flags.Set("temperature", a.Temperature)
	}
	if a.MaxTokens != 0 {
		exec.Flags.Set("max-tokens", a.MaxTokens)
	}
	if a.System != "" {
		exec.Flags.Set("system", a.System)
	}
	if a.ResponseFormat != "" {
		exec.Flags.Set("format", a.ResponseFormat)
	}
	// Use global output flag
	if ctx.CLI != nil && ctx.CLI.Output != "" {
		exec.Flags.Set("output", ctx.CLI.Output)
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
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}

	// Map flags
	if c.Resume != "" {
		exec.Flags.Set("resume", c.Resume)
	}
	if c.Model != "" {
		exec.Flags.Set("model", c.Model)
	}
	if len(c.Attach) > 0 {
		exec.Flags.Set("attach", c.Attach)
	}

	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "chat", exec)
}

// ConfigCmd handles the config command
type ConfigCmd struct {
	// Subcommands with brief descriptions
	Show     ConfigShowCmd     `cmd:"" help:"Show all configuration settings"`
	Get      ConfigGetCmd      `cmd:"" help:"Get a specific value"`
	Set      ConfigSetCmd      `cmd:"" help:"Set a configuration value"`
	Validate ConfigValidateCmd `cmd:"" help:"Validate configuration file"`
}

// ConfigShowCmd handles config show
type ConfigShowCmd struct{}

func (c *ConfigShowCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"list"},
		Flags:   command.NewFlags(nil),
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
		Flags:   command.NewFlags(nil),
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
		Flags:   command.NewFlags(nil),
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
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "config", exec)
}

// ModelCmd handles the model command
type ModelCmd struct {
	List   ModelListCmd   `cmd:"" help:"List available models"`
	Info   ModelInfoCmd   `cmd:"" help:"Show model information"`
	Select ModelSelectCmd `cmd:"" help:"Select default model"`
}

// ModelListCmd handles model list
type ModelListCmd struct {
	Provider   string `help:"Filter by provider"`
	Capability string `help:"Filter by capability"`
}

func (m *ModelListCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"list"},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	if m.Provider != "" {
		exec.Flags.Set("provider", m.Provider)
	}
	if m.Capability != "" {
		exec.Flags.Set("capability", m.Capability)
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "model", exec)
}

// ModelInfoCmd handles model info
type ModelInfoCmd struct {
	Model string `arg:"" required:"" help:"Model to show info for"`
}

func (m *ModelInfoCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"info", m.Model},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "model", exec)
}

// ModelSelectCmd handles model select
type ModelSelectCmd struct {
	Model string `arg:"" required:"" help:"Model to select"`
}

func (m *ModelSelectCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"select", m.Model},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "model", exec)
}

// ProfileCmd handles the profile command
type ProfileCmd struct {
	List   ProfileListCmd   `cmd:"" help:"List all profiles"`
	Create ProfileCreateCmd `cmd:"" help:"Create a new profile"`
	Switch ProfileSwitchCmd `cmd:"" help:"Switch to a profile"`
	Show   ProfileShowCmd   `cmd:"" help:"Show profile details"`
	Update ProfileUpdateCmd `cmd:"" help:"Update a profile"`
	Delete ProfileDeleteCmd `cmd:"" help:"Delete a profile"`
}

// ProfileListCmd handles profile list
type ProfileListCmd struct{}

func (p *ProfileListCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"list"},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "profile", exec)
}

// ProfileCreateCmd handles profile create
type ProfileCreateCmd struct {
	Name        string `arg:"" required:"" help:"Profile name"`
	Provider    string `help:"Provider name"`
	Model       string `help:"Model name"`
	Description string `help:"Profile description"`
}

func (p *ProfileCreateCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"create", p.Name},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	if p.Provider != "" {
		exec.Flags.Set("provider", p.Provider)
	}
	if p.Model != "" {
		exec.Flags.Set("model", p.Model)
	}
	if p.Description != "" {
		exec.Flags.Set("description", p.Description)
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "profile", exec)
}

// ProfileSwitchCmd handles profile switch
type ProfileSwitchCmd struct {
	Name string `arg:"" required:"" help:"Profile to switch to"`
}

func (p *ProfileSwitchCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"switch", p.Name},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "profile", exec)
}

// ProfileShowCmd handles profile show
type ProfileShowCmd struct {
	Name string `arg:"" help:"Profile to show (default: current)"`
}

func (p *ProfileShowCmd) Run(ctx *Context) error {
	args := []string{"show"}
	if p.Name != "" {
		args = append(args, p.Name)
	}
	exec := &command.ExecutionContext{
		Args:    args,
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "profile", exec)
}

// ProfileUpdateCmd handles profile update
type ProfileUpdateCmd struct {
	Name   string   `arg:"" required:"" help:"Profile to update"`
	Config []string `arg:"" required:"" help:"Configuration values (key=value)"`
}

func (p *ProfileUpdateCmd) Run(ctx *Context) error {
	args := []string{"update", p.Name}
	args = append(args, p.Config...)
	exec := &command.ExecutionContext{
		Args:    args,
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "profile", exec)
}

// ProfileDeleteCmd handles profile delete
type ProfileDeleteCmd struct {
	Name string `arg:"" required:"" help:"Profile to delete"`
}

func (p *ProfileDeleteCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"delete", p.Name},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "profile", exec)
}

// AliasCmd handles the alias command
type AliasCmd struct {
	Add    AliasAddCmd    `cmd:"" help:"Add an alias"`
	Remove AliasRemoveCmd `cmd:"" help:"Remove an alias"`
	List   AliasListCmd   `cmd:"" help:"List all aliases"`
	Show   AliasShowCmd   `cmd:"" help:"Show alias details"`
}

// AliasAddCmd handles alias add
type AliasAddCmd struct {
	Name    string   `arg:"" required:"" help:"Alias name"`
	Command []string `arg:"" required:"" help:"Command to alias"`
	Scope   string   `help:"Alias scope (cli, repl, all)"`
}

func (a *AliasAddCmd) Run(ctx *Context) error {
	args := []string{"add", a.Name}
	args = append(args, a.Command...)
	exec := &command.ExecutionContext{
		Args:    args,
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	if a.Scope != "" {
		exec.Flags.Set("scope", a.Scope)
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "alias", exec)
}

// AliasRemoveCmd handles alias remove
type AliasRemoveCmd struct {
	Name string `arg:"" required:"" help:"Alias to remove"`
}

func (a *AliasRemoveCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"remove", a.Name},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "alias", exec)
}

// AliasListCmd handles alias list
type AliasListCmd struct {
	Scope string `help:"Filter by scope (cli, repl, all)"`
}

func (a *AliasListCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"list"},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	if a.Scope != "" {
		exec.Flags.Set("scope", a.Scope)
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "alias", exec)
}

// AliasShowCmd handles alias show
type AliasShowCmd struct {
	Name string `arg:"" required:"" help:"Alias to show"`
}

func (a *AliasShowCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"show", a.Name},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "alias", exec)
}

// Context provides runtime context for commands
// HistoryCmd handles the history command
type HistoryCmd struct {
	List   HistoryListCmd   `cmd:"" help:"List all sessions"`
	Show   HistoryShowCmd   `cmd:"" help:"Show session details"`
	Delete HistoryDeleteCmd `cmd:"" help:"Delete a session"`
	Export HistoryExportCmd `cmd:"" help:"Export a session"`
	Search HistorySearchCmd `cmd:"" help:"Search sessions by content"`
}

// HistoryListCmd lists all sessions
type HistoryListCmd struct{}

// Run executes the history list command
func (h *HistoryListCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"list"},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "history", exec)
}

// HistoryShowCmd shows session details
type HistoryShowCmd struct {
	SessionID string `arg:"" required:"" help:"Session ID to show"`
}

// Run executes the history show command
func (h *HistoryShowCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"show", h.SessionID},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "history", exec)
}

// HistoryDeleteCmd deletes a session
type HistoryDeleteCmd struct {
	SessionID string `arg:"" required:"" help:"Session ID to delete"`
}

// Run executes the history delete command
func (h *HistoryDeleteCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"delete", h.SessionID},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "history", exec)
}

// HistoryExportCmd exports a session
type HistoryExportCmd struct {
	SessionID string `arg:"" required:"" help:"Session ID to export"`
	Format    string `default:"json" enum:"json,markdown" help:"Export format"`
}

// Run executes the history export command
func (h *HistoryExportCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"export", h.SessionID},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	exec.Flags.Set("format", h.Format)
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "history", exec)
}

// HistorySearchCmd searches sessions
type HistorySearchCmd struct {
	Query string `arg:"" required:"" help:"Search query"`
}

// Run executes the history search command
func (h *HistorySearchCmd) Run(ctx *Context) error {
	exec := &command.ExecutionContext{
		Args:    []string{"search", h.Query},
		Flags:   command.NewFlags(nil),
		Stdout:  ctx.Stdout,
		Stderr:  ctx.Stderr,
		Context: ctx.Ctx,
	}
	return ctx.Registry.GetExecutor().Execute(ctx.Ctx, "history", exec)
}

type Context struct {
	*kong.Context
	Registry *command.Registry
	Config   *config.Config
	Logger   *logging.Logger
	Stdout   io.Writer
	Stderr   io.Writer
	Ctx      context.Context
	CLI      *CLI // Reference to global CLI options
}

func main() {
	// Create the CLI parser
	parser := kong.Must(&CLI{},
		kong.Name("magellai"),
		kong.Description("A command-line interface for interacting with Large Language Models"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:             true,
			Summary:             true,
			NoExpandSubcommands: true,
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
	logLevel := "info"
	if envLevel := os.Getenv("MAGELLAI_LOG_LEVEL"); envLevel != "" {
		logLevel = envLevel
	}
	
	logConfig := logging.LogConfig{
		Level:      logLevel,
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
	if cli.ProfileName != "" {
		if err := cfg.SetProfile(cli.ProfileName); err != nil {
			logger.Error("Failed to set profile", "profile", cli.ProfileName, "error", err)
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

	askCmd := core.NewAskCommand(cfg)
	if err := registry.Register(askCmd); err != nil {
		logger.Error("failed to register ask command", "error", err)
		os.Exit(1)
	}

	chatCmd := core.NewChatCommand(cfg)
	if err := registry.Register(chatCmd); err != nil {
		logger.Error("failed to register chat command", "error", err)
		os.Exit(1)
	}

	historyCmd := core.NewHistoryCommand()
	if err := registry.Register(historyCmd); err != nil {
		logger.Error("failed to register history command", "error", err)
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
		CLI:      &cli,
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
