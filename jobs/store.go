package jobs

// EnumOptions defines the options for enumerators
type EnumOptions struct {
	PageSize  int
	Partition int
}

// Enumerable is a store which enumerates items
type Enumerable interface {
	Enumerate(EnumOptions) Enumerator
}

// Value is abstract form of stored object
type Value interface {
	Valid() bool
	Unmarshal(out interface{}) error
}

// Enumerator is used to enumerate items
type Enumerator interface {
	Next() ([]Value, error)
}

// KeyValueStore is simple K/V store
type KeyValueStore interface {
	Put(key string, value interface{}) error
	Get(key string) (Value, error)
	Remove(key string) (Value, error)
}

// PartitionedStore is the store partitions key/value pairs
// EnumOptions.Partition is used to scan keys
type PartitionedStore interface {
	Enumerable
	KeyValueStore
}

// Acquisition is the result of Acquire
type Acquisition interface {
	Acquired() bool
	Owner() string
	TTL() int
	Refresh() error
	Release()
}

// OrderedList is an enumerable list which obeys the order when
// keys are inserted
type OrderedList interface {
	Enumerable
	KeyValueStore
}

// Store is the persistent storage for jobs/tasks
type Store interface {
	// Bucket obtains a reference to a partitioned store
	Bucket(name string) PartitionedStore
	// OrderedList obtains a handle to an ordered list
	OrderedList(name string) OrderedList
	// Acquire acquires a lock
	Acquire(name string) Acquisition
}
