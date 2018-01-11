# A Golang JSON to TypeScript model converter

Use the commantline tool:

    tscriptify -package=package/with/your/models -target=target_ts_file.ts Model1 Model2

If all your structs are in one file, you can convert them with:

    tscriptify -package=package/with/your/models -target=target_ts_file.ts path/to/file/with/structs.go

Or by using it from your code:

    converter := typescriptify.New()
    converter.Add(Person{})
    converter.Add(Dummy{})
    err := converter.ConvertToFile("ts/models.ts")
    if err != nil {
        panic(err.Error())
    }

## Models and conversion

If the `Person` structs contain a reference to the `Address` struct, then you don't have to add `Address` explicitly. Only fields with a valid `json` tag will be converted to TypeScript models.

Example input structs:

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

Generated TypeScript:

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

In TypeScript you can just cast your javascript object in any of those models:

    var person = <Person> {"name":"Me myself","nicknames":["aaa", "bbb"]};
    console.log(person.name);
    // The TypeScript compiler will throw an error for this line
    console.log(person.something);

Any custom code can be added to Typescript models:

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

The lines between `//[Address:]` and `//[end]` will be left intact after `ConvertToFile()`.

If your custom code contain methods, then just casting yout object to the target class (with `<Person> {...}`) won't work because the casted object won't contain your methods.
In that case, you can configure the converter to create static `createFrom` methods:

    converter := typescriptify.New()
	converter.CreateFromMethod = true
	converter.Indent = "    "

The TypeScript code will now be:

    class Person {
        name: string;
        personal_info: PersonalInfo;
        nicknames: string[];
        addresses: Address[];

        static createFrom(source: any) {
            var result = new Person();
            result.name = source["name"];
            result.personal_info = source["personal_info"] ? PersonalInfo.createFrom(source["personal_info"]) : null;
            result.nicknames = source["nicknames"];
            result.addresses = source["addresses"] ? source["addresses"].map(function(element) { return Address.createFrom(element); }) : null;
            return result;
        }

        //[Person:]

        yourMethod = () => {
            return "name:" + this.name;
        }

        //[end]
    }

And now, instead of casting to `Person` you need to:

    var person = Person.createFrom({"name":"Me myself","nicknames":["aaa", "bbb"]});

If you use golang JSON structs as responses from your API, you may want to have a common prefix for all the generated models:

    converter := typescriptify.New()
    converter.Prefix("API_")
    converter.Add(Person{})

The model name will be `API_Person` instead of `Person`.

License
-------

This library is licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

