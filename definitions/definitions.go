// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package definitions

import (
	"karma.run/kvm/val"
)

const (
	MetaModel       = `MetaModel`
	TagModel        = `TagModel`
	TagBucket       = `TagBucket`
	GraphBucket     = `GraphBucket`
	PhargBucket     = `PhargBucket` // (inverse GraphBucket)
	MigrationBucket = `MigrationBucket`
	NoitargimBucket = `NoitargimBucket`
	MigrationModel  = `MigrationModel`
	ExpressionModel = `ExpressionModel`
	UserModel       = `UserModel`
	RoleModel       = `RoleModel`
	RootUser        = `RootUser`
)

var (
	MetaModelBytes       = []byte(MetaModel)
	TagModelBytes        = []byte(TagModel)
	TagBucketBytes       = []byte(TagBucket)
	GraphBucketBytes     = []byte(GraphBucket)
	PhargBucketBytes     = []byte(PhargBucket)
	MigrationBucketBytes = []byte(MigrationBucket)
	NoitargimBucketBytes = []byte(NoitargimBucket)
	MigrationModelBytes  = []byte(MigrationModel)
	ExpressionModelBytes = []byte(ExpressionModel)
	UserModelBytes       = []byte(UserModel)
	RoleModelBytes       = []byte(RoleModel)
	RootUserBytes        = []byte(RootUser)
)

func NewMetaModelValue(metaId string) val.Value {
	return val.Union{"recursion", val.StructFromMap(map[string]val.Value{
		"label": val.String("model"),
		"model": val.Union{"union", val.MapFromMap(map[string]val.Value{
			"recursive": val.Union{"struct", val.MapFromMap(map[string]val.Value{
				"top":    val.Union{"string", val.Struct{}},
				"models": val.Union{"map", val.Union{"recurse", val.String("model")}},
			})},
			"recursion": val.Union{"struct", val.MapFromMap(map[string]val.Value{
				"label": val.Union{"string", val.Struct{}},
				"model": val.Union{"recurse", val.String("model")},
			})},
			"recurse": val.Union{"string", val.Struct{}},
			"annotation": val.Union{"struct", val.MapFromMap(map[string]val.Value{
				"value": val.Union{"string", val.Struct{}},
				"model": val.Union{"recurse", val.String("model")},
			})},
			"set":      val.Union{"recurse", val.String("model")},
			"list":     val.Union{"recurse", val.String("model")},
			"tuple":    val.Union{"list", val.Union{"recurse", val.String("model")}},
			"struct":   val.Union{"map", val.Union{"recurse", val.String("model")}},
			"union":    val.Union{"map", val.Union{"recurse", val.String("model")}},
			"ref":      val.Union{"ref", val.Ref{metaId, metaId}},
			"map":      val.Union{"recurse", val.String("model")},
			"optional": val.Union{"recurse", val.String("model")},
			"enum":     val.Union{"set", val.Union{"string", val.Struct{}}},
			"bool":     val.Union{"struct", val.Map{}},
			"dateTime": val.Union{"struct", val.Map{}},
			"float":    val.Union{"struct", val.Map{}},
			"string":   val.Union{"struct", val.Map{}},
			"int8":     val.Union{"struct", val.Map{}},
			"int16":    val.Union{"struct", val.Map{}},
			"int32":    val.Union{"struct", val.Map{}},
			"int64":    val.Union{"struct", val.Map{}},
			"uint8":    val.Union{"struct", val.Map{}},
			"uint16":   val.Union{"struct", val.Map{}},
			"uint32":   val.Union{"struct", val.Map{}},
			"uint64":   val.Union{"struct", val.Map{}},
			"null":     val.Union{"struct", val.Map{}},
		})},
	})}
}

func NewTagModelValue(metaId string) val.Value {
	return val.Union{"struct", val.MapFromMap(map[string]val.Value{
		"tag":   val.Union{"string", val.Struct{}},
		"model": val.Union{"ref", val.Ref{metaId, metaId}},
	})}
}

func NewUserModelValue(metaId, roleId string) val.Value {
	return val.Union{"struct", val.MapFromMap(map[string]val.Value{
		"username": val.Union{"string", val.Struct{}},
		"password": val.Union{"string", val.Struct{}},
		"roles":    val.Union{"list", val.Union{"ref", val.Ref{metaId, roleId}}},
	})}
}

func NewRoleModelValue(metaId, exprId string) val.Value {
	return val.Union{"struct", val.MapFromMap(map[string]val.Value{
		"name": val.Union{"string", val.Struct{}},
		"permissions": val.Union{"struct", val.MapFromMap(map[string]val.Value{
			"create": val.Union{"ref", val.Ref{metaId, exprId}},
			"read":   val.Union{"ref", val.Ref{metaId, exprId}},
			"update": val.Union{"ref", val.Ref{metaId, exprId}},
			"delete": val.Union{"ref", val.Ref{metaId, exprId}},
		})},
	})}
}

func NewMigrationModelValue(metaId, exprId string) val.Value {
	return val.Union{"list", val.Union{"struct", val.MapFromMap(map[string]val.Value{
		"source": val.Union{"ref", val.Ref{metaId, metaId}},
		"target": val.Union{"ref", val.Ref{metaId, metaId}},
		"expression": val.Union{"union", val.MapFromMap(map[string]val.Value{
			"auto":   val.Union{"struct", val.Map{}},
			"manual": val.Union{"ref", val.Ref{metaId, exprId}},
		})},
	})}}
}
