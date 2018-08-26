"use strict";
/* Do not change, this code is generated from Golang structs */
exports.__esModule = true;
var Address = /** @class */ (function () {
    function Address() {
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
    }
    Person.createFrom = function (source) {
        if ('string' === typeof source)
            source = JSON.parse(source);
        var result = new Person();
        result.name = source['name'];
        result.personal_info = source['personal_info'] ? PersonalInfo.createFrom(source['personal_info']) : null;
        result.nicknames = source['nicknames'];
        result.addresses = source['addresses'] ? source['addresses'].map(function (element) { return Address.createFrom(element); }) : null;
        result.address = source['address'] ? Address.createFrom(source['address']) : null;
        if (source['children']) {
            result.children = {};
            for (var key in source['children'])
                result.children[key] = Person.createFrom(source[key]);
        }
        return result;
    };
    return Person;
}());
exports.Person = Person;
