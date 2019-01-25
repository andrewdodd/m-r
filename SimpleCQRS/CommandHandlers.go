package SimpleCQRS

type InventoryCommandHandlers struct {
	repo InventoryItemRepository
}

func NewInventoryCommandHandlers(repo InventoryItemRepository) InventoryCommandHandlers {
	return InventoryCommandHandlers{repo}
}

func (r *InventoryCommandHandlers) HandleCreateInventoryItem(m Command) error {
	message := m.(CreateInventoryItem)
	item := NewInventoryItem(message.InventoryItemId, message.Name)
	return r.repo.Save(item, -1)
}

func (r *InventoryCommandHandlers) HandleDeactivateInventoryItem(m Command) error {
	message := m.(DeactivateInventoryItem)
	ar, _ := r.repo.GetById(message.InventoryItemId)
	item := ar.(*InventoryItem)
	item.Deactivate()
	return r.repo.Save(item, message.OriginalVersion)
}

func (r *InventoryCommandHandlers) HandleRemoveItemsFromInventory(m Command) error {
	message := m.(RemoveItemsFromInventory)
	ar, _ := r.repo.GetById(message.InventoryItemId)
	item := ar.(*InventoryItem)
	item.Remove(message.Count)
	return r.repo.Save(item, message.OriginalVersion)
}

func (r *InventoryCommandHandlers) HandleCheckInItemsToInventory(m Command) error {
	message := m.(CheckInItemsToInventory)
	ar, _ := r.repo.GetById(message.InventoryItemId)
	item := ar.(*InventoryItem)
	item.CheckIn(message.Count)
	return r.repo.Save(item, message.OriginalVersion)
}

func (r *InventoryCommandHandlers) HandleRenameInventoryItem(m Command) error {
	message := m.(RenameInventoryItem)
	ar, _ := r.repo.GetById(message.InventoryItemId)
	item := ar.(*InventoryItem)
	item.ChangeName(message.NewName)
	return r.repo.Save(item, message.OriginalVersion)
}
