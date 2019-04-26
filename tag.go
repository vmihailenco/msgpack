package msgpack

import (
	"strings"
)

type tagOptions string

func (o tagOptions) Get(name string) (string, bool) {
	s := string(o)
	for len(s) > 0 {
		var next string
		idx := strings.IndexByte(s, ',')
		if idx >= 0 {
			s, next = s[:idx], s[idx+1:]
		}
		if strings.HasPrefix(s, name) {
			return s[len(name):], true
		}
		s = next
	}
	return "", false
}

func (o tagOptions) Contains(name string) bool {
	_, ok := o.Get(name)
	return ok
}

func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, tagOptions("")
}
