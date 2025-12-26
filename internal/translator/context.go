package translator

import (
	"sync"
)

const MaxContextSize = 5

type ConversationContext struct {
	mu         sync.RWMutex
	utterances []Utterance
	maxSize    int
}

type Utterance struct {
	Original   string
	Translated string
	SourceLang string
	TargetLang string
}

func NewConversationContext() *ConversationContext {
	return &ConversationContext{
		utterances: make([]Utterance, 0, MaxContextSize),
		maxSize:    MaxContextSize,
	}
}

func (c *ConversationContext) Add(original, translated, sourceLang, targetLang string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	u := Utterance{
		Original:   original,
		Translated: translated,
		SourceLang: sourceLang,
		TargetLang: targetLang,
	}

	c.utterances = append(c.utterances, u)

	if len(c.utterances) > c.maxSize {
		c.utterances = c.utterances[len(c.utterances)-c.maxSize:]
	}
}

func (c *ConversationContext) GetRecent() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, 0, len(c.utterances)*2)
	for _, u := range c.utterances {
		result = append(result, u.Original)
		if u.Translated != "" {
			result = append(result, "→ "+u.Translated)
		}
	}
	return result
}

func (c *ConversationContext) GetRecentOriginals() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, len(c.utterances))
	for i, u := range c.utterances {
		result[i] = u.Original
	}
	return result
}

func (c *ConversationContext) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.utterances = c.utterances[:0]
}

func (c *ConversationContext) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.utterances)
}

type ContextManager struct {
	mu       sync.RWMutex
	sessions map[string]*ConversationContext
}

func NewContextManager() *ContextManager {
	return &ContextManager{
		sessions: make(map[string]*ConversationContext),
	}
}

func (m *ContextManager) Get(sessionID string) *ConversationContext {
	m.mu.RLock()
	ctx, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if exists {
		return ctx
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if ctx, exists = m.sessions[sessionID]; exists {
		return ctx
	}

	ctx = NewConversationContext()
	m.sessions[sessionID] = ctx
	return ctx
}

func (m *ContextManager) Remove(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}
