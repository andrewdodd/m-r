package SimpleCQRS

type Event interface {
	Version() int
	SaveVersion(v int)
}

type BaseEvent struct {
	version int
}

func (e *BaseEvent) Version() int {
	return e.version
}

func (e *BaseEvent) SaveVersion(v int) {
	e.version = v
}

type InventoryItemCreated struct {
	*BaseEvent
	id   Guid
	name string
}

func NewInventoryItemCreated(id Guid, name string) InventoryItemCreated {
	be := BaseEvent{}
	return InventoryItemCreated{id: id, name: name, BaseEvent: &be}
}
func (o InventoryItemCreated) Id() Guid {
	return o.id
}
func (o InventoryItemCreated) Name() string {
	return o.name
}

type InventoryItemDeactivated struct {
	*BaseEvent
	id Guid
}

func NewInventoryItemDeactivated(id Guid) InventoryItemDeactivated {
	be := BaseEvent{}
	return InventoryItemDeactivated{id: id, BaseEvent: &be}
}
func (o InventoryItemDeactivated) Id() Guid {
	return o.id
}

type InventoryItemRenamed struct {
	*BaseEvent
	id      Guid
	newName string
}

func NewInventoryItemRenamed(id Guid, newName string) InventoryItemRenamed {
	be := BaseEvent{}
	return InventoryItemRenamed{id: id, newName: newName, BaseEvent: &be}
}
func (o InventoryItemRenamed) Id() Guid {
	return o.id
}
func (o InventoryItemRenamed) NewName() string {
	return o.newName
}

type ItemsCheckedInToInventory struct {
	*BaseEvent
	id    Guid
	count int
}

func NewItemsCheckedInToInventory(id Guid, count int) ItemsCheckedInToInventory {
	be := BaseEvent{}
	return ItemsCheckedInToInventory{id: id, count: count, BaseEvent: &be}
}
func (o ItemsCheckedInToInventory) Id() Guid {
	return o.id
}
func (o ItemsCheckedInToInventory) Count() int {
	return o.count
}

type ItemsRemovedFromInventory struct {
	*BaseEvent
	id    Guid
	count int
}

func NewItemsRemovedFromInventory(id Guid, count int) ItemsRemovedFromInventory {
	be := BaseEvent{}
	return ItemsRemovedFromInventory{id: id, count: count, BaseEvent: &be}
}
func (o ItemsRemovedFromInventory) Id() Guid {
	return o.id
}
func (o ItemsRemovedFromInventory) Count() int {
	return o.count
}
