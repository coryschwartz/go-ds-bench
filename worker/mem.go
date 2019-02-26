package worker

// #include <stdlib.h>
import "C"
import (
	"errors"
	"github.com/pbnjay/memory"
	"unsafe"
)

func constrainTo(newsize int64) (func(), error) {
	// <hack>
	total := memory.TotalMemory()
	toTake := int64(total) - newsize
	if toTake <= 1024 * 1024 * 128 { // leave 128M min
		return nil, errors.New("cannot create RAM from Go")
	}

	//  <hack terrifying="1">
	mptr := uintptr(C.malloc(C.ulong(toTake)))

	// make sure pages are actually initialized
	page := int64(4092) // usually the case
	for i := int64(0); i < toTake / page; i++ {
		C.GoBytes(unsafe.Pointer(mptr + uintptr(i * page)), 1)[0] = 42
	}

	//  </hack>

	return func() {
		C.free(unsafe.Pointer(mptr))
	}, nil
	// </hack>
}
