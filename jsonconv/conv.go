package jsonconv

import (
	"fmt"
	"log"
	"reflect"
)

type FieldType int

const (
	FieldTypeArray FieldType = iota
	FieldTypeObject

	FieldTypeNumber
	FieldTypeString
	FieldTypeBoolean
)

func (ft FieldType) IsComplex() bool {
	return ft == FieldTypeArray || ft == FieldTypeObject
}

var types = make(map[reflect.Kind]FieldType)

func init() {
	types[reflect.Bool] = FieldTypeBoolean

	types[reflect.Int] = FieldTypeNumber
	types[reflect.Int8] = FieldTypeNumber
	types[reflect.Int16] = FieldTypeNumber
	types[reflect.Int32] = FieldTypeNumber
	types[reflect.Int64] = FieldTypeNumber
	types[reflect.Uint] = FieldTypeNumber
	types[reflect.Uint8] = FieldTypeNumber
	types[reflect.Uint16] = FieldTypeNumber
	types[reflect.Uint32] = FieldTypeNumber
	types[reflect.Uint64] = FieldTypeNumber
	types[reflect.Float32] = FieldTypeNumber
	types[reflect.Float64] = FieldTypeNumber

	types[reflect.String] = FieldTypeString

	types[reflect.Slice] = FieldTypeArray
	types[reflect.Struct] = FieldTypeObject
}

type TemplateArgs struct {
	Entities            []JSONEntity
	JSONFieldTypeString func(JSONField) string
}

type JSONEntity struct {
	Name   string
	Fields []JSONField
}

type JSONField struct {
	JsonName string
	Type     FieldType

	// Used when Type is FieldTypeArray or FieldTypeObject:
	ElementType FieldType
	// Used with FieldTypeArray when the element type is FieldTypeObject
	ElementTypeName string
}

type EntityParser struct {
	golangTypes      []reflect.Type
	jsonEntitites    []JSONEntity
	alreadyConverted map[reflect.Type]bool
}

func NewEntityParser() *EntityParser {
	return &EntityParser{
		golangTypes:      []reflect.Type{},
		jsonEntitites:    []JSONEntity{},
		alreadyConverted: map[reflect.Type]bool{},
	}
}

func (p *EntityParser) Add(obj interface{}) {
	p.AddType(reflect.TypeOf(obj))
}

func (p *EntityParser) AddType(typeOf reflect.Type) {
	p.golangTypes = append(p.golangTypes, typeOf)
}

func (p *EntityParser) Parse() ([]JSONEntity, error) {
	for _, typeOf := range p.golangTypes {
		err := p.ParseType(typeOf)
		if err != nil {
			return nil, err
		}
		p.alreadyConverted[typeOf] = true
	}
	return p.jsonEntitites, nil
}

func (p *EntityParser) ParseType(typeOf reflect.Type) error {
	if _, found := p.alreadyConverted[typeOf]; found {
		return nil
	}

	res := JSONEntity{
		Name:   typeOf.Name(),
		Fields: []JSONField{},
	}

	fields := deepFields(typeOf)
loop:
	for _, field := range fields {
		jsonFieldName := field.Tag.Get("json")
		if jsonFieldName == "-" {
			log.Println("Ignored", field.Name)
			continue loop
		}

		jsonType, found := types[field.Type.Kind()]
		if !found {
			return fmt.Errorf("Can't convert %s", field.Type.String())
		}

		if jsonType == FieldTypeArray {
			// Array
			fieldElemKind := field.Type.Elem().Kind()
			elementType, found := types[fieldElemKind]
			if !found {
				panic(fmt.Sprintf("Cannot find json element type for %s", fieldElemKind.String()))
			}
			res.Fields = append(res.Fields, JSONField{
				JsonName:        jsonFieldName,
				Type:            jsonType,
				ElementType:     elementType,
				ElementTypeName: field.Type.Elem().Name(),
			})
			if elementType.IsComplex() {
				p.ParseType(field.Type.Elem())
			}
		} else if jsonType == FieldTypeObject {
			// Object/struct
			res.Fields = append(res.Fields, JSONField{
				JsonName:        jsonFieldName,
				Type:            jsonType,
				ElementTypeName: field.Type.Name(),
			})
			if jsonType.IsComplex() {
				p.ParseType(field.Type)
			}
		} else {
			res.Fields = append(res.Fields, JSONField{
				JsonName: jsonFieldName,
				Type:     jsonType,
			})
			// Simple type
		}
	}

	p.jsonEntitites = append(p.jsonEntitites, res)
	p.alreadyConverted[typeOf] = true

	return nil
}

func deepFields(typeOf reflect.Type) []reflect.StructField {
	fields := make([]reflect.StructField, 0)

	if typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
	}

	if typeOf.Kind() != reflect.Struct {
		return fields
	}

	for i := 0; i < typeOf.NumField(); i++ {
		f := typeOf.Field(i)

		kind := f.Type.Kind()
		if f.Anonymous && kind == reflect.Struct {
			//fmt.Println(v.Interface())
			fields = append(fields, deepFields(f.Type)...)
		} else {
			fields = append(fields, f)
		}
	}

	return fields
}
