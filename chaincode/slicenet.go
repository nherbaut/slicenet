package main

//nico
import (
	"log"
	"fmt"
	"strconv"

	b64 "encoding/base64"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"math"
	"github.com/pkg/errors"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Get the args from the transaction proposal
	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}

	// Set up any variables or assets here by calling stub.PutState()

	// We store the key and the value on the ledger
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "set" {
		result, err = set(stub, args)
	} else if fn == "priceTD" {
		edgeRes, _, err := priceTD(stub, args)
		if err == nil {
			return pb.Response{shim.OK, "ok", []byte(b64.StdEncoding.EncodeToString([]byte(edgeRes)))}

		} else {
			return shim.Error(err.Error())
		}

	} else if fn == "commit" {
		result, err = commit(stub, args[0], args[1])
		if err != nil {
			return shim.Error(err.Error())
		}
		result, err = set(stub, []string{args[0], result})

	} else { // assume 'get' even if fn is nil
		result, err = get(stub, args)
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

// generate a princing for a
func priceTD(stub shim.ChaincodeStubInterface, args []string) (string, float64, error) {
	if len(args) < 2 {
		return "", 0, fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	netId := args[0]
	requestPathData := args[1]

	competitorPrice := math.MaxFloat64

	if len(args) == 3 {
		userPrice, err := strconv.ParseFloat(args[2], 64)
		competitorPrice = userPrice
		if (err != nil) {
			log.Fatal(err)
		}
	}

	log.Println("competitor price:")
	log.Println(competitorPrice)

	netdata, err := get(stub, []string{netId})
	if err != nil {
		log.Fatal("failed to get network for id" + netId)

	}
	graphData, err := b64.StdEncoding.DecodeString(netdata)
	if err != nil {
		log.Fatal("failed to decode base64: " + netdata)

	}
	requestData, err := b64.StdEncoding.DecodeString(requestPathData)
	if err != nil {
		log.Fatal("failed to decode base64: " + netdata)

	}
	g, err := LoadGraphStr(graphData)

	if err != nil {
		log.Fatal("failed to load Graph : " + string(graphData))

	}

	request, err := LoadGraphStr(requestData)

	if err != nil {
		log.Fatal("failed to load request : " + string(requestData))

	}
	winnerGraph, winnerPrice, err := FitPathStr(g, request.Edges, DummyPricer, competitorPrice)

	if err != nil {
		return "", 0, errors.New("failed to fit request in graph ")
	}

	bestPrice, err := DummyPricer(competitorPrice, []float64{winnerPrice})

	return winnerGraph, bestPrice, err

}

func NetID2Graph(stub shim.ChaincodeStubInterface, netid string) (Graph, error) {
	datab64, err := get(stub, []string{netid})
	if err != nil {
		return Graph{}, err
	}
	data, err := b64.StdEncoding.DecodeString(datab64)
	graph, err := LoadGraphStr([]byte(data))
	return graph, err

}

// This function removes the provided edge array from the TD resources
func commit(stub shim.ChaincodeStubInterface, netId string, requestb64 string) (string, error) {

	requestData, err := b64.StdEncoding.DecodeString(requestb64)
	if err != nil {
		return "", err
	}

	request, err := LoadGraphStr(requestData)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	graph, err := NetID2Graph(stub, netId)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	graph, err = updateGraph(graph, request.Edges)
	if err != nil {
		return "", err
	}

	graphStr, err := GraphToStr(graph)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	return b64.StdEncoding.EncodeToString([]byte(graphStr)), nil

}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	return args[1], nil
}

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return string(value), nil
}

/**/
func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}

/*
// main function starts up the chaincode in the container during instantiate

func main() {
	//if err := shim.Start(new(SimpleAsset)); err != nil {
	//      fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	//}
	g, _ := LoadGraph("slicenet/net.yaml")

	for i := 0; i < 10; i++ {
		ae, _ := LoadGraph("slicenet/request" + strconv.Itoa(i) + ".yaml")

		fmt.Println(ae)

		_, _, err := FitPath(g, ae.Edges, DummyPricer, 9999999)
		if err != nil {
			fmt.Println(strconv.Itoa(i) + " no good!")
			continue
		}

		fmt.Println(strconv.Itoa(i) + " OK!")

		
		fmt.Println("commiting edge:")
		fmt.Println(graph.Edges[0])

		g2, err := updateGraph(g, graph.Edges)
		if err != nil {
			log.Fatal(err)
		}

		strRep, _ := yaml.Marshal(g2)
		fmt.Println(string(strRep))

	}

} /**/
