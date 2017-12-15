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
	"karma.run/codec"
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
	Codec      codec.Interface

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

func (vm VirtualMachine) ParseCompileAndExecute(v val.Value, argument, expected mdl.Model) (val.Value, mdl.Model, err.Error) {

	instructions, model, e := vm.ParseAndCompile(v, argument, expected)
	if e != nil {
		return nil, nil, e
	}

	v, e = vm.Execute(flattenSequences(instructions, nil), nil)
	if e != nil {
		return nil, nil, e
	}

	v, e = slurpIterators(v)
	return v, model, e
}

func (vm VirtualMachine) CompileAndExecute(node xpr.Expression, argument, expected mdl.Model) (val.Value, mdl.Model, err.Error) {
	typed, e := vm.TypeExpression(node, argument, expected)
	if e != nil {
		return nil, nil, e
	}
	instructions := flattenSequences(vm.Compile(typed), nil)
	v, e := vm.Execute(instructions, nil)
	if e != nil {
		return nil, nil, e
	}
	v, e = slurpIterators(v)
	return v, typed.Actual, e
}

func (vm VirtualMachine) ParseAndCompile(v val.Value, argument, expected mdl.Model) (inst.Sequence, mdl.Model, err.Error) {

	cacheKey := vm.MetaModelId() + string(val.Hash(v, nil).Sum(nil))

	if item, ok := compilerCache.Get(cacheKey); ok {
		entry := item.(compilerCacheEntry)
		return entry.i, entry.m, entry.e
	}

	typed, e := vm.Parse(v, argument, expected)
	if e != nil {
		compilerCache.Set(cacheKey, compilerCacheEntry{nil, nil, e})
		return nil, nil, e
	}

	instruction, model := vm.Compile(typed), typed.Actual
	instructions := flattenSequences(instruction, nil)
	compilerCache.Set(cacheKey, compilerCacheEntry{instructions, model, nil})
	return instructions, model, nil

}

func (vm VirtualMachine) Parse(v val.Value, argument, expected mdl.Model) (xpr.TypedExpression, err.Error) {
	ast := xpr.ExpressionFromValue(v)
	untyped, e := xpr.EliminateDoNotation(ast)
	if e != nil {
		return xpr.TypedExpression{}, e
	}
	typed, e := vm.TypeExpression(untyped, argument, expected)
	if e != nil {
		return xpr.TypedExpression{}, e
	}
	return FoldConstants(typed), nil
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

func FoldConstants(n xpr.TypedExpression) xpr.TypedExpression {
	return n.Transform(func(n xpr.Expression) xpr.Expression {
		if tx, ok := n.(xpr.TypedExpression); ok {
			if cm, ok := tx.Actual.(ConstantModel); ok {
				return xpr.TypedExpression{xpr.Literal{cm.Value}, tx.Expected, tx.Actual}
			}
		}
		return n
	}).(xpr.TypedExpression)
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

	bv, e := vm.Execute(is, v)
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

type Stack []val.Value

func (s *Stack) Push(v val.Value) {
	(*s) = append((*s), v)
}

func (s *Stack) Pop() val.Value {
	l := s.Len()
	v := (*s)[l-1]
	(*s) = (*s)[:l-1]
	return v
}

func (s Stack) Len() int {
	return len(s)
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
		ids, e := vm.Execute(inst.Sequence{
			inst.Constant{val.String(definitions.TagModel)},
			inst.Constant{definitions.NewTagModelValue(meta)},

			inst.Constant{val.String(definitions.ExpressionModel)},
			inst.Constant{mdl.ValueFromModel(meta, xpr.LanguageModel, nil)},

			inst.Constant{val.String(definitions.MigrationModel)},
			inst.Constant{definitions.NewMigrationModelValue(meta, definitions.ExpressionModel)},

			inst.Constant{val.String(definitions.RoleModel)},
			inst.Constant{definitions.NewRoleModelValue(meta, definitions.ExpressionModel)},

			inst.Constant{val.String(definitions.UserModel)},
			inst.Constant{definitions.NewUserModelValue(meta, definitions.RoleModel)},

			inst.BuildMap{5},
			inst.CreateMultiple{meta},
		}, nil)

		if e != nil {
			return e
		}

		ids.(val.Map).ForEach(func(k string, id val.Value) bool {
			if e_ := db.Put([]byte(k), []byte(id.(val.Ref)[1])); e_ != nil {
				e = err.InternalError{Problem: e.Error()}
				return false
			}
			return true
		})
		if e != nil {
			return e
		}
	}

	{ // create default tags
		_, e := vm.Execute(inst.Sequence{
			inst.Constant{val.String("_model")},
			inst.Constant{val.StructFromMap(map[string]val.Value{
				"tag":   val.String("_model"),
				"model": val.Ref{meta, vm.MetaModelId()},
			})},

			inst.Constant{val.String("_tag")},
			inst.Constant{val.StructFromMap(map[string]val.Value{
				"tag":   val.String("_tag"),
				"model": val.Ref{meta, vm.TagModelId()},
			})},

			inst.Constant{val.String("_expression")},
			inst.Constant{val.StructFromMap(map[string]val.Value{
				"tag":   val.String("_expression"),
				"model": val.Ref{meta, vm.ExpressionModelId()},
			})},

			inst.Constant{val.String("_migration")},
			inst.Constant{val.StructFromMap(map[string]val.Value{
				"tag":   val.String("_migration"),
				"model": val.Ref{meta, vm.MigrationModelId()},
			})},

			inst.Constant{val.String("_role")},
			inst.Constant{val.StructFromMap(map[string]val.Value{
				"tag":   val.String("_role"),
				"model": val.Ref{meta, vm.RoleModelId()},
			})},

			inst.Constant{val.String("_user")},
			inst.Constant{val.StructFromMap(map[string]val.Value{
				"tag":   val.String("_user"),
				"model": val.Ref{meta, vm.UserModelId()},
			})},

			inst.BuildMap{6},
			inst.CreateMultiple{vm.TagModelId()},
		}, nil)

		if e != nil {
			return e
		}
	}

	{ // create root user

		trueExpr, e := vm.Execute(inst.Sequence{
			inst.Constant{val.String("self")},
			inst.Constant{val.Bool(true)},
			inst.BuildMap{1},
			inst.CreateMultiple{vm.ExpressionModelId()},
			inst.Constant{val.String("self")},
			inst.Key{},
		}, nil)
		if e != nil {
			return e
		}

		sysRole, e := vm.Execute(inst.Sequence{
			inst.Constant{val.String("self")},
			inst.Constant{val.StructFromMap(map[string]val.Value{
				"name": val.String("admins"),
				"permissions": val.StructFromMap(map[string]val.Value{
					"create": trueExpr.(val.Ref),
					"read":   trueExpr.(val.Ref),
					"update": trueExpr.(val.Ref),
					"delete": trueExpr.(val.Ref),
				}),
			})},
			inst.BuildMap{1},
			inst.CreateMultiple{vm.RoleModelId()},
			inst.Constant{val.String("self")},
			inst.Key{},
		}, nil)
		if e != nil {
			return e
		}

		sysUser, e := vm.Execute(inst.Sequence{
			inst.Constant{val.String("self")},
			inst.Constant{val.StructFromMap(map[string]val.Value{
				"username": val.String("admin"),
				"password": val.String(""),
				"roles":    val.List{sysRole},
			})},
			inst.BuildMap{1},
			inst.CreateMultiple{vm.UserModelId()},
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
			typed, e := vm.Parse(v.Value, AnyModel, AnyModel)
			if e != nil {
				return err.ExecutionError{
					Problem: `there was an error compiling the expression to be persisted`,
					Child_:  e,
				}
			}
			v.Value = xpr.ValueFromExpression(typed)
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

		if edges := extractRefs(v.Value); len(edges) > 0 {

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
			// and do not contain: CRUD, static/contextual
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

					transformation, e := findAutoTransformation(sourceModel, targetModel)
					if e != nil {
						return err.ExecutionError{
							fmt.Sprintf(`failed inferring automatic migration for source model "%s" to target model "%s"`, sourceMID, targetMID),
							nil,
							// C: val.Map{
							// 	"source": source,
							// 	"target": target,
							// 	"error":  e.ToValue(),
							// },
						}
					}

					transformation = xpr.With{
						Value:  xpr.AssertModelRef{Value: xpr.Argument{}, Ref: xpr.Literal{source}},
						Return: transformation,
					}

					xid := common.RandomId()

					expr = xpr.ValueFromExpression(transformation)

					e = vm.Write(vm.ExpressionModelId(), map[string]val.Meta{
						xid: vm.WrapValueInMeta(expr, xid, vm.ExpressionModelId()),
					})
					if e != nil {
						return e
					}

					exprRef = val.Ref{vm.ExpressionModelId(), xid}
					expr = expr

					object.Set("expression", val.Union{"manual", exprRef})

				} else {

					exprRef = expression.Value.(val.Ref)

					exprMeta, e := vm.Get(exprRef[0], exprRef[1])
					if e != nil {
						return e
					}

					expr = exprMeta.Value

					// TODO: either revise mdl.Fits definition or truncate information content of outModel to targetModel
					//       in case of manual migration (auto is guaranteed to deliver sensible result)

					// TODO: check that expr doesn't create, delete or update data
				}

				instructions, outModel, e := vm.ParseAndCompile(expr, sourceModel, targetModel.Model)
				if e != nil {
					return e
				}

				if e := checkType(outModel, targetModel.Model); e != nil {
					return err.ExecutionError{
						Problem: `migration expression output does not fit target model`,
						Child_:  e,
					}
				}

				{ // register migration
					sourceBucket, e := migs.CreateBucketIfNotExists([]byte(sourceMID))
					if e != nil {
						log.Panicln(e)
					}
					if e := sourceBucket.Put([]byte(targetMID), []byte(exprRef[1])); e != nil {
						log.Panicln(e)
					}
					targetBucket, e := sgim.CreateBucketIfNotExists([]byte(targetMID))
					if e != nil {
						log.Panicln(e)
					}
					if e := targetBucket.Put([]byte(sourceMID), []byte(exprRef[1])); e != nil {
						log.Panicln(e)
					}
				}

				{ // actual object migration loop

					todo := make([]string, 0, 16)
					done := make(map[string]struct{})
					bundle := make(map[string]val.Meta, 16) // interdependent target objects
					sourceBucket := db.Bucket([]byte(sourceMID))

					e := sourceBucket.ForEach(func(k, _ []byte) error {
						id := string(k)
						if _, ok := done[id]; ok {
							return nil
						}
						todo = append(todo, id)
						for len(todo) > 0 {
							id := todo[0]
							todo = todo[1:]
							bs := sourceBucket.Get([]byte(id))
							v, _ := karma.Decode(bs, vm.WrapModelInMeta(sourceMID, sourceModel.Model))
							mv := DematerializeMeta(v.(val.Struct))
							migrated, e := vm.Execute(instructions, mv)
							if e != nil {
								return e
							}
							migrated, e = slurpIterators(migrated)
							if e != nil {
								return e
							}
							for _, rf := range extractRefs(migrated) {
								if rf[0] == targetMID {
									if _, ok := done[rf[1]]; !ok {
										todo = append(todo, rf[1])
									}
								}
							}
							out := mv.Copy().(val.Meta)
							out.Id = val.Ref{targetMID, id}
							out.Model = val.Ref{vm.MetaModelId(), targetMID}
							out.Value = migrated
							bundle[id] = out
							done[id] = struct{}{}
						}
						if e := vm.Write(targetMID, bundle); e != nil {
							return e
						}
						for k, _ := range bundle { // clear bundle
							delete(bundle, k)
						}
						return nil
					})

					if e != nil {
						if ke, ok := e.(err.Error); ok {
							return err.ExecutionError{
								Problem: `error during migration`,
								Child_:  ke,
								// C: val.Map{
								// 	"source": source,
								// 	"target": target,
								// 	"error":  ke.ToValue(),
								// },
							}
						} else {
							log.Panicln(e)
						}
					}
				}

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
	Migration val.Value // an expression
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

		it, mm, e := vm.ParseAndCompile(node.Migration, node.InModel, nil)
		if e != nil {
			log.Panicln(pretty.Sprint(e)) // TODO: should theoretically never happen, but... add lots of contextual info in case it does
		}

		vl, e := vm.Execute(it, node.InValue)
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
			inst.Identity{},
			inst.Deref{},
			inst.Field{Key: "permissions"},
			inst.MapStruct{inst.Sequence{
				inst.Identity{},
				inst.Field{Key: "value"},
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
	im := mdl.Any{}
	ci, cm, ce := vm.ParseAndCompile(orExpressions(c), im, mdl.Bool{})
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
	ri, rm, re := vm.ParseAndCompile(orExpressions(r), im, mdl.Bool{})
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
	ui, um, ue := vm.ParseAndCompile(orExpressions(u), im, mdl.Bool{})
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
	di, dm, de := vm.ParseAndCompile(orExpressions(d), im, mdl.Bool{})
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
		create: flattenSequences(ci, nil),
		read:   flattenSequences(ri, nil),
		update: flattenSequences(ui, nil),
		delete: flattenSequences(di, nil),
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
	switch m := m.(type) {
	case mdl.Or:
		if m[0] == (mdl.Null{}) {
			return m[1]
		}
		if m[1] == (mdl.Null{}) {
			return m[0]
		}
		return mdl.Or{
			deoptionalize(m[0]),
			deoptionalize(m[1]),
		}
	}
	return m
}

func modelTypeKey(m mdl.Model) string {
	switch m := m.Concrete().(type) {
	case mdl.Or:
		return modelTypeKey(m[0]) + "|" + modelTypeKey(m[1])
	case mdl.Null:
		return "null"
	case mdl.Set:
		return "set"
	case mdl.List:
		return "list"
	case mdl.Map:
		return "map"
	case mdl.Tuple:
		return "tuple"
	case mdl.Struct:
		return "struct"
	case mdl.Union:
		return "union"
	case mdl.String:
		return "string"
	case mdl.Enum:
		return "enum"
	case mdl.Float:
		return "float"
	case mdl.Bool:
		return "bool"
	case mdl.Any:
		return "any"
	case mdl.Ref:
		return "ref"
	case mdl.DateTime:
		return "dateTime"
	case mdl.Int8:
		return "int8"
	case mdl.Int16:
		return "int16"
	case mdl.Int32:
		return "int32"
	case mdl.Int64:
		return "int64"
	case mdl.Uint8:
		return "uint8"
	case mdl.Uint16:
		return "uint16"
	case mdl.Uint32:
		return "uint32"
	case mdl.Uint64:
		return "uint64"
	}
	panic(fmt.Sprintf(`unhandled modelTypeKey case: %T`, m))
}

func valueTypeKey(v val.Value) string {
	if v == val.Null {
		return "null"
	}
	switch v.(type) {
	case val.Raw:
		return "unknown"
	case val.Set:
		return "set"
	case val.List:
		return "list"
	case val.Map:
		return "map"
	case val.Tuple:
		return "tuple"
	case val.Struct:
		return "struct"
	case val.Union:
		return "union"
	case val.String:
		return "string"
	case val.Symbol:
		return "enum"
	case val.Float:
		return "float"
	case val.Bool:
		return "bool"
	case val.Ref:
		return "ref"
	case val.DateTime:
		return "dateTime"
	case val.Int8:
		return "int8"
	case val.Int16:
		return "int16"
	case val.Int32:
		return "int32"
	case val.Int64:
		return "int64"
	case val.Uint8:
		return "uint8"
	case val.Uint16:
		return "uint16"
	case val.Uint32:
		return "uint32"
	case val.Uint64:
		return "uint64"
	}
	panic(fmt.Sprintf(`unhandled valueTypeKey case: %T`, v))
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
