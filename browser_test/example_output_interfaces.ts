/* Do not change, this code is generated from Golang structs */


export interface Address {
    city: string;
    number: number;
    country?: string;
    //[Address:]
    /* Custom code here */

    [key: string]: any

    //[end]
}
export interface PersonalInfo {
    hobby: string[];
    pet_name: string;
    //[PersonalInfo:]
    /* Custom code here */

    [key: string]: any

    //[end]
}
export interface Person {
    name: string;
    personal_info: PersonalInfo;
    nicknames: string[];
    addresses: Address[];
    address: Address;
    metadata: {[key:string]:string};
    friends: Person[];
    //[Person:]
    /* Custom code here */

    [key: string]: any

    //[end]
}