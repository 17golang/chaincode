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
	ORDER_STATUS_CANINVEST		//订单有效
	ORDER_STATUS_FULL		//订单满额
	ORDER_STATUS_LOAN		//订单放款
	ORDER_STATUS_REFUND		//订单还款
	ORDER_STATUS_FINISHED		//订单完成
	//ORDER_STATUS_CANCEL		//订单取消
)
const (
	ADMIN = iota			//管理员
	SIMPLE_PERSON			//普通用户
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
	Amount float64    `json:"amount"`
	Role   int 	`json:"role"`
}

type Order struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Amount        float64            `json:"amount"`
	Current	      float64	     `json:"current"`
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
	UserName  string `json:"userName"`
	OrderId   string `json:"orderId"`
	Amount    float64    `json:"amount"`
}

type RefundRecord struct {
	ID        string `json:"id"`
	CreatorId string `json:"creatorId"`
	OrderId   string `json:"orderId"`
	Amount    float64    `json:"amount"`
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

	CreateUser("admin","18888888888",20000,ADMIN)

	CreateUser("Randy","18673692416",20000,SIMPLE_PERSON)

	CreateUser("Myra","18673692435",10000,SIMPLE_PERSON)

	return shim.Success(nil)


}


func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Unknown supported call")
}


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
		return t.deleteState(stub, args)
	}
	if args[0] == "query" {
		// queries an entity state
		return t.query(stub, args)
	}
	if args[0] == "createUser"{
		return t.createUser(stub,args)
	}
	if args[0] == "recharge"{
		return t.recharge(stub,args)
	}
	if args[0] == "createOrder"{
		return t.createOrder(stub,args)
	}
	if args[0] == "publish" {
		// Deletes an entity from its state
		return t.publish(stub, args)
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

	return shim.Error("Unknown action, check the first argument, must be one of 'delete', 'query', 'createOrder', 'publish', 'invest', 'loan' or 'refund'")
}


// Deletes an entity from state
func (t *SimpleChaincode) deleteState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
		//非admin账户不能看到CREATE状态下的订单
		var isAdmin bool = false
		user,ok := primaryKeyToUser[param]
		if ok{
			if(user.Role == ADMIN){
				isAdmin = true
			}
		}
		var orderList [] *Order
		for _,v := range creatorToOrder{
			for _,v2 := range v{
				if(!isAdmin && v2.Status == ORDER_STATUS_CREATE){
					continue
				}
				orderList  = append(orderList,v2)
			}
		}
		bytes,err := json.Marshal(orderList)
		if err != nil{
			return shim.Error("find OrderList error !")
		}
		return shim.Success(bytes)
	}
	case "userOrderList" :{
		bytes,err := json.Marshal(creatorToOrder[param])
		if err != nil{
			return shim.Error("find userOrderList error !")
		}
		return shim.Success(bytes)
	}
	case "order" :{
		return t.queryOrder(stub,args)
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
		return shim.Error("support subQuery method is 'userList','user','orderList','userOrderList','order','refundRecord','investRecord'!")
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

func CreateUser(name string,mobile string,amount float64,role int) *User{
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
	user.Role = role
	fmt.Printf("Create User successfully User = %v\n",user)

	primaryKeyToUser[user.ID] = &user
	nameToUser[user.Name] = &user
	return &user
}

func (t *SimpleChaincode) createUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name = args[1]
	var mobile = args[2]
	_,ok :=nameToUser[name]
	if(ok){
		return shim.Error("User name has exist!")
	}
	var user User
	user.ID = getUUID()
	user.Name = name
	user.Mobile = mobile
	user.Amount = 0
	user.Role = SIMPLE_PERSON

	primaryKeyToUser[user.ID] = &user
	nameToUser[user.Name] = &user
	return shim.Success(nil)
}

func (t *SimpleChaincode) recharge(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var userId = args[1]
	amount,err := strconv.ParseFloat(args[2],64)
	if err != nil{
		return shim.Error("recharge amount illegal!")
	}

	user,ok := primaryKeyToUser[userId]
	if !ok{
		shim.Error("User not exist!")
	}
	user.Amount = user.Amount + amount
	nameToUser[user.Name] = &user

	return shim.Success(nil)
}

func (t *SimpleChaincode) createOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var title = args[1]
	var creatorId = args[4]
	var createTime = args[5]
	var endTime = args[6]
	amount,err :=strconv.ParseFloat(args[2],32)
	if(err != nil){
		return shim.Error("param amount must be integer!")
	}
	rate,err := strconv.ParseFloat(args[3],32)
	if(err != nil){
		return shim.Error("param rate parse error,please have a good check!")
	}
	order:=Order{}
	order.ID=getUUID()
	order.Title=title
	order.Amount = amount
	order.Current = 0
	order.Status = ORDER_STATUS_CREATE
	order.Rate = rate
	order.CreatorId = creatorId
	order.CreateTime = createTime
	order.EndTime = endTime
	order.InvestRecords = []InvestRecord{}
	order.RefundRecords = []RefundRecord{}

	subMap := creatorToOrder[creatorId]
	if subMap == nil {
		subMap = map[string]*Order{}
	}
	subMap[order.ID] = &order
	creatorToOrder[creatorId] = subMap

	bytes,err := json.Marshal(&order)
	if(err != nil){
		shim.Error("Json marshal error!")
	}
	stub.PutState(order.ID,bytes)
	byt,err := stub.GetState(order.ID)
	if err != nil{
		return shim.Error("Order is not exist!")
	}

	return shim.Success(byt)
}

func (t *SimpleChaincode) publish(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var orderId = args[1]

	var order1 Order

	bytes,err := stub.GetState(orderId)
	if err != nil{
		return shim.Error("Order is not exist!")
	}
	json.Unmarshal(bytes,&order1)
	order := &order1
	if order == nil {
		return shim.Error("Order parse error !")
	}
	if order.Status != ORDER_STATUS_CREATE{
		return shim.Error("Order can't publish! status not create!")
	}
	order.Status = ORDER_STATUS_CANINVEST
	b,err := json.Marshal(order)
	if(err != nil){
		shim.Error("Json marshal error!")
	}
	stub.PutState(order.ID,b)

	var tempOrder *Order
	for _,m := range creatorToOrder{
		for  _,o := range m   {
			if(orderId == o.ID){
				tempOrder = o
				break
			}
		}
	}
	tempOrder.Status = order.Status
	return shim.Success(nil)
}

func (t *SimpleChaincode) invest(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var orderId = args[1]
	var creatorId = args[2]
	amount,err :=strconv.ParseFloat(args[3],32)
	if(err != nil){
		return shim.Error("param amount must be float!")
	}

	var order1 Order

	bytes,err := stub.GetState(orderId)
	if err != nil{
		return shim.Error("Order is not exist!")
	}
	json.Unmarshal(bytes,&order1)
	order := &order1
	if order == nil {
		return shim.Error("Order parse error !")
	}
	if order.Status != ORDER_STATUS_CANINVEST{
		return shim.Error("Order can't invest! status can not invest")
	}
	if(order.Amount < amount){
		return shim.Error("Invest amount can't greater than order left amount!")
	}
	//用户扣款
	var user *User
	user2,ok := primaryKeyToUser[creatorId]
	if ok{
		user = user2
	}
	if  user == nil{
		return shim.Error("User can't invest!")
	}
	if user.Amount < amount {
		return shim.Error("User has not enough money!")
	}
	//remove "amount" from user account
	if user.Amount < amount{
		return shim.Error("balance is not enough !")
	}
	user.Amount = user.Amount - amount

	var investRecord InvestRecord
	investRecord.ID = getUUID()
	investRecord.CreatorId = creatorId
	investRecord.Amount = amount
	investRecord.OrderId = order.ID
	investRecord.UserName = user.Name

	order.InvestRecords = append(order.InvestRecords,investRecord)
	order.Current += amount

	creatorToInvestRecord[creatorId] = append(creatorToInvestRecord[creatorId],&investRecord)

	//投资金额达到目标,修改订单状态为投资满额
	if order.Amount == order.Current{
		order.Status = ORDER_STATUS_FULL
	}

	b,err := json.Marshal(order)
	if(err != nil){
		shim.Error("Json marshal error!")
	}
	stub.PutState(order.ID,b)

	var tempOrder *Order
	for _,m := range creatorToOrder{
		for  _,o := range m   {
			if(orderId == o.ID){
				tempOrder = o
				break
			}
		}
	}
	tempOrder.Current = order.Current
	tempOrder.Status  = order.Status
	tempOrder.InvestRecords = order.InvestRecords

	return shim.Success(nil)
}

func (t *SimpleChaincode) loan(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var orderId = args[1]
	var order1 Order

	bytes,err := stub.GetState(orderId)
	if err != nil{
		return shim.Error("Order is not exist!")
	}
	json.Unmarshal(bytes,&order1)
	order := &order1
	if order.Status != ORDER_STATUS_FULL || order.Current != order.Amount{
		return shim.Error("Order amount is not enough !")
	}
	//放款
	user := primaryKeyToUser[order.CreatorId]
	user.Amount = user.Amount + order.Current

	//订单存放的投资金额清零
	order.Current = 0

	//订单状态改为已放款
	order.Status = ORDER_STATUS_LOAN

	b,err := json.Marshal(order)
	if(err != nil){
		shim.Error("Json marshal error!")
	}
	stub.PutState(order.ID,b)

	var tempOrder *Order
	for _,m := range creatorToOrder{
		for  _,o := range m   {
			if(orderId == o.ID){
				tempOrder = o
				break
			}
		}
	}
	tempOrder.Current = order.Current
	tempOrder.Status  = order.Status

	return shim.Success(nil)
}

func (t *SimpleChaincode)refund(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var orderId = args[1]

	var order1 Order

	bytes,err := stub.GetState(orderId)
	if err != nil{
		return shim.Error("Order is not exist!")
	}
	json.Unmarshal(bytes,&order1)
	order := &order1
	if order == nil{
		return shim.Error("Order is not exist !")
	}
	if order.Status != ORDER_STATUS_LOAN {
		return shim.Error("Order can't refund current status is not loan!")
	}
	user := primaryKeyToUser[order.CreatorId]
	if(user == nil){
		return shim.Error("User is not exist !")
	}
	var receiveAmount float64
	for _,v := range order.InvestRecords  {
		receiveAmount += v.Amount * (1+order.Rate)
	}
	if user.Amount < receiveAmount{
		return shim.Error("User balance is not enough !")
	}

	for _,v := range order.InvestRecords  {
		u := primaryKeyToUser[v.CreatorId]

		refundRecord := RefundRecord{}
		refundRecord.ID = getUUID()
		refundRecord.CreatorId = v.CreatorId
		var realAmount = v.Amount * (1+order.Rate)
		refundRecord.Amount = realAmount
		refundRecord.OrderId = v.OrderId

		if order.RefundRecords == nil{
			order.RefundRecords = []RefundRecord{}
		}
		order.RefundRecords = append(order.RefundRecords,refundRecord)
		u.Amount = u.Amount + realAmount
		user.Amount = user.Amount - realAmount

		creatorToRefundRecord[refundRecord.CreatorId] = append(creatorToRefundRecord[refundRecord.CreatorId],&refundRecord)
	}
	order.Status = ORDER_STATUS_REFUND

	var sum float64
	for _,v := range order.RefundRecords  {
		sum += v.Amount
	}
	if order.Amount <= sum{
		order.Status = ORDER_STATUS_FINISHED
	}

	b,err := json.Marshal(order)
	if(err != nil){
		shim.Error("Json marshal error!")
	}
	stub.PutState(order.ID,b)

	var tempOrder *Order
	for _,m := range creatorToOrder{
		for  _,o := range m   {
			if(orderId == o.ID){
				tempOrder = o
				break
			}
		}
	}
	tempOrder.Status  = order.Status
	tempOrder.RefundRecords = order.RefundRecords

	return shim.Success(nil)
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

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
