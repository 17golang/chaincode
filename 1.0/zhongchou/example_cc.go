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
	"fmt"
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
	"time"
	"math/rand"
)

const (
	ORDER_STATUS_CREATE = iota 	//订单创建
	ORDER_STATUS_FULL		//订单满额
	ORDER_STATUS_LOAD		//订单放款
	ORDER_STATUS_REFUND		//订单还款
	ORDER_STATUS_FINISHED		//订单完成
	ORDER_STATUS_CANCEL		//订单取消
)

//存放用户信息
var primaryKeyToUser = map[string]*User{}

var nameToUser = map[string]*User{}

//存放用户众筹信息
var creatorToOrder = map[string]map[string]*Order{}

//存放用户投资信息
var creatorToInvestRecord = map[string][]*InvestRecord{}

//存放用户还款信息
var creatorToRefundRecord = map[string][]*RefundRecord{}


type User struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Mobile string `json:"mobile"`
	Amount int    `json:"amount"`
}

type Order struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Amount        int            `json:"amount"`
	Current	      int	     `json:"current"`
	Status        int            `json:"status"`
	Rate          float64        `json:"rate"`
	CreatorId     string         `json:"creatorId"`
	CreateTime    string         `json:"createTime"`
	EndTime       string         `json:"endTime"`
	InvestRecords []InvestRecord `json:"investRecords"`
	RefundRecords []RefundRecord `json:"refundRecords"`
}

type InvestRecord struct {
	ID        string `json:"id"`
	CreatorId string `json:"creatorId"`
	OrderId   string `json:"orderId"`
	Amount    int    `json:"amount"`
}

type RefundRecord struct {
	ID        string `json:"id"`
	CreatorId string `json:"creatorId"`
	OrderId   string `json:"orderId"`
	Amount    int    `json:"amount"`
}


// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	fmt.Println("########### example_cc Init ###########")
	_, args := stub.GetFunctionAndParameters()
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}


	user1 := CreateUser("Randy","18673692416",20000)

	user2 :=CreateUser("Myra","18673692435",0)
	CreateUser("Randy","18673692416",0)

	o1 := CreateOrder("众筹_1",10000,0.05,user2.ID)

	//o2 := CreateOrder("众筹_2",10000,0.05,user2.ID)
	//fmt.Printf("o1=%d,o2=%d",o1,o2)
	fmt.Printf("投资前用户的账户信息:%v--->%v\n",primaryKeyToUser[user1.ID],primaryKeyToUser[user2.ID])
	Invest(o1.ID,user1.ID,10000)
	go Invest(o1.ID,user1.ID,10000)
	go Invest(o1.ID,user1.ID,10000)
	go Invest(o1.ID,user1.ID,10000)
	go Invest(o1.ID,user1.ID,10000)

	//user :=primaryKeyToUser[user1.ID]
	fmt.Printf("投资后用户的账户信息:%v--->%v\n",primaryKeyToUser[user1.ID],primaryKeyToUser[user2.ID])

	loan(o1.ID)
	fmt.Printf("放款后用户的账户信息:%v--->%v\n",primaryKeyToUser[user1.ID],primaryKeyToUser[user2.ID])

	refund(o1.ID)
	fmt.Printf("还款后用户的账户信息:%v--->%v\n",primaryKeyToUser[user1.ID],primaryKeyToUser[user2.ID])
	b,_ :=json.Marshal(creatorToOrder[user2.ID])
	fmt.Printf("订单信息:%v\n",string(b))

	bytes,_ := json.Marshal(o1)
	stub.PutState(o1.ID,bytes)
	return shim.Success(nil)


}



func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Unknown supported call")
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### example_cc Invoke ###########")
	function, args := stub.GetFunctionAndParameters()

	if function != "invoke" {
		return shim.Error("Unknown function call")
	}

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting at least 2")
	}

	if args[0] == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	}

	if args[0] == "query" {
		// queries an entity state
		return t.query(stub, args)
	}
	if args[0] == "move" {
		// Deletes an entity from its state
		return t.move(stub, args)
	}

	if args[0] == "createOrder"{
		return t.createOrder(stub,args)
	}
	if args[0] == "invest" {
		return t.invest(stub, args)
	}
	if args[0] == "loan" {
		// Deletes an entity from its state
		return t.loan(stub, args)
	}
	if args[0] == "refund" {
		return t.refund(stub, args)
	}

	if args[0] == "investOrder" {
		// Deletes an entity from its state
		return t.investOrder(stub, args)
	}
	if args[0] == "queryOrder" {
		// Deletes an entity from its state
		return t.queryOrder(stub, args)
	}

	return shim.Error("Unknown action, check the first argument, must be one of 'delete', 'query', 'refund', or 'move'")
}


func (t *SimpleChaincode) move(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// must be an invoke
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4, function followed by 2 names and 1 value")
	}

	A = args[1]
	B = args[2]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil);
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[1]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	// if len(args) != 2 {
	// 	return shim.Error("Incorrect number of arguments. Expecting 3 params")
	// }
	var subQueryMethod string = args[1]
	var param = args[2]
	switch subQueryMethod {
		case "userList" :{
			var userList [] *User
			for _,v := range primaryKeyToUser{

							userList  = append(userList,v)

			}
			bytes,err := json.Marshal(userList)
			if err != nil{
				return shim.Error("find UserList error !")
			}
			return shim.Success(bytes)
		}
		case "user" :{
			bytes,err := json.Marshal(primaryKeyToUser[param])
			if err != nil{
				return shim.Error("find User error !")
			}
			return shim.Success(bytes)
		}
		case "orderList" :{
		  var orderList [] *Order
			for _,v := range creatorToOrder{
					for _,v2 := range v{
							orderList  = append(orderList,v2)
					}
			}
			bytes,err := json.Marshal(orderList)
			if err != nil{
				return shim.Error("find OrderList error !")
			}
			return shim.Success(bytes)
		}
		case "Order" :{
			var order *Order
			for _,v := range creatorToOrder{
					for _,v2 := range v{
							if param == v2.ID{
								  order = v2
									break
							}
					}
			}
			bytes,err := json.Marshal(order)
			if err != nil{
				return shim.Error("find User error !")
			}
			return shim.Success(bytes)
		}
		case "investRecord" :{
			bytes,err := json.Marshal(creatorToInvestRecord[param])
			if err != nil{
				return shim.Error("find investRecord error !")
			}
			return shim.Success(bytes)
		}
	case "refundRecord" :{
		bytes,err := json.Marshal(creatorToRefundRecord[param])
		if err != nil{
			return shim.Error("find refundRecord error !")
		}
		return shim.Success(bytes)
	}
	default:{
		return shim.Error("support subQuery method is 'userList','user','orderList','order','refundRecord','investRecord'!")
	}

	}
}

//随机字符串
func getUUID() string {
	var size = 32
	var kind = 3
	ikind, kinds, result := kind, [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}, make([]byte, size)
	is_all := kind > 2 || kind < 0
	rand.Seed(time.Now().UnixNano())
	for i :=0; i < size; i++ {
		if is_all { // random ikind
			ikind = rand.Intn(3)
		}
		scope, base := kinds[ikind][0], kinds[ikind][1]
		result[i] = uint8(base+rand.Intn(scope))
	}
	return string(result)
}

func CreateUser(name string,mobile string,amount int) *User{
	_,ok :=nameToUser[name]
	if(ok){
		fmt.Println("User is exist!")
		return nil
	}
	var user User
	user.ID = getUUID()
	user.Name = name
	user.Mobile = mobile
	user.Amount = amount

	fmt.Printf("Create User successfully User = %v\n",user)

	primaryKeyToUser[user.ID] = &user
	nameToUser[user.Name] = &user
	return &user
}

func CreateOrder(title string,amount int,rate float64,creatorId string) Order{
	order:=Order{}
	order.ID=getUUID()
	order.Title=title
	order.Amount =amount
	order.Current = 0
	order.Status = ORDER_STATUS_CREATE
	order.Rate = rate
	order.CreatorId = creatorId
	order.CreateTime = time.Now().String()
	order.EndTime = time.Now().String()
	order.InvestRecords = []InvestRecord{}
	order.RefundRecords = []RefundRecord{}
	fmt.Printf("Create Order successfully Order = %v\n",order)

	submap :=creatorToOrder[creatorId]
	if submap == nil {
		submap = map[string]*Order{}
	}
	submap[order.ID] = &order
	creatorToOrder[creatorId] = submap

	return order
}


func (t *SimpleChaincode) createOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var creatorId = args[4]
	amount,_ :=strconv.Atoi(args[2])
	order:=Order{}
	order.ID=getUUID()
	order.Title=args[1]
	order.Amount = amount
	order.Current = 0
	order.Status = ORDER_STATUS_CREATE
	order.Rate = 0.05
	order.CreatorId = args[4]
	order.CreateTime = time.Now().String()
	order.EndTime = time.Now().String()
	order.InvestRecords = []InvestRecord{}
	order.RefundRecords = []RefundRecord{}
	fmt.Printf("Create Order successfully Order = %v\n",order)

	submap :=creatorToOrder[creatorId]
	if submap == nil {
		submap = map[string]*Order{}
	}
	submap[order.ID] = &order
	creatorToOrder[creatorId] = submap


	bytes,_ := json.Marshal(order)

	stub.PutState(order.ID,bytes)

	return shim.Success(nil)
}

func (t *SimpleChaincode) invest(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var orderId = args[1]
	var creatorId = args[2]
	amount,_ :=strconv.Atoi(args[3])
	var order *Order
	for _,v:= range creatorToOrder{
		for _,v:=range v {
			if(orderId == v.ID){
				order = v
				break;
			}
		}
	}
	if order == nil {
		fmt.Println("Order is not exist!")
		return shim.Error("Order is not exist !")
	}
	if order.Status != ORDER_STATUS_CREATE{
		fmt.Println("Order can't invest!")
		return shim.Error("Order can't invest!")
	}
	//用户扣款
	var user *User
	user2,ok := primaryKeyToUser[creatorId]
	if ok{
		user = user2
	}

	if  user == nil{
		fmt.Println("User is not exist!")
		return shim.Error("Order can't invest!")
	}
	if user.Amount < amount {
		fmt.Printf("User %v hasn't enough money!",user.Name)
		return shim.Error("Order can't invest!")
	}
	//remove "amount" from user account
	user.Amount = user.Amount - amount

	//
	var investRecord InvestRecord
	investRecord.ID = getUUID()
	investRecord.CreatorId = creatorId
	investRecord.Amount = amount
	investRecord.OrderId = order.ID

	order.InvestRecords = append(order.InvestRecords,investRecord)
	order.Current += amount

	creatorToInvestRecord[creatorId] = append(creatorToInvestRecord[creatorId],&investRecord)

	//投资金额达到目标,修改订单状态为投资满额
	if order.Amount == order.Current{
		order.Status = ORDER_STATUS_FULL
	}

	fmt.Printf("Invest successfully! user : %v invest %v and balance is %v\n",user.Name,amount,user.Amount)
	return shim.Success(nil)
}

func (t *SimpleChaincode) loan(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var orderid = args[1]
	var order *Order
	for _,v := range creatorToOrder{
		for  _,v2 := range v   {
			if(orderid == v2.ID){
				order = v2
				break
			}
		}
	}
	if order == nil{
		fmt.Println("Order is not exist !")
			return shim.Error("Order is not exist !")
	}
	if order.Status != 1 || order.Current != order.Amount{
		fmt.Println("Order amount is not enough")
		return shim.Error("Order amount is not enough !")
	}
	user := primaryKeyToUser[order.CreatorId]
	user.Amount = user.Amount + order.Amount

	order.Status = ORDER_STATUS_LOAD
	return shim.Success(nil)
}

func (t *SimpleChaincode)refund(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var orderid = args[1]
	var order *Order
	for _,v := range creatorToOrder{
		for  _,v2 := range v   {
			if(orderid == v2.ID){
				order = v2
				break
			}
		}
	}
	if order == nil{
		fmt.Println("Order is not exist !")
		return shim.Error("Order is not exist !")
	}
	if order.Status != 2 {
		fmt.Println("Order isn't loan!")
		return shim.Error("Order is not exist !")
	}
	user := primaryKeyToUser[order.CreatorId]
	if(user == nil){
		fmt.Println("User is not exist!")
		return shim.Error("Order is not exist !")
	}
	for _,v := range order.InvestRecords  {
		u := primaryKeyToUser[v.CreatorId]

		refundRecord := RefundRecord{}
		refundRecord.ID = getUUID()
		refundRecord.CreatorId = v.CreatorId
		refundRecord.Amount = v.Amount
		refundRecord.OrderId = v.OrderId

		if order.RefundRecords == nil{
			order.RefundRecords = []RefundRecord{}
		}
		order.RefundRecords = append(order.RefundRecords,refundRecord)
		u.Amount = u.Amount + v.Amount
		user.Amount = user.Amount - v.Amount

		creatorToRefundRecord[refundRecord.CreatorId] = append(creatorToRefundRecord[refundRecord.CreatorId],&refundRecord)
	}
	order.Status = ORDER_STATUS_REFUND
		return shim.Success(nil)
}

/**
项目认筹
create by wupeng
 */
func (t *SimpleChaincode) investOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Print("执行saveEntity方法: " )
	fmt.Println(args)
	var key, value string
	var err error
	var bytes []byte
	key = args[1]
	value = args[2]

	//声明结构体接收json参数
	var data  InvestRecord
	err = json.Unmarshal([]byte(value), &data)
	if err != nil {
		return shim.Error("json参数转struct失败")
	}

	bytes, err = stub.GetState(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	var order Order
	err = json.Unmarshal(bytes, &order)
	if err != nil {
		return shim.Error("GetState转struct失败--investOrder")
	}

	investRecords := order.InvestRecords
	investRecords = append(investRecords, data)
	order.InvestRecords = investRecords

	bytes, err = json.Marshal(order)
	if err != nil {
		return shim.Error("order转json失败")
	}

	err = stub.PutState(key, bytes)
	if err != nil {
		return shim.Error(err.Error())
	}

        return shim.Success(nil);
}


/**
项目认筹记录查询
create by wupeng
 */
func (t *SimpleChaincode) queryOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Print("执行queryEntity方法: " )
	fmt.Println(args)

	var key string
	key = args[1]

	orderBytes, err := stub.GetState(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	var order Order
	err1 := json.Unmarshal(orderBytes, &order)
	if err1 != nil {
		return shim.Error("GetState转struct失败--queryOrder")
	}

	bytes, err2 := json.Marshal(order)
	if err2 != nil {
		return shim.Error("order转json失败")
	}

	return shim.Success(bytes)
}



func Invest(orderId string,creatorId string,amount int){
	var order *Order
	for _,v:= range creatorToOrder{
		for _,v:=range v {
			if(orderId == v.ID){
				order = v
				break;
			}
		}
	}
	if order == nil {
		fmt.Println("Order is not exist!")
		return
	}
	if order.Status != ORDER_STATUS_CREATE{
		fmt.Println("Order can't invest!")
		return
	}
	//用户扣款
	var user *User
	user2,ok := primaryKeyToUser[creatorId]
	if ok{
		user = user2
	}

	if  user == nil{
		fmt.Println("User is not exist!")
		return
	}
	if user.Amount < amount {
		fmt.Printf("User %v hasn't enough money!",user.Name)
		return
	}
	//remove "amount" from user account
	user.Amount = user.Amount - amount

	//
	var investRecord InvestRecord
	investRecord.ID = getUUID()
	investRecord.CreatorId = creatorId
	investRecord.Amount = amount
	investRecord.OrderId = order.ID

	order.InvestRecords = append(order.InvestRecords,investRecord)
	order.Current += amount

	creatorToInvestRecord[creatorId] = append(creatorToInvestRecord[creatorId],&investRecord)

	//投资金额达到目标,修改订单状态为投资满额
	if order.Amount == order.Current{
		order.Status = ORDER_STATUS_FULL
	}

	fmt.Printf("Invest successfully! user : %v invest %v and balance is %v\n",user.Name,amount,user.Amount)
}

func loan(orderid string){
	var order *Order
	for _,v := range creatorToOrder{
		for  _,v2 := range v   {
			if(orderid == v2.ID){
				order = v2
				break
			}
		}
	}
	if order == nil{
		fmt.Println("Order is not exist !")
		return
	}
	if order.Status != 1 || order.Current != order.Amount{
		fmt.Println("Order amount is not enough")
		return
	}
	user := primaryKeyToUser[order.CreatorId]
	user.Amount = user.Amount + order.Amount

	order.Status = ORDER_STATUS_LOAD
}

func refund(orderid string){
	var order *Order
	for _,v := range creatorToOrder{
		for  _,v2 := range v   {
			if(orderid == v2.ID){
				order = v2
				break
			}
		}
	}
	if order == nil{
		fmt.Println("Order is not exist !")
		return
	}
	if order.Status != 2 {
		fmt.Println("Order isn't loan!")
		return
	}
	user := primaryKeyToUser[order.CreatorId]
	if(user == nil){
		fmt.Println("User is not exist!")
		return
	}
	for _,v := range order.InvestRecords  {
		u := primaryKeyToUser[v.CreatorId]

		refundRecord := RefundRecord{}
		refundRecord.ID = getUUID()
		refundRecord.CreatorId = v.CreatorId
		refundRecord.Amount = v.Amount
		refundRecord.OrderId = v.OrderId

		if order.RefundRecords == nil{
			order.RefundRecords = []RefundRecord{}
		}
		order.RefundRecords = append(order.RefundRecords,refundRecord)
		u.Amount = u.Amount + v.Amount
		user.Amount = user.Amount - v.Amount

		creatorToRefundRecord[refundRecord.CreatorId] = append(creatorToRefundRecord[refundRecord.CreatorId],&refundRecord)
	}
	order.Status = ORDER_STATUS_REFUND

}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
