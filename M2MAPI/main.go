package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"unsafe"

	"mecm2m-Emulator-PMNode/pkg/m2mapi"
	"mecm2m-Emulator-PMNode/pkg/m2mapp"
	"mecm2m-Emulator-PMNode/pkg/message"

	"github.com/joho/godotenv"
)

var (
	port                         string
	cloud_server_ip_port         string
	ip_address                   string
	connected_mec_server_ip_port string
	ad_cache                     map[string]m2mapi.AreaDescriptor = make(map[string]m2mapi.AreaDescriptor)
	vmnoder_port                 string
	vmnode_id                    string
)

func init() {
	// .envファイルの読み込み
	if err := godotenv.Load(os.Getenv("HOME") + "/.env"); err != nil {
		log.Fatal(err)
	}
	port = os.Getenv("M2M_API_PORT")
	ip_address = os.Getenv("IP_ADDRESS")
	cloud_server_ip_port = os.Getenv("CLOUD_SERVER_IP_PORT")
	connected_mec_server_ip_port = os.Getenv("CONNECTED_MEC_SERVER_IP_ADDRESS") + ":" + port
	vmnoder_port = os.Getenv("VMNODER_PORT")
	vmnode_id = os.Getenv("VMNODE_ID")
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
	http.HandleFunc("/m2mapi/area", resolveArea)
	http.HandleFunc("/m2mapi/area/extend", extendAD)
	http.HandleFunc("/m2mapi/node", resolveNode)
	http.HandleFunc("/m2mapi/data/past/node", resolvePastNode)
	http.HandleFunc("/m2mapi/data/current/node", resolveCurrentNode)
	http.HandleFunc("/m2mapi/data/condition/node", resolveConditionNode)
	http.HandleFunc("/m2mapi/data/past/area", resolvePastArea)
	http.HandleFunc("/m2mapi/data/current/area", resolveCurrentArea)
	http.HandleFunc("/m2mapi/data/condition/area", resolveConditionArea)
	http.HandleFunc("/m2mapi/actuate", actuate)
	http.HandleFunc("/hello", hello)

	log.Printf("Connect to http://%s:%s/ for M2M API", ip_address, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World\n")
}

func resolveArea(w http.ResponseWriter, r *http.Request) {
	// POST リクエストのみを受信する
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "resolvePoint: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapp.ResolveAreaInput{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "resolvePoint: Error missmatching packet format", http.StatusInternalServerError)
		}

		// GraphDBへの問い合わせ
		results := resolveAreaFunction(inputFormat.SW, inputFormat.NE)

		fmt.Fprintf(w, "%v\n", results)
	} else {
		http.Error(w, "resolvePoint: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
	fmt.Println(ad_cache)
}

func extendAD(w http.ResponseWriter, r *http.Request) {
	// POST リクエストのみを受信する
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "extendAD: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapi.ExtendAD{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "extendAD: Error missmatching packet format", http.StatusInternalServerError)
		}

		output := m2mapi.ExtendAD{}
		if value, ok := ad_cache[inputFormat.AD]; ok {
			for _, ad_detail := range value.AreaDescriptorDetail {
				ad_detail.TTL.Add(1 * time.Hour)
			}
			output.Flag = true
		} else {
			output.Flag = false
		}

		fmt.Fprintf(w, "%v\n", output)
	} else {
		http.Error(w, "extendAD: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func resolveNode(w http.ResponseWriter, r *http.Request) {
	// POST リクエストのみを受信する
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "resolveNode: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapp.ResolveNodeInput{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "resolveNode: Error missmatching packet format", http.StatusInternalServerError)
		}

		// GraphDBへの問い合わせ
		results := resolveNodeFunction(inputFormat.AD, inputFormat.Capability, inputFormat.NodeType)

		fmt.Fprintf(w, "%v\n", results)
	} else {
		http.Error(w, "resolveNode: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func resolvePastNode(w http.ResponseWriter, r *http.Request) {
	// POST リクエストのみを受信する
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "resolvePastNode: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapp.ResolveDataByNodeInput{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "resolvePastNode: Error missmatching packet format", http.StatusInternalServerError)
		}

		// VNodeへリクエスト転送
		m2mapi_results := resolvePastNodeFunction(inputFormat.VNodeID, inputFormat.SocketAddress, inputFormat.Capability, inputFormat.Period)
		// m2mapp用に成型
		results := m2mapp.ResolveDataByNodeOutput{}
		results.VNodeID = m2mapi_results.VNodeID
		for _, val := range m2mapi_results.Values {
			v := m2mapp.Value{
				Capability: val.Capability,
				Time:       val.Time,
				Value:      val.Value,
			}
			results.Values = append(results.Values, v)
		}

		fmt.Fprintf(w, "%v\n", results)
	} else {
		http.Error(w, "resolvePastNode: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func resolveCurrentNode(w http.ResponseWriter, r *http.Request) {
	// POST リクエストのみを受信する
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "resolveCurrentNode: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapp.ResolveDataByNodeInput{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "resolveCurrentNode: Error missmatching packet format", http.StatusInternalServerError)
		}

		// VNodeへリクエスト転送
		m2mapi_results := resolveCurrentNodeFunction(inputFormat.VNodeID, inputFormat.SocketAddress, inputFormat.Capability)
		// m2mapp用に成型
		results := m2mapp.ResolveDataByNodeOutput{}
		results.VNodeID = m2mapi_results.VNodeID
		for _, val := range m2mapi_results.Values {
			v := m2mapp.Value{
				Capability: val.Capability,
				Time:       val.Time,
				Value:      val.Value,
			}
			results.Values = append(results.Values, v)
		}

		fmt.Fprintf(w, "%v\n", results)
	} else {
		http.Error(w, "resolveCurrentNode: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func resolveConditionNode(w http.ResponseWriter, r *http.Request) {
	// POST リクエストのみを受信する
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "resolveConditionNode: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapp.ResolveDataByNodeInput{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "resolveConditionNode: Error missmatching packet format", http.StatusInternalServerError)
		}

		// VNode へリクエスト転送
		m2mapi_results := resolveConditionNodeFunction(inputFormat.VNodeID, inputFormat.SocketAddress, inputFormat.Capability, inputFormat.Condition)
		// m2mapp用に成型
		results := m2mapp.ResolveDataByNodeOutput{}
		results.VNodeID = m2mapi_results.VNodeID
		for _, val := range m2mapi_results.Values {
			v := m2mapp.Value{
				Capability: val.Capability,
				Time:       val.Time,
				Value:      val.Value,
			}
			results.Values = append(results.Values, v)
		}

		fmt.Fprintf(w, "%v\n", results)
	} else {
		http.Error(w, "resolveConditionNode: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func resolvePastArea(w http.ResponseWriter, r *http.Request) {
	// POST リクエストのみを受信する
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "resolvePastArea: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapi.ResolveDataByArea{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "resolvePastArea: Error missmatching packet format", http.StatusInternalServerError)
		}

		// VNode もしくは VMNode へリクエスト転送
		results := resolvePastAreaFunction(inputFormat.AD, inputFormat.Capability, inputFormat.NodeType, inputFormat.Period)

		fmt.Fprintf(w, "%v\n", results)
	} else {
		http.Error(w, "resolvePastArea: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func resolveCurrentArea(w http.ResponseWriter, r *http.Request) {
	// POST リクエストのみを受信する
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "resolveCurrentArea: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapi.ResolveDataByArea{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "resolveCurrentArea: Error missmatching packet format", http.StatusInternalServerError)
		}

		// VNode もしくは VMNode へリクエスト転送
		results := resolveCurrentAreaFunction(inputFormat.AD, inputFormat.Capability, inputFormat.NodeType)

		fmt.Fprintf(w, "%v\n", results)
	} else {
		http.Error(w, "resolveCurrentArea: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func resolveConditionArea(w http.ResponseWriter, r *http.Request) {
	// POST リクエストのみを受信する
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "resolveConditionArea: Error reading request body", http.StatusInternalServerError)
			return
		}
		inputFormat := &m2mapi.ResolveDataByArea{}
		if err := json.Unmarshal(body, inputFormat); err != nil {
			http.Error(w, "resolveConditionArea: Error missmatching packet format", http.StatusInternalServerError)
		}

		// VNode もしくは VMNode へリクエスト転送
		results := resolveConditionAreaFunction(inputFormat.AD, inputFormat.Capability, inputFormat.NodeType, inputFormat.Condition)

		fmt.Fprintf(w, "%v\n", results)
	} else {
		http.Error(w, "resolveCurrentArea: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func actuate(w http.ResponseWriter, r *http.Request) {
	// POST リクエストのみを受信する
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

		// VNode もしくは VMNode へリクエスト転送
		results := actuateFunction(inputFormat.VNodeID, inputFormat.Action, inputFormat.SocketAddress, inputFormat.Parameter)

		fmt.Fprintf(w, "%v\n", results)
	} else {
		http.Error(w, "actuate: Method not supported: Only POST request", http.StatusMethodNotAllowed)
	}
}

func resolveAreaFunction(sw, ne m2mapp.SquarePoint) m2mapp.ResolveAreaOutput {
	// 接続先MEC Serverに入力内容をそのまま転送する
	var results m2mapp.ResolveAreaOutput

	transmit_request := m2mapi.ResolveArea{
		SW:         m2mapi.SquarePoint{Lat: sw.Lat, Lon: sw.Lon},
		NE:         m2mapi.SquarePoint{Lat: ne.Lat, Lon: ne.Lon},
		PMNodeFlag: true,
	}
	transmit_data, _ := json.Marshal(transmit_request)
	mec_server_url := "http://" + connected_mec_server_ip_port + "/m2mapi/area"
	area_response, err := http.Post(mec_server_url, "application/json", bytes.NewBuffer(transmit_data))
	if err != nil {
		fmt.Println("Error making request: ", err)
	}
	defer area_response.Body.Close()

	body, err := io.ReadAll(area_response.Body)
	if err != nil {
		panic(err)
	}

	var area_output m2mapp.ResolveAreaOutput
	if err = json.Unmarshal(body, &area_output); err != nil {
		fmt.Println("Error Unmarshaling: ", err)
	}
	fmt.Println("area_output: ", area_output)

	var area_desc m2mapi.AreaDescriptor
	area_desc.AreaDescriptorDetail = make(map[string]m2mapi.AreaDescriptorDetail)
	area_desc.AreaDescriptorDetail = area_output.Descriptor.AreaDescriptorDetail

	ad := fmt.Sprintf("%x", uintptr(unsafe.Pointer(&area_desc)))
	ttl := time.Now().Add(1 * time.Hour)
	results.AD = ad
	results.TTL = ttl

	ad_cache[ad] = area_desc
	return results
}

func resolveNodeFunction(ad string, caps []string, node_type string) m2mapp.ResolveNodeOutput {
	// 接続先MEC Serverに入力内容をそのまま転送する
	area_desc := ad_cache[ad]
	transmit_request := m2mapi.ResolveNode{
		AreaDescriptorDetail: area_desc.AreaDescriptorDetail,
		Capability:           caps,
		NodeType:             node_type,
		PMNodeFlag:           true,
	}
	transmit_data, _ := json.Marshal(transmit_request)
	mec_server_url := "http://" + connected_mec_server_ip_port + "/m2mapi/node"
	node_response, err := http.Post(mec_server_url, "application/json", bytes.NewBuffer(transmit_data))
	if err != nil {
		fmt.Println("Error making request: ", err)
	}
	defer node_response.Body.Close()

	body, err := io.ReadAll(node_response.Body)
	if err != nil {
		panic(err)
	}

	var node_output m2mapp.ResolveNodeOutput
	if err = json.Unmarshal(body, &node_output); err != nil {
		fmt.Println("Error Unmarhsaling: ", err)
	}
	return node_output
}

func resolvePastNodeFunction(vnode_id, socket_address string, capability []string, period m2mapp.PeriodInput) m2mapi.ResolveDataByNode {
	null_data := m2mapi.ResolveDataByNode{VNodeID: "NULL"}
	var results m2mapi.ResolveDataByNode

	// 入力のVNodeIDが自身のVMNodeIDと一致するか比較する．
	// 一致すれば，自身のLocal GraphDBにVSNodeの検索をかける．一致しなければ，入力のSocketAddressに直接リクエスト転送する．

	if vnode_id == vmnode_id {
		var format_capability []string
		for _, cap := range capability {
			cap = "\\\"" + cap + "\\\""
			format_capability = append(format_capability, cap)
		}
		payload := `{"statements": [{"statement": "MATCH (vs:VSNode)-[:isPhysicalizedBy]->(ps:PSNode) WHERE ps.Capability IN [` + strings.Join(format_capability, ", ") + `] return vs.VNodeID, vs.SocketAddress;"}]}`
		graphdb_url := "http://" + os.Getenv("NEO4J_USERNAME") + ":" + os.Getenv("NEO4J_LOCAL_PASSWORD") + "@localhost:" + os.Getenv("NEO4J_LOCAL_PORT_GOLANG") + "/db/data/transaction/commit"
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
		var vsnode_set_own m2mapi.VNodeSet
		for _, v1 := range values {
			for k2, v2 := range v1.(map[string]interface{}) {
				if k2 == "data" {
					for _, v3 := range v2.([]interface{}) {
						for k4, v4 := range v3.(map[string]interface{}) {
							if k4 == "row" {
								row_data = v4
								dataArray := row_data.([]interface{})
								vsnode_set_own.VNodeID = dataArray[0].(string)
								vsnode_set_own.VNodeSocketAddress = dataArray[1].(string)
							}
						}
					}
				}
			}
		}
		// ここで，vsnode_set_own に格納されるVNodeの情報は1つだけであるという前提（ノード指定型データ取得だから）
		// マッチする情報が得られなかった場合，この時点でレスポンスする
		if vsnode_set_own.VNodeID == "" {
			return null_data
		}

		transmit_request := m2mapi.ResolveDataByNode{
			VNodeID:       vsnode_set_own.VNodeID,
			Capability:    capability,
			Period:        m2mapi.PeriodInput{Start: period.Start, End: period.End},
			SocketAddress: vsnode_set_own.VNodeSocketAddress,
		}
		transmit_data, err := json.Marshal(transmit_request)
		if err != nil {
			fmt.Println("Error marshaling data: ", err)
			return null_data
		}
		vmnoder_url := "http://localhost:" + vmnoder_port + "/primapi/data/past/node"
		response_data, err := http.Post(vmnoder_url, "application/json", bytes.NewBuffer(transmit_data))
		if err != nil {
			fmt.Println("Error making request:", err)
			return null_data
		}
		defer response_data.Body.Close()

		byteArray, _ = io.ReadAll(response_data.Body)
		if err = json.Unmarshal(byteArray, &results); err != nil {
			fmt.Println("Error unmarshaling data: ", err)
			return null_data
		}
	} else {
		transmit_request := m2mapi.ResolveDataByNode{
			VNodeID:    vnode_id,
			Capability: capability,
			Period:     m2mapi.PeriodInput{Start: period.Start, End: period.End},
		}
		transmit_data, err := json.Marshal(transmit_request)
		if err != nil {
			fmt.Println("Error marshaling data: ", err)
			return null_data
		}
		vsnode_url := "http://" + socket_address + "/primapi/data/past/node"
		response_data, err := http.Post(vsnode_url, "application/json", bytes.NewBuffer(transmit_data))
		if err != nil {
			fmt.Println("Error making request:", err)
			return null_data
		}
		defer response_data.Body.Close()

		byteArray, _ := io.ReadAll(response_data.Body)
		if err = json.Unmarshal(byteArray, &results); err != nil {
			fmt.Println("Error unmarshaling data: ", err)
			return null_data
		}
	}

	return results
}

func resolveCurrentNodeFunction(vnode_id, socket_address string, capability []string) m2mapi.ResolveDataByNode {
	null_data := m2mapi.ResolveDataByNode{VNodeID: "NULL"}
	var results m2mapi.ResolveDataByNode

	// 入力のVNodeIDが自身のVMNodeIDと一致するか比較する．
	// 一致すれば，自身のLocal GraphDBにVSNodeの検索をかける．一致しなければ，入力のSocketAddressに直接リクエスト転送する．

	if vnode_id == vmnode_id {
		var format_capability []string
		for _, cap := range capability {
			cap = "\\\"" + cap + "\\\""
			format_capability = append(format_capability, cap)
		}
		payload := `{"statements": [{"statement": "MATCH (vs:VSNode)-[:isPhysicalizedBy]->(ps:PSNode) WHERE ps.Capability IN [` + strings.Join(format_capability, ", ") + `] return vs.VNodeID, vs.SocketAddress;"}]}`
		graphdb_url := "http://" + os.Getenv("NEO4J_USERNAME") + ":" + os.Getenv("NEO4J_LOCAL_PASSWORD") + "@localhost:" + os.Getenv("NEO4J_LOCAL_PORT_GOLANG") + "/db/data/transaction/commit"
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
		var vsnode_set_own m2mapi.VNodeSet
		for _, v1 := range values {
			for k2, v2 := range v1.(map[string]interface{}) {
				if k2 == "data" {
					for _, v3 := range v2.([]interface{}) {
						for k4, v4 := range v3.(map[string]interface{}) {
							if k4 == "row" {
								row_data = v4
								dataArray := row_data.([]interface{})
								vsnode_set_own.VNodeID = dataArray[0].(string)
								vsnode_set_own.VNodeSocketAddress = dataArray[1].(string)
							}
						}
					}
				}
			}
		}
		// ここで，vsnode_set_own に格納されるVNodeの情報は1つだけであるという前提（ノード指定型データ取得だから）
		// マッチする情報が得られなかった場合，この時点でレスポンスする
		if vsnode_set_own.VNodeID == "" {
			return null_data
		}

		transmit_request := m2mapi.ResolveDataByNode{
			VNodeID:       vsnode_set_own.VNodeID,
			Capability:    capability,
			SocketAddress: vsnode_set_own.VNodeSocketAddress,
		}
		transmit_data, err := json.Marshal(transmit_request)
		if err != nil {
			fmt.Println("Error marshaling data: ", err)
			return null_data
		}
		vmnoder_url := "http://localhost:" + vmnoder_port + "/primapi/data/current/node"
		response_data, err := http.Post(vmnoder_url, "application/json", bytes.NewBuffer(transmit_data))
		if err != nil {
			fmt.Println("Error making request:", err)
			return null_data
		}
		defer response_data.Body.Close()

		byteArray, _ = io.ReadAll(response_data.Body)
		if err = json.Unmarshal(byteArray, &results); err != nil {
			fmt.Println("Error unmarshaling data: ", err)
			return null_data
		}
	} else {
		transmit_request := m2mapi.ResolveDataByNode{
			VNodeID:    vnode_id,
			Capability: capability,
		}
		transmit_data, err := json.Marshal(transmit_request)
		if err != nil {
			fmt.Println("Error marshaling data: ", err)
			return null_data
		}
		vsnode_url := "http://" + socket_address + "/primapi/data/current/node"
		response_data, err := http.Post(vsnode_url, "application/json", bytes.NewBuffer(transmit_data))
		if err != nil {
			fmt.Println("Error making request:", err)
			return null_data
		}
		defer response_data.Body.Close()

		byteArray, _ := io.ReadAll(response_data.Body)
		if err = json.Unmarshal(byteArray, &results); err != nil {
			fmt.Println("Error unmarshaling data: ", err)
			return null_data
		}
	}

	return results
}

func resolveConditionNodeFunction(vnode_id, socket_address string, capability []string, condition m2mapp.ConditionInput) m2mapi.ResolveDataByNode {
	null_data := m2mapi.ResolveDataByNode{VNodeID: "NULL"}
	var results m2mapi.ResolveDataByNode

	// 入力のVNodeIDが自身のVMNodeIDと一致するか比較する．
	// 一致すれば，自身のLocal GraphDBにVSNodeの検索をかける．一致しなければ，入力のSocketAddressに直接リクエスト転送する．

	if vnode_id == vmnode_id {
		var format_capability []string
		for _, cap := range capability {
			cap = "\\\"" + cap + "\\\""
			format_capability = append(format_capability, cap)
		}
		payload := `{"statements": [{"statement": "MATCH (vs:VSNode)-[:isPhysicalizedBy]->(ps:PSNode) WHERE ps.Capability IN [` + strings.Join(format_capability, ", ") + `] return vs.VNodeID, vs.SocketAddress;"}]}`
		graphdb_url := "http://" + os.Getenv("NEO4J_USERNAME") + ":" + os.Getenv("NEO4J_LOCAL_PASSWORD") + "@localhost:" + os.Getenv("NEO4J_LOCAL_PORT_GOLANG") + "/db/data/transaction/commit"
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
		var vsnode_set_own m2mapi.VNodeSet
		for _, v1 := range values {
			for k2, v2 := range v1.(map[string]interface{}) {
				if k2 == "data" {
					for _, v3 := range v2.([]interface{}) {
						for k4, v4 := range v3.(map[string]interface{}) {
							if k4 == "row" {
								row_data = v4
								dataArray := row_data.([]interface{})
								vsnode_set_own.VNodeID = dataArray[0].(string)
								vsnode_set_own.VNodeSocketAddress = dataArray[1].(string)
							}
						}
					}
				}
			}
		}
		// ここで，vsnode_set_own に格納されるVNodeの情報は1つだけであるという前提（ノード指定型データ取得だから）
		// マッチする情報が得られなかった場合，この時点でレスポンスする
		if vsnode_set_own.VNodeID == "" {
			return null_data
		}

		transmit_request := m2mapi.ResolveDataByNode{
			VNodeID:       vsnode_set_own.VNodeID,
			Capability:    capability,
			Condition:     m2mapi.ConditionInput{Limit: m2mapi.Range{LowerLimit: condition.Limit.LowerLimit, UpperLimit: condition.Limit.UpperLimit}, Timeout: condition.Timeout},
			SocketAddress: vsnode_set_own.VNodeSocketAddress,
		}
		transmit_data, err := json.Marshal(transmit_request)
		if err != nil {
			fmt.Println("Error marshaling data: ", err)
			return null_data
		}
		vmnoder_url := "http://localhost:" + vmnoder_port + "/primapi/data/condition/node"
		response_data, err := http.Post(vmnoder_url, "application/json", bytes.NewBuffer(transmit_data))
		if err != nil {
			fmt.Println("Error making request:", err)
			return null_data
		}
		defer response_data.Body.Close()

		byteArray, _ = io.ReadAll(response_data.Body)
		if err = json.Unmarshal(byteArray, &results); err != nil {
			fmt.Println("Error unmarshaling data: ", err)
			return null_data
		}
	} else {
		transmit_request := m2mapi.ResolveDataByNode{
			VNodeID:    vnode_id,
			Capability: capability,
			Condition:  m2mapi.ConditionInput{Limit: m2mapi.Range{LowerLimit: condition.Limit.LowerLimit, UpperLimit: condition.Limit.UpperLimit}, Timeout: condition.Timeout},
		}
		transmit_data, err := json.Marshal(transmit_request)
		if err != nil {
			fmt.Println("Error marshaling data: ", err)
			return null_data
		}
		vsnode_url := "http://" + socket_address + "/primapi/data/condition/node"
		response_data, err := http.Post(vsnode_url, "application/json", bytes.NewBuffer(transmit_data))
		if err != nil {
			fmt.Println("Error making request:", err)
			return null_data
		}
		defer response_data.Body.Close()

		byteArray, _ := io.ReadAll(response_data.Body)
		if err = json.Unmarshal(byteArray, &results); err != nil {
			fmt.Println("Error unmarshaling data: ", err)
			return null_data
		}
	}

	return results
}

func resolvePastAreaFunction(ad, capability, node_type string, period m2mapi.PeriodInput) m2mapi.ResolveDataByArea {
	null_data := m2mapi.ResolveDataByArea{AD: "NULL"}
	var results m2mapi.ResolveDataByArea

	// ADに含まれるすべてのVNodeIDに対して過去データ取得リクエストを転送したい．
	area_desc := ad_cache[ad]
	if node_type == "All" || node_type == "VSNode" {
		for _, vsnode := range area_desc.AreaDescriptorDetail[""].VNode {
			request_data := m2mapi.ResolveDataByNode{
				VNodeID: vsnode.VNodeID,
				//Capability: capability,
				Period: m2mapi.PeriodInput{Start: period.Start, End: period.End},
			}

			transmit_data, err := json.Marshal(request_data)
			if err != nil {
				fmt.Println("Error marshaling data: ", err)
				return null_data
			}
			transmit_url := "http://" + vsnode.VNodeSocketAddress + "/primapi/data/past/node"
			response_data, err := http.Post(transmit_url, "application/json", bytes.NewBuffer(transmit_data))
			if err != nil {
				fmt.Println("Error making request: ", err)
				return null_data
			}
			defer response_data.Body.Close()

			byteArray, _ := io.ReadAll(response_data.Body)
			var result m2mapi.ResolveDataByNode
			if err := json.Unmarshal(byteArray, &result); err != nil {
				fmt.Println("Error unmarshaling data: ", err)
				return null_data
			}

			data := m2mapi.SensorData{
				VNodeID: result.VNodeID,
				Values:  result.Values,
			}
			results.Datas = append(results.Datas, data)
		}
	}

	if node_type == "All" || node_type == "VMNode" {
		// はじめに，ADに登録されているPAreaIDに存在するPMNodeとそのソケットアドレスを検索する
		vmnode_results_by_resolve_node := resolveNodeFunction(ad, []string{capability}, node_type)
		for _, vmnode_result := range vmnode_results_by_resolve_node.VNode {
			request_data := m2mapi.ResolveDataByNode{
				VNodeID: vmnode_result.VNodeID,
				//Capability: capability,
				Period: m2mapi.PeriodInput{Start: period.Start, End: period.End},
			}

			transmit_data, err := json.Marshal(request_data)
			if err != nil {
				fmt.Println("Error marhsaling data: ", err)
				return null_data
			}
			transmit_url := "http://" + vmnode_result.VNodeSocketAddress + "/primpai/data/past/node"
			response_data, err := http.Post(transmit_url, "application/json", bytes.NewBuffer(transmit_data))
			if err != nil {
				fmt.Println("Error making request: ", err)
				return null_data
			}
			defer response_data.Body.Close()

			byteArray, _ := io.ReadAll(response_data.Body)
			var result m2mapi.ResolveDataByNode
			if err := json.Unmarshal(byteArray, &result); err != nil {
				fmt.Println("Error unmarshaling data: ", err)
				return null_data
			}

			data := m2mapi.SensorData{
				VNodeID: result.VNodeID,
				Values:  result.Values,
			}
			results.Datas = append(results.Datas, data)
		}
	}

	return results
}

func resolveCurrentAreaFunction(ad, capability, node_type string) m2mapi.ResolveDataByArea {
	null_data := m2mapi.ResolveDataByArea{AD: "NULL"}
	var results m2mapi.ResolveDataByArea

	// ADに含まれるすべてのVNodeIDに対して現在データ取得リクエストを転送したい．
	if node_type == "All" || node_type == "VSNode" {
		// はじめに，ADに登録されているVSNodeのうち，指定したCapabilityを持つものだけを抽出する
		vsnode_results_by_resolve_node := resolveNodeFunction(ad, []string{capability}, node_type)
		for _, vsnode_result := range vsnode_results_by_resolve_node.VNode {
			request_data := m2mapi.ResolveDataByNode{
				VNodeID: vsnode_result.VNodeID,
				//Capability: capability,
			}

			transmit_data, err := json.Marshal(request_data)
			if err != nil {
				fmt.Println("Error marshaling data: ", err)
				return null_data
			}
			// VSNodeへ転送
			transmit_url := "http://" + vsnode_result.VNodeSocketAddress + "/primapi/data/current/node"
			response_data, err := http.Post(transmit_url, "application/json", bytes.NewBuffer(transmit_data))
			if err != nil {
				fmt.Println("Error making request: ", err)
				return null_data
			}
			defer response_data.Body.Close()

			byteArray, _ := io.ReadAll(response_data.Body)
			var result m2mapi.ResolveDataByNode
			if err := json.Unmarshal(byteArray, &result); err != nil {
				fmt.Println("Error unmarshaling data: ", err)
				return null_data
			}

			data := m2mapi.SensorData{
				VNodeID: result.VNodeID,
				Values:  result.Values,
			}
			results.Datas = append(results.Datas, data)
		}
	}

	if node_type == "All" || node_type == "VMNode" {
		// はじめに，ADに登録されているPAreaIDに存在するPMNodeとそのソケットアドレスを検索する
		vmnode_results_by_resolve_node := resolveNodeFunction(ad, []string{capability}, node_type)
		for _, vmnode_result := range vmnode_results_by_resolve_node.VNode {
			request_data := m2mapi.ResolveDataByNode{
				VNodeID: vmnode_result.VNodeID,
				//Capability: capability,
			}

			transmit_data, err := json.Marshal(request_data)
			if err != nil {
				fmt.Println("Error marhsaling data: ", err)
				return null_data
			}
			// VSNodeへ転送
			transmit_url := "http://" + vmnode_result.VMNodeRSocketAddress + "/primpai/data/current/node"
			response_data, err := http.Post(transmit_url, "application/json", bytes.NewBuffer(transmit_data))
			if err != nil {
				fmt.Println("Error making request: ", err)
				return null_data
			}
			defer response_data.Body.Close()

			byteArray, _ := io.ReadAll(response_data.Body)
			var result m2mapi.ResolveDataByNode
			if err := json.Unmarshal(byteArray, &result); err != nil {
				fmt.Println("Error unmarshaling data: ", err)
				return null_data
			}

			data := m2mapi.SensorData{
				VNodeID: result.VNodeID,
				Values:  result.Values,
			}
			results.Datas = append(results.Datas, data)
		}
	}

	return results
}

func resolveConditionAreaFunction(ad, capability, node_type string, condition m2mapi.ConditionInput) m2mapi.ResolveDataByArea {
	null_data := m2mapi.ResolveDataByArea{AD: "NULL"}
	var results m2mapi.ResolveDataByArea

	// ADに含まれるすべてのVNodeIDに対して現在データ取得リクエストを転送したい．
	if node_type == "All" || node_type == "VSNode" {
		// はじめに，ADに登録されているVSNodeのうち，指定したCapabilityを持つものだけを抽出する
		vsnode_results_by_resolve_node := resolveNodeFunction(ad, []string{capability}, node_type)
		for _, vsnode_result := range vsnode_results_by_resolve_node.VNode {
			request_data := m2mapi.ResolveDataByNode{
				VNodeID: vsnode_result.VNodeID,
				//Capability: capability,
				Condition: condition,
			}

			transmit_data, err := json.Marshal(request_data)
			if err != nil {
				fmt.Println("Error marshaling data: ", err)
				return null_data
			}
			// VSNodeへ転送
			transmit_url := "http://" + vsnode_result.VNodeSocketAddress + "/primapi/data/condition/node"
			response_data, err := http.Post(transmit_url, "application/json", bytes.NewBuffer(transmit_data))
			if err != nil {
				fmt.Println("Error making request: ", err)
				return null_data
			}
			defer response_data.Body.Close()

			byteArray, _ := io.ReadAll(response_data.Body)
			var result m2mapi.ResolveDataByNode
			if err := json.Unmarshal(byteArray, &result); err != nil {
				fmt.Println("Error unmarshaling data: ", err)
				return null_data
			}

			data := m2mapi.SensorData{
				VNodeID: result.VNodeID,
				Values:  result.Values,
			}
			results.Datas = append(results.Datas, data)
		}
	}

	if node_type == "All" || node_type == "VMNode" {
		// はじめに，ADに登録されているPAreaIDに存在するPMNodeとそのソケットアドレスを検索する
		vmnode_results_by_resolve_node := resolveNodeFunction(ad, []string{capability}, node_type)
		for _, vmnode_result := range vmnode_results_by_resolve_node.VNode {
			request_data := m2mapi.ResolveDataByNode{
				VNodeID: vmnode_result.VNodeID,
				//Capability: capability,
				Condition: condition,
			}

			transmit_data, err := json.Marshal(request_data)
			if err != nil {
				fmt.Println("Error marhsaling data: ", err)
				return null_data
			}
			// VMNodeRへ転送
			transmit_url := "http://" + vmnode_result.VMNodeRSocketAddress + "/primpai/data/condition/node"
			response_data, err := http.Post(transmit_url, "application/json", bytes.NewBuffer(transmit_data))
			if err != nil {
				fmt.Println("Error making request: ", err)
				return null_data
			}
			defer response_data.Body.Close()

			byteArray, _ := io.ReadAll(response_data.Body)
			var result m2mapi.ResolveDataByNode
			if err := json.Unmarshal(byteArray, &result); err != nil {
				fmt.Println("Error unmarshaling data: ", err)
				return null_data
			}

			data := m2mapi.SensorData{
				VNodeID: result.VNodeID,
				Values:  result.Values,
			}
			results.Datas = append(results.Datas, data)
		}
	}

	return results
}

func actuateFunction(vnode_id, action, socket_address string, parameter float64) m2mapi.Actuate {
	null_data := m2mapi.Actuate{VNodeID: "NULL"}

	request_data := m2mapi.Actuate{
		VNodeID:   vnode_id,
		Action:    action,
		Parameter: parameter,
	}
	transmit_data, err := json.Marshal(request_data)
	if err != nil {
		fmt.Println("Error marshaling data: ", err)
		return null_data
	}
	transmit_url := "http://" + socket_address + "/primapi/actuate"
	response_data, err := http.Post(transmit_url, "application/json", bytes.NewBuffer(transmit_data))
	if err != nil {
		fmt.Println("Error making request: ", err)
		return null_data
	}
	defer response_data.Body.Close()

	byteArray, _ := io.ReadAll(response_data.Body)
	var results m2mapi.Actuate
	if err = json.Unmarshal(byteArray, &results); err != nil {
		fmt.Println("Error unmarshaling data: ", err)
		return null_data
	}

	return results
}

func bodyGraphDB(byteArray []byte) []interface{} {
	var jsonBody map[string]interface{}
	if err := json.Unmarshal(byteArray, &jsonBody); err != nil {
		message.MyError(err, "bodyGraphDB > json.Unmarshal")
		return nil
	}
	var values []interface{}
	for _, v1 := range jsonBody {
		switch t1 := v1.(type) {
		case []interface{}:
			for _, v2 := range v1.([]interface{}) {
				fmt.Println("v1([]interface{}): ", v2, "type: ", t1)
				values = v1.([]interface{})
			}
		case map[string]interface{}:
			for _, v2 := range v1.(map[string]interface{}) {
				switch t2 := v2.(type) {
				case []interface{}:
					values = v2.([]interface{})
				default:
					fmt.Println("type: ", t2)
				}
			}
		default:
			fmt.Println("Format Assertion False")
		}
	}
	return values
}

func addIfNotExists(slice []string, item string) []string {
	for _, v := range slice {
		if v == item {
			return slice
		}
	}
	return append(slice, item)
}
