package cllbpriority

// Adjusted maximum values for i, j, and k to ensure unique encoding within the range.
const (
	MaxI = 20  // e.g: max projects that use the listener
	MaxJ = 100 // e.g: max deployed instances per project
	MaxK = 20  // e.g: services per instance

)

var (
	// Offset where priorities start, useful if there needs to be room before.
	Offset = 1000
	// MaxEncoded value can be used to create priorities after this block.
	MaxEncoded = MaxI*baseI + MaxJ*baseJ + MaxK + Offset
)

// Calculate base values for encoding.
const (
	baseK = 1                  // Base for k is 1 as it's the smallest unit
	baseJ = MaxK + 1           // Base for j to accommodate maximum of k
	baseI = (MaxJ + 1) * baseJ // Base for i to accommodate maximum of j
)

// ListenerPriority will calculate a unique priority number given three integers. If values go
// past their max the priority will wrap around and might collide.
func ListenerPriority(i, j, k int) int {
	i, j, k = i%(MaxI+1), j%(MaxJ+1), k%(MaxK+1)

	return i*baseI + j*baseJ + k + Offset
}
