package server

type AAA struct {
	// input
	VNodeID_n  string
	HomeMECID  string
	Credential string

	// output
	Status bool
	// VNodeID_n string (dup)
}

// Serverのソケットファイル群
type ServerSocketFiles struct {
	M2MApi    string `json:"m2mApi"`
	LocalMgr  string `json:"localMgr"`
	PNodeMgr  string `json:"pnodeMgr"`
	AAA       string `json:"aaa"`
	LocalRepo string `json:"localRepo"`
	GraphDB   string `json:"graphDB"`
	SensingDB string `json:"sensingDB"`
}

// SessionKey 用のパケットフォーマット
type RequestSessionKey struct {
	// input
	PNodeID string

	// output
	SessionKey string
}
