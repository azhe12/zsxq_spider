package main

import (
	"encoding/json"
	"fmt"
)

type Cookie struct {
	domain string `json:"domain"`
	expirationDate float32 `json:"expirationDate"`
	hostOnly bool `json:"hostOnly"`
	path string `json:"path"`
	sameSite string `json:"sameSite"`
	secure bool `json:"secure"`
	session bool `json:"session"`
	storeId string `json:"storeId"`
	id int `json:"id"`
	name string `json:"name"`
	value string `json:"value"`
}

//func main() {
//	cookieStr := `[
//{
//    "domain": ".zsxq.com",
//    "expirationDate": 1568431615.21024,
//    "hostOnly": false,
//    "httpOnly": false,
//    "name": "abtest_env",
//    "path": "/",
//    "sameSite": "unspecified",
//    "secure": false,
//    "session": false,
//    "storeId": "0",
//    "value": "product",
//    "id": 1
//},
//{
//    "domain": ".zsxq.com",
//    "expirationDate": 1583915272,
//    "hostOnly": false,
//    "httpOnly": false,
//    "name": "UM_distinctid",
//    "path": "/",
//    "sameSite": "unspecified",
//    "secure": false,
//    "session": false,
//    "storeId": "0",
//    "value": "16d1f6fa00eec0-0d8bbe2bdcc961-38637501-1aeaa0-16d1f6fa00f435",
//    "id": 2
//}]`
//	cookies := []Cookie{}
//	json.Unmarshal([]byte(cookieStr), cookies)
//
//	fmt.Printf("cookies %v", cookies)
//
//}


func main ( ) {
	var jsonBlob = [ ] byte ( ` [
        { "Name" : "Platypus" , "Order" : "Monotremata" } ,
        { "Name" : "Quoll" ,     "Order" : "Dasyuromorphia" }
    ] ` )
	type Animal struct {
		//Name  string
		Order string
	}
	var animals [ ] Animal
	err := json.Unmarshal ( jsonBlob , & animals )
	if err != nil {
		fmt.Println ( "error:" , err )
	}
	fmt.Printf ( "%+v" , animals )
}