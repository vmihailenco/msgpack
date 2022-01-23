package msgpack

import (
	"reflect"
)

func (d *Decoder) newValue(t reflect.Type) reflect.Value {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	gen := d.typeGen[t]

	if gen == nil {
		if d.typeGen == nil {
			d.typeGen = map[reflect.Type]*ptrGen{}
		}
		gen = &ptrGen{typ: t, cap: 4096}
		d.typeGen[t] = gen
	}

	return gen.next()
}

type ptrGen struct {
	raw reflect.Value
	typ reflect.Type
	idx int
	cap int
}

func (p *ptrGen) next() (v reflect.Value) {
	if p.idx == p.cap || !p.raw.IsValid() {
		p.raw = reflect.MakeSlice(reflect.SliceOf(p.typ), p.cap, p.cap)
		p.idx = 0
	}

	v = p.raw.Index(p.idx).Addr()
	p.idx++

	return
}
