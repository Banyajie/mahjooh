package jx_nsq

import (
	"github.com/nsqio/go-nsq"
	"chess_alg_jx/config"
)

var (
	NsqProducer   *nsq.Producer
)

func InitNsq() error {
	var err error
	if NsqProducer, err = nsq.NewProducer(config.Config.NsqAddr, nsq.NewConfig()); err != nil {
		return err
	}

	return nil
}
