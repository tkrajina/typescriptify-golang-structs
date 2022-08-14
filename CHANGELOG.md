# Changelog

## v0.1.8, v0.1.9

- Typescript doc tags
- Handle fields that is not annotated with json tag
- Matrix testing

## v0.1.7

- Handle packages with hyphens

## v0.1.6

- Fix map keys if suffix/prefix specified
- process customImports on Params with other flags

## v0.1.5

- Fixed panic with arrays
- Use go modules for dependency management
- Example shell script how to create a typescript model directly from json

## v0.1.4

- fix ignored pointers
- interface cmdline flag

## v0.1.2

- Log field and type creation to make the order (and why a type was converted) simpler to follow
- Global custom types: Merge branch 'fix-33' of https://github.com/shackra/typescriptify-golang-structs into shackra-fix-33

## v0.1.1

- custom types (insted of setting `ts_type` and `ts_transform` every time)

## v0.1.0

- simplified conversion of objects
- Pointer anonymous structs
- more (and better) tests
- maps of objects
- convert in constructors (createFrom deprecated)
- custom imports
- Add `?` to field name if it's a pointer type
- New way of defining enums
