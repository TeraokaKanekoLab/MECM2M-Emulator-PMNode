package vsnode

type ResolvePastDataByNode struct {
	// input
	// Local SensingDB へのクエリ

	// output
	PNodeID    string  `json:"pnode-id"`
	Capability string  `json:"capability"`
	Timestamp  string  `json:"timestamp"`
	Value      float64 `json:"value"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
}

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
