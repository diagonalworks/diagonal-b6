#!/usr/bin/env python3

import sys
import json

from copy import copy
from datetime import datetime

SPECIAL_FUNCTIONS = ("map", "filter")

COLLECTION_PARENTS = {
    "FeatureIDPointFeatureCollection": "AnyPointCollection",
    "FeatureIDPathFeatureCollection": "FeatureIDPathCollection",
    "FeatureIDAreaFeatureCollection": "AnyAreaCollection",
}

GENERIC_COLLECTION_TRAITS = "AnyAnyCollectionTraits"
GENERIC_COLLECTION_RESULT = "AnyAnyCollectionResult"
GENERIC_COLLECTION_COLLECTION_RESULT = "AnyAnyAnyCollectionCollectionResult"

EXTRA_TRAITS = {
    "QueryResult": ["QueryConversionTraits"],
}

def name_for_function(name):
    if name in ["or", "and"]:
        return name + "_"
    return name.replace("-", "_")

def name_for_type(t):
    if t == "string":
        return "String"
    elif t == "int":
        return "Int"
    elif t == "float64":
        return "Float"
    elif t == "any":
        return "Any"
    elif t == "bool":
        return "Bool"
    return t

def name_for_traits(t):
    return "%sTraits" % name_for_type(t)

def name_for_collection_values(t):
    return "%sValues" % name_for_type(t)

def name_for_result(t):
    return "%sResult" % name_for_type(t)

def name_for_any_key_result(t):
    return "Any%sResult" % name_for_type(t)

def name_for_collection_of_traits(t, collections):
    for (name, (key, value)) in collections.items():
        if key == "Any" and value == t:
            return name_for_traits(name)
    return GENERIC_COLLECTION_TRAITS

def name_for_collection_of_result(t, collections):
    for (name, (key, value)) in collections.items():
        if key == "Any" and value == t:
            return name_for_result(name)
    if t.endswith("Collection"):
        return GENERIC_COLLECTION_COLLECTION_RESULT
    return GENERIC_COLLECTION_RESULT

BUILTIN_RESULTS = {
    str: name_for_result("string"),
    int: name_for_result("int"),
    float: name_for_result("float64"),
    bool: name_for_result("bool"),
    list: name_for_result("AnyAnyCollection"),
}

def output_traits(t, functions, collections, hints, parents):
    if len(parents.get(t, [])) > 0:
        print("class %s(%s):" % (name_for_traits(t), ", ".join([name_for_traits(p) for p in parents[t]])))
    else:
        print("class %s:" % name_for_traits(t))
    methods = 0
    for f in functions:            
        if len(f["Args"]) > 0 and f["Args"][0] == t:
            if methods == 0:
                print("")
            signature = ", ".join(["self"] + ["a%d: %s" % (i, hints[a]) for (i, a) in enumerate(f["Args"][1:])])
            print("    def %s(%s) -> %s:" % (name_for_function(f["Name"]), signature, hints[f["Result"]]))
            args = ", ".join(["self"] + ["a%d" % i for i in range(0, len(f["Args"][1:]))])
            print("        return %s(%s)" % (name_for_function(f["Name"]), args))
            print("")
            methods += 1
    print("    @classmethod")
    print("    def _collection(cls):")
    print("        return %s" % (name_for_collection_of_result(t, collections),))
    print("")

def output_collection_values_traits(t, functions, collections, hints, parents):
    n = name_for_collection_values(t)
    if len(parents.get(t, [])) > 0:
        print("class %s(%s):" % (name_for_traits(n), ", ".join([name_for_traits(name_for_collection_values(p)) for p in parents[t]])))
    else:
        print("class %s:" % name_for_traits(n))
    methods = 0
    for f in functions:            
        if len(f["Args"]) > 0 and f["Args"][0] == t:
            if methods == 0:
                print("")
            signature = ", ".join(["self"] + ["a%d: %s" % (i, hints[a]) for (i, a) in enumerate(f["Args"][1:])])
            print("    def %s(%s) -> %s:" % (name_for_function(f["Name"]), signature, name_for_collection_of_traits(f["Result"], collections)))
            args = ", ".join(["a%d" % i for i in range(0, len(f["Args"][1:]))])
            if len(args) > 0:
                print("        return self.map(Lambda(lambda x: %s(x, %s), [self._values()]))" % (name_for_function(f["Name"]), args))
            else:
                print("        return self.map(Lambda(%s, [self._values()]))" % name_for_function(f["Name"]))
            print("")
            methods += 1
    if methods == 0:
        print("    pass")
        print("")
    
    print("class %s(Result, %s, %s):" % (name_for_any_key_result(n), name_for_traits(n), GENERIC_COLLECTION_TRAITS))
    print("")
    print("    def __init__(self, node):")
    print("        Result.__init__(self, node)")
    print("")
    print("    @classmethod")
    print("    def _values(cls):")
    print("        return %s" % (name_for_result(t)))
    print("")

def output_function_arg_result(t, hints):
    print("class %s(Result, %s):" % (name_for_result(t["Name"]), hints[t["Name"]],))
    print("")
    print("    def __init__(self, node):")
    print("        Result.__init__(self, node)")
    print("")
    args = ", ".join(["a%d : %s" % (i, hints[at]) for (i, at) in enumerate(t["Args"])])
    print("    def __call__(self, %s) -> %s:" % (args, hints[t["Result"]]))
    print("        raise NotImplementedError()")
    print("")

def ancestors(t, parents):
    queue = copy(parents.get(t, []))
    ancestors = []
    while len(queue) > 0:        
        ancestors.append(queue.pop())
        queue.extend(parents.get(ancestors[-1], []))
    return ancestors

def main():
    api = json.load(sys.stdin)
    print("# Code generated by generate_api.py. DO NOT EDIT.")
    print("# Client library for Diagonal's geospatial analysis engine, b6.")
    print("")
    print("from __future__ import annotations")
    print("")
    print("from typing import Callable")
    print("")
    print("import diagonal_b6.expression")
    print("from diagonal_b6.expression import Call, Symbol, Lambda, Result, QueryConversionTraits, register_builtin_result")
    print("")
    print("VERSION = %s" % repr(api["Version"]))
    print("")

    traits = set()
    parents = {}
    hints = {}
    for t in ("any", "int", "float64", "bool", "string"):
        traits.add(t)
        parents[t] = []
        hints[t] = name_for_traits(t)
    for t in ("int", "float64"):
        parents[t].append("Number")
    for t in api["Interfaces"]:
        traits.add(t["Name"])
        hints[t["Name"]] = name_for_traits(t["Name"])
        parents[t["Name"]] = t["Implements"]
        for tt in t["Implements"]:
            traits.add(tt)
            hints[tt] = name_for_traits(tt)

    reference_counts = {}
    collections = {}
    collection_values = set()
    for t in api["Collections"]:
        traits.add(t["Name"])
        traits.add(t["Value"])
        collections[t["Name"]] = (t["Key"], t["Value"])
        collection_values.add(t["Value"])
        for n in (t["Value"], "AnyAnyCollection"):
            reference_counts[n] = reference_counts.get(n, 0) + 1
        for a in ancestors(t["Value"], parents):
            collection_values.add(a)
            reference_counts[a] = reference_counts.get(a, 0) + 1
        hints[t["Name"]] = name_for_result(t["Name"])
        if t["Name"] != "AnyAnyCollection":
            if t["Name"] in COLLECTION_PARENTS:
                parents[t["Name"]] = [COLLECTION_PARENTS[t["Name"]], name_for_collection_values(t["Value"])]
            else:
                parents[t["Name"]] = ["AnyAnyCollection", name_for_collection_values(t["Value"])]
    for t in api["FunctionArgs"]:
        hints[t["Name"]] = "Callable[[%s],%s]" % (",".join([hints[a] for a in t["Args"]]), hints[t["Result"]])

    for t in traits:
        for a in ancestors(t, parents):
            reference_counts[a] = reference_counts.get(a, 0) + 1
    traits = list(traits)
    traits.sort(key=lambda t: -reference_counts.get(t, 0))

    for t in traits:
        output_traits(t, api["Functions"], collections, hints, parents)
        if t in collection_values:
            output_collection_values_traits(t, api["Functions"], collections, hints, parents)

    for t in traits:
        parents = ["Result", name_for_traits(t)]
        parents.extend(EXTRA_TRAITS.get(name_for_result(t), []))
        print("class %s(%s):" % (name_for_result(t), ",".join(parents)))
        print("    def __init__(self, node):")
        print("        Result.__init__(self, node)")        
        print("")
        if t in collections:
            _, values = collections[t]
            print("    @classmethod")
            print("    def _values(cls):")
            print("        return %s" % name_for_result(values))
            print("")

    for t in api["FunctionArgs"]:
        output_function_arg_result(t, hints)

    for f in api["Functions"]:
        if f["Name"] in SPECIAL_FUNCTIONS:
            print("%s = diagonal_b6.expression._%s" % (f["Name"], f["Name"]))
        else:
            signature_args = ["a%d: %s" % (i, hints[a]) for (i, a) in enumerate(f["Args"])]
            if f["IsVariadic"]:
                signature_args[-1] = "*" + signature_args[-1]
            signature = ", ".join(signature_args)
            print("def %s(%s) -> %s:" % (name_for_function(f["Name"]), signature, hints[f["Result"]]))
            n = len(f["Args"])
            if f["IsVariadic"]:
                n -= 1
            print("    args = [%s]" % ", ".join(["a%d" % i for i in range(0, n)]))
            if f["IsVariadic"]:
                print("    args.extend(a%d)" % (len(f["Args"]) - 1))
            print("    return %s(Call(Symbol(%s), args))" % (name_for_result(f["Result"]), repr(f["Name"])))
        print("")

    for type, result in BUILTIN_RESULTS.items():
        print("register_builtin_result(%s,%s)" % (type.__name__, result))

if __name__ == "__main__":
    main()