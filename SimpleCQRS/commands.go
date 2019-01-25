package SimpleCQRS

type Command interface{}

type DeactivateInventoryItem struct {
	InventoryItemId Guid
	OriginalVersion int
}

type CreateInventoryItem struct {
	InventoryItemId Guid
	Name            string
}

type RenameInventoryItem struct {
	InventoryItemId Guid
	OriginalVersion int
	NewName         string
}
type CheckInItemsToInventory struct {
	InventoryItemId Guid
	OriginalVersion int
	Count           int
}

type RemoveItemsFromInventory struct {
	InventoryItemId Guid
	OriginalVersion int
	Count           int
}
