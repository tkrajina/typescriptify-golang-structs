package typescriptify

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"
	"regexp"
)

type (
	TagFn func(nameField *string, value string) (string, string)

	TypeScriptify struct {
		Prefix           string
		Suffix           string
		Indent           string
		CreateFromMethod bool
		BackupExtension  string // If empty no backup

		tagsHandler map[string]TagFn

		golangTypes []reflect.Type
		types       map[reflect.Kind]string

		// throwaway, used when converting
		alreadyConverted map[reflect.Type]bool
	}
)

var TagRegExp = regexp.MustCompile("(?i)([a-z]+):\"(.+?)\"\\s?")

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

	result.types = types

	result.tagsHandler = map[string]TagFn{}

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

func (this *TypeScriptify) Add(obj interface{}) {
	this.AddType(reflect.TypeOf(obj))
}

func (this *TypeScriptify) AddType(typeOf reflect.Type) {
	this.golangTypes = append(this.golangTypes, typeOf)
}

func (this *TypeScriptify) AddTag(tagName string, tagFn TagFn) {
	this.tagsHandler[tagName] = tagFn
}

func (this *TypeScriptify) RemoveTag(tagName string) {
	this.tagsHandler[tagName] = nil
}

func (this *TypeScriptify) GetTagsHandler() map[string]TagFn {
	return this.tagsHandler
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

	fileOut, err := os.Create(fmt.Sprintf("%s-%s.%s", fileName, time.Now().Format("2006-01-02T15_04_05.99"), this.BackupExtension))
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

func (this *TypeScriptify) convertType(typeOf reflect.Type, customCode map[string]string) (string, error) {
	if _, found := this.alreadyConverted[typeOf]; found { // Already converted
		return "", nil
	}

	entityName := fmt.Sprintf("%s%s%s", this.Prefix, this.Suffix, typeOf.Name())
	result := fmt.Sprintf("class %s {\n", entityName)
	builder := typeScriptClassBuilder{
		types:  this.types,
		indent: this.Indent,
	}

	fields := deepFields(typeOf)

	for _, field := range fields {
		jsonTag := field.Tag.Get("json")
		jsonFieldName := ""
		if len(jsonTag) > 0 {
			jsonTagParts := strings.Split(jsonTag, ",")
			if len(jsonTagParts) > 0 {
				jsonFieldName = strings.Trim(jsonTagParts[0], this.Indent)
			}
		}
		if len(jsonFieldName) > 0 && jsonFieldName != "-" {
			var err error

			switch field.Type.Kind() {
			case reflect.Struct: // Struct:
				typeScriptChunk, err := this.convertType(field.Type, customCode)
				if err != nil {
					return "", err
				}
				result = typeScriptChunk + "\n" + result
				builder.AddStructField(jsonFieldName, field.Type.Name())
			case reflect.Slice: // Slice:
				if field.Type.Elem().Kind() == reflect.Struct { // Slice of structs:
					typeScriptChunk, err := this.convertType(field.Type.Elem(), customCode)
					if err != nil {
						return "", err
					}
					result = typeScriptChunk + "\n" + result
					builder.AddArrayOfStructsField(jsonFieldName, field.Type.Elem().Name())
				} else { // Slice of simple fields:
					err = builder.AddSimpleArrayField(jsonFieldName, field.Type.Elem().Name(), field.Type.Elem().Kind())
				}
			default: // Simple field:
				err = builder.AddSimpleField(jsonFieldName, field.Type.Name(), field.Type.Kind())
			}
			if err != nil {
				return "", err
			}

			tags := TagRegExp.FindAllStringSubmatch(string(field.Tag), -1)
			for _, tag := range tags {
				if len(tag) > 2 && this.tagsHandler[tag[1]] != nil {
					builder.AddCustomField(this.tagsHandler[tag[1]](&jsonFieldName, tag[2]))
				}
			}
		}
	}

	result += builder.fields
	if this.CreateFromMethod {
		result += fmt.Sprintf("\n%sstatic createFrom(source: any) {\n", this.Indent)
		result += fmt.Sprintf("%s%svar result = new %s();\n", this.Indent, this.Indent, entityName)
		result += builder.createFromMethodBody
		result += fmt.Sprintf("%s%sreturn result;\n", this.Indent, this.Indent)
		result += fmt.Sprintf("%s}\n\n", this.Indent)
	}

	if customCode != nil {
		code := customCode[entityName]
		result += this.Indent + "//[" + entityName + ":]\n" + code + "\n\n" + this.Indent + "//[end]\n"
	}

	result += "}"

	this.alreadyConverted[typeOf] = true

	return result, nil
}

type typeScriptClassBuilder struct {
	types                map[reflect.Kind]string
	indent               string
	fields               string
	createFromMethodBody string
}

func (this *typeScriptClassBuilder) AddSimpleArrayField(fieldName, fieldType string, kind reflect.Kind) error {
	if typeScriptType, ok := this.types[kind]; ok {
		if len(fieldName) > 0 {
			this.fields += fmt.Sprintf("%s%s: %s[];\n", this.indent, fieldName, typeScriptType)
			this.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"];\n", this.indent, this.indent, fieldName, fieldName)
			return nil
		}
	}
	return errors.New(fmt.Sprintf("Cannot find type for %s (%s/%s)", kind.String(), fieldName, fieldType))
}

func (this *typeScriptClassBuilder) AddSimpleField(fieldName, fieldType string, kind reflect.Kind) error {
	if typeScriptType, ok := this.types[kind]; ok {
		if len(fieldName) > 0 {
			this.fields += fmt.Sprintf("%s%s: %s;\n", this.indent, fieldName, typeScriptType)
			this.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"];\n", this.indent, this.indent, fieldName, fieldName)
			return nil
		}
	}
	return errors.New("Cannot find type for " + fieldType)
}

func (this *typeScriptClassBuilder) AddStructField(fieldName, fieldType string) {
	this.AddCustomField(fieldName, fieldType)
	this.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"] ? %s.createFrom(source[\"%s\"]) : null;\n", this.indent, this.indent, fieldName, fieldName, fieldType, fieldName)
}

func (this *typeScriptClassBuilder) AddCustomField(fieldName, fieldType string) {
	this.fields += fmt.Sprintf("%s%s: %s;\n", this.indent, fieldName, fieldType)
}

func (this *typeScriptClassBuilder) AddArrayOfStructsField(fieldName, fieldType string) {
	this.fields += fmt.Sprintf("%s%s: %s[];\n", this.indent, fieldName, fieldType)
	this.createFromMethodBody += fmt.Sprintf("%s%sresult.%s = source[\"%s\"] ? source[\"%s\"].map(function(element) { return %s.createFrom(element); }) : null;\n", this.indent, this.indent, fieldName, fieldName, fieldName, fieldType)
}
