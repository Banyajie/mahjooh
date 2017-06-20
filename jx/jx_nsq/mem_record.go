package jx_nsq

import (
	"chess_alg_jx/config"
	"encoding/json"
)

type Nsq_match_record struct {
	MId   int64  `json:"m_id"`
	Score int    `json:"score"`
	CName string `json:"c_name"`
}

/*生成玩家游戏的胜平负记录*/
func NsqMatchRecordProducer(data Nsq_match_record) error {
	buff, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := NsqProducer.Publish(config.Config.NsqTopicMemberRecordToApi, buff); err != nil {
		return err
	}

	return nil
}
