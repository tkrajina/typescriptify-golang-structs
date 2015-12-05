package main

import "github.com/tkrajina/typescriptify-golang-structs/typescriptify"

type Address struct {
	// Used in html
	City    float64 `json:"city"`
	Country string  `json:"country,omitempty"`
}

type Person struct {
	Name      string    `json:"name"`
	Nicknames []string  `json:"nicknames"`
	Addresses []Address `json:"addresses"`
}

func main() {
	converter := typescriptify.New()

	converter.Add(Person{})

	err := converter.ConvertToFile("example_output.ts")
	if err != nil {
		panic(err.Error())
	}
}
