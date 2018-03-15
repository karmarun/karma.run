// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package kvm

import (
	"bytes"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/kr/pretty"
	"hash"
	"hash/fnv" // FNV-1 has a very low collision rate
	"karma.run/cc"
	"karma.run/codec/karma.v2"
	"karma.run/common"
	"karma.run/definitions"
	"karma.run/kvm/err"
	"karma.run/kvm/inst"
	"karma.run/kvm/mdl"
	"karma.run/kvm/val"
	"karma.run/kvm/xpr"
	"log"
	"time"
)

const SeparatorByte = '~'

type VirtualMachine struct {
	UserID     string
	RootBucket *bolt.Bucket

	permissions    *permissions
	permRecursions map[string]struct{}

	cache struct {
		UserModelId       string
		RoleModelId       string
		ExpressionModelId string
		MetaModelId       string
		TagModelId        string
		MigrationModelId  string
	}
}

func (vm VirtualMachine) RootUserId() string {
	return string(vm.RootBucket.Get(definitions.RootUserBytes))
}

func (vm VirtualMachine) isDefaultModelId(mid string) bool {
	return mid == vm.ExpressionModelId() ||
		mid == vm.MetaModelId() ||
		mid == vm.MigrationModelId() ||
		mid == vm.RoleModelId() ||
		mid == vm.TagModelId() ||
		mid == vm.UserModelId()
}

func (vm *VirtualMachine) UserModelId() string {
	if vm.cache.UserModelId != "" {
		return vm.cache.UserModelId
	}
	s := string(vm.RootBucket.Get(definitions.UserModelBytes))
	vm.cache.UserModelId = s
	return s
}

func (vm *VirtualMachine) RoleModelId() string {
	if vm.cache.RoleModelId != "" {
		return vm.cache.RoleModelId
	}
	s := string(vm.RootBucket.Get(definitions.RoleModelBytes))
	vm.cache.RoleModelId = s
	return s
}

func (vm *VirtualMachine) ExpressionModelId() string {
	if vm.cache.ExpressionModelId != "" {
		return vm.cache.ExpressionModelId
	}
	s := string(vm.RootBucket.Get(definitions.ExpressionModelBytes))
	vm.cache.ExpressionModelId = s
	return s
}

func (vm *VirtualMachine) MetaModelId() string {
	if vm.cache.MetaModelId != "" {
		return vm.cache.MetaModelId
	}
	s := string(vm.RootBucket.Get(definitions.MetaModelBytes))
	vm.cache.MetaModelId = s
	return s
}

func (vm *VirtualMachine) TagModelId() string {
	if vm.cache.TagModelId != "" {
		return vm.cache.TagModelId
	}
	s := string(vm.RootBucket.Get(definitions.TagModelBytes))
	vm.cache.TagModelId = s
	return s
}

func (vm *VirtualMachine) MigrationModelId() string {
	if vm.cache.MigrationModelId != "" {
		return vm.cache.MigrationModelId
	}
	s := string(vm.RootBucket.Get(definitions.MigrationModelBytes))
	vm.cache.MigrationModelId = s
	return s
}

func (vm VirtualMachine) ParseCompileAndExecute(v val.Value, scope *ModelScope, parameters []mdl.Model, expect mdl.Model, arguments ...val.Value) (val.Value, mdl.Model, err.Error) {

	instructions, model, e := vm.ParseAndCompile(v, scope, parameters, expect)
	if e != nil {
		return nil, nil, e
	}

	v, e = vm.Execute(instructions, nil, arguments...)
	if e != nil {
		return nil, nil, e
	}

	v, e = slurpIterators(v)
	return v, model, e
}

func (vm VirtualMachine) ParseAndCompile(v val.Value, scope *ModelScope, parameters []mdl.Model, expect mdl.Model) (inst.Sequence, mdl.Model, err.Error) {

	cacheKey := vm.MetaModelId() + string(val.Hash(v, nil).Sum(nil))

	if item, ok := compilerCache.Get(cacheKey); ok {
		entry := item.(compilerCacheEntry)
		return entry.i, entry.m, entry.e
	}

	typed, e := vm.Parse(v, scope, parameters, expect)
	if e != nil {
		compilerCache.Set(cacheKey, compilerCacheEntry{nil, nil, e})
		return nil, nil, e
	}

	instructions, model := vm.CompileFunction(typed), typed.Actual
	compilerCache.Set(cacheKey, compilerCacheEntry{instructions, model, nil})
	return instructions, model, nil

}

// NOTE: parameters == nil means don't check argument types. parameters == []mdl.Mode{} means check for niladic function.
func (vm VirtualMachine) Parse(v val.Value, scope *ModelScope, parameters []mdl.Model, expect mdl.Model) (xpr.TypedFunction, err.Error) {
	ast := xpr.FunctionFromValue(v)
	typed, e := vm.TypeFunction(ast, scope, expect)
	if e != nil {
		return xpr.TypedFunction{}, e
	}
	if parameters != nil {
		if e := checkArgumentTypes(typed, parameters...); e != nil {
			return typed, e
		}
	}
	return typed, nil
}

type compilerCacheEntry struct {
	i inst.Sequence
	m mdl.Model
	e err.Error
}

var compilerCache = cc.NewLru(1024)

// clears compiler cache (for all databases)
func ClearCompilerCache() {
	compilerCache.Clear()
}

func convertNumericType(v val.Value, m mdl.Model) val.Value {
	switch m.Concrete().(type) {

	case mdl.Int8:
		switch v := v.(type) {
		case val.Int8:
			return val.Int8(v)
		case val.Int16:
			return val.Int8(v)
		case val.Int32:
			return val.Int8(v)
		case val.Int64:
			return val.Int8(v)
		case val.Uint8:
			return val.Int8(v)
		case val.Uint16:
			return val.Int8(v)
		case val.Uint32:
			return val.Int8(v)
		case val.Uint64:
			return val.Int8(v)
		case val.Float:
			return val.Int8(v)
		}
	case mdl.Int16:
		switch v := v.(type) {
		case val.Int8:
			return val.Int16(v)
		case val.Int16:
			return val.Int16(v)
		case val.Int32:
			return val.Int16(v)
		case val.Int64:
			return val.Int16(v)
		case val.Uint8:
			return val.Int16(v)
		case val.Uint16:
			return val.Int16(v)
		case val.Uint32:
			return val.Int16(v)
		case val.Uint64:
			return val.Int16(v)
		case val.Float:
			return val.Int16(v)
		}
	case mdl.Int32:
		switch v := v.(type) {
		case val.Int8:
			return val.Int32(v)
		case val.Int16:
			return val.Int32(v)
		case val.Int32:
			return val.Int32(v)
		case val.Int64:
			return val.Int32(v)
		case val.Uint8:
			return val.Int32(v)
		case val.Uint16:
			return val.Int32(v)
		case val.Uint32:
			return val.Int32(v)
		case val.Uint64:
			return val.Int32(v)
		case val.Float:
			return val.Int32(v)
		}
	case mdl.Int64:
		switch v := v.(type) {
		case val.Int8:
			return val.Int64(v)
		case val.Int16:
			return val.Int64(v)
		case val.Int32:
			return val.Int64(v)
		case val.Int64:
			return val.Int64(v)
		case val.Uint8:
			return val.Int64(v)
		case val.Uint16:
			return val.Int64(v)
		case val.Uint32:
			return val.Int64(v)
		case val.Uint64:
			return val.Int64(v)
		case val.Float:
			return val.Int64(v)
		}
	case mdl.Uint8:
		switch v := v.(type) {
		case val.Int8:
			return val.Uint8(v)
		case val.Int16:
			return val.Uint8(v)
		case val.Int32:
			return val.Uint8(v)
		case val.Int64:
			return val.Uint8(v)
		case val.Uint8:
			return val.Uint8(v)
		case val.Uint16:
			return val.Uint8(v)
		case val.Uint32:
			return val.Uint8(v)
		case val.Uint64:
			return val.Uint8(v)
		case val.Float:
			return val.Uint8(v)
		}
	case mdl.Uint16:
		switch v := v.(type) {
		case val.Int8:
			return val.Uint16(v)
		case val.Int16:
			return val.Uint16(v)
		case val.Int32:
			return val.Uint16(v)
		case val.Int64:
			return val.Uint16(v)
		case val.Uint8:
			return val.Uint16(v)
		case val.Uint16:
			return val.Uint16(v)
		case val.Uint32:
			return val.Uint16(v)
		case val.Uint64:
			return val.Uint16(v)
		case val.Float:
			return val.Uint16(v)
		}
	case mdl.Uint32:
		switch v := v.(type) {
		case val.Int8:
			return val.Uint32(v)
		case val.Int16:
			return val.Uint32(v)
		case val.Int32:
			return val.Uint32(v)
		case val.Int64:
			return val.Uint32(v)
		case val.Uint8:
			return val.Uint32(v)
		case val.Uint16:
			return val.Uint32(v)
		case val.Uint32:
			return val.Uint32(v)
		case val.Uint64:
			return val.Uint32(v)
		case val.Float:
			return val.Uint32(v)
		}
	case mdl.Uint64:
		switch v := v.(type) {
		case val.Int8:
			return val.Uint64(v)
		case val.Int16:
			return val.Uint64(v)
		case val.Int32:
			return val.Uint64(v)
		case val.Int64:
			return val.Uint64(v)
		case val.Uint8:
			return val.Uint64(v)
		case val.Uint16:
			return val.Uint64(v)
		case val.Uint32:
			return val.Uint64(v)
		case val.Uint64:
			return val.Uint64(v)
		case val.Float:
			return val.Uint64(v)
		}
	case mdl.Float:
		switch v := v.(type) {
		case val.Int8:
			return val.Float(v)
		case val.Int16:
			return val.Float(v)
		case val.Int32:
			return val.Float(v)
		case val.Int64:
			return val.Float(v)
		case val.Uint8:
			return val.Float(v)
		case val.Uint16:
			return val.Float(v)
		case val.Uint32:
			return val.Float(v)
		case val.Uint64:
			return val.Float(v)
		case val.Float:
			return val.Float(v)
		}
	}
	panic(fmt.Sprintf("%T :: %T", v, m))
}

func stringSliceContains(ss []string, s string) bool {
	for _, t := range ss {
		if t == s {
			return true
		}
	}
	return false
}

func stringSliceToList(ss []string) val.List {
	ls := make(val.List, len(ss), len(ss))
	for i, s := range ss {
		ls[i] = val.String(s)
	}
	return ls
}

// convenience adapter for models associated with a bucket
// (Bucket = model's value collection)
type BucketModel struct {
	mdl.Model
	Bucket string
}

// convenience for dynamic compilation of constant values
type ConstantModel struct {
	mdl.Model
	Value val.Value
}

func (m ConstantModel) Zero() val.Value {
	return m.Value
}

func UnwrapBucket(model mdl.Model) mdl.Model {
	if bm, ok := model.(BucketModel); ok {
		return bm.Model
	}
	return model
}

func UnwrapConstant(model mdl.Model) mdl.Model {
	if cm, ok := model.(ConstantModel); ok {
		return cm.Model
	}
	return model
}

type Permission uint8

const (
	InvalidPermission Permission = iota // zero-guard
	CreatePermission
	ReadPermission
	UpdatePermission
	DeletePermission
)

// CheckPermissions checks permissions, recursively. The base case is nil / permission granted.
// This enables the definition of impure permissions, i.e. permissions that depend on data reads.
func (vm *VirtualMachine) CheckPermission(p Permission, v val.Meta) err.Error {

	is, recKey := (inst.Sequence)(nil), v.Id[0]+v.Id[1]

	switch p {
	case CreatePermission:
		recKey = "create" + recKey
		is = vm.permissions.create

	case ReadPermission:
		recKey = "read" + recKey
		is = vm.permissions.read

	case UpdatePermission:
		recKey = "update" + recKey
		is = vm.permissions.update

	case DeletePermission:
		recKey = "delete" + recKey
		is = vm.permissions.delete

	default:
		panic(fmt.Sprintln("invalid permission value", p))
	}

	if is == nil {
		return nil
	}

	if vm.permRecursions == nil {
		vm.permRecursions = make(map[string]struct{}, 128)
	}

	if _, ok := vm.permRecursions[recKey]; ok {
		return nil
	}

	vm.permRecursions[recKey] = struct{}{}
	defer delete(vm.permRecursions, recKey)

	bv, e := vm.Execute(is, nil, v)
	if e != nil {
		return e
	}

	if !unMeta(bv).(val.Bool) {
		return err.PermissionDeniedError{}
	}

	return nil

}

func (vm VirtualMachine) WrapModelInMeta(mid string, model mdl.Model) mdl.Model {
	m := mdl.NewStruct(5)
	m.Set("id", mdl.Ref{mid})
	m.Set("created", mdl.DateTime{})
	m.Set("updated", mdl.DateTime{})
	m.Set("model", mdl.Ref{vm.MetaModelId()})
	m.Set("value", model)
	return m
}

func (vm VirtualMachine) WrapValueInMeta(value val.Value, id, mid string) val.Meta {
	now := time.Now()
	return val.Meta{
		Id:      val.Ref{mid, id},
		Created: val.DateTime{now},
		Updated: val.DateTime{now},
		Model:   val.Ref{vm.MetaModelId(), mid},
		Value:   value,
	}
}

// turns structs (from persistence) into meta values for vm
func DematerializeMeta(s val.Struct) val.Meta {
	return val.Meta{
		Id:      s.Field("id").(val.Ref),
		Model:   s.Field("model").(val.Ref),
		Created: s.Field("created").(val.DateTime),
		Updated: s.Field("updated").(val.DateTime),
		Value:   s.Field("value"),
	}
}

// turns meta values into structs for persistence
func MaterializeMeta(m val.Meta) val.Struct {
	return val.StructFromMap(map[string]val.Value{
		"id":      m.Id,
		"model":   m.Model,
		"created": m.Created,
		"updated": m.Updated,
		"value":   m.Value,
	})
}

func slurpIterators(v val.Value) (val.Value, err.Error) {
	e := (err.Error)(nil)
	v = v.Transform(func(v val.Value) val.Value {
		if e != nil {
			return v
		}
		if iv, ok := v.(iteratorValue); ok {
			l, e_ := iteratorToList(iv.iterator)
			if e_ != nil {
				e = e_
			}
			return l
		}
		return v
	})
	if e != nil {
		return nil, e
	}
	return v, nil
}

func iteratorToList(ir iterator) (val.List, err.Error) {
	ls := make(val.List, 0, 1024)
	e := ir.forEach(func(v val.Value) err.Error {
		ls = append(ls, v)
		return nil
	})
	return ls, e
}

func (vm VirtualMachine) newReadPermissionFilterIterator(sub iterator) iterator {
	return newFilterIterator(sub, func(v val.Value) (bool, err.Error) {
		if vm.permissions != nil && vm.permissions.read != nil {
			if e := vm.CheckPermission(ReadPermission, v.(val.Meta)); e != nil {
				if _, ok := e.(err.PermissionDeniedError); ok {
					return false, nil
				}
				return false, e
			}
		}
		return true, nil
	})
}

func (vm VirtualMachine) UpdateModels() error {

	meta := vm.MetaModelId()

	v := vm.WrapValueInMeta(definitions.NewMetaModelValue(meta), meta, meta)
	if e := vm.RootBucket.Bucket([]byte(meta)).Put([]byte(meta), karma.Encode(MaterializeMeta(v), vm.WrapModelInMeta(meta, vm.MetaModel()))); e != nil {
		return e
	}

	expr := vm.ExpressionModelId()

	v = vm.WrapValueInMeta(mdl.ValueFromModel(meta, xpr.LanguageModel, nil), expr, meta)
	if e := vm.RootBucket.Bucket([]byte(meta)).Put([]byte(expr), karma.Encode(MaterializeMeta(v), vm.WrapModelInMeta(meta, vm.MetaModel()))); e != nil {
		return e
	}

	return nil
}

func (vm VirtualMachine) InitDB() error {

	db := vm.RootBucket

	meta := common.RandomId()

	if e := db.Put(definitions.MetaModelBytes, []byte(meta)); e != nil {
		return e
	}

	for _, bucket := range [][]byte{
		[]byte(meta),
		definitions.MigrationBucketBytes,
		definitions.NoitargimBucketBytes,
		definitions.TagBucketBytes,
		definitions.UniqueBucketBytes,
		definitions.GraphBucketBytes,
		definitions.PhargBucketBytes,
	} {
		if _, e := db.CreateBucket(bucket); e != nil {
			return e
		}
	}

	{ // create meta model
		v := vm.WrapValueInMeta(definitions.NewMetaModelValue(meta), meta, meta)
		if e := db.Bucket([]byte(meta)).Put([]byte(meta), karma.Encode(MaterializeMeta(v), vm.WrapModelInMeta(meta, vm.MetaModel()))); e != nil {
			return e
		}
	}

	{ // create default models

		e := (err.Error)(nil)

		createBuckets := func(k string, id val.Value) bool {
			if e_ := db.Put([]byte(k), []byte(id.(val.Ref)[1])); e_ != nil {
				e = err.InternalError{Problem: e.Error()}
				return false
			}
			return true
		}

		ids, e := vm.Execute(inst.Sequence{
			inst.CreateMultiple{meta, map[string]inst.Sequence{
				definitions.TagModel: {
					inst.Constant{
						definitions.NewTagModelValue(meta),
					},
				},
				definitions.ExpressionModel: {
					inst.Constant{
						mdl.ValueFromModel(meta, xpr.LanguageModel, nil),
					},
				},
			}},
		}, nil)

		if e != nil {
			return e
		}

		if ids.(val.Map).ForEach(createBuckets); e != nil {
			return e
		}

		ids, e = vm.Execute(inst.Sequence{
			inst.CreateMultiple{meta, map[string]inst.Sequence{
				definitions.MigrationModel: {
					inst.Constant{
						definitions.NewMigrationModelValue(meta, ids.(val.Map).Key(definitions.ExpressionModel).(val.Ref)[1]),
					},
				},
				definitions.RoleModel: {
					inst.Constant{
						definitions.NewRoleModelValue(meta, ids.(val.Map).Key(definitions.ExpressionModel).(val.Ref)[1]),
					},
				},
			}},
		}, nil)

		if e != nil {
			return e
		}

		if ids.(val.Map).ForEach(createBuckets); e != nil {
			return e
		}

		ids, e = vm.Execute(inst.Sequence{
			inst.CreateMultiple{meta, map[string]inst.Sequence{
				definitions.UserModel: {
					inst.Constant{
						definitions.NewUserModelValue(meta, ids.(val.Map).Key(definitions.RoleModel).(val.Ref)[1]),
					},
				},
			}},
		}, nil)

		if e != nil {
			return e
		}

		if ids.(val.Map).ForEach(createBuckets); e != nil {
			return e
		}

	}

	{ // create default tags
		_, e := vm.Execute(inst.Sequence{
			inst.CreateMultiple{vm.TagModelId(), map[string]inst.Sequence{
				"_model": inst.Sequence{
					inst.Constant{val.StructFromMap(map[string]val.Value{
						"tag":   val.String("_model"),
						"model": val.Ref{meta, vm.MetaModelId()},
					})},
				},
				"_tag": inst.Sequence{
					inst.Constant{val.StructFromMap(map[string]val.Value{
						"tag":   val.String("_tag"),
						"model": val.Ref{meta, vm.TagModelId()},
					})},
				},
				"_expression": inst.Sequence{
					inst.Constant{val.StructFromMap(map[string]val.Value{
						"tag":   val.String("_expression"),
						"model": val.Ref{meta, vm.ExpressionModelId()},
					})},
				},
				"_migration": inst.Sequence{
					inst.Constant{val.StructFromMap(map[string]val.Value{
						"tag":   val.String("_migration"),
						"model": val.Ref{meta, vm.MigrationModelId()},
					})},
				},
				"_role": inst.Sequence{
					inst.Constant{val.StructFromMap(map[string]val.Value{
						"tag":   val.String("_role"),
						"model": val.Ref{meta, vm.RoleModelId()},
					})},
				},
				"_user": inst.Sequence{
					inst.Constant{val.StructFromMap(map[string]val.Value{
						"tag":   val.String("_user"),
						"model": val.Ref{meta, vm.UserModelId()},
					})},
				},
			}},
		}, nil)

		if e != nil {
			return e
		}
	}

	{ // create root user

		trueExpr, e := vm.Execute(inst.Sequence{
			inst.CreateMultiple{vm.ExpressionModelId(), map[string]inst.Sequence{
				"self": inst.Sequence{
					inst.Constant{
						xpr.ValueFromFunction(xpr.NewFunction([]string{"_"}, xpr.Literal{val.Bool(true)})),
					},
				},
			}},
			inst.Constant{val.String("self")},
			inst.Key{},
		}, nil)
		if e != nil {
			return e
		}

		sysRole, e := vm.Execute(inst.Sequence{
			inst.CreateMultiple{vm.RoleModelId(), map[string]inst.Sequence{
				"self": inst.Sequence{
					inst.Constant{val.StructFromMap(map[string]val.Value{
						"name": val.String("admins"),
						"permissions": val.StructFromMap(map[string]val.Value{
							"create": trueExpr.(val.Ref),
							"read":   trueExpr.(val.Ref),
							"update": trueExpr.(val.Ref),
							"delete": trueExpr.(val.Ref),
						}),
					})},
				},
			}},
			inst.Constant{val.String("self")},
			inst.Key{},
		}, nil)
		if e != nil {
			return e
		}

		sysUser, e := vm.Execute(inst.Sequence{
			inst.CreateMultiple{vm.UserModelId(), map[string]inst.Sequence{
				"self": inst.Sequence{
					inst.Constant{val.StructFromMap(map[string]val.Value{
						"username": val.String("admin"),
						"password": val.String(""),
						"roles":    val.List{sysRole},
					})},
				},
			}},
			inst.Constant{val.String("self")},
			inst.Key{},
		}, nil)
		if e != nil {
			return e
		}
		if e := db.Put(definitions.RootUserBytes, []byte(sysUser.(val.Ref)[1])); e != nil {
			return e
		}

	}

	return nil
}

func (vm VirtualMachine) Get(mid, oid string) (val.Meta, err.Error) {

	mv, e := vm.get(mid, oid)
	if e != nil {
		return mv, e
	}

	if vm.permissions != nil && vm.permissions.read != nil {
		if e := vm.CheckPermission(ReadPermission, mv); e != nil {
			return val.Meta{}, e
		}
	}

	return mv, nil

}

func (vm VirtualMachine) get(mid, oid string) (val.Meta, err.Error) {

	bk := vm.RootBucket.Bucket([]byte(mid))
	if bk == nil {
		return val.Meta{}, err.ModelNotFoundError{
			err.ObjectNotFoundError{
				Ref: val.Ref{vm.MetaModelId(), mid},
			},
		}
	}

	m, e := vm.Model(mid)
	if e != nil {
		return val.Meta{}, e
	}

	dt := bk.Get([]byte(oid))
	if dt == nil {
		return val.Meta{}, err.ObjectNotFoundError{
			Ref: val.Ref{mid, oid},
		}
	}

	vl, _ := karma.Decode(dt, vm.WrapModelInMeta(mid, m.Model))
	mv := DematerializeMeta(vl.(val.Struct))

	return mv, nil

}

const ModelCacheCapacity = 512

var ModelCache = cc.NewLru(ModelCacheCapacity)

func (vm VirtualMachine) Model(mid string) (BucketModel, err.Error) {

	metaId := vm.MetaModelId()
	cacheKey := metaId + "/" + mid // metaId is distinct for every database

	if m, ok := ModelCache.Get(cacheKey); ok {
		return BucketModel{Bucket: mid, Model: m.(mdl.Model)}, nil
	}

	bs := vm.RootBucket.Bucket([]byte(metaId)).Get([]byte(mid))
	if bs == nil {
		return BucketModel{}, err.ModelNotFoundError{
			err.ObjectNotFoundError{
				Ref: val.Ref{vm.MetaModelId(), mid},
			},
		}
	}

	mv, _ := karma.Decode(bs, vm.WrapModelInMeta(metaId, vm.MetaModel()))

	m, e := mdl.ModelFromValue(vm.MetaModelId(), unMeta(DematerializeMeta(mv.(val.Struct))).(val.Union), nil)
	if e != nil {
		log.Panicln("failed decoding persisted model", e.Error(), mid)
	}

	ModelCache.Set(cacheKey, m)

	return BucketModel{Model: m, Bucket: mid}, nil // note: m.Copy _is_ necessary.
}

func (vm VirtualMachine) MetaModel() mdl.Model {
	m, e := mdl.ModelFromValue(vm.MetaModelId(), definitions.NewMetaModelValue(vm.MetaModelId()).(val.Union), nil)
	if e != nil {
		log.Panicln(e)
	}
	return m
}

func (vm VirtualMachine) InRefs(mid, id string) []val.Ref {

	out := ([]val.Ref)(nil)

	if bk := vm.RootBucket.Bucket(definitions.PhargBucketBytes).Bucket(encodeVertex(mid, id)); bk != nil {
		e := bk.ForEach(func(source, _ []byte) error {
			m, i := decodeVertex(source)
			out = append(out, val.Ref{m, i})
			return nil
		})
		if e != nil {
			log.Panicln(e)
		}
	}

	return out
}

func (vm VirtualMachine) Delete(mid string, id string) err.Error {

	db := vm.RootBucket

	v, e := vm.Get(mid, id)
	if e != nil {
		if _, ok := e.(err.ObjectNotFoundError); ok {
			return nil
		}
		return e
	}

	if mid == vm.UserModelId() && id == vm.RootUserId() {
		return err.ExecutionError{
			`cannot delete root user`,
			nil,
		}
	}

	if mid == vm.MetaModelId() {

		if vm.isDefaultModelId(id) {
			return err.ExecutionError{
				`cannot delete default models`,
				nil,
			}
		}

		// check that there are no incoming references to any object in model
		// except from the same model to itself
		pb := db.Bucket(definitions.PhargBucketBytes)
		e := pb.ForEach(func(target, _ []byte) error {
			if bytes.HasPrefix(target, []byte(id)) {
				return pb.Bucket(target).ForEach(func(source, _ []byte) error {
					if !bytes.HasPrefix(source, []byte(id)) {
						return err.ExecutionError{
							fmt.Sprintf(`some objects in model "%s" are still referenced by objects in other models`, id),
							nil,
						} // break
					}
					return nil
				})
			}
			return nil
		})
		if e != nil {
			if e, ok := e.(err.ExecutionError); ok {
				return e
			}
			log.Panicln(e)
		}

	}

	{ // check that there are no references to object (except self-references), this includes migrations
		for _, r := range vm.InRefs(mid, id) {
			if r[0] != mid || r[1] != id {
				return err.ExecutionError{
					`delete: there are graph relations to the object being deleted.`,
					nil,
					// C: val.Map{"source": r, "target": val.Ref{mid, id}},
				}
			}
		}
	}

	{ // delete graph edges

		key := encodeVertex(mid, id)
		gb, pb := db.Bucket(definitions.GraphBucketBytes), db.Bucket(definitions.PhargBucketBytes)

		targets := ([][]byte)(nil)
		if kb := gb.Bucket(key); kb != nil {
			e := kb.ForEach(func(k, v []byte) error {
				targets = append(targets, k)
				return nil
			})
			if e != nil {
				log.Panicln(e)
			}
			if e := gb.DeleteBucket(key); e != nil {
				log.Panicln(e)
			}
		}

		for _, target := range targets {
			if e := pb.Bucket(target).Delete(key); e != nil {
				log.Panicln(e)
			}
		}

	}

	{ // delete unique indices

		m, e := vm.Model(mid)
		if e != nil {
			return e
		}

		if uniqs := uniqueHashes(m, v.Value); uniqs != nil {

			ub := db.Bucket(definitions.UniqueBucketBytes).Bucket([]byte(mid))

			for _, uq := range uniqs {
				key := append(hashStringSlice(uq.Path), uq.Hash...)
				if e := ub.Delete(key); e != nil {
					log.Panicln(e)
				}
			}

		}

	}

	{ // delete object itself
		if e := db.Bucket([]byte(mid)).Delete([]byte(id)); e != nil {
			log.Panicln(e)
		}
	}

	// if deleting a model: remove data, graph and pharg buckets
	if mid == vm.MetaModelId() {

		if e := db.DeleteBucket([]byte(id)); e != nil {
			log.Panicln(e)
		}

		if e := db.Bucket(definitions.GraphBucketBytes).DeleteBucket(encodeVertex(mid, id)); e != nil {
			log.Panicln(e)
		}

		if e := db.Bucket(definitions.PhargBucketBytes).DeleteBucket(encodeVertex(mid, id)); e != nil {
			log.Panicln(e)
		}

	}

	if mid == vm.TagModelId() {

		tag := v.Value.(val.Struct).Field("tag").(val.String)

		if e := db.Bucket(definitions.TagBucketBytes).Delete([]byte(tag)); e != nil {
			log.Panicln(e)
		}

	}

	if mid == vm.MigrationModelId() {

		// FIXME: deleting a migration can break another migration-path involving relocateRef

		migration := v.Value.(val.List)

		migs := db.Bucket(definitions.MigrationBucketBytes)
		sgim := db.Bucket(definitions.NoitargimBucketBytes)

		for _, mig := range migration {

			object := mig.(val.Struct)

			source := object.Field("source").(val.Ref)
			sourceMID := source[1]

			target := object.Field("target").(val.Ref)
			targetMID := target[1]

			if e := migs.Bucket([]byte(sourceMID)).Delete([]byte(targetMID)); e != nil {
				log.Panicln(e)
			}

			if e := sgim.Bucket([]byte(targetMID)).Delete([]byte(sourceMID)); e != nil {
				log.Panicln(e)
			}

		}

	}

	if mid == vm.MetaModelId() {
		ModelCache.Remove(mid + "/" + id)
	}

	return nil

}

// Note: Write is idempontent i.e. also serves as update function
func (vm VirtualMachine) Write(mid string, values map[string]val.Meta) err.Error {

	db := vm.RootBucket

	for id, v := range values {

		md, e := vm.Model(mid)
		if e != nil {
			return e
		}

		if mid == vm.ExpressionModelId() {
			fun, e := vm.Parse(v.Value, nil, nil, nil)
			if e != nil {
				return err.ExecutionError{
					Problem: `there was an error compiling the function to be persisted`,
					Child_:  e,
				}
			}
			v.Value = xpr.ValueFromFunction(fun)
		}

		if uniqs := uniqueHashes(md, v.Value); len(uniqs) > 0 {

			ub, e := db.Bucket(definitions.UniqueBucketBytes).CreateBucketIfNotExists([]byte(mid))
			if e != nil {
				log.Panicln(e) // [1]
			}

			{ // check unique constraint violation
				for _, uq := range uniqs {
					key := append(hashStringSlice(uq.Path), uq.Hash...)
					if bs := ub.Get(key); bs != nil {
						if string(bs) != id { // only error if not updating self
							return err.ExecutionError{
								`unique constraint violation`,
								nil,
								// TODO: use err.ErrorPath to indicate where the violation occurred
								// C: val.Map{"path": stringSliceToList(uq.Path)},
							}
						}
					}
				}
			}

			{ // delete old unique hashes
				del := make([][]byte, 0, 16)
				e = ub.ForEach(func(k, v []byte) error {
					if string(v) == id {
						del = append(del, k)
					}
					return nil
				})
				if e != nil {
					log.Panicln(e) // [1]
				}
				for _, k := range del {
					if e := ub.Delete(k); e != nil {
						log.Panicln(e) // [1]
					}
				}
			}

			{ // write new unique hashes
				for _, uq := range uniqs {
					key := append(hashStringSlice(uq.Path), uq.Hash...)
					if e := ub.Put(key, []byte(id)); e != nil {
						log.Panicln(e)
					}
				}
			}

		}

		{

			edges := extractRefs(v.Value)

			for _, edge := range edges {
				if _, ok := values[edge[1]]; ok {
					continue
				}
				tb := db.Bucket([]byte(edge[0]))
				if tb == nil {
					log.Panicf(`ref to inexistent model %s in model %s`, edge[0], mid)
				}
				if tb.Get([]byte(edge[1])) == nil {
					return err.ExecutionError{
						Problem: `referenced object not found`,
						Child_: err.ObjectNotFoundError{
							Ref: val.Ref{edge[0], edge[1]},
						},
					}
				}
			}

			key := encodeVertex(mid, id)

			gb, e := db.Bucket(definitions.GraphBucketBytes).CreateBucketIfNotExists(key)
			if e != nil {
				log.Panicln(e) // [1]
			}

			{ // remove old graph relations (backward and forward)
				del := make([][]byte, 0, 16)
				e = gb.ForEach(func(k, _ []byte) error {
					del = append(del, k)
					return nil
				})
				if e != nil {
					log.Panicln(e)
				}
				for _, d := range del {
					if e := gb.Delete(d); e != nil {
						log.Panicln(e)
					}
					if e := db.Bucket(definitions.PhargBucketBytes).Bucket(d).Delete(key); e != nil {
						log.Panicln(e)
					}
				}
			}

			{ // create new graph reations (backward and forward)

				for _, edge := range edges {
					if e := gb.Put(encodeVertex(edge[0], edge[1]), []byte{}); e != nil {
						log.Panicln(e)
					}
					pb, e := db.Bucket(definitions.PhargBucketBytes).CreateBucketIfNotExists(encodeVertex(edge[0], edge[1]))
					if e != nil {
						log.Panicln(e)
					}
					if e := pb.Put(key, []byte{}); e != nil {
						log.Panicln(e)
					}
				}
			}

		}

		if mid == vm.RoleModelId() {
			// TODO: validate that permission expressions are boolean
			// and do not contain: CRUD
			// ^ analogously for migrations
		}

		if mid == vm.MetaModelId() {
			if _, e := db.CreateBucketIfNotExists([]byte(id)); e != nil {
				log.Panicln(e)
			}
			ModelCache.Remove(mid + "/" + id)
		}

		if mid == vm.TagModelId() {

			o := v.Value.(val.Struct)
			tag, model := o.Field("tag").(val.String), o.Field("model").(val.Ref)

			if e := db.Bucket(definitions.TagBucketBytes).Put([]byte(tag), []byte(model[1])); e != nil {
				log.Panicln(e)
			}

		}

		if mid == vm.MigrationModelId() {

			migration := v.Value.(val.List)

			// set of source model IDs
			sources := make(map[string]struct{}, len(migration))

			// dependentMID -> dependencyMID (single dependency is enough)
			dependencies := make(map[string]string, len(migration))

			pharg := db.Bucket(definitions.PhargBucketBytes)
			if pharg == nil {
				log.Panicln("pharg bucket missing")
			}

			migs := db.Bucket(definitions.MigrationBucketBytes)
			sgim := db.Bucket(definitions.NoitargimBucketBytes)

			// TODO: deny creating migrations from/to _model, _migration, _expression ?

			for _, mig := range migration {

				object := mig.(val.Struct)

				source := object.Field("source").(val.Ref)
				sourceMID := source[1]

				target := object.Field("target").(val.Ref)
				targetMID := target[1]

				if sb := migs.Bucket([]byte(sourceMID)); sb != nil {
					if bs := sb.Get([]byte(targetMID)); bs != nil {
						return err.ExecutionError{
							fmt.Sprintf(`there is already a migration from source model "%s" to target model "%s"`, sourceMID, targetMID),
							nil,
							// C: val.Map{
							// 	"source": source,
							// 	"target": target,
							// },
						}
					}
				}

				// resolve graph dependants
				vertex := encodeVertex(source[0], source[1])
				if db := pharg.Bucket(vertex); db != nil {
					e := db.ForEach(func(k, _ []byte) error {
						m, i := decodeVertex(k)
						if m != vm.MetaModelId() {
							return nil
						}
						dependencies[i] = sourceMID
						return nil
					})
					if e != nil {
						log.Panicln(e)
					}
				}

				sources[sourceMID] = struct{}{}

			}

			for dependant, dependency := range dependencies {
				if _, ok := sources[dependant]; !ok {
					return err.ExecutionError{
						fmt.Sprintf(`model "%s" has graph relations to model "%s" but is not included as source in migration`, dependant, dependency),
						nil,
						// C: val.Map{
						// 	"dependant":  val.Ref{vm.MetaModelId(), dependant},
						// 	"dependency": val.Ref{vm.MetaModelId(), dependency},
						// },
					}
				}
			}

			for _, mig := range migration {

				object := mig.(val.Struct)

				source := object.Field("source").(val.Ref)
				sourceMID := source[1]

				target := object.Field("target").(val.Ref)
				targetMID := target[1]

				sourceModel, e := vm.Model(sourceMID)
				if e != nil {
					return e
				}

				targetModel, e := vm.Model(targetMID)
				if e != nil {
					return e
				}

				exprRef := val.Ref{}
				expr := (val.Value)(nil)

				if expression := object.Field("expression").(val.Union); expression.Case == "auto" {
				}

				panic("todo")

				_ = sgim
				_ = sourceModel
				_ = targetModel
				_ = exprRef
				_ = expr

			}

		}

		// actual persistence of the value
		if e := db.Bucket([]byte(mid)).Put([]byte(id), karma.Encode(MaterializeMeta(v), vm.WrapModelInMeta(mid, md.Model))); e != nil {
			log.Panicln(e)
		}

	}

	return nil

}

func allConstants(args ...inst.Instruction) bool {
	for _, a := range args {
		if _, ok := a.(inst.Constant); !ok {
			return false
		}
	}
	return true
}

func extractRefs(v val.Value) []val.Ref {

	edges := ([]val.Ref)(nil)

	unMeta(v).Transform(func(v val.Value) val.Value {
		if r, ok := v.(val.Ref); ok {
			unique := true
			for _, d := range edges {
				if r == d {
					unique = false
					break
				}
			}
			if unique {
				edges = append(edges, r)
			}
		}
		return v
	})

	return edges
}

func uniqueHashes(m mdl.Model, v val.Value) []uniquePath {

	uniquePaths := uniquePaths(m)

	if uniquePaths == nil {
		return nil
	}

	hashes := make([]hash.Hash64, len(uniquePaths), len(uniquePaths))

	UnwrapBucket(m).TraverseValue(unMeta(v), func(v val.Value, m mdl.Model) {

		if u, ok := m.(mdl.Unique); ok {

			i := -1

			for j, p := range uniquePaths {
				if p.Unique == u {
					i = j
					break
				}
			}

			// panic if i == -1 -> programming error

			if hashes[i] == nil {
				hashes[i] = fnv.New64()
			}

			hashes[i].Write([]byte{SeparatorByte})     // separator
			hashes[i].Write(val.Hash(v, nil).Sum(nil)) // sorts keys in structs and maps

		}

	})

	for i, h := range hashes {
		uniquePaths[i].Hash = h.Sum(nil)
	}

	return uniquePaths
}

type uniquePath struct {
	Unique mdl.Unique
	Path   []string
	Hash   []byte
}

func uniquePaths(m mdl.Model) []uniquePath {
	paths := ([]uniquePath)(nil)
	m.Traverse(nil, func(p []string, m mdl.Model) {
		if u, ok := m.(mdl.Unique); ok {
			paths = append(paths, uniquePath{
				Unique: u, Path: p, Hash: nil,
			})
		}
	})
	return paths
}

func hashStringSlice(ss []string) []byte {
	hash := fnv.New64()
	for _, s := range ss {
		hash.Write([]byte{SeparatorByte}) // separator
		hash.Write([]byte(s))
	}
	return hash.Sum(nil)
}

func encodeVertex(model, id string) []byte {
	vert := make([]byte, 0, len(model)+len(id)+1)
	vert = append(vert, model...)
	vert = append(vert, SeparatorByte)
	vert = append(vert, id...)
	return vert
}

func decodeVertex(bs []byte) (string, string) {
	for i, b := range bs {
		if b == SeparatorByte {
			return string(bs[:i]), string(bs[i+1:])
		}
	}
	panic("never reached")
}

type migrationNode struct {
	InModel   mdl.Model
	InValue   val.Value
	Migration val.Value // a function
	Children  map[string]*migrationNode
}

func (vm VirtualMachine) applyMigrationTree(id string, tree map[string]*migrationNode, cache map[string]val.Meta) map[string]val.Meta {

	if cache == nil {
		cache = make(map[string]val.Meta)
	}

	if len(tree) == 0 {
		return cache
	}

	for mid, node := range tree {

		it, mm, e := vm.ParseAndCompile(node.Migration, nil, []mdl.Model{node.InModel}, nil)
		if e != nil {
			log.Panicln(pretty.Sprint(e)) // TODO: should theoretically never happen, but... add lots of contextual info in case it does
		}

		vl, e := vm.Execute(it, nil, node.InValue)
		if e != nil {
			log.Panicln(pretty.Sprint(e)) // TODO: should theoretically never happen, but... add lots of contextual info in case it does
		}

		mv := vm.WrapValueInMeta(vl, id, mid)

		cache[mid] = mv

		for _, child := range node.Children {
			child.InValue, child.InModel = mv, mm
		}

		vm.applyMigrationTree(id, node.Children, cache)
	}

	return cache

}

func modelsInMigrationTree(tree map[string]*migrationNode, cache []string) []string {

	if cache == nil {
		cache = make([]string, 0, 16)
	}

	for k, v := range tree {
		cache = append(cache, k)
		modelsInMigrationTree(v.Children, cache)
	}

	return cache
}

func (vm *VirtualMachine) migrationTree(from string, seen map[string]struct{}) map[string]*migrationNode {

	db := vm.RootBucket

	if seen == nil {
		seen = make(map[string]struct{})
	}

	seen[from] = struct{}{}

	mb := db.Bucket(definitions.MigrationBucketBytes)
	if mb == nil {
		return nil // initializing database, no migrations to apply
	}

	fb := mb.Bucket([]byte(from))
	if fb == nil {
		return nil
	}

	migs := make(map[string]*migrationNode)

	e := fb.ForEach(func(k, v []byte) error {

		x := string(k)

		if _, ok := seen[x]; ok {
			return nil // continue
		}

		exprID := string(v)
		exprMeta, e := vm.Get(vm.ExpressionModelId(), exprID)
		if e != nil {
			return e
		}

		migs[x] = &migrationNode{
			InModel:   nil,
			InValue:   nil,
			Migration: exprMeta.Value,
			Children:  vm.migrationTree(x, seen),
		}

		return nil // continue
	})

	if e != nil {
		log.Panicln(e)
	}

	return migs

}

func (vm *VirtualMachine) reverseMigrationTree(to string, seen map[string]struct{}) map[string]*migrationNode {

	db := vm.RootBucket

	if seen == nil {
		seen = make(map[string]struct{})
	}

	seen[to] = struct{}{}

	mb := db.Bucket(definitions.NoitargimBucketBytes)
	if mb == nil {
		return nil // initializing database, no migrations to apply
	}

	tb := mb.Bucket([]byte(to))
	if tb == nil {
		return nil
	}

	migs := make(map[string]*migrationNode)

	e := tb.ForEach(func(k, v []byte) error {

		x := string(k)

		if _, ok := seen[x]; ok {
			return nil // continue
		}

		exprID := string(v)
		exprMeta, e := vm.Get(vm.ExpressionModelId(), exprID)
		if e != nil {
			return e
		}

		migs[x] = &migrationNode{
			InModel:   nil,
			InValue:   nil,
			Migration: exprMeta.Value,
			Children:  vm.migrationTree(x, seen),
		}

		return nil // continue
	})

	if e != nil {
		log.Panicln(e)
	}

	return migs

}

type permissions struct {
	create inst.Sequence
	read   inst.Sequence
	update inst.Sequence
	delete inst.Sequence
}

func (vm *VirtualMachine) lazyLoadPermissions() err.Error {
	if vm.permissions != nil || vm.UserID == "" {
		return nil
	}
	vm.permissions = &permissions{
		create: inst.Sequence{inst.Constant{val.Bool(true)}},
		read:   inst.Sequence{inst.Constant{val.Bool(true)}},
		update: inst.Sequence{inst.Constant{val.Bool(true)}},
		delete: inst.Sequence{inst.Constant{val.Bool(true)}},
	}
	ps, e := vm.permissionsForUserId(vm.UserID)
	vm.permissions = nil
	if e != nil {
		return e
	}
	vm.permissions = ps
	return nil
}

// user -> role -> permission
func (vm *VirtualMachine) permissionsForUserId(uid string) (*permissions, err.Error) {
	v, e := vm.Execute(inst.Sequence{
		inst.Constant{val.Ref{vm.UserModelId(), uid}},
		inst.Deref{},
		inst.Field{Key: "roles"},
		inst.MapList{inst.Sequence{
			inst.Define("i"),
			inst.Pop{},
			inst.Define("role"),
			inst.Pop{},
			inst.Scope("role"),
			inst.Deref{},
			inst.Field{Key: "permissions"},
			inst.MapStruct{inst.Sequence{
				inst.Define("k"),
				inst.Pop{},
				inst.Define("permission"),
				inst.Pop{},
				inst.Scope("permission"),
				inst.Deref{},
			}},
		}},
	}, nil)
	if e != nil {
		if _, ok := e.(err.ObjectNotFoundError); ok {
			// TODO: better message for "user not found"
		}
		return nil, e
	}
	l := v.(val.List)
	c, r, u, d := make(val.List, 0, len(l)), make(val.List, 0, len(l)), make(val.List, 0, len(l)), make(val.List, 0, len(l))
	for _, m := range l {
		s := m.(val.Struct)
		c = append(c, unMeta(s.Field("create")))
		r = append(r, unMeta(s.Field("read")))
		u = append(u, unMeta(s.Field("update")))
		d = append(d, unMeta(s.Field("delete")))
	}

	ci, cm, ce := vm.ParseAndCompile(orExpressions(c), nil, []mdl.Model{AnyModel}, mdl.Bool{})
	if ce != nil {
		return nil, err.ExecutionError{
			Problem: `failed compiling create permissions`,
			Child_:  ce,
		}
	}
	if _, ok := cm.Concrete().(mdl.Bool); !ok {
		return nil, err.ExecutionError{
			fmt.Sprintf(`non-boolean create permission expression in user %s`, uid),
			nil,
		}
	}
	ri, rm, re := vm.ParseAndCompile(orExpressions(r), nil, []mdl.Model{AnyModel}, mdl.Bool{})
	if re != nil {
		return nil, err.ExecutionError{
			Problem: `failed compiling read permissions`,
			Child_:  re,
		}
	}
	if _, ok := rm.Concrete().(mdl.Bool); !ok {
		return nil, err.ExecutionError{
			fmt.Sprintf(`non-boolean read permission expression in user %s`, uid),
			nil,
		}
	}
	ui, um, ue := vm.ParseAndCompile(orExpressions(u), nil, []mdl.Model{AnyModel}, mdl.Bool{})
	if re != nil {
		return nil, err.ExecutionError{
			Problem: `failed compiling update permissions`,
			Child_:  ue,
		}
	}
	if _, ok := um.Concrete().(mdl.Bool); !ok {
		return nil, err.ExecutionError{
			fmt.Sprintf(`non-boolean update permission expression in user %s`, uid),
			nil,
		}
	}
	di, dm, de := vm.ParseAndCompile(orExpressions(d), nil, []mdl.Model{AnyModel}, mdl.Bool{})
	if de != nil {
		return nil, err.ExecutionError{
			Problem: `failed compiling delete permissions`,
			Child_:  de,
		}
	}
	if _, ok := dm.Concrete().(mdl.Bool); !ok {
		return nil, err.ExecutionError{
			fmt.Sprintf(`non-boolean delete permission expression in user %s`, uid),
			nil,
		}
	}
	return &permissions{
		create: ci,
		read:   ri,
		update: ui,
		delete: di,
	}, nil
}

func orExpressions(xs val.List /* of val.Unions */) val.Value {
	if len(xs) == 0 {
		return val.Union{"bool", val.Bool(false)}
	}
	if len(xs) == 1 {
		return xs[0]
	}
	return val.Union{"or", xs}
}

func findCycle(reachable map[string]map[string]struct{}, vertex string, path []string) []string {
	if inStringSlice(vertex, path) {
		return append(path, vertex)
	}
	for target, _ := range reachable[vertex] {
		if cycle := findCycle(reachable, target, append(path, vertex)); cycle != nil {
			return cycle
		}
	}
	return nil
}

func inStringSlice(s string, ss []string) bool {
	for _, z := range ss {
		if s == z {
			return true
		}
	}
	return false
}

func deoptionalize(m mdl.Model) mdl.Model {
	if m, ok := m.(mdl.Optional); ok {
		return m.Model
	}
	return m
}

func reverse(v val.List) {
	l := len(v)
	for i := 0; i < l/2; i++ {
		j := l - i - 1
		v[i], v[j] = v[j], v[i]
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
