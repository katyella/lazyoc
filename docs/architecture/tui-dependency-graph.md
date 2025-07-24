# TUI Dependencies and Component Map

## Current Dependencies Flow

```mermaid
graph TB
    subgraph "User Interface Layer"
        TUI[SimplifiedTUI<br/>2290 lines]
        Update[Update Method<br/>456 lines]
        View[View Method]
        Render[Rendering Methods]
    end

    subgraph "State Management"
        ConnState[Connection State]
        UIState[UI State]
        ResourceState[Resource State]
        ModalState[Modal State]
    end

    subgraph "Event Processing"
        KeyEvents[Keyboard Events]
        Messages[Tea Messages]
        Commands[Tea Commands]
    end

    subgraph "Kubernetes Layer"
        K8sClient[K8s Client]
        AuthProvider[Auth Provider]
        Monitor[Connection Monitor]
        Projects[Project Manager]
        Resources[Resource Client]
    end

    subgraph "External Dependencies"
        BubbleTea[Bubble Tea]
        Lipgloss[Lipgloss]
        K8sAPI[k8s.io/client-go]
    end

    subgraph "Internal Utilities"
        Constants[Constants]
        Logging[Logging]
        ErrorHandling[Error Handling]
        Navigation[Navigation]
    end

    %% User interactions
    User([User]) --> KeyEvents
    KeyEvents --> Update
    Messages --> Update
    
    %% Update method interactions
    Update --> ConnState
    Update --> UIState
    Update --> ResourceState
    Update --> ModalState
    Update --> Commands
    
    %% State to View
    ConnState --> View
    UIState --> View
    ResourceState --> View
    ModalState --> View
    
    %% View to Render
    View --> Render
    Render --> BubbleTea
    Render --> Lipgloss
    
    %% K8s interactions
    TUI --> K8sClient
    K8sClient --> AuthProvider
    K8sClient --> Monitor
    K8sClient --> Projects
    K8sClient --> Resources
    
    %% K8s to external
    K8sClient --> K8sAPI
    
    %% Utilities
    TUI --> Constants
    TUI --> Logging
    TUI --> ErrorHandling
    TUI --> Navigation
    
    %% Commands back to messages
    Commands --> Messages
```

## Component Coupling Analysis

### High Coupling Areas

1. **Update Method**
   - Directly modifies 15+ state fields
   - Handles 20+ different message types
   - Contains business logic mixed with UI logic

2. **Rendering Methods**
   - Direct access to all state fields
   - Complex layout calculations inline
   - Style definitions mixed with content

3. **K8s Operations**
   - Scattered throughout the file
   - No clear separation between UI and API calls
   - Error handling mixed with UI updates

### Circular Dependencies

```mermaid
graph LR
    A[State Updates] --> B[View Rendering]
    B --> C[User Events]
    C --> A
    
    D[K8s Operations] --> E[Error Display]
    E --> F[Recovery Actions]
    F --> D
```

## Proposed Decoupled Architecture

```mermaid
graph TB
    subgraph "Presentation Layer"
        App[App Controller]
        ViewManager[View Manager]
        ComponentRegistry[Component Registry]
    end

    subgraph "Component Layer"
        Header[Header Component]
        Content[Content Component]
        StatusBar[StatusBar Component]
        Modals[Modal Components]
    end

    subgraph "State Layer"
        StateManager[State Manager]
        ConnectionSlice[Connection State]
        ResourceSlice[Resource State]
        UISlice[UI State]
    end

    subgraph "Service Layer"
        K8sService[K8s Service]
        LogService[Log Service]
        ProjectService[Project Service]
    end

    subgraph "Infrastructure"
        EventBus[Event Bus]
        CommandQueue[Command Queue]
        ErrorHandler[Error Handler]
    end

    %% User flow
    User([User]) --> App
    App --> ViewManager
    ViewManager --> ComponentRegistry
    
    %% Components to State
    Header --> StateManager
    Content --> StateManager
    StatusBar --> StateManager
    Modals --> StateManager
    
    %% State slices
    StateManager --> ConnectionSlice
    StateManager --> ResourceSlice
    StateManager --> UISlice
    
    %% Services
    K8sService --> ConnectionSlice
    LogService --> ResourceSlice
    ProjectService --> ResourceSlice
    
    %% Infrastructure
    App --> EventBus
    EventBus --> StateManager
    StateManager --> CommandQueue
    CommandQueue --> K8sService
    
    %% Error flow
    K8sService --> ErrorHandler
    ErrorHandler --> StateManager
```

## Dependency Injection Plan

```go
// Core interfaces
type Component interface {
    Init(state StateReader) error
    Update(event Event) (Command, error)
    View() string
}

type StateReader interface {
    GetConnectionState() ConnectionState
    GetResourceState() ResourceState
    GetUIState() UIState
}

type Service interface {
    Execute(ctx context.Context, cmd Command) error
}

// Dependency injection
type App struct {
    state      *StateManager
    components map[string]Component
    services   map[string]Service
    eventBus   *EventBus
}
```

## Module Boundaries

### Clear Responsibilities

1. **Components**: Only UI rendering and local state
2. **State Manager**: Central state coordination
3. **Services**: Business logic and external integrations
4. **Event Bus**: Message routing and command dispatch

### Communication Rules

1. Components → State: Read only via interfaces
2. Components → Services: Via commands only
3. Services → State: Via events only
4. State → Components: Via subscriptions

## Testability Improvements

### Before (Current)
- Cannot test Update() without full setup
- Cannot isolate rendering logic
- K8s operations require live cluster

### After (Proposed)
- Mock state for component tests
- Mock services for integration tests
- Mock event bus for flow tests
- Isolated unit tests for each module