package psnode

import "time"

// 時刻受信用の型
type TimeSync struct {
	PNodeID     string
	CurrentTime time.Time
}

// センサデータ型
type DataForRegist struct {
	PNodeID    string  `json:"pnode-id"`
	Capability string  `json:"capability"`
	Timestamp  string  `json:"timestamp"`
	Value      float64 `json:"value"`
	PSinkID    string  `json:"psink-id"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
}

// PMNodeの移動を知らせるパッケージ
type Mobility struct {
	PNodeID string `json:"pnode-id"`
}
