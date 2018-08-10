"use strict";
/* Do not change, this code is generated from Golang structs */
exports.__esModule = true;
var Address = /** @class */ (function () {
    function Address() {
        var _this = this;
        //[Address:]
        /* Custom code here */
        this.getAddressString = function () {
            return _this.city + " " + _this.number;
        };
        //[end]
    }
    Address.createFrom = function (source) {
        if ('string' === typeof source)
            source = JSON.parse(source);
        var result = new Address();
        result.city = source['city'];
        result.number = source['number'];
        result.country = source['country'];
        return result;
    };
    return Address;
}());
exports.Address = Address;
var PersonalInfo = /** @class */ (function () {
    function PersonalInfo() {
        var _this = this;
        //[PersonalInfo:]
        this.getPersonalInfoString = function () {
            return "pet:" + _this.pet_name;
        };
        //[end]
    }
    PersonalInfo.createFrom = function (source) {
        if ('string' === typeof source)
            source = JSON.parse(source);
        var result = new PersonalInfo();
        result.hobby = source['hobby'];
        result.pet_name = source['pet_name'];
        return result;
    };
    return PersonalInfo;
}());
exports.PersonalInfo = PersonalInfo;
var Person = /** @class */ (function () {
    function Person() {
        var _this = this;
        //[Person:]
        this.getInfo = function () {
            return "name:" + _this.name;
        };
        //[end]
    }
    Person.createFrom = function (source) {
        if ('string' === typeof source)
            source = JSON.parse(source);
        var result = new Person();
        result.name = source['name'];
        result.personal_info = source['personal_info'] ? PersonalInfo.createFrom(source['personal_info']) : null;
        result.nicknames = source['nicknames'];
        result.addresses = source['addresses'] ? source['addresses'].map(function (element) { return Address.createFrom(element); }) : null;
        return result;
    };
    return Person;
}());
exports.Person = Person;
