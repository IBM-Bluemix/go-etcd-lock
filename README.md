etcd-lock
=========


[![Build Status](https://travis-ci.org/IBM-Bluemix/go-etcd-lock.svg?branch=master)](https://travis-ci.org/IBM-Bluemix/go-etcd-lock)
[![Coverage Status](https://coveralls.io/repos/github/IBM-Bluemix/go-etcd-lock/badge.svg?branch=master)](https://coveralls.io/github/IBM-Bluemix/go-etcd-lock?branch=master)

This is a basic client implementation of lock based on the logics in mod/lock

This library doesn't provide the `*etcd.Client` because it doesn't want to
manage the configuration of it (TLS or not, endpoints etc.) So a client has to
exist previously

Import
------

```
# Master via standard import
get get github.com/IBM-Bluemix/go-etcd-lock/lock
```

Example
-------

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/IBM-Bluemix/go-etcd-lock/lock"
	"github.com/coreos/etcd/client"
)

func main() {
	cfg := client.Config{
		Endpoints: []string{"http://127.0.0.1:2379"},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	locker := lock.NewEtcdLocker(&c)
	l, err := locker.Acquire("/foo", 60)

	if lockErr, ok := err.(*lock.Error); ok {
		// Key already locked
		fmt.Println(lockErr)
		return
	} else if err != nil {
		// Communication with etcd has failed or other error
		panic(err)
	}

	// It's ok, lock is granted for 60 secondes

	// When the opration is done we release the lock
	err = l.Release()
	if err != nil {
		// Something wrong can happen during release: connection problem with etcd
		panic(err)
	}
}

```

Testing
-------

You need a etcd instance running on `localhost:2379`, then:

```
go test ./...
```
