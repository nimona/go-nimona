package main

import (
	"sort"
	"sync"
	"time"

	"nimona.io/pkg/errors"
)

const (
	ErrNotFound = errors.Error("not found")
)

// models
type (
	Conversation struct {
		CID          string
		LastActivity time.Time
	}
	Message struct {
		CID             string
		ConversationCID string
		Body            string
		SenderKey       string
		Created         time.Time
	}
	Participant struct {
		Key             string
		ConversationCID string
		Nickname        string
		Updated         time.Time
	}
)

// store
type (
	conversations []*Conversation // helper, used for sorting
	messages      []*Message      // helper, used for sorting
	participants  []*Participant  // helper, used for sorting
	store         struct {
		conversations     conversations
		conversationsLock sync.RWMutex
		messages          map[string]messages
		messagesLock      sync.RWMutex
		participants      map[string]participants
		participantsLock  sync.RWMutex
	}
)

func NewMemoryStore() *store {
	return &store{
		conversations: conversations{},
		messages:      map[string]messages{},
		participants:  map[string]participants{},
	}
}

// store model helpers
func (a messages) Len() int {
	return len(a)
}

func (a messages) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a messages) Less(i, j int) bool {
	return a[i].Created.Before(a[j].Created)
}

func (a participants) Len() int {
	return len(a)
}

func (a participants) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a participants) Less(i, j int) bool {
	return a[i].Nickname < a[j].Nickname
}

func (s *store) GetConversations() ([]*Conversation, error) {
	s.conversationsLock.RLock()
	defer s.conversationsLock.RUnlock()
	return []*Conversation(s.conversations), nil
}

func (s *store) GetParticipants(conv string) ([]*Participant, error) {
	s.participantsLock.RLock()
	defer s.participantsLock.RUnlock()
	p, ok := s.participants[conv]
	if !ok {
		return nil, ErrNotFound
	}
	return []*Participant(p), nil
}

// GetMessages returns a conversation's messages in descending order.
func (s *store) GetMessages(conv string, limit, skip int) ([]*Message, error) {
	s.messagesLock.RLock()
	defer s.messagesLock.RUnlock()
	p, ok := s.messages[conv]
	if !ok {
		return nil, ErrNotFound
	}
	if limit == 0 && skip == 0 {
		return []*Message(p), nil
	}
	if skip > len(p) {
		return nil, nil
	}
	last := len(p) - skip
	first := len(p) - skip - limit
	if first < 0 {
		first = 0
	}
	return []*Message(p[first:last]), nil
}

func (s *store) PutConversation(con *Conversation) error {
	s.conversationsLock.Lock()
	defer s.conversationsLock.Unlock()
	s.conversations = append(
		s.conversations,
		con,
	)
	return nil
}

// PutParticipant adds a participant to a conversation and resorts all of them.
func (s *store) PutParticipant(par *Participant) error {
	s.participantsLock.Lock()
	defer s.participantsLock.Unlock()
	_, ok := s.participants[par.ConversationCID]
	if !ok {
		s.participants[par.ConversationCID] = []*Participant{}
	}
	for _, xpar := range s.participants[par.ConversationCID] {
		if xpar.Key == par.Key {
			if par.Nickname != "" {
				xpar.Nickname = par.Nickname
			}
			return nil
		}
	}
	s.participants[par.ConversationCID] = append(
		s.participants[par.ConversationCID],
		par,
	)
	sort.Sort(s.participants[par.ConversationCID])
	return nil
}

// PutMessage adds a message to a conversation and resorts all messages.
func (s *store) PutMessage(msg *Message) error {
	s.messagesLock.Lock()
	defer s.messagesLock.Unlock()
	_, ok := s.messages[msg.ConversationCID]
	if !ok {
		s.messages[msg.ConversationCID] = []*Message{}
	}
	for _, message := range s.messages[msg.ConversationCID] {
		if message.CID == msg.CID {
			return nil
		}
	}
	s.messages[msg.ConversationCID] = append(
		s.messages[msg.ConversationCID],
		msg,
	)
	sort.Sort(s.messages[msg.ConversationCID])
	return nil
}
