package m2mapp

import (
	"fmt"
	"mecm2m-Emulator-PMNode/pkg/m2mapi"
	"time"
)

type App struct {
	AppID       string
	Address     string
	Description string
	GID         uint64 // goroutine ID
}

type AppInputData struct {
	AppName    string
	NE         SquarePoint
	SW         SquarePoint
	Capability []string
	Period     PeriodInput
	Limit      Range
	Timeout    int
	Action     string
	Parameter  float64
}

type ResolveAreaInput struct {
	// input
	NE SquarePoint `json:"ne"`
	SW SquarePoint `json:"sw"`
}

type ResolveAreaOutput struct {
	// output
	AD  string    `json:"ad"`
	TTL time.Time `json:"ttl"`

	// etc.
	Descriptor m2mapi.AreaDescriptor `json:"area-descriptor"` // PMNodeにADに紐づく情報を与えるため
}

type ResolveNodeInput struct {
	// input
	AD         string   `json:"ad"`
	Capability []string `json:"capability"`
	NodeType   string   `json:"node-type"`
}

type ResolveNodeOutput struct {
	// output
	VNode []m2mapi.VNodeSet `json:"vnode"`
}

type ResolveDataByNodeInput struct {
	// input
	VNodeID       string         `json:"vnode-id"`
	Capability    []string       `json:"capability"`
	Period        PeriodInput    `json:"period"`
	Condition     ConditionInput `json:"condition"`
	SocketAddress string         `json:"socket-address"`
}

type ResolveDataByNodeOutput struct {
	// output
	VNodeID string  `json:"vnode-id"`
	Values  []Value `json:"values"`
}

type ResolveDataByAreaInput struct {
	// input
	AD         string         `json:"ad"`
	Capability []string       `json:"capability"`
	Period     PeriodInput    `json:"period"`
	Condition  ConditionInput `json:"condition"`
	NodeType   string         `json:"node-type"`
}

type ResolveDataByAreaOutput struct {
	// output
	Values map[string][]Value `json:"values"`
}

type ActuateInput struct {
	// input
	VNodeID       string  `json:"vnode-id"`
	Capability    string  `json:"capability"`
	Action        string  `json:"action"`
	Parameter     float64 `json:"parameter"`
	SocketAddress string  `json:"socket-address"`
}

type ActuateOutput struct {
	// output
	Status bool `json:"status"`
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
	Capability string  `json:"capability"`
	Time       string  `json:"time"`
	Value      float64 `json:"value"`
}

type ConditionInput struct {
	Limit   Range         `json:"limit"`
	Timeout time.Duration `json:"timeout"`
}

type Range struct {
	LowerLimit float64 `json:"lower-limit"`
	UpperLimit float64 `json:"upper-limit"`
}

func (a *App) String() string {
	return fmt.Sprintf("AppID: %s, Address: %s, Description: %s, GID: %d", a.AppID, a.Address, a.Description, a.GID)
}
