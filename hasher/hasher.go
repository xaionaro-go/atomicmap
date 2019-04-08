package hasher

type Hasher struct{}

func New() *Hasher {
	return &Hasher{}
}

func (h *Hasher) PreHash(key interface{}) (uint64, uint8, bool) {
	return PreHash(key)
}

func (h *Hasher) PreHashBytes(key []byte) (uint64, uint8, bool) {
	return PreHashBytes(key)
}

func (h *Hasher) PreHashUint64(key uint64) (uint64, uint8, bool) {
	return PreHashUint64(key)
}

func (h *Hasher) CompleteHash(keyPreHash uint64, keyTypeID uint8) uint64 {
	return CompleteHash(keyPreHash, keyTypeID)
}
func (h *Hasher) Hash(key interface{}) uint64 {
	return Hash(key)
}
