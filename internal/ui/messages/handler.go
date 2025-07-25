package messages

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// MessagePriority defines the priority level for messages
type MessagePriority int

const (
	PriorityLow MessagePriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// MessageHandler provides utilities for message handling and processing
type MessageHandler struct {
	// Message batching support
	batchedMessages []tea.Msg
	batchSize       int

	// Performance tracking
	messageCount  map[string]int
	highFreqTypes map[string]bool
}

// NewMessageHandler creates a new message handler with default settings
func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		batchedMessages: make([]tea.Msg, 0, 10),
		batchSize:       10,
		messageCount:    make(map[string]int),
		highFreqTypes:   make(map[string]bool),
	}
}

// GetMessagePriority determines the priority of a message
func (h *MessageHandler) GetMessagePriority(msg tea.Msg) MessagePriority {
	switch msg.(type) {
	// Critical messages that must be processed immediately
	case tea.QuitMsg:
		return PriorityCritical
	case ErrorMsg:
		return PriorityCritical

	// High priority messages that affect UI state
	case tea.WindowSizeMsg:
		return PriorityHigh
	case InitMsg:
		return PriorityHigh
	case ConnectedMsg, DisconnectedMsg:
		return PriorityHigh

	// Normal priority messages
	case tea.KeyMsg:
		return PriorityNormal
	case LoadingMsg:
		return PriorityNormal
	case StatusMsg:
		return PriorityNormal

	// Low priority messages
	case RefreshMsg:
		return PriorityLow

	default:
		return PriorityNormal
	}
}

// TrackMessage tracks message frequency for performance optimization
func (h *MessageHandler) TrackMessage(msg tea.Msg) {
	msgType := getMessageType(msg)
	h.messageCount[msgType]++

	// Mark as high frequency if we've seen more than 100 in the session
	if h.messageCount[msgType] > 100 {
		h.highFreqTypes[msgType] = true
	}
}

// IsHighFrequency checks if a message type is high frequency
func (h *MessageHandler) IsHighFrequency(msg tea.Msg) bool {
	return h.highFreqTypes[getMessageType(msg)]
}

// ShouldBatch determines if a message should be batched
func (h *MessageHandler) ShouldBatch(msg tea.Msg) bool {
	switch msg.(type) {
	// Never batch critical messages
	case tea.QuitMsg, ErrorMsg, tea.WindowSizeMsg:
		return false

	// Batch similar status updates
	case StatusMsg, LoadingMsg:
		return h.IsHighFrequency(msg)

	default:
		return false
	}
}

// AddToBatch adds a message to the batch queue
func (h *MessageHandler) AddToBatch(msg tea.Msg) {
	h.batchedMessages = append(h.batchedMessages, msg)
}

// ProcessBatch processes and clears the batched messages
func (h *MessageHandler) ProcessBatch() []tea.Msg {
	if len(h.batchedMessages) == 0 {
		return nil
	}

	// Deduplicate similar messages (keep only the latest)
	processed := h.deduplicateMessages(h.batchedMessages)

	// Clear the batch
	h.batchedMessages = h.batchedMessages[:0]

	return processed
}

// deduplicateMessages removes duplicate messages keeping only the latest
func (h *MessageHandler) deduplicateMessages(messages []tea.Msg) []tea.Msg {
	// Use a map to track the latest instance of each message type
	latest := make(map[string]tea.Msg)
	order := make([]string, 0)

	for _, msg := range messages {
		key := getMessageKey(msg)
		if _, exists := latest[key]; !exists {
			order = append(order, key)
		}
		latest[key] = msg
	}

	// Reconstruct in order
	result := make([]tea.Msg, 0, len(order))
	for _, key := range order {
		result = append(result, latest[key])
	}

	return result
}

// getMessageType returns a string representation of the message type
func getMessageType(msg tea.Msg) string {
	switch msg.(type) {
	case tea.KeyMsg:
		return "KeyMsg"
	case tea.WindowSizeMsg:
		return "WindowSizeMsg"
	case ErrorMsg:
		return "ErrorMsg"
	case LoadingMsg:
		return "LoadingMsg"
	case StatusMsg:
		return "StatusMsg"
	case InitMsg:
		return "InitMsg"
	case RefreshMsg:
		return "RefreshMsg"
	case ConnectedMsg:
		return "ConnectedMsg"
	case DisconnectedMsg:
		return "DisconnectedMsg"
	default:
		return "Unknown"
	}
}

// getMessageKey returns a unique key for deduplication
func getMessageKey(msg tea.Msg) string {
	switch m := msg.(type) {
	case StatusMsg:
		// Group status messages by type
		return fmt.Sprintf("StatusMsg:%d", m.Type)
	case LoadingMsg:
		// Group loading messages together
		return "LoadingMsg"
	default:
		// Use type name for other messages
		return getMessageType(msg)
	}
}

// MessageProcessor is a function that processes a message and returns commands
type MessageProcessor func(tea.Msg) (bool, []tea.Cmd)

// ProcessWithHandlers processes a message through a series of handlers
func ProcessWithHandlers(msg tea.Msg, handlers ...MessageProcessor) []tea.Cmd {
	var allCmds []tea.Cmd

	for _, handler := range handlers {
		handled, cmds := handler(msg)
		if len(cmds) > 0 {
			allCmds = append(allCmds, cmds...)
		}
		if handled {
			// Stop processing if a handler consumed the message
			break
		}
	}

	return allCmds
}
