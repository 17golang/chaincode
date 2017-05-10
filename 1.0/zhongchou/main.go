package main

import (
	_"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"strings"
	_"errors"
	"time"
	_"reflect"
	"encoding/json"
)

//存放用户信息
var primaryKeyToUser = map[string]User{}

//存放用户众筹信息
var creatorToOrder = map[string]map[string]Order{}

//存放用户投资信息
//var creatorToInvestRecord = map[string]InvestRecord{}

//存放用户还款信息
//var creatorToRefundRecord = map[string]RefundRecord{}

//var m = map[string]interface{}{}

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

func init() {
	//初始化数据

	//User1 := User{"user_1", "Randy", "18673692416", 0}
	//User2 := User{"user_2", "Myra", "18673692222", 0}
	//
	//primaryKeyToUser[User1.ID] = User1
	//primaryKeyToUser[User2.ID] = User2
	//
	//
	//order1 := Order{getUUID(), "众筹_1", 10000, 0, 0.05, "user_1", "2017-05-08 15:52:00", "2017-05-31 23:59;59", []InvestRecord{}, []RefundRecord{}}
	//order2 := Order{getUUID(), "众筹_2", 10000, 0, 0.06, "user_2", "2017-05-08 15:52:00", "2017-05-31 23:59;59", []InvestRecord{}, []RefundRecord{}}
	//creatorToOrder[order1.CreatorId] = append(creatorToOrder[order1.CreatorId], order1)
	//creatorToOrder[order2.CreatorId] = append(creatorToOrder[order2.CreatorId], order2)
	//creatorToOrder[order1.CreatorId] = append(creatorToOrder[order1.CreatorId], order1)
	//submap := creatorToOrder[order1.]

	//order1.

	//o1, _ := json.Marshal(order1)
	//m["order_1"] = o1
	//o2, _ := json.Marshal(order2)
	//m["order_2"] = o2

}

func getUUID() string{
	return strings.Replace(uuid.NewV4().String(),"-","",-1)
}

func CreateUser(name string,mobile string,amount int) User{
	_,ok :=primaryKeyToUser[name]
	if(ok){
		fmt.Println("User is exist!")
		//return nil
	}
	user := User{getUUID(),name,mobile,amount}

	fmt.Printf("Create User successfully User = %v\n",user)

	primaryKeyToUser[user.ID] = user
	return user
}

func CreateOrder(title string,amount int,rate float64,creatorId string) Order{
	order:=Order{}
	order.ID=getUUID()
	order.Title=title
	order.Amount =amount
	order.Current = 0
	order.Status = 0
	order.Rate = rate
	order.CreatorId = creatorId
	order.CreateTime = time.Now().String()
	order.EndTime = time.Now().String()
	order.InvestRecords = []InvestRecord{}
	order.RefundRecords = []RefundRecord{}
	fmt.Printf("Create Order successfully Order = %v\n",order)

	submap :=creatorToOrder[creatorId]
	if submap == nil {
		submap = map[string]Order{}
	}
	submap[order.ID] = order
	creatorToOrder[creatorId] = submap

	return order
}

func Invset(orderId string,creatorId string,amount int){
	var order *Order
	for _,v:= range creatorToOrder{
		for _,v:=range v {
			if(orderId == v.ID){
				order = &v
				break;
			}
		}
	}
	if order == nil {
		fmt.Println("Order is not exist!")
		return
	}
	if order.Status == 1{
		fmt.Println("Order has full of amount!")
	}
	//用户扣款
	user := primaryKeyToUser[creatorId]
	if  *user ==nil{
		fmt.Println("User is not exist!")
	}
	if user.Amount < amount {
		fmt.Printf("用户:%v 余额不足",user)
		return
	}

	//
	var investRecord InvestRecord
	investRecord.ID = getUUID()
	investRecord.CreatorId = creatorId
	investRecord.Amount = amount
	investRecord.OrderId = order.ID

	order.InvestRecords = append(order.InvestRecords,investRecord)
	order.Current += amount

	//投资金额达到目标,修改订单状态为投资满额
	fmt.Printf("hasInvestAmount=%v,orderAmount=%d\n",order.Current,order.Amount)
	if order.Amount == order.Current{
		order.Status = 1
	}
	for _,v:= range creatorToOrder{
		for _,o:=range v {
			if(orderId == o.ID){
				v[orderId] = *order
				break;
			}
		}
	}

}

func main() {
	user1 := CreateUser("Randy","18673692416",0)
	user2 := CreateUser("Myra","18673692435",0)
	CreateUser("Randy","18673692416",0)

	o1 := CreateOrder("众筹_1",10000,0.05,user1.ID)

	o2 := CreateOrder("众筹_2",10000,0.05,user2.ID)
	//fmt.Printf("o1=%d,o2=%d",o1,o2)

	Invset(o1.ID,user1.ID,10000)


	for _,v:= range creatorToOrder{
		//v := map[string]Order{}(v)
		//fmt.Println(reflect.TypeOf(v).String())
		//fmt.Println(v)
		for k,v:=range v {
			b,_ :=json.Marshal(v)
			if(o1.ID == v.ID){
				fmt.Printf( " k = %v,v=%v\n",k,string(b))
			}
			if(o2.ID == v.ID){
				fmt.Printf( " k2 = %v,v=%v\n",k,v)
			}
		}
	}


}
