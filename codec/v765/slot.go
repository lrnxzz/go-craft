package v765

import gocraft "github.com/lrnxzz/go-craft"

type Slot struct {
	Present gocraft.Bool
	Item    gocraft.VarInt
	Count   gocraft.Byte
	Data    gocraft.NBT
}

func slotOf(stack gocraft.ItemStack) Slot {
	if stack.Empty() {
		return Slot{}
	}

	return Slot{
		Present: true,
		Item:    gocraft.VarInt(stack.Item),
		Count:   gocraft.Byte(stack.Count),
		Data:    stack.Data,
	}
}

func (s Slot) Append(dst []byte) []byte {
	if !s.Present.Bool() {
		return s.Present.Append(dst)
	}

	return gocraft.AppendAll(dst, s.Present, s.Item, s.Count, s.Data)
}

func (s *Slot) Decode(r *gocraft.Reader) error {
	if err := s.Present.Decode(r); err != nil {
		return err
	}
	if !s.Present.Bool() {
		*s = Slot{}

		return nil
	}

	return gocraft.DecodeAll(r, &s.Item, &s.Count, &s.Data)
}

func (s Slot) Stack() gocraft.ItemStack {
	if !s.Present.Bool() {
		return gocraft.ItemStack{}
	}

	return gocraft.ItemStack{
		Item:  gocraft.ItemID(s.Item),
		Count: s.Count.Int(),
		Data:  s.Data,
	}
}

type ChangedSlot struct {
	Index gocraft.Short
	Item  Slot
}

func (c ChangedSlot) Append(dst []byte) []byte {
	return gocraft.AppendAll(dst, c.Index, c.Item)
}

func (c *ChangedSlot) Decode(r *gocraft.Reader) error {
	return gocraft.DecodeAll(r, &c.Index, &c.Item)
}
