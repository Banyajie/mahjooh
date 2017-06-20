package jx_nsq

import (
	"chess_alg_jx/config"
	"encoding/json"
)

type NsqMatchPlayback struct {
	MId       int64  `json:"m_id"`
	McNo      int64  `json:"mc_no"`
	SId       int    `json:"s_id"`
	CurId     int    `json:"cur_id"`
	CurName   string `json:"cur_name"`
	MsgType   int    `json:"msg_type"`
	MsgData   string `json:"msg_data"`
	TimeStamp int64  `json:"time_stamp"`
	LeftNum   int    `json:"left_num"`
	CName     string `json:"c_name"`
}

/*生成玩家游戏的流水记录*/
func NsqMatchPlaybackProducer(data NsqMatchPlayback) error {
	buff, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err := NsqProducer.Publish(config.Config.NsqTopicMatchPlaybackToApi, buff); err != nil {
		return err
	}

	return nil
}
