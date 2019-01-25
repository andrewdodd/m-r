package SimpleCQRS

import (
	"errors"
	"math/rand"
	"reflect"
	"time"
)

type CommandHandler func(cmd Command) error
type EventProcessor func(cmd Event) error

type FakeBus struct {
	commandHandlers map[reflect.Type]CommandHandler
	eventProcessors map[reflect.Type][]EventProcessor
	induceDelay     bool
}

func NewFakeBus(induceDelay bool) FakeBus {
	return FakeBus{
		commandHandlers: make(map[reflect.Type]CommandHandler),
		eventProcessors: make(map[reflect.Type][]EventProcessor),
		induceDelay:     induceDelay,
	}
}

func (fb *FakeBus) SetCommandHandler(cmdType reflect.Type, handler CommandHandler) error {
	if _, ok := fb.commandHandlers[cmdType]; ok {
		return errors.New("command handler already registered")
	}
	fb.commandHandlers[cmdType] = handler
	return nil
}

func (fb *FakeBus) AddEventProcessor(eventType reflect.Type, processor EventProcessor) error {
	processors, ok := fb.eventProcessors[eventType]
	if !ok {
		processors = make([]EventProcessor, 0)
	}
	for _, p := range processors {
		if reflect.DeepEqual(p, processor) {
			return errors.New("processor already registered")
		}
	}
	processors = append(processors, processor)
	fb.eventProcessors[eventType] = processors
	return nil
}

func (fb *FakeBus) Dispatch(cmd Command) error {
	if handler, ok := fb.commandHandlers[reflect.TypeOf(cmd)]; ok {
		if fb.induceDelay {
			time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second) // Have possible command race conditions too
		}
		return handler(cmd)
	}
	return errors.New("no handler registered")
}

func (fb *FakeBus) Publish(evt Event) error {
	if processors, ok := fb.eventProcessors[reflect.TypeOf(evt)]; ok {
		for _, processor := range processors {
			go func(p EventProcessor) {
				if fb.induceDelay {
					time.Sleep(time.Duration(rand.Intn(10)) * time.Second) // Have a variable degree of eventual consistency
				}
				p(evt)
			}(processor)
		}
		return nil
	}
	return errors.New("no processor registered")
}

type CommandDispatcher interface {
	Dispatch(e Command) error
}

type EventPublisher interface {
	Publish(e Event) error
}
