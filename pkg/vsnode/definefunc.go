package vsnode

type ResolveCurrentDataByNode struct {
	// input
	// PSNode へのリクエスト転送
	PNodeID    string `json:"pnode-id"`
	Capability string `json:"capability"`

	// output
	//Capability (dup)
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}
