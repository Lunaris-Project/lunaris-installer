package messages

import (
	"sync"
)

// Queue represents a message queue
type Queue struct {
	messages []Message
	maxSize  int
	mu       sync.Mutex
}

// NewQueue creates a new message queue
func NewQueue(maxSize int) *Queue {
	return &Queue{
		messages: make([]Message, 0),
		maxSize:  maxSize,
	}
}

// Add adds a message to the queue
func (q *Queue) Add(msg Message) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Add the message
	q.messages = append(q.messages, msg)

	// Trim the queue if it exceeds the maximum size
	if len(q.messages) > q.maxSize {
		// Keep the first quarter and last three quarters
		firstQuarter := q.maxSize / 4
		lastThreeQuarters := q.maxSize - firstQuarter - 1 // -1 for the truncation message

		truncatedMessages := make([]Message, 0, q.maxSize)
		truncatedMessages = append(truncatedMessages, q.messages[:firstQuarter]...)
		
		// Add a truncation message
		truncatedMessages = append(truncatedMessages, NewInfoMessage("... (messages truncated) ...", "system"))
		
		// Add the last three quarters
		truncatedMessages = append(truncatedMessages, q.messages[len(q.messages)-lastThreeQuarters:]...)
		
		q.messages = truncatedMessages
	}
}

// Get returns all messages
func (q *Queue) Get() []Message {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Return a copy of the messages
	messages := make([]Message, len(q.messages))
	copy(messages, q.messages)
	
	return messages
}

// GetLast returns the last n messages
func (q *Queue) GetLast(n int) []Message {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.messages) <= n {
		// Return a copy of all messages
		messages := make([]Message, len(q.messages))
		copy(messages, q.messages)
		return messages
	}

	// Return a copy of the last n messages
	messages := make([]Message, n)
	copy(messages, q.messages[len(q.messages)-n:])
	
	return messages
}

// Clear clears all messages
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.messages = make([]Message, 0)
}

// Size returns the number of messages
func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.messages)
}

// Filter returns messages that match the filter
func (q *Queue) Filter(filter func(Message) bool) []Message {
	q.mu.Lock()
	defer q.mu.Unlock()

	filtered := make([]Message, 0)
	for _, msg := range q.messages {
		if filter(msg) {
			filtered = append(filtered, msg)
		}
	}
	
	return filtered
}

// FilterByType returns messages of the specified type
func (q *Queue) FilterByType(msgType MessageType) []Message {
	return q.Filter(func(msg Message) bool {
		return msg.Type == msgType
	})
}

// FilterBySource returns messages from the specified source
func (q *Queue) FilterBySource(source string) []Message {
	return q.Filter(func(msg Message) bool {
		return msg.Source == source
	})
}
