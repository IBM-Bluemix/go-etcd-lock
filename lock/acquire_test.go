package lock

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/errgo.v1"
)

func TestAcquire(t *testing.T) {
	locker := NewEtcdLocker(client(), true)
	Convey("A lock shouldn't be acquired twice", t, func() {
		lock, err := locker.Acquire("/lock", 10)
		defer lock.Release()
		So(err, ShouldBeNil)
		So(lock, ShouldNotBeNil)
		lock, err = locker.Acquire("/lock", 10)
		So(err, ShouldNotBeNil)
		So(errgo.Cause(err), ShouldHaveSameTypeAs, &Error{})
		So(lock, ShouldBeNil)
	})

	Convey("After expiration, a lock should be acquirable again", t, func() {
		lock, err := locker.Acquire("/lock-expire", 1)
		So(err, ShouldBeNil)
		So(lock, ShouldNotBeNil)

		time.Sleep(2 * time.Second)

		lock, err = locker.Acquire("/lock-expire", 1)
		So(err, ShouldBeNil)
		So(lock, ShouldNotBeNil)
		lock.Release()
	})
}
