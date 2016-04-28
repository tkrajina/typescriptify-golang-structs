package main

import (
	"fmt"

	"github.com/tkrajina/typescriptify-golang-structs/example"
	"github.com/tkrajina/typescriptify-golang-structs/jsonconv"
)

func main() {
	converter := jsonconv.NewEntityParser()
	converter.Add(example.Person{})
	entities, err := converter.Parse()
	if err != nil {
		panic(err.Error())
	}
	{
		result := jsonconv.T__typescript(jsonconv.TemplateArgs{
			Entities:            entities,
			JSONFieldTypeString: GolandFieldTypeResolver,
		})
		fmt.Println(result)
	}
	{
		result := jsonconv.T__java(jsonconv.TemplateArgs{
			Entities:            entities,
			JSONFieldTypeString: JavaFieldTypeResolver,
		})
		fmt.Println(result)
	}
}

func GolandFieldTypeResolver(field jsonconv.JSONField) string {
	simpleTypes := map[jsonconv.FieldType]string{
		jsonconv.FieldTypeNumber:  "number",
		jsonconv.FieldTypeString:  "string",
		jsonconv.FieldTypeBoolean: "boolean",
	}

	if simple, found := simpleTypes[field.Type]; found {
		return simple
	}

	if field.Type == jsonconv.FieldTypeArray {
		if simple, found := simpleTypes[field.ElementType]; found {
			return fmt.Sprintf("[]%s", simple)
		} else if len(field.ElementTypeName) > 0 {
			return fmt.Sprintf("[]%s", field.ElementTypeName)
		} else {
			panic(fmt.Sprintf("No element type name for %v", field))
		}
	} else if field.Type == jsonconv.FieldTypeObject {
		return field.ElementTypeName
	}

	panic(fmt.Sprintf("Cannot find name for %v", field))
}

func JavaFieldTypeResolver(field jsonconv.JSONField) string {
	simpleTypes := map[jsonconv.FieldType]string{
		jsonconv.FieldTypeNumber:  "Double",
		jsonconv.FieldTypeString:  "String",
		jsonconv.FieldTypeBoolean: "Boolean",
	}

	if simple, found := simpleTypes[field.Type]; found {
		return simple
	}

	if field.Type == jsonconv.FieldTypeArray {
		if simple, found := simpleTypes[field.ElementType]; found {
			return fmt.Sprintf("List<%s>", simple)
		} else if len(field.ElementTypeName) > 0 {
			return fmt.Sprintf("List<%s>", field.ElementTypeName)
		} else {
			panic(fmt.Sprintf("No element type name for %v", field))
		}
	} else if field.Type == jsonconv.FieldTypeObject {
		return field.ElementTypeName
	}

	panic(fmt.Sprintf("Cannot find name for %v", field))
}
