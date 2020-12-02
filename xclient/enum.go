package xclient

// SelectMode defines the algorithm of selecting a services from candidates.
type SelectMode int

const (
	//RandomSelect is selecting randomly
	Random SelectMode = iota

	//RoundRobin is selecting by round robin
	RoundRobin

	//WeightedRoundRobin is selecting by weighted round robin
	WeightedRoundRobin

	//ConsistentHash is selecting by hashing
	ConsistentHash
)

