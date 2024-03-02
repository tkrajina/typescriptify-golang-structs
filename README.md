# A Golang JSON to TypeScript model converter

## Installation

The command-line tool:

```
go install github.com/tkrajina/typescriptify-golang-structs/tscriptify
```

The library:

```
go get github.com/tkrajina/typescriptify-golang-structs
```

## Usage

Use the command line tool:

```
tscriptify -package=package/with/your/models -target=target_ts_file.ts Model1 Model2
```

If you need to import a custom type in Typescript, you can pass the import string:

```
tscriptify -package=package/with/your/models -target=target_ts_file.ts -import="import { Decimal } from 'decimal.js'" Model1 Model2
```

If all your structs are in one file, you can convert them with:

```
tscriptify -package=package/with/your/models -target=target_ts_file.ts path/to/file/with/structs.go
```

Or by using it from your code:

```golang
converter := typescriptify.New().
    Add(Person{}).
    Add(Dummy{})
err := converter.ConvertToFile("ts/models.ts")
if err != nil {
    panic(err.Error())
}
```

Command line options:

```
$ tscriptify --help
Usage of tscriptify:
-backup string
        Directory where backup files are saved
-package string
        Path of the package with models
-target string
        Target typescript file
```

## Models and conversion

If the `Person` structs contain a reference to the `Address` struct, then you don't have to add `Address` explicitly. Only fields with a valid `json` tag will be converted to TypeScript models.

Example input structs:

```golang
type Address struct {
	City    string  `json:"city"`
	Number  float64 `json:"number"`
	Country string  `json:"country,omitempty"`
}

type PersonalInfo struct {
	Hobbies []string `json:"hobby"`
	PetName string   `json:"pet_name"`
}

type Person struct {
	Name         string       `json:"name"`
	PersonalInfo PersonalInfo `json:"personal_info"`
	Nicknames    []string     `json:"nicknames"`
	Addresses    []Address    `json:"addresses"`
	Address      *Address     `json:"address"`
	Metadata     []byte       `json:"metadata" ts_type:"{[key:string]:string}"`
	Friends      []*Person    `json:"friends"`
}
```

Generated TypeScript:

```typescript
export class Address {
    city: string;
    number: number;
    country?: string;

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.city = source["city"];
        this.number = source["number"];
        this.country = source["country"];
    }
}
export class PersonalInfo {
    hobby: string[];
    pet_name: string;

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.hobby = source["hobby"];
        this.pet_name = source["pet_name"];
    }
}
export class Person {
    name: string;
    personal_info: PersonalInfo;
    nicknames: string[];
    addresses: Address[];
    address?: Address;
    metadata: {[key:string]:string};
    friends: Person[];

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.name = source["name"];
        this.personal_info = this.convertValues(source["personal_info"], PersonalInfo);
        this.nicknames = source["nicknames"];
        this.addresses = this.convertValues(source["addresses"], Address);
        this.address = this.convertValues(source["address"], Address);
        this.metadata = source["metadata"];
        this.friends = this.convertValues(source["friends"], Person);
    }

	convertValues(a: any, classs: any, asMap: boolean = false): any {
		if (!a) {
			return a;
		}
		if (a.slice) {
			return (a as any[]).map(elem => this.convertValues(elem, classs));
		} else if ("object" === typeof a) {
			if (asMap) {
				for (const key of Object.keys(a)) {
					a[key] = new classs(a[key]);
				}
				return a;
			}
			return new classs(a);
		}
		return a;
	}
}
```

If you prefer interfaces, the output is:

```typescript
export interface Address {
    city: string;
    number: number;
    country?: string;
}
export interface PersonalInfo {
    hobby: string[];
    pet_name: string;
}
export interface Person {
    name: string;
    personal_info: PersonalInfo;
    nicknames: string[];
    addresses: Address[];
    address?: Address;
    metadata: {[key:string]:string};
    friends: Person[];
}
```

In TypeScript you can just cast your json object in any of those models:

```typescript
var person = <Person> {"name":"Me myself","nicknames":["aaa", "bbb"]};
console.log(person.name);
// The TypeScript compiler will throw an error for this line
console.log(person.something);
```

## Custom Typescript code

Any custom code can be added to Typescript models:

```typescript
class Address {
        street : string;
        no : number;
        //[Address:]
        country: string;
        getStreetAndNumber() {
            return street + " " + number;
        }
        //[end]
}
```

The lines between `//[Address:]` and `//[end]` will be left intact after `ConvertToFile()`.

If your custom code contain methods, then just casting yout object to the target class (with `<Person> {...}`) won't work because the casted object won't contain your methods.

In that case use the constructor:

```typescript
var person = new Person({"name":"Me myself","nicknames":["aaa", "bbb"]});
```

If you use golang JSON structs as responses from your API, you may want to have a common prefix for all the generated models:

```golang
converter := typescriptify.New().
converter.Prefix = "API_"
converter.Add(Person{})
```

The model name will be `API_Person` instead of `Person`.

## Field comments

Field documentation comments can be added with the `ts_doc` tag:

```golang
type Person struct {
	Name string `json:"name" ts_doc:"This is a comment"`
}
```

Generated typescript:

```typescript
export class Person {
	/** This is a comment */
	name: string;
}
```

## Custom types

If your field has a type not supported by typescriptify which can be JSONized as is, then you can use the `ts_type` tag to specify the typescript type to use:

```golang
type Data struct {
    Counters map[string]int `json:"counters" ts_type:"CustomType"`
}
```

...will create:

```typescript
export class Data {
        counters: CustomType;
}
```

If the JSON field needs some special handling before converting it to a javascript object, use `ts_transform`.
For example:

```golang
type Data struct {
    Time time.Time `json:"time" ts_type:"Date" ts_transform:"new Date(__VALUE__)"`
}
```

Generated typescript:

```typescript
export class Date {
	time: Date;

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.time = new Date(source["time"]);
    }
}
```

In this case, you should always use `new Data(json)` instead of just casting `<Data>json`.

If you use a custom type that has to be imported, you can do the following:

```golang
converter := typescriptify.New()
converter.AddImport("import Decimal from 'decimal.js'")
```

This will put your import on top of the generated file.

## Global custom types

Additionally, you can tell the library to automatically use a given Typescript type and custom transformation for a type:

```golang
converter := New()
converter.ManageType(time.Time{}, TypeOptions{TSType: "Date", TSTransform: "new Date(__VALUE__)"})
```

If you only want to change `ts_transform` but not `ts_type`, you can pass an empty string.

## Enums

There are two ways to create enums. 

### Enums with TSName()

In this case you must provide a list of enum values and the enum type must have a `TSName() string` method

```golang
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

var AllWeekdays = []Weekday{ Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, }

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
```

If this is too verbose for you, you can also provide a list of enums and enum names:

```golang
var AllWeekdays = []struct {
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
```

Then, when converting models `AddEnum()` to specify the enum:

```golang
    converter := New().
        AddEnum(AllWeekdays)
```

The resulting code will be:

```typescript
export enum Weekday {
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
}
```

## License

This library is licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

