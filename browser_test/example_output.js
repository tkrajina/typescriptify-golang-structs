/* Do not change, this code is generated from Golang structs */
var Address = (function () {
    function Address() {
        var _this = this;
        //[Address:]
        /* Custom code here */
        this.getAddressString = function () {
            return _this.city + " " + _this.number;
        };
    }
    Address.createFrom = function (source) {
        var result = new Address();
        result.city = source["city"];
        result.number = source["number"];
        result.country = source["country"];
        return result;
    };
    return Address;
})();
var PersonalInfo = (function () {
    function PersonalInfo() {
        var _this = this;
        //[PersonalInfo:]
        this.getPersonalInfoString = function () {
            return "pet:" + _this.pet_name;
        };
    }
    PersonalInfo.createFrom = function (source) {
        var result = new PersonalInfo();
        result.hobby = source["hobby"];
        result.pet_name = source["pet_name"];
        return result;
    };
    return PersonalInfo;
})();
var Person = (function () {
    function Person() {
        var _this = this;
        //[Person:]
        this.getInfo = function () {
            return "name:" + _this.name;
        };
    }
    Person.createFrom = function (source) {
        var result = new Person();
        result.name = source["name"];
        result.personal_info = source["personal_info"] ? PersonalInfo.createFrom(source["personal_info"]) : null;
        result.nicknames = source["nicknames"];
        result.addresses = source["addresses"] ? source["addresses"].map(function (element) { return Address.createFrom(element); }) : null;
        return result;
    };
    return Person;
})();
