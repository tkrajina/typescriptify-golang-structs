# A Golang JSON to TypeScript model converter

    converter := typescriptify.New()
    converter.Add(Person{})
    converter.Add(Dummy{})
    err := converter.ConvertToFile("ts/models.ts")
    if err != nil {
        panic(err.Error())
    }

If the `Person` structs contain a reference to the `Address` struct, then you don't have to add `Address` explicitly. Only fields with a valid `json` tag will be converted to TypeScript models.

Example input structs:

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

In TypeScript you can just cast your JSON in any of those models:

    var json = <Person> {"name":"Me myself","nicknames":["aaa", "bbb"]};
    console.log(json.name);
    // The TypeScript compiler will throw an error for this line
    console.log(json.something);

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

If you use golang JSON structs as responses from your API, you may want to have a common prefix for all the generated models:

    converter := typescriptify.New()
    converter.Prefix("API_")
    converter.Add(Person{})

The model name will be `API_Person` instead of `Person`.

License
-------

This library is licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

