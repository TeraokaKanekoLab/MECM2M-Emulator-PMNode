package main

import (
	"context"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"mecm2m-Emulator-PMNode/pkg/m2mapi"
	"mecm2m-Emulator-PMNode/pkg/message"
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
var bufferSensorData = make(map[string]m2mapi.DataForRegist)
var mu sync.Mutex
var buffer_chan = make(chan string)

func init() {
	// .envファイルの読み込み
	if err := godotenv.Load(os.Getenv("HOME") + "/.env"); err != nil {
		log.Fatal(err)
	}
}

func resolvePastNode(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "resolvePastNode: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapi.ResolveDataByNode{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "resolvePastNode: Error missmatching packet format", http.StatusInternalServerError)
		}

		// 入力のVNodeIDをPNodeIDに変換
		pnode_id := convertID(inputFormat.VNodeID, 63, 61)
		capability := inputFormat.Capability
		start := inputFormat.Period.Start
		end := inputFormat.Period.End

		// SensingDBを開く
		mysql_path := os.Getenv("MYSQL_USERNAME") + ":" + os.Getenv("MYSQL_PASSWORD") + "@tcp(127.0.0.1:" + os.Getenv("MYSQL_PORT") + ")/" + os.Getenv("MYSQL_LOCAL_DB")
		DBConnection, err := sql.Open("mysql", mysql_path)
		if err != nil {
			http.Error(w, "resolvePastNode: Error opening SensingDB", http.StatusInternalServerError)
		}
		defer DBConnection.Close()
		if err := DBConnection.Ping(); err != nil {
			http.Error(w, "resolvePastNode: Error connecting SensingDB", http.StatusInternalServerError)
		} else {
			message.MyMessage("DB Connection Success")
		}
		defer DBConnection.Close()
		// DBへの同時接続数の制限
		//DBConnection.SetMaxOpenConns(50)

		var cmd string
		table := os.Getenv("MYSQL_TABLE")
		cmd = "SELECT * FROM " + table + " WHERE PNodeID = \"" + pnode_id + "\" AND Capability = \"" + capability + "\" AND Timestamp > \"" + start + "\" AND Timestamp <= \"" + end + "\";"

		rows, err := DBConnection.Query(cmd)
		if err != nil {
			http.Error(w, "resolvePastNode: Error querying SensingDB", http.StatusInternalServerError)
		}
		defer rows.Close()

		var vnode_id string
		var vals []m2mapi.Value
		for rows.Next() {
			field := []string{"0", "0", "0", "0", "0", "0", "0"}
			// PNodeID, Capability, Timestamp, Value, PSinkID, Lat, Lon
			err := rows.Scan(&field[0], &field[1], &field[2], &field[3], &field[4], &field[5], &field[6])
			if err != nil {
				fmt.Println(err)
				http.Error(w, "resolvePastNode: Error scanning sensing data", http.StatusInternalServerError)
				return
			}
			vnode_id = convertID(field[0], 63)
			valFloat, _ := strconv.ParseFloat(field[3], 64)
			val := m2mapi.Value{Capability: field[1], Time: field[2], Value: valFloat}
			vals = append(vals, val)
		}
		result := m2mapi.ResolveDataByNode{
			VNodeID: vnode_id,
			Values:  vals,
		}

		jsonData, err := json.Marshal(result)
		if err != nil {
			http.Error(w, "resolvePastNode: Error marshaling data", http.StatusInternalServerError)
			return
		}

		fmt.Println(string(jsonData))
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

		// PSNodeへリクエストを送信するためにリンクプロセスを噛ます
		pnode_id := convertID(inputFormat.VNodeID, 63, 61)
		link_process_socket_address := link_process_socket_address_path + "/access-network_" + pnode_id + ".sock"
		connLinkProcess, err := net.Dial(protocol, link_process_socket_address)
		if err != nil {
			message.MyError(err, "resolveCurrentNode > net.Dial")
		}
		decoderLinkProcess := gob.NewDecoder(connLinkProcess)
		encoderLinkProcess := gob.NewEncoder(connLinkProcess)

		syncFormatClient("CurrentNode", decoderLinkProcess, encoderLinkProcess)

		if err := encoderLinkProcess.Encode(inputFormat); err != nil {
			message.MyError(err, "resolveCurrentNode > encoderLinkProcess.Encode")
		}

		// PSNodeへ

		// 受信する型は m2mapi.ResolveDataByNode
		results := m2mapi.ResolveDataByNode{}
		if err := decoderLinkProcess.Decode(&results); err != nil {
			message.MyError(err, "resolveCurrentNode > decoderLinkProcess.Decode")
		}

		// 最後にM2M APIへ返送
		jsonData, err := json.Marshal(results)
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

		// PSNodeからの定期的なセンサデータ登録で受信するセンサデータを読み込み，Conditionと合致する内容であれば，M2M APIへ返送する
		inputPNodeID := convertID(inputFormat.VNodeID, 63, 61)
		buffer_data := bufferSensorData[inputPNodeID]
		val := buffer_data.Value

		lowerLimit := inputFormat.Condition.Limit.LowerLimit
		upperLimit := inputFormat.Condition.Limit.UpperLimit
		timeout := inputFormat.Condition.Timeout

		timeoutContext, cancelFunc := context.WithTimeout(context.Background(), timeout)
		defer cancelFunc()

		fmt.Println("Wait for Condition Data...")
	Loop:
		for {
			select {
			case <-timeoutContext.Done():
				fmt.Println("Timeout Deadline")
				nullData := m2mapi.ResolveDataByNode{
					VNodeID: "Timeout",
				}
				jsonData, err := json.Marshal(nullData)
				if err != nil {
					http.Error(w, "resolveConditionNode: Error marshaling data", http.StatusInternalServerError)
					break Loop
				}
				fmt.Fprintf(w, "%v\n", string(jsonData))
				return
			case <-buffer_chan:
				mu.Lock()
				if val != bufferSensorData[inputPNodeID].Value {
					// バッファデータ更新
					val = bufferSensorData[inputPNodeID].Value
				}
				mu.Unlock()

				if val >= lowerLimit && val < upperLimit {
					// 条件を満たすので，M2M APIへ結果を転送
					register_data := bufferSensorData[inputPNodeID]
					values := []m2mapi.Value{}
					value := m2mapi.Value{
						Capability: register_data.Capability,
						Time:       register_data.Timestamp,
						Value:      register_data.Value,
					}
					values = append(values, value)
					data := m2mapi.ResolveDataByNode{
						Values: values,
					}
					jsonData, err := json.Marshal(data)
					if err != nil {
						http.Error(w, "resolveConditionNode: Error marshaling data", http.StatusInternalServerError)
						return
					}
					fmt.Fprintf(w, "%v\n", string(jsonData))
					bufferSensorData[inputPNodeID] = m2mapi.DataForRegist{}
					return
				} else {
					continue Loop
				}
			}
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
	var wg sync.WaitGroup

	// 初期環境構築時に作成したVSNodeのポート分だけ必要
	initial_environment_file := os.Getenv("HOME") + os.Getenv("PROJECT_PATH") + "/VSNode/initial_environment.json"
	file, err := os.Open(initial_environment_file)
	if err != nil {
		fmt.Println("Error opening file: ", err)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file: ", err)
		return
	}

	var ports Ports
	err = json.Unmarshal(data, &ports)
	if err != nil {
		fmt.Println("Error decoding JSON: ", err)
		return
	}

	for _, port := range ports.Port {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			startServer(port)
		}(port)
	}

	wg.Wait()
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
