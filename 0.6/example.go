/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"
	"encoding/base64"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// This chaincode implements a simple map that is stored in the state.
// The following operations are available.

// Invoke operations
// put - requires two arguments, a key and value
// remove - requires a key

// Query operations
// get - requires one argument, a key, and returns a value
// keys - requires no arguments, returns all keys

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

const (  
    base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"  
)  
  
var coder = base64.NewEncoding(base64Table)

// Init is a no-op
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return nil, nil
}

// Invoke has two functions
// put - takes two arguements, a key and value, and stores them in the state
// remove - takes one argument, a key, and removes if from the state
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	switch function {
	case "write":
		id:=args[0]
		json:=args[1]


		//jsonByte,_:=coder.DecodeString(string(json))
		err:=stub.PutState(string(id),json)
		if err!=nil{
			return nil,fmt.Errorf("put fund error %s",err)
		}

		return nil,nil


	default:
		return nil, errors.New("Unsupported operation")
	}
}

// Query has two functions
// get - takes one argument, a key, and returns the value for the key
// keys - returns all keys stored in this chaincode
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	switch function {

	case "queryFund":
		if len(args) < 1 {
			return nil, errors.New("get operation must include one argument, a key")
		}
		key := args[0]
		value, err := stub.GetState(key)
		if err != nil {
			return nil, fmt.Errorf("get operation failed. Error accessing state: %s", err)
		}
		return value, nil

	case "queryFunds":
		var json="["
		ids:=args[0]
		idArray := strings.Split(string(ids), "##")
		for _, v := range idArray { 
			value, _ := stub.GetState(string(v))
			json += string(value)	
		}
		json+="]"

		return []byte(json) ,nil

	default:
		return nil, errors.New("Unsupported operation")
	}
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
