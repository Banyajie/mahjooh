package jx_nsq

import (
	"chess_alg_jx/config"
	"encoding/json"
)

type Nsq_match_sum struct {
	McNo     int64  `json:"mc_no"`
	EastMid  int64  `json:"east_mid"`
	EastSum  int    `json:"east_sum"`
	SouthMid int64  `json:"south_mid"`
	SouthSum int    `json:"south_sum"`
	WestMid  int64  `json:"west_mid"`
	WestSum  int    `json:"west_sum"`
	NorthMid int64  `json:"north_mid"`
	NorthSum int    `json:"north_sum"`
	CName    string `json:"c_name"`
}

func NsqMatchSumProducer(data Nsq_match_sum) error {
	buff, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := NsqProducer.Publish(config.Config.NsqTopicMatchSumToApi, buff); err != nil {
		return err
	}

	return nil
}
