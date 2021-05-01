package data

import(
	"bytes"
)

type Struct struct {
	Bindings map[StructKey]Value
	Error error
}

var EmptyStruct *Struct = NewStruct()

func NewStruct() *Struct {
	return &Struct{Bindings: make(map[StructKey]Value)}
}

// MakeStruct - create a new <struct> object from the arguments, which can be other structs, or key/value pairs
func MakeStruct(fieldvals []Value) (*Struct, error) {
	strct := &Struct{
		Bindings: make(map[StructKey]Value),
	}
	count := len(fieldvals)
	i := 0
	var bindings map[StructKey]Value
	for i < count {
		o := fieldvals[i]
		if p, ok := o.(*Instance); ok {
			o = p.Value
		}
		i++
		switch o.Type() {
		case StructType: // not a valid key, just copy bindings from it
			p := o.(*Struct)
			if bindings == nil {
				bindings = make(map[StructKey]Value, len(p.Bindings))
			}
			for k, v := range p.Bindings {
				bindings[k] = v
			}
		case StringType, SymbolType, KeywordType, TypeType:
			if i == count {
				return nil, NewError(ArgumentErrorKey, "Mismatched keyword/value in arglist: ", o)
			}
			if bindings == nil {
				bindings = make(map[StructKey]Value)
			}
			bindings[newStructKey(o)] = fieldvals[i]
			i++
		default:
			return nil, NewError(ArgumentErrorKey, "Bad struct key: ", o)
		}
	}
	if bindings == nil {
		strct.Bindings = make(map[StructKey]Value)
	} else {
		strct.Bindings = bindings
	}
	return strct, nil
}

   // Equal returns true if the object is equal to the argument
func (s1 *Struct) Equals(another Value) bool {
	if s2, ok := another.(*Struct); ok {
		bindings1 := s1.Bindings
		size := len(bindings1)
		bindings2 := s2.Bindings
		if size == len(bindings2) {
			for k, v := range bindings1 {
				v2, ok := bindings2[k]
				if !ok {
					return false
				}
				if !Equal(v, v2) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (d *Struct) Type() Value {
	return StructType
}

func (d *Struct) String() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	first := true
	for k, v := range d.Bindings {
		if first {
			first = false
		} else {
			buf.WriteString(" ")
		}
		buf.WriteString(k.Value)
		buf.WriteString(" ")
		buf.WriteString(v.String())
	}
	buf.WriteString("}")
	return buf.String()
}

type StructKey struct {
	Value string
	Type string
}

func newStructKey(key Value) StructKey {
	if IsValidStructKey(key) {
		return StructKey{key.String(), key.Type().String()}
	}
	return StructKey{}
}

func IsValidStructKey(d Value) bool {
	switch d.Type() {
	case StringType, SymbolType, KeywordType, TypeType:
		return true
	}
	return false
}

func (k StructKey) ToValue() Value {
	if k.Type == "<string>" {
		return NewString(k.Value)
	}
	return Intern(k.Value)
}

func (d *Struct) Value() Value {
	return d
}

func (data *Struct) Length() int {
	return len(data.Bindings)
}

func (strct *Struct) Get(key Value) Value {
	if IsValidStructKey(key) {
		k := newStructKey(key)
		result, ok := strct.Bindings[k]
		if ok {
			return result
		}
	}
	return Null
}

func (strct *Struct) Has(key Value) bool {
	tmp := strct.Get(key)
	return tmp != Null
}

func (strct *Struct) Put(key Value, val Value) *Struct {
	k := newStructKey(key)
	if k.Value == "" {
		//strct.Error = fmt.Errorf("Bad key for struct: %v", key)
		//I'd like to return an Error, but then the Put method cannot be chained. Unless Value had all methods. Maybe?
	} else {
		strct.Bindings[k] = val
	}
	return strct
}

func (strct *Struct) Unput(key Value) *Struct {
	k := newStructKey(key)
	delete(strct.Bindings, k)
	return strct
}
