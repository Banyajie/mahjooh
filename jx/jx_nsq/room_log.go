package jx_nsq

import (
	"chess_alg_jx/config"
	"encoding/json"
)

type Nsq_room_log struct {
	MId    int64  `json:"m_id"`
	RlType int    `json:"rl_type"`
	CName  string `json:"c_name"`
}

/*生成玩家游戏的流水记录*/
func NsqRoomLogProducer(data Nsq_room_log) error {
	buff, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err := NsqProducer.Publish(config.Config.NsqTopicRoomLogToApi, buff); err != nil {
		return err
	}

	return nil
}
