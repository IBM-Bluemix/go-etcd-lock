package lock

import (
	"fmt"
	"os"
	"sort"
	"time"

	ect "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"gopkg.in/errgo.v1"
)

type Error struct {
	hostname string
}

func (e *Error) Error() string {
	return fmt.Sprintf("key is already locked by %s", e.hostname)
}

type Locker interface {
	Acquire(key string, ttl uint64) (Lock, error)
	Wait(key string) error
}

type EtcdLocker struct {
	client ect.Client
	sync   bool
}

func NewEtcdLocker(client ect.Client, sync bool) Locker {
	return &EtcdLocker{client: client, sync: sync}
}

type Lock interface {
	Release() error
}

type EtcdLock struct {
	client ect.Client
	key    string
	index  uint64
}

func (locker *EtcdLocker) Acquire(key string, ttl uint64 /*, wait bool*/) (Lock, error) {
	hasLock := false
	key = addPrefix(key)
	lock, err := locker.addLockDirChild(key, ttl)
	if err != nil {
		return nil, errgo.Mask(err)
	}

	kapi := ect.NewKeysAPI(locker.client)
	ctx := context.Background()
	for !hasLock {
		res, err := kapi.Get(ctx, key, &ect.GetOptions{Recursive: true, Sort: true, Quorum: true})
		if err != nil {
			return nil, errgo.Mask(err)
		}

		if len(res.Node.Nodes) > 1 {
			sort.Sort(res.Node.Nodes)
			if res.Node.Nodes[0].CreatedIndex != lock.Node.CreatedIndex {
				kapi.Delete(ctx, lock.Node.Key, &ect.DeleteOptions{Recursive: false})
				return nil, &Error{res.Node.Nodes[0].Value}
			} else {
				// if the first index is the current one, it's our turn to lock the key
				hasLock = true
			}
		} else {
			// If there are only 1 node, it's our, lock is acquired
			hasLock = true
		}
	}

	// If we get the lock, set the ttl and return it
	_, err = kapi.Set(ctx, lock.Node.Key, lock.Node.Value, &ect.SetOptions{PrevExist: ect.PrevExist, TTL: time.Duration(ttl) * time.Second})
	if err != nil {
		return nil, errgo.Mask(err)
	}

	return &EtcdLock{locker.client, lock.Node.Key, lock.Node.CreatedIndex}, nil
}

func (locker *EtcdLocker) addLockDirChild(key string, ttl uint64) (*ect.Response, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, errgo.Notef(err, "fail to get hostname")
	}

	ctx := context.Background()
	if locker.sync {
		locker.client.Sync(ctx)
	}

	kapi := ect.NewKeysAPI(locker.client)
	return kapi.CreateInOrder(ctx, key, hostname, &ect.CreateInOrderOptions{TTL: time.Duration(ttl) * time.Second})
}
