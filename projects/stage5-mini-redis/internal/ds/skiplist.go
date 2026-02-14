package ds

import "sort"

// ZItem 有序集合项。
type ZItem struct {
	Member string
	Score  float64
}

// SkipList 简化跳表实现（底层切片有序维护）。
type SkipList struct {
	items map[string]float64
}

// NewSkipList 创建结构。
func NewSkipList() *SkipList {
	return &SkipList{items: make(map[string]float64)}
}

// Insert 插入成员。
func (s *SkipList) Insert(member string, score float64) {
	s.items[member] = score
}

// Delete 删除成员。
func (s *SkipList) Delete(member string) bool {
	if _, ok := s.items[member]; !ok {
		return false
	}
	delete(s.items, member)
	return true
}

// RangeByScore 按分数区间查询。
func (s *SkipList) RangeByScore(min, max float64) []ZItem {
	result := make([]ZItem, 0)
	for member, score := range s.items {
		if score >= min && score <= max {
			result = append(result, ZItem{Member: member, Score: score})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Score == result[j].Score {
			return result[i].Member < result[j].Member
		}
		return result[i].Score < result[j].Score
	})
	return result
}
