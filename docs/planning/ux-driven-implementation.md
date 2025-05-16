# Modular Library-First Design for **magellai**

This revision treats the *core intelligence*—provider APIs, prompt orchestration, tool execution, agent state machines, workflow runner, session storage—as a reusable Go module.
Two thin front-ends, **CLI** and **REPL**, simply wire user I/O into that library. Everything else (plugins, flows, config, completions) adapts to the new separation.

---



## 1 . Guiding Principles

| Goal                              | Design Choice                                                                                                        |
| --------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| **Keep the core tiny**            | Ship *only* “ask”, “chat”, and basic “config/history” in the root binary. Everything else is a plugin.               |
| **Play well with shells**         | Pure STDIN/STDOUT, no hidden interactive prompts when input is non-TTY. Unix pipes first; PowerShell tested.         |
| **Grow like Git/kubectl**         | Discover extra functionality by executable-name convention (`magellai-xxx`). No code changes needed to add commands. |
| **Single source of truth for UX** | Unified help tree, one completion script, one flag style, one config file format.                                    |


| Objective                      | Library-centric Decision                                                                                                                                                |
| ------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Isolate domain logic**       | Put all LLM, tool, agent, flow, session, and config code in `github.com/lexlapax/magellai` (importable by other Go apps).                                               |
| **Keep surfaces small**        | Library exposes a handful of composable services (e.g., `LLMClient`, `ToolExecutor`, `Conversation`, `WorkflowRunner`). Front-ends depend **only** on those interfaces. |
| **Swap UIs freely**            | CLI, REPL, HTTP server, or GUI can coexist, each built on the same API and sharing plugins/config/history.                                                              |
| **Enable external automation** | Third-party Go programs can call `magellai.Ask()` or embed a `Conversation` without spawning a subprocess.                                                              |
| **Unbreakable plugin story**   | Binary plugins still work; they call the library via JSON-RPC or import it directly if written in Go.                                                                   |

---

## 2 . Achitecture & Command Topology (Stable for Users)
### architecture
```text
┌──────────────────────────────┐
│           Front-Ends         │
│  CLI cmd/magellai            │
│  REPL pkg/repl               │
│  (future) HTTP, gRPC, TUI    │
└──────────────┬───────────────┘
               │ clean interfaces
┌──────────────┴───────────────┐
│        Core Library          │
│  llm      → provider drivers │
│  tool     → exec / RPC hooks │
│  agent    → multi-step flows │
│  workflow → YAML / code DAGs │
│  session  → storage & recall │
│  config   → profiles & env   │
└──────────────┬───────────────┘
               │ plug-in boundary
┌──────────────┴───────────────┐
│      External Extensions     │
│  magellai-tool-*, etc.       │
│  Go plugins (optional)       │
└──────────────────────────────┘
```

*Front-ends* import the library and wire user input/output.
*Plugins* either shell out (`exec.Command`) or invoke exported Go APIs if they share the same build.

### command topology

```text
magellai <primary> [sub] [flags] [args]

PRIMARY ACTIONS
  ask       one-shot query  (scriptable)
  chat      interactive REPL
  tool      run <tool>…            (plugin)
  agent     run <agent>…           (plugin)
  flow      run <flow>…            (plugin)

HOUSEKEEPING
  config    set / get / edit YAML
  history   list / show / rm
  plugin    list / install / update / remove
  help      any-level help
```

* **Reserved namespace**: nothing else becomes a primary verb.
* Everything under `tool|agent|flow` is delegated to a plugin binary (`magellai-tool-<name>`, etc.).
* Sub-verbs inside those namespaces (`list`, `install`, etc.) are handled by the core so they’re always available.

---

## 3 . Flag Scheme (Simple & Predictable)



| Scope             | Examples                                                       | Notes                                                        |
| ----------------- | -------------------------------------------------------------- | ------------------------------------------------------------ |
| **Global**        | `-v/--verbosity`, `--output`, `--no-color`, `--profile=<name>` | Acceptable anywhere; parsed first.                           |
| **Generation**    | `-m/--model`, `--temperature`, `--max-tokens`                  | Shared by *ask, chat, agent, flow* so scripts stay portable. |
| **I/O**           | `-a/--attach <file>` (repeatable), `--stream/--no-stream`      | Works identically across modes.                              |
| **Command-local** | e.g. `--depth` for a specific flow                             | Namespaced in plugin help; no collisions.                    |

*Short flags only for the top 3 most-used options (`-m`, `-o`, `-v`).*
Anything else is long-form for clarity.

### Command Surface (unchanged for users, simplified inside)

```text
magellai <verb> [sub] [flags] [args]

ask | chat | tool | agent | flow | config | history | plugin | help
```

CLI still discovers plugins via `magellai-<kind>-<name>`; the wrapper marshals CLI flags into the library’s structs and prints results.


### Flag Strategy

Front-ends register one **flag set** that fills a `Request` or `Conversation` struct:

| Category   | Examples                                             | Routed To                      |
| ---------- | ---------------------------------------------------- | ------------------------------ |
| Global     | `--verbosity`, `--output`, `--no-color`, `--profile` | `config.Resolve()`             |
| LLM        | `--model`, `--temperature`, `--max-tokens`           | `PromptParams`                 |
| I/O        | `--stream`, `--attach/-a`                            | Attachment slice & stream bool |
| Verb-local | flow-specific like `--depth`                         | `FlowOpts` etc.                |

Because parsing is front-end, the library remains flag-free and testable.

---

## 4 . Plugin Architecture (Power with Restraint)

###  Public Library API Sketch

```go
// High-level one-shot
func Ask(ctx context.Context, req Request) (Response, error)

// Stateful conversation
type Conversation struct {
    ID string
    PromptConfig PromptParams
    Stream(bool) // toggles streaming callbacks
    Send(ctx context.Context, userMsg string, att []Attachment) (Response, error)
}

// Multi-step agent
func RunAgent(ctx context.Context, name string, args []string, opts AgentOpts) error

// Workflow by name or file
func RunFlow(ctx context.Context, flowID string, opts FlowOpts) error
```

Every front-end sticks to these calls; no duplicative LLM code lives in the UI layers.
---
1. **Name-convention execs**
   *Discovery*: at runtime scan `$PATH` and `~/.magellai/plugins`
   *Mapping*:

   ```
   magellai tool calculator  → exec magellai-tool-calculator
   magellai agent researcher → exec magellai-agent-researcher
   ```
2. **Thin gRPC or JSON-over-stdin/stdout contract**

   * Core sends a JSON envelope: `{ "args": [...], "stdin": "...", "env": {…} }`
   * Plugin replies with streamed JSON `{ "event":"chunk", "data":"..." }` so the core can forward/format.
3. **Optional Go-plugin interface** *(future)*

   * Keep core interface stable (`Plugin.Run(ctx, io.Reader, io.Writer)`).
   * When Go’s `plugin` pkg is available (Linux/macOS), you can dlopen shared objects for performance.

Benefits: language-agnostic today, tighter integration tomorrow.

---

## 5 . REPL Design (Chat Mode)

* **Prompt**: `User ›` / `AI ›` with ANSI colors when TTY.
* **Slash commands** (never sent to LLM):

  ```
  /help, /model gpt-4, /save, /load <id>, /reset, /exit
  ```
* **History persistence**: each session ID and metadata stored in `~/.magellai/sessions/*.json`.
* **Attachments inside chat**: `:attach file.pdf` (adds to context; identical to `-a` in CLI).


### REPL Front-End

*Package `repl`* consumes the library:

```go
conv := magellai.NewConversation(profile)
// loop over user lines
resp, _ := conv.Send(ctx, line, currentAttachments)
ui.Render(resp)
```

Slash commands mutate the `Conversation` or call helper library functions (`session.Save(conv)`).

---
### Plugin Contract

### Exec-style (language-agnostic)

```
$ magellai tool calculator 2+2
└─ CLI passes JSON to magellai-tool-calculator stdin
└─ Plugin replies with JSON events
└─ CLI decodes → library.StreamRenderer
```

### Go-native (best performance)

```go
// plugins/go/calculator/main.go
func init() {
    magellai.RegisterTool("calculator", func(ctx context.Context, in io.Reader, out io.Writer, args []string) error {
        // call library if desired
    })
}
```

The plugin binary imports the same core module, guaranteeing API parity.

## 6 . Configuration
## 8 . Configuration



```yaml
profiles:
  personal:
    provider: openai
    model: gpt-4o
    defaults:
      temperature: 0.7
history_dir: ~/.magellai/sessions
plugin_dir:  ~/.magellai/plugins
```



---

`pkg/config` centralizes:

```yaml
# ~/.config/magellai/config.yaml
default_profile: personal
profiles:
  personal:
    provider: openai
    model: gpt-4o
    temperature: 0.7
  work:
    provider: anthropic
    model: claude-3-sonnet
```
CLIs use `config.Load(profile)`; REPL uses the same.

Users edit via:

```bash
magellai config set profile.work.model gemini-1.5-pro
magellai config edit    # opens $EDITOR
```

Environment variable override: `MAGELLAI_MODEL`, `MAGELLAI_API_KEY`, etc.

---

## 7 . Project Layout (Go)

```
magellai/
  cmd/
    magellai/          → Cobra flags → core APIs
  pkg/
    llm/               → drivers (openai, anthropic, gemini…)
    tool/              → registry, exec bridge
    agent/             → state machines
    workflow/          → YAML parser, DAG engine
    session/           → save/load JSON
    config/            → Viper wrapper
    repl/              → interactive loop (imports core)
  plugins/             → sample external binaries
```

Everything under **pkg/** is buildable as a standalone library (`go get github.com/lexlapax/magellai/pkg/...`).


* **Use Cobra** for hierarchical commands + completions.
* **Use Viper** for YAML config with profiles.
* **Keep provider implementations in their own Go files** (not plugins) so the core can fall back to them even if no external plugin is installed.

---

## 8 . Implementation Roadmap


| Week | Deliverable                                  |                                    |
| ---- | -------------------------------------------- | ---------------------------------- |
| 1    | Core data models, OpenAI driver, `Ask()`     |                                    |
| 2    | Conversation API, REPL wrapper               |                                    |
| 3    | Cobra CLI wrapping core, YAML config loader  |                                    |
| 4    | Exec plugin runner, sample `calculator` tool |                                    |
| 5    | Agent & workflow engine                      |                                    |
| 6    | Plugin manager (\`plugin install             | list\`), docs & completion scripts |

Unit tests run **only** on the library; CLI + REPL get shallow integration tests.

---

### Payoff

* **Re-usability** – other Go apps (\~microservices, Slack bots, VS Code extensions) call the same package.
* **Consistency** – one canonical implementation of providers, tools, and workflows.
* **Maintainability** – UI tweaks never touch LLM logic; new providers never touch UI code.
* **Community growth** – plugin authors write Go or any language, confident that the library API is stable.

### Why This Hits the Sweet Spot

* **Simple on day 1**: only two built-ins (`ask`, `chat`), zero code needed to add a new tool.
* **Discoverable**: uniform help, autocompletion, clear flag grouping.
* **Extensible forever**: external binary model scales like Git; later you can offer a tighter Go-plugin path without breaking users.
* **Portable**: plain stdin/stdout contracts + YAML config keep it cross-platform.
* **Maintainable**: narrow, layered Go modules mean each concern (CLI, LLM provider, plugin runner, REPL) evolves independently.

Adopt this skeleton, and you can deliver a delightful developer UX now while cultivating a healthy plugin ecosystem that lets *others* expand Magellai’s horizons for you.
A library-first magellai is a platform, not just a terminal toy.

