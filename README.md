Vulgar is a powerful, modular workflow automation engine that combines Go's performance with Lua's scripting flexibility. With 60+ built-in modules covering cloud services, databases, AI, and developer tools, Vulgar lets you build automation scripts rapidly while maintaining full control.

## Key Capabilities

- **Script-First Automation**: Write workflows in Lua with zero boilerplate, powered by Go's performance
- **60+ Built-in Modules**: HTTP, databases, cloud storage, git platforms, notifications, and more
- **AI-Ready**: Integrations for OpenAI, Anthropic, Ollama, and HuggingFace (coming soon)
- **Google Workspace Native**: First-class support for Sheets, Drive, and Calendar
- **Interactive REPL**: Develop and test workflows interactively with command history and module preloading
- **Event-Driven**: Built-in support for cron jobs, file watchers, timers, and webhooks

## Quick Start

### Installation

Build from source (requires Go 1.21+):

```bash
git clone https://github.com/zepzeper/vulgar
cd vulgar
make build
make install  # Installs to ~/go/bin/
```

### Run Your First Workflow

Create `hello.lua`:
```lua
local http = require("http")
local json = require("json")

-- Fetch data from an API
local resp, err = http.get("https://api.github.com/zen")
if err then
    log.error("Request failed", { error = err })
    return
end

log.info("GitHub says: " .. resp.body)
```

Run it:
```bash
vulgar hello.lua
```

### Interactive REPL

Explore and experiment with modules interactively:

```bash
vulgar repl --preload
```

```lua
lua> client = gsheets.configure()
lua> sheets = client:list()
lua> for _, s in ipairs(sheets) do print(s.name) end
```

## Modules

Vulgar ships with a comprehensive module library organized into four categories:

### Core Modules
Essential utilities available in every workflow:

| Module | Description |
|--------|-------------|
| `http` | HTTP client with full request/response control |
| `json` | JSON encoding and decoding |
| `log` | Structured logging with levels (DEBUG, INFO, WARN, ERROR) |
| `fs` | File system operations |
| `path` | Path manipulation utilities |
| `time` | Time and duration handling |
| `env` | Environment variable access |
| `crypto` | Hashing and encryption |
| `uuid` | UUID generation |

### Standard Library
Extended functionality for common automation tasks:

| Module | Description |
|--------|-------------|
| `stdlib.cron` | Schedule recurring tasks |
| `stdlib.timer` | Delayed and periodic execution |
| `stdlib.shell` | Shell command execution |
| `stdlib.process` | Process management |
| `stdlib.filewatch` | File system event monitoring |
| `stdlib.yaml` / `xml` / `csv` | Data format parsing |
| `stdlib.regex` | Regular expressions |
| `stdlib.template` | Text templating |
| `stdlib.validator` | Input validation |
| `stdlib.retry` | Retry logic with backoff |
| `stdlib.workflow` | Workflow orchestration |
| `stdlib.compress.*` | Gzip, tar, and zip support |

### Integrations
Connect to external services and platforms:

| Category | Modules |
|----------|---------|
| **Google Workspace** | `integrations.gsheets`, `integrations.gdrive`, `integrations.gcalendar` |
| **Git Platforms** | `integrations.github`, `integrations.gitlab`, `integrations.codeberg` |
| **Notifications** | `integrations.slack`, `integrations.smtp` |
| **Infrastructure** | `integrations.ssh`, `integrations.docker`, `integrations.k8s` |
| **Databases** | `integrations.postgres`, `integrations.sqlite`, `integrations.mongodb`, `integrations.redis` |
| **Cloud Storage** | `integrations.s3` |
| **Messaging** | `integrations.kafka`, `integrations.rabbitmq`, `integrations.nats` |

### AI Modules *(Coming Soon)*
| Module | Description |
|--------|-------------|
| `ai.openai` | ChatGPT and GPT-4 integration |
| `ai.anthropic` | Claude API |
| `ai.ollama` | Local LLM inference |
| `ai.huggingface` | HuggingFace models |

List all available modules:
```bash
vulgar --list-modules
```

## Examples

### Google Sheets Automation

```lua
local gsheets = require("integrations.gsheets")

local client = gsheets.configure()
local sheet = client:open("My Spreadsheet")

-- Read data
local data = sheet:read("Sheet1!A1:C10")
for _, row in ipairs(data) do
    print(row[1], row[2], row[3])
end

-- Write data
sheet:write("Sheet1!A1", {{"Hello", "World"}})
```

### Scheduled Tasks with Cron

```lua
local cron = require("stdlib.cron")
local slack = require("integrations.slack")

cron.schedule("0 9 * * 1-5", function()
    slack.send("#general", "Good morning, team! ☀️")
end)

cron.run()  -- Start the scheduler
```

### HTTP API with Retry Logic

```lua
local http = require("http")
local json = require("json")
local retry = require("stdlib.retry")

local function fetch_data()
    local resp, err = http.get("https://api.example.com/data")
    if err then return nil, err end
    return json.decode(resp.body)
end

local data, err = retry.call(fetch_data, {
    max_attempts = 3,
    delay = "1s",
    backoff = "exponential"
})
```

### File Watching

```lua
local filewatch = require("stdlib.filewatch")
local shell = require("stdlib.shell")

filewatch.watch("./src", function(event)
    if event.name:match("%.lua$") then
        log.info("File changed: " .. event.name)
        shell.exec("make test")
    end
end)
```

## CLI Reference

```
vulgar [script] [args...]

Flags:
  -e, --eval string      Execute Lua code directly
  -l, --log-level string Log level: DEBUG, INFO, WARN, ERROR (default "INFO")
  -v, --verbose          Enable verbose logging (DEBUG level)
  -t, --timeout string   Execution timeout (e.g., 30s, 5m, 1h)
  -c, --check            Syntax check only, do not execute
      --dry-run          Parse and validate without side effects
      --list-modules     List all available modules
      --profile          Enable CPU profiling
      --trace            Enable execution tracing
      --version          Show version information

Commands:
  repl        Interactive Lua REPL
  init        Initialize authentication for integrations
  ui          Launch the terminal UI
```

## Development

```bash
# Build
make build

# Run tests
make test

# Format code
make fmt

# Lint
make lint

# Build for all platforms
make release-all
```

## Resources

- 📚 [Module Documentation](docs/) *(coming soon)*
- 💡 [Example Workflows](examples/) *(coming soon)*
- 🐛 [Issue Tracker](https://github.com/zepzeper/vulgar/issues)

## Support

Found a bug or have a feature request? Open an issue on GitHub!

## License

Vulgar is open source software licensed under the [Apache License 2.0](LICENSE).

---

## What does "Vulgar" mean?

The namae comes from the ide of making automation *accessible to everyone* - like "vulgar" in its original Latin sense meaning "of the common people." Just as Lua means "moon" in Portuguese and illuminates the darkness, Vulgar aims to make powerful automation tools available to all developers, not just those with specialized knowledge.
