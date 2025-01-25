#!/usr/bin/env python

# TODO:
#
#   - [ ] Be a bit more precise about the Python types; for this some
#   inspection/understanding of "generate_api.py" would be required.
#
import sys
import json
import itertools

if __name__ == "__main__":
  docs = json.loads(sys.stdin.read())

  print("# API documentation")

  all_interfaces = [ i["Name"] for i in docs["Interfaces"] ]
  all_collections = [ i["Name"] for i in docs["Collections"] ]

  def maybe_link (typ):
    if typ in all_interfaces:
      return f"[{typ}](#{typ.lower()})"
    if typ in all_collections:
      return f"[{typ}](#{typ.lower()})"
    return f"`{typ}`"

  print("## Functions")

  print("""
  This is documentation generated from the `b6-api` binary written assuming
  you are interacting with it via the Python API.

  Below are all the functions, with their Python function name, and the
  corresponding argument names and types. Note that the types are the **b6 go**
  type definitions; not the python ones; nevertheless it is indicative of what
  type to expect to construct on the Python side.
  """)

  for function in sorted(docs["Functions"], key=lambda f: f["Name"]):
    name = function["Name"].replace("-", "_")
    argTypes = function["ArgTypes"]
    argNames = function["ArgNames"]
    result = function["Result"]
    doc = function["Doc"]
    isVariadic = function["IsVariadic"]
    print("")
    print(f"### <tt>{name}</tt> ")

    # Only show the arg for now; the type can come at a later point when we
    # can format it correctly, and, moreover, when we can be more precise
    # about the Python type.
    args = ", ".join([ f"{arg}" for (arg, _) in zip(argNames, argTypes) ])

    func_def = f"def {name}({args}) -> {result}"

    print("```python title='Indicative Python type signature'")
    print(f"{func_def}")
    print("```")

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
    print(f"- {maybe_link(result)}")

    misc_items = []

    if isVariadic:
      misc_items.append(" - [x] Function is _variadic_ (has a variable number of arguments.)")

    if len(misc_items) > 0:
      print("#### Misc")
      for i in misc_items:
        print(i)

  print("")
  print("## Functions by Return Type")
  sorted_by_type = sorted(docs["Functions"], key=lambda f: f["Result"])
  for group, funcs in itertools.groupby(sorted_by_type, key=lambda f: f["Result"]):
    print(f"### <tt>{group}</tt>")
    for function in sorted(funcs, key=lambda f: f["Name"]):
      name = function["Name"].replace("-", "_")
      print(f" - <tt>[{name}](#{name})</tt>")
    print("")

  print("## Collections")
  for collection in sorted(docs["Collections"], key=lambda i: i["Name"]):
    name = collection["Name"]
    key = collection["Key"]
    value = collection["Value"]
    print("")
    print(f"### <tt>{name}</tt>")
    print("")
    print("|Key|Value|")
    print("|---|-----|")
    print(f"{maybe_link(key)}|{maybe_link(value)}")


  print("## Interfaces")

  for interface in sorted(docs["Interfaces"], key=lambda i: i["Name"]):
    name = interface["Name"]
    print("")
    print(f"### <tt>{name}</tt>")
    print("")

    if interface["Implements"]:
      print("#### Implements")
      for impl in sorted(interface["Implements"]):
        print(f"- [{impl}](#{impl.lower()})")
