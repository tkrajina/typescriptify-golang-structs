/* Do not change, this code is generated from Golang structs */


class Address {
    city: string;
    number: number;
    country: string;
    static fromJSON(json: any) {
        var result = new Address();
        result.city = json["city"];
        result.number = json["number"];
        result.country = json["country"];
        return result;
    }
    //[Address:]
    /* Custom code here */

    getAddressString = () => {
        return this.city + " " + this.number;
    }

    //[end]
}
class PersonalInfo {
    hobby: string[];
    pet_name: string;
    static fromJSON(json: any) {
        var result = new PersonalInfo();
        result.hobby = json["hobby"];
        result.pet_name = json["pet_name"];
        return result;
    }
    //[PersonalInfo:]

    getPersonalInfoString = () => {
        return "pet:" + this.pet_name;
    }

    //[end]
}
class Person {
    name: string;
    personal_info: PersonalInfo;
    nicknames: string[];
    addresses: Address[];
    static fromJSON(json: any) {
        var result = new Person();
        result.name = json["name"];
        result.personal_info = PersonalInfo.fromJSON(json["personal_info"]);
        result.nicknames = json["nicknames"];
        if (json["addresses"]) {
            result.addresses = json["addresses"].map(function(element) { return Address.fromJSON(element); });
        }
        return result;
    }
    //[Person:]

    getInfo = () => {
        return "name:" + this.name;
    }

    //[end]
}