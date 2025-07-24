# TUI Refactoring Analysis

## Current State Analysis - tui_simplified.go

### File Overview
- **Total Lines**: 2,290
- **Primary Purpose**: Monolithic TUI implementation for LazyOC
- **Framework**: Bubble Tea (Charm)
- **Current Issues**: 
  - Single file containing all TUI logic
  - Tight coupling between components
  - Difficult to test individual parts
  - Complex state management

### Component Breakdown

#### 1. Core Structure (Lines 31-108)
The `SimplifiedTUI` struct contains all application state:
- Connection state (k8s client, auth, monitoring)
- Resource state (pods, logs)
- UI state (theme, panels, visibility)
- Modal state (project selection, error display)
- Navigation state (selected items, scroll positions)

#### 2. Initialization Layer (Lines 110-210)
- Constructor: `NewSimplifiedTUI`
- Kubeconfig setup: `SetKubeconfig`
- Bubbletea initialization: `Init`

#### 3. Event Handling (Lines 212-668)
The massive `Update` method handles:
- Keyboard events (navigation, commands)
- Message processing (API responses, timers)
- State mutations
- Command generation

#### 4. Rendering System (Lines 670-1335)
- Main view composition
- Component rendering (header, tabs, content, status bar)
- Modal overlays (help, project selection, errors)
- Layout calculations

#### 5. Kubernetes Integration (Lines 1383-1715)
- Client initialization
- Resource loading (pods, logs)
- Connection monitoring
- Error handling

#### 6. Project Management (Lines 1717-2095)
- Project/namespace switching
- Project list modal
- OpenShift integration

#### 7. Utilities (Lines 2183-2290)
- Helper functions
- Log formatting and coloring
- Scroll calculations

### Dependencies Map

```
SimplifiedTUI
├── Internal Packages
│   ├── k8s (client, cluster detection)
│   ├── auth (authentication provider)
│   ├── monitor (connection monitoring)
│   ├── projects (project management)
│   ├── resources (pod operations)
│   ├── logging (debug logging)
│   ├── ui/components (error display)
│   ├── ui/errors (error mapping)
│   ├── ui/messages (message types)
│   ├── ui/models (base app model)
│   ├── ui/navigation (navigation controller)
│   └── constants (configuration values)
└── External Packages
    ├── bubbletea (TUI framework)
    ├── lipgloss (styling)
    └── k8s.io/client-go (Kubernetes client)
```

### State Flow

```
User Input → Update() → State Mutation → View() → Screen
     ↑                          ↓
     └── Commands ← Tea.Cmd ← ─┘
```

### Identified Problems

1. **Monolithic Update Method**: 456 lines handling all events
2. **Mixed Responsibilities**: UI, business logic, and K8s operations intertwined
3. **Complex State Management**: All state in single struct
4. **Testing Challenges**: Cannot test components in isolation
5. **Code Duplication**: Similar patterns repeated (loading, error handling)

### Refactoring Strategy

#### Phase 1: Extract Event Handling
- Move keyboard handling to `handlers/keyboard.go`
- Extract message handling to `handlers/messages.go`
- Create event router to dispatch events

#### Phase 2: Component Extraction
- Extract each visual component to separate files
- Create component interfaces for consistency
- Implement component lifecycle (Init, Update, View)

#### Phase 3: State Management
- Create centralized state manager
- Implement state slices for different domains
- Add state change notifications

#### Phase 4: Service Layer
- Extract K8s operations to service layer
- Create abstractions for resource operations
- Implement proper error propagation

#### Phase 5: Testing & Documentation
- Add unit tests for each module
- Create integration tests
- Document component interactions

### Target Architecture

```
internal/tui/
├── app.go                 # Main application struct
├── components/            # UI components
│   ├── header.go
│   ├── tabs.go
│   ├── content.go
│   ├── statusbar.go
│   └── panel.go
├── views/                 # View logic
│   ├── pods.go
│   ├── logs.go
│   └── resources.go
├── handlers/              # Event handling
│   ├── keyboard.go
│   ├── messages.go
│   └── router.go
├── state/                 # State management
│   ├── manager.go
│   ├── connection.go
│   ├── resources.go
│   └── ui.go
├── services/              # Business logic
│   ├── kubernetes.go
│   ├── logs.go
│   └── projects.go
├── modals/               # Modal dialogs
│   ├── help.go
│   ├── project.go
│   └── error.go
└── utils/                # Utilities
    ├── colors.go
    ├── layout.go
    └── formatting.go
```

### Implementation Plan

1. **Week 1**: Extract event handlers and create basic component structure
2. **Week 2**: Implement state management and service layer
3. **Week 3**: Complete component extraction and add tests
4. **Week 4**: Integration testing and documentation

### Success Metrics

- [ ] No single file over 300 lines
- [ ] 80%+ test coverage per module
- [ ] Clear separation of concerns
- [ ] Improved build times
- [ ] Easier debugging and maintenance

### Risks and Mitigations

1. **Risk**: Breaking existing functionality
   - **Mitigation**: Incremental refactoring with tests

2. **Risk**: Performance degradation
   - **Mitigation**: Benchmark before and after

3. **Risk**: Increased complexity
   - **Mitigation**: Clear documentation and examples