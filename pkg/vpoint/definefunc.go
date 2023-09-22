package vpoint

// VPointのソケットファイル群
type VPointSocketFiles struct {
	VPoints []string `json:"vpoints"`
}

// 地点指定型現在・条件充足データ取得時のVNode解決用フォーマット
type CurrentPointVNode struct {
	// input
	VPointID   string
	Capability string

	// output
	VNodeID       []string
	VNodeSockAddr []string
}
