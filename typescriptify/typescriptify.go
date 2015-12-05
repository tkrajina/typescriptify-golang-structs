package typescriptify

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

type TypeScriptify struct {
	prefix      string
	suffix      string
	golangTypes []reflect.Type
	types       map[reflect.Kind]string
	indent      string

	// throwaway, used when converting
	alreadyConverted map[reflect.Type]bool
}

func New() *TypeScriptify {
	result := new(TypeScriptify)
	result.indent = "\t"

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

	return result
}

func (this TypeScriptify) getTypescriptFieldLine(fieldName string, kind reflect.Kind, array bool) (string, error) {
	arr := ""
	if array {
		arr = "[]"
	}
	if typeScriptType, ok := this.types[kind]; ok {
		if len(fieldName) > 0 {
			return fmt.Sprintf("%s%s: %s%s;\n", this.indent, fieldName, typeScriptType, arr), nil
		}
	}
	return "", errors.New(fmt.Sprintf("don't know how to translate type %s (field:%s)", kind.String(), fieldName))
}

func (this *TypeScriptify) Indent(indent string) {
	this.indent = indent
}

func (this *TypeScriptify) Prefix(prefix string) {
	this.prefix = prefix
}

func (this *TypeScriptify) Suffix(suffix string) {
	this.suffix = suffix
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
		result += "\n" + strings.Trim(typeScriptCode, " "+this.indent+"\r\n")
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

func (this TypeScriptify) ConvertToFile(fileName string) error {
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
	entityName := fmt.Sprintf("%s%s%s", this.prefix, this.suffix, typeOf.Name())
	result := fmt.Sprintf("class %s {\n", entityName)

	if _, found := this.alreadyConverted[typeOf]; found {
		// Already converted
		return "", nil
	}

	for i := 0; i < typeOf.NumField(); i++ {
		val := typeOf.Field(i)
		//fmt.Println("kind=", val.Type.Kind().String())
		jsonTag := val.Tag.Get("json")
		jsonFieldName := ""
		if len(jsonTag) > 0 {
			jsonTagParts := strings.Split(jsonTag, ",")
			if len(jsonTagParts) > 0 {
				jsonFieldName = strings.Trim(jsonTagParts[0], this.indent)
			}
		}
		if len(jsonFieldName) > 0 && jsonFieldName != "-" {
			if val.Type.Kind() == reflect.Struct {
				// Struct:
				typeScriptChunk, err := this.convertType(val.Type, customCode)
				if err != nil {
					return "", err
				}
				result = typeScriptChunk + "\n" + result + fmt.Sprintf("%s%s: %s;\n", this.indent, jsonFieldName, val.Type.Name())
			} else if val.Type.Kind() == reflect.Slice {
				// Slice:
				if val.Type.Elem().Kind() == reflect.Struct {
					// Slice of structs:
					typeScriptChunk, err := this.convertType(val.Type.Elem(), customCode)
					if err != nil {
						return "", err
					}
					result = typeScriptChunk + "\n" + result + fmt.Sprintf("%s%s: %s[];\n", this.indent, jsonFieldName, val.Type.Elem().Name())
				} else {
					// Slice of simple fields:
					line, err := this.getTypescriptFieldLine(jsonFieldName, val.Type.Elem().Kind(), true)
					if err != nil {
						return "", err
					}
					result += line
				}
			} else {
				// simple field:
				line, err := this.getTypescriptFieldLine(jsonFieldName, val.Type.Kind(), false)
				if err != nil {
					return "", err
				}
				result += line
			}
		}
	}

	if customCode != nil {
		code := customCode[entityName]
		result += this.indent + "//[" + entityName + ":]\n" + code + "\n\n" + this.indent + "//[end]\n"
	}

	result += "}"

	this.alreadyConverted[typeOf] = true

	return result, nil
}
