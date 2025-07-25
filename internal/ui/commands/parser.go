package commands

import (
	"fmt"
	"strings"
)

// CommandType represents the type of command
type CommandType int

const (
	CommandTypeUnknown CommandType = iota
	CommandTypeQuit
	CommandTypeHelp
	CommandTypeConnect
	CommandTypeNamespace
	CommandTypeContext
	CommandTypeRefresh
	CommandTypeExec
	CommandTypeDelete
	CommandTypeLogs
	CommandTypeDescribe
	CommandTypeEdit
	CommandTypeScale
	CommandTypeRestart
	CommandTypeExport
	CommandTypeApply
)

// Command represents a parsed command
type Command struct {
	Type     CommandType
	Resource string
	Name     string
	Args     []string
	Flags    map[string]string
	RawInput string
}

// ParseCommand parses a command string into a Command structure
func ParseCommand(input string) (*Command, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty command")
	}

	cmd := &Command{
		RawInput: input,
		Flags:    make(map[string]string),
	}

	// Split the command into parts
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	// Parse the main command
	switch strings.ToLower(parts[0]) {
	case "q", "quit", "exit":
		cmd.Type = CommandTypeQuit
	case "h", "help", "?":
		cmd.Type = CommandTypeHelp
		if len(parts) > 1 {
			cmd.Args = parts[1:]
		}
	case "connect":
		cmd.Type = CommandTypeConnect
		if len(parts) > 1 {
			cmd.Args = parts[1:]
		}
	case "ns", "namespace":
		cmd.Type = CommandTypeNamespace
		if len(parts) > 1 {
			cmd.Name = parts[1]
		}
	case "ctx", "context":
		cmd.Type = CommandTypeContext
		if len(parts) > 1 {
			cmd.Name = parts[1]
		}
	case "r", "refresh", "reload":
		cmd.Type = CommandTypeRefresh
	case "exec":
		cmd.Type = CommandTypeExec
		if len(parts) < 2 {
			return nil, fmt.Errorf("exec requires a resource name")
		}
		cmd.Name = parts[1]
		if len(parts) > 2 {
			cmd.Args = parts[2:]
		}
	case "delete", "del", "rm":
		cmd.Type = CommandTypeDelete
		if len(parts) < 3 {
			return nil, fmt.Errorf("delete requires resource type and name")
		}
		cmd.Resource = parts[1]
		cmd.Name = parts[2]
	case "logs", "log":
		cmd.Type = CommandTypeLogs
		if len(parts) < 2 {
			return nil, fmt.Errorf("logs requires a pod name")
		}
		cmd.Name = parts[1]
		// Parse flags like -f for follow
		for i := 2; i < len(parts); i++ {
			if strings.HasPrefix(parts[i], "-") {
				key := strings.TrimPrefix(parts[i], "-")
				cmd.Flags[key] = "true"
			}
		}
	case "describe", "desc":
		cmd.Type = CommandTypeDescribe
		if len(parts) < 3 {
			return nil, fmt.Errorf("describe requires resource type and name")
		}
		cmd.Resource = parts[1]
		cmd.Name = parts[2]
	case "edit":
		cmd.Type = CommandTypeEdit
		if len(parts) < 3 {
			return nil, fmt.Errorf("edit requires resource type and name")
		}
		cmd.Resource = parts[1]
		cmd.Name = parts[2]
	case "scale":
		cmd.Type = CommandTypeScale
		if len(parts) < 3 {
			return nil, fmt.Errorf("scale requires deployment name and replica count")
		}
		cmd.Name = parts[1]
		cmd.Args = parts[2:]
	case "restart":
		cmd.Type = CommandTypeRestart
		if len(parts) < 2 {
			return nil, fmt.Errorf("restart requires a deployment name")
		}
		cmd.Name = parts[1]
	case "export":
		cmd.Type = CommandTypeExport
		if len(parts) < 3 {
			return nil, fmt.Errorf("export requires resource type and name")
		}
		cmd.Resource = parts[1]
		cmd.Name = parts[2]
	case "apply":
		cmd.Type = CommandTypeApply
		if len(parts) < 2 {
			return nil, fmt.Errorf("apply requires a file path")
		}
		cmd.Args = parts[1:]
	default:
		cmd.Type = CommandTypeUnknown
		return nil, fmt.Errorf("unknown command: %s", parts[0])
	}

	return cmd, nil
}

// GetCommandHelp returns help text for a specific command
func GetCommandHelp(cmdType CommandType) string {
	switch cmdType {
	case CommandTypeQuit:
		return "quit/q/exit - Exit the application"
	case CommandTypeHelp:
		return "help/h/? [command] - Show help for a command"
	case CommandTypeConnect:
		return "connect [context] - Connect to a Kubernetes cluster"
	case CommandTypeNamespace:
		return "namespace/ns [name] - Switch to a different namespace"
	case CommandTypeContext:
		return "context/ctx [name] - Switch to a different context"
	case CommandTypeRefresh:
		return "refresh/r/reload - Refresh the current view"
	case CommandTypeExec:
		return "exec <pod> [command...] - Execute a command in a pod"
	case CommandTypeDelete:
		return "delete/del/rm <type> <name> - Delete a resource"
	case CommandTypeLogs:
		return "logs/log <pod> [-f] - View pod logs (-f to follow)"
	case CommandTypeDescribe:
		return "describe/desc <type> <name> - Describe a resource"
	case CommandTypeEdit:
		return "edit <type> <name> - Edit a resource"
	case CommandTypeScale:
		return "scale <deployment> <replicas> - Scale a deployment"
	case CommandTypeRestart:
		return "restart <deployment> - Restart a deployment"
	case CommandTypeExport:
		return "export <type> <name> - Export a resource as YAML"
	case CommandTypeApply:
		return "apply <file> - Apply a configuration file"
	default:
		return "Unknown command"
	}
}

// GetAllCommands returns help text for all commands
func GetAllCommands() string {
	commands := []CommandType{
		CommandTypeQuit,
		CommandTypeHelp,
		CommandTypeConnect,
		CommandTypeNamespace,
		CommandTypeContext,
		CommandTypeRefresh,
		CommandTypeExec,
		CommandTypeDelete,
		CommandTypeLogs,
		CommandTypeDescribe,
		CommandTypeEdit,
		CommandTypeScale,
		CommandTypeRestart,
		CommandTypeExport,
		CommandTypeApply,
	}

	var help strings.Builder
	help.WriteString("Available Commands:\n\n")

	for _, cmd := range commands {
		help.WriteString("  " + GetCommandHelp(cmd) + "\n")
	}

	return help.String()
}
