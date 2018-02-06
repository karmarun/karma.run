// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package xpr

import (
	"karma.run/common"
)

var LanguageModel = common.QuickModel(model)

const model = `
[
    "recursion", {
        "label": "x",
        "model": [
            "union", {

                "null":     ["null", {}],
                "bool":     ["bool", {}],
                "dateTime": ["dateTime", {}],
                "string":   ["string", {}],
                "float":    ["float", {}],
                "int8":     ["int8", {}],
                "int16":    ["int16", {}],
                "int32":    ["int32", {}],
                "int64":    ["int64", {}],
                "uint8":    ["uint8", {}],
                "uint16":   ["uint16", {}],
                "uint32":   ["uint32", {}],
                "uint64":   ["uint64", {}],
                "symbol":   ["string", {}],


                "map":      ["map",   ["recurse", "x"]],
                "list":     ["list",  ["recurse", "x"]],
                "set":      ["set",   ["recurse", "x"]],
                "struct":   ["map",   ["recurse", "x"]],
                "tuple":    ["list",  ["recurse", "x"]],
                "union":    ["tuple", [["string", {}], ["recurse", "x"]]],

                "ref":      ["tuple", [["recurse", "x"], ["recurse", "x"]]],

                "id":          ["struct", {}],
                "arg":         ["struct", {}],
                "currentUser": ["struct", {}],

                "all":            ["recurse", "x"],
                "assertPresent":  ["recurse", "x"],
                "delete":         ["recurse", "x"],
                "extractStrings": ["recurse", "x"],
                "first":          ["recurse", "x"],
                "floatToInt":     ["recurse", "x"],
                "get":            ["recurse", "x"],
                "intToFloat":     ["recurse", "x"],
                "isPresent":      ["recurse", "x"],
                "length":         ["recurse", "x"],
                "metarialize":    ["recurse", "x"],
                "model":          ["recurse", "x"],
                "modelOf":        ["recurse", "x"],
                "not":            ["recurse", "x"],
                "presentOrZero":  ["recurse", "x"],
                "refTo":          ["recurse", "x"],
                "resolveAllRefs": ["recurse", "x"],
                "reverseList":    ["recurse", "x"],
                "stringToLower":  ["recurse", "x"],
                "tag":            ["recurse", "x"],
                "zero":           ["recurse", "x"],
                "allReferrers":   ["recurse", "x"],

                "after":       ["tuple", [["recurse","x"], ["recurse","x"]]],
                "before":      ["tuple", [["recurse","x"], ["recurse","x"]]],
                "less":        ["tuple", [["recurse","x"], ["recurse","x"]]],
                "greater":     ["tuple", [["recurse","x"], ["recurse","x"]]],
                "add":         ["tuple", [["recurse","x"], ["recurse","x"]]],
                "subtract":    ["tuple", [["recurse","x"], ["recurse","x"]]],
                "multiply":    ["tuple", [["recurse","x"], ["recurse","x"]]],
                "divide":      ["tuple", [["recurse","x"], ["recurse","x"]]],
                "equal":       ["tuple", [["recurse","x"], ["recurse","x"]]],
                "concatLists": ["tuple", [["recurse","x"], ["recurse","x"]]],

                "setField": ["struct", {
                    "name":  ["string", {}],
                    "value": ["recurse", "x"],
                    "in":    ["recurse", "x"]
                }],

                "setKey": ["struct", {
                    "name":  ["string", {}],
                    "value": ["recurse", "x"],
                    "in":    ["recurse", "x"]
                }],

                "field": ["struct", {
                    "name":  ["string", {}],
                    "value": ["recurse", "x"]
                }],

                "key": ["struct", {
                    "name":  ["string", {}],
                    "value": ["recurse", "x"]
                }],

                "relocateRef": ["struct", {
                    "ref":   ["recurse", "x"],
                    "model": ["recurse", "x"]
                }],

                "referrers": ["struct", {
                    "of": ["recurse", "x"],
                    "in": ["recurse", "x"]
                }],

                "referred": ["struct", {
                    "from": ["recurse", "x"],
                    "in":   ["recurse", "x"]
                }],

                "inList": ["struct", {
                    "value": ["recurse", "x"],
                    "in":    ["recurse", "x"]
                }],

                "resolveRefs": ["struct", {
                    "value":  ["recurse", "x"],
                    "models": ["set", ["recurse", "x"]]
                }],

                "matchRegex": ["struct", {
                    "value":           ["recurse", "x"],
                    "regex":           ["string", {}],
                    "caseInsensitive": ["bool", {}],
                    "multiLine":       ["bool", {}]
                }],

                "searchRegex": ["struct", {
                    "value":           ["recurse", "x"],
                    "regex":           ["string", {}],
                    "caseInsensitive": ["bool", {}],
                    "multiLine":       ["bool", {}]
                }],

                "searchAllRegex": ["struct", {
                    "value":           ["recurse", "x"],
                    "regex":           ["string", {}],
                    "caseInsensitive": ["bool", {}],
                    "multiLine":       ["bool", {}]
                }],

                "slice": ["struct", {
                    "value":  ["recurse", "x"],
                    "offset": ["recurse", "x"],
                    "length": ["recurse", "x"]
                }],

                "graphFlow": ["struct", {
                    "start": ["recurse", "x"],
                    "flow":  ["set", ["struct", {
                        "from":     ["recurse", "x"],
                        "forward":  ["set", ["recurse", "x"]],
                        "backward": ["set", ["recurse", "x"]]
                    }]]
                }],

                "assertModelRef": ["struct", {
                    "value": ["recurse", "x"],
                    "ref" :  ["recurse", "x"]
                }],

                "switchCase": ["struct", {
                    "value":   ["recurse", "x"],
                    "default": ["recurse", "x"],
                    "cases":   ["map", ["recurse", "x"]]
                }],

                "switchModelRef": ["struct", {
                    "value":   ["recurse", "x"],
                    "default": ["recurse", "x"],
                    "cases":   ["set", ["struct", {
                        "match": ["recurse", "x"],
                        "return": ["recurse", "x"]
                    }]]
                }],

                "if": ["struct", {
                    "condition": ["recurse", "x"],
                    "then":      ["recurse", "x"],
                    "else":      ["recurse", "x"]
                }],

                "with": ["struct", {
                    "value":  ["recurse", "x"],
                    "return": ["recurse", "x"]
                }],

                "assertCase": ["struct", {
                    "value": ["recurse", "x"],
                    "case":  ["string", {}]
                }],

                "isCase": ["struct", {
                    "value": ["recurse", "x"],
                    "case":  ["recurse", "x"]
                }],

                "memSort": ["struct", {
                    "value":      ["recurse", "x"],
                    "expression": ["recurse", "x"]
                }],

                "mapSet": ["struct", {
                    "value":      ["recurse", "x"],
                    "expression": ["recurse", "x"]
                }],

                "mapList": ["struct", {
                    "value":      ["recurse", "x"],
                    "expression": ["recurse", "x"]
                }],

                "mapMap": ["struct", {
                    "value":      ["recurse", "x"],
                    "expression": ["recurse", "x"]
                }],

                "filterList": ["struct", {
                    "value":      ["recurse", "x"],
                    "expression": ["recurse", "x"]
                }],

                "reduceList": ["struct", {
                    "value":      ["recurse", "x"],
                    "expression": ["recurse", "x"]
                }],

                "indexTuple": ["struct", {
                    "value":  ["recurse", "x"],
                    "number": ["int64", {}]
                }],

                "and": ["set",["recurse", "x"]],

                "or": ["set",["recurse", "x"]],

                "create": ["struct", {
                    "in":    ["recurse", "x"],
                    "value": ["recurse", "x"]
                }],

                "update": ["struct", {
                    "ref":   ["recurse", "x"],
                    "value": ["recurse", "x"]
                }],

                "createMultiple": ["struct", {
                    "in":     ["recurse", "x"],
                    "values": ["map", ["recurse", "x"]]
                }],

                "joinStrings": ["struct", {
                    "strings":   ["recurse", "x"],
                    "separator": ["recurse", "x"]
                }]


            }
        ]
    }
]
`

// note: switchType left out when ors were deprecated
//  filter -> filterList
//  index  -> indexTuple
//  some functions that used to take lists now take sets where it makes sense
