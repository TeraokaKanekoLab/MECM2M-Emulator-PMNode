package m2mapi

import "time"

type Area struct {
	// data property
	AreaID      string
	Address     string
	NE          SquarePoint
	SW          SquarePoint
	Description string
	// Name string
	// IoTSPIDs []string

	// object property
	//contains PSink
}

type PSink struct {
	// data property
	VPointID_n  string
	Address     string
	Lat         float64
	Lon         float64
	Description string
	VPointID    string //追加
	ServerIP    string //追加
	//Policy         string
	//IoTSPID        string
}

type ResolveArea struct {
	// input
	NE SquarePoint
	SW SquarePoint

	// output
	AD  string
	TTL time.Time

	// etc.
	DestSocketAddr string // リンクプロセスが宛先のソケットアドレスを認識するために必要
}

// エリアマッピングを取得するためのフォーマット
type AreaMapping struct {
	// input
	NE SquarePoint `json:"ne"`
	SW SquarePoint `json:"sw"`

	// output
	ServerIPs []string `json:"serverips"`
}

type ExtendAD struct {
	// input
	AD string

	// output
	Flag bool
}

type ResolveNode struct {
	// input
	AD           string
	Capabilities []string
	NodeType     string

	// output
	VNodeID       string
	SocketAddress string

	// etc.
	DestSocketAddr string // リンクプロセスが宛先のソケットアドレスを認識するために必要
}

type ResolveDataByNode struct {
	// input
	VNodeID       string      //`json:"vnode_id"`
	Capability    string      //`json:"capability"`
	Period        PeriodInput //`json:"period"`
	Condition     ConditionInput
	SocketAddress string //`json:"socket_address"` // センサデータ取得対象となるVNodeのソケットアドレス

	// output
	//VNodeID 	string (dup)
	Values []Value

	// etc.
	DestSocketAddr string // リンクプロセスが宛先のソケットアドレスを認識するために必要
}

type ResolveDataByArea struct {
	// input
	AD         string
	Capability string
	Period     PeriodInput
	Condition  ConditionInput
	NodeType   string // VSNode or VMNode or Both
	//SocketAddress string

	// output
	Datas []SensorData
}

type Actuate struct {
	// input
	VNodeID       string
	Action        string
	Parameter     float64
	SocketAddress string // 動作指示対象となるVNodeのソケットアドレス

	// output
	Status bool

	// etc.
	DestSocketAddr string // リンクプロセスが宛先のソケットアドレスを認識するために必要
}

type ConditionInput struct {
	Limit   Range
	Timeout time.Duration
}

type AreaDescriptor struct {
	PAreaID  []string
	VSNode   map[string]string // VNodeID と SocketAddressのマッピング
	TTL      time.Time
	ServerIP []string // ノード解決時に使いたい，MECサーバのIPアドレス
}

type SensorData struct {
	VNodeID string
	Values  []Value
}

type SquarePoint struct {
	Lat float64
	Lon float64
}

type PeriodInput struct {
	Start string
	End   string
}

type Value struct {
	Capability string
	Time       string
	Value      float64
}

type Range struct {
	LowerLimit float64
	UpperLimit float64
}

type DataForRegist struct {
	PNodeID    string
	Capability string
	Timestamp  string
	Value      float64
	PSinkID    string
	Lat        float64
	Lon        float64
}
