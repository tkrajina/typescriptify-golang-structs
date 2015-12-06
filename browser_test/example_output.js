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
    Address.fromJSON = function (json) {
        var result = new Address();
        result.city = json["city"];
        result.number = json["number"];
        result.country = json["country"];
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
    PersonalInfo.fromJSON = function (json) {
        var result = new PersonalInfo();
        result.hobby = json["hobby"];
        result.pet_name = json["pet_name"];
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
    Person.fromJSON = function (json) {
        var result = new Person();
        result.name = json["name"];
        result.personal_info = PersonalInfo.fromJSON(json["personal_info"]);
        result.nicknames = json["nicknames"];
        if (json["addresses"]) {
            result.addresses = json["addresses"].map(function (element) { return Address.fromJSON(element); });
        }
        return result;
    };
    return Person;
})();
