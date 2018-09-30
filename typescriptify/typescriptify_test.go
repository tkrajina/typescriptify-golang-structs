package typescriptify

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
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

type HasName struct {
	Name string `json:"name"`
}

type Person struct {
	HasName
	Nicknames []string  `json:"nicknames"`
	Addresses []Address `json:"addresses"`
	Address   *Address  `json:"address"`
	Metadata  []byte    `json:"metadata" ts_type:"{[key:string]:string}"`
	Friends   []*Person `json:"friends"`
	Dummy     Dummy     `json:"a"`
}

func TestTypescriptifyWithTypes(t *testing.T) {
	converter := New()

	converter.AddType(reflect.TypeOf(Person{}))
	converter.CreateFromMethod = false
	converter.BackupDir = ""

	desiredResult := `export class Dummy {
        something: string;
}
export class Address {
        duration: number;
        text: string;
}
export class Person {
        name: string;
        nicknames: string[];
		addresses: Address[];
		address: Address;
		metadata: {[key:string]:string};
		friends: Person[];
        a: Dummy;
}`
	testConverter(t, converter, desiredResult)
}

func TestTypescriptifyWithInstances(t *testing.T) {
	converter := New()

	converter.Add(Person{})
	converter.Add(Dummy{})
	converter.CreateFromMethod = false
	converter.DontExport = true
	converter.BackupDir = ""

	desiredResult := `class Dummy {
        something: string;
}
class Address {
        duration: number;
        text: string;
}
class Person {
        name: string;
        nicknames: string[];
		addresses: Address[];
		address: Address;
		metadata: {[key:string]:string};
		friends: Person[];
        a: Dummy;
}`
	testConverter(t, converter, desiredResult)
}

func TestTypescriptifyWithDoubleClasses(t *testing.T) {
	converter := New()

	converter.AddType(reflect.TypeOf(Person{}))
	converter.AddType(reflect.TypeOf(Person{}))
	converter.CreateFromMethod = false
	converter.BackupDir = ""

	desiredResult := `export class Dummy {
        something: string;
}
export class Address {
        duration: number;
        text: string;
}
export class Person {
        name: string;
		nicknames: string[];
		addresses: Address[];
		address: Address;
		metadata: {[key:string]:string};
		friends: Person[];
        a: Dummy;
}`
	testConverter(t, converter, desiredResult)
}

func TestWithPrefixes(t *testing.T) {
	converter := New()

	converter.Prefix = "test_"

	converter.Add(Address{})
	converter.Add(Dummy{})
	converter.CreateFromMethod = false
	converter.DontExport = true
	converter.BackupDir = ""

	desiredResult := `class test_Address {
        duration: number;
        text: string;
}
class test_Dummy {
        something: string;
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
					os.Stderr.WriteString(fmt.Sprintf("%d. line don't match: `%s` != `%s`\n", i+1, line1, line2))
					os.Stderr.WriteString(fmt.Sprintf("Expected:\n%s\n\nGot:\n%s\n", desiredResult, typeScriptCode))
					t.Fail()
				}
			}
		}
	}
}

func TestTypescriptifyCustomType(t *testing.T) {
	type TestCustomType struct {
		Map map[string]int `json:"map" ts_type:"{[key: string]: number}"`
	}

	converter := New()

	converter.AddType(reflect.TypeOf(TestCustomType{}))
	converter.CreateFromMethod = false
	converter.BackupDir = ""

	desiredResult := `export class TestCustomType {
        map: {[key: string]: number};
}`
	testConverter(t, converter, desiredResult)
}

func TestDate(t *testing.T) {
	type TestCustomType struct {
		Time time.Time `json:"time" ts_type:"Date" ts_transform:"new Date(__VALUE__)"`
	}

	converter := New()

	converter.AddType(reflect.TypeOf(TestCustomType{}))
	converter.CreateFromMethod = true
	converter.BackupDir = ""

	desiredResult := `export class TestCustomType {
    time: Date;

    static createFrom(source: any) {
        if ('string' === typeof source) source = JSON.parse(source);
        const result = new TestCustomType();
        result.time = new Date(source["time"]);
        return result;
    }

}`
	testConverter(t, converter, desiredResult)
}

func TestRecursive(t *testing.T) {
	type Test struct {
		Children []Test `json:"children"`
	}

	converter := New()

	converter.AddType(reflect.TypeOf(Test{}))
	converter.CreateFromMethod = true
	converter.BackupDir = ""

	desiredResult := `export class Test {
    children: Test[];

    static createFrom(source: any) {
        if ('string' === typeof source) source = JSON.parse(source);
        const result = new Test();
        result.children = source["children"] ? source["children"].map(function(element) { return Test.createFrom(element); }) : null;
        return result;
    }

}`
	testConverter(t, converter, desiredResult)
}

func TestAny(t *testing.T) {
	type Test struct {
		Any interface{} `json:"field"`
	}

	converter := New()

	converter.AddType(reflect.TypeOf(Test{}))
	converter.CreateFromMethod = true
	converter.BackupDir = ""

	desiredResult := `export class Test {
    field: any;

    static createFrom(source: any) {
        if ('string' === typeof source) source = JSON.parse(source);
        const result = new Test();
        result.field = source["field"];
        return result;
    }

}`
	testConverter(t, converter, desiredResult)
}
