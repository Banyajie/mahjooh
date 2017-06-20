package jx_nsq

import (
	"chess_alg_jx/config"
	"encoding/json"
)

type Nsq_match_turn struct {
	McNo     int64  `json:"mc_no"`
	MId      int64  `json:"m_id"`
	TurnType int    `json:"turn_type"`
	CName    string `json:"c_name"`
}

/*生成玩家游戏的流水记录*/
func NsqMatchTurnProducer(data Nsq_match_turn) error {
	buff, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := NsqProducer.Publish(config.Config.NsqTopicMatchTurnToApi, buff); err != nil {
		return err
	}

	return nil
}
