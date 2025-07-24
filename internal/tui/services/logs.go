package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/katyella/lazyoc/internal/tui/views"
)

// LogsService manages log streaming and buffering
type LogsService struct {
	mu sync.RWMutex
	
	// Kubernetes service
	k8s *KubernetesService
	
	// Active streams
	streams map[string]*logStream
	
	// Log buffers
	buffers map[string]*LogBuffer
	
	// Observers
	observers []LogsObserver
}

// logStream represents an active log stream
type logStream struct {
	podName       string
	containerName string
	context       context.Context
	cancel        context.CancelFunc
	logChan       <-chan string
	errChan       <-chan error
}

// LogBuffer stores log entries for a resource
type LogBuffer struct {
	Resource  string
	Container string
	Entries   []views.LogEntry
	MaxSize   int
}

// LogsObserver receives log events
type LogsObserver interface {
	OnLogsReceived(resource, container string, entries []views.LogEntry)
	OnLogError(resource, container string, err error)
}

// NewLogsService creates a new logs service
func NewLogsService(k8s *KubernetesService) *LogsService {
	return &LogsService{
		k8s:       k8s,
		streams:   make(map[string]*logStream),
		buffers:   make(map[string]*LogBuffer),
		observers: make([]LogsObserver, 0),
	}
}

// AddObserver adds a logs observer
func (l *LogsService) AddObserver(observer LogsObserver) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.observers = append(l.observers, observer)
}

// RemoveObserver removes a logs observer
func (l *LogsService) RemoveObserver(observer LogsObserver) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	for i, obs := range l.observers {
		if obs == observer {
			l.observers = append(l.observers[:i], l.observers[i+1:]...)
			break
		}
	}
}

// StartStreaming starts streaming logs for a pod
func (l *LogsService) StartStreaming(podName, containerName string, lines int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Stop existing stream if any
	streamKey := l.getStreamKey(podName, containerName)
	if existing, ok := l.streams[streamKey]; ok {
		existing.cancel()
		delete(l.streams, streamKey)
	}
	
	// Create context for stream
	ctx, cancel := context.WithCancel(context.Background())
	
	// Start streaming
	logChan, errChan := l.k8s.StreamPodLogs(ctx, podName, containerName, lines)
	
	// Create stream record
	stream := &logStream{
		podName:       podName,
		containerName: containerName,
		context:       ctx,
		cancel:        cancel,
		logChan:       logChan,
		errChan:       errChan,
	}
	
	l.streams[streamKey] = stream
	
	// Start processing logs
	go l.processLogStream(stream)
	
	return nil
}

// StopStreaming stops streaming logs for a pod
func (l *LogsService) StopStreaming(podName, containerName string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	streamKey := l.getStreamKey(podName, containerName)
	if stream, ok := l.streams[streamKey]; ok {
		stream.cancel()
		delete(l.streams, streamKey)
	}
}

// StopAllStreams stops all log streams
func (l *LogsService) StopAllStreams() {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	for _, stream := range l.streams {
		stream.cancel()
	}
	
	l.streams = make(map[string]*logStream)
}

// GetLogs retrieves logs for a pod (non-streaming)
func (l *LogsService) GetLogs(podName, containerName string, lines int64) ([]views.LogEntry, error) {
	logs, err := l.k8s.GetPodLogs(podName, containerName, lines, false)
	if err != nil {
		return nil, err
	}
	
	// Parse logs into entries
	entries := l.parseLogEntries(logs)
	
	// Store in buffer
	l.updateBuffer(podName, containerName, entries)
	
	return entries, nil
}

// GetBuffer returns the log buffer for a resource
func (l *LogsService) GetBuffer(podName, containerName string) []views.LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	bufferKey := l.getStreamKey(podName, containerName)
	if buffer, ok := l.buffers[bufferKey]; ok {
		return buffer.Entries
	}
	
	return nil
}

// ClearBuffer clears the log buffer for a resource
func (l *LogsService) ClearBuffer(podName, containerName string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	bufferKey := l.getStreamKey(podName, containerName)
	delete(l.buffers, bufferKey)
}

// ClearAllBuffers clears all log buffers
func (l *LogsService) ClearAllBuffers() {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	l.buffers = make(map[string]*LogBuffer)
}

// processLogStream processes logs from a stream
func (l *LogsService) processLogStream(stream *logStream) {
	batchSize := 10
	batchTimeout := 100 * time.Millisecond
	batch := make([]views.LogEntry, 0, batchSize)
	timer := time.NewTimer(batchTimeout)
	
	flushBatch := func() {
		if len(batch) > 0 {
			l.updateBuffer(stream.podName, stream.containerName, batch)
			l.notifyObservers(stream.podName, stream.containerName, batch)
			batch = batch[:0]
		}
	}
	
	for {
		select {
		case line, ok := <-stream.logChan:
			if !ok {
				flushBatch()
				return
			}
			
			entry := l.parseLogEntry(line)
			batch = append(batch, entry)
			
			if len(batch) >= batchSize {
				flushBatch()
				timer.Reset(batchTimeout)
			}
			
		case err, ok := <-stream.errChan:
			if ok && err != nil {
				l.notifyError(stream.podName, stream.containerName, err)
			}
			flushBatch()
			return
			
		case <-timer.C:
			flushBatch()
			timer.Reset(batchTimeout)
			
		case <-stream.context.Done():
			flushBatch()
			return
		}
	}
}

// parseLogEntries parses a log string into entries
func (l *LogsService) parseLogEntries(logs string) []views.LogEntry {
	lines := strings.Split(logs, "\n")
	entries := make([]views.LogEntry, 0, len(lines))
	
	for _, line := range lines {
		if line != "" {
			entries = append(entries, l.parseLogEntry(line))
		}
	}
	
	return entries
}

// parseLogEntry parses a single log line
func (l *LogsService) parseLogEntry(line string) views.LogEntry {
	// Try to parse timestamp
	timestamp := time.Now()
	logLine := line
	
	// Common timestamp formats (simplified)
	if len(line) > 30 {
		// Try to parse ISO timestamp
		if t, err := time.Parse(time.RFC3339, line[:30]); err == nil {
			timestamp = t
			logLine = strings.TrimSpace(line[30:])
		}
	}
	
	return views.LogEntry{
		Timestamp: timestamp,
		Line:      logLine,
		Level:     views.DetectLogLevel(logLine),
	}
}

// updateBuffer updates the log buffer
func (l *LogsService) updateBuffer(podName, containerName string, entries []views.LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	bufferKey := l.getStreamKey(podName, containerName)
	buffer, exists := l.buffers[bufferKey]
	
	if !exists {
		buffer = &LogBuffer{
			Resource:  podName,
			Container: containerName,
			Entries:   make([]views.LogEntry, 0, 1000),
			MaxSize:   1000,
		}
		l.buffers[bufferKey] = buffer
	}
	
	// Append entries
	buffer.Entries = append(buffer.Entries, entries...)
	
	// Trim if too large
	if len(buffer.Entries) > buffer.MaxSize {
		start := len(buffer.Entries) - buffer.MaxSize
		buffer.Entries = buffer.Entries[start:]
	}
}

// notifyObservers notifies observers of new log entries
func (l *LogsService) notifyObservers(podName, containerName string, entries []views.LogEntry) {
	l.mu.RLock()
	observers := make([]LogsObserver, len(l.observers))
	copy(observers, l.observers)
	l.mu.RUnlock()
	
	for _, observer := range observers {
		observer.OnLogsReceived(podName, containerName, entries)
	}
}

// notifyError notifies observers of a log error
func (l *LogsService) notifyError(podName, containerName string, err error) {
	l.mu.RLock()
	observers := make([]LogsObserver, len(l.observers))
	copy(observers, l.observers)
	l.mu.RUnlock()
	
	for _, observer := range observers {
		observer.OnLogError(podName, containerName, err)
	}
}

// getStreamKey returns a unique key for a pod/container combination
func (l *LogsService) getStreamKey(podName, containerName string) string {
	if containerName == "" {
		return podName
	}
	return fmt.Sprintf("%s/%s", podName, containerName)
}

// IsStreaming checks if logs are being streamed for a pod
func (l *LogsService) IsStreaming(podName, containerName string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	streamKey := l.getStreamKey(podName, containerName)
	_, exists := l.streams[streamKey]
	return exists
}

// GetActiveStreams returns the list of active log streams
func (l *LogsService) GetActiveStreams() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	streams := make([]string, 0, len(l.streams))
	for key := range l.streams {
		streams = append(streams, key)
	}
	
	return streams
}