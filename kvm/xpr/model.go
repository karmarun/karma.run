// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package xpr

import (
	"karma.run/kvm/mdl"
)

var LanguageModel = mdl.DefineRecursion("function", func(function *mdl.Recursion) mdl.Model {

	expression := mdl.DefineRecursion("expression", func(expression *mdl.Recursion) mdl.Model {

		data := mdl.DefineRecursion("data", func(data *mdl.Recursion) mdl.Model {
			return mdl.UnionFromMap(map[string]mdl.Model{
				"null":     mdl.Null{},
				"bool":     mdl.Bool{},
				"dateTime": mdl.DateTime{},
				"string":   mdl.String{},
				"float":    mdl.Float{},
				"int8":     mdl.Int8{},
				"int16":    mdl.Int16{},
				"int32":    mdl.Int32{},
				"int64":    mdl.Int64{},
				"uint8":    mdl.Uint8{},
				"uint16":   mdl.Uint16{},
				"uint32":   mdl.Uint32{},
				"uint64":   mdl.Uint64{},
				"symbol":   mdl.String{},
				"map":      mdl.Map{data},
				"list":     mdl.List{data},
				"set":      mdl.Set{data},
				"struct":   mdl.Map{data},
				"tuple":    mdl.List{data},
				"union":    mdl.Tuple{mdl.String{}, data},
				"ref":      mdl.Tuple{mdl.String{}, mdl.String{}},
				"expr":     expression,
			})
		})

		return mdl.UnionFromMap(map[string]mdl.Model{

			"data":      data,
			"signature": function,

			"define": mdl.Tuple{mdl.String{}, expression},
			"scope":  mdl.String{},

			"id":             mdl.EmptyStruct,
			"arg":            mdl.EmptyStruct,
			"currentUser":    mdl.EmptyStruct,
			"all":            expression,
			"assertPresent":  expression,
			"delete":         expression,
			"extractStrings": expression,
			"first":          expression,
			"floatToInt":     expression,
			"get":            expression,
			"intToFloat":     expression,
			"isPresent":      expression,
			"length":         expression,
			"metarialize":    expression,
			"model":          expression,
			"modelOf":        expression,
			"not":            expression,
			"presentOrZero":  expression,
			"refTo":          expression,
			"resolveAllRefs": expression,
			"reverseList":    expression,
			"stringToLower":  expression,
			"tag":            expression,
			"zero":           expression,
			"allReferrers":   expression,

			"addInt64":  mdl.Tuple{expression, expression},
			"addInt32":  mdl.Tuple{expression, expression},
			"addInt16":  mdl.Tuple{expression, expression},
			"addInt8":   mdl.Tuple{expression, expression},
			"addUint64": mdl.Tuple{expression, expression},
			"addUint32": mdl.Tuple{expression, expression},
			"addUint16": mdl.Tuple{expression, expression},
			"addUint8":  mdl.Tuple{expression, expression},

			"after":       mdl.Tuple{expression, expression},
			"before":      mdl.Tuple{expression, expression},
			"equal":       mdl.Tuple{expression, expression},
			"concatLists": mdl.Tuple{expression, expression},
			"setField": mdl.StructFromMap(map[string]mdl.Model{
				"name":  mdl.String{},
				"value": expression,
				"in":    expression,
			}),
			"and": mdl.Set{expression},
			"or":  mdl.Set{expression},
			"setKey": mdl.StructFromMap(map[string]mdl.Model{
				"name":  mdl.String{},
				"value": expression,
				"in":    expression,
			}),
			"field": mdl.StructFromMap(map[string]mdl.Model{
				"name":  mdl.String{},
				"value": expression,
			}),
			"key": mdl.StructFromMap(map[string]mdl.Model{
				"name":  mdl.String{},
				"value": expression,
			}),
			"relocateRef": mdl.StructFromMap(map[string]mdl.Model{
				"ref":   expression,
				"model": expression,
			}),
			"referrers": mdl.StructFromMap(map[string]mdl.Model{
				"of": expression,
				"in": expression,
			}),
			"referred": mdl.StructFromMap(map[string]mdl.Model{
				"from": expression,
				"in":   expression,
			}),
			"inList": mdl.StructFromMap(map[string]mdl.Model{
				"value": expression,
				"in":    expression,
			}),
			"resolveRefs": mdl.StructFromMap(map[string]mdl.Model{
				"value":  expression,
				"models": mdl.Set{expression},
			}),
			"matchRegex": mdl.StructFromMap(map[string]mdl.Model{
				"value":           expression,
				"regex":           mdl.String{},
				"caseInsensitive": mdl.Bool{},
				"multiLine":       mdl.Bool{},
			}),
			"searchRegex": mdl.StructFromMap(map[string]mdl.Model{
				"value":           expression,
				"regex":           mdl.String{},
				"caseInsensitive": mdl.Bool{},
				"multiLine":       mdl.Bool{},
			}),
			"searchAllRegex": mdl.StructFromMap(map[string]mdl.Model{
				"value":           expression,
				"regex":           mdl.String{},
				"caseInsensitive": mdl.Bool{},
				"multiLine":       mdl.Bool{},
			}),
			"slice": mdl.StructFromMap(map[string]mdl.Model{
				"value":  expression,
				"offset": expression,
				"length": expression,
			}),
			"graphFlow": mdl.StructFromMap(map[string]mdl.Model{
				"start": expression,
				"flow": mdl.Set{mdl.StructFromMap(map[string]mdl.Model{
					"from":     expression,
					"forward":  mdl.Set{expression},
					"backward": mdl.Set{expression},
				})},
			}),
			"assertModelRef": mdl.StructFromMap(map[string]mdl.Model{
				"value": expression,
				"ref":   expression,
			}),
			"switchCase": mdl.StructFromMap(map[string]mdl.Model{
				"value":   expression,
				"default": expression,
				"cases":   mdl.Map{expression},
			}),
			"switchModelRef": mdl.StructFromMap(map[string]mdl.Model{
				"value":   expression,
				"default": expression,
				"cases": mdl.Set{mdl.StructFromMap(map[string]mdl.Model{
					"match":  expression,
					"return": expression,
				})},
			}),
			"if": mdl.StructFromMap(map[string]mdl.Model{
				"condition": expression,
				"then":      expression,
				"else":      expression,
			}),
			"with": mdl.StructFromMap(map[string]mdl.Model{
				"value":  expression,
				"return": expression,
			}),
			"assertCase": mdl.StructFromMap(map[string]mdl.Model{
				"value": expression,
				"case":  mdl.String{},
			}),
			"isCase": mdl.StructFromMap(map[string]mdl.Model{
				"value": expression,
				"case":  expression,
			}),
			"memSort": mdl.StructFromMap(map[string]mdl.Model{
				"value":      expression,
				"expression": expression,
			}),
			"mapSet": mdl.StructFromMap(map[string]mdl.Model{
				"value":      expression,
				"expression": expression,
			}),
			"mapList": mdl.StructFromMap(map[string]mdl.Model{
				"value":      expression,
				"expression": function,
			}),
			"mapMap": mdl.StructFromMap(map[string]mdl.Model{
				"value":      expression,
				"expression": function,
			}),
			"filterList": mdl.StructFromMap(map[string]mdl.Model{
				"value":      expression,
				"expression": expression,
			}),
			"reduceList": mdl.StructFromMap(map[string]mdl.Model{
				"value":      expression,
				"expression": expression,
			}),
			"indexTuple": mdl.StructFromMap(map[string]mdl.Model{
				"value":  expression,
				"number": mdl.Int64{},
			}),
			"create": mdl.StructFromMap(map[string]mdl.Model{
				"in":    expression,
				"value": expression,
			}),
			"update": mdl.StructFromMap(map[string]mdl.Model{
				"ref":   expression,
				"value": expression,
			}),
			"createMultiple": mdl.StructFromMap(map[string]mdl.Model{
				"in":     expression,
				"values": mdl.Map{expression},
			}),
			"joinStrings": mdl.StructFromMap(map[string]mdl.Model{
				"strings":   expression,
				"separator": expression,
			}),
		})
	})

	arguments := mdl.List{mdl.String{}}

	return mdl.UnionFromMap(map[string]mdl.Model{
		"function": mdl.Tuple{arguments, mdl.List{expression}},
	})
})
