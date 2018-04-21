package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"github.com/pkg/errors"
	"fmt"
	"strconv"
)

type Edge struct {
	Id    string
	Node1 string
	Node2 string
	Attrs map[string][]string
}

type Graph struct {
	Edges [] Edge
	Price float64
}

func GraphToStr(graph Graph) (string, error) {
	graphstr, err := yaml.Marshal(graph)
	if err != nil {
		return "", err
	}

	return string(graphstr), nil
}

func LoadGraphStr(data []byte) (Graph, error) {

	g := Graph{}
	err := yaml.Unmarshal(data, &g)
	if err != nil {
		log.Fatalf("error: %v", err)
		return Graph{}, err
	}

	return g, nil
}

func LoadGraph(fpath string) (Graph, error) {
	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		log.Fatal(err)
		return Graph{}, err
	}

	return LoadGraphStr(data)

}

func LoadRequestStr(data []byte) (Edge, error) {
	e := Edge{}
	err := yaml.Unmarshal([]byte(data), &e)
	if err != nil {
		log.Fatalf("error: %v", err)
		return Edge{}, err
	}

	return e, nil

}

func LoadRequest(fpath string) (Edge, error) {
	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		log.Fatal(err)
		return Edge{}, err
	}
	return LoadRequestStr(data)

}

func FitRequestStr(graph Graph, request Edge) (string, float64, error) {

	edge, price, err := FitRequest(graph, request)
	if err != nil {
		log.Fatal(err)
	}
	edgeStr, err := yaml.Marshal(edge)
	if err != nil {
		log.Fatal(err)
	}

	return string(edgeStr), price, err

}
func FitRequest(graph Graph, request Edge) (Edge, float64, error) {
	//log.Println("embedding")
	//log.Println(request)
	candidateEdges := []Edge{}
	for _, e := range graph.Edges {
		winner := true
		//we have the same start and end nodes
		if request.Node1 == e.Node1 && request.Node2 == e.Node2 {
			attrs := e.Attrs

			//make sure we matche on every edge
			for k, dataEdge := range request.Attrs {
				if dataGraph, ok := attrs[k]; ok {

					if ! doesAttrFit(dataGraph[0], dataGraph[1], dataEdge [0], dataEdge [1]) {
						winner = false
						break
					}
				} else {
					log.Fatal("failed to get attr from graph")

				}

			}
			if (!winner) {
				continue
			} else {
				candidateEdges = append(candidateEdges, e)

			}
		}
	}

	if len(candidateEdges) == 0 {
		return Edge{}, -1, errors.New("failed to embed request")
	} else {
		winner := candidateEdges[0]
		winnerPrice, _ := strconv.ParseFloat(winner.Attrs["price"][1], 64)
		for _, candidate := range candidateEdges[1:] {
			if candidatePrice, err := strconv.ParseFloat(candidate.Attrs["price"][1], 64); err == nil && candidatePrice < winnerPrice {
				winner = candidate //MAKE SURE TO RETURN REQUEST NOT NETWORK!
				winnerPrice = candidatePrice

			}

		}

		request.Id = winner.Id
		return request, winnerPrice, nil

	}

}
func DummyPricer(target float64, prices []float64) (float64, error) {
	log.Println("DummyPricer")
	log.Println(target)
	log.Println(prices)
	var sum = 0.0
	for _, p := range prices {
		sum += p
	}
	if sum < target {
		if 2*sum < target {
			return sum * 2, nil
		} else {
			return target * 0.9, nil
		}

	} else {
		return 0.0, errors.New("can't compete with price: " + fmt.Sprintf("%.2f", target) + " with our best offer+" + fmt.Sprintf("%.2f", sum))

	}
}

func FitPathStr(graph Graph, request []Edge, princeFunc func(target float64, prices []float64) (float64, error), target float64) (string, float64, error) {

	g, price, err := FitPath(graph, request, princeFunc, target)

	if err != nil {
		return "", 0, err
	} else {

		graphstr, err := yaml.Marshal(g)
		return string(graphstr), price, err

	}

}
func FitPath(graph Graph, requests []Edge, princeFunc func(target float64, prices []float64) (float64, error), target float64) (Graph, float64, error) {

	res := make([]Edge, 0)
	prices := make([]float64, len(requests))

	for _, e := range requests {
		if edge, price, err := FitRequest(graph, e); err == nil {
			res = append(res, edge)
			prices = append(prices, price)

		} else {
			return Graph{}, 0.0, errors.New("path cannot be embedded on the network due to lack of resource ")
		}
	}

	if lowerPrice, err := princeFunc(target, prices); err == nil {
		g := Graph{Edges: res}
		g.Price = lowerPrice
		return g, lowerPrice, nil
	} else {
		return Graph{}, 0.0, errors.New("path cannot be embedded on the network with competitive price")
	}
}

func doesAttrFit(graphComparator string, graphValue string, edgeComparator string, edgeValue string) bool {
	if edgeComparator == "=" {
		return graphValue == edgeValue
	} else if edgeComparator == "<" {
		return StringToFloat(graphValue) <= StringToFloat(edgeValue)
	} else if edgeComparator == ">" {
		return StringToFloat(graphValue) >= StringToFloat(edgeValue)
	} else {
		log.Fatalf("illegal comparator")
		panic("illegal comparator")
	}

}

func StringToFloat(s string) float64 {
	res, _ := strconv.ParseFloat(s, 64)
	return res
}

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func updateGraph(graph Graph, requestEdges []Edge) (Graph, error) {
	//var res Graph = graph
	for _, requestEdge := range requestEdges {
		for _, candidateEdge := range graph.Edges {
			//log.Println(candidateEdge.Id)
			//log.Println(requestEdge.Id)
			if candidateEdge.Id == requestEdge.Id {
				for k, dataEdge := range (candidateEdge.Attrs) {

					if len(dataEdge) == 2 || dataEdge[2] != "resource" { //request does not consume resource
						//log.Println("skipped")
						continue
					} else {

						previous, _ := strconv.ParseFloat(dataEdge[1], 64)
						consumed, _ := strconv.ParseFloat(requestEdge.Attrs[k][1], 64)
						var newResourceValue = previous - consumed

						if newResourceValue < 0 {

							return Graph{}, errors.New("failed to allocate enough resources for edge " + requestEdge.Id)
						}

						candidateEdge.Attrs[k][1] = FloatToString(newResourceValue)

					}
				}
			}
		}

	}

	return graph, nil
}
