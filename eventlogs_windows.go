// +build windows

package main

import (
	"fmt"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const (
	replacementChar = '\uFFFD'     // Unicode replacement character
	maxRune         = '\U0010FFFF' // Maximum valid Unicode code point.
)

const (
	// 0xd800-0xdc00 encodes the high 10 bits of a pair.
	// 0xdc00-0xe000 encodes the low 10 bits of a pair.
	// the value is those 20 bits plus 0x10000.
	surr1 = 0xd800
	surr2 = 0xdc00
	surr3 = 0xe000

	surrSelf = 0x10000
)

const (
	EvtQueryChannelPath         = 0x1
	EvtQueryFilePath            = 0x2
	EvtQueryForwardDirection    = 0x100
	EvtQueryReverseDirection    = 0x200
	EvtQueryTolerateQueryErrors = 0x1000
)

var (
	modwevtapi = syscall.NewLazyDLL("wevtapi.dll")

	procEvtRender = modwevtapi.NewProc("EvtRender")
	procEvtQuery  = modwevtapi.NewProc("EvtQuery")
	procEvtClose  = modwevtapi.NewProc("EvtClose")
	procEvtNext   = modwevtapi.NewProc("EvtNext")
)

type EvtHandle uintptr

func _EvtNext(resultSet EvtHandle, eventArraySize uint32, eventArray *EvtHandle, timeout uint32, flags uint32, numReturned *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procEvtNext.Addr(), 6, uintptr(resultSet), uintptr(eventArraySize), uintptr(unsafe.Pointer(eventArray)), uintptr(timeout), uintptr(flags), uintptr(unsafe.Pointer(numReturned)))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _EvtRender(context EvtHandle, fragment EvtHandle, flags EvtRenderFlag, bufferSize uint32, buffer *byte, bufferUsed *uint32, propertyCount *uint32) (err error) {
	r1, _, e1 := syscall.Syscall9(procEvtRender.Addr(), 7, uintptr(context), uintptr(fragment), uintptr(flags), uintptr(bufferSize), uintptr(unsafe.Pointer(buffer)), uintptr(unsafe.Pointer(bufferUsed)), uintptr(unsafe.Pointer(propertyCount)), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _EvtClose(object EvtHandle) (err error) {

	r1, _, e1 := syscall.Syscall(procEvtClose.Addr(), 1, uintptr(object), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func _EvtQuery(query string) (handle EvtHandle, err error) {
	r0, _, e1 := syscall.Syscall9(procEvtQuery.Addr(), 3, 0, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(query))), EvtQueryChannelPath|EvtQueryReverseDirection, 0, 0, 0, 0, 0)
	handle = EvtHandle(r0)
	if handle == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func closeEvtHandles(eventNextHandles []EvtHandle) {
	//fmt.Println("closeEvtHandles", len(eventNextHandles))
	for i := 0; i < len(eventNextHandles); i++ {
		//fmt.Println(eventNextHandles[i])
		_EvtClose(eventNextHandles[i])
	}
}

func _ReadEventLogs(query string, maxItemsToRead int) (result []string, err error) {

	queryEvtHandle, err := _EvtQuery(query)
	defer _EvtClose(queryEvtHandle)

	if err != nil {
		return nil, err
	}

	var nextEventsRead, renderBufferUsed, renderBytesRead uint32
	var renderBuffer [10 * 1024]byte

	eventList := make([]string, maxItemsToRead)
	var eventListIdx int = 0
	var eventNextHandles = make([]EvtHandle, maxItemsToRead)

	err = _EvtNext(queryEvtHandle, uint32(len(eventNextHandles)), &eventNextHandles[0], 10*1000 /*timeout(ms)*/, 0 /*flags*/, &nextEventsRead)
	if nextEventsRead < 1 {
		if err.Error() == `No more data is available.` {
			err = nil
		}
	}

	if err == nil {
		for i := uint32(0); i < nextEventsRead; i++ {
			evtHandle := eventNextHandles[i]
			//defer _EvtClose(evtHandle)

			err = _EvtRender(0, evtHandle, EvtRenderEventXml, uint32(len(renderBuffer)), &renderBuffer[0], &renderBufferUsed, &renderBytesRead)

			var event string
			if renderBufferUsed > uint32(len(renderBuffer)) {
				alternativeBuffer := make([]byte, renderBufferUsed)
				err = _EvtRender(0, evtHandle, EvtRenderEventXml, uint32(len(alternativeBuffer)), &alternativeBuffer[0], &renderBufferUsed, &renderBytesRead)
				event, _, err = UTF16BytesToString(alternativeBuffer[:renderBufferUsed])
				if err != nil {
					eventList = append(eventList, event)
				}
			} else if err != nil {
				return nil, err
			} else {
				event, _, err = UTF16BytesToString(renderBuffer[:renderBufferUsed])
			}

			eventList[eventListIdx] = event
			eventListIdx++
		}
	}

	closeEvtHandles(eventNextHandles[:nextEventsRead])

	return eventList[:eventListIdx], err
}

func UTF16BytesToString(b []byte) (string, int, error) {
	if len(b)%2 != 0 {
		return "", 0, fmt.Errorf("Slice must have an even length (length=%d)", len(b))
	}

	offset := -1

	if nullIndex := indexNullTerminator(b); nullIndex > -1 {
		if len(b) > nullIndex+2 {
			offset = nullIndex + 2
		}

		b = b[:nullIndex]
	}

	s := make([]uint16, len(b)/2)
	for i := range s {
		s[i] = uint16(b[i*2]) + uint16(b[(i*2)+1])<<8
	}

	return string(utf16.Decode(s)), offset, nil
}

func indexNullTerminator(b []byte) int {
	if len(b) < 2 {
		return -1
	}

	for i := 0; i < len(b); i += 2 {
		if b[i] == 0 && b[i+1] == 0 {
			return i
		}
	}

	return -1
}
