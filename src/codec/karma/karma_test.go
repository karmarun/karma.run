package karma

import (
	"github.com/karmarun/karma.run/common"
	"github.com/karmarun/karma.run/kvm/mdl"
	"github.com/karmarun/karma.run/kvm/val"
	"github.com/kr/pretty"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestKarmaCodec(t *testing.T) {

	for i := 3; i < 20; i++ {
		m := randomModel(i)
		pretty.Println(m)
		v := randomValue(i, m)
		pretty.Println(v)
		bs := Encode(v, nil)
		pretty.Println(bs)
		w, bs := Decode(bs)
		if len(bs) != 0 {
			t.Log("len(bs) != 0")
		}
		if !v.Equals(w) {
			t.Log("!v.Equals(w)")
		}
	}

}

func randomModel(r int) mdl.Model {
	if r == 0 {
		r = 1
	}
	l := rand.Int() % r // length parameter 0..(r-1)
	switch rand.Int() % 13 {
	case 0:
		return mdl.List{randomModel(r - 1)}
	case 1:
		return mdl.Map{randomModel(r - 1)}
	case 2:
		m := make(mdl.Tuple, l, l)
		for i := 0; i < l; i++ {
			m[i] = randomModel(r - 1)
		}
		return &m
	case 3:
		m := make(mdl.Struct, l)
		for i := 0; i < l; i++ {
			m[common.ExcelVariableName(i)] = randomModel(r - 1)
		}
		return &m
	case 4:
		if l == 0 {
			l = 1 // union with zero elements is illegal
		}
		m := make(mdl.Union, l)
		for i := 0; i < l; i++ {
			m[common.ExcelVariableName(i)] = randomModel(r - 1)
		}
		return &m
	case 5:
		return mdl.String{}
	case 6:
		return mdl.Int{}
	case 7:
		return mdl.Float{}
	case 8:
		return mdl.Bool{}
	case 9:
		return mdl.Any{}
	case 10:
		return mdl.DateTime{}
	case 11:
		return mdl.Ref{"abcdefghijklmnop"}
	case 12:
		return mdl.Optional{randomModel(r - 1)}
	}
	panic("never reached")
}

func randomValue(r int, m mdl.Model) val.Value {
	if r == 0 {
		r = 1
	}
	l := rand.Int() % r // length parameter 0..(r-1)
	switch m := m.(type) {
	case mdl.List:
		v := make(val.List, l, l)
		for i := 0; i < l; i++ {
			v[i] = randomValue(r-1, m.Elements)
		}
		return v
	case mdl.Map:
		v := make(val.Map, l)
		for i := 0; i < l; i++ {
			v[common.ExcelVariableName(i)] = randomValue(r-1, m.Elements)
		}
		return v
	case mdl.Tuple:
		v := make(val.Tuple, len(*m), len(*m))
		for i, m := range *m {
			v[i] = randomValue(r-1, m)
		}
		return v
	case mdl.Struct:
		v := make(val.Struct, len(*m))
		for k, m := range *m {
			v[k] = randomValue(r-1, m)
		}
		return v
	case mdl.Union:
		for k, m := range *m {
			return val.Union{Case: k, Value: randomValue(r-1, m)}
		}
	case mdl.String:
		return val.String(strings.Repeat("äöü", l*7)) // prime number
	case mdl.Int:
		return val.Int(rand.Int())
	case mdl.Float:
		return val.Float(rand.Int())
	case mdl.Bool:
		return val.Bool(l > (r / 2))
	case mdl.Any:
		return val.Raw{'0', '1', '2', '3'}
	case mdl.Ref:
		return val.Ref{"abcdefghijklmnop", "abcdefghijklmnop"}
	case mdl.DateTime:
		return val.DateTime{time.Now()}
	case mdl.Optional:
		if l > (r / 2) {
			return val.Null{}
		}
		return randomValue(r-1, m.Model)
	}
	panic("never reached")
}
