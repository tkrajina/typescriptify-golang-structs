"use strict";
/* Do not change, this code is generated from Golang structs */
Object.defineProperty(exports, "__esModule", { value: true });
exports.Person = exports.PersonalInfo = exports.Address = void 0;
var Address = /** @class */ (function () {
    function Address(source) {
        if (source === void 0) { source = {}; }
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
    return Address;
}());
exports.Address = Address;
var PersonalInfo = /** @class */ (function () {
    function PersonalInfo(source) {
        if (source === void 0) { source = {}; }
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
    return PersonalInfo;
}());
exports.PersonalInfo = PersonalInfo;
var Person = /** @class */ (function () {
    function Person(source) {
        if (source === void 0) { source = {}; }
        var _this = this;
        //[Person:]
        this.getInfo = function () {
            return "name:" + _this.name;
        };
        if ('string' === typeof source)
            source = JSON.parse(source);
        this.name = source["name"];
        this.personal_info = this.convertValues(source["personal_info"], PersonalInfo);
        this.nicknames = source["nicknames"];
        this.addresses = this.convertValues(source["addresses"], Address);
        this.address = this.convertValues(source["address"], Address);
        this.metadata = source["metadata"];
        this.friends = this.convertValues(source["friends"], Person);
    }
    Person.prototype.convertValues = function (a, classs, asMap) {
        var _this = this;
        if (asMap === void 0) { asMap = false; }
        if (!a) {
            return a;
        }
        if (a.slice) {
            return a.map(function (elem) { return _this.convertValues(elem, classs); });
        }
        else if ("object" === typeof a) {
            if (asMap) {
                for (var _i = 0, _a = Object.keys(a); _i < _a.length; _i++) {
                    var key = _a[_i];
                    a[key] = new classs(a[key]);
                }
                return a;
            }
            return new classs(a);
        }
        return a;
    };
    return Person;
}());
exports.Person = Person;
