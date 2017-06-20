package room

import (
	"strconv"
	"time"

	"chess_alg_jx/jx/jx_nsq"
	"chess_alg_jx/logger"
	"chess_alg_jx/timer"
	"chess_alg_jx/utils"
)

//玩家在房间中的基本信息
type Player_data struct {
	UserId     int64  `json:"user_id"`
	UserName   string `json:"user_name"`
	UserHead   string `json:"user_head"`
	UserSeat   int    `json:"user_seat"`
	UserStatus bool   `json:"user_status"`
}

/*请求当前房间中的庄家ID*/
func (room *Room) RcvReqBanker(s_id int) {
	if err := room.GameNsqProducer(SEND_BANKER_ID, s_id, room.BankerId); err != nil {
		return
	}
}

/*再来一局*/
func (room *Room) RcvRmRematch(seatId int) {
	if room.GameNum == room.RoomSet.McCnt*8 {
		for i := 0; i < 4; i++ {
			if err := room.GameNsqProducer(SEND_MATCH_USEUP, i, nil); err != nil {
				return
			}
		}
		return
	}
	room.Player[seatId].State = PLAYER_STATE_READY
	room.CurrentIndex = seatId
	for i := 0; i < 4; i++ {
		if err := room.GameNsqProducer(SEND_RMREMATCH, i, seatId); err != nil {
			return
		}
	}
	for i := 0; i < 4; i++ {
		if room.Player[i].UserId != 0 {
			if room.Player[i].State != PLAYER_STATE_READY {
				room.SyncRedis()
				return
			}
		}
	}
	room.StartGame()
}

/*申请解散房间*/
func (room *Room) RcvApplyDiscardRoom(s_id int) {
	room.Player[s_id].Discard = true
	room.SyncRedis()
	if room.GameStatus != GAME_STATUS_FREE {
		for i := 1; i < 4; i++ {
			if room.Player[i].UserId != 0 {
				if room.Player[i].State == PLAYER_STATE_LEAF || room.Player[i].State == PLAYER_STATE_DOWN {
					continue
				}
				if err := room.GameNsqProducer(SEND_APPLY_DISCARD_ROOM, (i+s_id)%4, s_id); err != nil {
					return
				}
			}
		}
	}
	return
}

/*确认解散房间*/
func (room *Room) RcvConfirmDiscardRoom(s_id int) {
	room.Player[s_id].Discard = true
	room.SyncRedis()
	if room.GameStatus != GAME_STATUS_FREE {
		for i := 1; i < 4; i++ {
			if room.Player[i].UserId != 0 {
				if room.Player[i].State != PLAYER_STATE_DOWN && room.Player[i].State != PLAYER_STATE_LEAF {
					if err := room.GameNsqProducer(SEND_DISCARD_CONFIRM, (i+s_id)%4, s_id); err != nil {
						return
					}
				}
			}
		}
		//检测是否所有玩家确认取消，如果确认则解散房间
		for i := 0; i < 4; i++ {
			if room.Player[i].UserId != 0 {
				if room.Player[i].State != PLAYER_STATE_DOWN && room.Player[i].State != PLAYER_STATE_LEAF {
					if room.Player[i].Discard == false {
						return
					}
				}
			}
		}
	}
	//解散房间
	room_log := jx_nsq.Nsq_room_log{
		MId:    room.Player[s_id].UserId,
		RlType: ROOM_ACTION_TYPE_DISCARD,
		CName:  "chess_alg_jx",
	}
	if err := jx_nsq.NsqRoomLogProducer(room_log); err != nil {
		logger.Error("NsqMatchCardProducer: ", err)
		return
	}
	if room.GameNum == 0 {
		rcard := jx_nsq.Nsq_Rcard{
			MId:   room.Player[room.BankerId].UserId,
			Rid:   room.RoomId,
			Type:  jx_nsq.RCARD_TYPE_RETURN,
			Value: room.RoomSet.McCnt,
			CName: "chess_alg_jx",
		}
		if err := jx_nsq.NsqRcardProducer(rcard); err != nil {
			logger.Error("NsqRcardProducer: ", err)
			return
		}
	}

	for i := 0; i < 4; i++ {
		if room.Player[i].UserId != 0 {
			if room.Player[i].State != PLAYER_STATE_LEAF && room.Player[i].State != PLAYER_STATE_DOWN {
				if err := room.GameNsqProducer(SEND_DISCARD_ROOM, i, nil); err != nil {
					return
				}
			}
		}
	}
	if err := utils.RedisClient.SRem("master", strconv.FormatInt(room.MasterId, 10)); err != nil {
		logger.Error("SyncRedis SRem", err)
		return
	}
	utils.RedisClient.Del("room_" + strconv.FormatInt(room.RoomId, 10))
	utils.RedisClient.Del("return" + strconv.FormatInt(room.MasterId, 10))
	return
}

/*取消解散房间*/
func (room *Room) RcvConcelDiscardRoom(s_id int) {
	room.Player[s_id].Discard = false
	room.SyncRedis()
	for i := 1; i < 4; i++ {
		if room.Player[i].UserId != 0 {
			if room.Player[i].State != PLAYER_STATE_LEAF && room.Player[i].State != PLAYER_STATE_DOWN {
				if err := room.GameNsqProducer(SEND_DISCARD_CENCEL, (i+s_id)%4, s_id); err != nil {
					return
				}
			}
		}
	}
}

/*收到玩家掉线信息*/
func (room *Room) RcvPlayerDown(seat_id int) {
	room.Player[seat_id].State = PLAYER_STATE_DOWN
	room.SyncRedis()
	//2：通知其他的三个玩家
	for i := 0; i < 4; i++ {
		if i != seat_id && room.Player[i].State != PLAYER_STATE_DOWN {
			if err := room.GameNsqProducer(SEND_PLAYER_DOWN, i, seat_id); err != nil {
				return
			}
		}
	}

	return
}

//玩家断线重新连接的时候发给客户端数据
type reconnection struct {
	SeatId      int      `json:"seat_id"`
	UserId      int64    `json:"user_id"`
	CurrentId   int      `json:"current_id"`
	CurrentName string   `json:"current_name"`
	ChuSeatId   int      `json:"chu_seat_id"`
	Head        string   `json:"head"`
	BankId      int      `json:"bank_id"`
	Score       int      `json:"score"`
	MScore      int      `json:"m_score"`
	Hand        []int    `json:"hand"`
	Eat         []EatPai `json:"eat"`
	Alt         []int    `json:"alt"`
	Ming        []int    `json:"ming"`
	Dark        []int    `json:"dark"`
	HavePlay    []int    `json:"have_play"`
}

/*玩家掉线重连*/
func (room *Room) RcvPlayerReconnnection(s_id int) {
	if room.GameStatus == GAME_STATUS_FREE {
		if err := room.GameNsqProducer(SEND_MATCH_OVER, s_id, nil); err != nil {
			return
		}
		return
	}
	room.Player[s_id].State = PLAYER_STATE_PLAING
	for i := 0; i < 4; i++ {
		data := reconnection{
			SeatId:      s_id,
			UserId:      room.Player[i].UserId,
			CurrentId:   i,
			CurrentName: room.Player[i].UserName,
			ChuSeatId:   room.ChuPaiIndex,
			Head:        room.Player[i].Head,
			BankId:      room.BankerId,
			Score:       room.Player[i].Score,
			MScore:      room.Player[i].MScore,
			Hand:        nil,
			Eat:         room.Player[i].HandCard.Eat,
			Alt:         room.Player[i].HandCard.Alt,
			Ming:        room.Player[i].HandCard.Bright,
			Dark:        room.Player[i].HandCard.Dark,
			HavePlay:    room.Player[i].Played,
		}
		if i == s_id {
			data.Hand = room.Player[i].HandCard.ReverChange()
		}
		if err := room.GameNsqProducer(SEND_PLAYER_RECONNECTION, s_id, data); err != nil {
			return
		}
	}
	for i := 0; i < 4; i++ {
		if i != s_id {
			if err := room.GameNsqProducer(SEND_PLAYER_UP, i, s_id); err != nil {
				return
			}
		}
	}
	if err := room.GameNsqProducer(SEND_SYNC_POINT, s_id, nil); err != nil {
		return
	}
	room.SyncRedis()

	for i := 0; i < len(room.MayMes); i++ {
		if room.MayMes[i].SeatId == s_id {
			for i := 0; i < len(room.CurrentMes); i++ {
				if room.CurrentMes[i].SeatId == s_id {
					return
				}
			}
			if err := room.GameNsqProducer(SEND_MAY_ACTION, s_id, []int{room.MayMes[i].MessType, room.SharePai}); err != nil {
				return
			}
			return
		}
	}

	return
}

/*房间进入玩家, 并确定玩家的座位号*/
func (room *Room) PlayerIntoRoom(user_id int64, user_name string, head string) (seatid int) {
	//1: 首先判断玩家是否已经进入过房间
	for i := 0; i < 4; i++ {
		if room.Player[i].UserId == user_id {
			seatid = i
			room.Player[i].State = PLAYER_STATE_READY
			room.Player[i].UserId = user_id
			room.Player[i].UserName = user_name
			room.Player[i].Head = head
			room.Player[i].MayHu = true
			return
		}
	}
	//2: 玩家还没进入过此房间, 优先放空位置, 如果没有空位置，放在离开玩家的位置
	for i := 0; i < 4; i++ {
		if room.Player[i].UserId == 0 {
			seatid = i
			room.Player[i] = Player{
				UserId:    user_id,
				UserName:  user_name,
				Head:      head,
				Score:     0,
				MScore:    0,
				MayHu:     true,
				Discard:   false,
				State:     PLAYER_STATE_READY,
				DrawTimes: 0,
				Opt:       nil,
				HandCard: Hand{
					Pai: [][]int{{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
					Eat:    nil,
					Alt:    nil,
					Bright: nil,
					Dark:   nil,
				},
				WinData: Win{
					Selfdrawn: false,
					Award:     nil,
					Share:     0,
					Loser:     -1,
				},
				ScoreWater: nil,
				Played:     nil,
			}
			//清空已有玩家的数据
			for j := 0; j < 4; j++ {
				if room.Player[j].UserId != 0 && room.Player[j].State != PLAYER_STATE_LEAF {
					room.Player[j].Score = 0
				}
			}
			return seatid
		}
	}
	//3：没有空位置
	for i := 1; i < 4; i++ {
		if room.Player[i].State == PLAYER_STATE_LEAF {
			seatid = i
			room.Player[i] = Player{
				UserId:    user_id,
				UserName:  user_name,
				Head:      head,
				Score:     0,
				MScore:    0,
				MayHu:     true,
				State:     PLAYER_STATE_READY,
				DrawTimes: 0,
				Opt:       nil,
				HandCard: Hand{
					Pai: [][]int{{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
					Eat:    nil,
					Alt:    nil,
					Bright: nil,
					Dark:   nil,
				},
				WinData: Win{
					Selfdrawn: false,
					Award:     nil,
					Share:     0,
					Loser:     -1,
				},
				ScoreWater: nil,
				Played:     nil,
			}
			//清空已有玩家的数据
			for j := 0; j < 4; j++ {
				room.Player[j].Score = 0
			}
			return seatid
		}
	}
	return
}

/*收到玩家进入房间消息*/
func (room *Room) RcvPlayerInRoom(user_id int64, user_name string, head string) {
	s_id := room.PlayerIntoRoom(user_id, user_name, head)
	room.SyncRedis()
	room_log := jx_nsq.Nsq_room_log{
		MId:    user_id,
		RlType: ROOM_ACTION_TYPE_JOIN,
		CName:  "chess_alg_jx",
	}
	if err := jx_nsq.NsqRoomLogProducer(room_log); err != nil {
		logger.Error("NsqMatchCardProducer: ", err)
		return
	}
	for i := 0; i < 4; i++ {
		if room.Player[i].UserId != 0 && room.Player[i].State != PLAYER_STATE_LEAF {
			if s_id == i {
				//1：将此玩家信息告诉已经在房间的玩家, 包括自己
				for j := 0; j < 4; j++ {
					room.CurrentIndex = j
					if room.Player[j].UserId != 0 && room.Player[j].State != PLAYER_STATE_LEAF {
						status := false
						if room.Player[j].State == PLAYER_STATE_READY {
							status = true
						}
						if err := room.GameNsqProducer(SEND_PLAYER_IN_ROOM, s_id, Player_data{
							UserId:     room.Player[j].UserId,
							UserName:   room.Player[j].UserName,
							UserHead:   room.Player[j].Head,
							UserSeat:   j,
							UserStatus: status,
						}); err != nil {
							return
						}
					}
				}
			} else {
				//2：将已经在房间的玩家告诉刚进入房间的玩家
				room.CurrentIndex = s_id
				if err := room.GameNsqProducer(SEND_PLAYER_IN_ROOM, i, Player_data{
					UserId:     room.Player[s_id].UserId,
					UserName:   room.Player[s_id].UserName,
					UserHead:   room.Player[s_id].Head,
					UserSeat:   s_id,
					UserStatus: true,
				}); err != nil {
					return
				}
			}
		}
	}
	//发送每个玩家的总分
	for i := 0; i < 4; i++ {
		room.CurrentIndex = i
		if room.Player[i].UserId != 0 && room.Player[i].State != PLAYER_STATE_LEAF {
			for j := 0; j < 4; j++ {
				if room.Player[j].UserId != 0 && room.Player[j].State != PLAYER_STATE_LEAF {
					if err := room.GameNsqProducer(SEND_TOTAL_SCORE, j, room.Player[i].Score); err != nil {
						return
					}
				}
			}
		}
	}

	for i := 0; i < 4; i++ {
		if room.Player[i].UserId == 0 || room.Player[i].State != PLAYER_STATE_READY {
			return
		}
	}
	room.StartGame()
	return
}

/*玩家退出房间*/
func (room *Room) PlayerOutRoom(s_id int) {
	room_log := jx_nsq.Nsq_room_log{
		MId:    room.Player[s_id].UserId,
		RlType: ROOM_ACTION_TYPE_LEAVE,
		CName:  "chess_alg_jx",
	}
	if err := jx_nsq.NsqRoomLogProducer(room_log); err != nil {
		logger.Error("NsqMatchCardProducer: ", err)
		return
	}
	for i := 0; i < 4; i++ {
		if room.Player[i].UserId != 0 && room.Player[i].State != PLAYER_STATE_LEAF {
			if err := room.GameNsqProducer(SEND_PLAYER_OUT_ROOM, i, s_id); err != nil {
				return
			}
		}
	}
	//玩家退出房间先保留玩家数据
	room.Player[s_id].State = PLAYER_STATE_LEAF
	room.SyncRedis()

	if room.GameNum == room.RoomSet.McCnt*8 {
		for i := 0; i < 4; i++ {
			if room.Player[i].State != PLAYER_STATE_LEAF {
				return
			}
		}
		logger.Debug("---八局完成--玩家退出---解散房间-------")
		room_log := jx_nsq.Nsq_room_log{
			MId:    room.Player[s_id].UserId,
			RlType: ROOM_ACTION_TYPE_DISCARD,
			CName:  "chess_alg_jx",
		}
		if err := jx_nsq.NsqRoomLogProducer(room_log); err != nil {
			logger.Error("NsqMatchCardProducer: ", err)
			return
		}

		if err := utils.RedisClient.SRem("master", strconv.FormatInt(room.MasterId, 10)); err != nil {
			logger.Error("SyncRedis SRem", err)
			return
		}
		utils.RedisClient.Del("room_" + strconv.FormatInt(room.RoomId, 10))
		utils.RedisClient.Del("return" + strconv.FormatInt(room.MasterId, 10))
	}
}

//收到玩家洗牌完成
func (room *Room) RcvPlayerShuffled(s_id int) {
	if room.Player[s_id].State != PLAYER_STATE_FREE {
		return
	}
	room.Player[s_id].State = PLAYER_STATE_SHUFFLED
	room.SyncRedis()
	for i := 0; i < 4; i++ {
		if room.Player[i].State != PLAYER_STATE_SHUFFLED {
			return
		}
	}
	timer.DelTimer("SHUFFLE")
	room.DecideBanker()
}

//收到玩家定庄色子效果完成
func (room *Room) RcvPlayerDiceBanker(s_id int) {
	if room.Player[s_id].State != PLAYER_STATE_SHUFFLED {
		return
	}
	room.Player[s_id].State = PLAYER_STATE_DICE_BANKERED
	room.SyncRedis()

	for i := 0; i < 4; i++ {
		if room.Player[i].State != PLAYER_STATE_DICE_BANKERED {
			return
		}
	}
	timer.DelTimer("DECIDEBANKER")
	//定庄完成后首轮发牌
	room.PreRoundDeal()
	//发牌后定精牌
	room.DecideUniverse()
}

//收到玩家展示精牌效果完成
func (room *Room) RcvPlayerShowUniverse(s_id int) {
	if room.Player[s_id].State != PLAYER_STATE_DICE_BANKERED {
		return
	}
	room.Player[s_id].State = PLAYER_STATE_SHOW_UNIVERSE_END
	for i := 0; i < 4; i++ {
		if room.Player[i].State != PLAYER_STATE_SHOW_UNIVERSE_END {
			room.SyncRedis()
			return
		}
	}
	timer.DelTimer("DECIDEUNIVERSE")
	//如果时开局上下翻
	if room.RoomSet.TurnRound == true {
		room.CreateOpeningScore()
	}
	time.Sleep(2 * time.Second)
	room.CurrentIndex = room.BankerId
	room.MoPaiIndex = room.BankerId
	room.Deal()
}

/*
收到玩家出牌信息
	1：关闭上次定时器
	2：首先判断手牌中是否已经听牌，修改听牌标志，然后将玩家所出的牌从手牌中除去
	3：设置room.sharePai为当前出的牌
	4：修改控牌玩家ID，对控牌玩家ID做相应的判断
*/
func (room *Room) RcvPlayAHand(s_id int, pai int) {
	if room.Player[s_id].HandCard.Pai[pai/10][pai%10] == 0 {
		room.syncPlayerHander(s_id)
		return
	}
	room.CurrentIndex = s_id
	room.ChuPaiIndex = s_id
	//将玩家出的牌从玩家手牌中除取
	room.Player[s_id].HandCard.Pai[pai/10][0] -= 1
	room.Player[s_id].HandCard.Pai[pai/10][pai%10] -= 1
	//更新玩家游戏状态
	room.Player[s_id].Opt = append(room.Player[s_id].Opt, PLAYING_STATUS_PLAY)
	room.Player[s_id].Played = append(room.Player[s_id].Played, pai)
	//设置当前麻将最后出的牌值
	room.SharePai = pai

	//判断是否抄庄
	if s_id == (room.BankerId+3)%4 && room.Player[s_id].DrawTimes == 1 &&
		len(room.Player[(s_id+1)%4].Played) == 1 && pai == room.Player[(s_id+1)%4].Played[0] &&
		len(room.Player[(s_id+2)%4].Played) == 1 && pai == room.Player[(s_id+2)%4].Played[0] &&
		len(room.Player[(s_id+3)%4].Played) == 1 && pai == room.Player[(s_id+3)%4].Played[0] {
		//抄庄，游戏结束
		room.CreateSearchBankerScore()
		room.GameOver()
		return
	}
	//nsq通知其他玩家
	for i := 0; i < 4; i++ {
		if err := room.GameNsqProducer(SEND_PLAY_A_HAND, i, pai); err != nil {
			return
		}
	}
	//对其他玩家进行判断
	room.JudgeOther()
	room.SyncRedis()
}

/*处理玩家吃牌消息*/
func (room *Room) processEat(s_id int, pai []int) {
	room.MayMes = room.MayMes[:0]
	room.CurrentIndex = s_id
	/*告诉所有玩家有玩家吃牌，用于渲染桌面*/
	for i := 0; i < 4; i++ {
		if err := room.GameNsqProducer(SEND_EAT, i, []int{room.ChuPaiIndex, pai[0], pai[1], pai[2]}); err != nil {
			return
		}
		if err := room.GameNsqProducer(SEND_SYNC_POINT, i, nil); err != nil {
			return
		}
	}
	room.Player[room.ChuPaiIndex].Played = room.Player[room.ChuPaiIndex].Played[:len(room.Player[room.ChuPaiIndex].Played)-1]
	room.Player[s_id].HandCard.Eat = append(room.Player[s_id].HandCard.Eat, pai)
	room.Player[s_id].Opt = append(room.Player[s_id].Opt, PLAYING_STATUS_EAT)
	room.Player[s_id].MayHu = true
	for _, v := range pai {
		if v != room.SharePai {
			room.Player[s_id].HandCard.Pai[v/10][v%10]--
			room.Player[s_id].HandCard.Pai[v/10][0]--
		}
	}
	room.SyncRedis()
	//玩家吃牌后检查当前牌中是否有暗杠
	if dackBar := room.Player[s_id].HandCard.IsConcealedKong(); dackBar != 0 {
		if err := room.GameNsqProducer(SEND_MAY_ACTION, s_id, []int{ACTION_DACK_BAR, dackBar}); err != nil {
			return
		}
	}

	room.SyncRedis()
	return
}

/*收到玩家吃牌消息*/
func (room *Room) RcvEat(s_id int, pai []int) {
	//检测当前消息是否重复
	for i := 0; i < len(room.CurrentMes); i++ {
		if room.CurrentMes[i].SeatId == s_id {
			return
		}
	}

	if len(room.MayMes) == 1 {
		room.processEat(s_id, pai)
	} else {
		room.CurrentMes = append(room.CurrentMes, Monitor{SeatId: s_id, MessType: ACTION_EAT, Data: pai})
		if len(room.MayMes) == len(room.CurrentMes) {
			room.ProcessJudge()
		}
	}
	room.SyncRedis()
	return
}

/*处理玩家碰牌消息*/
func (room *Room) ProcessAlt(s_id int) {
	room.CurrentIndex = s_id
	for i := 0; i < 4; i++ {
		if err := room.GameNsqProducer(SEND_SYNC_POINT, i, nil); err != nil {
			return
		}
		if err := room.GameNsqProducer(SEND_ALT, i, []int{room.ChuPaiIndex, room.SharePai}); err != nil {
			return
		}
	}

	room.Player[room.ChuPaiIndex].Played = room.Player[room.ChuPaiIndex].Played[:len(room.Player[room.ChuPaiIndex].Played)-1]

	room.Player[s_id].MayHu = true
	room.Player[s_id].HandCard.Alt = append(room.Player[s_id].HandCard.Alt, room.SharePai)
	room.Player[s_id].Opt = append(room.Player[s_id].Opt, PLAYING_STATUS_ALT)
	room.Player[s_id].HandCard.Pai[room.SharePai/10][room.SharePai%10] -= 2
	room.Player[s_id].HandCard.Pai[room.SharePai/10][0] -= 2
	room.SyncRedis()

	if dackBar := room.Player[room.CurrentIndex].HandCard.IsConcealedKong(); dackBar != 0 {
		if err := room.GameNsqProducer(SEND_MAY_ACTION, room.CurrentIndex, []int{ACTION_DACK_BAR, dackBar}); err != nil {
			return
		}
	}
	return
}

/*
收到玩家碰牌信息：
	1：关闭上次定时器
	2：将当前出的牌添加入碰牌结构，
 	3：将当前控牌玩家的此张牌从手牌中除去
 	4：将此玩家碰牌消息通知给其他三个玩家
*/
func (room *Room) RcvAlt(s_id int) {
	for i := 0; i < len(room.CurrentMes); i++ {
		if room.CurrentMes[i].SeatId == s_id {
			return
		}
	}
	if len(room.MayMes) == 1 {
		room.ProcessAlt(s_id)
	} else {
		for i := 0; i < len(room.MayMes); i++ {
			if (room.MayMes[i].MessType & ACTION_HU) != 0 {
				room.CurrentMes = append(room.CurrentMes, Monitor{SeatId: s_id, MessType: ACTION_ALT})
				if len(room.MayMes) == len(room.CurrentMes) {
					room.ProcessJudge()
				}
				room.SyncRedis()
				return
			}
		}
		room.ProcessAlt(s_id)
	}
	room.MayMes = room.MayMes[:0]
	room.CurrentMes = room.CurrentMes[:0]
	room.SyncRedis()
	return
}

/*
收到玩家明杠-加杠--可抢杠-信息：
	1：关闭上次定时器
	2：修改玩家手牌结构
  	3：为玩家继续发牌
  	4：抢杠---------------------
*/
func (room *Room) RcvAddBar(s_id int) {
	//更新玩家游戏状态
	room.ChuPaiIndex = s_id
	room.CurrentIndex = s_id
	room.Player[s_id].Opt = append(room.Player[s_id].Opt, PLAYING_STATUS_ADD_BAR)
	room.Player[s_id].MayHu = true
	room.Player[s_id].HandCard.Pai[room.SharePai/10][0] -= 1
	room.Player[s_id].HandCard.Pai[room.SharePai/10][room.SharePai%10] -= 1
	for i := 0; i < 4; i++ {
		if err := room.GameNsqProducer(SEND_ADD_BAR, i, room.SharePai); err != nil {
			return
		}
	}
	//检测玩家是否可以抢杠
	for i := 1; i < 4; i++ {
		index := (s_id + i) % 4
		if hu := room.Player[index].HandCard.MayHu(room.UpUniverse, room.SharePai); hu == true {
			logger.Debugf("玩家 %d 可以抢杠", index)
			room.MayMes = append(room.MayMes, Monitor{SeatId: index, MessType: ACTION_HU})
			if err := room.GameNsqProducer(SEND_MAY_ACTION, index, []int{ACTION_HU, room.SharePai}); err != nil {
				return
			}
		}
	}
	if len(room.MayMes) > 0 {
		room.GameStatus = GAME_STATUS_GRAB_BAR //更改游戏状态为抢杠
		room.SyncRedis()
		return
	}
	room.CreateBrightBarScore()
	room.MoPaiIndex = room.CurrentIndex
	room.SyncRedis()
	room.Deal()
}

/*收到暗杠信息*/
func (room *Room) RcvDarkBar(s_id int, pai int) {
	room.CurrentIndex = s_id
	for i := 0; i < 4; i++ {
		if err := room.GameNsqProducer(SEND_SYNC_POINT, i, nil); err != nil {
			return
		}
		if err := room.GameNsqProducer(SEND_DARK_BAR, i, pai); err != nil {
			return
		}
	}
	//更新玩家游戏状态
	room.Player[s_id].Opt = append(room.Player[room.CurrentIndex].Opt, PLAYING_STATUS_DARK_BAR)
	room.Player[room.CurrentIndex].MayHu = true
	room.Player[room.CurrentIndex].HandCard.Dark = append(room.Player[room.CurrentIndex].HandCard.Dark, pai)
	room.Player[room.CurrentIndex].HandCard.Pai[pai/10][pai%10] -= 4
	room.Player[room.CurrentIndex].HandCard.Pai[pai/10][0] -= 4
	room.CreateDardBarScore()
	room.MoPaiIndex = room.CurrentIndex
	room.SyncRedis()
	room.Deal()
}

/*处理玩家明杠*/
func (room *Room) ProcessBrightBar(s_id int) {
	room.CurrentIndex = s_id
	for i := 0; i < 4; i++ {
		if err := room.GameNsqProducer(SEND_SYNC_POINT, i, nil); err != nil {
			return
		}
		if err := room.GameNsqProducer(SEND_BRIGHT_BAR, i, []int{room.ChuPaiIndex, room.SharePai}); err != nil {
			return
		}
	}
	//更新出牌玩家信息
	room.Player[room.ChuPaiIndex].Played = room.Player[room.ChuPaiIndex].Played[:len(room.Player[room.ChuPaiIndex].Played)-1]
	//更新杠牌玩家手中杠牌的数据
	room.Player[s_id].HandCard.Bright = append(room.Player[s_id].HandCard.Bright, room.SharePai)
	room.Player[room.CurrentIndex].HandCard.Pai[room.SharePai/10][room.SharePai%10] -= 3
	room.Player[room.CurrentIndex].HandCard.Pai[room.SharePai/10][0] -= 3
	room.Player[room.CurrentIndex].Opt = append(room.Player[room.CurrentIndex].Opt, PLAYING_STATUS_BRIGHT_BAR)
	room.CreateBrightBarScore()
	room.MoPaiIndex = room.CurrentIndex
	room.SyncRedis()
	room.Deal()
}

/*收到明杠信息*/
func (room *Room) RcvBrightBar(s_id int) {
	for i := 0; i < len(room.CurrentMes); i++ {
		if room.CurrentMes[i].SeatId == s_id {
			return
		}
	}

	if len(room.MayMes) == 1 {
		room.ProcessBrightBar(s_id)
	} else {
		for i := 0; i < len(room.MayMes); i++ {
			if (room.MayMes[i].MessType & ACTION_HU) != 0 {
				room.CurrentMes = append(room.CurrentMes, Monitor{SeatId: s_id, MessType: ACTION_BRIGHT_BAR})
				if len(room.MayMes) == len(room.CurrentMes) {
					room.ProcessJudge()
				}
				room.SyncRedis()
				return
			}
		}
		room.ProcessBrightBar(s_id)
	}
	room.MayMes = room.MayMes[:0]
	room.CurrentMes = room.CurrentMes[:0]
	room.SyncRedis()
	return
}

/*收到玩家 过 信息*/
func (room *Room) RcvPass(s_id int) {
	for i := 0; i < len(room.CurrentMes); i++ {
		if room.CurrentMes[i].SeatId == s_id {
			return
		}
	}

	for i := 0; i < len(room.MayMes); i++ {
		if room.MayMes[i].SeatId == s_id && room.MayMes[i].MessType&ACTION_HU != 0 {
			room.Player[s_id].MayHu = false
		}
	}
	if room.GameStatus == GAME_STATUS_CONTROL {
		//玩家控牌阶段收到过
		if len(room.MayMes) == 1 {
			room.MayMes = room.MayMes[:0]
			room.CurrentIndex = (room.ChuPaiIndex + 1) % 4
			room.MoPaiIndex = room.CurrentIndex
			room.SyncRedis()
			room.Deal()
			return
		} else {
			room.CurrentMes = append(room.CurrentMes, Monitor{SeatId: s_id, MessType: ACTION_PASS})
			if len(room.MayMes) == len(room.CurrentMes) {
				room.ProcessJudge()
			}
			room.SyncRedis()
			return
		}
	} else if room.GameStatus == GAME_STATUS_GRAB_BAR {
		if len(room.MayMes) == 1 {
			room.CreateBrightBarScore()
			room.MayMes = room.MayMes[:0]
			room.MoPaiIndex = room.CurrentIndex
			room.SyncRedis()
			room.Deal()
		} else {
			room.CurrentMes = append(room.CurrentMes, Monitor{SeatId: s_id, MessType: ACTION_PASS})
			if len(room.MayMes) == len(room.CurrentMes) {
				room.ProcessJudge()
			}
			room.SyncRedis()
			return
		}
	} else if room.GameStatus == GAME_STATUS_CUT_HU {
		room.ProcessHu(room.CurrentMes[0].SeatId)
	}
}

/*处理玩家胡牌信息*/
func (room *Room) ProcessHu(s_id int) {
	//将胡的牌放入玩家手牌中做胡牌计算
	room.CurrentIndex = s_id
	if room.GameStatus == GAME_STATUS_MOPAI {
		room.Player[s_id].WinData.Selfdrawn = true
		if (s_id == room.BankerId) && room.Player[s_id].DrawTimes == 1 {
			//检查是否是天胡, 当前玩家是庄家而且第一次摸牌
			room.CreateGodLandHuScore(true)
			room.GameOver()
			return
		}
	} else if room.GameStatus == GAME_STATUS_CONTROL {
		room.Player[s_id].HandCard.Pai[room.SharePai/10][0] += 1
		room.Player[s_id].HandCard.Pai[room.SharePai/10][room.SharePai%10] += 1
		if (s_id != room.BankerId) && room.Player[s_id].DrawTimes == 0 {
			//检查是否是地胡, 旁家胡的庄家的第一张牌
			room.CreateGodLandHuScore(false)
			room.GameOver()
			return
		}
	}
	//保存玩家胡牌信息
	room.Player[s_id].WinData.Loser = room.ChuPaiIndex
	room.Player[s_id].WinData.Share = room.SharePai
	//生成玩家胡牌的流水
	room.CreatHuScore(s_id)
	for i := 0; i < 4; i++ {
		if err := room.GameNsqProducer(SEND_WIN, i, room.Player[s_id].WinData); err != nil {
			return
		}
	}
	room.GameOver()
	return
}

/*收到玩家胡牌信息：检测是否可以进入截胡阶段*/
func (room *Room) RcvHu(s_id int) {
	if room.GameStatus == GAME_STATUS_MOPAI {
		room.ProcessHu(s_id)
		return
	}
	room.MayMes = room.MayMes[:0]
	room.CurrentMes = room.CurrentMes[:0]
	room.CurrentMes = append(room.CurrentMes, Monitor{SeatId: s_id, MessType: ACTION_HU})
	for i := 1; i < 4; i++ {
		index := (room.ChuPaiIndex + i) % 4
		if index == s_id {
			break
		}
		if room.Player[index].MayHu == true {
			if hu := room.Player[index].HandCard.MayHu(room.UpUniverse, room.SharePai); hu == true {
				if room.Player[index].HandCard.isUniverseDiao(room.UpUniverse) == false {
					logger.Debug("收到玩家胡牌信息后，发现有玩家可以截胡: ")
					room.MayMes = append(room.MayMes, Monitor{SeatId: index, MessType: ACTION_HU})
				}
			}
		}
	}
	if len(room.MayMes) != 0 {
		room.SyncRedis()
		return
	}
	room.ProcessHu(s_id)
	return
}

func (room *Room) RcvBeat(s_id int) {
	if err := room.GameNsqProducer(SEND_BEAT, s_id, nil); err != nil {
		return
	}
}
