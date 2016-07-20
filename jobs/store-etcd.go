package jobs

import (
    client "github.com/coreos/etcd/clientv3"
)

type EtcdStore struct{
    Client client
}

func (etcd EtcdStore) Insert(key, value string) (StoreResponse, error) {
    return nil, nil
}

func (etcd EtcdStore) Update(key, value string) (StoreResponse, error) {
    return nil, nil
}

func (etcd EtcdStore) Get(key string) (StoreResponse, error) {
    return nil, nil
}

func (etcd EtcdStore) Delete(key string) (StoreResponse, error) {
    return nil, nil
}

func (etcd EtcdStore) InsertWithOrdering(key, value string) (StoreResponse, error) {
    return nil, nil
}