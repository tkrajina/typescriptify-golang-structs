/* Do not change, this code is generated from Golang structs */


export interface Address {
    city: string;
    number: number;
    country?: string;
}
export interface PersonalInfo {
    hobby: string[];
    pet_name: string;
}
export interface Person {
    name: string;
    personal_info: PersonalInfo;
    nicknames: string[];
    addresses: Address[];
    address?: Address;
    metadata: {[key:string]:string};
    friends: Person[];
    //[Person:]
    /* Custom code here */

    [key: string]: any

    //[end]
}