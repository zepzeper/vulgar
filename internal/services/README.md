# Service Layer Architecture

This directory contains shared service clients that provide API logic consumed by both CLI commands (`cmd/vulgar/discover/`) and Lua modules (`internal/modules/integrations/`).

## Purpose

The service layer creates a **single source of truth** for API integrations. Instead of duplicating HTTP client logic in both CLI and Lua modules, the service layer:

1. Centralizes API logic in reusable Go packages
2. Ensures consistent behavior between CLI and Lua workflows
3. Simplifies testing (test service layer independently)
4. Makes adding new integrations faster

## Architecture

```
                 ┌─────────────────┐    ┌─────────────────┐
                 │  CLI Commands   │    │  Lua Modules    │
                 │  (cmd/vulgar/)  │    │  (integrations/)│
                 └────────┬────────┘    └────────┬────────┘
                          │                      │
                          └──────────┬───────────┘
                                     │
                          ┌──────────▼──────────┐
                          │   Service Layer     │
                          │ (internal/services/)│
                          └──────────┬──────────┘
                                     │
                 ┌───────────────────┼───────────────────┐
                 │                   │                   │
         ┌───────▼───────┐   ┌───────▼───────┐   ┌───────▼───────┐
         │    config     │   │  httpclient   │   │  External API │
         │ (config.toml) │   │               │   │               │
         └───────────────┘   └───────────────┘   └───────────────┘
```

## Service Structure

Each service follows this pattern:

```
internal/services/<name>/
├── client.go    # API client with methods
├── types.go     # Request/response types
└── options.go   # Query options (optional)
```

### client.go

```go
package <name>

// Client wraps the HTTP client with service-specific logic
type Client struct {
    http *httpclient.Client
    // service-specific fields
}

// ClientOptions for explicit configuration
type ClientOptions struct {
    Token string
    URL   string
    // other options
}

// NewClient creates client with explicit options
func NewClient(opts ClientOptions) (*Client, error) { ... }

// NewClientFromConfig creates client from config.toml
func NewClientFromConfig() (*Client, error) { ... }

// API methods return typed responses
func (c *Client) ListFoo(ctx context.Context, opts ListOptions) ([]Foo, error) { ... }
```

### types.go

```go
package <name>

// Define shared types used by both CLI and Lua modules
type Foo struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    // ...
}

// Request types for create/update operations
type CreateFooRequest struct {
    Name string `json:"name"`
    // ...
}
```

## Consuming Services

### From CLI Commands

```go
// cmd/vulgar/discover/<name>/<name>.go
package <name>

import (
    "context"
    "github.com/zepzeper/vulgar/internal/services/<name>"
)

func listCmd() *cobra.Command {
    return &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // Create client from config
            client, err := <name>.NewClientFromConfig()
            if err != nil {
                return err
            }
            
            // Use service methods
            items, err := client.ListFoo(context.Background(), opts)
            if err != nil {
                return err
            }
            
            // CLI-specific: format and print
            printTable(items)
            return nil
        },
    }
}
```

### From Lua Modules

```go
// internal/modules/integrations/<name>/<name>_module.go
package <name>

import (
    lua "github.com/yuin/gopher-lua"
    svc "<name>service"
)

// Wrap service client
type luaClient struct {
    svc *svc.Client
}

func luaListFoo(L *lua.LState) int {
    client := checkClient(L, 1)
    
    // Use service methods
    items, err := client.svc.ListFoo(context.Background(), opts)
    if err != nil {
        return util.PushError(L, err.Error())
    }
    
    // Lua-specific: convert to Lua tables
    return util.PushSuccess(L, itemsToLua(L, items))
}
```

## Available Services

| Service | Path | CLI Commands | Lua Module |
|---------|------|--------------|------------|
| GitLab | `services/gitlab/` | `vulgar gitlab` | `integrations.gitlab` |
| GitHub | `services/github/` | `vulgar github` | `integrations.github` |
| Slack | `services/slack/` | `vulgar slack` | `integrations.slack` |
| Codeberg | `services/codeberg/` | `vulgar codeberg` | `integrations.codeberg` |
| Google Drive | `services/google/drive/` | `vulgar gdrive` | `integrations.gdrive` |
| Google Sheets | `services/google/sheets/` | `vulgar gsheets` | `integrations.gsheets` |
| Google Calendar | `services/google/calendar/` | `vulgar gcalendar` | `integrations.gcalendar` |

## Adding a New Service

1. **Create service directory**: `internal/services/<name>/`

2. **Define types** in `types.go`:
   - Response types with JSON tags
   - Request types for mutations

3. **Implement client** in `client.go`:
   - `ClientOptions` struct
   - `NewClient(opts)` constructor
   - `NewClientFromConfig()` config-based constructor
   - API methods returning typed responses

4. **Add config support** in `internal/config/config.go`:
   - Add config struct fields
   - Add getter functions

5. **Update CLI** in `cmd/vulgar/discover/<name>/`:
   - Use service client instead of raw httpclient
   - Convert service types to display types

6. **Update Lua module** in `internal/modules/integrations/<name>/`:
   - Wrap service client in userdata
   - Convert service types to Lua tables
   - Register module in `all/all.go`

## Configuration

All services read from `~/.config/vulgar/config.toml`:

```toml
[gitlab]
token = "glpat-..."
url = "https://gitlab.com"
projects = ["group/project"]

[github]
token = "ghp_..."
default_owner = "username"

[slack]
token = "xoxb-..."
default_channel = "#general"

[codeberg]
token = "..."
url = "https://codeberg.org"

[google]
credentials_file = "/path/to/credentials.json"
```

## Testing

Test service layer independently:

```bash
go test ./internal/services/...
```

Test Lua modules:

```bash
go test ./internal/modules/integrations/...
```
