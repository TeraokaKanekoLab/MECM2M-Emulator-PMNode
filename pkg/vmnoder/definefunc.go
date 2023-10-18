package vmnoder

import "time"

type ResolveNode struct {
	// input
	// neo4j へのクエリ

	// output
	VNode []VNodeSet `json:"vnode"`
}

type ResolvePastDataByNode struct {
	// input
	VNodeID    string      `json:"vnode-id"`
	Capability []string    `json:"capability"`
	Period     PeriodInput `json:"period"`

	// output
	Values []Value `json:"values"`
}

type ResolveCurrentDataByNode struct {
	// input
	VNodeID    string   `json:"vnode-id"`
	Capability []string `json:"capability"`

	// output
	// VNodeID (dup)
	Values []Value `json:"values"`
}

type ResolveConditionDataByNode struct {
	// input
	VNodeID    string         `json:"vnode-id"`
	Capability []string       `json:"capability"`
	Condition  ConditionInput `json:"condition"`

	// output
	Values Value
}

type Actuate struct {
	// input
	VNodeID       string  `json:"vnode-id"`
	Capability    string  `json:"capability"`
	Action        string  `json:"action"`
	Parameter     float64 `json:"parameter"`
	SocketAddress string  `json:"socket-address"` // 動作指示対象となるVNodeのソケットアドレス

	// output
	Status bool `json:"status"`
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

type ConditionInput struct {
	Limit   Range         `json:"limit"`
	Timeout time.Duration `json:"timeout"`
}

type Range struct {
	LowerLimit float64 `json:"lower-limit"`
	UpperLimit float64 `json:"upper-limit"`
}
