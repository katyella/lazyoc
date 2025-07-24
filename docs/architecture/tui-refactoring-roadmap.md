# TUI Refactoring Roadmap

## Overview
This document outlines the step-by-step plan to refactor the monolithic `tui_simplified.go` (2,290 lines) into a modular architecture.

## Goals
1. **Reduce complexity**: No file over 300 lines
2. **Improve testability**: 80%+ test coverage
3. **Enable team collaboration**: Clear module boundaries
4. **Maintain functionality**: Zero regression
5. **Enhance maintainability**: Clear separation of concerns

## Module Structure

```
internal/tui/
├── app.go                    # Core application (~150 lines)
├── components/
│   ├── component.go          # Component interface (~50 lines)
│   ├── header.go            # Header component (~100 lines)
│   ├── tabs.go              # Tab bar component (~80 lines)
│   ├── content.go           # Content router (~100 lines)
│   ├── statusbar.go         # Status bar (~150 lines)
│   └── panel.go             # Panel base (~100 lines)
├── views/
│   ├── view.go              # View interface (~50 lines)
│   ├── pods.go              # Pod list view (~200 lines)
│   ├── logs.go              # Log view (~200 lines)
│   ├── resources.go         # Resource views (~150 lines)
│   └── details.go           # Detail views (~150 lines)
├── handlers/
│   ├── handler.go           # Handler interface (~50 lines)
│   ├── keyboard.go          # Keyboard handler (~200 lines)
│   ├── messages.go          # Message handler (~200 lines)
│   └── router.go            # Event router (~100 lines)
├── state/
│   ├── manager.go           # State manager (~200 lines)
│   ├── connection.go        # Connection state (~100 lines)
│   ├── resources.go         # Resource state (~100 lines)
│   ├── ui.go                # UI state (~100 lines)
│   └── subscriptions.go     # State subscriptions (~100 lines)
├── services/
│   ├── service.go           # Service interface (~50 lines)
│   ├── kubernetes.go        # K8s service (~200 lines)
│   ├── logs.go              # Log service (~150 lines)
│   ├── projects.go          # Project service (~150 lines)
│   └── monitor.go           # Monitor service (~100 lines)
├── modals/
│   ├── modal.go             # Modal interface (~50 lines)
│   ├── help.go              # Help modal (~100 lines)
│   ├── project.go           # Project modal (~150 lines)
│   └── error.go             # Error modal (~100 lines)
└── utils/
    ├── colors.go            # Color utilities (~100 lines)
    ├── layout.go            # Layout calculations (~100 lines)
    └── formatting.go        # Text formatting (~100 lines)
```

## Implementation Phases

### Phase 1: Foundation (Week 1)
**Goal**: Create base structure and interfaces

#### Day 1-2: Project Setup
- [ ] Create directory structure
- [ ] Define core interfaces
- [ ] Set up base types
- [ ] Create app.go skeleton

#### Day 3-4: State Management
- [ ] Implement state manager
- [ ] Create state slices
- [ ] Add subscription system
- [ ] Write state tests

#### Day 5: Event System
- [ ] Create event bus
- [ ] Define event types
- [ ] Implement command queue
- [ ] Add event tests

### Phase 2: Handler Extraction (Week 2)
**Goal**: Extract and modularize event handling

#### Day 1-2: Message Handlers
- [ ] Extract message handling from Update()
- [ ] Create message router
- [ ] Implement handler registry
- [ ] Add handler tests

#### Day 3-4: Keyboard Handlers
- [ ] Extract keyboard handling
- [ ] Create key mapping system
- [ ] Implement navigation logic
- [ ] Add keyboard tests

#### Day 5: Integration
- [ ] Wire handlers to app
- [ ] Test event flow
- [ ] Verify state updates
- [ ] Fix integration issues

### Phase 3: Component System (Week 3)
**Goal**: Extract UI components

#### Day 1: Component Framework
- [ ] Implement component interface
- [ ] Create component registry
- [ ] Add lifecycle methods
- [ ] Write component tests

#### Day 2: Header & StatusBar
- [ ] Extract header component
- [ ] Extract statusbar component
- [ ] Implement component state
- [ ] Add rendering tests

#### Day 3: Content Components
- [ ] Extract main panel
- [ ] Extract detail panel
- [ ] Extract log panel
- [ ] Add panel tests

#### Day 4: Modal Components
- [ ] Extract help modal
- [ ] Extract project modal
- [ ] Extract error modal
- [ ] Add modal tests

#### Day 5: View System
- [ ] Implement view manager
- [ ] Create view registry
- [ ] Add view switching
- [ ] Test view transitions

### Phase 4: Service Layer (Week 4)
**Goal**: Extract business logic

#### Day 1-2: Kubernetes Service
- [ ] Extract K8s operations
- [ ] Create service interface
- [ ] Implement error handling
- [ ] Add service tests

#### Day 3: Log Service
- [ ] Extract log streaming
- [ ] Implement log formatting
- [ ] Add log caching
- [ ] Write log tests

#### Day 4: Project Service
- [ ] Extract project operations
- [ ] Implement project switching
- [ ] Add project caching
- [ ] Write project tests

#### Day 5: Integration
- [ ] Wire services to app
- [ ] Test service interactions
- [ ] Verify error propagation
- [ ] Performance testing

### Phase 5: Finalization (Week 5)
**Goal**: Complete refactoring and documentation

#### Day 1-2: Migration
- [ ] Migrate remaining code
- [ ] Remove old file
- [ ] Update imports
- [ ] Fix build issues

#### Day 3: Testing
- [ ] Integration tests
- [ ] Regression tests
- [ ] Performance tests
- [ ] Coverage report

#### Day 4: Documentation
- [ ] API documentation
- [ ] Architecture guide
- [ ] Migration guide
- [ ] Example usage

#### Day 5: Review
- [ ] Code review
- [ ] Security review
- [ ] Performance review
- [ ] Final adjustments

## Migration Strategy

### Step 1: Parallel Development
- Keep `tui_simplified.go` functional
- Build new modules alongside
- Use feature flags for switching

### Step 2: Incremental Migration
```go
// temporary bridge
type App struct {
    legacy *SimplifiedTUI  // old implementation
    new    *tui.App       // new implementation
    useNew bool           // feature flag
}
```

### Step 3: Component by Component
1. Start with stateless components (utils)
2. Move to simple components (header, statusbar)
3. Then complex components (content, modals)
4. Finally core logic (state, handlers)

### Step 4: Testing at Each Step
- Unit tests for new modules
- Integration tests for combinations
- Regression tests against legacy
- Performance benchmarks

## Risk Mitigation

### Risk 1: Breaking Changes
- **Mitigation**: Comprehensive test suite
- **Mitigation**: Feature flags for rollback
- **Mitigation**: Incremental deployment

### Risk 2: Performance Regression
- **Mitigation**: Benchmark before/after
- **Mitigation**: Profile critical paths
- **Mitigation**: Optimize hot spots

### Risk 3: Team Disruption
- **Mitigation**: Clear communication
- **Mitigation**: Pair programming
- **Mitigation**: Daily syncs

## Success Criteria

### Code Quality
- [ ] No file exceeds 300 lines
- [ ] Cyclomatic complexity < 10
- [ ] Test coverage > 80%
- [ ] No circular dependencies

### Functionality
- [ ] All features working
- [ ] No performance regression
- [ ] Improved error handling
- [ ] Better user experience

### Maintainability
- [ ] Clear module boundaries
- [ ] Comprehensive documentation
- [ ] Easy to extend
- [ ] Simple to debug

## Timeline Summary

| Week | Phase | Deliverables |
|------|-------|--------------|
| 1 | Foundation | Core structure, state management |
| 2 | Handlers | Event handling system |
| 3 | Components | UI component system |
| 4 | Services | Business logic layer |
| 5 | Finalization | Testing, docs, deployment |

## Next Steps

1. Review and approve roadmap
2. Set up tracking system
3. Begin Phase 1 implementation
4. Daily progress updates
5. Weekly demos