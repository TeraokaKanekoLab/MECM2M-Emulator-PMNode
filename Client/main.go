package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mecm2m-Emulator-PMNode/pkg/m2mapi"
	"mecm2m-Emulator-PMNode/pkg/m2mapp"
	"mecm2m-Emulator-PMNode/pkg/psnode"
	"net/http"
	"os"
	"time"
)

func main() {
	var data any
	var url string
	args := os.Args

	switch args[1] {
	case "area":
		data = m2mapp.ResolveAreaInput{
			NE: m2mapp.SquarePoint{Lat: 35.531, Lon: 139.531},
			SW: m2mapp.SquarePoint{Lat: 35.53, Lon: 139.53},
		}
		url = "http://localhost:8080/m2mapi/area"
	case "node":
		data = m2mapp.ResolveNodeInput{
			AD:         args[2],
			Capability: []string{"MaxTemp", "MaxHumid", "MaxWind"},
			NodeType:   "VSNode",
		}
		url = "http://localhost:8080/m2mapi/node"
	case "past_node":
		data = m2mapi.ResolveDataByNode{
			VNodeID:       "9223372036854775808",
			Capability:    []string{"MaxTemp"},
			Period:        m2mapi.PeriodInput{Start: "2023-09-25 11:59:50 +0900 JST", End: "2023-09-25 12:00:05 +0900 JST"},
			SocketAddress: "192.168.1.1:11000",
		}
		url = "http://localhost:8080/m2mapi/data/past/node"
	case "current_node":
		data = m2mapi.ResolveDataByNode{
			VNodeID:       "9223372036854775808",
			Capability:    []string{"MaxTemp"},
			SocketAddress: "192.168.1.1:11000",
		}
		url = "http://localhost:8080/m2mapi/data/current/node"
	case "condition_node":
		data = m2mapi.ResolveDataByNode{
			VNodeID:       "9223372036854775808",
			Capability:    []string{"MaxTemp"},
			Condition:     m2mapi.ConditionInput{Limit: m2mapi.Range{LowerLimit: 33, UpperLimit: 37}, Timeout: 10 * time.Second},
			SocketAddress: "192.168.1.1:11000",
		}
		url = "http://localhost:8080/m2mapi/data/condition/node"
	case "past_area":
		data = m2mapi.ResolveDataByArea{
			AD:         args[2],
			Capability: "MaxTemp",
			Period:     m2mapi.PeriodInput{Start: "2023-08-16 04:55:50 +0900 JST", End: "2023-08-16 04:56:00 +0900 JST"},
			NodeType:   "VSNode",
		}
		url = "http://localhost:8080/m2mapi/data/past/area"
	case "current_area":
		data = m2mapi.ResolveDataByArea{
			AD:         args[2],
			Capability: "MaxHumid",
			NodeType:   "VSNode",
		}
		url = "http://localhost:8080/m2mapi/data/current/area"
	case "condition_area":
		data = m2mapi.ResolveDataByArea{
			AD:         args[2],
			Capability: "MaxTemp",
			Condition:  m2mapi.ConditionInput{Limit: m2mapi.Range{LowerLimit: 33, UpperLimit: 37}, Timeout: 10 * time.Second},
			NodeType:   "VSNode",
		}
		url = "http://localhost:8080/m2mapi/data/condition/area"
	case "extend_ad":
		data = m2mapi.ExtendAD{
			AD: "c000121688",
		}
		url = "http://localhost:8080/m2mapi/area/extend"
	case "time":
		data = psnode.TimeSync{
			PNodeID:     "2305843009213693952",
			CurrentTime: time.Now(),
		}
		url = "http://localhost:14000/time"
	default:
		fmt.Println("There is no args")
		log.Fatal()
	}

	client_data, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling data: ", err)
		return
	}
	response, err := http.Post(url, "application/json", bytes.NewBuffer(client_data))
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Server Response:", string(body))
}
