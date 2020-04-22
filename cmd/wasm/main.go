// +build js

package main

import (
	"bytes"
	"syscall/js"

	"github.com/qmuntal/go3mf"
	"github.com/qmuntal/go3mf/beamlattice"
	"github.com/qmuntal/go3mf/materials"
	"github.com/qmuntal/go3mf/production"
	"github.com/qmuntal/go3mf/slices"
)

func main() {
	fnUnmarshal3MF := js.FuncOf(unmarshal3MF)
	fnRelease := js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		go3mf.JSRelease()
		return nil
	})
	js.Global().Set("unmarshal3MF", fnUnmarshal3MF)
	js.Global().Set("go3mfRelease", fnRelease)
	select {}
	fnUnmarshal3MF.Release()
	fnRelease.Release()
}

func unmarshal3MF(this js.Value, inputs []js.Value) interface{} {
	model := new(go3mf.Model)
	model.WithSpec(&beamlattice.Spec{})
	model.WithSpec(&production.Spec{})
	model.WithSpec(&slices.Spec{})
	model.WithSpec(&materials.Spec{})
	data := inputs[0]
	buff := make([]uint8, data.Get("byteLength").Int())
	js.CopyBytesToGo(buff, data)
	d := go3mf.NewDecoder(bytes.NewReader(buff), int64(len(buff)))
	err := d.Decode(model)
	if err != nil {
		panic(err.Error())
	}
	return model
}
