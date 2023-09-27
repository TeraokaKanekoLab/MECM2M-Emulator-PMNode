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
	NE SquarePoint `json:"ne"`
	SW SquarePoint `json:"sw"`

	// output
	AD  string    `json:"ad"`
	TTL time.Time `json:"ttl"`

	// etc.
	Descriptor   AreaDescriptor `json:"area-descriptor"` // 転送相手からADの内容を得るため
	TransmitFlag bool           `json:"transmit-flag"`   // M2M API -> M2M API のリクエスト転送か否かのフラグ
}

// エリアマッピングを取得するためのフォーマット
type AreaMapping struct {
	// input
	NE SquarePoint `json:"ne"`
	SW SquarePoint `json:"sw"`

	// output
	MECCoverArea MECCoverArea `json:"mec-cover-area"`
}

type ExtendAD struct {
	// input
	AD string `json:"ad"`

	// output
	Flag bool `json:"flag"`
}

type ResolveNode struct {
	// input
	AD           string   `json:"ad"`
	Capabilities []string `json:"capabilities"`
	NodeType     string   `json:"node-type"`

	// output
	VNode []VNodeSet `json:"vnode"`
}

type ResolveDataByNode struct {
	// input
	VNodeID       string         `json:"vnode-id"`
	Capability    string         `json:"capability"`
	Period        PeriodInput    `json:"period"`
	Condition     ConditionInput `json:"condition"`
	SocketAddress string         `json:"socket-address"` // センサデータ取得対象となるVNodeのソケットアドレス

	// output
	//VNodeID 	string (dup)
	Values []Value `json:"values"`
}

type ResolveDataByArea struct {
	// input
	AD         string         `json:"ad"`
	Capability string         `json:"capability"`
	Period     PeriodInput    `json:"period"`
	Condition  ConditionInput `json:"condition"`
	NodeType   string         `json:"node-type"` // VSNode or VMNode or Both

	// output
	Datas []SensorData `json:"datas"`
}

type Actuate struct {
	// input
	VNodeID       string  `json:"vnode-id"`
	Action        string  `json:"action"`
	Parameter     float64 `json:"parameter"`
	SocketAddress string  `json:"socket-address"` // 動作指示対象となるVNodeのソケットアドレス

	// output
	Status bool `json:"status"`
}

type ConditionInput struct {
	Limit   Range         `json:"limit"`
	Timeout time.Duration `json:"timeout"`
}

type AreaDescriptor struct {
	// ServerIPをkey, AreaDescriptorDetailをvalue
	AreaDescriptorDetail map[string]AreaDescriptorDetail `json:"area-descriptor-detail"`
}

type AreaDescriptorDetail struct {
	PAreaID []string   `json:"parea-id"`
	VNode   []VNodeSet `json:"vnode"`
	TTL     time.Time  `json:"ttl"`
}

type SensorData struct {
	VNodeID string  `json:"vnode-id"`
	Values  []Value `json:"values"`
}

type SquarePoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type PeriodInput struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Value struct {
	Capability string  `json:"capability"`
	Time       string  `json:"time"`
	Value      float64 `json:"value"`
}

type Range struct {
	LowerLimit float64 `json:"lower-limit"`
	UpperLimit float64 `json:"upper-limit"`
}

type DataForRegist struct {
	PNodeID    string  `json:"pnode-id"`
	Capability string  `json:"capability"`
	Timestamp  string  `json:"timestamp"`
	Value      float64 `json:"value"`
	PSinkID    string  `json:"psink-id"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
}

type VNodeSet struct {
	VNodeID              string `json:"vnode-id"`
	VNodeSocketAddress   string `json:"vnode-socket-address"`
	VMNodeRSocketAddress string `json:"vmnoder-socket-address"`
}

type MECCoverArea struct {
	ServerIP string  `json:"server-ip"`
	MinLat   float64 `json:"min-lat"`
	MaxLat   float64 `json:"max-lat"`
	MinLon   float64 `json:"min-lon"`
	MaxLon   float64 `json:"max-lon"`
}
