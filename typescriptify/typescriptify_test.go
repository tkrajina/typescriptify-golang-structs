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
		address?: Address;
		metadata: {[key:string]:string};
		friends: Person[];
        a: Dummy;
}`
	testConverter(t, converter, desiredResult)
}

func TestTypescriptifyWithCustomImports(t *testing.T) {
	converter := New()

	converter.AddType(reflect.TypeOf(Person{}))
	converter.CreateFromMethod = false
	converter.BackupDir = ""
	converter.AddImport("import { Decimal } from 'decimal.js'")

	desiredResult := `
import { Decimal } from 'decimal.js'

export class Dummy {
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
		address?: Address;
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
		address?: Address;
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
		address?: Address;
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
		address?: Address;
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

    static createFrom(source: any = {}) {
		return new test_Dummy_test(source);
	}

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.something = source["something"];
    }
}
class test_Address_test {
    duration: number;
	text?: string;

    static createFrom(source: any = {}) {
		return new test_Address_test(source);
	}

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.duration = source["duration"];
        this.text = source["text"];
    }
}
class test_Person_test {
    name: string;
    nicknames: string[];
    addresses: test_Address_test[];
    address?: test_Address_test;
    metadata: {[key:string]:string};
    friends: test_Person_test[];
	a: test_Dummy_test;

    static createFrom(source: any = {}) {
		return new test_Person_test(source);
	}

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.name = source["name"];
        this.nicknames = source["nicknames"];
        this.addresses = source["addresses"] && source["addresses"].map((element: any) => new test_Address_test(element));
        this.address = source["address"] && new test_Address_test(source["address"]);
        this.metadata = source["metadata"];
        this.friends = source["friends"] && source["friends"].map((element: any) => new test_Person_test(element));
        this.a = source["a"] && new test_Dummy_test(source["a"]);
    }
}`
	testConverter(t, converter, desiredResult)
}

func testConverter(t *testing.T, converter *TypeScriptify, desiredResult string) {
	typeScriptCode, err := converter.Convert(nil)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("----------------------------------------------------------------------------------------------------")
	fmt.Println(typeScriptCode)
	fmt.Println("----------------------------------------------------------------------------------------------------")

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

    static createFrom(source: any = {}) {
        return new TestCustomType(source);
	}

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.time = new Date(source["time"]);
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

    static createFrom(source: any = {}) {
        return new Test(source);
	}

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.children = source["children"] && source["children"].map((element: any) => new Test(element));
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

    static createFrom(source: any = {}) {
        return new Key(source);
    }

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.key = source["key"];
    }
}
export class Keyboard {
    keys: Key[][];

    static createFrom(source: any = {}) {
        return new Keyboard(source);
    }

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.keys = source["keys"] && source["keys"].map((element: any) => new Key(element));
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

    static createFrom(source: any = {}) {
		return new Test(source);
	}

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.field = source["field"];
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

// One way to specify enums is to list all values and then every one must have a TSName() method
var allWeekdaysV1 = []Weekday{
	Sunday,
	Monday,
	Tuesday,
	Wednesday,
	Thursday,
	Friday,
	Saturday,
}

// Another way to specify enums:
var allWeekdaysV2 = []struct {
	Value  Weekday
	TSName string
}{
	{Sunday, "SUNDAY"},
	{Monday, "MONDAY"},
	{Tuesday, "TUESDAY"},
	{Wednesday, "WEDNESDAY"},
	{Thursday, "THURSDAY"},
	{Friday, "FRIDAY"},
	{Saturday, "SATURDAY"},
}

type Holliday struct {
	Name    string  `json:"name"`
	Weekday Weekday `json:"weekday"`
}

func TestEnum(t *testing.T) {
	for _, allWeekdays := range []interface{}{allWeekdaysV1, allWeekdaysV2} {
		converter := New().
			AddType(reflect.TypeOf(Holliday{})).
			AddEnum(allWeekdays).
			WithConstructor(false).
			WithCreateFromMethod(true).
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

	static createFrom(source: any = {}) {
        return new Holliday(source);
    }

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.name = source["name"];
        this.weekday = source["weekday"];
    }
}`
		testConverter(t, converter, desiredResult)
	}
}

type Gender string

const (
	MaleStr   Gender = "m"
	FemaleStr Gender = "f"
)

var allGenders = []struct {
	Value  Gender
	TSName string
}{
	{MaleStr, "MALE"},
	{FemaleStr, "FEMALE"},
}

func TestEnumWithStringValues(t *testing.T) {
	converter := New().
		AddEnum(allGenders).
		WithConstructor(false).
		WithCreateFromMethod(false).
		WithBackupDir("")

	desiredResult := `
export enum Gender {
	MALE = "m",
	FEMALE = "f",
}
`
	testConverter(t, converter, desiredResult)
}

func TestConstructorWithReferences(t *testing.T) {
	converter := New().
		AddType(reflect.TypeOf(Person{})).
		AddEnum(allWeekdaysV2).
		WithConstructor(true).
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
export class Dummy {
    something: string;

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.something = source["something"];
    }
}
export class Address {
    duration: number;
    text?: string;

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.duration = source["duration"];
        this.text = source["text"];
    }
}
export class Person {
    name: string;
    nicknames: string[];
    addresses: Address[];
    address?: Address;
    metadata: {[key:string]:string};
    friends: Person[];
    a: Dummy;

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.name = source["name"];
        this.nicknames = source["nicknames"];
        this.addresses = source["addresses"] && source["addresses"].map((element: any) => new Address(element));
        this.address = source["address"] && new Address(source["address"]);
        this.metadata = source["metadata"];
        this.friends = source["friends"] && source["friends"].map((element: any) => new Person(element));
        this.a = source["a"] && new Dummy(source["a"]);
    }
}`
	testConverter(t, converter, desiredResult)
}

type WithMap struct {
	Map        map[string]int      `json:"simpleMap"`
	MapObjects map[string]Address  `json:"mapObjects"`
	PtrMap     *map[string]Address `json:"ptrMapObjects"`
}

func TestMaps(t *testing.T) {
	converter := New().
		AddType(reflect.TypeOf(WithMap{})).
		WithConstructor(true).
		WithCreateFromMethod(false).
		WithBackupDir("")

	desiredResult := `
      export class Address {
          duration: number;
          text?: string;
      
          constructor(source: any = {}) {
              if ('string' === typeof source) source = JSON.parse(source);
              this.duration = source["duration"];
              this.text = source["text"];
          }
      }
      export class WithMap {
          simpleMap: {[key: string]: number};
      
          mapObjects: {[key: string]: Address};

          ptrMapObjects?: {[key: string]: Address};


          constructor(source: any = {}) {
              if ('string' === typeof source) source = JSON.parse(source);
              this.simpleMap = source["simpleMap"] ? source["simpleMap"] : null;
              this.mapObjects = source["mapObjects"] ? source["mapObjects"] : null;
              this.ptrMapObjects = source["ptrMapObjects"] ? source["ptrMapObjects"] : null;

          }
      }
`
	testConverter(t, converter, desiredResult)
}

func TestPTR(t *testing.T) {
	type Person struct {
		Name *string `json:"name"`
	}

	converter := New()
	converter.CreateFromMethod = false
	converter.BackupDir = ""
	converter.Add(Person{})

	desiredResult := `export class Person {
    name?: string;
}`
	testConverter(t, converter, desiredResult)
}
