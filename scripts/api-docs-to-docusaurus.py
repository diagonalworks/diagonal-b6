#!/usr/bin/env python

import sys
import json

if __name__ == "__main__":
  docs = json.loads(sys.stdin.read())

  print("""---
sidebar_position: 1
---
""")

  print("# b6 API documentation")

  all_interfaces = [ i["Name"] for i in docs["Interfaces"] ]

  def maybe_link (typ):
    if typ in all_interfaces:
      return f"[`{typ}`](#{typ.lower()})"
    return f"`{typ}`"

  print("## Functions")

  print("""
  This is documentationed generated from the `b6-api` binary.

  Below are all the functions, with their Python function name, and the
  corresponding argument names and types. Note that the types are the **b6 go**
  type definitions; not the python ones; nevertheless it is indicative of what
  type to expect to construct on the python side.
  """)

  for function in sorted(docs["Functions"], key=lambda f: f["Name"]):
    name = function["Name"].replace("-", "_")
    argTypes = function["ArgTypes"]
    argNames = function["ArgNames"]
    result = function["Result"]
    doc = function["Doc"]
    isVariadic = function["IsVariadic"]
    print("")
    print(f"### *b6.{name}* <span style={{{{fontSize: 12 +'px', fontWeight: 'normal'}}}}>:: {result}</span>")
    print("")
    print(doc)
    print("#### Arguments")
    print("")

    if len(argTypes) != len(argNames):
      print(f"Error! {function} doesn't have equal number of args and types.")
      exit(2)

    for (arg, typ) in zip(argNames, argTypes):
      print(f"- `{arg}` of type {maybe_link(typ)}")

    print("")
    print("#### Returns")
    print(f"{maybe_link(result)}")

    misc_items = []

    if isVariadic:
      misc_items.append(f" - [x] Function is _variadic_ (has a variable number of arguments.)")

    if len(misc_items) > 0:
      print("#### Misc")
      for i in misc_items:
        print(i)


  print("## Interfaces")

  for interface in sorted(docs["Interfaces"], key=lambda i: i["Name"]):
    name = interface["Name"]
    print("")
    print(f"### *{name}*")
    print("")

    if interface["Implements"]:
      print("#### Implements")
      for impl in sorted(interface["Implements"]):
        print(f"- [{impl}](#{impl.lower()})")
