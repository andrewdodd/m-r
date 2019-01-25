package SimpleCQRS

import "errors"

type EventStore interface {
	SaveEvents(aggregateId Guid, events []Event, expectedVersion int) error
	GetEventsForAggregate(aggregateId Guid) ([]Event, error)
}

type es struct {
	publisher EventPublisher
	current   map[Guid][]EventDescriptor
}

func NewEventStore(p EventPublisher) EventStore {
	return &es{p, make(map[Guid][]EventDescriptor)}
}

type EventDescriptor struct {
	data    Event
	id      Guid
	version int
}

func (e *es) SaveEvents(aggregateId Guid, events []Event, expectedVersion int) error {
	eventDescriptors, ok := e.current[aggregateId]

	if !ok {
		eventDescriptors = make([]EventDescriptor, 0)
		e.current[aggregateId] = eventDescriptors
	} else if expectedVersion != -1 {
		lastEvent := eventDescriptors[len(eventDescriptors)-1]
		if lastEvent.data.Version() != expectedVersion {
			return errors.New("Concurrency Error")
		}
	}

	i := expectedVersion

	// iterate through current aggregate events increasing version with each processed even
	for _, event := range events {
		i++
		event.SaveVersion(i)

		ed := EventDescriptor{data: event, id: aggregateId, version: i}
		// push event to the event descriptors list for current aggregate
		eventDescriptors = append(eventDescriptors, ed)

		// publish current event to the bus for further processing by subscribers
		e.publisher.Publish(event)
	}
	e.current[aggregateId] = eventDescriptors

	return nil
}

// collect all processed events for given aggregate and return them as a list
// used to build up an aggregate from its history (Domain.LoadsFromHistory)
func (e *es) GetEventsForAggregate(aggregateId Guid) ([]Event, error) {
	eventDescriptors, ok := e.current[aggregateId]

	if !ok {
		return nil, errors.New("Aggregate not found")
	}
	events := make([]Event, len(eventDescriptors))

	for i, ed := range eventDescriptors {
		events[i] = ed.data
	}

	return events, nil
}
