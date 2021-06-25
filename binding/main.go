package main

// #include <stdint.h>
// #include <stdlib.h>
// typedef struct { void* message; int size; char* error; } BytesReturn;
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/rs/xid"

	"nimona.io/pkg/chore"
	"nimona.io/pkg/context"
	"nimona.io/pkg/object"
	"nimona.io/pkg/version"
)

var (
	nimonaProvider     *Provider
	subscriptionsMutex sync.RWMutex
	subscriptions      map[string]object.ReadCloser
)

func renderBytes(b []byte, err error) *C.BytesReturn {
	r := (*C.BytesReturn)(C.malloc(C.size_t(C.sizeof_BytesReturn)))
	if err != nil {
		fmt.Println("++ renderBytes() ERROR", err)
		r.error = C.CString(err.Error())
		return r
	}
	r.error = nil
	r.message = C.CBytes(b)
	r.size = C.int(len(b))
	return r
}

func marshalObject(o *object.Object) ([]byte, error) {
	m, err := object.Copy(o).MarshalMap()
	if err != nil {
		return nil, err
	}
	m["_hash"] = o.Hash()
	return json.Marshal(m)
}

func renderObject(o *object.Object) *C.BytesReturn {
	b, err := marshalObject(o)
	fmt.Println("++ renderObject() RESP body=", string(b))
	return renderBytes(b, err)
}

//export NimonaBridgeCall
func NimonaBridgeCall(
	name *C.char,
	payload unsafe.Pointer,
	payloadSize C.int,
) *C.BytesReturn {
	nameString := C.GoString(name)
	payloadBytes := C.GoBytes(payload, payloadSize)
	fmt.Printf("++ Called %s with %s\n", nameString, string(payloadBytes))

	switch nameString {
	case "init":
		fmt.Println("++ Call(get) RESP version=", version.Version)
		if nimonaProvider != nil {
			return renderBytes(nil, nil)
		}
		req := &InitRequest{}
		if err := json.Unmarshal(payloadBytes, req); err != nil {
			return renderBytes(nil, err)
		}
		nimonaProvider = New(req)
		subscriptionsMutex = sync.RWMutex{}
		subscriptions = map[string]object.ReadCloser{}
		return renderBytes([]byte("ok"), nil)
	case "get":
		ctx := context.New(
			context.WithTimeout(3 * time.Second),
		)
		req := GetRequest{}
		if err := json.Unmarshal(payloadBytes, &req); err != nil {
			return renderBytes(nil, err)
		}
		r, err := nimonaProvider.Get(ctx, req)
		if err != nil {
			fmt.Println("++ Call(get) ERROR", err)
			return renderBytes(nil, err)
		}
		os := []string{}
		for {
			o, err := r.Read()
			if err != nil || o == nil {
				break
			}
			b, err := marshalObject(o)
			if err != nil {
				fmt.Println("++ Call(err) ERROR", err)
				return renderBytes(nil, err)
			}
			os = append(os, string(b))
		}
		res := &GetResponse{
			ObjectBodies: os,
		}
		b, err := json.Marshal(res)
		fmt.Println("++ Call(get) RESP body=", string(b))
		return renderBytes(b, err)
	case "version":
		fmt.Println("++ Call(get) RESP version=", version.Version)
		return renderBytes([]byte(version.Version), nil)
	case "subscribe":
		ctx := context.New(
			context.WithTimeout(3 * time.Second),
		)
		req := SubscribeRequest{}
		if err := json.Unmarshal(payloadBytes, &req); err != nil {
			return renderBytes(nil, err)
		}
		r, err := nimonaProvider.Subscribe(ctx, req)
		if err != nil {
			fmt.Println("++ Call(subscribe) ERROR", err)
			return renderBytes(nil, err)
		}
		key := xid.New().String()
		subscriptionsMutex.Lock()
		subscriptions[key] = r
		subscriptionsMutex.Unlock()
		fmt.Println("++ Call(subscribe) RESP key=", key)
		return renderBytes([]byte(key), nil)
	case "pop":
		subscriptionsMutex.RLock()
		r, ok := subscriptions[string(payloadBytes)]
		if !ok {
			return renderBytes(nil, errors.New("missing subscription key"))
		}
		subscriptionsMutex.RUnlock()
		o, err := r.Read()
		if err != nil {
			fmt.Println("++ Call(pop) ERROR", err)
			return renderBytes(nil, err)
		}
		return renderObject(o)
	case "cancel":
		subscriptionsMutex.Lock()
		r, ok := subscriptions[string(payloadBytes)]
		if !ok {
			return renderBytes(nil, errors.New("missing subscription key"))
		}
		r.Close()
		delete(subscriptions, string(payloadBytes))
		subscriptionsMutex.Unlock()
		return renderBytes(nil, nil)
	case "requestStream":
		ctx := context.New(
			context.WithTimeout(10 * time.Second),
		)
		if err := nimonaProvider.RequestStream(
			ctx,
			chore.Hash(string(payloadBytes)),
		); err != nil {
			return renderBytes(nil, err)
		}
		return renderBytes(nil, nil)
	case "put":
		ctx := context.New(
			context.WithTimeout(3 * time.Second),
		)
		o := &object.Object{}
		if err := json.Unmarshal(payloadBytes, o); err != nil {
			return renderBytes(nil, err)
		}
		u, err := nimonaProvider.Put(ctx, o)
		if err != nil {
			fmt.Println("++ Call(put) ERROR", err)
			return renderBytes(nil, err)
		}
		return renderObject(u)
	case "getFeedRootHash":
		feedRootHash := nimonaProvider.GetFeedRootHash(string(payloadBytes))
		return renderBytes([]byte(feedRootHash), nil)
	case "getConnectionInfo":
		o, _ := object.Marshal(nimonaProvider.GetConnectionInfo())
		return renderObject(o)
	}

	return renderBytes([]byte("error"), errors.New(nameString+" not implemented"))
}

// Unused
func main() {}
