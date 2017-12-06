// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package xpr

import (
	"common"
)

var LanguageModel = common.QuickModel(model)

const model = `
{
    "recursive": {
        "top": "top",
        "models": {
            "top": {
                "or": [
                    {
                        "recurse": "primitive"
                    },
                    {
                        "recurse": "primitive-constructor"
                    },
                    {
                        "recurse": "composite-constructor"
                    },
                    {
                        "recurse": "expression"
                    }
                ]
            },
            "expression": {
                "union": {
                    "tag": {
                        "recurse": "top"
                    },
                    "model": {
                        "recurse": "top"
                    },
                    "reverseList": {
                        "recurse": "top"
                    },
                    "stringToLower": {
                        "recurse": "top"
                    },
                    "setField": {
                        "struct": {
                            "name": {
                                "recurse": "top"
                            },
                            "value": {
                                "recurse": "top"
                            },
                            "in": {
                                "recurse": "top"
                            }
                        }
                    },
                    "setKey": {
                        "struct": {
                            "name": {
                                "recurse": "top"
                            },
                            "value": {
                                "recurse": "top"
                            },
                            "in": {
                                "recurse": "top"
                            }
                        }
                    },
                    "field": {
                        "or": [
                            {
                                "string": {}
                            },
                            {
                                "struct": {
                                    "value": {
                                        "recurse": "top"
                                    },
                                    "name": {
                                        "recurse": "top"
                                    }
                                }
                            }
                        ]
                    },
                    "key": {
                        "or": [
                            {
                                "string": {}
                            },
                            {
                                "struct": {
                                    "value": {
                                        "recurse": "top"
                                    },
                                    "name": {
                                        "or": [
                                            {
                                                "string": {}
                                            },
                                            {
                                                "recurse": "top"
                                            }
                                        ]
                                    }
                                }
                            }
                        ]
                    },
                    "relocateRef": {
                        "struct": {
                            "ref": {
                                "recurse": "expression"
                            },
                            "model": {
                                "recurse": "expression"
                            }
                        }
                    },
                    "metarialize": {
                        "recurse": "expression"
                    },
                    "all": {
                        "or": [
                            {
                                "recurse": "expression"
                            }
                        ]
                    },
                    "assertPresent": {
                        "recurse": "top"
                    },
                    "isPresent": {
                        "recurse": "top"
                    },
                    "presentOrZero": {
                        "recurse": "top"
                    },
                    "referrers": {
                        "struct": {
                            "of": {
                                "recurse": "top"
                            },
                            "in": {
                                "recurse": "top"
                            }
                        }
                    },
                    "referred": {
                        "struct": {
                            "from": {
                                "recurse": "top"
                            },
                            "in": {
                                "recurse": "top"
                            }
                        }
                    },
                    "inList": {
                        "struct": {
                            "value": {
                                "recurse": "top"
                            },
                            "in": {
                                "recurse": "top"
                            }
                        }
                    },
                    "resolveRefs": {
                        "struct": {
                            "value": {
                                "recurse": "top"
                            },
                            "models": {
                                "list": {
                                    "recurse": "top"
                                }
                            }
                        }
                    },
                    "matchRegex": {
                        "or": [
                            {
                                "string": {}
                            },
                            {
                                "struct": {
                                    "value": {
                                        "optional": {
                                            "recurse": "top"
                                        }
                                    },
                                    "regex": {
                                        "recurse": "top"
                                    },
                                    "caseInsensitive": {
                                        "optional": {
                                            "recurse": "top"
                                        }
                                    },
                                    "multiLine": {
                                        "optional": {
                                            "recurse": "top"
                                        }
                                    }
                                }
                            }
                        ]
                    },
                    "searchRegex": {
                        "or": [
                            {
                                "string": {}
                            },
                            {
                                "struct": {
                                    "value": {
                                        "optional": {
                                            "recurse": "top"
                                        }
                                    },
                                    "regex": {
                                        "recurse": "top"
                                    },
                                    "caseInsensitive": {
                                        "optional": {
                                            "recurse": "top"
                                        }
                                    },
                                    "multiLine": {
                                        "optional": {
                                            "recurse": "top"
                                        }
                                    }
                                }
                            }
                        ]
                    },
                    "searchAllRegex": {
                        "or": [
                            {
                                "string": {}
                            },
                            {
                                "struct": {
                                    "value": {
                                        "optional": {
                                            "recurse": "top"
                                        }
                                    },
                                    "regex": {
                                        "recurse": "top"
                                    },
                                    "caseInsensitive": {
                                        "optional": {
                                            "recurse": "top"
                                        }
                                    },
                                    "multiLine": {
                                        "optional": {
                                            "recurse": "top"
                                        }
                                    }
                                }
                            }
                        ]
                    },
                    "slice": {
                        "struct": {
                            "value": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "offset": {
                                "recurse": "top"
                            },
                            "length": {
                                "recurse": "top"
                            }
                        }
                    },
                    "resolveAllRefs": {
                        "recurse": "top"
                    },
                    "graphFlow": {
                        "struct": {
                            "start": {
                                "recurse": "top"
                            },
                            "flow": {
                                "list": {
                                    "struct": {
                                        "from": {
                                            "recurse": "top"
                                        },
                                        "forward": {
                                            "optional": {
                                                "list": {
                                                    "recurse": "top"
                                                }
                                            }
                                        },
                                        "backward": {
                                            "optional": {
                                                "list": {
                                                    "recurse": "top"
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "assertModelRef": {
                        "or": [
                            {
                                "struct": {
                                    "value": {
                                        "optional": {
                                            "recurse": "top"
                                        }
                                    },
                                    "ref": {
                                        "recurse": "top"
                                    }
                                }
                            }
                        ]
                    },
                    "switchType": {
                        "struct": {
                            "value": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "null": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "set": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "list": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "map": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "tuple": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "struct": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "union": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "string": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "enum": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "float": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "bool": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "any": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "ref": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "dateTime": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "int": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "int8": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "int16": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "int32": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "int64": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "uint": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "uint8": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "uint16": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "uint32": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "uint64": {
                                "optional": {
                                    "recurse": "top"
                                }
                            }
                        }
                    },
                    "switchCase":{
                        "struct": {
                            "value": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "cases": {
                                "map": {
                                    "recurse": "top"
                                }
                            }
                        }
                    },
                    "switchModelRef": {
                        "struct": {
                            "value": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "default": {
                                "recurse": "top"
                            },
                            "cases": {
                                "list": {
                                    "struct": {
                                        "match": {
                                            "recurse": "top"
                                        },
                                        "return": {
                                            "recurse": "top"
                                        }
                                    }
                                }
                            }
                        }
                    },
                    "if": {
                        "struct": {
                            "condition": {
                                "recurse": "top"
                            },
                            "then": {
                                "recurse": "top"
                            },
                            "else": {
                                "recurse": "top"
                            }
                        }
                    },
                    "with": {
                        "struct": {
                            "value": {
                                "recurse": "top"
                            },
                            "return": {
                                "recurse": "top"
                            }
                        }
                    },
                    "assertCase": {
                        "struct": {
                            "value": {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            "case": {
                                "or": [
                                    {
                                        "string": {}
                                    },
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        }
                    },
                    "isCase": {
                        "struct": {
                            "value": {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            "case": {
                                "or": [
                                    {
                                        "string": {}
                                    },
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        }
                    },
                    "refTo": {
                        "or": [
                            {
                                "recurse": "top"
                            }
                        ]
                    },
                    "first": {
                        "or": [
                            {
                                "recurse": "top"
                            }
                        ]
                    },
                    "memSort":{
                        "struct": {
                            "value": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "expression": {
                                "recurse": "top"
                            }
                        }
                    },
                    "mapSet":{
                        "struct": {
                            "value": {
                                "optional": {
                                    "recurse": "top"
                                }
                            },
                            "expression": {
                                "recurse": "top"
                            }
                        }
                    },
                    "mapList": {
                        "struct": {
                            "value": {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            "expression": {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        }
                    },
                    "reduceList": {
                        "struct": {
                            "value": {
                                "or": [
                                    {
                                        "list": {
                                            "recurse": "top"
                                        }
                                    },
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            "expression": {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        }
                    },
                    "index": {
                        "struct": {
                            "value": {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            "number": {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        }
                    },
                    "mapMap": {
                        "struct": {
                            "value": {
                                "or": [
                                    {
                                        "recurse": "top"
                                    },
                                    {
                                        "map": {
                                            "recurse": "top"
                                        }
                                    }
                                ]
                            },
                            "expression": {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        }
                    },
                    "filter": {
                        "struct": {
                            "value": {
                                "or": [
                                    {
                                        "recurse": "expression"
                                    }
                                ]
                            },
                            "expression": {
                                "or": [
                                    {
                                        "recurse": "bool"
                                    },
                                    {
                                        "recurse": "expression"
                                    }
                                ]
                            }
                        }
                    },
                    "after": {
                        "tuple": [
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        ]
                    },
                    "before": {
                        "tuple": [
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        ]
                    },
                    "less": {
                        "tuple": [
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        ]
                    },
                    "greater": {
                        "tuple": [
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        ]
                    },
                    "add": {
                        "tuple": [
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        ]
                    },
                    "subtract": {
                        "tuple": [
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        ]
                    },
                    "multiply": {
                        "tuple": [
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        ]
                    },
                    "divide": {
                        "tuple": [
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        ]
                    },
                    "equal": {
                        "or": [
                            {
                                "tuple": [
                                    {
                                        "recurse": "top"
                                    },
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            {
                                "recurse": "top"
                            }
                        ]
                    },
                    "and": {
                        "list": {
                            "or": [
                                {
                                    "recurse": "top"
                                }
                            ]
                        }
                    },
                    "or": {
                        "list": {
                            "or": [
                                {
                                    "recurse": "top"
                                }
                            ]
                        }
                    },
                    "intToFloat": {
                        "recurse": "top"
                    },
                    "floatToInt": {
                        "recurse": "top"
                    },
                    "modelOf": {
                        "recurse": "top"
                    },
                    "not": {
                        "recurse": "top"
                    },
                    "length": {
                        "recurse": "top"
                    },
                    "get": {
                        "or": [
                            {
                                "recurse": "top"
                            }
                        ]
                    },
                    "contextual": {
                        "any": {}
                    },
                    "static": {
                        "any": {}
                    },
                    "create": {
                        "struct": {
                            "in": {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            "value": {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        }
                    },
                    "createMultiple": {
                        "struct": {
                            "in": {
                                "recurse": "top"
                            },
                            "values": {
                                "map": {
                                    "recurse": "top"
                                }
                            }
                        }
                    },
                    "update": {
                        "struct": {
                            "ref": {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            "value": {
                                "or": [
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        }
                    },
                    "delete": {
                        "recurse": "top"
                    },
                    "extractStrings": {
                        "recurse": "top"
                    },
                    "concatLists": {
                        "tuple": [
                            {
                                "recurse": "top"
                            },
                            {
                                "recurse": "top"
                            }
                        ]
                    },
                    "joinStrings": {
                        "struct": {
                            "strings": {
                                "recurse": "top"
                            },
                            "separator": {
                                "optional": {
                                    "recurse": "top"
                                }
                            }
                        }
                    },
                    "id": {
                        "struct": {}
                    },
                    "arg": {
                        "struct": {}
                    },
                    "zero": {
                        "struct": {}
                    },
                    "currentUser": {
                        "struct": {}
                    },
                    "do": {
                        "map": {
                            "recurse": "top"
                        }
                    },
                    "bind": {
                        "string": {}
                    }
                }
            },
            "primitive": {
                "or": [
                    {
                        "recurse": "null"
                    },
                    {
                        "recurse": "bool"
                    },
                    {
                        "recurse": "int8"
                    },
                    {
                        "recurse": "uint8"
                    },
                    {
                        "recurse": "int16"
                    },
                    {
                        "recurse": "uint16"
                    },
                    {
                        "recurse": "int32"
                    },
                    {
                        "recurse": "uint32"
                    },
                    {
                        "recurse": "int64"
                    },
                    {
                        "recurse": "uint64"
                    },
                    {
                        "recurse": "float"
                    },
                    {
                        "recurse": "dateTime"
                    },
                    {
                        "recurse": "string"
                    }
                ]
            },
            "primitive-constructor": {
                "union": {
                    "newBool": {
                        "or": [{"bool":{}}, {"recurse": "top"}]
                    },
                    "newDateTime": {
                        "or": [{"dateTime":{}}, {"recurse": "top"}]
                    },
                    "newString": {
                        "or": [{"string":{}}, {"recurse": "top"}]
                    },
                    "newFloat": {
                        "or": [{"float":{}}, {"recurse": "top"}]
                    },
                    "newInt": {
                        "or": [{"int":{}}, {"recurse": "top"}]
                    },
                    "newInt8": {
                        "or": [{"int8":{}}, {"recurse": "top"}]
                    },
                    "newInt16": {
                        "or": [{"int16":{}}, {"recurse": "top"}]
                    },
                    "newInt32": {
                        "or": [{"int32":{}}, {"recurse": "top"}]
                    },
                    "newInt64": {
                        "or": [{"int64":{}}, {"recurse": "top"}]
                    },
                    "newUint": {
                        "or": [{"uint":{}}, {"recurse": "top"}]
                    },
                    "newUint8": {
                        "or": [{"uint8":{}}, {"recurse": "top"}]
                    },
                    "newUint16": {
                        "or": [{"uint16":{}}, {"recurse": "top"}]
                    },
                    "newUint32": {
                        "or": [{"uint32":{}}, {"recurse": "top"}]
                    },
                    "newUint64": {
                        "or": [{"uint64":{}}, {"recurse": "top"}]
                    }
                }
            },
            "composite-constructor": {
                "union": {
                    "newList": {
                        "list": {
                            "recurse": "top"
                        }
                    },
                    "newTuple": {
                        "list": {
                            "recurse": "top"
                        }
                    },
                    "newMap": {
                        "map": {
                            "recurse": "top"
                        }
                    },
                    "newStruct": {
                        "map": {
                            "recurse": "top"
                        }
                    },
                    "newUnion": {
                        "struct": {
                            "case": {
                                "recurse": "top"
                            },
                            "value": {
                                "recurse": "top"
                            }
                        }
                    },
                    "newRef": {
                        "struct": {
                            "model": {
                                "or": [
                                    {
                                        "string": {}
                                    },
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            },
                            "id": {
                                "or": [
                                    {
                                        "string": {}
                                    },
                                    {
                                        "recurse": "top"
                                    }
                                ]
                            }
                        }
                    }
                }
            },
            "bool": {
                "bool": {}
            },
            "int": {
                "int": {}
            },
            "float": {
                "float": {}
            },
            "dateTime": {
                "dateTime": {}
            },
            "string": {
                "string": {}
            },
            "null": {
                "null": {}
            },
            "int8": {
                "int8": {}
            },
            "uint8": {
                "uint8": {}
            },
            "int16": {
                "int16": {}
            },
            "uint16": {
                "uint16": {}
            },
            "int32": {
                "int32": {}
            },
            "uint32": {
                "uint32": {}
            },
            "int64": {
                "int64": {}
            },
            "uint64": {
                "uint64": {}
            }
        }
    }
}
`
