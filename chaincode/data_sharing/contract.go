package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type SimpleChaincode struct {
}

type data struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Name       string `json:"name"`
	Content    string `json:"content"`
	Date       string `json:"date"`
	Time       string `json:"time"`
	Owner      string `json:"owner"`
}

type request_data struct {
	ObjectType string `json:"docType"`
	Name       string `json:"name"`
	DataTxid   string `json:"datatxid"`
	Requestor  string `json:"requestor"`
}

type reply struct {
	ObjectType  string `json:"docType"`
	Name        string `json:"name"`
	RequestTxid string `json:"requesttxid"`
	reply       string `json:"reply"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	if function == "publishData" {
		return t.publishData(stub, args)
	} else if function == "showDataInfo" {
		return t.showDataInfo(stub, args)
	} else if function == "showPendingRequests" {
		return t.showPendingRequests(stub, args)
	} else if function == "requestData" {
		return t.requestData(stub, args)
	} else if function == "handleRequest" {
		return t.handleRequest(stub, args)
	}

	fmt.Println("invoke did not find func: " + function)
	return shim.Error("Received unknown function invocation")
}

func (t *SimpleChaincode) publishData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	fmt.Println("- start publishing data")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	name := args[0]
	content := args[1]
	date := strings.ToLower(args[2])
	time := strings.ToLower(args[3])
	owner := strings.ToLower(args[4])

	dataAsBytes, err := stub.GetState(name)
	if err != nil {
		return shim.Error("Failed to get data: " + err.Error())
	} else if dataAsBytes != nil {
		fmt.Println("This data already exists: " + name)
		return shim.Error("This data already exists: " + name)
	}

	objectType := "data"
	data := &data{objectType, name, content, date, time, owner}
	dataJSONasBytes, err := json.Marshal(data)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(name, dataJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end publishing data")
	return shim.Success(nil)
}

func (t *SimpleChaincode) showDataInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the marble to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

func (t *SimpleChaincode) showPendingRequests(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	owner := strings.ToLower(args[0])
	var data_list []string
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"data\",\"owner\":\"%s\"}}", owner)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		data_list = append(data_list, queryResponse.Key)
	}
	queryResults, err := getQueryListForRequests(stub, data_list)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (t *SimpleChaincode) requestData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	fmt.Println("- start requesting data")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	name := args[0]
	datatxid := args[1]
	requestor := strings.ToLower(args[2])

	dataAsBytes, err := stub.GetState(datatxid)
	if err != nil {
		return shim.Error("Failed to get data: " + err.Error())
	}
	fmt.Println(dataAsBytes)

	objectType := "request"
	data := &request_data{objectType, name, datatxid, requestor}
	dataJSONasBytes, err := json.Marshal(data)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(name, dataJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end requesting data")
	return shim.Success(nil)
}

func (t *SimpleChaincode) handleRequest(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	fmt.Println("- start handling data")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	name := args[0]
	requesttxid := args[1]
	response := args[2]

	dataAsBytes, err := stub.GetState(requesttxid)
	if err != nil {
		return shim.Error("Failed to get request: " + err.Error())
	}
	fmt.Println(dataAsBytes)

	objectType := "response"
	data := &reply{objectType, name, requesttxid, response}
	dataJSONasBytes, err := json.Marshal(data)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(name, dataJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.DelState(requesttxid)
	if err != nil {
		return shim.Error("Failed to delete request:" + err.Error())
	}
	fmt.Println("- end handingling request")
	return shim.Success(nil)
}

//returns a json of Requests based on a list of Data
func getQueryListForRequests(stub shim.ChaincodeStubInterface, querylist []string) ([]byte, error) {

	fmt.Printf("- getQueryListForRequests querylist:\n%s\n", querylist)
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for _, element := range querylist{
		requestQueryString := fmt.Sprintf("{\"selector\":{\"docType\":\"request\",\"datatxid\":\"%s\"}}", element)

		resultsIterator, err := stub.GetQueryResult(requestQueryString)
		if err != nil {
			return nil, err
		}
		defer resultsIterator.Close()

		// buffer is a JSON array containing QueryRecords

		for resultsIterator.HasNext() {
			queryResponse, err := resultsIterator.Next()
			if err != nil {
				return nil, err
			}
			// Add a comma before array members, suppress it for the first array member
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}
			buffer.WriteString("{\"Key\":")
			buffer.WriteString("\"")
			buffer.WriteString(queryResponse.Key)
			buffer.WriteString("\"")

			buffer.WriteString(", \"Record\":")
			// Record is a JSON object, so we write as-is
			buffer.WriteString(string(queryResponse.Value))
			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
		}
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}