package jx_nsq

import (
	"encoding/json"
	"chess_alg_jx/config"
)

type Nsq_match struct {
	RId      int64  `json:"r_id"`
	McNo     int64  `json:"mc_no"`
	EastMid  int64  `json:"east_mid"`
	SouthMid int64  `json:"south_mid"`
	WestMid  int64  `json:"west_mid"`
	NorthMid int64  `json:"north_mid"`
	CName    string `json:"c_name"`
}

func NsqMatchProducer(data Nsq_match) error {
	buff, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := NsqProducer.Publish(config.Config.NsqTopicMatchToApi, buff); err != nil {
		return err
	}

	return nil
}