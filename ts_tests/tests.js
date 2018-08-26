"use strict";
exports.__esModule = true;
var example_output_1 = require("./example_output");
var Tests = /** @class */ (function () {
    function Tests() {
    }
    Tests.prototype.testMisc = function () {
        var person = example_output_1.Person.createFrom({
            "name": "Ovo Ono",
            "nicknames": ["aaa", "bbb"],
            "personal_info": {
                "hobby": ["1", "2"],
                "pet_name": "nera"
            },
            "addresses": [
                { "city": "aaa", "number": 13 },
                { "city": "bbb", "number": 14 }
            ]
        });
        if (person.getInfo() != "name:Ovo Ono") {
            throw new Error("Person method not there");
        }
        if (person.personal_info.hobby.length != 2) {
            alert("No hobbies found");
            return;
        }
        if (person.addresses.length != 2) {
            alert("No addresses found");
            return;
        }
        if (person.addresses[1].getAddressString() != "bbb 14") {
            alert("Address methodincorrect");
            return;
        }
        console.log("OK");
    };
    Tests.prototype.testMaps = function () {
        var person = example_output_1.Person.createFrom({
            "children": {
                "eve": {
                    "name": "Eve",
                    "nicknames": ["the one"]
                },
                "adam": {
                    "name": "Adam"
                }
            },
            "children_age": {
                "eve": 19,
                "adam": 20
            }
        });
        if (!person.children["adam"]) {
            throw new Error("No Adam");
        }
        if (!person.children["eve"]) {
            throw new Error("No Eve");
        }
        if (person.children["eve"].nicknames[0] != "the one") {
            throw new Error("No Eve");
        }
        if (person.children_age["adam"] != 20) {
            throw new Error("No Adam");
        }
        if (person.children_age["eve"] != 19) {
            throw new Error("No Eve");
        }
        console.log("OK");
    };
    return Tests;
}());
var tests = new Tests();
for (var testMethod in tests) {
    tests[testMethod]();
}
