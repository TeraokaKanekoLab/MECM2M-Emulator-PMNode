package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"mecm2m-Emulator-PMNode/pkg/m2mapi"
	"mecm2m-Emulator-PMNode/pkg/message"
	"mecm2m-Emulator-PMNode/pkg/psnode"
	"mecm2m-Emulator-PMNode/pkg/vmnoder"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

const (
	protocol                         = "unix"
	link_process_socket_address_path = "/tmp/mecm2m/link-process"
)

type Format struct {
	FormType string
}

// 充足条件データ取得用のセンサデータのバッファ．(key, value) = (PNodeID, DataForRegist)
var (
	bufferSensorData = make(map[string]m2mapi.DataForRegist)
	mu               sync.Mutex
	buffer_chan      = make(chan string)
	ip_address       string
	vmnoder_port     string
	pmnode_id        string
	vmnode_ip_port   string
)

func init() {
	// .envファイルの読み込み
	if err := godotenv.Load(os.Getenv("HOME") + "/.env"); err != nil {
		log.Fatal(err)
	}
	ip_address = os.Getenv("IP_ADDRESS")
	vmnoder_port = os.Getenv("VMNODER_PORT")
	pmnode_id = os.Getenv("PMNODE_ID")
	vmnode_ip_port = os.Getenv("VMNODE_IP_PORT")
}

func resolvePastNode(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// センサデータ取得対象であるVNodeのSocketAddress (IP:Port) にリクエスト転送
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "resolvePastNode: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapi.ResolveDataByNode{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "resolvePastNode: Error missmatching packet format", http.StatusInternalServerError)
		}

		transmit_request := vmnoder.ResolvePastDataByNode{
			VNodeID:    inputFormat.VNodeID,
			Capability: inputFormat.Capability,
			Period:     vmnoder.PeriodInput{Start: inputFormat.Period.Start, End: inputFormat.Period.End},
		}
		transmit_data, err := json.Marshal(transmit_request)
		if err != nil {
			fmt.Println("Error marshaling data: ", err)
			return
		}
		transmit_url := "http://" + inputFormat.SocketAddress + "/primapi/data/past/node"
		response_data, err := http.Post(transmit_url, "application/json", bytes.NewBuffer(transmit_data))
		if err != nil {
			fmt.Println("Error making request:", err)
			return
		}
		defer response_data.Body.Close()

		byteArray, _ := io.ReadAll(response_data.Body)
		var vsnode_results vmnoder.ResolvePastDataByNode
		if err = json.Unmarshal(byteArray, &vsnode_results); err != nil {
			fmt.Println("Error unmarshaling data: ", err)
			return
		}

		results := m2mapi.ResolveDataByNode{}
		results.VNodeID = vsnode_results.VNodeID
		for _, val := range vsnode_results.Values {
			v := m2mapi.Value{
				Capability: val.Capability,
				Time:       val.Time,
				Value:      val.Value,
			}
			results.Values = append(results.Values, v)
		}

		jsonData, err := json.Marshal(results)
		if err != nil {
			http.Error(w, "resolvePastNode: Error marshaling data", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%v\n", string(jsonData))
	} else {
		http.Error(w, "resolvePastNode: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func resolveCurrentNode(w http.ResponseWriter, r *http.Request) {
	// リクエストの送信元が自PMNode内部のM2M APIかそれ以外で分岐する必要あり
	// 自PMNode内部のM2M APIからのリクエストの場合，Local GraphDBでの検索は不必要
	remoteAdder := r.RemoteAddr
	host, _, err := net.SplitHostPort(remoteAdder)
	if err != nil {
		fmt.Println("Error splitting IP Port: ", err)
		return
	}

	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "resolveCurrentNode: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapi.ResolveDataByNode{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "resolveCurrentNode: Error missmatching packet format", http.StatusInternalServerError)
		}
		if host == "127.0.0.1" {
			// 入力をそのまま対象のVSNodeに転送
			transmit_request := vmnoder.ResolveCurrentDataByNode{
				VNodeID:    inputFormat.VNodeID,
				Capability: inputFormat.Capability,
			}
			transmit_data, err := json.Marshal(transmit_request)
			if err != nil {
				fmt.Println("Error marshaling data: ", err)
				return
			}
			transmit_url := "http://" + inputFormat.SocketAddress + "/primapi/data/current/node"
			response_data, err := http.Post(transmit_url, "application/json", bytes.NewBuffer(transmit_data))
			if err != nil {
				fmt.Println("Error making request:", err)
				return
			}
			defer response_data.Body.Close()

			byteArray, _ := io.ReadAll(response_data.Body)
			var vsnode_results vmnoder.ResolveCurrentDataByNode
			if err = json.Unmarshal(byteArray, &vsnode_results); err != nil {
				fmt.Println("Error unmarshaling data: ", err)
				return
			}

			results := m2mapi.ResolveDataByNode{}
			results.VNodeID = vsnode_results.VNodeID
			for _, val := range vsnode_results.Values {
				v := m2mapi.Value{
					Capability: val.Capability,
					Time:       val.Time,
					Value:      val.Value,
				}
				results.Values = append(results.Values, v)
			}

			jsonData, err := json.Marshal(results)
			if err != nil {
				http.Error(w, "resolveCurrentNode: Error marshaling data", http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "%v\n", string(jsonData))
		} else {
			// inputFormatにはVMNodeIDが含まれる (VMNodeRIDはない)

			// Capabilityの情報をもとに，対象となるVSNodeのVSNodeIDとソケットアドレスを検索する
			var format_capability []string
			for _, cap := range inputFormat.Capability {
				cap = "\\\"" + cap + "\\\""
				format_capability = append(format_capability, cap)
			}
			payload := `{"statements": [{"statement": "MATCH (vs:VSNode)-[:isPhysicalizedBy]->(ps:PSNode) WHERE ps.Capability IN [` + strings.Join(format_capability, ", ") + `] return vs.VNodeID, vs.SocketAddress;"}]}`
			graphdb_url := "http://" + os.Getenv("NEO4J_USERNAME") + ":" + os.Getenv("NEO4J_LOCAL_PASSWORD") + "@" + "localhost:" + os.Getenv("NEO4J_LOCAL_PORT_GOLANG") + "/db/data/transaction/commit"
			req, _ := http.NewRequest("POST", graphdb_url, bytes.NewBuffer([]byte(payload)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "*/*")

			client := new(http.Client)
			resp, err := client.Do(req)
			if err != nil {
				message.MyError(err, "resolvePointFunction > client.Do()")
			}
			defer resp.Body.Close()

			byteArray, _ := io.ReadAll(resp.Body)
			values := bodyGraphDB(byteArray)

			var row_data interface{}
			var vsnode_result vmnoder.ResolveNode
			for _, v1 := range values {
				for k2, v2 := range v1.(map[string]interface{}) {
					if k2 == "data" {
						for _, v3 := range v2.([]interface{}) {
							for k4, v4 := range v3.(map[string]interface{}) {
								if k4 == "row" {
									row_data = v4
									dataArray := row_data.([]interface{})
									vnode_set := vmnoder.VNodeSet{
										VNodeID:       dataArray[0].(string),
										SocketAddress: dataArray[1].(string),
									}
									vsnode_result.VNode = append(vsnode_result.VNode, vnode_set)
								}
							}
						}
					}
				}
			}

			// 取得したVSNodeのそれぞれにデータ取得要求を転送
			var wg sync.WaitGroup
			var vmnoder_results m2mapi.ResolveDataByNode
			for _, vnode_set := range vsnode_result.VNode {
				wg.Add(1)
				go func(vnode_set vmnoder.VNodeSet) {
					defer wg.Done()

					vmnoder_data := vmnoder.ResolveCurrentDataByNode{
						VNodeID:    vnode_set.VNodeID,
						Capability: inputFormat.Capability,
					}
					transmit_data, err := json.Marshal(vmnoder_data)
					if err != nil {
						fmt.Println("Error marshaling data: ", err)
						return
					}
					transmit_url := "http://" + vnode_set.SocketAddress + "/primapi/data/current/node"
					response_data, err := http.Post(transmit_url, "application/json", bytes.NewBuffer(transmit_data))
					if err != nil {
						fmt.Println("Error making request: ", err)
						return
					}
					defer response_data.Body.Close()

					byteArray, _ := io.ReadAll(response_data.Body)
					var results vmnoder.ResolveCurrentDataByNode
					if err = json.Unmarshal(byteArray, &results); err != nil {
						fmt.Println("Error unmarshaling data: ", err)
						return
					}

					for _, val := range results.Values {
						m2mapi_value := m2mapi.Value{
							Capability: val.Capability,
							Time:       val.Time,
							Value:      val.Value,
						}
						vmnoder_results.Values = append(vmnoder_results.Values, m2mapi_value)
					}
				}(vnode_set)
			}
			wg.Wait()

			vmnoder_results.VNodeID = inputFormat.VNodeID

			// 最後にM2M APIへ返送
			jsonData, err := json.Marshal(vmnoder_results)
			if err != nil {
				http.Error(w, "resolveCurrentNode: Error marshaling data", http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "%v\n", string(jsonData))
		}
	} else {
		http.Error(w, "resolveCurrentNode: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func resolveConditionNode(w http.ResponseWriter, r *http.Request) {
	// リクエストの送信元が自PMNode内部のM2M APIかそれ以外で分岐する必要あり
	// 自PMNode内部のM2M APIからのリクエストの場合，Local GraphDBでの検索は不必要
	remoteAdder := r.RemoteAddr
	host, _, err := net.SplitHostPort(remoteAdder)
	if err != nil {
		fmt.Println("Error splitting IP Port: ", err)
		return
	}

	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "resolveConditionNode: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapi.ResolveDataByNode{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "resolveConditionNode: Error missmatching packet format", http.StatusInternalServerError)
		}
		if host == "127.0.0.1" {
			// 入力をそのまま対象のVSNodeに転送
			transmit_request := vmnoder.ResolveConditionDataByNode{
				VNodeID:    inputFormat.VNodeID,
				Capability: inputFormat.Capability,
				Condition:  vmnoder.ConditionInput{Limit: vmnoder.Range{LowerLimit: inputFormat.Condition.Limit.LowerLimit, UpperLimit: inputFormat.Condition.Limit.UpperLimit}, Timeout: inputFormat.Condition.Timeout},
			}
			transmit_data, err := json.Marshal(transmit_request)
			if err != nil {
				fmt.Println("Error marshaling data: ", err)
				return
			}
			transmit_url := "http://" + inputFormat.SocketAddress + "/primapi/data/condition/node"
			response_data, err := http.Post(transmit_url, "application/json", bytes.NewBuffer(transmit_data))
			if err != nil {
				fmt.Println("Error making request:", err)
				return
			}
			defer response_data.Body.Close()

			byteArray, _ := io.ReadAll(response_data.Body)
			var vsnode_results vmnoder.ResolveConditionDataByNode
			if err = json.Unmarshal(byteArray, &vsnode_results); err != nil {
				fmt.Println("Error unmarshaling data: ", err)
				return
			}

			fmt.Println("vsnode_results: ", vsnode_results)

			results := m2mapi.ResolveDataByNode{}
			results.VNodeID = vsnode_results.VNodeID
			results.Values = append(results.Values, m2mapi.Value(vsnode_results.Values))

			jsonData, err := json.Marshal(results)
			if err != nil {
				http.Error(w, "resolveCurrentNode: Error marshaling data", http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "%v\n", string(jsonData))
		} else {
			// inputFormatにはVMNodeIDが含まれる (VMNodeRIDはない)

			// Capabilityの情報をもとに，対象となるVSNodeのVSNodeIDとソケットアドレスを検索する
			var format_capability []string
			for _, cap := range inputFormat.Capability {
				cap = "\\\"" + cap + "\\\""
				format_capability = append(format_capability, cap)
			}
			payload := `{"statements": [{"statement": "MATCH (vs:VSNode)-[:isPhysicalizedBy]->(ps:PSNode) WHERE ps.Capability IN [` + strings.Join(format_capability, ", ") + `] return vs.VNodeID, vs.SocketAddress;"}]}`
			graphdb_url := "http://" + os.Getenv("NEO4J_USERNAME") + ":" + os.Getenv("NEO4J_LOCAL_PASSWORD") + "@" + "localhost:" + os.Getenv("NEO4J_LOCAL_PORT_GOLANG") + "/db/data/transaction/commit"
			req, _ := http.NewRequest("POST", graphdb_url, bytes.NewBuffer([]byte(payload)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "*/*")

			client := new(http.Client)
			resp, err := client.Do(req)
			if err != nil {
				message.MyError(err, "resolvePointFunction > client.Do()")
			}
			defer resp.Body.Close()

			byteArray, _ := io.ReadAll(resp.Body)
			values := bodyGraphDB(byteArray)

			var row_data interface{}
			var vsnode_result vmnoder.ResolveNode
			for _, v1 := range values {
				for k2, v2 := range v1.(map[string]interface{}) {
					if k2 == "data" {
						for _, v3 := range v2.([]interface{}) {
							for k4, v4 := range v3.(map[string]interface{}) {
								if k4 == "row" {
									row_data = v4
									dataArray := row_data.([]interface{})
									vnode_set := vmnoder.VNodeSet{
										VNodeID:       dataArray[0].(string),
										SocketAddress: dataArray[1].(string),
									}
									vsnode_result.VNode = append(vsnode_result.VNode, vnode_set)
								}
							}
						}
					}
				}
			}

			// 取得したVSNodeのそれぞれにデータ取得要求を転送
			var wg sync.WaitGroup
			var vmnoder_results m2mapi.ResolveDataByNode
			for _, vnode_set := range vsnode_result.VNode {
				wg.Add(1)
				go func(vnode_set vmnoder.VNodeSet) {
					defer wg.Done()
					vmnoder_data := vmnoder.ResolveConditionDataByNode{
						VNodeID:    vnode_set.VNodeID,
						Capability: inputFormat.Capability,
						Condition:  vmnoder.ConditionInput{Limit: vmnoder.Range{LowerLimit: inputFormat.Condition.Limit.LowerLimit, UpperLimit: inputFormat.Condition.Limit.UpperLimit}, Timeout: inputFormat.Condition.Timeout},
					}
					transmit_data, err := json.Marshal(vmnoder_data)
					if err != nil {
						fmt.Println("Error marshaling data: ", err)
						return
					}
					transmit_url := "http://" + vnode_set.SocketAddress + "/primapi/data/condition/node"
					response_data, err := http.Post(transmit_url, "application/json", bytes.NewBuffer(transmit_data))
					if err != nil {
						fmt.Println("Error making request: ", err)
						return
					}
					defer response_data.Body.Close()

					byteArray, _ := io.ReadAll(response_data.Body)
					var results vmnoder.ResolveConditionDataByNode
					if err = json.Unmarshal(byteArray, &results); err != nil {
						fmt.Println("Error unmarshaling data: ", err)
						return
					}

					m2mapi_value := m2mapi.Value{
						Capability: results.Values.Capability,
						Time:       results.Values.Time,
						Value:      results.Values.Value,
					}
					vmnoder_results.Values = append(vmnoder_results.Values, m2mapi_value)
				}(vnode_set)
			}
			wg.Wait()

			vmnoder_results.VNodeID = inputFormat.VNodeID

			// 最後にM2M APIへ返送
			jsonData, err := json.Marshal(vmnoder_results)
			if err != nil {
				http.Error(w, "resolveConditionNode: Error marshaling data", http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "%v\n", string(jsonData))
		}
	} else {
		http.Error(w, "resolveConditionNode: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func actuate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "actuate: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapi.Actuate{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "actuate: Error missmatching packet format", http.StatusInternalServerError)
		}

		// 入力をそのまま対象のVSNodeに転送
		transmit_request := vmnoder.Actuate{
			VNodeID:    inputFormat.VNodeID,
			Capability: inputFormat.Capability,
			Action:     inputFormat.Action,
			Parameter:  inputFormat.Parameter,
		}
		transmit_data, err := json.Marshal(transmit_request)
		if err != nil {
			fmt.Println("Error marshaling data: ", err)
			return
		}
		transmit_url := "http://" + inputFormat.SocketAddress + "/primapi/actuate"
		response_data, err := http.Post(transmit_url, "application/json", bytes.NewBuffer(transmit_data))
		if err != nil {
			fmt.Println("Error making request: ", err)
			return
		}
		defer response_data.Body.Close()

		byteArray, _ := io.ReadAll(response_data.Body)
		var vsnode_results vmnoder.Actuate
		if err = json.Unmarshal(byteArray, &vsnode_results); err != nil {
			fmt.Println("Error unmarshaling data: ", err)
			return
		}

		results := m2mapi.Actuate{}
		results.VNodeID = vsnode_results.VNodeID
		results.Status = vsnode_results.Status

		jsonData, err := json.Marshal(results)
		if err != nil {
			http.Error(w, "actuate: Error marshaling data", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%v\n", string(jsonData))
	} else {
		http.Error(w, "actuate: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func dataRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "dataRegister: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &psnode.DataForRegist{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "dataRegister: Error missmatching packet format", http.StatusInternalServerError)
		}

		// MEC ServerのSensingDBに登録する際は，センサデータのPNodeIDはPMNodeIDに変換する
		inputFormat.PNodeID = pmnode_id

		// Home MEC Server のVMNodeにセンサデータを転送
		transmit_data, err := json.Marshal(&inputFormat)
		if err != nil {
			fmt.Println("Error marshaling data: ", err)
			return
		}
		transmit_url := "http://" + vmnode_ip_port + "/data/register"
		response_data, err := http.Post(transmit_url, "application/json", bytes.NewBuffer(transmit_data))
		if err != nil {
			fmt.Println("Error making request: ", err)
			return
		}

		byteArray, _ := io.ReadAll(response_data.Body)
		fmt.Fprintf(w, "%v\n", string(byteArray))
	} else {
		http.Error(w, "dataRegister: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func startServer(port int) {
	mux := http.NewServeMux() // 新しいServeMuxインスタンスを作成
	mux.HandleFunc("/primapi/data/past/node", resolvePastNode)
	mux.HandleFunc("/primapi/data/current/node", resolveCurrentNode)
	mux.HandleFunc("/primapi/data/condition/node", resolveConditionNode)
	mux.HandleFunc("/primapi/actuate", actuate)
	mux.HandleFunc("/data/register", dataRegister)

	address := fmt.Sprintf(":%d", port)
	log.Printf("Starting server on %s", address)

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error starting server on port %d: %v", port, err)
	}
}

type Ports struct {
	Port []int `json:"ports"`
}

func main() {
	/*
		// Mainプロセスのコマンドラインからシミュレーション実行開始シグナルを受信するまで待機
		signals_from_main := make(chan os.Signal, 1)

		// 停止しているプロセスを再開するために送信されるシグナル，SIGCONT(=18)を受信するように設定
		signal.Notify(signals_from_main, syscall.SIGCONT)

		// シグナルを待機
		fmt.Println("Waiting for signal...")
		sig := <-signals_from_main

		// 受信したシグナルを表示
		fmt.Printf("Received signal: %v\n", sig)
	*/

	vmnoder_port_int, _ := strconv.Atoi(vmnoder_port)
	startServer(vmnoder_port_int)
}

func syncFormatClient(command string, decoder *gob.Decoder, encoder *gob.Encoder) {
	format := &Format{}
	switch command {
	case "CurrentNode":
		format.FormType = "CurrentNode"
	}
	if err := encoder.Encode(format); err != nil {
		message.MyError(err, "syncFormatClient > "+command+" > encoder.Encode")
	}
}

func convertID(id string, pos ...int) string {
	id_int := new(big.Int)

	_, ok := id_int.SetString(id, 10)
	if !ok {
		message.MyMessage("Failed to convert string to big.Int")
	}

	for _, position := range pos {
		mask := new(big.Int).Lsh(big.NewInt(1), uint(position))
		id_int.Xor(id_int, mask)
	}
	return id_int.String()
}

func bodyGraphDB(byteArray []byte) []interface{} {
	var jsonBody map[string]interface{}
	if err := json.Unmarshal(byteArray, &jsonBody); err != nil {
		message.MyError(err, "bodyGraphDB > json.Unmarshal")
		return nil
	}
	var values []interface{}
	for _, v1 := range jsonBody {
		switch v1.(type) {
		case []interface{}:
			for range v1.([]interface{}) {
				values = v1.([]interface{})
			}
		case map[string]interface{}:
			for _, v2 := range v1.(map[string]interface{}) {
				switch v2.(type) {
				case []interface{}:
					values = v2.([]interface{})
				default:
				}
			}
		default:
			fmt.Println("Format Assertion False")
		}
	}
	return values
}
