package SimpleCQRS

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

type CommandHandler func(cmd Command) error
type EventProcessor func(cmd Event) error

type FakeBus struct {
	commandQueue    chan queuedCommand
	commandHandlers map[reflect.Type]CommandHandler
	eventProcessors map[reflect.Type][]EventProcessor
	induceDelay     bool
}

type queuedCommand struct {
	cmd                 Command
	synchronousResponse chan CommandProcessingError
}

func NewFakeBus(induceDelay bool) *FakeBus {
	fb := &FakeBus{
		commandQueue:    make(chan queuedCommand),
		commandHandlers: make(map[reflect.Type]CommandHandler),
		eventProcessors: make(map[reflect.Type][]EventProcessor),
		induceDelay:     induceDelay,
	}

	go fb.processCommands()
	return fb
}

func (fb *FakeBus) processCommands() {
	for {
		select {
		case cmdReq := <-fb.commandQueue:
			cmd := cmdReq.cmd
			fmt.Println("Processing command:", cmd)
			resp := cmdReq.synchronousResponse
			handler, _ := fb.commandHandlers[reflect.TypeOf(cmd)]

			if fb.induceDelay {
				time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second) // Have possible command race conditions too
			}

			result := handler(cmd)
			fmt.Println("Processed command, result:", result)
			select {
			case resp <- result:
			default:
			}
		}
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

func (fb *FakeBus) Dispatch(cmd Command, syncResp chan CommandProcessingError) CommandSubmissionError {
	if _, ok := fb.commandHandlers[reflect.TypeOf(cmd)]; ok {
		fmt.Println("Queuing command:", cmd)
		fb.commandQueue <- queuedCommand{cmd, syncResp}
		return nil
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

type CommandProcessingError error
type CommandSubmissionError error

type CommandDispatcher interface {
	Dispatch(e Command,
		synchronousResponse chan CommandProcessingError) CommandSubmissionError
}

type EventPublisher interface {
	Publish(e Event) error
}
