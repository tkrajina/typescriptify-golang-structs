import { Person } from "./example_output";

class Tests {
    test1() {
        const person = Person.createFrom({
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
    }

    test2() {
        Person.createFrom({
            "name": "Ovo Ono",
            "nicknames": ["aaa", "bbb"],
            "personal_info": {
                "hobby": ["1", "2"],
                "pet_name": "nera"
            }
        });
        console.log("OK");
    }

    test3() {
        const person = Person.createFrom({
            "name": "Ovo Ono",
            "nicknames": ["aaa", "bbb"],
            "personal_info": {}
        });
        console.log("OK");
    }

    test4() {
        const person = Person.createFrom({
        });
        console.log("OK");
    }
}

const tests = new Tests();

for (const testMethod in tests) {
    tests[testMethod]();
}