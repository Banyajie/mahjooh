package room

import (
	"encoding/json"
	"github.com/nsqio/go-nsq"
	"strconv"
	"time"

	"chess_alg_jx/config"
	"chess_alg_jx/jx/jx_nsq"
	"chess_alg_jx/logger"
)

//接收的websocket消息类型
const (
	RCV_PLAYER_INTO_ROOM    = iota //进入房间
	RCV_PLAYER_OUT_ROOM            //退出房间
	RCV_REQ_BANKER                 //请求庄家ID
	RCV_PLAY_A_HAND                //出牌
	RCV_EAT                        //吃牌
	RCV_ALT                        //碰牌
	RCV_BRIGHT_BAR                 //明杠
	RCV_ADD_BAR                    //补杠
	RCV_DARK_BAR                   //暗杠
	RCV_PASS                       //pass
	RCV_WIN                        //胡牌
	RCV_PLAYER_DOWN                //掉线
	RCV_PLAYER_RECONNECTION        //掉线重连
	RCV_RMREMATCH                  //再来一局
	RCV_APPLY_DISCARD_ROOM         //解散房间请求
	RCV_SHUFFLE                    //洗牌完成
	RCV_DECIDE_BANKER              //定庄色子完成
	RCV_SHOW_UNIVERSE_END          //精牌展示完成
	RCV_DISCARD_CONFIRM            //确认解散房间
	RCV_DISCARD_CANCEL             //取消解散房间
	RCV_BEAT                       //心跳
)

//一个玩家可以对当前牌局有那些动作
const (
	ACTION_EAT        = 1  //吃
	ACTION_ALT        = 2  //碰
	ACTION_BRIGHT_BAR = 4  //明杠
	ACTION_ADD_BAR    = 8  //加杠
	ACTION_DACK_BAR   = 16 //暗杠
	ACTION_HU         = 32 //胡牌
	ACTION_PASS       = 64 //过
)

//nsq推送的websocket消息类型
const (
	SEND_PREROUND_DEAL       = iota //首轮发牌
	SEND_DEAL                       //发牌
	SEND_PLAY_A_HAND                //玩家出牌
	SEND_MAY_ACTION                 //玩家可以对其他玩家出的牌有所动作
	SEND_EAT                        //玩家吃牌
	SEND_ALT                        //玩家碰牌
	SEND_BRIGHT_BAR                 //玩家明杠
	SEND_ADD_BAR                    //玩家补杠
	SEND_DARK_BAR                   //玩家暗杠
	SEND_WIN                        //玩家胡牌
	SEND_GAME_OVER                  //游戏结束
	SEND_SYNC_HAND                  //同步玩家手牌
	SEND_SYNC_POINT                 //告诉玩家当前控牌ID
	SEND_PLAYER_IN_ROOM             //玩家进入房间
	SEND_BANKER_ID                  //告诉玩家庄家ID
	SEND_PLAYER_OUT_ROOM            //玩家退出房间
	SEND_PLAYER_HAND                //玩家的手牌
	SEND_PLAYER_DOWN                //玩家掉线
	SEND_PLAYER_RECONNECTION        //玩家掉线重连
	SEND_PLAYER_MONEY               //玩家的金额
	SEND_MATCH_OVER                 //游戏已结束
	SEND_DICE_BANKER                //定庄色子
	SEND_DICE_UNIVERSE              //定精色子
	SEND_UNIVERSE                   //精牌
	SEND_RMREMATCH                  //再来一局
	SEND_APPLY_DISCARD_ROOM         //申请解散房间
	SEND_MATCH_USEUP                //房卡局数已用完
	SEND_RESETTIMER                 //重置时间
	SEND_UNIVERSE_NUM               //精牌的数量
	SEND_START_SHUFFLE              //开始洗牌
	SEND_TOTAL_SCORE                //玩家总分
	SEND_PLAYER_UP                  //玩家上线
	SEND_DISCARD_CONFIRM            //确认解散房间
	SEND_DISCARD_CENCEL             //取消解散房间
	SEND_DISCARD_ROOM               //解散房间
	SEND_BEAT                       //心跳
)

//通信过程中服务器发送的消息格式
type Message struct {
	RoomId      int64       `json:"room_id"`             //房间ID
	UserId      int64       `json:"user_id"`             //玩家ID
	SeatId      int         `json:"seat_id"`             //玩家座位ID
	CurrentId   int         `json:"current_id"`          //当前控牌玩家ID
	CurrentName string      `json:"current_player_name"` //当前控牌玩家名字
	TimeStamp   int64       `json:"time_stamp"`          //发送消息时的时间戳
	MsgType     int         `json:"mess_type"`           //type
	MsgData     interface{} `json:"mess_data"`           //data
	MatchNum    int         `json:"match_num"`           //当前局数
	LeftNum     int         `json:"left_num"`            //当前牌局还剩多少牌
}

//通信过程中客户端发送的消息格式
type MessClient struct {
	RoomId   int64  `json:"room_id"`   //房间ID
	UserId   int64  `json:"user_id"`   //玩家ID
	UserName string `json:"user_name"` //玩家姓名
	HeadImg  string `json:"head_img"`  //玩家头像
	SeatId   int    `json:"seat_id"`   //玩家座位ID
	MessType int    `json:"mess_type"` //type
	MessData []int  `json:"mess_data"` //data
}

func GameNsqHandle(msg *nsq.Message) error {
	var mess MessClient
	if err := json.Unmarshal(msg.Body, &mess); err != nil {
		logger.Error("GameNsqHandle json.Unmarshal err: ", err)
		return err
	}

	rm := Room{}
	if err := rm.GetRoomData(strconv.FormatInt(mess.RoomId, 10)); err != nil {
		return nil
	}
	if mess.MessType != RCV_BEAT {
		logger.Noticef("Rcv %d 玩家消息 MessType: %d-%s, Msg: %+v", mess.SeatId, mess.MessType, rcvMsgHint(mess.MessType), mess)
	}
	rm.ManagerGame(mess)

	return nil
}

func GameNsqConsumer() {
	nsqConfig := nsq.NewConfig()

	wsConsumer, err := nsq.NewConsumer(config.Config.NsqTopicWsToGame, config.Config.NsqChannelWsToGame, nsqConfig)
	if err != nil {
		logger.Error("game NewConsumer", err)
		return
	}

	wsConsumer.AddConcurrentHandlers(nsq.HandlerFunc(GameNsqHandle), 5)
	if err := wsConsumer.ConnectToNSQD(config.Config.NsqAddr); err != nil {
		logger.Error("ConnectToNSQD", err)
		return
	}
}

func (room *Room) GameNsqProducer(mt int, s_id int, data interface{}) error {
	//游戏回放
	data_str, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if mt == SEND_PREROUND_DEAL || mt == SEND_DEAL || mt == SEND_PLAY_A_HAND ||
		mt == SEND_MAY_ACTION || mt == SEND_EAT || mt == SEND_ALT || mt == SEND_BRIGHT_BAR ||
		mt == SEND_ADD_BAR || mt == SEND_DARK_BAR || mt == SEND_WIN || mt == SEND_GAME_OVER ||
		mt == SEND_BANKER_ID || mt == SEND_PLAYER_HAND || mt == SEND_PLAYER_MONEY || mt == SEND_DICE_BANKER ||
		mt == SEND_DICE_UNIVERSE || mt == SEND_UNIVERSE || mt == SEND_UNIVERSE_NUM || mt == SEND_START_SHUFFLE || mt == SEND_TOTAL_SCORE {
		playback := jx_nsq.NsqMatchPlayback{
			MId:       room.Player[s_id].UserId,
			McNo:      room.McNo,
			SId:       s_id,
			CurId:     room.CurrentIndex,
			CurName:   room.Player[room.CurrentIndex].UserName,
			MsgType:   mt,
			MsgData:   string(data_str),
			TimeStamp: time.Now().Unix(),
			LeftNum:   room.LeftNum,
			CName:     "chess_alg_jx",
		}
		if err := jx_nsq.NsqMatchPlaybackProducer(playback); err != nil {
			return err
		}
	}
	message := &Message{
		RoomId:      room.RoomId,
		UserId:      room.Player[s_id].UserId,
		SeatId:      s_id,
		CurrentId:   room.CurrentIndex,
		CurrentName: room.Player[room.CurrentIndex].UserName,
		TimeStamp:   time.Now().Unix(),
		MsgType:     mt,
		MsgData:     data,
		MatchNum:    room.GameNum,
		LeftNum:     room.LeftNum,
	}
	if message.MsgType != SEND_SYNC_POINT && message.MsgType != SEND_RESETTIMER && message.MsgType != SEND_BEAT {
		logger.Debugf("Send MessType: %d-%s, Msg: %+v", message.MsgType, sendMsgHint(message.MsgType), message)
	}
	buff, err := json.Marshal(message)
	if err != nil {
		return err
	}
	if err := jx_nsq.NsqProducer.Publish(config.Config.NsqTopicGameToWs, buff); err != nil {
		return err
	}

	return nil
}

func rcvMsgHint(id int) string {
	switch id {
	case RCV_PLAYER_INTO_ROOM:
		return "进入房间"
	case RCV_PLAYER_OUT_ROOM:
		return "退出房间"
	case RCV_REQ_BANKER:
		return "请求庄家id"
	case RCV_PLAY_A_HAND:
		return "出牌"
	case RCV_EAT:
		return "吃牌"
	case RCV_ALT:
		return "碰牌"
	case RCV_BRIGHT_BAR:
		return "明杠"
	case RCV_ADD_BAR:
		return "补杠"
	case RCV_DARK_BAR:
		return "暗杠"
	case RCV_PASS:
		return "pass"
	case RCV_WIN:
		return "胡牌"
	case RCV_PLAYER_DOWN:
		return "掉线"
	case RCV_PLAYER_RECONNECTION:
		return "掉线重连"
	case RCV_RMREMATCH:
		return "再来一局"
	case RCV_APPLY_DISCARD_ROOM:
		return "申请解散房间"
	case RCV_SHUFFLE:
		return "洗牌完成"
	case RCV_DECIDE_BANKER:
		return "定庄色子效果完成"
	case RCV_SHOW_UNIVERSE_END:
		return "精牌展示完成"
	case RCV_DISCARD_CONFIRM:
		return "确认解散房间"
	case RCV_DISCARD_CANCEL:
		return "取消解散房间"
	case RCV_BEAT:
		return "心跳"
	default:
		return ""
	}
	return ""
}

func sendMsgHint(messId int) string {
	switch messId {
	case SEND_PREROUND_DEAL:
		return "首轮发牌"
	case SEND_DEAL:
		return "发牌"
	case SEND_PLAY_A_HAND:
		return "玩家出牌"
	case SEND_MAY_ACTION:
		return "玩家对当前出牌可以吃-碰-杠-胡"
	case SEND_EAT:
		return "玩家吃牌"
	case SEND_ALT:
		return "玩家碰牌"
	case SEND_BRIGHT_BAR:
		return "玩家明杠"
	case SEND_ADD_BAR:
		return "玩家加杠"
	case SEND_DARK_BAR:
		return "玩家暗杠"
	case SEND_WIN:
		return "玩家胡牌"
	case SEND_GAME_OVER:
		return "游戏结束"
	case SEND_SYNC_HAND:
		return "同步玩家手牌"
	case SEND_SYNC_POINT:
		return "同步控牌id"
	case SEND_PLAYER_IN_ROOM:
		return "玩家进入房间"
	case SEND_BANKER_ID:
		return "庄家id"
	case SEND_PLAYER_OUT_ROOM:
		return "退出房间"
	case SEND_PLAYER_HAND:
		return "玩家手牌"
	case SEND_PLAYER_DOWN:
		return "玩家掉线"
	case SEND_PLAYER_RECONNECTION:
		return "玩家掉线重联"
	case SEND_PLAYER_MONEY:
		return "玩家积分"
	case SEND_MATCH_OVER:
		return "游戏已结束"
	case SEND_DICE_BANKER:
		return "定庄时色子"
	case SEND_DICE_UNIVERSE:
		return "定精时色子"
	case SEND_UNIVERSE:
		return "精牌"
	case SEND_RMREMATCH:
		return "再来一局"
	case SEND_APPLY_DISCARD_ROOM:
		return "申请解散房间"
	case SEND_MATCH_USEUP:
		return "房卡局数用完"
	case SEND_RESETTIMER:
		return "重置时间"
	case SEND_UNIVERSE_NUM:
		return "精牌数量"
	case SEND_START_SHUFFLE:
		return "开始洗牌"
	case SEND_TOTAL_SCORE:
		return "玩家总分"
	case SEND_PLAYER_UP:
		return "玩家上线"
	case SEND_DISCARD_CONFIRM:
		return "确认解散房间"
	case SEND_DISCARD_CENCEL:
		return "取消解散房间"
	case SEND_DISCARD_ROOM:
		return "解散房间"
	case SEND_BEAT:
		return "心跳"
	default:
		return ""
	}
	return ""
}
