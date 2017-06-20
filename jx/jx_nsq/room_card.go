package jx_nsq

import (
	"chess_alg_jx/config"
	"encoding/json"
)

//房卡流水类型
const (
	RCARD_TYPE_FIRSTBUY = 1 //第一次购买
	RCARD_TYPE_DEDUCT   = 2 //消耗
	RCARD_TYPE_RETURN   = 3 //退回
	RCARD_TYPE_PRESENT  = 4 //首次赠送
)

type Nsq_Rcard struct {
	MId   int64  `json:"m_id"`
	Rid   int64  `json:"rid"`
	Type  int    `json:"type"`
	Value int    `json:"value"`
	CName string `json:"c_name"`
}

/*生成玩家游戏的流水记录*/
func NsqRcardProducer(data Nsq_Rcard) error {
	buff, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := NsqProducer.Publish(config.Config.NsqTopicRoomCardToApi, buff); err != nil {
		return err
	}

	return nil
}
