package gocraft_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
)

func TestInventoryLoadAndSlotBounds(t *testing.T) {
	stacks := make([]gocraft.ItemStack, gocraft.InventorySize)
	stacks[gocraft.SlotMainStart] = gocraft.ItemStack{
		Item:  7,
		Count: 3,
	}

	var inventory gocraft.Inventory
	inventory.Load(stacks)

	got := inventory.Slot(gocraft.SlotMainStart)
	if !got.Is(7) || got.Count != 3 {
		t.Errorf("slot = %+v, want item 7 count 3", got)
	}
	if !inventory.Slot(-1).Empty() || !inventory.Slot(gocraft.InventorySize).Empty() {
		t.Error("out of range slots should read as empty")
	}
}

func TestInventoryFindPrefersHotbar(t *testing.T) {
	var inventory gocraft.Inventory
	inventory.SetSlot(gocraft.SlotMainStart, gocraft.ItemStack{
		Item:  5,
		Count: 1,
	})
	inventory.SetSlot(gocraft.HotbarSlot(4), gocraft.ItemStack{
		Item:  5,
		Count: 1,
	})

	slot, ok := inventory.FindItem(5)
	if !ok {
		t.Fatal("item should be found")
	}
	if slot != gocraft.HotbarSlot(4) {
		t.Errorf("found slot %d, want hotbar slot %d", slot, gocraft.HotbarSlot(4))
	}
}

func TestInventoryCountSumsStorage(t *testing.T) {
	var inventory gocraft.Inventory
	inventory.SetSlot(gocraft.SlotMainStart, gocraft.ItemStack{
		Item:  9,
		Count: 30,
	})
	inventory.SetSlot(gocraft.HotbarSlot(0), gocraft.ItemStack{
		Item:  9,
		Count: 12,
	})
	inventory.SetSlot(gocraft.SlotOffhand, gocraft.ItemStack{
		Item:  9,
		Count: 1,
	})
	inventory.SetSlot(gocraft.SlotHead, gocraft.ItemStack{
		Item:  9,
		Count: 1,
	})

	if got := inventory.Count(9); got != 43 {
		t.Errorf("count = %d, want 43 (armor slots are not storage)", got)
	}
}

func TestInventorySwapAndHeld(t *testing.T) {
	var inventory gocraft.Inventory
	inventory.SetSlot(gocraft.SlotMainStart, gocraft.ItemStack{
		Item:  2,
		Count: 1,
	})

	inventory.Swap(gocraft.SlotMainStart, gocraft.HotbarSlot(0))
	if !inventory.Hotbar(0).Is(2) || !inventory.Slot(gocraft.SlotMainStart).Empty() {
		t.Error("swap should move the stack into the hotbar")
	}

	inventory.SelectHeld(0)
	if !inventory.Held().Is(2) {
		t.Error("held stack should follow the selected hotbar index")
	}

	inventory.SelectHeld(9)
	if inventory.HeldIndex() != 0 {
		t.Error("out of range held index should be ignored")
	}
}

func TestInventoryFirstEmpty(t *testing.T) {
	var inventory gocraft.Inventory
	for index := gocraft.SlotHotbarStart; index <= gocraft.SlotOffhand; index++ {
		inventory.SetSlot(index, gocraft.ItemStack{
			Item:  1,
			Count: 1,
		})
	}

	slot, ok := inventory.FirstEmpty()
	if !ok {
		t.Fatal("main storage should still have room")
	}
	if slot != gocraft.SlotMainStart {
		t.Errorf("first empty = %d, want %d", slot, gocraft.SlotMainStart)
	}
}
