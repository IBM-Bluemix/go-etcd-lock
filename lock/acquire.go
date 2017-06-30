package lock

import (
	"fmt"
	"os"
	"strings"
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

func (locker *EtcdLocker) Acquire(key string, ttl uint64) (Lock, error) {
	key = strings.Replace(key, "/", "_", -1)
	key = addPrefix(key)

	kapi := ect.NewKeysAPI(locker.client)
	ctx := context.Background()
	hostname, err := os.Hostname()
	if err != nil {
		return nil, errgo.Notef(err, "fail to get hostname")
	}
	_, err = kapi.Set(ctx, key, hostname, &ect.SetOptions{PrevExist: ect.PrevNoExist, TTL: time.Duration(ttl) * time.Second})
	if err != nil {
		if cErr, ok := err.(ect.Error); ok {
			if cErr.Code == ect.ErrorCodeNodeExist {
				return nil, &Error{"another host"}
			}
		}
		return nil, errgo.Mask(err)
	}
	return &EtcdLock{client: locker.client, key: key}, nil
}
