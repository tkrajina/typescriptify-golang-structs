package typescriptify

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

type Address struct {
	// Used in html
	Duration float64 `json:"duration"`
	Text1    string  `json:"text,omitempty"`
	// Ignored:
	Text2 string `json:",omitempty"`
	Text3 string `json:"-"`
}

type Dummy struct {
	Something string `json:"something"`
}

type Person struct {
	Name      string    `json:"name"`
	Nicknames []string  `json:"nicknames"`
	Addresses []Address `json:"addresses"`
	Dummy     Dummy     `json:"a"`
}

func TestTypescriptifyWithTypes(t *testing.T) {
	converter := New()

	converter.AddType(reflect.TypeOf(Person{}))

	desiredResult := `class Dummy {
        something : string;
}
class Address {
        duration : number;
        text : string;
}
class Person {
        name : string;
        nicknames : string[];
        addresses : Address[];
        a : Dummy;
}`
	testConverter(t, converter, desiredResult)
}

func TestTypescriptifyWithInstances(t *testing.T) {
	converter := New()

	converter.Add(Person{})
	converter.Add(Dummy{})

	desiredResult := `class Dummy {
        something : string;
}
class Address {
        duration : number;
        text : string;
}
class Person {
        name : string;
        nicknames : string[];
        addresses : Address[];
        a : Dummy;
}`
	testConverter(t, converter, desiredResult)
}

func TestTypescriptifyWithDoubleClasses(t *testing.T) {
	converter := New()

	converter.AddType(reflect.TypeOf(Person{}))
	converter.AddType(reflect.TypeOf(Person{}))

	desiredResult := `class Dummy {
        something : string;
}
class Address {
        duration : number;
        text : string;
}
class Person {
        name : string;
        nicknames : string[];
        addresses : Address[];
        a : Dummy;
}`
	testConverter(t, converter, desiredResult)
}

func TestWithPrefixes(t *testing.T) {
	converter := New()

	converter.Prefix("test_")

	converter.Add(Address{})
	converter.Add(Dummy{})

	desiredResult := `class test_Address {
        duration : number;
        text : string;
}
class test_Dummy {
        something : string;
}`
	testConverter(t, converter, desiredResult)
}

func testConverter(t *testing.T, converter *TypeScriptify, desiredResult string) {
	typeScriptCode, err := converter.Convert(nil)
	if err != nil {
		panic(err.Error())
	}

	typeScriptCode = strings.Trim(typeScriptCode, " \t\n\r")
	if typeScriptCode != desiredResult {
		lines1 := strings.Split(typeScriptCode, "\n")
		lines2 := strings.Split(desiredResult, "\n")

		if len(lines1) != len(lines2) {
			os.Stderr.WriteString(fmt.Sprintf("Lines: %d != %d\n", len(lines1), len(lines2)))
			os.Stderr.WriteString(fmt.Sprintf("Expected:\n%s\n\nGot:\n%s\n", desiredResult, typeScriptCode))
			t.Fail()
		} else {
			for i := 0; i < len(lines1); i++ {
				line1 := strings.Trim(lines1[i], " \t\r\n")
				line2 := strings.Trim(lines2[i], " \t\r\n")
				if line1 != line2 {
					os.Stderr.WriteString(fmt.Sprintf("%d. line don't match: `%s` != `%s`\n", line1, line2))
					os.Stderr.WriteString(fmt.Sprintf("Expected:\n%s\n\nGot:\n%s\n", desiredResult, typeScriptCode))
					t.Fail()
				}
			}
		}
	}
}
