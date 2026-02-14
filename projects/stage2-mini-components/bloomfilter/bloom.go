package bloomfilter

import (
	"hash/fnv"
	"math"
)

// Filter 是布隆过滤器。
type Filter struct {
	bits    []byte
	size    uint64
	hashNum uint64
	count   uint64
}

// New 创建布隆过滤器。
func New(expectedItems uint64, falsePositiveRate float64) *Filter {
	if expectedItems == 0 {
		expectedItems = 1
	}
	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		falsePositiveRate = 0.01
	}

	size := uint64(math.Ceil(-float64(expectedItems) * math.Log(falsePositiveRate) / (math.Ln2 * math.Ln2)))
	hashNum := uint64(math.Ceil((float64(size) / float64(expectedItems)) * math.Ln2))
	if hashNum == 0 {
		hashNum = 1
	}

	return &Filter{
		bits:    make([]byte, (size+7)/8),
		size:    size,
		hashNum: hashNum,
	}
}

// Add 添加元素。
func (f *Filter) Add(data []byte) {
	for i := uint64(0); i < f.hashNum; i++ {
		idx := f.hash(data, i) % f.size
		f.setBit(idx)
	}
	f.count++
}

// Contains 检查元素是否可能存在。
func (f *Filter) Contains(data []byte) bool {
	for i := uint64(0); i < f.hashNum; i++ {
		idx := f.hash(data, i) % f.size
		if !f.getBit(idx) {
			return false
		}
	}
	return true
}

// EstimateFalsePositiveRate 估算当前误判率。
func (f *Filter) EstimateFalsePositiveRate() float64 {
	if f.count == 0 {
		return 0
	}
	m := float64(f.size)
	k := float64(f.hashNum)
	n := float64(f.count)
	return math.Pow(1-math.Exp(-k*n/m), k)
}

func (f *Filter) hash(data []byte, seed uint64) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte{byte(seed), byte(seed >> 8), byte(seed >> 16), byte(seed >> 24)})
	_, _ = h.Write(data)
	return h.Sum64()
}

func (f *Filter) setBit(idx uint64) {
	byteIdx := idx / 8
	bitIdx := idx % 8
	f.bits[byteIdx] |= 1 << bitIdx
}

func (f *Filter) getBit(idx uint64) bool {
	byteIdx := idx / 8
	bitIdx := idx % 8
	return f.bits[byteIdx]&(1<<bitIdx) != 0
}
