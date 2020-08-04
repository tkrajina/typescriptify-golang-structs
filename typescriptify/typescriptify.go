package typescriptify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"time"
)

const (
	tsTransformTag = "ts_transform"
	tsType         = "ts_type"
)

type TypeScriptify struct {
	Prefix           string
	Suffix           string
	Indent           string
	CreateFromMethod bool
	BackupDir        string // If empty no backup
	DontExport       bool
	CreateInterface  bool

	golangTypes []reflect.Type
	enumTypes   []reflect.Type
	enumValues  map[reflect.Type][]interface{}
	types       map[reflect.Kind]string

	// throwaway, used when converting
	alreadyConverted map[reflect.Type]bool
}

func New() *TypeScriptify {
	result := new(TypeScriptify)
	result.Indent = "\t"
	result.BackupDir = "."

	types := make(map[reflect.Kind]string)

	types[reflect.Bool] = "boolean"
	types[reflect.Interface] = "any"

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

	result.types = types

	result.Indent = "    "
	result.CreateFromMethod = true

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

func (t *TypeScriptify) AddEnum(values interface{}) {
	if t.enumValues == nil {
		t.enumValues = map[reflect.Type][]interface{}{}
	}

	items := reflect.ValueOf(values)
	if items.Kind() != reflect.Slice {
		panic(fmt.Sprintf("Values for %T isn't a slice", values))
	}

	var ty reflect.Type
	for i := 0; i < items.Len(); i++ {
		item := items.Index(i)
		if i == 0 {
			ty = item.Type()
		}
		t.enumValues[ty] = append(t.enumValues[ty], item.Interface())
	}

	t.enumTypes = append(t.enumTypes, ty)
}

// AddEnumValues is deprecated, use `AddEnum()`
func (t *TypeScriptify) AddEnumValues(typeOf reflect.Type, values interface{}) {
	t.AddEnum(values)
}

func (t *TypeScriptify) Convert(customCode map[string]string) (string, error) {
	t.alreadyConverted = make(map[reflect.Type]bool)

	result := ""
	for _, typeof := range t.enumTypes {

		typeScriptCode, err := t.convertEnum(typeof, t.enumValues[typeof])
		if err != nil {
			return "", err
		}
		result += "\n" + strings.Trim(typeScriptCode, " "+t.Indent+"\r\n")
	}

	for _, typeof := range t.golangTypes {

		typeScriptCode, err := t.convertType(typeof, customCode)
		if err != nil {
			return "", err
		}
		result += "\n" + strings.Trim(typeScriptCode, " "+t.Indent+"\r\n")
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

	_, backupFn := path.Split(fmt.Sprintf("%s-%s.backup", fileName, time.Now().Format("2006-01-02T15_04_05.99")))
	if t.BackupDir != "" {
		backupFn = path.Join(t.BackupDir, backupFn)
	}

	return ioutil.WriteFile(backupFn, bytes, os.FileMode(0700))
}

func (t TypeScriptify) ConvertToFile(fileName string) error {
	if len(t.BackupDir) > 0 {
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

type TSNamer interface {
	TSName() string
}

func (t *TypeScriptify) convertEnum(typeOf reflect.Type, values []interface{}) (string, error) {
	if _, found := t.alreadyConverted[typeOf]; found { // Already converted
		return "", nil
	}
	t.alreadyConverted[typeOf] = true

	entityName := t.Prefix + typeOf.Name() + t.Suffix
	result := "enum " + entityName + " {\n"

	for _, val := range values {
		byts, _ := json.Marshal(val)
		name := fmt.Sprintf("VALUE_%s", string(byts))
		if withName, is := val.(TSNamer); is {
			name = withName.TSName()
		} else if s, is := val.(fmt.Stringer); is {
			name = strings.ToUpper(s.String())
		}
		result += fmt.Sprintf("%s%s = %s,\n", t.Indent, name, string(byts))
	}

	result += "}"

	if !t.DontExport {
		result = "export " + result
	}

	return result, nil
}

func (t *TypeScriptify) convertType(typeOf reflect.Type, customCode map[string]string) (string, error) {
	if _, found := t.alreadyConverted[typeOf]; found { // Already converted
		return "", nil
	}
	t.alreadyConverted[typeOf] = true

	entityName := t.Prefix + typeOf.Name() + t.Suffix
	result := ""
	if t.CreateInterface {
		result += fmt.Sprintf("interface %s {\n", entityName)
	} else {
		result += fmt.Sprintf("class %s {\n", entityName)
	}
	if !t.DontExport {
		result = "export " + result
	}
	builder := typeScriptClassBuilder{
		types:  t.types,
		indent: t.Indent,
		prefix: t.Prefix,
		suffix: t.Suffix,
	}

	fields := deepFields(typeOf)
	for _, field := range fields {
		if field.Type.Kind() == reflect.Ptr {
			field.Type = field.Type.Elem()
		}
		jsonTag := field.Tag.Get("json")
		jsonFieldName := ""
		if len(jsonTag) > 0 {
			jsonTagParts := strings.Split(jsonTag, ",")
			if len(jsonTagParts) > 0 {
				jsonFieldName = strings.Trim(jsonTagParts[0], t.Indent)
			}
			for _, t := range jsonTagParts {
				if t == "" {
					break
				}
				if t == "omitempty" {
					jsonFieldName = fmt.Sprintf("%s?", jsonFieldName)
				}
			}
		}
		if len(jsonFieldName) > 0 && jsonFieldName != "-" {
			var err error
			customTransformation := field.Tag.Get(tsTransformTag)
			customTSType := field.Tag.Get(tsType)
			if customTransformation != "" {
				err = builder.AddSimpleField(jsonFieldName, field)
			} else if _, isEnum := t.enumValues[field.Type]; isEnum {
				builder.AddEnumField(jsonFieldName, field)
			} else if customTSType != "" { // Struct:
				err = builder.AddSimpleField(jsonFieldName, field)
			} else if field.Type.Kind() == reflect.Struct { // Struct:
				typeScriptChunk, err := t.convertType(field.Type, customCode)
				if err != nil {
					return "", err
				}
				if typeScriptChunk != "" {
					result = typeScriptChunk + "\n" + result
				}
				builder.AddStructField(jsonFieldName, field)
			} else if field.Type.Kind() == reflect.Slice { // Slice:
				if field.Type.Elem().Kind() == reflect.Ptr { //extract ptr type
					field.Type = field.Type.Elem()
				}

				arrayDepth := 1
				for field.Type.Elem().Kind() == reflect.Slice { // Slice of slices:
					field.Type = field.Type.Elem()
					arrayDepth++
				}

				if field.Type.Elem().Kind() == reflect.Struct { // Slice of structs:
					typeScriptChunk, err := t.convertType(field.Type.Elem(), customCode)
					if err != nil {
						return "", err
					}
					if typeScriptChunk != "" {
						result = typeScriptChunk + "\n" + result
					}
					builder.AddArrayOfStructsField(jsonFieldName, field, arrayDepth)
				} else { // Slice of simple fields:
					err = builder.AddSimpleArrayField(jsonFieldName, field, arrayDepth)
				}
			} else { // Simple field:
				err = builder.AddSimpleField(jsonFieldName, field)
			}
			if err != nil {
				return "", err
			}
		}
	}

	result += builder.fields
	if !t.CreateInterface && t.CreateFromMethod {
		result += fmt.Sprintf("\n%sstatic createFrom(source: any) {\n", t.Indent)
		result += fmt.Sprintf("%s%sif ('string' === typeof source) source = JSON.parse(source);\n", t.Indent, t.Indent)
		result += fmt.Sprintf("%s%sconst result = new %s();\n", t.Indent, t.Indent, entityName)
		result += builder.createFromMethodBody
		result += fmt.Sprintf("%s%sreturn result;\n", t.Indent, t.Indent)
		result += fmt.Sprintf("%s}\n\n", t.Indent)
	}

	if customCode != nil {
		code := customCode[entityName]
		if len(code) != 0 {
			result += t.Indent + "//[" + entityName + ":]\n" + code + "\n\n" + t.Indent + "//[end]\n"
		}
	}

	result += "}"

	return result, nil
}

type typeScriptClassBuilder struct {
	types                map[reflect.Kind]string
	indent               string
	fields               string
	createFromMethodBody string
	prefix, suffix       string
}

func (t *typeScriptClassBuilder) AddSimpleArrayField(fieldName string, field reflect.StructField, arrayDepth int) error {
	fieldType, kind := field.Type.Elem().Name(), field.Type.Elem().Kind()
	typeScriptType := t.types[kind]

	if len(fieldName) > 0 {
		strippedFieldName := strings.ReplaceAll(fieldName, "?", "")
		customTSType := field.Tag.Get(tsType)
		if len(customTSType) > 0 {
			t.fields += fmt.Sprintf("%s%s: %s;\n", t.indent, fieldName, customTSType)
			t.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"];\n", t.indent, t.indent, strippedFieldName, strippedFieldName)
			return nil
		} else if len(typeScriptType) > 0 {
			t.fields += fmt.Sprintf("%s%s: %s%s;\n", t.indent, fieldName, typeScriptType, strings.Repeat("[]", arrayDepth))
			t.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"];\n", t.indent, t.indent, strippedFieldName, strippedFieldName)
			return nil
		}
	}

	return errors.New(fmt.Sprintf("cannot find type for %s (%s/%s)", kind.String(), fieldName, fieldType))
}

func (t *typeScriptClassBuilder) AddSimpleField(fieldName string, field reflect.StructField) error {
	fieldType, kind := field.Type.Name(), field.Type.Kind()
	customTSType := field.Tag.Get(tsType)

	typeScriptType := t.types[kind]
	if len(customTSType) > 0 {
		typeScriptType = customTSType
	}

	customTransformation := field.Tag.Get(tsTransformTag)

	if len(typeScriptType) > 0 && len(fieldName) > 0 {
		strippedFieldName := strings.ReplaceAll(fieldName, "?", "")
		t.fields += fmt.Sprintf("%s%s: %s;\n", t.indent, fieldName, typeScriptType)
		if customTransformation == "" {
			t.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"];\n", t.indent, t.indent, strippedFieldName, strippedFieldName)
		} else {
			val := fmt.Sprintf(`source["%s"]`, strippedFieldName)
			expression := strings.Replace(customTransformation, "__VALUE__", val, -1)
			t.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = %s;\n", t.indent, t.indent, strippedFieldName, expression)
		}
		return nil
	}

	return errors.New("Cannot find type for " + fieldType + ", fideld: " + fieldName)
}

func (t *typeScriptClassBuilder) AddEnumField(fieldName string, field reflect.StructField) {
	fieldType := field.Type.Name()
	t.fields += fmt.Sprintf("%s%s: %s;\n", t.indent, fieldName, t.prefix+fieldType+t.suffix)
}

func (t *typeScriptClassBuilder) AddStructField(fieldName string, field reflect.StructField) {
	fieldType := field.Type.Name()
	strippedFieldName := strings.ReplaceAll(fieldName, "?", "")
	t.fields += fmt.Sprintf("%s%s: %s;\n", t.indent, fieldName, t.prefix+fieldType+t.suffix)
	t.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"] ? %s.createFrom(source[\"%s\"]) : null;\n", t.indent, t.indent, strippedFieldName, strippedFieldName, t.prefix+fieldType+t.suffix, strippedFieldName)
}

func (t *typeScriptClassBuilder) AddArrayOfStructsField(fieldName string, field reflect.StructField, arrayDepth int) {
	fieldType := field.Type.Elem().Name()
	strippedFieldName := strings.ReplaceAll(fieldName, "?", "")
	t.fields += fmt.Sprintf("%s%s: %s%s;\n", t.indent, fieldName, t.prefix+fieldType+t.suffix, strings.Repeat("[]", arrayDepth))
	t.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"] ? source[\"%s\"].map(function(element: any) { return %s.createFrom(element); }) : null;\n", t.indent, t.indent, strippedFieldName, strippedFieldName, strippedFieldName, t.prefix+fieldType+t.suffix)
}
