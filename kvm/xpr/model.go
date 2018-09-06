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
				// primitive
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

				// composite
				"map":    mdl.Map{data},
				"list":   mdl.List{data},
				"set":    mdl.Set{data},
				"struct": mdl.Map{data},
				"tuple":  mdl.List{data},
				"union":  mdl.Tuple{mdl.String{}, data},
				"ref":    mdl.Tuple{mdl.String{}, mdl.String{}},
				"expr":   expression,
			})
		})

		return mdl.UnionFromMap(map[string]mdl.Model{

			// convenience primitive constructors
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

			"data":      data,
			"signature": function,

			"define": mdl.Tuple{mdl.String{}, expression},
			"scope":  mdl.String{},

			"dateTimeNow":    mdl.EmptyStruct,
			"currentUser":    mdl.EmptyStruct,
			"all":            expression,
			"assertPresent":  expression,
			"delete":         expression,
			"extractStrings": expression,
			"first":          expression,
			"get":            expression,
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
			"allReferrers":   expression,
			"tagExists":      expression,
			"toFloat":        expression,
			"toInt8":         expression,
			"toInt16":        expression,
			"toInt32":        expression,
			"toInt64":        expression,
			"toUint8":        expression,
			"toUint16":       expression,
			"toUint32":       expression,
			"toUint64":       expression,
			"zero":           mdl.EmptyStruct,

			"stringContains":  mdl.Tuple{expression, expression},
			"substringIndex":  mdl.Tuple{expression, expression},
			"memSortFunction": mdl.Tuple{expression, function},

			"leftFoldList":  mdl.Tuple{expression, expression, function}, // (list, initial, reducer)
			"rightFoldList": mdl.Tuple{expression, expression, function}, // (list, initial, reducer)
			// "someList":      mdl.Tuple{expression, function},
			// "everyList":     mdl.Tuple{expression, function},

			"addFloat":  mdl.Tuple{expression, expression},
			"addInt64":  mdl.Tuple{expression, expression},
			"addInt32":  mdl.Tuple{expression, expression},
			"addInt16":  mdl.Tuple{expression, expression},
			"addInt8":   mdl.Tuple{expression, expression},
			"addUint64": mdl.Tuple{expression, expression},
			"addUint32": mdl.Tuple{expression, expression},
			"addUint16": mdl.Tuple{expression, expression},
			"addUint8":  mdl.Tuple{expression, expression},

			"subFloat":  mdl.Tuple{expression, expression},
			"subInt64":  mdl.Tuple{expression, expression},
			"subInt32":  mdl.Tuple{expression, expression},
			"subInt16":  mdl.Tuple{expression, expression},
			"subInt8":   mdl.Tuple{expression, expression},
			"subUint64": mdl.Tuple{expression, expression},
			"subUint32": mdl.Tuple{expression, expression},
			"subUint16": mdl.Tuple{expression, expression},
			"subUint8":  mdl.Tuple{expression, expression},

			"divFloat":  mdl.Tuple{expression, expression},
			"divInt64":  mdl.Tuple{expression, expression},
			"divInt32":  mdl.Tuple{expression, expression},
			"divInt16":  mdl.Tuple{expression, expression},
			"divInt8":   mdl.Tuple{expression, expression},
			"divUint64": mdl.Tuple{expression, expression},
			"divUint32": mdl.Tuple{expression, expression},
			"divUint16": mdl.Tuple{expression, expression},
			"divUint8":  mdl.Tuple{expression, expression},

			"mulFloat":  mdl.Tuple{expression, expression},
			"mulInt64":  mdl.Tuple{expression, expression},
			"mulInt32":  mdl.Tuple{expression, expression},
			"mulInt16":  mdl.Tuple{expression, expression},
			"mulInt8":   mdl.Tuple{expression, expression},
			"mulUint64": mdl.Tuple{expression, expression},
			"mulUint32": mdl.Tuple{expression, expression},
			"mulUint16": mdl.Tuple{expression, expression},
			"mulUint8":  mdl.Tuple{expression, expression},

			"gtFloat":  mdl.Tuple{expression, expression},
			"gtInt64":  mdl.Tuple{expression, expression},
			"gtInt32":  mdl.Tuple{expression, expression},
			"gtInt16":  mdl.Tuple{expression, expression},
			"gtInt8":   mdl.Tuple{expression, expression},
			"gtUint64": mdl.Tuple{expression, expression},
			"gtUint32": mdl.Tuple{expression, expression},
			"gtUint16": mdl.Tuple{expression, expression},
			"gtUint8":  mdl.Tuple{expression, expression},

			"ltFloat":  mdl.Tuple{expression, expression},
			"ltInt64":  mdl.Tuple{expression, expression},
			"ltInt32":  mdl.Tuple{expression, expression},
			"ltInt16":  mdl.Tuple{expression, expression},
			"ltInt8":   mdl.Tuple{expression, expression},
			"ltUint64": mdl.Tuple{expression, expression},
			"ltUint32": mdl.Tuple{expression, expression},
			"ltUint16": mdl.Tuple{expression, expression},
			"ltUint8":  mdl.Tuple{expression, expression},

			"after":       mdl.Tuple{expression, expression},
			"before":      mdl.Tuple{expression, expression},
			"equal":       mdl.Tuple{expression, expression},
			"concatLists": mdl.Tuple{expression, expression},
			"field":       mdl.Tuple{mdl.String{}, expression},
			"key":         mdl.Tuple{expression, expression},

			"with":    mdl.Tuple{expression, function},
			"mapSet":  mdl.Tuple{expression, function},
			"mapList": mdl.Tuple{expression, function},
			"mapMap":  mdl.Tuple{expression, function},

			"and": mdl.List{expression},
			"or":  mdl.List{expression},

			"create":     mdl.Tuple{expression, function},
			"filterList": mdl.Tuple{expression, function},
			"memSort":    mdl.Tuple{expression, function},

			"createMultiple": mdl.Tuple{expression, mdl.Map{function}},
			"resolveRefs":    mdl.Tuple{expression, mdl.Set{expression}},

			"indexTuple": mdl.Tuple{expression, mdl.Int64{}},

			"switchCase": mdl.Tuple{expression, mdl.Map{function}},

			"reduceList": mdl.StructFromMap(map[string]mdl.Model{
				"value":   expression,
				"initial": expression,
				"reducer": function,
			}),

			"setField": mdl.StructFromMap(map[string]mdl.Model{
				"name":  mdl.String{},
				"value": expression,
				"in":    expression,
			}),
			"setKey": mdl.StructFromMap(map[string]mdl.Model{
				"name":  mdl.String{},
				"value": expression,
				"in":    expression,
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

			"assertCase": mdl.StructFromMap(map[string]mdl.Model{
				"case":  mdl.String{},
				"value": expression,
			}),
			"isCase": mdl.StructFromMap(map[string]mdl.Model{
				"value": expression,
				"case":  expression,
			}),

			"update": mdl.StructFromMap(map[string]mdl.Model{
				"ref":   expression,
				"value": expression,
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
