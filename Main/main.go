package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"mecm2m-Emulator-PMNode/pkg/m2mapi"
	"mecm2m-Emulator-PMNode/pkg/message"
	"mecm2m-Emulator-PMNode/pkg/psnode"

	"github.com/joho/godotenv"
)

const (
	protocol           = "unix"
	timeSock           = "/tmp/mecm2m/time.sock"
	data_send_interval = 10
	layout             = "2006-01-02 15:04:05 +0900 JST"
)

type Format struct {
	FormType string
}

type Ports struct {
	Port []int `json:"ports"`
}

var vmnode_ip_port string

func init() {
	// .envファイルの読み込み
	if err := godotenv.Load(os.Getenv("HOME") + "/.env"); err != nil {
		log.Fatal(err)
	}
	vmnode_ip_port = os.Getenv("VMNODE_IP_PORT")
}

func main() {
	if err := godotenv.Load(os.Getenv("HOME") + "/.env"); err != nil {
		fmt.Println("There is no ~/.env file")
	}

	// 1. 各プロセスファイルの実行
	// 各プロセスのPIDを格納

	processIds := []int{}
	/*
		// 2-1. M2M API の実行
		fmt.Println("----------")
		m2m_api_path := os.Getenv("HOME") + os.Getenv("PROJECT_NAME") + "/M2MAPI/main"
		cmdM2MAPI := exec.Command(m2m_api_path)
		errCmdM2MAPI := cmdM2MAPI.Start()
		if errCmdM2MAPI != nil {
			message.MyError(errCmdM2MAPI, "exec.Command > M2M API > Start")
		} else {
			fmt.Println("M2M API is running")
		}
		processIds = append(processIds, cmdM2MAPI.Process.Pid)

		// 2-2. Local Manager の実行

		// 2-3. Local AAA の実行

		// 2-4. Local Repository の実行

		// 2-5. VSNode の実行
		fmt.Println("----------")
		vsnode_path := os.Getenv("HOME") + os.Getenv("PROJECT_NAME") + "/VSNode/main"
		cmdVSNode := exec.Command(vsnode_path)
		errCmdVSNode := cmdVSNode.Start()
		if errCmdVSNode != nil {
			message.MyError(errCmdVSNode, "exec.Command > VSNode > Start")
		} else {
			fmt.Println("VSNode is running")
		}
		processIds = append(processIds, cmdVSNode.Process.Pid)

		// 2-6. VMNode の実行

		// 2-7. PSNode の実行
		fmt.Println("----------")
		psnode_path := os.Getenv("HOME") + os.Getenv("PROJECT_NAME") + "/PSNode/main"
		cmdPSNode := exec.Command(psnode_path)
		errCmdPSNode := cmdPSNode.Start()
		if errCmdPSNode != nil {
			message.MyError(errCmdPSNode, "exec.Command > PSNode > Start")
		} else {
			fmt.Println("PSNode is running")
		}
		processIds = append(processIds, cmdPSNode.Process.Pid)

		fmt.Println("Process pid: ", processIds)
	*/
	// 2. Global/Local SensingDB のsensordataテーブルを作成する
	create_sensing_db_table_index := os.Getenv("HOME") + os.Getenv("PROJECT_NAME") + "/setup/SensingDB/create_db_table_index.py"
	cmdSensingDB := exec.Command("python3", create_sensing_db_table_index)
	errCmdSensingDB := cmdSensingDB.Run()
	if errCmdSensingDB != nil {
		message.MyError(errCmdSensingDB, "start > cmdSensingDB.Run")
	}

	// main()を走らす前に，startコマンドを入力することで，各プロセスにシグナルを送信する

	inputChan := make(chan string)
	vsnode_initial_environment_file := os.Getenv("HOME") + os.Getenv("PROJECT_PATH") + "/VSNode/initial_environment.json"
	file, err := os.Open(vsnode_initial_environment_file)
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

	var vsnode_ports Ports
	err = json.Unmarshal(data, &vsnode_ports)
	if err != nil {
		fmt.Println("Error decoding JSON: ", err)
		return
	}

	// 物理デバイスが定期的にセンサデータ登録するための時刻配布
	sliceLength := len(vsnode_ports.Port)
	numSlices := 29
	subSliceLength := sliceLength / numSlices
	subSlices := make([][]int, numSlices)
	for i := 0; i < numSlices; i++ {
		start := i * subSliceLength
		end := (i + 1) * subSliceLength
		if i == numSlices-1 {
			end = sliceLength
		}
		subSlices[i] = vsnode_ports.Port[start:end]
	}

	//go ticker(inputChan, subSlices)
	go mobility(inputChan, subSlices)

	// シミュレータ開始前
	reader := bufio.NewReader(os.Stdin)
	for {
		// クライアントがコマンドを入力
		var command string
		fmt.Printf("Inactive > ")
		command, _ = reader.ReadString('\n')

		// 何も入力がなければそのままcontinue
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}

		// シミュレーション前のコマンドを制御
		if commandExecutionBeforeEmulator(command, processIds, inputChan) {
			break
		} else {
			continue
		}
	}

	// シミュレータ開始後
	for {
		// クライアントがコマンドを入力
		var command string
		fmt.Printf("Simulating > ")
		command, _ = reader.ReadString('\n')

		// 何も入力がなければそのままcontinue
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}

		options := loadInput(command)
		sockAddr := selectSocketFile(command)
		if len(options) == 0 && sockAddr == "" {
			// コマンドが見つからない
			fmt.Println(command, ": command not found")
			continue
		} else if sockAddr == "basic command" {
			commandExecutionAfterEmulator(command, processIds)
			continue
		}
		conn, err := net.Dial(protocol, sockAddr)
		if err != nil {
			message.MyError(err, "main > net.Dial")
		}

		decoder := gob.NewDecoder(conn)
		encoder := gob.NewEncoder(conn)
		// commandExecutionする前に，Server側と型の同期を取りたい
		syncFormatClient(command, decoder, encoder)
		// APIを叩くコマンドを制御
		commandAPIExecution(command, decoder, encoder, options, processIds)
	}
}

func ticker(inputChan chan string, subSlices [][]int) {
	<-inputChan
	// 時間間隔指定
	// センサデータ登録の時間間隔を一定の閾値を設けてランダムに設定．PSNodeごとに違う時間間隔を設けたい（未完成）
	t := time.NewTicker(time.Duration(data_send_interval) * time.Second)
	defer t.Stop()

	// シグナル受信用チャネル
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(sig)

	/*
		for {
			var wg sync.WaitGroup
			select {
			case now := <-t.C:
				// 現在時刻(now)の送信
				for _, port := range psnode_ports {
					wg.Add(1)
					go func(port int) {
						defer wg.Done()
						port_str := strconv.Itoa(port)
						pnode_id := trimPNodeID(port)
						send_data := psnode.TimeSync{
							PNodeID:     pnode_id,
							CurrentTime: now,
						}
						url := "http://localhost:" + port_str + "/time"
						client_data, err := json.Marshal(send_data)
						if err != nil {
							fmt.Println("Error marshaling data: ", err)
							return
						}
						response, err := http.Post(url, "application/json", bytes.NewBuffer(client_data))
						if err != nil {
							fmt.Println("Error making request: ", err)
							return
						}
						defer response.Body.Close()
					}(port)
				}
			// シグナルを受信した場合
			case s := <-sig:
				switch s {
				case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
					fmt.Println("Stop!")
					return
				}
			}
		}
	*/
	var wg sync.WaitGroup
	for _, slices := range subSlices {
		for _, port := range slices {
			wg.Add(1)
			go func(port int) {
				defer wg.Done()
				port_str := strconv.Itoa(port)
				pnode_id := trimPNodeID(port)
				send_data := psnode.TimeSync{
					PNodeID:     pnode_id,
					CurrentTime: time.Now(),
				}
				sensordata := generateSensordata(&send_data)
				url := "http://localhost:" + port_str + "/data/register"
				client_data, err := json.Marshal(sensordata)
				if err != nil {
					fmt.Println("Error marshaling data: ", err)
					return
				}
				response, err := http.Post(url, "application/json", bytes.NewBuffer(client_data))
				if err != nil {
					fmt.Println("Error making request: ", err)
					return
				}
				defer response.Body.Close()
			}(port)
		}
		time.Sleep(1 * time.Second)
	}
	wg.Wait()

}

func mobility(inputChan chan string, subSlices [][]int) {
	<-inputChan
	var wg sync.WaitGroup
	for _, slices := range subSlices {
		for _, port := range slices {
			wg.Add(1)
			go func(port int) {
				defer wg.Done()
				send_data := psnode.Mobility{
					PNodeID: os.Getenv("PMNODE_ID"),
				}
				url := "http://" + vmnode_ip_port + "/mobility"
				client_data, err := json.Marshal(send_data)
				if err != nil {
					fmt.Println("Error marshaling data: ", err)
					return
				}
				response, err := http.Post(url, "application/json", bytes.NewBuffer(client_data))
				if err != nil {
					fmt.Println("Error making request: ", err)
					return
				}
				defer response.Body.Close()
			}(port)
			time.Sleep(1000 * time.Millisecond)
		}
	}
	wg.Wait()
}

// 入力したコマンドに対応するAPIの入力内容を取得
func loadInput(command string) []string {
	var file string
	var options []string
	switch command {
	case "point":
		// SWLat,SWLon,NELat,NELon
		file = "option_file/m2m_api/point.csv"
	case "node":
		// VPointID_n, Caps
		file = "option_file/m2m_api/node.csv"
	case "past_node":
		// VNodeID_n, Cap, Period{Start, End}
		file = "option_file/m2m_api/past_node.csv"
	case "past_point":
		// VPointID_n, Cap, Period{Start, End}
		file = "option_file/m2m_api/past_point.csv"
	case "current_node":
		// VNodeID_n, Cap
		file = "option_file/m2m_api/current_node.csv"
	case "current_point":
		// VPointID_n, Cap
		file = "option_file/m2m_api/current_point.csv"
	case "condition_node":
		// VNodeID_n, Cap, LowerLimit, UpperLimit, Timeout(s)
		file = "option_file/m2m_api/condition_node.csv"
	case "condition_point":
		// VPointID_n, Cap, LowerLimit, UpperLimit, Timeout(s)
		file = "option_file/m2m_api/condition_point.csv"
	case "actuate":
		// VNodeID_n, Action, Parameter
		file = "option_file/m2m_api/actuate.csv"
	case "exit", "register", "help":
		options = append(options, "basic command")
		return options
	default:
		return options
	}
	fp, err := os.Open(file)
	if err != nil {
		message.MyError(err, "loadInput > os.Open")
	}
	defer fp.Close()
	r := csv.NewReader(fp)
	rows, err := r.ReadAll()
	if err != nil {
		message.MyError(err, "loadInput > r.ReadAll")
	}
	options = rows[0]
	return options
}

// 入力したコマンドに対応するソケットファイルを選択
func selectSocketFile(command string) string {
	var sockAddr string
	defaultAddr := "/tmp/mecm2m"
	defaultExt := ".sock"
	switch command {
	case "point", "node", "past_node", "past_point", "current_node", "current_point", "condition_node", "condition_point", "actuate":
		sockAddr = defaultAddr + "/svr_1_m2mapi" + defaultExt
	case "exit", "register", "help":
		sockAddr = "basic command"
	default:
		sockAddr = ""
	}
	return sockAddr
}

func syncFormatClient(command string, decoder *gob.Decoder, encoder *gob.Encoder) {
	format := &Format{}
	switch command {
	case "point":
		format.FormType = "Point"
	case "node":
		format.FormType = "Node"
	case "past_node":
		format.FormType = "PastNode"
	case "past_point":
		format.FormType = "PastPoint"
	case "current_node":
		format.FormType = "CurrentNode"
	case "current_point":
		format.FormType = "CurrentPoint"
	case "condition_node":
		format.FormType = "ConditionNode"
	case "condition_point":
		format.FormType = "ConditionPoint"
	case "actuate":
		format.FormType = "Actuate"
	}
	if err := encoder.Encode(format); err != nil {
		message.MyError(err, "syncFormatClient > "+command+" > encoder.Encode")
	}
}

func trimPNodeID(port int) string {
	base_port, _ := strconv.Atoi(os.Getenv("VSNODE_BASE_PORT"))
	id_index := port - base_port
	pnode_id_int := int(0b0010<<60) + id_index
	pnode_id := strconv.Itoa(pnode_id_int)
	return pnode_id
}

// センサデータの登録
func generateSensordata(inputFormat *psnode.TimeSync) psnode.DataForRegist {
	var result psnode.DataForRegist
	// PSNodeのconfigファイルを検索し，ソケットファイルと一致する情報を取得する
	psnode_json_file_path := os.Getenv("HOME") + os.Getenv("PROJECT_NAME") + "/setup/GraphDB/config/config_main_psnode.json"
	psnodeJsonFile, err := os.Open(psnode_json_file_path)
	if err != nil {
		fmt.Println(err)
	}
	defer psnodeJsonFile.Close()
	psnodeByteValue, _ := io.ReadAll(psnodeJsonFile)

	var psnodeResult map[string][]interface{}
	json.Unmarshal(psnodeByteValue, &psnodeResult)

	psnodes := psnodeResult["psnodes"]
	for _, v := range psnodes {
		psnode_format := v.(map[string]interface{})
		psnode := psnode_format["psnode"].(map[string]interface{})
		//psnode_relation_label := psnode["relation-label"].(map[string]interface{})
		psnode_data_property := psnode["data-property"].(map[string]interface{})
		pnode_id := psnode_data_property["PNodeID"].(string)
		if pnode_id == inputFormat.PNodeID {
			result.PNodeID = pnode_id
			result.Capability = psnode_data_property["Capability"].(string)
			result.Timestamp = inputFormat.CurrentTime.Format(layout)
			randomFloat := randomFloat64()
			min := 30.0
			//max := 40.0
			value_value := min + randomFloat
			result.Value = value_value
			//result.PSinkID = psnode_relation_label["PSink"].(string)
			result.PSinkID = "PSink"
			position := psnode_data_property["Position"].([]interface{})
			result.Lat = position[0].(float64)
			result.Lon = position[1].(float64)
		}
	}
	return result
}

func randomFloat64() float64 {
	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		panic(err)
	}
	floatValue := new(big.Float).SetInt(n)
	float64Value, _ := floatValue.Float64()
	f := float64Value / 100
	return f
}

// APIを叩く以外のコマンドの実行 (シミュレーション実行，exit, デバイスの登録)
func commandExecutionBeforeEmulator(command string, processIds []int, inputChan chan string) bool {
	flag := false
	switch command {
	// シミュレーションの実行
	case "start":
		inputChan <- "start"
		for _, pid := range processIds {
			process, err := os.FindProcess(pid)
			if err != nil {
				message.MyError(err, "commandExecutionBeforeEmulator > start > os.FindProcess")
			}

			if err := process.Signal(syscall.Signal(syscall.SIGCONT)); err != nil {
				message.MyError(err, "commandExecutionBeforeEmulator > start > process.Signal")
			}
		}
		message.MyMessage("		*** Emulator is running *** 	")
		flag = true
		return flag
	// シミュレーションの終了
	case "exit":
		// 各プロセスの削除

		for _, pid := range processIds {
			process, err := os.FindProcess(pid)
			if err != nil {
				message.MyError(err, "commandExecutionBeforeEmulator > exit > os.FindProcess")
			}

			err = process.Signal(os.Interrupt)
			if err != nil {
				message.MyError(err, "commandExecutionBeforeEmulator > exit > process.Signal")
			} else {
				fmt.Printf("process (%d) is killed\n", pid)
			}
		}

		message.MyMessage("Bye")
		os.Exit(0)
	// デバイスの登録
	case "register":
		// あらかじめ，登録するデバイスの情報をconfigファイルとして用意しておく
		// ./option_file/register/register_psnode.py を実行
		// 2023-05-10 一旦PSNodeの登録のみ
		register_psnode_script := "./option_file/register/register_psnode.py"
		register_psnode_file := "./option_file/register/register_psnode.json"
		cmdGraphDB := exec.Command("python3", register_psnode_script, register_psnode_file)
		errCmdGraphDB := cmdGraphDB.Run()
		if errCmdGraphDB != nil {
			message.MyError(errCmdGraphDB, "commandExecutionBeforeEmulator > register > cmdGraphDB.Run")
		}

		message.MyMessage("		*** Succeed: Register device *** 	")
		return flag
	// helpコマンド
	case "help":
		fmt.Println("[start]: 	Start MECM2M Emulator")
		fmt.Println("[exit]: 	Exit MECM2M Emulator")
		fmt.Println("[register]: 	Register PSNode")
		return flag
	default:
		// コマンドが見つからない
		fmt.Println(command, ": command not found")
		return flag
	}

	return flag
}

// APIを叩く以外のコマンドの実行 (シミュレーション実行，exit, デバイスの登録)
func commandExecutionAfterEmulator(command string, processIds []int) {
	switch command {
	// シミュレーションの終了
	case "exit":
		// 各プロセスの削除

		for _, pid := range processIds {
			process, err := os.FindProcess(pid)
			if err != nil {
				message.MyError(err, "commandExecutionBeforeEmulator > exit > os.FindProcess")
			}

			err = process.Signal(os.Interrupt)
			if err != nil {
				message.MyError(err, "commandExecutionBeforeEmulator > exit > process.Signal")
			} else {
				fmt.Printf("process (%d) is killed\n", pid)
			}
		}

		message.MyMessage("Bye")
		os.Exit(0)
	// デバイスの登録
	case "register":
		// あらかじめ，登録するデバイスの情報をconfigファイルとして用意しておく
		// ./option_file/register/register_psnode.py を実行
		// 2023-05-10 一旦PSNodeの登録のみ
		// 本来はこの場所ではない．APIExecutionに入れる (2023-05-25)
		register_psnode_script := "./option_file/register/register_psnode.py"
		register_psnode_file := "./option_file/register/register_psnode.json"
		cmdGraphDB := exec.Command("python3", register_psnode_script, register_psnode_file)
		errCmdGraphDB := cmdGraphDB.Run()
		if errCmdGraphDB != nil {
			message.MyError(errCmdGraphDB, "commandExecutionAfterEmulator > register > cmdGraphDB.Run")
		}

		message.MyMessage("		*** Succeed: Register device *** 	")
	// helpコマンド
	case "help":
		fmt.Println("[exit]: 		Exit MECM2M Emulator")
		fmt.Println("[register]: 		Register PSNode")
		fmt.Println("[point]: 		Resolve Point")
		fmt.Println("[node]: 		Resolve Node")
		fmt.Println("[past_node]: 		Resolve Past Data By Node")
		fmt.Println("[current_node]: 	Resolve Current Data By Node")
		fmt.Println("[past_point]: 		Resolbe Past Data By Point")
		fmt.Println("[current_point]: 	Resolve Current Data By Point")
		fmt.Println("[actuate]:		Actuate a Node")
	default:
		// コマンドが見つからない
		fmt.Println(command, ": command not found")
	}
}

// APIを叩くコマンドの実行
func commandAPIExecution(command string, decoder *gob.Decoder, encoder *gob.Encoder, options []string, processIds []int) {
	switch command {
	case "point":
		var swlat, swlon, nelat, nelon float64
		swlat, _ = strconv.ParseFloat(options[0], 64)
		swlon, _ = strconv.ParseFloat(options[1], 64)
		nelat, _ = strconv.ParseFloat(options[2], 64)
		nelon, _ = strconv.ParseFloat(options[3], 64)
		point_input := &m2mapi.ResolveArea{
			SW: m2mapi.SquarePoint{Lat: swlat, Lon: swlon},
			NE: m2mapi.SquarePoint{Lat: nelat, Lon: nelon},
		}
		if err := encoder.Encode(point_input); err != nil {
			message.MyError(err, "commandAPIExecution > point > encoder.Encode")
		}
		message.MyWriteMessage(*point_input)

		// ポイント解決の結果を受信する (PsinkのVPointID_n，Address)
		point_output := []m2mapi.ResolveArea{}
		if err := decoder.Decode(&point_output); err != nil {
			message.MyError(err, "commandAPIExecution > point > decoder.Decode")
		}
		message.MyReadMessage(point_output)
	case "node":
		var VPointID string
		Caps := make([]string, len(options)-1)
		for i, option := range options {
			if i == 0 {
				VPointID = option
			} else {
				Caps[i-1] = option
			}
		}
		node_input := &m2mapi.ResolveNode{
			NodeType: VPointID,
		}
		if err := encoder.Encode(node_input); err != nil {
			message.MyError(err, "commandAPIExecution > node > encoder.Encode")
		}
		message.MyWriteMessage(*node_input)

		// ノード解決の結果を受信する（PNodeのVNodeID_n, Cap）
		node_output := []m2mapi.ResolveNode{}
		if err := decoder.Decode(&node_output); err != nil {
			message.MyError(err, "commandAPIExecution > node > decoder.Decode")
		}
		message.MyReadMessage(node_output)
	case "past_node":
		var VNodeID, Capability, Start, End, SocketAddress string
		VNodeID = options[0]
		Capability = options[1]
		Start = options[2]
		End = options[3]
		SocketAddress = options[4]
		past_node_input := &m2mapi.ResolveDataByNode{
			VNodeID:       VNodeID,
			Capability:    []string{Capability},
			Period:        m2mapi.PeriodInput{Start: Start, End: End},
			SocketAddress: SocketAddress,
		}
		if err := encoder.Encode(past_node_input); err != nil {
			message.MyError(err, "commandAPIExecution > past_node > encoder.Encode")
		}
		message.MyWriteMessage(*past_node_input)

		// ノードの過去データ解決を受信する（VNodeID_n, Value, Cap, Time）
		past_node_output := m2mapi.ResolveDataByNode{}
		if err := decoder.Decode(&past_node_output); err != nil {
			message.MyError(err, "commandAPIExecution > past_node > decoder.Decode")
		}
		message.MyReadMessage(past_node_output)
	case "past_point":
		var VPointID_n, Capability, Start, End, SocketAddress string
		VPointID_n = options[0]
		Capability = options[1]
		Start = options[2]
		End = options[3]
		SocketAddress = options[4]
		past_point_input := &m2mapi.ResolveDataByNode{
			VNodeID:       VPointID_n,
			Capability:    []string{Capability},
			Period:        m2mapi.PeriodInput{Start: Start, End: End},
			SocketAddress: SocketAddress,
		}
		if err := encoder.Encode(past_point_input); err != nil {
			message.MyError(err, "commandAPIExecution > past_point > encoder.Encode")
		}
		message.MyWriteMessage(*past_point_input)

		// ポイントの過去データ解決を受信する（VNodeID_n, Value, Cap, Time）
		past_point_output := m2mapi.ResolveDataByNode{}
		if err := decoder.Decode(&past_point_output); err != nil {
			message.MyError(err, "commandAPIExecution > past_point > decoder.Decode")
		}
		message.MyReadMessage(past_point_output)
	case "current_node":
		var VNodeID, Capability, SocketAddress string
		VNodeID = options[0]
		Capability = options[1]
		SocketAddress = options[2]
		current_node_input := &m2mapi.ResolveDataByNode{
			VNodeID:       VNodeID,
			Capability:    []string{Capability},
			SocketAddress: SocketAddress,
		}
		if err := encoder.Encode(current_node_input); err != nil {
			message.MyError(err, "commandAPIExecution > current_node > encoder.Encode")
		}
		message.MyWriteMessage(current_node_input)

		// ノードの現在データ解決を受信する（Value, Cap, Time)
		current_node_output := m2mapi.ResolveDataByNode{}
		if err := decoder.Decode(&current_node_output); err != nil {
			message.MyError(err, "commandAPIExecution > current_node > decoder.Decode")
		}
		message.MyReadMessage(current_node_output)
	case "current_point":
		var VPointID_n, Capability, SocketAddress string
		VPointID_n = options[0]
		Capability = options[1]
		SocketAddress = options[2]
		current_point_input := &m2mapi.ResolveDataByNode{
			VNodeID:       VPointID_n,
			Capability:    []string{Capability},
			SocketAddress: SocketAddress,
		}
		if err := encoder.Encode(current_point_input); err != nil {
			message.MyError(err, "commandAPIExecution > current_point > encoder.Encode")
		}
		message.MyWriteMessage(current_point_input)

		// ポイントの現在データ解決を受信する（VNodeID_n, Value, Cap, Time）
		current_point_output := m2mapi.ResolveDataByNode{}
		if err := decoder.Decode(&current_point_output); err != nil {
			message.MyError(err, "commandAPIExecution > current_point > decoder.Decode")
		}
		message.MyReadMessage(current_point_output)
	case "condition_node":
		var VNodeID_n, Capability, SocketAddress string
		var LowerLimit, UpperLimit float64
		var Timeout time.Duration
		VNodeID_n = options[0]
		Capability = options[1]
		LowerLimit, _ = strconv.ParseFloat(options[2], 64)
		UpperLimit, _ = strconv.ParseFloat(options[3], 64)
		Timeout, _ = time.ParseDuration(options[4])
		SocketAddress = options[5]
		condition_node_input := &m2mapi.ResolveDataByNode{
			VNodeID:       VNodeID_n,
			Capability:    []string{Capability},
			Condition:     m2mapi.ConditionInput{Limit: m2mapi.Range{LowerLimit: LowerLimit, UpperLimit: UpperLimit}, Timeout: Timeout},
			SocketAddress: SocketAddress,
		}
		if err := encoder.Encode(condition_node_input); err != nil {
			message.MyError(err, "commandAPIExecution > condition_node > encoder.Encode")
		}
		message.MyWriteMessage(condition_node_input)

		// ノードの現在データを受信する（Value, Cap, Time）
		condition_node_output := m2mapi.DataForRegist{}
		if err := decoder.Decode(&condition_node_output); err != nil {
			message.MyError(err, "commandAPIExecution > condition_node > decoder.Decode")
		}
		message.MyReadMessage(condition_node_output)
	case "condition_point":
		var VPointID_n, Capability, SocketAddress string
		var LowerLimit, UpperLimit float64
		var Timeout time.Duration
		VPointID_n = options[0]
		Capability = options[1]
		LowerLimit, _ = strconv.ParseFloat(options[2], 64)
		UpperLimit, _ = strconv.ParseFloat(options[3], 64)
		Timeout, _ = time.ParseDuration(options[4])
		SocketAddress = options[5]
		condition_point_input := &m2mapi.ResolveDataByNode{
			VNodeID:       VPointID_n,
			Capability:    []string{Capability},
			Condition:     m2mapi.ConditionInput{Limit: m2mapi.Range{LowerLimit: LowerLimit, UpperLimit: UpperLimit}, Timeout: Timeout},
			SocketAddress: SocketAddress,
		}
		if err := encoder.Encode(condition_point_input); err != nil {
			message.MyError(err, "commandAPIExecution > condition_point > encoder.Encode")
		}
		message.MyWriteMessage(condition_point_input)

		// ノードの現在データを受信する（Value, Cap, Time）
		condition_point_output := m2mapi.DataForRegist{}
		if err := decoder.Decode(&condition_point_output); err != nil {
			message.MyError(err, "commandAPIExecution > condition_point > decoder.Decode")
		}
		message.MyReadMessage(condition_point_output)
	case "actuate":
		var VNodeID_n, Action, SocketAddress string
		var Parameter float64
		VNodeID_n = options[0]
		Action = options[1]
		Parameter, _ = strconv.ParseFloat(options[2], 64)
		SocketAddress = options[3]
		actuate_input := &m2mapi.Actuate{
			VNodeID:       VNodeID_n,
			Action:        Action,
			Parameter:     Parameter,
			SocketAddress: SocketAddress,
		}
		if err := encoder.Encode(actuate_input); err != nil {
			message.MyError(err, "commandAPIExecution > actuate > encoder.Encode")
		}
		message.MyWriteMessage(actuate_input)

		// アクチュエートによる状態を受信する (Status)
		actuate_output := m2mapi.Actuate{}
		if err := decoder.Decode(&actuate_output); err != nil {
			message.MyError(err, "commandAPIExecution > actuate > decoder.Decode")
		}
		message.MyReadMessage(actuate_output)
	default:
		// コマンドが見つからない
		fmt.Println(command, ": command not found")
	}
}
