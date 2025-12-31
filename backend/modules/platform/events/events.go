package events

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventType defines the type of event
type EventType string

const (
	EventClusterCreated   EventType = "cluster.created"
	EventClusterUpdated   EventType = "cluster.updated"
	EventClusterDeleted   EventType = "cluster.deleted"
	EventClusterDeploying EventType = "cluster.deploying"
	EventClusterConnected EventType = "cluster.connected"
	EventClusterError     EventType = "cluster.error"

	EventHypervisorCreated   EventType = "hypervisor.created"
	EventHypervisorUpdated   EventType = "hypervisor.updated"
	EventHypervisorDeleted   EventType = "hypervisor.deleted"
	EventHypervisorDeploying EventType = "hypervisor.deploying"
	EventHypervisorConnected EventType = "hypervisor.connected"
	EventHypervisorError     EventType = "hypervisor.error"

	EventContainerEngineCreated   EventType = "container_engine.created"
	EventContainerEngineUpdated   EventType = "container_engine.updated"
	EventContainerEngineDeleted   EventType = "container_engine.deleted"
	EventContainerEngineConnected EventType = "container_engine.connected"
	EventContainerEngineError     EventType = "container_engine.error"

	// Firewall Security Events
	EventFirewallRuleCreated      EventType = "firewall_rule.created"
	EventFirewallRuleUpdated      EventType = "firewall_rule.updated"
	EventFirewallRuleDeleted      EventType = "firewall_rule.deleted"
	EventFirewallProfileCreated   EventType = "firewall_profile.created"
	EventFirewallProfileUpdated   EventType = "firewall_profile.updated"
	EventFirewallProfileDeleted   EventType = "firewall_profile.deleted"
	EventFirewallTemplateCreated  EventType = "firewall_template.created"
	EventFirewallTemplateUpdated  EventType = "firewall_template.updated"
	EventFirewallTemplateDeleted  EventType = "firewall_template.deleted"
	EventFirewallDeployStarted    EventType = "firewall_deploy.started"
	EventFirewallDeployCompleted  EventType = "firewall_deploy.completed"
	EventFirewallDeployFailed     EventType = "firewall_deploy.failed"
	EventFirewallRollbackStarted  EventType = "firewall_rollback.started"
	EventFirewallRollbackCompleted EventType = "firewall_rollback.completed"
)

// Event represents a domain event
type Event struct {
	ID         string                 `json:"id"`
	Type       EventType              `json:"type"`
	TenantID   uuid.UUID              `json:"tenantId"`
	ResourceID string                 `json:"resourceId"`
	Payload    map[string]interface{} `json:"payload"`
	Timestamp  time.Time              `json:"timestamp"`
}

// Handler is a function that handles an event
type Handler func(ctx context.Context, event Event)

// EventBus manages event subscriptions and publishing
type EventBus struct {
	mu       sync.RWMutex
	handlers map[EventType][]Handler
}

var globalBus *EventBus
var once sync.Once

// GetEventBus returns the global event bus singleton
func GetEventBus() *EventBus {
	once.Do(func() {
		globalBus = &EventBus{
			handlers: make(map[EventType][]Handler),
		}
	})
	return globalBus
}

// Subscribe registers a handler for an event type
func (b *EventBus) Subscribe(eventType EventType, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

// SubscribeAll registers a handler for all event types
func (b *EventBus) SubscribeAll(handler Handler) {
	// Subscribe to all known event types
	eventTypes := []EventType{
		EventClusterCreated, EventClusterUpdated, EventClusterDeleted,
		EventClusterDeploying, EventClusterConnected, EventClusterError,
		EventHypervisorCreated, EventHypervisorUpdated, EventHypervisorDeleted,
		EventHypervisorDeploying, EventHypervisorConnected, EventHypervisorError,
		EventContainerEngineCreated, EventContainerEngineUpdated, EventContainerEngineDeleted,
		EventContainerEngineConnected, EventContainerEngineError,
		EventFirewallRuleCreated, EventFirewallRuleUpdated, EventFirewallRuleDeleted,
		EventFirewallProfileCreated, EventFirewallProfileUpdated, EventFirewallProfileDeleted,
		EventFirewallTemplateCreated, EventFirewallTemplateUpdated, EventFirewallTemplateDeleted,
		EventFirewallDeployStarted, EventFirewallDeployCompleted, EventFirewallDeployFailed,
		EventFirewallRollbackStarted, EventFirewallRollbackCompleted,
	}

	for _, t := range eventTypes {
		b.Subscribe(t, handler)
	}
}

// Publish sends an event to all subscribed handlers
func (b *EventBus) Publish(ctx context.Context, event Event) {
	b.mu.RLock()
	handlers := b.handlers[event.Type]
	b.mu.RUnlock()

	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	for _, handler := range handlers {
		go func(h Handler) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Events] Handler panic for event %s: %v", event.Type, r)
				}
			}()
			h(ctx, event)
		}(handler)
	}
}

// PublishAsync publishes an event asynchronously
func (b *EventBus) PublishAsync(event Event) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		b.Publish(ctx, event)
	}()
}

// NewEvent creates a new event
func NewEvent(eventType EventType, tenantID uuid.UUID, resourceID string, payload map[string]interface{}) Event {
	return Event{
		ID:         uuid.New().String(),
		Type:       eventType,
		TenantID:   tenantID,
		ResourceID: resourceID,
		Payload:    payload,
		Timestamp:  time.Now(),
	}
}

// ToJSON converts the event to JSON
func (e Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
