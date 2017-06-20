package jx_nsq

import (
	"chess_alg_jx/config"
	"encoding/json"
)

const (
	UNIVERSE_TYPE_UP    = 1
	UNIVERSE_TYPE_DOWN  = 2
	UNIVERSE_TYPE_SMILE = 3
)

type Nsq_Universe struct {
	McNo    int64  `json:"mc_no"`
	Type    int    `json:"type"`
	Manager int    `json:"manager"`
	Deputy  int    `json:"deputy"`
	CName   string `json:"c_name"`
}

/*生成玩家游戏的流水记录*/
func NsqUniverseProducer(data Nsq_Universe) error {
	buff, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := NsqProducer.Publish(config.Config.NsqTopicUniverseToApi, buff); err != nil {
		return err
	}

	return nil
}
