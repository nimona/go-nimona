// Code generated by nimona.io/tools/objectify. DO NOT EDIT.

// +build !generate

package example

import (
	"github.com/mitchellh/mapstructure"
	"nimona.io/pkg/object"
)

const (
	InnerFooType = "test/inn"
)

// ToObject returns a f12n object
func (s InnerFoo) ToObject() object.Object {
	o := object.New()
	o.SetType(InnerFooType)
	if s.InnerBar != "" {
		o.SetRaw("inner_bar", s.InnerBar)
	}
	if len(s.InnerBars) > 0 {
		o.SetRaw("inner_bars", s.InnerBars)
	}
	if len(s.MoreInnerFoos) > 0 {
		o.SetRaw("inner_foos", s.MoreInnerFoos)
	}
	o.SetRaw("i", s.I)
	o.SetRaw("i8", s.I8)
	o.SetRaw("i16", s.I16)
	o.SetRaw("i32", s.I32)
	o.SetRaw("i64", s.I64)
	o.SetRaw("u", s.U)
	o.SetRaw("u8", s.U8)
	o.SetRaw("u16", s.U16)
	o.SetRaw("u32", s.U32)
	o.SetRaw("f32", s.F32)
	o.SetRaw("f64", s.F64)
	if len(s.Ai8) > 0 {
		o.SetRaw("ai8", s.Ai8)
	}
	if len(s.Ai16) > 0 {
		o.SetRaw("ai16", s.Ai16)
	}
	if len(s.Ai32) > 0 {
		o.SetRaw("ai32", s.Ai32)
	}
	if len(s.Ai64) > 0 {
		o.SetRaw("ai64", s.Ai64)
	}
	if len(s.Au16) > 0 {
		o.SetRaw("au16", s.Au16)
	}
	if len(s.Au32) > 0 {
		o.SetRaw("au32", s.Au32)
	}
	if len(s.Af32) > 0 {
		o.SetRaw("af32", s.Af32)
	}
	if len(s.Af64) > 0 {
		o.SetRaw("af64", s.Af64)
	}
	if len(s.AAi) > 0 {
		o.SetRaw("aAi", s.AAi)
	}
	if len(s.AAf) > 0 {
		o.SetRaw("aAf", s.AAf)
	}
	if len(s.AAs) > 0 {
		o.SetRaw("aAs", s.AAs)
	}
	o.SetRaw("b", s.B)
	return o
}

func anythingToAnythingForInnerFoo(
	from interface{},
	to interface{},
) error {
	config := &mapstructure.DecoderConfig{
		Result:  to,
		TagName: "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	if err := decoder.Decode(from); err != nil {
		return err
	}

	return nil
}

// FromObject populates the struct from a f12n object
func (s *InnerFoo) FromObject(o object.Object) error {
	atoa := anythingToAnythingForInnerFoo
	if err := atoa(o.GetRaw("inner_bar"), &s.InnerBar); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("inner_bars"), &s.InnerBars); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("inner_foos"), &s.MoreInnerFoos); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("i"), &s.I); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("i8"), &s.I8); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("i16"), &s.I16); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("i32"), &s.I32); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("i64"), &s.I64); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("u"), &s.U); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("u8"), &s.U8); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("u16"), &s.U16); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("u32"), &s.U32); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("f32"), &s.F32); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("f64"), &s.F64); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("ai8"), &s.Ai8); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("ai16"), &s.Ai16); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("ai32"), &s.Ai32); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("ai64"), &s.Ai64); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("au16"), &s.Au16); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("au32"), &s.Au32); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("af32"), &s.Af32); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("af64"), &s.Af64); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("aAi"), &s.AAi); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("aAf"), &s.AAf); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("aAs"), &s.AAs); err != nil {
		return err
	}
	if err := atoa(o.GetRaw("b"), &s.B); err != nil {
		return err
	}

	if ao, ok := interface{}(s).(interface{ afterFromObject() }); ok {
		ao.afterFromObject()
	}

	return nil
}

// GetType returns the object's type
func (s InnerFoo) GetType() string {
	return InnerFooType
}
