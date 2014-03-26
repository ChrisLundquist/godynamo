package main

import (
	"fmt"
	"encoding/json"
	ep "github.com/smugmug/godynamo/endpoint"
	query "github.com/smugmug/godynamo/endpoints/query"
	conf_iam "github.com/smugmug/godynamo/conf_iam"
	"github.com/smugmug/godynamo/conf"
	"github.com/smugmug/godynamo/conf_file"
	"log"
	keepalive "github.com/smugmug/godynamo/keepalive"
)

func main() {
	conf_file.Read()
	conf.Vals.ConfLock.RLock()
	if conf.Vals.Initialized == false {
		panic("the conf.Vals global conf struct has not been initialized")
	}

	// launch a background poller to keep conns to aws alive
	if conf.Vals.Network.DynamoDB.KeepAlive {
		log.Printf("launching background keepalive")
		go keepalive.KeepAlive([]string{})
	}

	// deal with iam, or not
	if conf.Vals.UseIAM {
		iam_ready_chan := make(chan bool)
		go conf_iam.GoIAM(iam_ready_chan)
		_ = <- iam_ready_chan
	}
	conf.Vals.ConfLock.RUnlock()

	tn := "test-godynamo-livetest"
	q := query.NewQuery()
	q.TableName = tn
	q.Select = ep.SELECT_ALL
	k_v1 := fmt.Sprintf("AHashKey%d",100)
	var kc query.KeyCondition
	kc.AttributeValueList = make([]ep.AttributeValue,1)
	kc.AttributeValueList[0] = ep.AttributeValue{S:k_v1}
	kc.ComparisonOperator = query.OP_EQ
	q.Limit = 10000
	q.KeyConditions["TheHashKey"] = kc
	json,_ := json.Marshal(q)
	fmt.Printf("JSON:%s\n",string(json))
	body,code,err := q.EndpointReq()
	fmt.Printf("%v\n%v\n%v\n",body,code,err)
}
