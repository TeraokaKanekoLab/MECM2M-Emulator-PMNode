package psnode

import "time"

// PSNodeのソケットファイル群
type PSNodeSocketFiles struct {
	PSNodes []string `json:"psnodes"`
}

// 時刻受信用の型
type TimeSync struct {
	PNodeID     string
	CurrentTime time.Time
}
