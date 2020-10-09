# A Golang JSON to TypeScript model converter

## Installation

The command-line tool:

```
go get github.com/tkrajina/typescriptify-golang-structs/tscriptify
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
```

Generated TypeScript:

```typescript
class Dummy {
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
}
```

In TypeScript you can just cast your javascript object in any of those models:

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

```golang
converter := typescriptify.New().
    Add(Address)
```

The TypeScript code will now be:

```typescript
export class Person {
    name: string;
    nicknames: string[];
    addresses: Address[];
    address: Address;
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

    //[Person:]

    yourMethod = () => {
        return "name:" + this.name;
    }

    //[end]
}
```

And now, instead of casting to `Person` you need to:

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

## Custom types

If your field has a type not supported by typescriptify which can be JSONized as is, then you can use the `ts_type` tag to specify the typescript type to use:

```golang
type Data struct {
    Counters map[string]int `json:"counters" ts_type:"{[key: string]: number}"`
}
```

...will create:

```typescript
export class Data {
        counters: {[key: string]: number};
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

Additionally, you can tell the library to automatically use a given Typescript type and custom transformation for a type:

```golang
converter := typescriptify.New()
converter.ManageType(time.Time{}, "Date", "new Date(__VALUE__)")
```

This is how now `time.Time` is manage in the library by default.

If you only want to change `ts_transform` but not `ts_type`, you can pass an empty string:

```golang
converter := typescriptify.New()
converter.ManageType(time.Time{}, "", "new Date(__VALUE__)")
```

Calling `.ManageType` multiple times with the same type overrides previous additions.

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

