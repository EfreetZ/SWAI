package cluster

const TotalSlots = 16384

// SlotRange slot 范围。
type SlotRange struct {
	Start uint16
	End   uint16
}

// Contains 判断 slot 是否命中范围。
func (r SlotRange) Contains(slot uint16) bool {
	return slot >= r.Start && slot <= r.End
}
