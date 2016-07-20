package jobs

type StoreConfig struct {
    Adapter string
    Name string
    Endpoints []string
    Username string
    Password string
}

type Store struct {
    Config StoreConfig
}

type KVStore interface {
    Put(storeId string, key string, value string) error
    Get(storeId string, key string) (string, error)
    Del(storeId string, key string) error
}

type ListStore interface {
    Put(key string, value string) error
    Get(key string) (string, error)
    Del(key string) error
    GetList(size int) ([]string, error)
    GetAll() ([]string, error)
}

type ClaimStore interface {
    Claim(key string, ownerId string) error
    Unclaim(key string) error
}

type kvStore struct {
    StoreAdapter StoreAdapter
    Name string
}

type listStore struct {
    StoreAdapter StoreAdapter
    Name string
}

type claimStore struct {
    StoreAdapter StoreAdapter
    Name string
}

func NewStore(config StoreConfig) Store {
    return Store{Config: config}
}

func (s Store) CreateKVStore(name string) KVStore {
    return nil
}

func (s Store) CreateListStore(name string) ListStore {
    return nil
}

func (s Store) CreateClaimStore(name string) ClaimStore {
    return nil
}
