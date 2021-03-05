package typescriptify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	Metadata  string    `json:"metadata" ts_type:"{[key:string]:string}" ts_transform:"JSON.parse(__VALUE__ || \"{}\")"`
	Friends   []*Person `json:"friends"`
	Dummy     Dummy     `json:"a"`
}

func TestTypescriptifyWithTypes(t *testing.T) {
	t.Parallel()
	converter := New()

	converter.AddType(reflect.TypeOf(Person{}))
	converter.CreateFromMethod = false
	converter.CreateConstructor = false
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
	testConverter(t, converter, false, desiredResult, nil)
}

func TestTypescriptifyWithCustomImports(t *testing.T) {
	t.Parallel()
	converter := New()

	converter.AddType(reflect.TypeOf(Person{}))
	converter.CreateFromMethod = false
	converter.BackupDir = ""
	converter.AddImport("//import { Decimal } from 'decimal.js'")
	converter.CreateConstructor = false

	desiredResult := `
//import { Decimal } from 'decimal.js'

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
	testConverter(t, converter, false, desiredResult, nil)
}

func TestTypescriptifyWithInstances(t *testing.T) {
	t.Parallel()
	converter := New()

	converter.Add(Person{})
	converter.Add(Dummy{})
	converter.CreateFromMethod = false
	converter.DontExport = true
	converter.BackupDir = ""
	converter.CreateConstructor = false

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
	testConverter(t, converter, false, desiredResult, nil)
}

func TestTypescriptifyWithInterfaces(t *testing.T) {
	t.Parallel()
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
	testConverter(t, converter, true, desiredResult, nil)
}

func TestTypescriptifyWithDoubleClasses(t *testing.T) {
	t.Parallel()
	converter := New()

	converter.AddType(reflect.TypeOf(Person{}))
	converter.AddType(reflect.TypeOf(Person{}))
	converter.CreateFromMethod = false
	converter.CreateConstructor = false
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
	testConverter(t, converter, false, desiredResult, nil)
}

func TestWithPrefixes(t *testing.T) {
	t.Parallel()
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
        this.addresses = this.convertValues(source["addresses"], test_Address_test);
        this.address = this.convertValues(source["address"], test_Address_test);
		this.metadata = JSON.parse(source["metadata"] || "{}");
		this.friends = this.convertValues(source["friends"], test_Person_test);
		this.a = this.convertValues(source["a"], test_Dummy_test);
    }

	` + tsConvertValuesFunc + `
}`
	jsn := jsonizeOrPanic(Person{
		Address:   &Address{Text1: "txt1"},
		Addresses: []Address{{Text1: "111"}},
		Metadata:  `{"something": "aaa"}`,
	})
	testConverter(t, converter, true, desiredResult, []string{
		`new test_Person_test()`,
		`JSON.stringify(new test_Person_test()?.metadata) === "{}"`,
		`!(new test_Person_test()?.address)`,
		`!(new test_Person_test()?.addresses)`,
		`!(new test_Person_test()?.addresses)`,

		`new test_Person_test(` + jsn + ` as any)`,
		`new test_Person_test(` + jsn + ` as any)?.metadata?.something === "aaa"`,
		`(new test_Person_test(` + jsn + ` as any)?.address as test_Address_test).text === "txt1"`,
		`new test_Person_test(` + jsn + ` as any)?.addresses?.length === 1`,
		`(new test_Person_test(` + jsn + ` as any)?.addresses[0] as test_Address_test)?.text === "111"`,
	})
}

func testConverter(t *testing.T, converter *TypeScriptify, strictMode bool, desiredResult string, tsExpressionAndDesiredResults []string) {
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
			if assert.Equal(t, strings.TrimSpace(expectedLine), strings.TrimSpace(gotLine), "line #%d", 1+i) {
				fmt.Printf("OK:       %s\n", gotLine)
			} else {
				t.FailNow()
			}
		}
	}

	if t.Failed() {
		t.FailNow()
	}

	testTypescriptExpression(t, strictMode, typeScriptCode, tsExpressionAndDesiredResults)
}

func testTypescriptExpression(t *testing.T, strictMode bool, baseScript string, tsExpressionAndDesiredResults []string) {
	f, err := ioutil.TempFile(os.TempDir(), "*.ts")
	assert.Nil(t, err)
	assert.NotNil(t, f)

	_, _ = f.WriteString(baseScript)
	_, _ = f.WriteString("\n")
	for n, expr := range tsExpressionAndDesiredResults {
		_, _ = f.WriteString("// " + expr + "\n")
		_, _ = f.WriteString(`if (` + expr + `) { console.log("#` + fmt.Sprint(1+n) + ` OK") } else { throw new Error() }`)
		_, _ = f.WriteString("\n\n")
	}

	fmt.Println("tmp ts: ", f.Name())
	var byts []byte
	if strictMode {
		byts, err = exec.Command("tsc", "--strict", f.Name()).CombinedOutput()
	} else {
		byts, err = exec.Command("tsc", f.Name()).CombinedOutput()
	}
	assert.Nil(t, err, string(byts))

	jsFile := strings.Replace(f.Name(), ".ts", ".js", 1)
	fmt.Println("executing:", jsFile)
	byts, err = exec.Command("node", jsFile).CombinedOutput()
	assert.Nil(t, err, string(byts))
}

func TestTypescriptifyCustomType(t *testing.T) {
	t.Parallel()
	type TestCustomType struct {
		Map map[string]int `json:"map" ts_type:"{[key: string]: number}"`
	}

	converter := New()

	converter.AddType(reflect.TypeOf(TestCustomType{}))
	converter.CreateFromMethod = false
	converter.BackupDir = ""
	converter.CreateConstructor = false

	desiredResult := `export class TestCustomType {
        map: {[key: string]: number};
}`
	testConverter(t, converter, false, desiredResult, nil)
}

func TestDate(t *testing.T) {
	t.Parallel()
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

	jsn := jsonizeOrPanic(TestCustomType{Time: time.Date(2020, 10, 9, 8, 9, 0, 0, time.UTC)})
	testConverter(t, converter, true, desiredResult, []string{
		`new TestCustomType(` + jsonizeOrPanic(jsn) + `).time instanceof Date`,
		//`console.log(new TestCustomType(` + jsonizeOrPanic(jsn) + `).time.toJSON())`,
		`new TestCustomType(` + jsonizeOrPanic(jsn) + `).time.toJSON() === "2020-10-09T08:09:00.000Z"`,
	})
}

func TestDateWithoutTags(t *testing.T) {
	t.Parallel()
	type TestCustomType struct {
		Time time.Time `json:"time"`
	}

	// Test with custom field options defined per-one-struct:
	converter1 := New()
	converter1.Add(NewStruct(TestCustomType{}).WithFieldOpts(time.Time{}, TypeOptions{TSType: "Date", TSTransform: "new Date(__VALUE__)"}))
	converter1.CreateFromMethod = true
	converter1.BackupDir = ""

	// Test with custom field options defined globally:
	converter2 := New()
	converter2.Add(reflect.TypeOf(TestCustomType{}))
	converter2.ManageType(time.Time{}, TypeOptions{TSType: "Date", TSTransform: "new Date(__VALUE__)"})
	converter2.CreateFromMethod = true
	converter2.BackupDir = ""

	for _, converter := range []*TypeScriptify{converter1, converter2} {
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

		jsn := jsonizeOrPanic(TestCustomType{Time: time.Date(2020, 10, 9, 8, 9, 0, 0, time.UTC)})
		testConverter(t, converter, true, desiredResult, []string{
			`new TestCustomType(` + jsonizeOrPanic(jsn) + `).time instanceof Date`,
			//`console.log(new TestCustomType(` + jsonizeOrPanic(jsn) + `).time.toJSON())`,
			`new TestCustomType(` + jsonizeOrPanic(jsn) + `).time.toJSON() === "2020-10-09T08:09:00.000Z"`,
		})
	}
}

func TestRecursive(t *testing.T) {
	t.Parallel()
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
        this.children = this.convertValues(source["children"], Test);
    }

	` + tsConvertValuesFunc + `
}`
	testConverter(t, converter, true, desiredResult, nil)
}

func TestArrayOfArrays(t *testing.T) {
	t.Parallel()
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
        this.keys = this.convertValues(source["keys"], Key);
    }

	` + tsConvertValuesFunc + `
}`
	testConverter(t, converter, true, desiredResult, nil)
}

func TestFixedArray(t *testing.T) {
	t.Parallel()
	type Sub struct{}
	type Tmp struct {
		Arr  [3]string `json:"arr"`
		Arr2 [3]Sub    `json:"arr2"`
	}

	converter := New()

	converter.AddType(reflect.TypeOf(Tmp{}))
	converter.CreateFromMethod = false
	converter.BackupDir = ""

	desiredResult := `export class Sub {


    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);

    }
}
export class Tmp {
    arr: string[];
    arr2: Sub[];

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.arr = source["arr"];
        this.arr2 = this.convertValues(source["arr2"], Sub);
    }

	` + tsConvertValuesFunc + `
}
`
	testConverter(t, converter, true, desiredResult, nil)
}

func TestAny(t *testing.T) {
	t.Parallel()
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
	testConverter(t, converter, true, desiredResult, nil)
}

type NumberTime time.Time

func (t NumberTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", time.Time(t).Unix())), nil
}

func TestTypeAlias(t *testing.T) {
	t.Parallel()
	type Person struct {
		Birth NumberTime `json:"birth" ts_type:"number"`
	}

	converter := New()

	converter.AddType(reflect.TypeOf(Person{}))
	converter.CreateFromMethod = false
	converter.BackupDir = ""
	converter.CreateConstructor = false

	desiredResult := `export class Person {
    birth: number;
}`
	testConverter(t, converter, false, desiredResult, nil)
}

type MSTime struct {
	time.Time
}

func (MSTime) UnmarshalJSON([]byte) error   { return nil }
func (MSTime) MarshalJSON() ([]byte, error) { return []byte("1111"), nil }

func TestOverrideCustomType(t *testing.T) {
	t.Parallel()

	type SomeStruct struct {
		Time MSTime `json:"time" ts_type:"number"`
	}
	var _ json.Marshaler = new(MSTime)
	var _ json.Unmarshaler = new(MSTime)

	converter := New()

	converter.AddType(reflect.TypeOf(SomeStruct{}))
	converter.CreateFromMethod = false
	converter.BackupDir = ""
	converter.CreateConstructor = false

	desiredResult := `export class SomeStruct {
    time: number;
}`
	testConverter(t, converter, false, desiredResult, nil)

	byts, _ := json.Marshal(SomeStruct{Time: MSTime{Time: time.Now()}})
	assert.Equal(t, `{"time":1111}`, string(byts))
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
	t.Parallel()
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
		testConverter(t, converter, true, desiredResult, nil)
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
	t.Parallel()
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
	testConverter(t, converter, true, desiredResult, nil)
}

type Animal int

const (
	Cat Animal = iota
	Dog
)

func (a Animal) String() string {
	switch a {
	case Cat:
		return "CAT"
	case Dog:
		return "DOG"
	default:
		return "???"
	}
}

var allAnimalsV1 = []Animal{
	Cat,
	Dog,
}

// Another way to specify enums:
var allAnimalsV2 = []struct {
	Value  Animal
	TSName string
}{
	{Cat, "CAT"},
	{Dog, "DOG"},
}

func TestStringer(t *testing.T) {
	t.Parallel()
	for _, allAnimals := range []interface{}{allAnimalsV1, allAnimalsV2} {
		converter := New().
			AddEnum(allAnimals).
			WithConstructor(false).
			WithCreateFromMethod(true).
			WithBackupDir("")

		desiredResult := `export enum Animal {
	CAT = 0,
	DOG = 1,
}
`
		testConverter(t, converter, true, desiredResult, nil)
	}
}

func TestConstructorWithReferences(t *testing.T) {
	t.Parallel()
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
        this.addresses = this.convertValues(source["addresses"], Address);
        this.address = this.convertValues(source["address"], Address);
		this.metadata = JSON.parse(source["metadata"] || "{}");
        this.friends = this.convertValues(source["friends"], Person);
        this.a = this.convertValues(source["a"], Dummy);
    }

	` + tsConvertValuesFunc + `
}`
	testConverter(t, converter, true, desiredResult, nil)
}

type WithMap struct {
	Map        map[string]int      `json:"simpleMap"`
	MapObjects map[string]Address  `json:"mapObjects"`
	PtrMap     *map[string]Address `json:"ptrMapObjects"`
}

func TestMaps(t *testing.T) {
	t.Parallel()
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
              this.simpleMap = source["simpleMap"];
			  this.mapObjects = this.convertValues(source["mapObjects"], Address, true);
			  this.ptrMapObjects = this.convertValues(source["ptrMapObjects"], Address, true);
		  }

		  ` + tsConvertValuesFunc + `
      }
`

	json := WithMap{
		Map:        map[string]int{"aaa": 1},
		MapObjects: map[string]Address{"bbb": {Duration: 1.0, Text1: "txt1"}},
		PtrMap:     &map[string]Address{"ccc": {Duration: 2.0, Text1: "txt2"}},
	}

	testConverter(t, converter, true, desiredResult, []string{
		`new WithMap(` + jsonizeOrPanic(json) + `).simpleMap.aaa == 1`,
		`(new WithMap(` + jsonizeOrPanic(json) + `).mapObjects.bbb) instanceof Address`,
		`!((new WithMap(` + jsonizeOrPanic(json) + `).mapObjects.bbb) instanceof WithMap)`,
		`new WithMap(` + jsonizeOrPanic(json) + `).mapObjects.bbb.duration == 1`,
		`new WithMap(` + jsonizeOrPanic(json) + `).mapObjects.bbb.text === "txt1"`,
		`(new WithMap(` + jsonizeOrPanic(json) + `)?.ptrMapObjects?.ccc) instanceof Address`,
		`!((new WithMap(` + jsonizeOrPanic(json) + `)?.ptrMapObjects?.ccc) instanceof WithMap)`,
		`new WithMap(` + jsonizeOrPanic(json) + `)?.ptrMapObjects?.ccc?.duration === 2`,
		`new WithMap(` + jsonizeOrPanic(json) + `)?.ptrMapObjects?.ccc?.text === "txt2"`,
	})
}

func TestPTR(t *testing.T) {
	t.Parallel()
	type Person struct {
		Name *string `json:"name"`
	}

	converter := New()
	converter.CreateFromMethod = false
	converter.BackupDir = ""
	converter.CreateConstructor = false
	converter.Add(Person{})

	desiredResult := `export class Person {
    name?: string;
}`
	testConverter(t, converter, true, desiredResult, nil)
}

type PersonWithPtrName struct {
	*HasName
}

func TestAnonymousPtr(t *testing.T) {
	t.Parallel()
	var p PersonWithPtrName
	p.HasName = &HasName{}
	p.Name = "JKLJKL"
	converter := New().
		AddType(reflect.TypeOf(PersonWithPtrName{})).
		WithConstructor(true).
		WithCreateFromMethod(false).
		WithBackupDir("")

	desiredResult := `
      export class PersonWithPtrName {
          name: string;
      
          constructor(source: any = {}) {
              if ('string' === typeof source) source = JSON.parse(source);
              this.name = source["name"];
          }
      }
`
	testConverter(t, converter, true, desiredResult, nil)
}

func jsonizeOrPanic(i interface{}) string {
	byts, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return string(byts)
}

func TestTestConverter(t *testing.T) {
	t.Parallel()

	ts := `class Converter {
		` + tsConvertValuesFunc + `
}
const converter = new Converter();

class Address {
    street: string;
    number: number;
    
    constructor(a: any) {
        this.street = a["street"];
        this.number = a["number"];
    }
}
`

	testTypescriptExpression(t, true, ts, []string{
		`(converter.convertValues(null, Address)) === null`,
		`(converter.convertValues([], Address)).length === 0`,
		`(converter.convertValues({}, Address)) instanceof Address`,
		`!(converter.convertValues({}, Address, true) instanceof Address)`,

		`(converter.convertValues([{street: "aaa", number: 19}] as any, Address) as Address[]).length == 1`,
		`(converter.convertValues([{street: "aaa", number: 19}] as any, Address) as Address[])[0] instanceof Address`,
		`(converter.convertValues([{street: "aaa", number: 19}] as any, Address) as Address[])[0].number === 19`,
		`(converter.convertValues([{street: "aaa", number: 19}] as any, Address) as Address[])[0].street === "aaa"`,

		`(converter.convertValues([[{street: "aaa", number: 19}]] as any, Address) as Address[]).length == 1`,
		`(converter.convertValues([[{street: "aaa", number: 19}]] as any, Address) as Address[][])[0][0] instanceof Address`,
		`(converter.convertValues([[{street: "aaa", number: 19}]] as any, Address) as Address[][])[0][0].number === 19`,
		`(converter.convertValues([[{street: "aaa", number: 19}]] as any, Address) as Address[][])[0][0].street === "aaa"`,

		`Object.keys((converter.convertValues({"first": {street: "aaa", number: 19}}, Address, true) as {[_: string]: Address})).length == 1`,
		`(converter.convertValues({"first": {street: "aaa", number: 19}} as any, Address, true) as {[_: string]: Address})["first"] instanceof Address`,
		`(converter.convertValues({"first": {street: "aaa", number: 19}} as any, Address, true) as {[_: string]: Address})["first"].number === 19`,
		`(converter.convertValues({"first": {street: "aaa", number: 19}} as any, Address, true) as {[_: string]: Address})["first"].street === "aaa"`,
	})
}

func TestIgnoredPTR(t *testing.T) {
	t.Parallel()
	type PersonWithIgnoredPtr struct {
		Name     string  `json:"name"`
		Nickname *string `json:"-"`
	}

	converter := New()
	converter.CreateFromMethod = false
	converter.BackupDir = ""
	converter.Add(PersonWithIgnoredPtr{})

	desiredResult := `
      export class PersonWithIgnoredPtr {
          name: string;
      
          constructor(source: any = {}) {
              if ('string' === typeof source) source = JSON.parse(source);
              this.name = source["name"];
          }
      }
`
	testConverter(t, converter, true, desiredResult, nil)
}
