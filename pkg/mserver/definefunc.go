package mserver

type ConnectNew struct {
	//input
	VNodeID_n  string
	PN_Type    string
	Time       string
	Position   SquarePoint
	Capability string
	HomeMECID  string

	//output
	Status bool
	//VNodeID_n string (dup)
	SessionKey string
}

//新しい接続先のVPointへのメッセージフォーマットとしても活用
type ConnectForModule struct {
	//input
	VNodeID_n  string
	PN_Type    string
	Time       string
	Position   SquarePoint
	Capability string
	HomeMECID  string

	//output
	Status bool
	//VNodeID_n string (dup)
}

type Disconnect struct {
	//input
	VNodeID_n string

	//output
	Status bool
}

type SquarePoint struct {
	Lat float64
	Lon float64
}
