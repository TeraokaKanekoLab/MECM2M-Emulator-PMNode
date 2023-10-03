package vmnoder

type ResolveNode struct {
	// input
	// neo4j へのクエリ

	// output
	VNode []VNodeSet `json:"vnode"`
}

type ResolvePastDataByNode struct {
	// input
	VNodeID    string      `json:"vnode-id"`
	Capability string      `json:"capability"`
	Period     PeriodInput `json:"period"`
}

type ResolveCurrentDataByNode struct {
	// input
	VNodeID    string `json:"vnode-id"`
	Capability string `json:"capability"`

	// output
	// VNodeID (dup)
	Values []Value `json:"values"`
}

type VNodeSet struct {
	VNodeID       string `json:"vnode-id"`
	SocketAddress string `json:"socket-address"`
}
type Value struct {
	Capability string  `json:"capability"`
	Time       string  `json:"time"`
	Value      float64 `json:"value"`
}

type PeriodInput struct {
	Start string `json:"start"`
	End   string `json:"end"`
}
