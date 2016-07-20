package jobs

type StoreResponse struct {
    Response string
}

type StoreAdapter interface {
    Insert(key string, value string) (StoreResponse, error)
    Update(key string, value string) (StoreResponse, error)
    Get(key string) (StoreResponse, error)
    Delete(key string) (StoreResponse, error)
    InsertWithOrdering(key string, value string) (StoreResponse, error)
}

