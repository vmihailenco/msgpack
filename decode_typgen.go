package msgpack

import (
	"reflect"
	"sync"
)

var (
	ptrGens    map[reflect.Type]*ptrGen
	ptrGenCap  = 4096
	ptrGensMux sync.RWMutex
)

func SetTypeGenSliceCap(n int) {
	ptrGensMux.Lock()
	defer ptrGensMux.Unlock()

	if n < 128 {
		n = 128
	}
	ptrGenCap = n
}

func getTypeGen(t reflect.Type) *ptrGen {
	ptrGensMux.RLock()
	g := ptrGens[t]
	ptrGensMux.RUnlock()
	if g != nil {
		return g
	}

	ptrGensMux.Lock()
	defer ptrGensMux.Unlock()
	if g = ptrGens[t]; g != nil {
		return g
	}

	g = &ptrGen{typ: t, cap: ptrGenCap, idx: -1}

	if ptrGens == nil {
		ptrGens = map[reflect.Type]*ptrGen{}
	}
	ptrGens[t] = g

	return g
}

func (d *Decoder) newValue(t reflect.Type) reflect.Value {
	if d.flags&usePreallocateValues == 0 {
		return reflect.New(t)
	}

	// log.Println("using typegen")

	return getTypeGen(t).next()
}

type ptrGen struct {
	sync.Mutex
	raw reflect.Value
	typ reflect.Type
	idx int
	cap int
}

func (p *ptrGen) next() (v reflect.Value) {
	p.Lock()
	defer p.Unlock()

	if p.idx == p.cap || p.idx == -1 {
		p.raw = reflect.MakeSlice(reflect.SliceOf(p.typ), p.cap, p.cap)
		p.idx = 0
	}

	v = p.raw.Index(p.idx).Addr()
	p.idx++

	return
}
