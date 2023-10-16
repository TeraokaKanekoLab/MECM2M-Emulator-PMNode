package main

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"mecm2m-Emulator-PMNode/pkg/m2mapi"
	"mecm2m-Emulator-PMNode/pkg/message"
	"mecm2m-Emulator-PMNode/pkg/vmnoder"
	"net"
	"net/http"
	"os"
	"strconv"
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
	vmnoder_port     string
)

func init() {
	// .envファイルの読み込み
	if err := godotenv.Load(os.Getenv("HOME") + "/.env"); err != nil {
		log.Fatal(err)
	}
	vmnoder_port = os.Getenv("VMNODER_PORT")
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
		// inputFormatにはVMNodeIDが含まれる (VMNodeRIDはない)

		// Capabilityの情報をもとに，対象となるVSNodeのVSNodeIDとソケットアドレスを検索する
		payload := `{"statements": [{"statement": "MATCH (vs:VSNode)-[:isPhysicalizedBy]->(ps:PSNode) WHERE ps.Capability = ` + inputFormat.Capability[0] + ` return vs.VNodeID, vs.SocketAddress;"}]}`
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
				//null_data := vmnoder.ResolveCurrentDataByNode{VNodeID: "NULL"}

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

		vmnoder_results.VNodeID = inputFormat.VNodeID

		// 最後にM2M APIへ返送
		jsonData, err := json.Marshal(vmnoder_results)
		if err != nil {
			http.Error(w, "resolveCurrentNode: Error marshaling data", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%v\n", string(jsonData))
	} else {
		http.Error(w, "resolveCurrentNode: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func resolveConditionNode(w http.ResponseWriter, r *http.Request) {
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
		// inputFormatにはVMNodeIDが含まれる (VMNodeRIDはない)

		// Capabilityの情報をもとに，対象となるVSNodeのVSNodeIDとソケットアドレスを検索する
		capability := "\\\"" + inputFormat.Capability[0] + "\\\""
		payload := `{"statements": [{"statement": "MATCH (vs:VSNode)-[:isPhysicalizedBy]->(ps:PSNode) WHERE ps.Capability = ` + capability + ` return vs.VNodeID, vs.SocketAddress;"}]}`
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

		// PSNodeへリクエストを送信するためにリンクプロセスを噛ます
		pnode_id := convertID(inputFormat.VNodeID, 63, 61)
		link_process_socket_address := link_process_socket_address_path + "/access-network_" + pnode_id + ".sock"
		connLinkProcess, err := net.Dial(protocol, link_process_socket_address)
		if err != nil {
			message.MyError(err, "resolveCurrentNode > net.Dial")
		}
		decoderLinkProcess := gob.NewDecoder(connLinkProcess)
		encoderLinkProcess := gob.NewEncoder(connLinkProcess)

		syncFormatClient("Actuate", decoderLinkProcess, encoderLinkProcess)

		if err := encoderLinkProcess.Encode(inputFormat); err != nil {
			message.MyError(err, "resolveCurrentNode > encoderLinkProcess.Encode")
		}

		// PSNodeへ

		// 受信する型は m2mapi.Actuate
		results := m2mapi.Actuate{}
		if err := decoderLinkProcess.Decode(&results); err != nil {
			message.MyError(err, "actuate > decoderLinkProcess.Decode")
		}

		// 最後にM2M APIへ返送
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
		inputFormat := &m2mapi.DataForRegist{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "dataRegister: Error missmatching packet format", http.StatusInternalServerError)
		}

		// Local GraphDB に対して，VSNode が PSNode のセッションキーをキャッシュしていない場合に聞きに行く工程がある

		// SensingDBを開く
		mysql_path := os.Getenv("MYSQL_USERNAME") + ":" + os.Getenv("MYSQL_PASSWORD") + "@tcp(127.0.0.1:" + os.Getenv("MYSQL_PORT") + ")/" + os.Getenv("MYSQL_LOCAL_DB")
		DBConnection, err := sql.Open("mysql", mysql_path)
		if err != nil {
			http.Error(w, "dataRegister: Error opening SensingDB", http.StatusInternalServerError)
		}
		defer DBConnection.Close()
		if err := DBConnection.Ping(); err != nil {
			http.Error(w, "dataRegister: Error connecting SensingDB", http.StatusInternalServerError)
		} else {
			message.MyMessage("DB Connection Success")
		}
		defer DBConnection.Close()
		// DBへの同時接続数の制限
		//DBConnection.SetMaxOpenConns(50)

		// データの挿入
		var cmd string
		table := os.Getenv("MYSQL_TABLE")
		cmd = "INSERT INTO " + table + "(PNodeID, Capability, Timestamp, Value, PSinkID, Lat, Lon) VALUES(?, ?, ?, ?, ?, ?, ?)"
		//cmd = "INSERT INTO " + table + "(PNodeID, Capability, Timestamp) VALUES(?, ?, ?)"
		stmt, err := DBConnection.Prepare(cmd)
		if err != nil {
			http.Error(w, "dataRegister: Error preparing SensingDB", http.StatusInternalServerError)
		}
		defer stmt.Close()

		_, err = stmt.Exec(inputFormat.PNodeID, inputFormat.Capability, inputFormat.Timestamp, inputFormat.Value, inputFormat.PSinkID, inputFormat.Lat, inputFormat.Lon)
		//_, err = stmt.Exec(inputFormat.PNodeID, inputFormat.Capability, inputFormat.Timestamp[:30])
		if err != nil {
			http.Error(w, "dataRegister: Error exec SensingDB", http.StatusInternalServerError)
		}

		fmt.Println("Data Inserted Successfully!")

		// バッファにセンサデータ登録
		mu.Lock()
		registerPNodeID := inputFormat.PNodeID
		bufferSensorData[registerPNodeID] = *inputFormat
		mu.Unlock()
		// チャネルに知らせる
		buffer_chan <- "buffered"
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
