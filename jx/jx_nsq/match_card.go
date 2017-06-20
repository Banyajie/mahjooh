package jx_nsq

import (
	"chess_alg_jx/config"
	"encoding/json"
)

type Nsq_match_card struct {
	McNo      int64  `json:"mc_no"`
	MId       int64  `json:"m_id"`
	HandCards string `json:"hand_cards"`
	CardType  int    `json:"card_type"`
	CName     string `json:"c_name"`
}

/*生成玩家游戏的流水记录*/
func NsqMatchCardProducer(data Nsq_match_card) error {
	buff, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := NsqProducer.Publish(config.Config.NsqTopicMatchCardToApi, buff); err != nil {
		return err
	}

	return nil
}
