package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

var (
	confFile string
	Config   = &config{}
)

type config struct {
	ApiAddr        string `json:"api_addr"`
	NsqAddr        string `json:"nsq_addr"`
	NsqLookupdAddr string `json:"nsq_lookupd_addr"`

	NsqTopicWsToGame   string `json:"nsq_topic_ws_to_game"`
	NsqTopicGameToWs   string `json:"nsq_topic_game_to_ws"`
	NsqChannelWsToGame string `json:"nsq_channel_ws_to_game"`
	NsqChannelGameToWs string `json:"nsq_channel_game_to_ws"`

	NsqTopicMatchToApi         string `json:"nsq_topic_match_to_api"`
	NsqTopicMatchPlaybackToApi string `json:"nsq_topic_match_playback_to_api"`
	NsqTopicMatchBillToApi     string `json:"nsq_topic_match_bill_to_api"`
	NsqTopicMatchTurnToApi     string `json:"nsq_topic_match_turn_to_api"`
	NsqTopicMatchCardToApi     string `json:"nsq_topic_match_card_to_api"`
	NsqTopicMatchSumToApi      string `json:"nsq_topic_match_score_to_api"`
	NsqTopicUniverseToApi      string `json:"nsq_topic_match_universe_to_api"`
	NsqTopicMemberRecordToApi  string `json:"nsq_topic_member_record_to_api"`
	NsqTopicRoomCardToApi      string `json:"nsq_topic_room_card_to_api"`
	NsqTopicRoomLogToApi       string `json:"nsq_topic_room_log_to_api"`

	NsqChannelMatchToApi         string `json:"nsq_channel_match_to_api"`
	NsqChannelMatchPlaybackToApi string `json:"nsq_channel_match_playback_to_api"`
	NsqChannelMatchBillToApi     string `json:"nsq_channel_match_bill_to_api"`
	NsqChannelMatchTurnToApi     string `json:"nsq_channel_match_turn_to_api"`
	NsqChannelMatchCardToApi     string `json:"nsq_channel_match_card_to_api"`
	NsqChannelMatchSumToApi      string `json:"nsq_channel_match_score_to_api"`
	NsqChannelUniverseToApi      string `json:"nsq_channel_match_universe_to_api"`
	NsqChannelMemberRecordToApi  string `json:"nsq_channel_member_record_to_api"`
	NsqChannelRoomCardToApi      string `json:"nsq_channel_room_card_to_api"`
	NsqChannelRoomLogToApi       string `json:"nsq_channel_room_log_to_api"`

	RedisAddr   string `json:"redis_addr"`
	RedisPrefix string `json:"redis_prefix"`
	MysqlDb     string `json:"mysql_db"`
	LogLevel    string `json:"log_level"`
	DealSpead   int    `json:"deal_spead"`
}

func InitConfig() error {
	flag.StringVar(&confFile, "c", "", "conf filename")
	flag.Parse()
	var err error
	var jsonBlob []byte
	if jsonBlob, err = ioutil.ReadFile(confFile); err != nil {
		log.Fatal("read conffile err", err)
		return err
	}
	if err = json.Unmarshal(jsonBlob, Config); err != nil {
		log.Fatal("unmarshal conffile err", err)
		return err
	}
	return nil
}
