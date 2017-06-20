package room

import (
	"strconv"

	"chess_alg_jx/logger"
	"chess_alg_jx/timer"
)

/*
	当玩家出牌后，对其他玩家进行判断
	：玩家是否可以吃、胡，碰，明杠
	：默认过
*/
func (room *Room) JudgeOther() {
	room.MayMes = room.MayMes[:0]
	for i := 1; i < 4; i++ {
		index := (room.ChuPaiIndex + i) % 4
		value := 0
		//判断当前玩家是否可以吃
		if i == 1 {
			if eat := room.Player[index].HandCard.IsMayEat(room.SharePai); eat == true {
				logger.Warningf("座位 %d 玩家可以吃  ", index)
				value = value | ACTION_EAT
			}
		}
		//判断当前玩家是否可以碰
		if room.Player[index].HandCard.Pai[room.SharePai/10][room.SharePai%10] >= 2 {
			logger.Warningf("座位 %d 玩家可以碰  ", index)
			value = value | ACTION_ALT
		}
		//判断当前玩家是否可以明杠
		if room.Player[index].HandCard.Pai[room.SharePai/10][room.SharePai%10] == 3 {
			logger.Warningf("座位 %d 玩家可以明杠  ", index)
			value = value | ACTION_BRIGHT_BAR
		}
		//将牌放入手牌中判断是否可以胡牌 / 如果是精钓则不能胡牌
		if room.Player[index].MayHu == true {
			if hu := room.Player[index].HandCard.MayHu(room.UpUniverse, room.SharePai); hu == true {
				logger.Warningf("座位 %d 玩家可以胡牌  ", index)
				if room.Player[index].HandCard.isUniverseDiao(room.UpUniverse) == false {
					value = value | ACTION_HU
				} else {
					logger.Warning("精调－－不能胡牌")
				}
			}
		}

		if value != 0 {
			if err := room.GameNsqProducer(SEND_MAY_ACTION, index, []int{value, room.SharePai}); err != nil {
				return
			}
			room.MayMes = append(room.MayMes, Monitor{SeatId: index, MessType: value})
		}
	}
	if len(room.MayMes) == 0 {
		room.CurrentIndex = (room.ChuPaiIndex + 1) % 4
		room.MoPaiIndex = room.CurrentIndex
		room.Deal()
	} else {
		for i := 0; i < 4; i++ {
			if err := room.GameNsqProducer(SEND_RESETTIMER, i, nil); err != nil {
				return
			}
		}
		room.GameStatus = GAME_STATUS_CONTROL
		room.SyncRedis()
	}
	return
}

/*处理多个玩家回复阶段*/
func (room *Room) ProcessJudge() {
	logger.Debugf("处理多个玩家回复阶段---", room.MayMes, room.CurrentMes)
	timer.Delete(timer.TimerMap[strconv.FormatInt(room.RoomId, 10)])
	//1:处理收到的每条消息，首先有碰/杠则不处理吃 如果没有则处理吃, 如果都是pass则继续游戏
	flag := false
	for i := 0; i < len(room.CurrentMes); i++ {
		if room.CurrentMes[i].MessType == ACTION_ALT {
			flag = true
			//处理碰牌
			room.ProcessAlt(room.CurrentMes[i].SeatId)
			//2:初始化room.CurrentMes room.CurrentNum
			room.MayMes = room.MayMes[:0]
			room.CurrentMes = room.CurrentMes[:0]
			room.SyncRedis()
			return
		} else if room.CurrentMes[i].MessType == ACTION_BRIGHT_BAR {
			//处理明杠
			flag = true
			room.ProcessBrightBar(room.CurrentMes[i].SeatId)
			//2:初始化room.CurrentMes room.CurrentNum
			room.MayMes = room.MayMes[:0]
			room.CurrentMes = room.CurrentMes[:0]
			room.SyncRedis()
			return
		}
	}
	eat_flag := false
	if flag == false {
		for i := 0; i < len(room.CurrentMes); i++ {
			if room.CurrentMes[i].MessType == ACTION_EAT {
				eat_flag = true
				room.processEat(room.CurrentMes[i].SeatId, room.CurrentMes[i].Data)
				room.MayMes = room.MayMes[:0]
				room.CurrentMes = room.CurrentMes[:0]
				room.SyncRedis()
				return
			}
		}
	}

	if flag == false && eat_flag == false {
		if room.GameStatus == GAME_STATUS_GRAB_BAR {
			room.CreateBrightBarScore()
			room.MayMes = room.MayMes[:0]
			room.CurrentMes = room.CurrentMes[:0]
			room.MoPaiIndex = room.CurrentIndex
			room.Deal()
			return
		} else {
			//都是过，继续游戏
			logger.Debug("都是过--继续发牌")
			room.MayMes = room.MayMes[:0]
			room.CurrentMes = room.CurrentMes[:0]
			room.CurrentIndex = (room.ChuPaiIndex + 1) % 4
			room.MoPaiIndex = room.CurrentIndex
			room.Deal()
			return
		}
	}

	return
}
