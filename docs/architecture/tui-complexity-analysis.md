# TUI Code Complexity Analysis

## Cyclomatic Complexity Report

### High Complexity Methods

| Method | Lines | Cyclomatic Complexity | Issues |
|--------|-------|----------------------|---------|
| Update() | 456 | ~60 | Massive switch statement, nested conditions |
| renderContent() | 243 | ~25 | Complex layout calculations, multiple render paths |
| InitializeK8sClient() | 96 | ~15 | Multiple error paths, nested initialization |
| renderProjectModal() | 114 | ~20 | Complex modal state handling |
| handleProjectModalKeys() | 54 | ~12 | Multiple key handling branches |

### Code Duplication Analysis

#### Pattern 1: Loading Operations
```go
// Repeated pattern across pods, logs, projects
func (t *SimplifiedTUI) loadResource() tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(...)
        defer cancel()
        
        resource, err := t.service.Get(ctx, ...)
        if err != nil {
            return ErrorMsg{...}
        }
        return ResourceLoadedMsg{...}
    }
}
```
**Occurrences**: 5 times (pods, logs, projects, cluster info, pod details)

#### Pattern 2: Error Display
```go
// Repeated error handling pattern
if err != nil {
    t.lastError = err
    t.errorModalVisible = true
    t.App.SetNotification(...)
    return t, nil
}
```
**Occurrences**: 12 times throughout the file

#### Pattern 3: Modal Rendering
```go
// Similar structure for help, project, and error modals
if t.modalVisible {
    overlay := lipgloss.NewStyle().
        Width(width).
        Height(height).
        Align(...).
        Render(modal)
    return lipgloss.Place(...)
}
```
**Occurrences**: 3 times

### Cognitive Complexity Issues

#### 1. State Mutation Complexity
The SimplifiedTUI struct has **40+ mutable fields** that can be modified from:
- 20+ locations in Update()
- 10+ locations in initialization
- 5+ locations in callbacks

#### 2. Control Flow Complexity
```
Update() method flow:
├── Check connection state (3 branches)
├── Handle messages (25+ message types)
│   ├── Each with 2-5 sub-branches
│   └── Many modifying multiple state fields
├── Handle keyboard events (30+ keys)
│   ├── Modal-specific handling (3 modals)
│   ├── Panel-specific handling (3 panels)
│   └── Tab-specific handling (5 tabs)
└── Generate commands (10+ command types)
```

#### 3. Rendering Complexity
```
View() rendering tree:
├── Main layout calculation
├── Conditional rendering (5+ conditions)
│   ├── Connection state
│   ├── Modal visibility
│   ├── Panel visibility
│   └── Resource availability
├── Dynamic styling (theme-dependent)
└── Overlay compositing
```

### Lines of Code by Functionality

```
Total: 2,290 lines

Initialization:    100 lines (4%)
Event Handling:    456 lines (20%)
Rendering:         665 lines (29%)
K8s Operations:    332 lines (14%)
Project Mgmt:      378 lines (17%)
Utilities:         107 lines (5%)
State/Types:       252 lines (11%)
```

### Complexity Metrics Summary

| Metric | Current | Target | Improvement Needed |
|--------|---------|--------|-------------------|
| Avg Method Length | 57 lines | <20 lines | 65% reduction |
| Max Method Length | 456 lines | <50 lines | 89% reduction |
| Cyclomatic Complexity | 60 (max) | <10 | 83% reduction |
| File Length | 2,290 lines | <300 lines | 87% reduction |
| State Fields | 40+ | <10 per module | 75% reduction |

### Refactoring Priority Matrix

| Component | Complexity | Impact | Priority |
|-----------|-----------|---------|----------|
| Update() method | Very High | Critical | P0 |
| Event handling | High | High | P0 |
| State management | High | High | P0 |
| Rendering system | Medium | Medium | P1 |
| K8s operations | Medium | Medium | P1 |
| Modal handling | Low | Low | P2 |
| Utilities | Low | Low | P2 |

### Complexity Reduction Strategy

#### Phase 1: Break Update() Method
- Extract message handlers: ~200 lines reduction
- Extract keyboard handlers: ~150 lines reduction
- Extract command generation: ~50 lines reduction
- **Result**: Update() reduced to <50 lines

#### Phase 2: Component Extraction
- Header component: ~50 lines
- Content component: ~200 lines per view
- StatusBar component: ~195 lines
- Modal components: ~100 lines each
- **Result**: No component >200 lines

#### Phase 3: State Refactoring
- Connection state slice: 5 fields
- Resource state slice: 8 fields
- UI state slice: 10 fields
- Modal state slice: 8 fields
- **Result**: Logical grouping, easier testing

#### Phase 4: Service Layer
- K8s service: All K8s operations
- Log service: Log streaming
- Project service: Project management
- **Result**: Clear separation of concerns

### Expected Improvements

1. **Testability**: From 0% to 80%+ coverage
2. **Maintainability**: 87% reduction in file size
3. **Readability**: 65% reduction in method complexity
4. **Debugging**: Clear module boundaries
5. **Team velocity**: Parallel development possible