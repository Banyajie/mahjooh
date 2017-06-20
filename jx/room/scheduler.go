package room

import (
	"encoding/json"
	"errors"
	"strconv"

	"chess_alg_jx/logger"
	"chess_alg_jx/utils"
)

//监控游戏需要的运行状态
const (
	GAME_STATUS_FREE     = iota //空转
	GAME_STATUS_MOPAI           //摸牌
	GAME_STATUS_CONTROL         //控牌
	GAME_STATUS_GRAB_BAR        //抢杠
	GAME_STATUS_CUT_HU          //截胡
)

func (room *Room) syncPlayerHander(s_id int) {
	//检测是否是当前控牌玩家信息
	logger.Notice("该玩家不是控牌玩家，同步该玩家信息")
	data := reconnection{
		SeatId:      s_id,
		UserId:      room.Player[s_id].UserId,
		CurrentId:   s_id,
		CurrentName: room.Player[s_id].UserName,
		ChuSeatId:   room.ChuPaiIndex,
		Head:        room.Player[s_id].Head,
		BankId:      room.BankerId,
		Score:       room.Player[s_id].Score,
		MScore:      room.Player[s_id].MScore,
		Hand:        room.Player[s_id].HandCard.ReverChange(),
		Eat:         room.Player[s_id].HandCard.Eat,
		Alt:         room.Player[s_id].HandCard.Alt,
		Ming:        room.Player[s_id].HandCard.Bright,
		Dark:        room.Player[s_id].HandCard.Dark,
		HavePlay:    room.Player[s_id].Played,
	}
	if err := room.GameNsqProducer(SEND_SYNC_HAND, s_id, data); err != nil {
		return
	}
	return
}

func (room *Room) ManagerGame(mess MessClient) {
	switch mess.MessType {
	case RCV_PLAYER_INTO_ROOM:
		//玩家进入房间
		room.RcvPlayerInRoom(mess.UserId, mess.UserName, mess.HeadImg)
	case RCV_PLAYER_OUT_ROOM:
		//玩家退出房间
		room.PlayerOutRoom(mess.SeatId)
	case RCV_APPLY_DISCARD_ROOM:
		//申请解散房间
		room.RcvApplyDiscardRoom(mess.SeatId)
	case RCV_DISCARD_CONFIRM:
		//确认解散房间
		room.RcvConfirmDiscardRoom(mess.SeatId)
	case RCV_DISCARD_CANCEL:
		//取消解散房间
		room.RcvConcelDiscardRoom(mess.SeatId)
	case RCV_RMREMATCH:
		//再来一局
		room.RcvRmRematch(mess.SeatId)
	case RCV_SHUFFLE:
		//洗牌完成
		room.RcvPlayerShuffled(mess.SeatId)
	case RCV_DECIDE_BANKER:
		//定庄色子完成
		room.RcvPlayerDiceBanker(mess.SeatId)
	case RCV_SHOW_UNIVERSE_END:
		//开局上下翻精牌展示完成
		room.RcvPlayerShowUniverse(mess.SeatId)
	case RCV_PLAY_A_HAND:
		//收到玩家出牌信息，检测其他三个玩家是否有可以碰/杠/胡的可能，如果有nsq推送消息，并开启定时器等待回应
		room.RcvPlayAHand(mess.SeatId, mess.MessData[0])
	case RCV_EAT:
		//收到玩家吃牌信息
		room.RcvEat(mess.SeatId, mess.MessData)
	case RCV_ALT:
		//收到玩家碰牌信息
		room.RcvAlt(mess.SeatId)
	case RCV_BRIGHT_BAR:
		//收到玩家明杠信息
		room.RcvBrightBar(mess.SeatId)
	case RCV_ADD_BAR:
		//收到玩家加杠信息
		room.RcvAddBar(mess.SeatId)
	case RCV_DARK_BAR:
		//收到玩家暗杠信息
		room.RcvDarkBar(mess.SeatId, mess.MessData[0])
	case RCV_PASS:
		//收到玩家过的信息
		room.RcvPass(mess.SeatId)
	case RCV_WIN:
		//收到玩家胡牌
		room.RcvHu(mess.SeatId)
	case RCV_PLAYER_DOWN:
		//收到玩家掉线
		room.RcvPlayerDown(mess.SeatId)
	case RCV_PLAYER_RECONNECTION:
		//收到玩家掉线重连
		room.RcvPlayerReconnnection(mess.SeatId)
	case RCV_REQ_BANKER:
		//请求当前房间的庄家ID
		room.RcvReqBanker(mess.SeatId)
	case RCV_BEAT:
		//心跳包
		room.RcvBeat(mess.SeatId)
	default:
		return
	}
	return
}

/*将当前room的数据同步到redis*/
func (room *Room) SyncRedis() error {
	var err error
	var data []byte

	data, err = json.Marshal(room)
	if err != nil {
		logger.Error("SyncRedis Marshal the data", err)
		return err
	}
	err = utils.RedisClient.Set("room_"+strconv.FormatInt(room.RoomId, 10), utils.ByteToString(data))
	if err != nil {
		logger.Error("SyncRedis set", err)
		return err
	}

	return nil
}

/*从redis中取出房间数据*/
func (room *Room) GetRoomData(r_id string) error {
	if r_id == "" {
		return errors.New("the r_id is nil")
	}
	if _, err := utils.RedisClient.IsKeyExit("room_" + r_id); err != nil {
		return errors.New("room is no exit")
	}
	var roomData string
	var err error
	if roomData, err = utils.RedisClient.Get("room_" + r_id); err != nil {
		return err
	}
	if err = json.Unmarshal([]byte(roomData), &room); err != nil {
		return err
	}
	return nil
}
