package room

import (
	"chess_alg_jx/jx/jx_nsq"
	"chess_alg_jx/logger"
	"chess_alg_jx/utils"
	"time"
)

const (
	UNIVERSE_TYPE_UP       = 1 //上精
	UNIVERSE_TYPE_SMILE    = 2 //回头一笑
	UNIVERSE_TYPE_LAND     = 3 //埋地雷
	UNIVERSE_TYPE_OPENGING = 4 //开局上下翻
)

/*计算每个玩家的该精牌分数, 区分是只算手牌还是所有的牌*/
func (game *Room) CountUniverseNum(seatId int, universe int, all bool) int {
	//手牌中精牌的个数
	num := game.Player[seatId].HandCard.Pai[universe/10][universe%10]
	if all == true {
		//计算吃、碰、杠、胡、手持的、打出的
		for _, data := range game.Player[seatId].HandCard.Eat {
			if have, _ := utils.Contain(universe, data); have == true {
				num++
			}
		}
		if have, _ := utils.Contain(universe, game.Player[seatId].HandCard.Alt); have == true {
			num += 3
		}
		if have, _ := utils.Contain(universe, game.Player[seatId].HandCard.Bright); have == true {
			num += 4
		}
		if have, _ := utils.Contain(universe, game.Player[seatId].HandCard.Dark); have == true {
			num += 4
		}
		//考虑精牌打出是否还算分
		if game.RoomSet.Universe == true {
			for _, v := range game.Player[seatId].Played {
				if v == universe {
					num++
				}
			}
		}
	}

	return num
}

/*玩家是否有杠精*/
func (game *Room) IsBarJIng(seatId int, universe int) bool {
	if have, _ := utils.Contain(universe, game.Player[seatId].HandCard.Bright); have == true {
		return true
	}
	if have, _ := utils.Contain(universe, game.Player[seatId].HandCard.Dark); have == true {
		return true
	}
	if have, _ := utils.Contain(NextOrder(universe), game.Player[seatId].HandCard.Bright); have == true {
		return true
	}
	if have, _ := utils.Contain(NextOrder(universe), game.Player[seatId].HandCard.Dark); have == true {
		return true
	}

	return false
}

/*检测玩家是否是霸王霸王精*/
func (game *Room) IsArchload(seatId int, universe int) bool {
	//检测其他三个玩家是否有精牌
	for j := 1; j < 4; j++ {
		pai := game.Player[(seatId+j)%4].HandCard.ReverChange()
		for _, v := range pai {
			if v == universe || v == NextOrder(universe) {
				return false
			}
		}
	}
	return true
}

/*计算精牌的流水: 每局游戏结束时，单独拿出来上精算分, 考虑是否时霸王精或者可以冲关*/
func (game *Room) CreateUniverseScore() {
	manager := game.UpUniverse
	deputy := NextOrder(manager)

	//1：告诉玩家所有玩家手牌中精牌的个数
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			num1 := game.CountUniverseNum(j, manager, game.RoomSet.Universe)
			num2 := game.CountUniverseNum(j, deputy, game.RoomSet.Universe)
			if err := game.GameNsqProducer(SEND_UNIVERSE_NUM, i, []int{UNIVERSE_TYPE_UP, j, manager, num1, deputy, num2}); err != nil {
				return
			}
		}
	}
	time.Sleep(2 * time.Second)
	//2：计算玩家精牌得分并插入流水数据库
	for i := 0; i < 4; i++ {
		score := 0
		for j := 0; j < 4; j++ {
			universe := 0
			if game.Player[j].HandCard.Pai[manager/10][manager%10] != 0 || game.Player[j].HandCard.Pai[deputy/10][deputy%10] != 0 {
				universe = game.Player[j].HandCard.Pai[manager/10][manager%10]*2 + game.Player[j].HandCard.Pai[deputy/10][deputy%10]
				//检测是否可以冲关
				if universe >= 5 {
					universe = universe * (universe - 3)
				}
				//检测是否是霸王
				if is := game.IsArchload(j, manager); is == true {
					if game.RoomSet.Snake == true {
						universe = universe * 2
					} else {
						universe += 10
					}
				}
				//是否含有杠精
				if is := game.IsBarJIng(j, manager); is == true {
					universe += 10
				}

				if i == j {
					score += universe * 3

				} else {
					score -= universe
					data := jx_nsq.Nsq_match_bill{
						McNo:      game.McNo,
						FMid:      game.Player[i].UserId,
						TMid:      game.Player[j].UserId,
						FNickname: game.Player[i].UserName,
						TNickname: game.Player[j].UserName,
						SType:     SCORE_TYPE_UP_JING,
						WinType:   0,
						Amount:    universe,
						CName:     "chess_alg_jx",
					}
					if err := jx_nsq.NsqMatchBillProducer(data); err != nil {
						logger.Error("玩家埋地雷，bill_to_api: ", err)
						return
					}
				}
			}
		}
		game.Player[i].ScoreWater = append(game.Player[i].ScoreWater, Score{
			UserName: game.Player[i].UserName,
			Type:     SCORE_TYPE_UP_JING,
			Score:    score,
		})
		game.Player[i].Score += score
		game.Player[i].MScore += score
		if score != 0 {
			for k := 0; k < 4; k++ {
				if err := game.GameNsqProducer(SEND_PLAYER_MONEY, k, []int{i, score}); err != nil {
					return
				}
			}
		}
	}
}

/*生成回头一笑的流水, 不考虑冲关和霸王*/
func (game *Room) CreateSmileScore() {
	manager := game.UpUniverse
	deputy := NextOrder(manager)

	//1：告诉玩家所有玩家手牌中精牌的个数
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			num1 := game.Player[j].HandCard.Pai[manager/10][manager%10]
			num2 := game.Player[j].HandCard.Pai[deputy/10][deputy%10]
			if err := game.GameNsqProducer(SEND_UNIVERSE_NUM, i, []int{UNIVERSE_TYPE_SMILE, j, manager, num1, deputy, num2}); err != nil {
				return
			}
		}
	}
	time.Sleep(2 * time.Second)
	//2：计算玩家精牌得分并插入流水数据库
	for i := 0; i < 4; i++ {
		score := 0
		for j := 0; j < 4; j++ {
			universe := 0
			if game.Player[j].HandCard.Pai[manager/10][manager%10] != 0 || game.Player[j].HandCard.Pai[deputy/10][deputy%10] != 0 {
				universe = game.Player[j].HandCard.Pai[manager/10][manager%10]*2 + game.Player[j].HandCard.Pai[deputy/10][deputy%10]
				//检测是否可以冲关
				if universe >= 5 {
					universe = universe * (universe - 3)
				}
				//检测是否是霸王
				if is := game.IsArchload(j, game.UpUniverse); is == true {
					if game.RoomSet.Snake == true {
						universe = universe * 2
					} else {
						universe = universe + 10
					}
				}

				if i == j {
					score += universe * 3

				} else {
					score -= universe
					data := jx_nsq.Nsq_match_bill{
						McNo:      game.McNo,
						FMid:      game.Player[i].UserId,
						TMid:      game.Player[j].UserId,
						FNickname: game.Player[i].UserName,
						TNickname: game.Player[j].UserName,
						SType:     SCORE_TYPE_SMILE,
						WinType:   0,
						Amount:    universe,
						CName:     "chess_alg_jx",
					}
					if err := jx_nsq.NsqMatchBillProducer(data); err != nil {
						logger.Error("玩家回头一笑，bill_to_api: ", err)
						return
					}
				}
			}
		}
		game.Player[i].ScoreWater = append(game.Player[i].ScoreWater, Score{
			UserName: game.Player[i].UserName,
			Type:     SCORE_TYPE_SMILE,
			Score:    score,
		})
		game.Player[i].Score += score
		game.Player[i].MScore += score
		if score != 0 {
			for k := 0; k < 4; k++ {
				if err := game.GameNsqProducer(SEND_PLAYER_MONEY, k, []int{i, score}); err != nil {
					return
				}
			}
		}
	}
}

/*生成埋地雷的流水, 游戏结束的时候用下精翻精计算，也包括冲关和霸王*/
func (game *Room) CreateLandMinesScore() {
	manager := game.DownUniverse
	deputy := NextOrder(game.DownUniverse)

	//1：告诉玩家所有玩家手牌中精牌的个数
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			num1 := game.CountUniverseNum(j, manager, game.RoomSet.Universe)
			num2 := game.CountUniverseNum(j, deputy, game.RoomSet.Universe)
			if err := game.GameNsqProducer(SEND_UNIVERSE_NUM, i, []int{UNIVERSE_TYPE_LAND, j, manager, num1, deputy, num2}); err != nil {
				return
			}
		}
	}
	time.Sleep(2 * time.Second)
	//2：计算玩家精牌得分并插入流水数据库
	for i := 0; i < 4; i++ {
		score := 0
		for j := 0; j < 4; j++ {
			universe := 0
			if game.Player[j].HandCard.Pai[manager/10][manager%10] != 0 || game.Player[j].HandCard.Pai[deputy/10][deputy%10] != 0 {
				universe = game.Player[j].HandCard.Pai[manager/10][manager%10]*2 + game.Player[j].HandCard.Pai[deputy/10][deputy%10]
				//检测是否可以冲关
				if universe >= 5 {
					universe = universe * (universe - 3)
				}
				//检测是否是霸王
				if is := game.IsArchload(j, manager); is == true {
					if game.RoomSet.Snake == true {
						universe = universe * 2
					} else {
						universe += 10
					}
				}
				//是否含有杠精
				if is := game.IsBarJIng(j, manager); is == true {
					universe += 10
				}

				if i == j {
					score += universe * 3

				} else {
					score -= universe
					data := jx_nsq.Nsq_match_bill{
						McNo:      game.McNo,
						FMid:      game.Player[i].UserId,
						TMid:      game.Player[j].UserId,
						FNickname: game.Player[i].UserName,
						TNickname: game.Player[j].UserName,
						SType:     SCORE_TYPE_LANDMINES,
						WinType:   0,
						Amount:    universe,
						CName:     "chess_alg_jx",
					}
					if err := jx_nsq.NsqMatchBillProducer(data); err != nil {
						logger.Error("玩家埋地雷，bill_to_api: ", err)
						return
					}
				}
			}
		}
		game.Player[i].ScoreWater = append(game.Player[i].ScoreWater, Score{
			UserName: game.Player[i].UserName,
			Type:     SCORE_TYPE_LANDMINES,
			Score:    score,
		})
		game.Player[i].Score += score
		game.Player[i].MScore += score
		if score != 0 {
			for k := 0; k < 4; k++ {
				if err := game.GameNsqProducer(SEND_PLAYER_MONEY, k, []int{i, score}); err != nil {
					return
				}
			}
		}
	}
}

/*生成开局上下翻精--特殊玩法的流水, 游戏结束的时候用下精翻精计算，也包括冲关和霸王*/
func (game *Room) CreateOpeningScore() {
	manager := game.DownUniverse
	deputy := NextOrder(game.DownUniverse)

	//1：告诉玩家所有玩家手牌中精牌的个数
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			num1 := game.Player[j].HandCard.Pai[manager/10][manager%10]
			num2 := game.Player[j].HandCard.Pai[deputy/10][deputy%10]
			if err := game.GameNsqProducer(SEND_UNIVERSE_NUM, i, []int{UNIVERSE_TYPE_OPENGING, j, manager, num1, deputy, num2}); err != nil {
				return
			}
		}
	}
	time.Sleep(2 * time.Second)
	//2：计算玩家精牌得分并插入流水数据库
	for i := 0; i < 4; i++ {
		score := 0
		for j := 0; j < 4; j++ {
			universe := 0
			if game.Player[j].HandCard.Pai[manager/10][manager%10] != 0 || game.Player[j].HandCard.Pai[deputy/10][deputy%10] != 0 {
				universe = game.Player[j].HandCard.Pai[manager/10][manager%10]*2 + game.Player[j].HandCard.Pai[deputy/10][deputy%10]
				//检测是否可以冲关
				if universe >= 5 {
					universe = universe * (universe - 3)
				}
				//检测是否是霸王
				if is := game.IsArchload(j, manager); is == true {
					if game.RoomSet.Snake == true {
						universe = universe * 2
					} else {
						universe += 10
					}
				}
				//是否含有杠精
				if is := game.IsBarJIng(j, manager); is == true {
					universe += 10
				}

				if i == j {
					score += universe * 3

				} else {
					score -= universe
					data := jx_nsq.Nsq_match_bill{
						McNo:      game.McNo,
						FMid:      game.Player[i].UserId,
						TMid:      game.Player[j].UserId,
						FNickname: game.Player[i].UserName,
						TNickname: game.Player[j].UserName,
						SType:     SCORE_TYPE_OPENING,
						WinType:   0,
						Amount:    universe,
						CName:     "chess_alg_jx",
					}
					if err := jx_nsq.NsqMatchBillProducer(data); err != nil {
						logger.Error("玩家开局上下翻，bill_to_api: ", err)
						return
					}
				}
			}
		}
		game.Player[i].ScoreWater = append(game.Player[i].ScoreWater, Score{
			UserName: game.Player[i].UserName,
			Type:     SCORE_TYPE_OPENING,
			Score:    score,
		})
		game.Player[i].Score += score
		game.Player[i].MScore += score
		if score != 0 {
			for k := 0; k < 4; k++ {
				if err := game.GameNsqProducer(SEND_PLAYER_MONEY, k, []int{i, score}); err != nil {
					return
				}
			}
		}
	}
}
