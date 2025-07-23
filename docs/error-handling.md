# Error Handling and Recovery System

LazyOC implements a comprehensive error handling and recovery system designed to provide user-friendly error messages and automated recovery mechanisms for common Kubernetes/OpenShift issues.

## Features

### 1. User-Friendly Error Messages

The system transforms low-level Kubernetes errors into clear, actionable messages:

- **Connection Errors**: "Cannot connect to the Kubernetes cluster. The cluster may be down or unreachable."
- **Authentication Errors**: "Your credentials are invalid or have expired. Run 'oc login' to refresh your authentication."
- **Permission Errors**: "You don't have permission to perform this operation. Contact your cluster administrator."
- **Project/Namespace Errors**: "The requested project or namespace does not exist or you don't have access to it."

### 2. Error Categories and Severity

Errors are classified by:

- **Categories**: Connection, Authentication, Resource, Configuration, Network, Permission, Project, General
- **Severity**: Info, Warning, Error, Critical
- **Retryability**: Whether the error can be automatically retried

### 3. Automatic Recovery

The system provides automated recovery mechanisms:

- **Auto-retry**: Connection errors are automatically retried up to 3 times with exponential backoff
- **Smart retry logic**: Only retries errors that are likely to succeed (e.g., network timeouts)
- **Backoff strategy**: Prevents overwhelming the cluster with rapid retry attempts

### 4. Interactive Error Management

Users can interact with errors through:

- **Error Modal**: Press `e` to view detailed error information
- **Recovery Actions**: Navigate with ↑↓ arrows and press Enter to execute
- **Manual Retry**: Press `r` to manually retry connections or refresh resources

## Error Display Components

### Status Bar Integration

- Error indicators appear in the status bar with appropriate icons
- Connection status shows error state with red indicators
- Hints show `e errors` when errors are present

### Error Modal

The error modal provides:

- **Title and Icon**: Clear identification of error type
- **User Message**: Simplified explanation of what went wrong
- **Technical Details**: Raw error information for debugging
- **Recovery Actions**: Suggested actions to resolve the issue
- **Suggested Action**: Contextual advice based on error category

## Keyboard Shortcuts

- `e` - Show error modal (when errors exist)
- `r` - Manual retry/reconnect or refresh resources
- `esc` - Close error modal
- `↑/↓` - Navigate recovery actions in error modal
- `Enter` - Execute selected recovery action
- `c` - Clear all errors (in error modal)

## Recovery Actions

Available recovery actions depend on error type:

### Connection/Network Errors
- **Retry Connection**: Attempt to reconnect to the cluster
- **Check Network**: Manual verification of network connectivity

### Authentication Errors
- **Re-authenticate**: Run 'oc login' to refresh authentication

### Project Errors
- **Refresh Projects**: Reload the project list
- **Switch Project**: Try switching to a different project

### Resource Errors
- **Refresh Resources**: Reload the current resource list

### General Actions
- **Refresh Application**: Perform full refresh of all data

## Auto-Retry Behavior

The system automatically retries:

- **Connection failures** due to network issues
- **Timeout errors** from slow responses
- **Temporary resource unavailability**

It does NOT auto-retry:

- **Authentication failures** (requires user action)
- **Permission errors** (requires admin intervention)
- **Configuration errors** (requires manual fixes)

## Implementation Details

### Error Mapping

The `errors.MapKubernetesError()` function maps common Kubernetes errors to user-friendly equivalents by analyzing error messages for keywords like:

- "connection refused" → Connection Failed
- "unauthorized" → Authentication Failed
- "forbidden" → Access Denied
- "not found" → Resource/Project Not Found
- "timeout" → Connection Timeout

### Retry Strategy

Default retry configuration:
- **Max Attempts**: 3
- **Initial Delay**: 1 second
- **Max Delay**: 30 seconds
- **Backoff Factor**: 2.0 (exponential)

Connection-specific retry:
- **Max Attempts**: 5
- **Initial Delay**: 2 seconds
- **Max Delay**: 60 seconds
- **Backoff Factor**: 1.5

### Error Persistence

- Errors are stored in the ErrorDisplayComponent
- Maximum of 10 recent errors are kept
- Errors automatically clear on successful operations
- Manual clearing available through error modal

## Usage Examples

### Viewing Errors

```bash
# If connection fails, you'll see:
❌ Connection failed: Cannot connect to the Kubernetes cluster

# Press 'e' to open error modal with detailed information and recovery options
```

### Manual Recovery

```bash
# Press 'r' to retry connection or refresh resources
# Press 'ctrl+p' to switch projects if current project has issues
# Press 'e' then navigate to recovery actions for guided fixes
```

### Automatic Recovery

```bash
# Connection errors automatically retry:
🔄 Auto-retry attempt 1/3 in 5 seconds...
🔄 Attempting reconnection (attempt 1/3)...
✨ Connection restored after 1 retries
```

This error handling system ensures users have clear visibility into issues and multiple pathways to recovery, whether automatic or manual.