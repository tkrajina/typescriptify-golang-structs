package typescriptify

import (
	"encoding/json"
	"fmt"
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
        text?: string;
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
        text?: string;
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

func TestTypescriptifyWithInterfaces(t *testing.T) {
	converter := New()

	converter.Add(Person{})
	converter.Add(Dummy{})
	converter.CreateFromMethod = false
	converter.DontExport = true
	converter.BackupDir = ""
	converter.CreateInterface = true

	desiredResult := `interface Dummy {
        something: string;
}
interface Address {
        duration: number;
        text?: string;
}
interface Person {
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
        text?: string;
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
	converter.Suffix = "_test"

	converter.Add(Person{})
	converter.CreateFromMethod = false
	converter.DontExport = true
	converter.BackupDir = ""
	converter.CreateFromMethod = true

	desiredResult := `class test_Dummy_test {
    something: string;

    static createFrom(source: any) {
        if ('string' === typeof source) source = JSON.parse(source);
        const result = new test_Dummy_test();
        result.something = source["something"];
        return result;
    }

}
class test_Address_test {
    duration: number;
    text?: string;

    static createFrom(source: any) {
        if ('string' === typeof source) source = JSON.parse(source);
        const result = new test_Address_test();
        result.duration = source["duration"];
        result.text = source["text"];
        return result;
    }

}
class test_Person_test {
    name: string;
    nicknames: string[];
    addresses: test_Address_test[];
    address: test_Address_test;
    metadata: {[key:string]:string};
    friends: test_Person_test[];
    a: test_Dummy_test;

    static createFrom(source: any) {
        if ('string' === typeof source) source = JSON.parse(source);
        const result = new test_Person_test();
        result.name = source["name"];
        result.nicknames = source["nicknames"];
        result.addresses = source["addresses"] ? source["addresses"].map(function(element: any) { return test_Address_test.createFrom(element); }) : null;
        result.address = source["address"] ? test_Address_test.createFrom(source["address"]) : null;
        result.metadata = source["metadata"];
        result.friends = source["friends"] ? source["friends"].map(function(element: any) { return test_Person_test.createFrom(element); }) : null;
        result.a = source["a"] ? test_Dummy_test.createFrom(source["a"]) : null;
        return result;
    }

}`
	testConverter(t, converter, desiredResult)
}

func testConverter(t *testing.T, converter *TypeScriptify, desiredResult string) {
	typeScriptCode, err := converter.Convert(nil)
	if err != nil {
		panic(err.Error())
	}

	desiredResult = strings.TrimSpace(desiredResult)
	typeScriptCode = strings.Trim(typeScriptCode, " \t\n\r")
	if typeScriptCode != desiredResult {
		gotLines1 := strings.Split(typeScriptCode, "\n")
		expectedLines2 := strings.Split(desiredResult, "\n")

		max := len(gotLines1)
		if len(expectedLines2) > max {
			max = len(expectedLines2)
		}

		for i := 0; i < max; i++ {
			var gotLine, expectedLine string
			if i < len(gotLines1) {
				gotLine = gotLines1[i]
			}
			if i < len(expectedLines2) {
				expectedLine = expectedLines2[i]
			}
			if strings.TrimSpace(gotLine) == strings.TrimSpace(expectedLine) {
				fmt.Printf("OK:       %s\n", gotLine)
			} else {
				fmt.Printf("GOT:      %s\n", gotLine)
				fmt.Printf("EXPECTED: %s\n", expectedLine)
				t.Fail()
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
        result.children = source["children"] ? source["children"].map(function(element: any) { return Test.createFrom(element); }) : null;
        return result;
    }

}`
	testConverter(t, converter, desiredResult)
}

func TestArrayOfArrays(t *testing.T) {
	type Key struct {
		Key string `json:"key"`
	}
	type Keyboard struct {
		Keys [][]Key `json:"keys"`
	}

	converter := New()

	converter.AddType(reflect.TypeOf(Keyboard{}))
	converter.CreateFromMethod = true
	converter.BackupDir = ""

	desiredResult := `export class Key {
    key: string;

    static createFrom(source: any) {
        if ('string' === typeof source) source = JSON.parse(source);
        const result = new Key();
        result.key = source["key"];
        return result;
    }

}
export class Keyboard {
    keys: Key[][];

    static createFrom(source: any) {
        if ('string' === typeof source) source = JSON.parse(source);
        const result = new Keyboard();
        result.keys = source["keys"] ? source["keys"].map(function(element: any) { return Key.createFrom(element); }) : null;
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

type NumberTime time.Time

func (t NumberTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", time.Time(t).Unix())), nil
}

func TestTypeAlias(t *testing.T) {
	type Person struct {
		Birth NumberTime `json:"birth" ts_type:"number"`
	}

	converter := New()

	converter.AddType(reflect.TypeOf(Person{}))
	converter.CreateFromMethod = false
	converter.BackupDir = ""

	desiredResult := `export class Person {
    birth: number;
}`
	testConverter(t, converter, desiredResult)
}

type MSTime struct {
	time.Time
}

func (MSTime) UnmarshalJSON([]byte) error   { return nil }
func (MSTime) MarshalJSON() ([]byte, error) { return []byte("1111"), nil }

func TestOverrideCustomType(t *testing.T) {

	type SomeStruct struct {
		Time MSTime `json:"time" ts_type:"number"`
	}
	var _ json.Marshaler = new(MSTime)
	var _ json.Unmarshaler = new(MSTime)

	converter := New()

	converter.AddType(reflect.TypeOf(SomeStruct{}))
	converter.CreateFromMethod = false
	converter.BackupDir = ""

	desiredResult := `export class SomeStruct {
    time: number;
}`
	testConverter(t, converter, desiredResult)

	byts, _ := json.Marshal(SomeStruct{Time: MSTime{Time: time.Now()}})
	if string(byts) != `{"time":1111}` {
		t.Error("marhshalling failed")
	}
}

type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

func (w Weekday) TSName() string {
	switch w {
	case Sunday:
		return "SUNDAY"
	case Monday:
		return "MONDAY"
	case Tuesday:
		return "TUESDAY"
	case Wednesday:
		return "WEDNESDAY"
	case Thursday:
		return "THURSDAY"
	case Friday:
		return "FRIDAY"
	case Saturday:
		return "SATURDAY"
	default:
		return "???"

	}
}

var AllWeekdays = []Weekday{
	Sunday,
	Monday,
	Tuesday,
	Wednesday,
	Thursday,
	Friday,
	Saturday,
}

type Holliday struct {
	Name    string  `json:"name"`
	Weekday Weekday `json:"weekday"`
}

func TestEnum(t *testing.T) {
	converter := New().
		AddType(reflect.TypeOf(Holliday{})).
		AddEnum(AllWeekdays).
		WithCreateFromMethod(false).
		WithBackupDir("")

	desiredResult := `export enum Weekday {
	SUNDAY = 0,
	MONDAY = 1,
	TUESDAY = 2,
	WEDNESDAY = 3,
	THURSDAY = 4,
	FRIDAY = 5,
	SATURDAY = 6,
}
export class Holliday {
	name: string;
	weekday: Weekday;
}`
	testConverter(t, converter, desiredResult)
}
