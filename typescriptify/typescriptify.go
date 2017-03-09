package typescriptify

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var TimeType = reflect.TypeOf(time.Now())

type TypeScriptify struct {
	Prefix            string
	Suffix            string
	Indent            string
	CreateConstructor bool
	CreateEmptyObject bool
	UseInterface      bool
	BackupExtension   string // If empty no backup

	golangTypes []reflect.Type
	types       map[reflect.Kind]string
	structTypes map[string]reflect.Type

	// throwaway, used when converting
	alreadyConverted map[reflect.Type]bool

	AllOptional bool
}

func New() *TypeScriptify {
	result := new(TypeScriptify)
	result.Indent = "\t"
	result.BackupExtension = "backup"

	types := make(map[reflect.Kind]string)

	types[reflect.Bool] = "boolean"

	types[reflect.Int] = "number"
	types[reflect.Int8] = "number"
	types[reflect.Int16] = "number"
	types[reflect.Int32] = "number"
	types[reflect.Int64] = "number"
	types[reflect.Uint] = "number"
	types[reflect.Uint8] = "number"
	types[reflect.Uint16] = "number"
	types[reflect.Uint32] = "number"
	types[reflect.Uint64] = "number"
	types[reflect.Float32] = "number"
	types[reflect.Float64] = "number"

	types[reflect.String] = "string"

	types[reflect.Interface] = "any"

	result.types = types
	result.structTypes = make(map[string]reflect.Type)

	result.Indent = "    "
	result.CreateConstructor = true

	return result
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

func (this *TypeScriptify) Add(obj interface{}) {
	this.AddType(reflect.TypeOf(obj))
}

func (this *TypeScriptify) AddType(typeOf reflect.Type) {
	this.golangTypes = append(this.golangTypes, typeOf)
}

func (this *TypeScriptify) Convert(customCode map[string]string) (string, error) {
	this.alreadyConverted = make(map[reflect.Type]bool)

	result := ""
	for _, typeof := range this.golangTypes {
		typeScriptCode, err := this.convertType(typeof, customCode)
		if err != nil {
			return "", err
		}
		result += "\n" + strings.Trim(typeScriptCode, " "+this.Indent+"\r\n")
	}

	structItems := ""
	for tsStructTypeName, _ := range this.structTypes {
		structItems += fmt.Sprintf("\"%s\":%s,\n", tsStructTypeName, tsStructTypeName)
	}

	result += fmt.Sprintf("\nexport let AllModelTypes = {\n%s}\n", structItems)

	return result, nil
}

func loadCustomCode(fileName string) (map[string]string, error) {
	result := make(map[string]string)
	f, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return result, err
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return result, err
	}

	var currentName string
	var currentValue string
	lines := strings.Split(string(bytes), "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "//[") && strings.HasSuffix(trimmedLine, ":]") {
			currentName = strings.Replace(strings.Replace(trimmedLine, "//[", "", -1), ":]", "", -1)
			currentValue = ""
		} else if trimmedLine == "//[end]" {
			result[currentName] = strings.TrimRight(currentValue, " \t\r\n")
			currentName = ""
			currentValue = ""
		} else if len(currentName) > 0 {
			currentValue += line + "\n"
		}
	}

	return result, nil
}

func (this TypeScriptify) backup(fileName string) error {
	fileIn, err := os.Open(fileName)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// No neet to backup, just return:
		return nil
	}
	defer fileIn.Close()

	bytes, err := ioutil.ReadAll(fileIn)
	if err != nil {
		return err
	}

	fileOut, err := os.Create(fmt.Sprintf("%s-%s.%s", fileName, time.Now().Format("2006-01-02T15:04:05.99"), this.BackupExtension))
	if err != nil {
		return err
	}
	defer fileOut.Close()

	_, err = fileOut.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (this TypeScriptify) ConvertToFile(fileName string) error {
	if len(this.BackupExtension) > 0 {
		err := this.backup(fileName)
		if err != nil {
			return err
		}
	}

	customCode, err := loadCustomCode(fileName)
	if err != nil {
		return err
	}

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	converted, err := this.Convert(customCode)
	if err != nil {
		return err
	}

	f.WriteString("/* Do not change, this code is generated from Golang structs */\n\n")
	f.WriteString(converted)
	if err != nil {
		return err
	}

	return nil
}

func IsExported(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(ch)
}

func (this *TypeScriptify) convertType(typeOf reflect.Type, customCode map[string]string, customName ...string) (string, error) {
	if typeOf == TimeType {
		return "", nil
	}

	isAnonymousStruct := len(customName) > 0 && customName[0] != typeOf.Name()

	if !isAnonymousStruct {
		if _, found := this.alreadyConverted[typeOf]; found {
			// Already converted
			return "", nil
		}

		this.alreadyConverted[typeOf] = true
	}

	entityName := fmt.Sprintf("%s%s%s", this.Prefix, this.Suffix, typeOf.Name())
	if len(customName) > 0 {
		entityName = fmt.Sprintf("%s%s%s", this.Prefix, this.Suffix, customName[0])
	}

	typeKind := "class"
	if this.UseInterface {
		typeKind = "interface"
	}
	this.structTypes[entityName] = typeOf
	result := fmt.Sprintf("export %s %s {\n", typeKind, entityName)
	builder := typeScriptClassBuilder{
		types:       this.types,
		structTypes: this.structTypes,
		indent:      this.Indent,
		AllOptional: this.AllOptional,
	}

	fields := deepFields(typeOf)
	fmt.Println(typeOf.Name(), typeOf.Kind(), entityName, "fields:", fields)
	for _, field := range fields {
		if !IsExported(field.Name) {
			continue // skip unexported field
		}
		jsonTag := field.Tag.Get("json")
		jsonFieldName := ""
		fieldType := field.Type
		isPtr := fieldType.Kind() == reflect.Ptr
		if isPtr {
			fieldType = field.Type.Elem()
		}
		if len(jsonTag) > 0 {
			jsonTagParts := strings.Split(jsonTag, ",")
			if len(jsonTagParts) > 0 {
				jsonFieldName = strings.Trim(jsonTagParts[0], this.Indent)
			}
		} else {
			if field.Name != "" {
				jsonFieldName = field.Name
			} else {
				jsonFieldName = fieldType.Name()
			}
		}
		fmt.Println("jsonFieldName", jsonFieldName)

		if len(jsonFieldName) > 0 && jsonFieldName != "-" {
			var err error

			if fieldType.Kind() == reflect.Interface {
				// empty interface
				builder.AddStructField(jsonFieldName, "any")

			} else if fieldType.Kind() == reflect.Map {
				// map[string]interface{}
				fmt.Println(fieldType.Key())
				keyType := "string"
				if kt, ok := this.types[fieldType.Key().Kind()]; ok {
					keyType = kt
				}
				valType := "any"
				mapValueType := fieldType.Elem()
				if mapValueType.Kind() == reflect.Ptr {
					mapValueType = mapValueType.Elem()
				}
				if mapValueType.Kind() == reflect.Struct {
					valType = mapValueType.Name()
					typeScriptChunk, err := this.convertType(mapValueType, customCode)
					if err != nil {
						return "", err
					}
					if typeScriptChunk != "" {
						result = typeScriptChunk + "\n" + result
					}
				} else if vt, ok := this.types[mapValueType.Kind()]; ok {
					valType = vt
				}
				builder.AddStructField(jsonFieldName, "{[key: "+keyType+"]: "+valType+"}", true)

			} else if fieldType.Kind() == reflect.Struct {
				// Struct:
				fieldTypeName := fieldType.Name()
				if fieldTypeName == "" {
					fieldTypeName = "__" + entityName + "_" + jsonFieldName // inline struct declaration
				}
				typeScriptChunk, err := this.convertType(fieldType, customCode, fieldTypeName)
				if err != nil {
					return "", err
				}
				result = typeScriptChunk + "\n" + result
				if fieldType == TimeType {
					//fieldTypeName = "Date"
					fieldTypeName = "string"
				} else {
					this.structTypes[fieldTypeName] = fieldType
				}
				builder.AddStructField(jsonFieldName, fieldTypeName, isPtr)

			} else if fieldType.Kind() == reflect.Slice {
				// Slice:
				elemType := fieldType.Elem()
				if elemType.Kind() == reflect.Ptr {
					fmt.Println("Ptr type", fieldType)
					elemType = elemType.Elem()
				}

				if elemType.Kind() == reflect.Struct {
					// Slice of structs:
					elemTypeName := elemType.Name()
					if elemTypeName == "" {
						elemTypeName = "__" + entityName + "_" + jsonFieldName // inline struct declaration
					}
					typeScriptChunk, err := this.convertType(elemType, customCode, elemTypeName)
					if err != nil {
						return "", err
					}
					result = typeScriptChunk + "\n" + result
					this.structTypes[elemTypeName] = elemType
					if elemType.Name() != "" {
						builder.AddArrayOfStructsField(jsonFieldName, elemType.Name())
					} else {
						builder.AddArrayOfStructsField(jsonFieldName, elemTypeName)
					}

				} else if elemType.Kind() == reflect.Interface {
					err = builder.AddSimpleArrayField(jsonFieldName, elemType.Name(), elemType.Kind())
				} else {
					// Slice of simple fields:
					err = builder.AddSimpleArrayField(jsonFieldName, elemType.Name(), elemType.Kind())
				}

			} else {
				// Simple field:
				err = builder.AddSimpleField(jsonFieldName, fieldType.Name(), fieldType.Kind(), isPtr)
			}
			if err != nil {
				return "", err
			}
		}
	}

	result += builder.fields
	if this.CreateConstructor {
		result += fmt.Sprintf("\n%sconstructor(init?: %s) {\n", this.Indent, entityName)
		result += fmt.Sprintf("%s%sif (!init) return\n", this.Indent, this.Indent)
		// result += fmt.Sprintf("%s%svar result = new %s()\n", this.Indent, this.Indent, entityName)
		result += builder.createFromMethodBody
		// result += fmt.Sprintf("%s%sreturn result\n", this.Indent, this.Indent)
		result += fmt.Sprintf("%s}\n\n", this.Indent)
	}

	if this.CreateEmptyObject {
		result += fmt.Sprintf("\n%sstatic emptyObject(): %s {\n", this.Indent, entityName)
		result += fmt.Sprintf("%s%svar result = new %s()\n", this.Indent, this.Indent, entityName)
		result += builder.createEmptyObjectBody
		result += fmt.Sprintf("%s%sreturn result\n", this.Indent, this.Indent)
		result += fmt.Sprintf("%s}\n\n", this.Indent)
	}

	// if this.CreateConstructor {
	// 	result += fmt.Sprintf("\n%sstatic createFrom(source: any) {\n", this.Indent)
	// 	result += fmt.Sprintf("%s%svar result = new %s()\n", this.Indent, this.Indent, entityName)
	// 	result += builder.createFromMethodBody
	// 	result += fmt.Sprintf("%s%sreturn result\n", this.Indent, this.Indent)
	// 	result += fmt.Sprintf("%s}\n\n", this.Indent)
	// }

	if customCode != nil {
		code := customCode[entityName]
		if code != "" {
			result += this.Indent + "//[" + entityName + ":]\n" + code + "\n\n" + this.Indent + "//[end]\n"
		}
	}

	result += "}"

	return result, nil
}

type typeScriptClassBuilder struct {
	types                 map[reflect.Kind]string
	structTypes           map[string]reflect.Type
	indent                string
	fields                string
	createFromMethodBody  string
	createEmptyObjectBody string

	AllOptional bool
}

func (this *typeScriptClassBuilder) AddSimpleArrayField(fieldName, fieldType string, kind reflect.Kind) error {
	if typeScriptType, ok := this.types[kind]; ok {
		if len(fieldName) > 0 {
			this.fields += fmt.Sprintf("%s%s: %s[]\n", this.indent, fieldName, typeScriptType)
			// this.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"]\n", this.indent, this.indent, fieldName, fieldName)
			fieldEmptyValue := "[]"
			this.createEmptyObjectBody += fmt.Sprintf("%s%sresult.%s = %s\n", this.indent, this.indent, fieldName, fieldEmptyValue)
			this.createFromMethodBody += fmt.Sprintf("%s%sif (init.%s) this.%s = init.%s\n", this.indent, this.indent, fieldName, fieldName, fieldName)
			return nil
		}
	}
	return errors.New(fmt.Sprintf("Cannot find type for %s (%s/%s)", kind.String(), fieldName, fieldType))
}

func (this *typeScriptClassBuilder) AddSimpleField(fieldName, fieldType string, kind reflect.Kind, isPtr ...bool) error {
	optional := ""
	if this.AllOptional || len(isPtr) > 0 && isPtr[0] {
		optional = "?"
	}
	if typeScriptType, ok := this.types[kind]; ok {
		if len(fieldName) > 0 {
			this.fields += fmt.Sprintf("%s%s%s: %s\n", this.indent, fieldName, optional, typeScriptType)
			// this.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"]\n", this.indent, this.indent, fieldName, fieldName)
			fieldEmptyValue := ""
			if typeScriptType == "string" {
				fieldEmptyValue = "\"\""
			} else if typeScriptType == "number" {
				fieldEmptyValue = "0"
			} else if typeScriptType == "boolean" {
				fieldEmptyValue = "false"
			} else if typeScriptType == "any" {
				fieldEmptyValue = "null"
			}
			this.createEmptyObjectBody += fmt.Sprintf("%s%sresult.%s = %s\n", this.indent, this.indent, fieldName, fieldEmptyValue)
			this.createFromMethodBody += fmt.Sprintf("%s%sif (init.%s) this.%s = init.%s\n", this.indent, this.indent, fieldName, fieldName, fieldName)
			return nil
		}
	}
	return errors.New("Cannot find type for " + fieldType)
}

func (this *typeScriptClassBuilder) AddStructField(fieldName, fieldType string, isPtr ...bool) {
	optional := ""
	if this.AllOptional || len(isPtr) > 0 && isPtr[0] {
		optional = "?"
	}
	this.fields += fmt.Sprintf("%s%s%s: %s\n", this.indent, fieldName, optional, fieldType)
	// createCall := fieldType + ".createFrom"
	// if fieldType == "Date" || fieldType == "string" || fieldType == "any" || strings.HasPrefix(fieldType, "{") {
	// 	createCall = "" // for Date, keep the string..., because JS won't deserialize to Date object automatically...
	// }
	// this.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"] ? %s(source[\"%s\"]) : null\n", this.indent, this.indent, fieldName, fieldName, createCall, fieldName)
	fieldEmptyValue := fmt.Sprintf("%s.emptyObject()", fieldType)
	if fieldType == "Date" || fieldType == "string" || fieldType == "any" || strings.HasPrefix(fieldType, "{") {
		fieldEmptyValue = "\"\"" // for Date, keep the string..., because JS won't deserialize to Date object automatically...
	}
	this.createEmptyObjectBody += fmt.Sprintf("%s%sresult.%s = %s\n", this.indent, this.indent, fieldName, fieldEmptyValue)

	this.createFromMethodBody += fmt.Sprintf("%s%sif (init.%s) this.%s = init.%s\n", this.indent, this.indent, fieldName, fieldName, fieldName)
}

func (this *typeScriptClassBuilder) AddArrayOfStructsField(fieldName, fieldType string) {
	this.fields += fmt.Sprintf("%s%s: %s[]\n", this.indent, fieldName, fieldType)
	// createCall := fieldType + ".createFrom"
	// if fieldType == "Date" || fieldType == "string" {
	// 	createCall = ""
	// }
	// this.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"] ? source[\"%s\"].map(function(element) { return %s(element); }) : null;\n", this.indent, this.indent, fieldName, fieldName, fieldName, createCall)
	fieldEmptyValue := "[]"
	this.createEmptyObjectBody += fmt.Sprintf("%s%sresult.%s = %s\n", this.indent, this.indent, fieldName, fieldEmptyValue)

	this.createFromMethodBody += fmt.Sprintf("%s%sif (init.%s) this.%s = init.%s\n", this.indent, this.indent, fieldName, fieldName, fieldName)
}
