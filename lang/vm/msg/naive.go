package msg

type NaiveMessage struct {
	v interface{}
}

func Naive(v interface{}) Message {
	return &NaiveMessage{v: v}
}

func (msg *NaiveMessage) Orig() interface{} {
	return msg.v
}

func (msg *NaiveMessage) Member(k string) (Message, bool) {
	m, ok := msg.v.(map[string]interface{})
	if !ok {
		return nil, false
	}
	child, ok := m[k]
	if !ok {
		return nil, false
	}
	return &NaiveMessage{v: child}, true
}

func (msg *NaiveMessage) Each() ([]Message, bool) {
	elems, ok := msg.v.([]interface{})
	if !ok {
		return nil, false
	}
	var out []Message
	for _, el := range elems {
		out = append(out, &NaiveMessage{v: el})
	}
	return out, true
}

func (msg *NaiveMessage) Range(from, to int) ([]Message, bool) {
	elems, ok := msg.v.([]interface{})
	if !ok {
		return nil, false
	}
	if from >= len(elems) || to > len(elems) {
		return nil, false
	}
	var out []Message
	for _, el := range elems[from:to] {
		out = append(out, &NaiveMessage{v: el})
	}
	return out, true
}

func (msg *NaiveMessage) Index(i int) (Message, bool) {
	elems, ok := msg.v.([]interface{})
	if !ok {
		return nil, false
	}
	if i >= len(elems) {
		return nil, false
	}
	return &NaiveMessage{v: elems[i]}, true
}
