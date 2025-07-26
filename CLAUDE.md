# LazyOC - Claude Code Project Memory

## Project Overview
**LazyOC** is a lazy terminal UI for OpenShift and Kubernetes clusters built in Go. It provides an intuitive, vim-like interface for managing cluster resources without memorizing kubectl/oc commands.

- **Language**: Go 1.24.5
- **Main Framework**: Bubble Tea (TUI)  
- **Target**: OpenShift/Kubernetes cluster management
- **Version**: 0.1.0 (from VERSION file)
- **Status**: Full OpenShift support implemented with navigation and details panels

## Essential Build & Development Commands

### Building & Running
```bash
# ALWAYS build to correct location
go build -o ./bin/lazyoc ./cmd/lazyoc

# Run the built binary
./bin/lazyoc

# Development mode (run without building)
go run ./cmd/lazyoc

# Use Makefile for standardized builds
make build          # Builds to bin/lazyoc
make dev           # Runs in development mode
make run           # Build then run
```

**CRITICAL**: Never use `go build ./cmd/lazyoc` as it creates binary in wrong location.

### Testing & Quality
```bash
# Run tests
go test ./...
make test

# Coverage
make test-coverage

# Linting (required before commits)
golangci-lint run
make lint

# Format code
go fmt ./...
make fmt

# All checks
make all           # clean fmt vet test build
```

## Project Architecture

### Directory Structure
```
lazyoc/
├── cmd/lazyoc/main.go           # Application entry point
├── internal/                    # Private application code
│   ├── ui/                     # Main TUI implementation (primary)
│   │   ├── components/         # UI components (header, panels, tabs, etc.)
│   │   ├── views/             # Resource-specific views (pods, logs, etc.)
│   │   ├── navigation/        # Keyboard/mouse handling, help system
│   │   ├── messages/          # Event handling system
│   │   ├── models/            # Bubble Tea models
│   │   └── styles/            # Theme and visual styling
│   ├── tui/                   # Alternative/experimental TUI (secondary)
│   ├── k8s/                   # Kubernetes client logic
│   │   ├── client.go          # Main K8s client
│   │   ├── auth/              # Authentication & kubeconfig
│   │   ├── resources/         # Resource operations (CRUD)
│   │   ├── projects/          # Namespace/project management
│   │   └── monitor/           # Resource monitoring & watching
│   ├── constants/             # Application constants
│   ├── errors/                # Error handling
│   └── logging/               # Logging utilities
├── pkg/                       # Public libraries (future)
├── api/                       # API definitions (future)
├── configs/                   # Configuration files
├── scripts/                   # Build and utility scripts
└── docs/                      # Architecture documentation
```

### Key Components

**Main Entry**: `cmd/lazyoc/main.go`
- Uses Cobra for CLI commands
- Flags: `--debug`, `--kubeconfig`, `--no-alt-screen`
- Calls `ui.RunTUI()` to start interface

**Primary TUI**: `internal/ui/tui.go` (Simplified Architecture)
- Built with Bubble Tea framework - single TUI implementation
- Program entry: `internal/ui/program.go`
- Main app model: `internal/ui/models/app.go`
- **Architecture Cleanup**: Removed dual TUI systems, now uses unified TUI implementation

**K8s/OpenShift Integration**: `internal/k8s/`
- Client wrapper around k8s.io/client-go and openshift/client-go
- Authentication via kubeconfig (`oc login` compatible)
- Full OpenShift resource support (BuildConfigs, ImageStreams, Routes)
- Resource operations and monitoring for both platforms

## Code Patterns & Conventions

### Error Handling
- Use `internal/errors/` package for structured errors
- Error recovery in `internal/ui/errors/recovery.go`
- Consistent error mapping in `internal/ui/errors/mapper.go`

### Event System
- Message-based architecture using Bubble Tea
- Messages defined in `internal/ui/messages/`
- Handlers in `internal/ui/messages/handler.go`

### Resource Management
- Interface-based design in `internal/k8s/interface.go`
- Resource operations in `internal/k8s/resources/`
- Retry logic for API calls

### UI Components
- Reusable components in `internal/ui/components/`
- View-specific logic in `internal/ui/views/`
- Consistent styling via `internal/ui/styles/`

## Authentication & Kubeconfig

LazyOC uses standard Kubernetes authentication:
- Default: `~/.kube/config`
- Custom: `--kubeconfig=/path/to/config`
- Supports all K8s auth methods (tokens, certs, OIDC, etc.)
- Works with `oc login` authentication flow

## Testing Strategy

### Test Files Location
- Unit tests: `*_test.go` files alongside source
- Integration tests: `internal/ui/integration_test.go`
- Auth tests: `internal/k8s/auth/auth_test.go`

### Test Commands
```bash
go test ./...                    # All tests
go test -race ./...             # Race condition detection
go test -v ./internal/k8s/...   # Specific package tests
```

## Development Workflow

### Before Implementing
1. Check existing patterns in relevant `internal/` directories
2. Follow interface-based design (see `internal/k8s/interface.go`)
3. Use constants from `internal/constants/`
4. Implement proper error handling with `internal/errors/`

### Code Style
- Use `go fmt` for formatting
- Follow Go naming conventions
- Use interfaces for testability
- Keep components small and focused
- Document public APIs

### Dependencies
- **TUI**: `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`
- **K8s**: `k8s.io/client-go`, `k8s.io/api`, `k8s.io/apimachinery`
- **OpenShift**: `github.com/openshift/client-go`, `github.com/openshift/api`
- **CLI**: `github.com/spf13/cobra`

## Common File Locations

### Entry Points
- Main: `cmd/lazyoc/main.go:19`
- TUI Start: `internal/ui/program.go` (RunTUI function)
- Main TUI: `internal/ui/tui.go` (unified implementation)
- App Model: `internal/ui/models/app.go`

### Key Implementations
- K8s/OpenShift Client: `internal/k8s/client.go`
- OpenShift Resources: `internal/k8s/resources/openshift_client.go`
- Resource Types: `internal/k8s/resources/types.go`
- TUI Navigation: `internal/ui/tui.go` (key handling around lines 390-490)
- Details Panels: `internal/ui/tui.go` (updatePodDetails, updateBuildConfigDetails, etc.)
- Tab Management: `internal/ui/models/app.go` (NextTab, PrevTab, ActiveTab)

### OpenShift Integration
- Resource Constants: `internal/constants/ui.go` (ResourceTabs with 8 tabs)
- Message Types: `internal/ui/messages/k8s_messages.go` (OpenShift-specific messages)
- Client Factory: `internal/k8s/client.go` (OpenShift detection and client initialization)

### Configuration
- Constants: `internal/constants/*.go`
- Version: `VERSION` file (currently 0.1.0)

## Git & Release

### Version Management
- Version stored in `VERSION` file
- Git tags: `v{VERSION}` (e.g., `v0.1.0`)
- Automated releases via GitHub Actions

### Commit Patterns
- Use conventional commits: `feat:`, `fix:`, `docs:`, etc.
- Reference issues: `fix: resolve auth timeout (#123)`
- Keep commits focused and atomic

## Performance Targets
- Memory: <100MB baseline
- CPU: <5% average, <10% peak  
- Startup: <2 seconds
- API Latency: <500ms for queries

## Troubleshooting

### Common Issues
- Build in wrong location: Use `go build -o ./bin/lazyoc ./cmd/lazyoc`
- Auth failures: Check `kubectl config current-context`
- TUI issues: Try `--debug` flag for logging to `lazyoc.log`

### Debug Mode
```bash
./bin/lazyoc --debug    # Enables logging to lazyoc.log
```

---

## Task Master AI Integration

### Essential Commands

#### Core Workflow Commands
```bash
# Project Setup
task-master init                                    # Initialize Task Master in current project
task-master parse-prd .taskmaster/docs/prd.txt      # Generate tasks from PRD document
task-master models --setup                        # Configure AI models interactively

# Daily Development Workflow
task-master list                                   # Show all tasks with status
task-master next                                   # Get next available task to work on
task-master show <id>                             # View detailed task information (e.g., task-master show 1.2)
task-master set-status --id=<id> --status=done    # Mark task complete

# Task Management
task-master add-task --prompt="description" --research        # Add new task with AI assistance
task-master expand --id=<id> --research --force              # Break task into subtasks
task-master update-task --id=<id> --prompt="changes"         # Update specific task
task-master update --from=<id> --prompt="changes"            # Update multiple tasks from ID onwards
task-master update-subtask --id=<id> --prompt="notes"        # Add implementation notes to subtask

# Analysis & Planning
task-master analyze-complexity --research          # Analyze task complexity
task-master complexity-report                      # View complexity analysis
task-master expand --all --research               # Expand all eligible tasks

# Dependencies & Organization
task-master add-dependency --id=<id> --depends-on=<id>       # Add task dependency
task-master move --from=<id> --to=<id>                       # Reorganize task hierarchy
task-master validate-dependencies                            # Check for dependency issues
task-master generate                                         # Update task markdown files (usually auto-called)
```

### Key Files & Project Structure

#### Core Files
- `.taskmaster/tasks/tasks.json` - Main task data file (auto-managed)
- `.taskmaster/config.json` - AI model configuration (use `task-master models` to modify)
- `.taskmaster/docs/prd.txt` - Product Requirements Document for parsing
- `.taskmaster/tasks/*.txt` - Individual task files (auto-generated from tasks.json)
- `.env` - API keys for CLI usage

#### Claude Code Integration Files
- `CLAUDE.md` - Auto-loaded context for Claude Code (this file)
- `.claude/settings.json` - Claude Code tool allowlist and preferences
- `.claude/commands/` - Custom slash commands for repeated workflows
- `.mcp.json` - MCP server configuration (project-specific)

#### Directory Structure
```
project/
├── .taskmaster/
│   ├── tasks/              # Task files directory
│   │   ├── tasks.json      # Main task database
│   │   ├── task-1.md      # Individual task files
│   │   └── task-2.md
│   ├── docs/              # Documentation directory
│   │   ├── prd.txt        # Product requirements
│   ├── reports/           # Analysis reports directory
│   │   └── task-complexity-report.json
│   ├── templates/         # Template files
│   │   └── example_prd.txt  # Example PRD template
│   └── config.json        # AI models & settings
├── .claude/
│   ├── settings.json      # Claude Code configuration
│   └── commands/         # Custom slash commands
├── .env                  # API keys
├── .mcp.json            # MCP configuration
└── CLAUDE.md            # This file - auto-loaded by Claude Code
```

### MCP Integration

Task Master provides an MCP server that Claude Code can connect to. Configure in `.mcp.json`:

```json
{
  "mcpServers": {
    "task-master-ai": {
      "command": "npx",
      "args": ["-y", "--package=task-master-ai", "task-master-ai"],
      "env": {
        "ANTHROPIC_API_KEY": "your_key_here",
        "PERPLEXITY_API_KEY": "your_key_here",
        "OPENAI_API_KEY": "OPENAI_API_KEY_HERE",
        "GOOGLE_API_KEY": "GOOGLE_API_KEY_HERE",
        "XAI_API_KEY": "XAI_API_KEY_HERE",
        "OPENROUTER_API_KEY": "OPENROUTER_API_KEY_HERE",
        "MISTRAL_API_KEY": "MISTRAL_API_KEY_HERE",
        "AZURE_OPENAI_API_KEY": "AZURE_OPENAI_API_KEY_HERE",
        "OLLAMA_API_KEY": "OLLAMA_API_KEY_HERE"
      }
    }
  }
}
```

#### Essential MCP Tools
```javascript
help; // = shows available taskmaster commands
// Project setup
initialize_project; // = task-master init
parse_prd; // = task-master parse-prd

// Daily workflow
get_tasks; // = task-master list
next_task; // = task-master next
get_task; // = task-master show <id>
set_task_status; // = task-master set-status

// Task management
add_task; // = task-master add-task
expand_task; // = task-master expand
update_task; // = task-master update-task
update_subtask; // = task-master update-subtask
update; // = task-master update

// Analysis
analyze_project_complexity; // = task-master analyze-complexity
complexity_report; // = task-master complexity-report
```

### Claude Code Workflow Integration

#### Standard Development Workflow

##### 1. Project Initialization
```bash
# Initialize Task Master
task-master init

# Create or obtain PRD, then parse it
task-master parse-prd .taskmaster/docs/prd.txt

# Analyze complexity and expand tasks
task-master analyze-complexity --research
task-master expand --all --research
```

If tasks already exist, another PRD can be parsed (with new information only!) using parse-prd with --append flag. This will add the generated tasks to the existing list of tasks.

##### 2. Daily Development Loop
```bash
# Start each session
task-master next                           # Find next available task
task-master show <id>                     # Review task details

# During implementation, check in code context into the tasks and subtasks
task-master update-subtask --id=<id> --prompt="implementation notes..."

# Complete tasks
task-master set-status --id=<id> --status=done
```

##### 3. Multi-Claude Workflows
For complex projects, use multiple Claude Code sessions:

```bash
# Terminal 1: Main implementation
cd project && claude

# Terminal 2: Testing and validation
cd project-test-worktree && claude

# Terminal 3: Documentation updates
cd project-docs-worktree && claude
```

#### Custom Slash Commands

Create `.claude/commands/taskmaster-next.md`:
```markdown
Find the next available Task Master task and show its details.

Steps:

1. Run `task-master next` to get the next task
2. If a task is available, run `task-master show <id>` for full details
3. Provide a summary of what needs to be implemented
4. Suggest the first implementation step
```

Create `.claude/commands/taskmaster-complete.md`:
```markdown
Complete a Task Master task: $ARGUMENTS

Steps:

1. Review the current task with `task-master show $ARGUMENTS`
2. Verify all implementation is complete
3. Run any tests related to this task
4. Mark as complete: `task-master set-status --id=$ARGUMENTS --status=done`
5. Show the next available task with `task-master next`
```

#### Tool Allowlist Recommendations

Add to `.claude/settings.json`:
```json
{
  "allowedTools": [
    "Edit",
    "Bash(task-master *)",
    "Bash(git commit:*)",
    "Bash(git add:*)",
    "Bash(npm run *)",
    "mcp__task_master_ai__*"
  ]
}
```

### Configuration & Setup

#### API Keys Required
At least **one** of these API keys must be configured:

- `ANTHROPIC_API_KEY` (Claude models) - **Recommended**
- `PERPLEXITY_API_KEY` (Research features) - **Highly recommended**
- `OPENAI_API_KEY` (GPT models)
- `GOOGLE_API_KEY` (Gemini models)
- `MISTRAL_API_KEY` (Mistral models)
- `OPENROUTER_API_KEY` (Multiple models)
- `XAI_API_KEY` (Grok models)

An API key is required for any provider used across any of the 3 roles defined in the `models` command.

#### Model Configuration
```bash
# Interactive setup (recommended)
task-master models --setup

# Set specific models
task-master models --set-main claude-3-5-sonnet-20241022
task-master models --set-research perplexity-llama-3.1-sonar-large-128k-online
task-master models --set-fallback gpt-4o-mini
```

### Task Structure & IDs

#### Task ID Format
- Main tasks: `1`, `2`, `3`, etc.
- Subtasks: `1.1`, `1.2`, `2.1`, etc.
- Sub-subtasks: `1.1.1`, `1.1.2`, etc.

#### Task Status Values
- `pending` - Ready to work on
- `in-progress` - Currently being worked on
- `done` - Completed and verified
- `deferred` - Postponed
- `cancelled` - No longer needed
- `blocked` - Waiting on external factors

#### Task Fields
```json
{
  "id": "1.2",
  "title": "Implement user authentication",
  "description": "Set up JWT-based auth system",
  "status": "pending",
  "priority": "high",
  "dependencies": ["1.1"],
  "details": "Use bcrypt for hashing, JWT for tokens...",
  "testStrategy": "Unit tests for auth functions, integration tests for login flow",
  "subtasks": []
}
```

### Claude Code Best Practices with Task Master

#### Context Management
- Use `/clear` between different tasks to maintain focus
- This CLAUDE.md file is automatically loaded for context
- Use `task-master show <id>` to pull specific task context when needed

#### Iterative Implementation
1. `task-master show <subtask-id>` - Understand requirements
2. Explore codebase and plan implementation
3. `task-master update-subtask --id=<id> --prompt="detailed plan"` - Log plan
4. `task-master set-status --id=<id> --status=in-progress` - Start work
5. Implement code following logged plan
6. `task-master update-subtask --id=<id> --prompt="what worked/didn't work"` - Log progress
7. `task-master set-status --id=<id> --status=done` - Complete task

#### Complex Workflows with Checklists
For large migrations or multi-step processes:

1. Create a markdown PRD file describing the new changes: `touch task-migration-checklist.md` (prds can be .txt or .md)
2. Use Taskmaster to parse the new prd with `task-master parse-prd --append` (also available in MCP)
3. Use Taskmaster to expand the newly generated tasks into subtasks. Consider using `analyze-complexity` with the correct --to and --from IDs (the new ids) to identify the ideal subtask amounts for each task. Then expand them.
4. Work through items systematically, checking them off as completed
5. Use `task-master update-subtask` to log progress on each task/subtask and/or updating/researching them before/during implementation if getting stuck

#### Git Integration
Task Master works well with `gh` CLI:

```bash
# Create PR for completed task
gh pr create --title "Complete task 1.2: User authentication" --body "Implements JWT auth system as specified in task 1.2"

# Reference task in commits
git commit -m "feat: implement JWT auth (task 1.2)"
```

#### Parallel Development with Git Worktrees
```bash
# Create worktrees for parallel task development
git worktree add ../project-auth feature/auth-system
git worktree add ../project-api feature/api-refactor

# Run Claude Code in each worktree
cd ../project-auth && claude    # Terminal 1: Auth work
cd ../project-api && claude     # Terminal 2: API work
```

### Troubleshooting

#### AI Commands Failing
```bash
# Check API keys are configured
cat .env                           # For CLI usage

# Verify model configuration
task-master models

# Test with different model
task-master models --set-fallback gpt-4o-mini
```

#### MCP Connection Issues
- Check `.mcp.json` configuration
- Verify Node.js installation
- Use `--mcp-debug` flag when starting Claude Code
- Use CLI as fallback if MCP unavailable

#### Task File Sync Issues
```bash
# Regenerate task files from tasks.json
task-master generate

# Fix dependency issues
task-master fix-dependencies
```

DO NOT RE-INITIALIZE. That will not do anything beyond re-adding the same Taskmaster core files.

### Important Notes

#### AI-Powered Operations
These commands make AI calls and may take up to a minute:

- `parse_prd` / `task-master parse-prd`
- `analyze_project_complexity` / `task-master analyze-complexity`
- `expand_task` / `task-master expand`
- `expand_all` / `task-master expand --all`
- `add_task` / `task-master add-task`
- `update` / `task-master update`
- `update_task` / `task-master update-task`
- `update_subtask` / `task-master update-subtask`

#### File Management
- Never manually edit `tasks.json` - use commands instead
- Never manually edit `.taskmaster/config.json` - use `task-master models`
- Task markdown files in `tasks/` are auto-generated
- Run `task-master generate` after manual changes to tasks.json

#### Claude Code Session Management
- Use `/clear` frequently to maintain focused context
- Create custom slash commands for repeated Task Master workflows
- Configure tool allowlist to streamline permissions
- Use headless mode for automation: `claude -p "task-master next"`

#### Multi-Task Updates
- Use `update --from=<id>` to update multiple future tasks
- Use `update-task --id=<id>` for single task updates
- Use `update-subtask --id=<id>` for implementation logging

#### Research Mode
- Add `--research` flag for research-based AI enhancement
- Requires a research model API key like Perplexity (`PERPLEXITY_API_KEY`) in environment
- Provides more informed task creation and updates
- Recommended for complex technical tasks

**Note**: Task Master commands are optional workflow tools - core LazyOC development can proceed without them.

---

## Recent Major Implementation (January 2025)

### OpenShift Integration Complete ✅
**Status**: Task 7 completed with commit `bb2060c`

#### Architecture Cleanup
- **Problem**: LazyOC had two competing TUI implementations causing confusion
- **Solution**: Removed complex component-based TUI, unified on `SimplifiedTUI` (renamed to `TUI`)
- **Result**: Single, clean TUI architecture in `internal/ui/tui.go`

#### OpenShift Resource Support
**Added full support for OpenShift-specific resources:**

1. **Resource Types** (`internal/k8s/resources/types.go`):
   - `BuildConfigInfo` - Build configurations with strategies, sources, triggers
   - `ImageStreamInfo` - Container image streams with tags and repositories  
   - `RouteInfo` - HTTP/HTTPS routes with TLS configuration
   - Full type definitions with proper field mapping

2. **Client Integration** (`internal/k8s/client.go`):
   - OpenShift cluster detection via API group discovery
   - Automatic fallback to Kubernetes-only mode
   - `InitializeOpenShiftAfterSetup()` method for proper client initialization
   - Support for both `kubectl` and `oc login` authentication

3. **UI Implementation** (`internal/ui/tui.go`):
   - Extended from 5 to 8 resource tabs (added BuildConfigs, ImageStreams, Routes)
   - Tab-aware navigation with 'j'/'k' keys working correctly in all tabs
   - Resource-specific selection indices and display functions
   - Proper async loading with Bubble Tea message system

#### Navigation & Details Panel Fix
**Critical Fix**: Resolved navigation and details panel issues

**Before**:
- 'j'/'k' keys only worked in Pods tab
- Details panel always showed pod information regardless of current tab
- OpenShift tabs showed "Coming soon" placeholders

**After**:
- Navigation works correctly in all 8 tabs with proper resource selection
- Details panels show relevant resource information:
  - **BuildConfigs**: Source info, build strategies, statistics, last build status
  - **ImageStreams**: Repository URLs, tag listings, image counts
  - **Routes**: Service targets, TLS configuration, routing policies
- Full functionality for OpenShift resource browsing

#### Key Files Modified
- `internal/ui/tui.go`: Navigation logic (lines 390-490), details functions (lines 2006-2117)
- `internal/constants/ui.go`: Extended ResourceTabs to 8 tabs
- `internal/k8s/client.go`: OpenShift client initialization and detection
- `internal/ui/models/app.go`: Tab management with OpenShift support
- `internal/ui/messages/k8s_messages.go`: OpenShift-specific message types

#### Testing Resources
- Created `setup-openshift-resources.sh` script for testing
- Adds sample BuildConfigs and Routes to current OpenShift project
- Verified with real OpenShift cluster (rm3.7wse.p1.openshiftapps.com)

#### Result
- ✅ Complete OpenShift integration with navigation and details
- ✅ Unified TUI architecture eliminates confusion
- ✅ Production-ready OpenShift resource management
- ✅ Maintains full Kubernetes compatibility

---

*This memory file provides Claude Code with comprehensive LazyOC project context to minimize file searching and improve development efficiency.*

