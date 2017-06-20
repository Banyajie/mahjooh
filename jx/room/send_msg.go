package room

import (
	"math/rand"
	"strconv"
	"time"

	"chess_alg_jx/jx/jx_nsq"
	"chess_alg_jx/logger"
	"chess_alg_jx/timer"
	"chess_alg_jx/utils"
)

/*1: 洗牌*/
func (room *Room) Shuffle() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < len(MajPool); i++ {
		x := r.Intn(len(MajPool))
		t := room.Mahj[len(MajPool)-i-1]
		room.Mahj[len(MajPool)-i-1] = room.Mahj[x]
		room.Mahj[x] = t
	}

	for i := 0; i < 4; i++ {
		if err := room.GameNsqProducer(SEND_START_SHUFFLE, i, nil); err != nil {
			return
		}
	}
	room.SyncRedis()
	timer.SetTimer("SHUFFLE", 10, room.checkShuffle, nil)
}

func (room *Room) checkShuffle(interface{}) {
	timer.DelTimer("SHUFFLE")
	if err := room.GetRoomData(strconv.FormatInt(room.RoomId, 10)); err != nil {
		return
	}
	for i := 0; i < 4; i++ {
		if room.Player[i].State != PLAYER_STATE_SHUFFLED {
			if err := room.GameNsqProducer(SEND_START_SHUFFLE, i, nil); err != nil {
				return
			}
		}
	}
	timer.SetTimer("SHUFFLE", 10, room.checkShuffle, nil)
	return
}

/*2: 摇赛子定庄*/
func (room *Room) DecideBanker() {
	logger.Notice("2: 定庄 ")
	//1: 产生两个小于六的随机数，定庄家
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num1 := r.Intn(6) + 1
	num2 := r.Intn(6) + 1
	room.Shai = []int{num1, num2}
	for i := 0; i < 4; i++ {
		if err := room.GameNsqProducer(SEND_DICE_BANKER, i, []int{num1, num2}); err != nil {
			return
		}
	}
	room.BankerId = (num1 + num2 - 1) % 4
	room.SyncRedis()
	timer.SetTimer("DECIDEBANKER", 10, room.checkDecideBanker, nil)
}

func (room *Room) checkDecideBanker(data interface{}) {
	timer.DelTimer("DECIDEBANKER")
	if err := room.GetRoomData(strconv.FormatInt(room.RoomId, 10)); err != nil {
		return
	}
	for i := 0; i < 4; i++ {
		if room.Player[i].State != PLAYER_STATE_DICE_BANKERED {
			if err := room.GameNsqProducer(SEND_DICE_BANKER, i, room.Shai); err != nil {
				return
			}
		}
	}
	timer.SetTimer("DECIDEBANKER", 10, room.checkDecideBanker, nil)
	return
}

/*3: 首轮发牌 每个玩家13张牌*/
func (room *Room) PreRoundDeal() {
	for k := 1; k <= 4; k++ {
		for i := 0; i < 4; i++ {
			room.CurrentIndex = (room.BankerId + i) % 4
			if k == 4 {
				value := DealOneCard(room.Mahj)
				room.LeftNum--
				room.Player[(room.BankerId+i)%4].HandCard.Pai[value/10][value%10] += 1
				room.Player[(room.BankerId+i)%4].HandCard.Pai[value/10][0] += 1
				if err := room.GameNsqProducer(SEND_PREROUND_DEAL, (room.BankerId+i)%4, value); err != nil {
					return
				}
			} else {
				var tmp []int
				for j := 0; j < 4; j++ {
					value := DealOneCard(room.Mahj)
					room.LeftNum--
					tmp = append(tmp, value)
					room.Player[(room.BankerId+i)%4].HandCard.Pai[value/10][value%10] += 1
					room.Player[(room.BankerId+i)%4].HandCard.Pai[value/10][0] += 1
				}
				if err := room.GameNsqProducer(SEND_PREROUND_DEAL, (room.BankerId+i)%4, tmp); err != nil {
					return
				}
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
	time.Sleep(1 * time.Second)
	//检测回头一笑
	if room.RoomSet.SmileBack == true && room.GameNum > 1 {
		room.CreateSmileScore()
		universe := jx_nsq.Nsq_Universe{
			McNo:    room.McNo,
			Type:    jx_nsq.UNIVERSE_TYPE_SMILE,
			Manager: room.UpUniverse,
			Deputy:  NextOrder(room.UpUniverse),
			CName:   "chess_alg_jx",
		}
		if err := jx_nsq.NsqUniverseProducer(universe); err != nil {
			logger.Error("NsqUniverseProducer: ", err)
			return
		}
	}
	room.SyncRedis()
}

/*选定精牌*/
func (room *Room) DecideUniverse() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num1 := r.Intn(6) + 1
	num2 := r.Intn(6) + 1
	room.Shai = []int{num1, num2}
	for i := 0; i < 4; i++ {
		if err := room.GameNsqProducer(SEND_DICE_UNIVERSE, i, []int{num1, num2}); err != nil {
			return
		}
	}
	room.UpUniverse = room.Mahj[len(MajPool)-(num1+num2)*2]
	universe := jx_nsq.Nsq_Universe{
		McNo:    room.McNo,
		Type:    jx_nsq.UNIVERSE_TYPE_UP,
		Manager: room.UpUniverse,
		Deputy:  NextOrder(room.UpUniverse),
		CName:   "chess_alg_jx",
	}

	if err := jx_nsq.NsqUniverseProducer(universe); err != nil {
		logger.Error("NsqUniverseProducer: ", err)
		return
	}
	if room.RoomSet.DownUniverse == false {
		room.DownUniverse = room.Mahj[len(MajPool)-(num1+num2)*2-1]
		universe := jx_nsq.Nsq_Universe{
			McNo:    room.McNo,
			Type:    jx_nsq.UNIVERSE_TYPE_DOWN,
			Manager: room.DownUniverse,
			Deputy:  NextOrder(room.DownUniverse),
			CName:   "chess_alg_jx",
		}

		if err := jx_nsq.NsqUniverseProducer(universe); err != nil {
			logger.Error("NsqUniverseProducer: ", err)
			return
		}
	}
	//定精色子完成后发送精牌
	for i := 0; i < 4; i++ {
		if err := room.GameNsqProducer(SEND_UNIVERSE, i, []int{room.UpUniverse, room.DownUniverse}); err != nil {
			return
		}
	}
	room.SyncRedis()
	timer.SetTimer("DECIDEUNIVERSE", 10, room.checkDecideUniverse, nil)
}

func (room *Room) checkDecideUniverse(interface{}) {
	timer.DelTimer("DECIDEUNIVERSE")
	if err := room.GetRoomData(strconv.FormatInt(room.RoomId, 10)); err != nil {
		return
	}
	for i := 0; i < 4; i++ {
		if room.Player[i].State != PLAYER_STATE_SHOW_UNIVERSE_END {
			//重新展示精牌
			if err := room.GameNsqProducer(SEND_DICE_UNIVERSE, i, room.Shai); err != nil {
				return
			}
		}
	}
	timer.SetTimer("DECIDEUNIVERSE", 10, room.checkDecideUniverse, nil)
	return
}

/*
为当前控牌玩家发牌
	1：首先检测当前控牌玩家是否已经胡牌，如果胡牌则跳过此玩家
	2：为此玩家发牌后，判断此玩家当前可以做：胡/听/刮风/暗杠判断
	3：开启定时器，等待玩家做出回应，如超时则默认出牌策略
*/
func (room *Room) Deal() {
	logger.Debugf("为 %d 玩家发牌，等待玩家回应： ", room.MoPaiIndex)
	dealPai := DealOneCard(room.Mahj)
	if dealPai == 0 {
		room.GameOver()
		return
	}
	room.SharePai = dealPai
	room.LeftNum--
	//将牌发送给玩家
	for i := 0; i < 4; i++ {
		if i == room.MoPaiIndex {
			if err := room.GameNsqProducer(SEND_DEAL, i, dealPai); err != nil {
				return
			}
		} else {
			if err := room.GameNsqProducer(SEND_DEAL, i, 0); err != nil {
				return
			}
		}
		//告诉其他三个玩家当前控牌ID
		if err := room.GameNsqProducer(SEND_SYNC_POINT, i, 1); err != nil {
			return
		}
	}
	room.GameStatus = GAME_STATUS_MOPAI
	room.Player[room.MoPaiIndex].MayHu = true
	//将新发的牌放入当前摸牌玩家的手牌中
	room.Player[room.MoPaiIndex].HandCard.Pai[dealPai/10][0]++
	room.Player[room.MoPaiIndex].HandCard.Pai[dealPai/10][dealPai%10]++
	//修改玩家摸牌次数
	room.Player[room.MoPaiIndex].DrawTimes++
	room.CurrentIndex = room.MoPaiIndex

	value := 0
	//玩家可以胡牌
	if hu := room.Player[room.MoPaiIndex].HandCard.MayHu(room.UpUniverse, 0); hu == true {
		value = value | ACTION_HU
	}
	//玩家可以加杠 -- 已经碰过当前牌
	if ok, _ := utils.Contain(dealPai, room.Player[room.MoPaiIndex].HandCard.Alt); ok == true {
		value = value | ACTION_ADD_BAR
	}
	//玩家可以暗杠 -- 玩家手中有四张当前牌/或者有其余已经是四张牌
	dack := 0
	if dack = room.Player[room.MoPaiIndex].HandCard.IsConcealedKong(); dack != 0 {
		value = value | ACTION_DACK_BAR
	}

	if value != 0 {
		if dack != 0 {
			if err := room.GameNsqProducer(SEND_MAY_ACTION, room.MoPaiIndex, []int{value, dack}); err != nil {
				return
			}
		} else {
			if err := room.GameNsqProducer(SEND_MAY_ACTION, room.MoPaiIndex, []int{value, room.SharePai}); err != nil {
				return
			}
		}
	}

	room.SyncRedis()
	return
}

/*
本局游戏结束
	1：计算每个玩家的输赢
	2：保存游戏记录
	3：初始化房间结构对下一局游戏准备
*/
func (room *Room) GameOver() {
	if err := room.GetRoomData(strconv.FormatInt(room.RoomId, 10)); err != nil {
		return
	}
	if room.RoomSet.LandMines == true {
		room.CreateLandMinesScore()
	} else {
		room.CreateUniverseScore()
	}
	room.GameStatus = GAME_STATUS_FREE
	room.SyncRedis()

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			room.CurrentIndex = j
			if err := room.GameNsqProducer(SEND_GAME_OVER, i, room.Player[j].ScoreWater); err != nil {
				return
			}
			if err := room.GameNsqProducer(SEND_PLAYER_HAND, i, room.Player[j].HandCard.ReverChange()); err != nil {
				return
			}
		}
	}
	//将玩家手排插入数据库
	room.InsertUserMatchCards()
	//更新玩家战绩数据库
	for i := 0; i < 4; i++ {
		data := jx_nsq.Nsq_match_record{
			MId:   room.Player[i].UserId,
			Score: room.Player[i].MScore,
			CName: "chess_alg_jx",
		}
		if err := jx_nsq.NsqMatchRecordProducer(data); err != nil {
			logger.Error("NsqMatchCardProducer: ", err)
			return
		}
	}
	data := jx_nsq.Nsq_match_sum{
		McNo:     room.McNo,
		EastMid:  room.Player[0].UserId,
		EastSum:  room.Player[0].MScore,
		SouthMid: room.Player[1].UserId,
		SouthSum: room.Player[1].MScore,
		WestMid:  room.Player[2].UserId,
		WestSum:  room.Player[2].MScore,
		NorthMid: room.Player[3].UserId,
		NorthSum: room.Player[3].MScore,
		CName:    "chess_alg_jx",
	}
	if err := jx_nsq.NsqMatchSumProducer(data); err != nil {
		logger.Error("NsqMatchSumProducer: ", err)
		return
	}
	/*初始化房间数据----开始下一局游戏*/
	if room.GameNum == room.RoomSet.McCnt*8 {
		for i := 0; i < 4; i++ {
			if err := room.GameNsqProducer(SEND_MATCH_USEUP, i, nil); err != nil {
				return
			}
		}
		return
	}
	room.InitRoomGame()
	return
}

func (room *Room) InsertUserMatchCards() {
	//1：保存玩家本局游戏中结束时的手牌
	for i := 0; i < 4; i++ {
		//1：手牌
		if room.Player[i].HandCard.Pai != nil {
			data := jx_nsq.Nsq_match_card{
				McNo:      room.McNo,
				MId:       room.Player[i].UserId,
				HandCards: utils.ReverToString(room.Player[i].HandCard.ReverChange()),
				CardType:  HANDCARDS_TYPE_HAND,
				CName:     "chess_alg_jx",
			}
			if err := jx_nsq.NsqMatchCardProducer(data); err != nil {
				logger.Error("NsqMatchCardProducer: ", err)
				return
			}
		}
		//2：碰牌
		if len(room.Player[i].HandCard.Alt) != 0 {
			data := jx_nsq.Nsq_match_card{
				McNo:      room.McNo,
				MId:       room.Player[i].UserId,
				HandCards: utils.ReverToString(room.Player[i].HandCard.Alt),
				CardType:  HANDCARDS_TYPE_ALT,
				CName:     "chess_alg_jx",
			}
			if err := jx_nsq.NsqMatchCardProducer(data); err != nil {
				logger.Error("NsqMatchCardProducer: ", err)
				return
			}
		}
		//3：明杠
		if len(room.Player[i].HandCard.Bright) != 0 {
			data := jx_nsq.Nsq_match_card{
				McNo:      room.McNo,
				MId:       room.Player[i].UserId,
				HandCards: utils.ReverToString(room.Player[i].HandCard.Bright),
				CardType:  HANDCARDS_TYPE_BRIGHT,
				CName:     "chess_alg_jx",
			}
			if err := jx_nsq.NsqMatchCardProducer(data); err != nil {
				logger.Error("NsqMatchCardProducer: ", err)
				return
			}
		}
		//4：暗杠
		if room.Player[i].HandCard.Dark != nil {
			data := jx_nsq.Nsq_match_card{
				McNo:      room.McNo,
				MId:       room.Player[i].UserId,
				HandCards: utils.ReverToString(room.Player[i].HandCard.Dark),
				CardType:  HANDCARDS_TYPE_DARK,
				CName:     "chess_alg_jx",
			}
			if err := jx_nsq.NsqMatchCardProducer(data); err != nil {
				logger.Error("NsqMatchCardProducer: ", err)
				return
			}
		}
		//5：吃牌
		if room.Player[i].HandCard.Eat != nil {
			for _, data := range room.Player[i].HandCard.Eat {
				data := jx_nsq.Nsq_match_card{
					McNo:      room.McNo,
					MId:       room.Player[i].UserId,
					HandCards: utils.ReverToString(data),
					CardType:  HANDCARDS_TYPE_EAT,
					CName:     "chess_alg_jx",
				}
				if err := jx_nsq.NsqMatchCardProducer(data); err != nil {
					logger.Error("NsqMatchCardProducer: ", err)
					return
				}
			}
		}
	}
}
