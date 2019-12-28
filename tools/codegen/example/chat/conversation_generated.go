// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package chat

import (
	json "encoding/json"

	object "nimona.io/pkg/object"
	schema "nimona.io/pkg/schema"
)

type (
	ConversationCreated struct {
		Name string `json:"name:s,omitempty"`
	}
	ConversationTopicUpdated struct {
		Topic string `json:"topic:s,omitempty"`
	}
	ConversationMessageAdded struct {
		Message *message `json:"message:o,omitempty"`
	}
	ConversationMessageRemoved struct {
		Message *message `json:"message:o,omitempty"`
	}
	MessageCreated struct {
		Body string `json:"body:s,omitempty"`
	}
)

func (e ConversationCreated) GetType() string {
	return "mochi.io/conversation.Created"
}

func (e ConversationCreated) GetSchema() *schema.Object {
	return &schema.Object{
		Properties: []*schema.Property{
			&schema.Property{
				Name:       "name",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
		},
		Links: []*schema.Link{},
	}
}

func (e ConversationCreated) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "mochi.io/conversation.Created"
	if e.Name != "" {
		m["name:s"] = e.Name
	}

	if schema := e.GetSchema(); schema != nil {
		m["$schema:o"] = schema.ToObject().ToMap()
	}
	return object.Object(m)
}

func (e *ConversationCreated) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e ConversationTopicUpdated) GetType() string {
	return "mochi.io/conversation.TopicUpdated"
}

func (e ConversationTopicUpdated) GetSchema() *schema.Object {
	return &schema.Object{
		Properties: []*schema.Property{
			&schema.Property{
				Name:       "topic",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
		},
		Links: []*schema.Link{},
	}
}

func (e ConversationTopicUpdated) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "mochi.io/conversation.TopicUpdated"
	if e.Topic != "" {
		m["topic:s"] = e.Topic
	}

	if schema := e.GetSchema(); schema != nil {
		m["$schema:o"] = schema.ToObject().ToMap()
	}
	return object.Object(m)
}

func (e *ConversationTopicUpdated) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e ConversationMessageAdded) GetType() string {
	return "mochi.io/conversation.MessageAdded"
}

func (e ConversationMessageAdded) GetSchema() *schema.Object {
	return &schema.Object{
		Properties: []*schema.Property{
			&schema.Property{
				Name:       "message",
				Type:       "mochi.io/message",
				Hint:       "o",
				IsRepeated: false,
				IsOptional: false,
			},
		},
		Links: []*schema.Link{},
	}
}

func (e ConversationMessageAdded) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "mochi.io/conversation.MessageAdded"
	if e.Message != nil {
		m["message:o"] = e.Message.ToObject().ToMap()
	}

	if schema := e.GetSchema(); schema != nil {
		m["$schema:o"] = schema.ToObject().ToMap()
	}
	return object.Object(m)
}

func (e *ConversationMessageAdded) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e ConversationMessageRemoved) GetType() string {
	return "mochi.io/conversation.MessageRemoved"
}

func (e ConversationMessageRemoved) GetSchema() *schema.Object {
	return &schema.Object{
		Properties: []*schema.Property{
			&schema.Property{
				Name:       "message",
				Type:       "mochi.io/message",
				Hint:       "o",
				IsRepeated: false,
				IsOptional: false,
			},
		},
		Links: []*schema.Link{},
	}
}

func (e ConversationMessageRemoved) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "mochi.io/conversation.MessageRemoved"
	if e.Message != nil {
		m["message:o"] = e.Message.ToObject().ToMap()
	}

	if schema := e.GetSchema(); schema != nil {
		m["$schema:o"] = schema.ToObject().ToMap()
	}
	return object.Object(m)
}

func (e *ConversationMessageRemoved) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e MessageCreated) GetType() string {
	return "mochi.io/message.Created"
}

func (e MessageCreated) GetSchema() *schema.Object {
	return &schema.Object{
		Properties: []*schema.Property{
			&schema.Property{
				Name:       "body",
				Type:       "string",
				Hint:       "s",
				IsRepeated: false,
				IsOptional: false,
			},
		},
		Links: []*schema.Link{},
	}
}

func (e MessageCreated) ToObject() object.Object {
	m := map[string]interface{}{}
	m["@type:s"] = "mochi.io/message.Created"
	if e.Body != "" {
		m["body:s"] = e.Body
	}

	if schema := e.GetSchema(); schema != nil {
		m["$schema:o"] = schema.ToObject().ToMap()
	}
	return object.Object(m)
}

func (e *MessageCreated) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}
