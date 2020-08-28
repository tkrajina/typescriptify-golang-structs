/* Do not change, this code is generated from Golang structs */


export class Address {
    city: string;
    number: number;
    country?: string;

    static createFrom(source: any = {}) {
        return new Address(source);
    }

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.city = source["city"];
        this.number = source["number"];
        this.country = source["country"];
    }
    //[Address:]
    /* Custom code here */

    getAddressString = () => {
        return this.city + " " + this.number;
    }

    //[end]
}
export class PersonalInfo {
    hobby: string[];
    pet_name: string;

    static createFrom(source: any = {}) {
        return new PersonalInfo(source);
    }

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.hobby = source["hobby"];
        this.pet_name = source["pet_name"];
    }
    //[PersonalInfo:]

    getPersonalInfoString = () => {
        return "pet:" + this.pet_name;
    }

    //[end]
}
export class Person {
    name: string;
    personal_info: PersonalInfo;
    nicknames: string[];
    addresses: Address[];
    address: Address;
    metadata: {[key:string]:string};
    friends: Person[];

    static createFrom(source: any = {}) {
        return new Person(source);
    }

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.name = source["name"];
        this.personal_info = source["personal_info"] && new PersonalInfo(source["personal_info"]);
        this.nicknames = source["nicknames"];
        this.addresses = source["addresses"] && source["addresses"].map((element: any) => new Address(element));
        this.address = source["address"] && new Address(source["address"]);
        this.metadata = source["metadata"];
        this.friends = source["friends"] && source["friends"].map((element: any) => new Person(element));
    }
    //[Person:]

    getInfo = () => {
        return "name:" + this.name;
    }

    //[end]
}