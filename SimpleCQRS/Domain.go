package SimpleCQRS

import (
	"errors"
)

type Guid string

type AggregateRoot interface {
	Id() Guid
	GetUncommittedChanges() []Event
	MarkChangesAsCommitted()
	LoadsFromHistory(history []Event) error
	ApplyChange(e Event) error
	ApplyChangeInternal(e Event, isNew bool) error
}

type AggRoot struct {
	changes    []Event
	_version   int
	id         Guid
	InnerApply func(e Event) error
}

func NewEmptyAggRoot() AggRoot {
	ag := AggRoot{
		changes: make([]Event, 0),
	}
	return ag
}

func (ag *AggRoot) Id() Guid {
	return ag.id
}

func (ag *AggRoot) Version() int {
	return ag._version
}

func (ag *AggRoot) version(v int) {
	ag._version = v
}

func (ag *AggRoot) GetUncommittedChanges() []Event {
	return ag.changes
}
func (ag *AggRoot) MarkChangesAsCommitted() {
	ag.changes = make([]Event, 0)
}

func (ag *AggRoot) LoadsFromHistory(history []Event) error {
	for _, e := range history {
		err := ag.ApplyChangeInternal(e, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ag *AggRoot) ApplyChange(e Event) error {
	return ag.ApplyChangeInternal(e, true)
}

// push atomic aggregate changes to local history for further processing (EventStore.SaveEvents)
func (ag *AggRoot) ApplyChangeInternal(e Event, isNew bool) error {
	err := ag.InnerApply(e)
	if err != nil {
		return err
	}
	if isNew {
		ag.changes = append(ag.changes, e)
	}
	return nil
}

type InventoryItem struct {
	AggRoot
	activated bool
}

// used to create in repository ... many ways to avoid this, eg making private constructor
func NewEmptyInventoryItem() *InventoryItem {
	i := &(InventoryItem{
		AggRoot: NewEmptyAggRoot(),
	})
	i.AggRoot.InnerApply = i.HandleEvent
	return i
}

func NewInventoryItem(id Guid, name string) *InventoryItem {
	i := NewEmptyInventoryItem()
	i.ApplyChange(NewInventoryItemCreated(id, name))
	return i
}

func (ii *InventoryItem) ChangeName(newName string) error {
	if newName == "" {
		return errors.New("newName cannot be empty")
	}
	ii.AggRoot.ApplyChange(NewInventoryItemRenamed(ii.id, newName))
	return nil
}

func (ii *InventoryItem) Remove(count int) error {
	if count <= 0 {
		return errors.New("cannot remove negative count from inventory")
	}
	ii.AggRoot.ApplyChange(NewItemsRemovedFromInventory(ii.id, count))
	return nil
}

func (ii *InventoryItem) CheckIn(count int) error {
	if count <= 0 {
		return errors.New("must have a count greater than 0 to add to inventory")
	}
	ii.AggRoot.ApplyChange(NewItemsCheckedInToInventory(ii.id, count))
	return nil
}

func (ii *InventoryItem) Deactivate() error {
	if !ii.activated {
		return errors.New("already deactivated")
	}
	ii.AggRoot.ApplyChange(NewInventoryItemDeactivated(ii.id))
	return nil
}

func (ii *InventoryItem) HandleEvent(event Event) error {
	switch e := event.(type) {
	case InventoryItemCreated:
		ii.id = e.Id()
		ii.activated = true
	case InventoryItemDeactivated:
		ii.activated = false
	}
	return nil
}

type Repository interface {
	Save(ar AggregateRoot, expectedVersion int) error
	GetById(id Guid) (AggregateRoot, error)
}

type InventoryItemRepository struct {
	Storage EventStore
}

func (repo *InventoryItemRepository) Save(ar AggregateRoot, expectedVersion int) error {
	return repo.Storage.SaveEvents(ar.Id(),
		ar.GetUncommittedChanges(),
		expectedVersion)
}

func (repo *InventoryItemRepository) GetById(id Guid) (AggregateRoot, error) {
	obj := NewEmptyInventoryItem()
	events, err := repo.Storage.GetEventsForAggregate(id)
	if err != nil {
		return obj, err
	}
	obj.LoadsFromHistory(events)
	return obj, err
}
