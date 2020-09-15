package typescriptify

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/tkrajina/go-reflector/reflector"
)

const (
	tsTransformTag = "ts_transform"
	tsType         = "ts_type"
)

type enumElement struct {
	value interface{}
	name  string
}

type TypeScriptify struct {
	Prefix            string
	Suffix            string
	Indent            string
	CreateFromMethod  bool
	CreateConstructor bool
	BackupDir         string // If empty no backup
	DontExport        bool
	CreateInterface   bool

	golangTypes []reflect.Type
	enumTypes   []reflect.Type
	enums       map[reflect.Type][]enumElement
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

func (t *TypeScriptify) WithCreateFromMethod(b bool) *TypeScriptify {
	t.CreateFromMethod = b
	return t
}

func (t *TypeScriptify) WithConstructor(b bool) *TypeScriptify {
	t.CreateConstructor = b
	return t
}

func (t *TypeScriptify) WithIndent(i string) *TypeScriptify {
	t.Indent = i
	return t
}

func (t *TypeScriptify) WithBackupDir(b string) *TypeScriptify {
	t.BackupDir = b
	return t
}

func (t *TypeScriptify) WithPrefix(p string) *TypeScriptify {
	t.Prefix = p
	return t
}

func (t *TypeScriptify) WithSuffix(s string) *TypeScriptify {
	t.Suffix = s
	return t
}

func (t *TypeScriptify) Add(obj interface{}) *TypeScriptify {
	t.AddType(reflect.TypeOf(obj))
	return t
}

func (t *TypeScriptify) AddType(typeOf reflect.Type) *TypeScriptify {
	t.golangTypes = append(t.golangTypes, typeOf)
	return t
}

func (t *TypeScriptify) AddEnum(values interface{}) *TypeScriptify {
	if t.enums == nil {
		t.enums = map[reflect.Type][]enumElement{}
	}
	items := reflect.ValueOf(values)
	if items.Kind() != reflect.Slice {
		panic(fmt.Sprintf("Values for %T isn't a slice", values))
	}

	var elements []enumElement
	for i := 0; i < items.Len(); i++ {
		item := items.Index(i)

		var el enumElement
		if item.Kind() == reflect.Struct {
			r := reflector.New(item.Interface())
			val, err := r.Field("Value").Get()
			if err != nil {
				panic(fmt.Sprint("missing Type field in ", item.Type().String()))
			}
			name, err := r.Field("TSName").Get()
			if err != nil {
				panic(fmt.Sprint("missing TSName field in ", item.Type().String()))
			}
			el.value = val
			el.name = name.(string)
		} else {
			el.value = item.Interface()
			if tsNamer, is := item.Interface().(TSNamer); is {
				el.name = tsNamer.TSName()
			} else {
				panic(fmt.Sprint(item.Type().String(), " has no TSName method"))
			}
		}

		elements = append(elements, el)
	}
	ty := reflect.TypeOf(elements[0].value)
	t.enums[ty] = elements
	t.enumTypes = append(t.enumTypes, ty)

	return t
}

// AddEnumValues is deprecated, use `AddEnum()`
func (t *TypeScriptify) AddEnumValues(typeOf reflect.Type, values interface{}) *TypeScriptify {
	t.AddEnum(values)
	return t
}

func (t *TypeScriptify) Convert(customCode map[string]string) (string, error) {
	t.alreadyConverted = make(map[reflect.Type]bool)

	result := ""
	for _, typeof := range t.enumTypes {
		elements := t.enums[typeof]
		typeScriptCode, err := t.convertEnum(typeof, elements)
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

func (t *TypeScriptify) convertEnum(typeOf reflect.Type, elements []enumElement) (string, error) {
	if _, found := t.alreadyConverted[typeOf]; found { // Already converted
		return "", nil
	}
	t.alreadyConverted[typeOf] = true

	entityName := t.Prefix + typeOf.Name() + t.Suffix
	result := "enum " + entityName + " {\n"

	for _, val := range elements {
		result += fmt.Sprintf("%s%s = %#v,\n", t.Indent, val.name, val.value)
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
			} else if _, isEnum := t.enums[field.Type]; isEnum {
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

	if t.CreateFromMethod {
		fmt.Fprintln(os.Stderr, "CREATEFROM METHOD IS DEPRECATED AND WILL BE REMOVED!!!!!!")
		t.CreateConstructor = true
	}

	result += strings.Join(builder.fields, "\n") + "\n"
	if !t.CreateInterface {
		if t.CreateFromMethod {
			result += fmt.Sprintf("\n%sstatic createFrom(source: any = {}) {\n", t.Indent)
			result += fmt.Sprintf("%s%sreturn new %s(source);\n", t.Indent, t.Indent, entityName)
			result += fmt.Sprintf("%s}\n", t.Indent)
		}
		if t.CreateConstructor {
			result += fmt.Sprintf("\n%sconstructor(source: any = {}) {\n", t.Indent)
			result += t.Indent + t.Indent + "if ('string' === typeof source) source = JSON.parse(source);\n"
			result += strings.Join(builder.constructorBody, "\n") + "\n"
			result += fmt.Sprintf("%s}\n", t.Indent)
		}
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
	fields               []string
	createFromMethodBody []string
	constructorBody      []string
	prefix, suffix       string
}

func (t *typeScriptClassBuilder) AddSimpleArrayField(fieldName string, field reflect.StructField, arrayDepth int) error {
	fieldType, kind := field.Type.Elem().Name(), field.Type.Elem().Kind()
	typeScriptType := t.types[kind]

	if len(fieldName) > 0 {
		strippedFieldName := strings.ReplaceAll(fieldName, "?", "")
		customTSType := field.Tag.Get(tsType)
		if len(customTSType) > 0 {
			t.addField(fieldName, customTSType)
			t.addInitializerFieldLine(strippedFieldName, fmt.Sprintf("source[\"%s\"]", strippedFieldName))
			return nil
		} else if len(typeScriptType) > 0 {
			t.addField(fieldName, fmt.Sprint(typeScriptType, strings.Repeat("[]", arrayDepth)))
			t.addInitializerFieldLine(strippedFieldName, fmt.Sprintf("source[\"%s\"]", strippedFieldName))
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
		t.addField(fieldName, typeScriptType)
		if customTransformation == "" {
			t.addInitializerFieldLine(strippedFieldName, fmt.Sprintf("source[\"%s\"]", strippedFieldName))
		} else {
			val := fmt.Sprintf(`source["%s"]`, strippedFieldName)
			expression := strings.Replace(customTransformation, "__VALUE__", val, -1)
			t.addInitializerFieldLine(strippedFieldName, expression)
		}
		return nil
	}

	return errors.New("Cannot find type for " + fieldType + ", fideld: " + fieldName)
}

func (t *typeScriptClassBuilder) AddEnumField(fieldName string, field reflect.StructField) {
	fieldType := field.Type.Name()
	t.addField(fieldName, t.prefix+fieldType+t.suffix)
	strippedFieldName := strings.ReplaceAll(fieldName, "?", "")
	t.addInitializerFieldLine(strippedFieldName, fmt.Sprintf("source[\"%s\"]", strippedFieldName))
}

func (t *typeScriptClassBuilder) AddStructField(fieldName string, field reflect.StructField) {
	fieldType := field.Type.Name()
	strippedFieldName := strings.ReplaceAll(fieldName, "?", "")
	t.addField(fieldName, t.prefix+fieldType+t.suffix)
	t.addInitializerFieldLine(strippedFieldName, fmt.Sprintf("source[\"%s\"] && new %s(source[\"%s\"])", strippedFieldName, t.prefix+fieldType+t.suffix, strippedFieldName))
}

func (t *typeScriptClassBuilder) AddArrayOfStructsField(fieldName string, field reflect.StructField, arrayDepth int) {
	fieldType := field.Type.Elem().Name()
	strippedFieldName := strings.ReplaceAll(fieldName, "?", "")
	t.addField(fieldName, fmt.Sprint(t.prefix+fieldType+t.suffix, strings.Repeat("[]", arrayDepth)))
	t.addInitializerFieldLine(strippedFieldName, fmt.Sprintf("source[\"%s\"] && source[\"%s\"].map((element: any) => new %s(element))", strippedFieldName, strippedFieldName, t.prefix+fieldType+t.suffix))
}

func (t *typeScriptClassBuilder) addInitializerFieldLine(fld, initializer string) {
	t.createFromMethodBody = append(t.createFromMethodBody, fmt.Sprint(t.indent, t.indent, "result.", fld, " = ", initializer, ";"))
	t.constructorBody = append(t.constructorBody, fmt.Sprint(t.indent, t.indent, "this.", fld, " = ", initializer, ";"))
}

func (t *typeScriptClassBuilder) addField(fld, fldType string) {
	t.fields = append(t.fields, fmt.Sprint(t.indent, fld, ": ", fldType, ";"))
}
