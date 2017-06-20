package jx_nsq

import (
	"chess_alg_jx/config"
	"encoding/json"
)

type Nsq_match_bill struct {
	McNo      int64  `json:"mc_no"`
	FMid      int64  `json:"f_mid"`
	TMid      int64  `json:"t_mid"`
	FNickname string `json:"f_nickname"`
	TNickname string `json:"t_nickname"`
	WinType   int    `json:"win_type"`
	SType     int    `json:"s_type"`
	Amount    int    `json:"amount"`
	CName     string `json:"c_name"`
}

/*生成玩家游戏的流水记录*/
func NsqMatchBillProducer(data Nsq_match_bill) error {
	buff, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := NsqProducer.Publish(config.Config.NsqTopicMatchBillToApi, buff); err != nil {
		return err
	}

	return nil
}
