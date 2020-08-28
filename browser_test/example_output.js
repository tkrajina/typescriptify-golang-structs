"use strict";
/* Do not change, this code is generated from Golang structs */
exports.__esModule = true;
var Address = /** @class */ (function () {
    function Address(source) {
        var _this = this;
        //[Address:]
        /* Custom code here */
        this.getAddressString = function () {
            return _this.city + " " + _this.number;
        };
        if ('string' === typeof source)
            source = JSON.parse(source);
        this.city = source["city"];
        this.number = source["number"];
        this.country = source["country"];
    }
    Address.createFrom = function (source) {
        return new Address(source);
    };
    return Address;
}());
exports.Address = Address;
var PersonalInfo = /** @class */ (function () {
    function PersonalInfo(source) {
        var _this = this;
        //[PersonalInfo:]
        this.getPersonalInfoString = function () {
            return "pet:" + _this.pet_name;
        };
        if ('string' === typeof source)
            source = JSON.parse(source);
        this.hobby = source["hobby"];
        this.pet_name = source["pet_name"];
    }
    PersonalInfo.createFrom = function (source) {
        return new PersonalInfo(source);
    };
    return PersonalInfo;
}());
exports.PersonalInfo = PersonalInfo;
var Person = /** @class */ (function () {
    function Person(source) {
        var _this = this;
        //[Person:]
        this.getInfo = function () {
            return "name:" + _this.name;
        };
        if ('string' === typeof source)
            source = JSON.parse(source);
        this.name = source["name"];
        this.personal_info = source["personal_info"] && new PersonalInfo(source["personal_info"]);
        this.nicknames = source["nicknames"];
        this.addresses = source["addresses"] && source["addresses"].map(function (element) { return new Address(element); });
        this.address = source["address"] && new Address(source["address"]);
        this.metadata = source["metadata"];
        this.friends = source["friends"] && source["friends"].map(function (element) { return new Person(element); });
    }
    Person.createFrom = function (source) {
        return new Person(source);
    };
    return Person;
}());
exports.Person = Person;
