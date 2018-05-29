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

type TypeScriptify struct {
	Prefix             string
	Suffix             string
	Indent             string
	CreateConstructor  bool
	CreateEmptyObject  bool
	CreateAllModelType bool
	UseInterface       bool
	BackupExtension    string // If empty no backup

	golangTypes []reflect.Type
	types       map[reflect.Kind]string
	structTypes map[string]reflect.Type
	dateTypes   []reflect.Type

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

	result.dateTypes = []reflect.Type{
		reflect.TypeOf(time.Now()),
	}

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

func (t *TypeScriptify) Add(obj interface{}) {
	t.AddType(reflect.TypeOf(obj))
}

func (t *TypeScriptify) AddType(typeOf reflect.Type) {
	t.golangTypes = append(t.golangTypes, typeOf)
}

func (t *TypeScriptify) RegisterDateType(typeOf reflect.Type) {
	t.dateTypes = append(t.dateTypes, typeOf)
}

func (t *TypeScriptify) Convert(customCode map[string]string) (string, error) {
	t.alreadyConverted = make(map[reflect.Type]bool)

	result := ""
	for _, typeof := range t.golangTypes {
		typeScriptCode, err := t.convertType(typeof, customCode)
		if err != nil {
			return "", err
		}
		result += "\n" + strings.Trim(typeScriptCode, " "+t.Indent+"\r\n")
	}

	if t.CreateAllModelType {
		structItems := ""
		for tsStructTypeName := range t.structTypes {
			structItems += fmt.Sprintf("\"%s\":%s,\n", tsStructTypeName, tsStructTypeName)
		}

		result += fmt.Sprintf("\nexport let AllModelTypes = {\n%s}\n", structItems)
	}

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

func (t TypeScriptify) backup(fileName string) error {
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

	fileOut, err := os.Create(fmt.Sprintf("%s-%s.%s", fileName, time.Now().Format("2006-01-02T15_04_05.99"), t.BackupExtension))
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

func (t TypeScriptify) ConvertToFile(fileName string) error {
	if len(t.BackupExtension) > 0 {
		err := t.backup(fileName)
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

	converted, err := t.Convert(customCode)
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

func (t *TypeScriptify) convertType(typeOf reflect.Type, customCode map[string]string, customName ...string) (string, error) {
	for _, v := range t.dateTypes {
		if v == typeOf {
			return "", nil
		}
	}

	isAnonymousStruct := len(customName) > 0 && customName[0] != typeOf.Name()

	if !isAnonymousStruct {
		if _, found := t.alreadyConverted[typeOf]; found {
			// Already converted
			return "", nil
		}

		t.alreadyConverted[typeOf] = true
	}

	entityName := fmt.Sprintf("%s%s%s", t.Prefix, t.Suffix, typeOf.Name())
	if len(customName) > 0 {
		entityName = fmt.Sprintf("%s%s%s", t.Prefix, t.Suffix, customName[0])
	}

	typeKind := "class"
	if t.UseInterface {
		typeKind = "interface"
	}
	t.structTypes[entityName] = typeOf
	result := fmt.Sprintf("export %s %s {\n", typeKind, entityName)
	builder := typeScriptClassBuilder{
		types:       t.types,
		structTypes: t.structTypes,
		indent:      t.Indent,
		AllOptional: t.AllOptional,
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
				jsonFieldName = strings.Trim(jsonTagParts[0], t.Indent)
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
			isOptional, err := t.parseTag(field)

			if fieldType.Kind() == reflect.Interface {
				// empty interface
				builder.AddStructField(jsonFieldName, "any", false, isOptional)

			} else if fieldType.Kind() == reflect.Map {
				// map[string]interface{}
				fmt.Println(fieldType.Key())
				keyType := "string"
				if kt, ok := t.types[fieldType.Key().Kind()]; ok {
					keyType = kt
				}
				valType := "any"
				mapValueType := fieldType.Elem()
				if mapValueType.Kind() == reflect.Ptr {
					mapValueType = mapValueType.Elem()
				}
				if mapValueType.Kind() == reflect.Struct {
					valType = t.Prefix + mapValueType.Name() + t.Suffix
					typeScriptChunk, err := t.convertType(mapValueType, customCode)
					if err != nil {
						return "", err
					}
					if typeScriptChunk != "" {
						result = typeScriptChunk + "\n" + result
					}
				} else if vt, ok := t.types[mapValueType.Kind()]; ok {
					valType = vt
				}
				builder.AddStructField(jsonFieldName, "{[key: "+keyType+"]: "+valType+"}", true, isOptional)

			} else if fieldType.Kind() == reflect.Struct {
				// Struct:
				fieldTypeName := fieldType.Name()
				if fieldTypeName == "" {
					fieldTypeName = "__" + entityName + "_" + jsonFieldName // inline struct declaration
				}
				typeScriptChunk, err := t.convertType(fieldType, customCode, fieldTypeName)
				if err != nil {
					return "", err
				}
				result = typeScriptChunk + "\n" + result

				isDateField := false
				for _, v := range t.dateTypes {
					if v != fieldType {
						continue
					}

					isDateField = true
					fieldTypeName = "Date"
				}
				if !isDateField {
					t.structTypes[fieldTypeName] = fieldType
					fieldTypeName = t.Prefix + fieldTypeName
				}
				builder.AddStructField(jsonFieldName, fieldTypeName, isPtr, isOptional)

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
					typeScriptChunk, err := t.convertType(elemType, customCode, elemTypeName)
					if err != nil {
						return "", err
					}
					result = typeScriptChunk + "\n" + result
					t.structTypes[elemTypeName] = elemType
					if elemType.Name() != "" {
						builder.AddArrayOfStructsField(jsonFieldName, t.Prefix + elemType.Name() + t.Suffix, isPtr, isOptional)
					} else {
						builder.AddArrayOfStructsField(jsonFieldName, t.Prefix + elemTypeName + t.Suffix, isPtr, isOptional)
					}

				} else if elemType.Kind() == reflect.Interface {
					err = builder.AddSimpleArrayField(jsonFieldName, elemType.Name(), elemType.Kind(), isPtr,  isOptional)
				} else {
					// Slice of simple fields:
					err = builder.AddSimpleArrayField(jsonFieldName, elemType.Name(), elemType.Kind(), isPtr, isOptional)
				}

			} else {
				// Simple field:
				err = builder.AddSimpleField(jsonFieldName, fieldType.Name(), fieldType.Kind(), isPtr, isOptional)
			}
			if err != nil {
				return "", err
			}
		}
	}

	result += builder.fields
	if t.CreateConstructor {
		result += fmt.Sprintf("\n%sconstructor(init?: %s) {\n", t.Indent, entityName)
		result += fmt.Sprintf("%s%sif (!init) return\n", t.Indent, t.Indent)
		// result += fmt.Sprintf("%s%svar result = new %s()\n", this.Indent, this.Indent, entityName)
		result += builder.createFromMethodBody
		// result += fmt.Sprintf("%s%sreturn result\n", this.Indent, this.Indent)
		result += fmt.Sprintf("%s}\n\n", t.Indent)
	}

	if t.CreateEmptyObject {
		result += fmt.Sprintf("\n%sstatic emptyObject(): %s {\n", t.Indent, entityName)
		result += fmt.Sprintf("%s%svar result = new %s()\n", t.Indent, t.Indent, entityName)
		result += builder.createEmptyObjectBody
		result += fmt.Sprintf("%s%sreturn result\n", t.Indent, t.Indent)
		result += fmt.Sprintf("%s}\n\n", t.Indent)
	}

	if customCode != nil {
		code := customCode[entityName]
		if code != "" {
			result += t.Indent + "//[" + entityName + ":]\n" + code + "\n\n" + t.Indent + "//[end]\n"
		}
	}

	result += "}"

	return result, nil
}

func (t *TypeScriptify) parseTag(field reflect.StructField) (optional bool, err error) {
	tag := field.Tag.Get("typescriptify")
	for _, v := range strings.Split(tag, ",") {
		switch v {
		case "optional":
			optional = true
		}
	}
	return
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

func (t *typeScriptClassBuilder) AddSimpleArrayField(fieldName, fieldType string, kind reflect.Kind, isPtr, isOptional bool) error {
	optional := ""
	if t.AllOptional || isPtr || isOptional {
		optional = "?"
	}

	if typeScriptType, ok := t.types[kind]; ok {
		if len(fieldName) > 0 {
			t.fields += fmt.Sprintf("%s%s%s: %s[];\n", t.indent, fieldName, optional, typeScriptType)
			// this.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"]\n", this.indent, this.indent, fieldName, fieldName)
			fieldEmptyValue := "[]"
			t.createEmptyObjectBody += fmt.Sprintf("%s%sresult.%s = %s\n", t.indent, t.indent, fieldName, fieldEmptyValue)
			t.createFromMethodBody += fmt.Sprintf("%s%sif (init.%s) this.%s = init.%s\n", t.indent, t.indent, fieldName, fieldName, fieldName)
			return nil
		}
	}
	return errors.New(fmt.Sprintf("Cannot find type for %s (%s/%s)", kind.String(), fieldName, fieldType))
}

func (t *typeScriptClassBuilder) AddSimpleField(fieldName, fieldType string, kind reflect.Kind, isPtr bool, isOptional bool) error {
	optional := ""
	if t.AllOptional || isPtr || isOptional {
		optional = "?"
	}
	if typeScriptType, ok := t.types[kind]; ok {
		if len(fieldName) > 0 {
			t.fields += fmt.Sprintf("%s%s%s: %s;\n", t.indent, fieldName, optional, typeScriptType)
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
			t.createEmptyObjectBody += fmt.Sprintf("%s%sresult.%s = %s\n", t.indent, t.indent, fieldName, fieldEmptyValue)
			t.createFromMethodBody += fmt.Sprintf("%s%sif (init.%s) this.%s = init.%s\n", t.indent, t.indent, fieldName, fieldName, fieldName)
			return nil
		}
	}
	return errors.New("Cannot find type for " + fieldType)
}

func (t *typeScriptClassBuilder) AddStructField(fieldName, fieldType string, isPtr bool, isOptional bool) {
	optional := ""
	if t.AllOptional || isPtr || isOptional {
		optional = "?"
	}
	t.fields += fmt.Sprintf("%s%s%s: %s;\n", t.indent, fieldName, optional, fieldType)
	// createCall := fieldType + ".createFrom"
	// if fieldType == "Date" || fieldType == "string" || fieldType == "any" || strings.HasPrefix(fieldType, "{") {
	// 	createCall = "" // for Date, keep the string..., because JS won't deserialize to Date object automatically...
	// }
	// this.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"] ? %s(source[\"%s\"]) : null\n", this.indent, this.indent, fieldName, fieldName, createCall, fieldName)
	fieldEmptyValue := fmt.Sprintf("%s.emptyObject()", fieldType)
	if fieldType == "string" || fieldType == "any" {
		fieldEmptyValue = "\"\"" // for Date, keep the string..., because JS won't deserialize to Date object automatically...
	}
	if strings.HasPrefix(fieldType, "{") {
		fieldEmptyValue = "null"
	}
	if fieldType == "Date" {
		fieldEmptyValue = "null"
		t.createFromMethodBody += fmt.Sprintf("%s%sif (init.%s) this.%s = new Date(init.%s as any)\n", t.indent, t.indent, fieldName, fieldName, fieldName)
	} else {
		t.createFromMethodBody += fmt.Sprintf("%s%sif (init.%s) this.%s = init.%s\n", t.indent, t.indent, fieldName, fieldName, fieldName)
	}
	t.createEmptyObjectBody += fmt.Sprintf("%s%sresult.%s = %s\n", t.indent, t.indent, fieldName, fieldEmptyValue)

}

func (t *typeScriptClassBuilder) AddArrayOfStructsField(fieldName, fieldType string, isPtr, isOptional bool) {
	optional := ""
	if t.AllOptional || isPtr || isOptional {
		optional = "?"
	}
	t.fields += fmt.Sprintf("%s%s%s: %s[];\n", t.indent, fieldName, optional, fieldType)
	// createCall := fieldType + ".createFrom"
	// if fieldType == "Date" || fieldType == "string" {
	// 	createCall = ""
	// }
	// this.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"] ? source[\"%s\"].map(function(element) { return %s(element); }) : null;\n", this.indent, this.indent, fieldName, fieldName, fieldName, createCall)
	fieldEmptyValue := "[]"
	t.createEmptyObjectBody += fmt.Sprintf("%s%sresult.%s = %s\n", t.indent, t.indent, fieldName, fieldEmptyValue)

	t.createFromMethodBody += fmt.Sprintf("%s%sif (init.%s) this.%s = init.%s\n", t.indent, t.indent, fieldName, fieldName, fieldName)
}
