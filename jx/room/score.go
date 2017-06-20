package room

import (
	"chess_alg_jx/jx/jx_nsq"
	"chess_alg_jx/logger"
	"chess_alg_jx/utils"
)

/*记分流水类型*/
const (
	SCORE_TYPE_BRIGHT_BAR    = iota + 1 //明杠 +1
	SCORE_TYPE_DARK_BAR                 //暗杠 +2
	SCORE_TYPE_UP_JING                  //上精  正精+2  副精+1
	SCORE_TYPE_SEARCH_BANKER            //抄庄 +5
	SCORE_TYPE_PING_HU                  //平胡
	SCORE_TYPE_SELFDRAWN_HU             //自摸
	SCORE_TYPE_SMILE                    //回头一笑
	SCORE_TYPE_LANDMINES                //埋地雷
	SCORE_TYPE_OPENING                  //开局上下翻
)

/*胡牌的奖励*/
const (
	HU_AWARD_TYPE_SELFDRAWN          = iota + 1 //自摸   *2
	HU_AWARD_TYPE_GRABBAR                       //抢杠   *2
	HU_AWARD_TYPE_BAR                           //杠开   *2
	HU_AWARD_TYPE_GOD                           //天胡   +20
	HU_AWARD_TYPE_LAND                          //地胡   +20
	HU_AWARD_TYPE_JING_DIAO                     //精钓   *2
	HU_AWARD_TYPE_LITTLE_SEVEN                  //小七对 *2
	HU_AWARD_TYPE_BIG_SEVEN                     //大七对 *2
	HU_AWARD_TYPE_THIRTEEN                      //十三烂 *2
	HU_AWARD_TYPE_SEVENSTAR_THIRTEEN            //七星十三烂 *4
	HU_AWARD_TYPE_DEGUO                         //德国  自摸时每家×2+5  否则点炮者×2
	HU_AWARD_TYPE_DEZHONGDE                     //德中德  自摸时每家×4+5  否则点炮者×4
)

/*玩家在一局游戏中的流水*/
type Score struct {
	UserName string `json:"user_name"` //产生此条流水的玩家名字
	Type     int    `json:"type"`      //流水类型
	Score    int    `json:"score"`     //分数
}

/*胡牌的奖励*/
type HuAward struct {
	Type int `json:"type"` //胡牌奖励类型
}

/*生成明杠的流水*/
func (game *Room) CreateBrightBarScore() {
	//更新玩家手牌中的加杠信息
	game.Player[game.CurrentIndex].HandCard.Bright = append(game.Player[game.CurrentIndex].HandCard.Bright, game.SharePai)
	//删除原来的碰牌
	for index, alt := range game.Player[game.CurrentIndex].HandCard.Alt {
		if alt == game.SharePai {
			game.Player[game.CurrentIndex].HandCard.Alt = append(game.Player[game.CurrentIndex].HandCard.Alt[:index],
				game.Player[game.CurrentIndex].HandCard.Alt[index+1:]...)
		}
	}

	i := game.CurrentIndex

	//生成开杠玩家的流水
	game.Player[i].Score += 3
	game.Player[i].MScore += 3
	game.Player[i].ScoreWater = append(game.Player[i].ScoreWater, Score{
		UserName: game.Player[i].UserName,
		Type:     SCORE_TYPE_BRIGHT_BAR,
		Score:    3,
	})
	for k := 0; k < 4; k++ {
		if err := game.GameNsqProducer(SEND_PLAYER_MONEY, k,
			[]int{i, 3}); err != nil {
			return
		}
	}
	//生成被开杠玩家的流水
	for j := 1; j < 4; j++ {
		game.Player[(i+j)%4].Score -= 1
		game.Player[(i+j)%4].MScore -= 1
		game.Player[(i+j)%4].ScoreWater = append(game.Player[(i+j)%4].ScoreWater, Score{
			UserName: game.Player[i].UserName,
			Type:     SCORE_TYPE_BRIGHT_BAR,
			Score:    -1,
		})
		for k := 0; k < 4; k++ {
			if err := game.GameNsqProducer(SEND_PLAYER_MONEY, k,
				[]int{(i + j) % 4, -1}); err != nil {
				return
			}
		}
		//添加流水数据库
		data := jx_nsq.Nsq_match_bill{
			McNo:      game.McNo,
			FMid:      game.Player[(i+j)%4].UserId,
			TMid:      game.Player[i].UserId,
			FNickname: game.Player[(i+j)%4].UserName,
			TNickname: game.Player[i].UserName,
			SType:     SCORE_TYPE_BRIGHT_BAR,
			WinType:   0,
			Amount:    1,
			CName:     "chess_alg_jx",
		}
		if err := jx_nsq.NsqMatchBillProducer(data); err != nil {
			logger.Error("玩家明杠，bill_to_api: ", err)
			return
		}
	}
}

/*生成暗杠的流水*/
func (game *Room) CreateDardBarScore() {
	i := game.CurrentIndex

	//生成开杠玩家的流水
	game.Player[i].Score += 6
	game.Player[i].MScore += 6
	game.Player[i].ScoreWater = append(game.Player[i].ScoreWater, Score{
		UserName: game.Player[i].UserName,
		Type:     SCORE_TYPE_DARK_BAR,
		Score:    6,
	})
	for k := 0; k < 4; k++ {
		if err := game.GameNsqProducer(SEND_PLAYER_MONEY, k,
			[]int{i, 6}); err != nil {
			return
		}
	}
	//生成被开杠玩家的流水
	for j := 1; j < 4; j++ {
		game.Player[(i+j)%4].Score -= 2
		game.Player[(i+j)%4].MScore -= 2
		game.Player[(i+j)%4].ScoreWater = append(game.Player[(i+j)%4].ScoreWater, Score{
			UserName: game.Player[i].UserName,
			Type:     SCORE_TYPE_DARK_BAR,
			Score:    -2,
		})
		for k := 0; k < 4; k++ {
			if err := game.GameNsqProducer(SEND_PLAYER_MONEY, k,
				[]int{(i + j) % 4, -2}); err != nil {
				return
			}
		}
		//添加流水数据库
		data := jx_nsq.Nsq_match_bill{
			McNo:      game.McNo,
			FMid:      game.Player[(i+j)%4].UserId,
			TMid:      game.Player[i].UserId,
			FNickname: game.Player[(i+j)%4].UserName,
			TNickname: game.Player[i].UserName,
			SType:     SCORE_TYPE_DARK_BAR,
			WinType:   0,
			Amount:    2,
			CName:     "chess_alg_jx",
		}
		if err := jx_nsq.NsqMatchBillProducer(data); err != nil {
			logger.Error("玩家暗杠，bill_to_api: ", err)
			return
		}
	}
}

/*检查玩家胡牌是否是杠开*/
func (room *Room) CheckIsBarWin() bool {
	if room.Player[room.CurrentIndex].WinData.Selfdrawn == true {
		lenth := len(room.Player[room.CurrentIndex].Opt)
		if room.Player[room.CurrentIndex].Opt[lenth-1] == PLAYING_STATUS_DARK_BAR ||
			room.Player[room.CurrentIndex].Opt[lenth-1] == PLAYING_STATUS_ADD_BAR ||
			room.Player[room.CurrentIndex].Opt[lenth-1] == PLAYING_STATUS_BRIGHT_BAR {
			return true
		}
	}

	return false
}

/*检查当前游戏结束时玩家胡牌是否是德中德*/
func (game *Room) isDeZhongDe() bool {
	zhengJing := game.UpUniverse
	fuJing := NextOrder(game.UpUniverse)

	for i := 0; i < 4; i++ {
		pai := game.Player[i].HandCard.ReverChange()
		z, _ := utils.Contain(zhengJing, pai)
		f, _ := utils.Contain(fuJing, pai)
		if z == true || f == true {
			return false
		}
	}
	return true
}

/*玩家胡牌生成玩家的流水, 胡牌玩家为当前控牌玩家*/
func (game *Room) CreatHuScore(seatId int) {
	i := seatId //当前控牌玩家
	score := 4
	//检查是否是抢杠
	if game.GameStatus == GAME_STATUS_GRAB_BAR {
		score = score * 2
		game.Player[i].WinData.Award = append(game.Player[i].WinData.Award, HU_AWARD_TYPE_GRABBAR)
		data := jx_nsq.Nsq_match_turn{
			McNo:     game.McNo,
			MId:      game.Player[i].UserId,
			TurnType: HU_AWARD_TYPE_GRABBAR,
			CName:    "chess_alg_jx",
		}
		if err := jx_nsq.NsqMatchTurnProducer(data); err != nil {
			logger.Error("玩家胡牌奖励--抢杠： ", err)
			return
		}
	}
	//检查是否时杠开
	if game.CheckIsBarWin() == true {
		score = score * 2
		game.Player[i].WinData.Award = append(game.Player[i].WinData.Award, HU_AWARD_TYPE_BAR)
		data := jx_nsq.Nsq_match_turn{
			McNo:     game.McNo,
			MId:      game.Player[i].UserId,
			TurnType: HU_AWARD_TYPE_BAR,
			CName:    "chess_alg_jx",
		}
		if err := jx_nsq.NsqMatchTurnProducer(data); err != nil {
			logger.Error("玩家胡牌奖励--杠开： ", err)
			return
		}
	}
	//是否是小七对
	if game.Player[i].HandCard.littleSevenPairs(game.UpUniverse) == true {
		score = score * 2
		game.Player[i].WinData.Award = append(game.Player[i].WinData.Award, HU_AWARD_TYPE_LITTLE_SEVEN)
		data := jx_nsq.Nsq_match_turn{
			McNo:     game.McNo,
			MId:      game.Player[i].UserId,
			TurnType: HU_AWARD_TYPE_LITTLE_SEVEN,
			CName:    "chess_alg_jx",
		}
		if err := jx_nsq.NsqMatchTurnProducer(data); err != nil {
			logger.Error("玩家胡牌奖励--小七对： ", err)
			return
		}
	}
	//是否是大七对
	if game.Player[i].HandCard.bigSevenPairs(game.UpUniverse) == true {
		score = score * 2
		game.Player[i].WinData.Award = append(game.Player[i].WinData.Award, HU_AWARD_TYPE_BIG_SEVEN)
		data := jx_nsq.Nsq_match_turn{
			McNo:     game.McNo,
			MId:      game.Player[i].UserId,
			TurnType: HU_AWARD_TYPE_BIG_SEVEN,
			CName:    "chess_alg_jx",
		}
		if err := jx_nsq.NsqMatchTurnProducer(data); err != nil {
			logger.Error("玩家胡牌奖励--大七对： ", err)
			return
		}
	}
	//是否是十三烂
	if game.Player[i].HandCard.isThirteen() == true {
		score = score * 2
		game.Player[i].WinData.Award = append(game.Player[i].WinData.Award, HU_AWARD_TYPE_THIRTEEN)
		data := jx_nsq.Nsq_match_turn{
			McNo:     game.McNo,
			MId:      game.Player[i].UserId,
			TurnType: HU_AWARD_TYPE_THIRTEEN,
			CName:    "chess_alg_jx",
		}
		if err := jx_nsq.NsqMatchTurnProducer(data); err != nil {
			logger.Error("玩家胡牌奖励--大七对： ", err)
			return
		}
	}
	//是否是七星十三烂
	if game.Player[i].HandCard.isSevenThirteen() == true {
		score = score * 4
		game.Player[i].WinData.Award = append(game.Player[i].WinData.Award, HU_AWARD_TYPE_SEVENSTAR_THIRTEEN)
		data := jx_nsq.Nsq_match_turn{
			McNo:     game.McNo,
			MId:      game.Player[i].UserId,
			TurnType: HU_AWARD_TYPE_SEVENSTAR_THIRTEEN,
			CName:    "chess_alg_jx",
		}
		if err := jx_nsq.NsqMatchTurnProducer(data); err != nil {
			logger.Error("玩家胡牌奖励--大七对： ", err)
			return
		}
	}

	//有抢杠/杠开	三家支付
	if game.Player[i].WinData.Selfdrawn == true || game.CheckIsBarWin() == true || game.GameStatus == GAME_STATUS_GRAB_BAR {
		//是否是精钓
		if game.Player[i].HandCard.isUniverseDiao(game.UpUniverse) == true {
			score = score * 2
			game.Player[i].WinData.Award = append(game.Player[i].WinData.Award, HU_AWARD_TYPE_JING_DIAO)
			data := jx_nsq.Nsq_match_turn{
				McNo:     game.McNo,
				MId:      game.Player[i].UserId,
				TurnType: HU_AWARD_TYPE_JING_DIAO,
				CName:    "chess_alg_jx",
			}
			if err := jx_nsq.NsqMatchTurnProducer(data); err != nil {
				logger.Error("玩家胡牌奖励--精钓： ", err)
				return
			}
		}

		if game.Player[i].WinData.Selfdrawn == true {
			//生成自摸的奖励
			game.Player[i].WinData.Award = append(game.Player[i].WinData.Award, HU_AWARD_TYPE_SELFDRAWN)
			//插入数据库
			data := jx_nsq.Nsq_match_turn{
				McNo:     game.McNo,
				MId:      game.Player[i].UserId,
				TurnType: HU_AWARD_TYPE_SELFDRAWN,
				CName:    "chess_alg_jx",
			}
			if err := jx_nsq.NsqMatchTurnProducer(data); err != nil {
				logger.Error("玩家胡牌加番--杠上炮： ", err)
				return
			}
		}

		if game.Player[i].HandCard.IsDeGuo() == true {
			//检查是否是德国
			score = score*2 + 5
			game.Player[i].WinData.Award = append(game.Player[i].WinData.Award, HU_AWARD_TYPE_DEGUO)
			data := jx_nsq.Nsq_match_turn{
				McNo:     game.McNo,
				MId:      game.Player[i].UserId,
				TurnType: HU_AWARD_TYPE_DEGUO,
				CName:    "chess_alg_jx",
			}
			if err := jx_nsq.NsqMatchTurnProducer(data); err != nil {
				logger.Error("玩家胡牌奖励--德国： ", err)
				return
			}
		} else if game.isDeZhongDe() == true {
			//检查是否是德中德
			score = score*4 + 5
			game.Player[i].WinData.Award = append(game.Player[i].WinData.Award, HU_AWARD_TYPE_DEZHONGDE)
			data := jx_nsq.Nsq_match_turn{
				McNo:     game.McNo,
				MId:      game.Player[i].UserId,
				TurnType: HU_AWARD_TYPE_DEZHONGDE,
				CName:    "chess_alg_jx",
			}
			if err := jx_nsq.NsqMatchTurnProducer(data); err != nil {
				logger.Error("玩家胡牌奖励--德国： ", err)
				return
			}
		} else {
			score = score * 2
		}

		game.Player[i].ScoreWater = append(game.Player[i].ScoreWater, Score{
			UserName: game.Player[i].UserName,
			Type:     SCORE_TYPE_SELFDRAWN_HU,
			Score:    score * 3,
		})
		game.Player[i].Score += score * 3
		game.Player[i].MScore += score * 3
		for j := 0; j < 4; j++ {
			if err := game.GameNsqProducer(SEND_PLAYER_MONEY, j,
				[]int{i, score * 3}); err != nil {
				return
			}
		}
		//生成其他三个玩家中输钱的流水
		for j := 1; j < 4; j++ {
			//生成此玩家输钱的流水
			game.Player[(i+j)%4].Score -= score
			game.Player[(i+j)%4].MScore -= score
			game.Player[(i+j)%4].ScoreWater = append(game.Player[(i+j)%4].ScoreWater, Score{
				UserName: game.Player[i].UserName,
				Type:     SCORE_TYPE_SELFDRAWN_HU,
				Score:    -score,
			})
			//修改玩家桌面上的钱
			for k := 0; k < 4; k++ {
				if err := game.GameNsqProducer(SEND_PLAYER_MONEY, k,
					[]int{(i + j) % 4, 0 - score}); err != nil {
					return
				}
			}
			//添加流水数据库
			data := jx_nsq.Nsq_match_bill{
				McNo:      game.McNo,
				FMid:      game.Player[(i+j)%4].UserId,
				TMid:      game.Player[i].UserId,
				FNickname: game.Player[(i+j)%4].UserName,
				TNickname: game.Player[i].UserName,
				SType:     SCORE_TYPE_SELFDRAWN_HU,
				WinType:   0,
				Amount:    score,
				CName:     "chess_alg_jx",
			}
			if err := jx_nsq.NsqMatchBillProducer(data); err != nil {
				logger.Error("玩家自摸，bill_to_api: ", err)
				return
			}
		}
	} else {
		//不是自摸，考虑点炮是一家付钱还是三家付钱
		win_score := 0
		dian_score := 0
		else_score := 0
		if game.Player[i].HandCard.IsDeGuo() == true {
			dian_score = score*2 + 5
			else_score = score * 2
			game.Player[i].WinData.Award = append(game.Player[i].WinData.Award, HU_AWARD_TYPE_DEGUO)
			data := jx_nsq.Nsq_match_turn{
				McNo:     game.McNo,
				MId:      game.Player[i].UserId,
				TurnType: HU_AWARD_TYPE_DEGUO,
				CName:    "chess_alg_jx",
			}
			if err := jx_nsq.NsqMatchTurnProducer(data); err != nil {
				logger.Error("玩家胡牌奖励--德国： ", err)
				return
			}
		} else if game.isDeZhongDe() == true {
			dian_score = score*4 + 5
			else_score = score * 4
			game.Player[i].WinData.Award = append(game.Player[i].WinData.Award, HU_AWARD_TYPE_DEZHONGDE)
			data := jx_nsq.Nsq_match_turn{
				McNo:     game.McNo,
				MId:      game.Player[i].UserId,
				TurnType: HU_AWARD_TYPE_DEZHONGDE,
				CName:    "chess_alg_jx",
			}
			if err := jx_nsq.NsqMatchTurnProducer(data); err != nil {
				logger.Error("玩家胡牌奖励--德国： ", err)
				return
			}
		} else {
			dian_score = score
			else_score = score
		}

		if game.RoomSet.Option == true {
			win_score = dian_score + else_score*2
		} else {
			win_score = dian_score
		}

		//胡牌玩家
		game.Player[i].Score += win_score
		game.Player[i].MScore += win_score
		game.Player[i].ScoreWater = append(game.Player[i].ScoreWater, Score{
			UserName: game.Player[i].UserName,
			Type:     SCORE_TYPE_PING_HU,
			Score:    win_score,
		})
		//修改赢钱玩家桌面上的钱
		for k := 0; k < 4; k++ {
			if err := game.GameNsqProducer(SEND_PLAYER_MONEY, k,
				[]int{i, win_score}); err != nil {
				return
			}
		}

		for j := 1; j < 4; j++ {
			if (i+j)%4 == game.ChuPaiIndex {
				game.Player[game.ChuPaiIndex].Score -= dian_score
				game.Player[game.ChuPaiIndex].MScore -= dian_score
				game.Player[game.ChuPaiIndex].ScoreWater = append(game.Player[game.ChuPaiIndex].ScoreWater, Score{
					UserName: game.Player[i].UserName,
					Type:     SCORE_TYPE_PING_HU,
					Score:    0 - dian_score,
				})
				//修改玩家桌面上的钱
				for k := 0; k < 4; k++ {
					if err := game.GameNsqProducer(SEND_PLAYER_MONEY, k,
						[]int{game.ChuPaiIndex, 0 - dian_score}); err != nil {
						return
					}
				}
				//添加流水数据库
				data := jx_nsq.Nsq_match_bill{
					McNo:      game.McNo,
					FMid:      game.Player[game.ChuPaiIndex].UserId,
					TMid:      game.Player[i].UserId,
					FNickname: game.Player[game.ChuPaiIndex].UserName,
					TNickname: game.Player[i].UserName,
					SType:     SCORE_TYPE_PING_HU,
					WinType:   0,
					Amount:    dian_score,
					CName:     "chess_alg_jx",
				}
				if err := jx_nsq.NsqMatchBillProducer(data); err != nil {
					logger.Error("玩家平胡，bill_to_api: ", err)
					return
				}
			} else {
				if game.RoomSet.Option == true {
					//生成此玩家输钱的流水
					game.Player[(i+j)%4].Score -= else_score
					game.Player[(i+j)%4].MScore -= else_score
					game.Player[(i+j)%4].ScoreWater = append(game.Player[(i+j)%4].ScoreWater, Score{
						UserName: game.Player[i].UserName,
						Type:     SCORE_TYPE_PING_HU,
						Score:    -else_score,
					})
					//修改玩家桌面上的钱
					for k := 0; k < 4; k++ {
						if err := game.GameNsqProducer(SEND_PLAYER_MONEY, k,
							[]int{(i + j) % 4, 0 - else_score}); err != nil {
							return
						}
					}
					//添加流水数据库
					data := jx_nsq.Nsq_match_bill{
						McNo:      game.McNo,
						FMid:      game.Player[(i+j)%4].UserId,
						TMid:      game.Player[i].UserId,
						FNickname: game.Player[(i+j)%4].UserName,
						TNickname: game.Player[i].UserName,
						SType:     SCORE_TYPE_PING_HU,
						WinType:   0,
						Amount:    else_score,
						CName:     "chess_alg_jx",
					}
					if err := jx_nsq.NsqMatchBillProducer(data); err != nil {
						logger.Error("玩家点炮胡，bill_to_api: ", err)
						return
					}
				}
			}
		}
	}
	game.SyncRedis()
	return
}

/*生成天胡地胡的流水*/
func (game *Room) CreateGodLandHuScore(god bool) {
	award_type := 0
	score_type := 0
	if god == true {
		award_type = HU_AWARD_TYPE_GOD
		score_type = SCORE_TYPE_PING_HU
	} else {
		award_type = HU_AWARD_TYPE_LAND
		score_type = SCORE_TYPE_SELFDRAWN_HU
	}

	i := game.CurrentIndex //当前控牌玩家
	score := 20
	//如果是精钓则score = 40
	if game.Player[i].HandCard.isUniverseDiao(game.UpUniverse) == true {
		score = 40
	}
	//生成胡牌奖励
	game.Player[i].WinData.Award = append(game.Player[i].WinData.Award, award_type)
	data := jx_nsq.Nsq_match_turn{
		McNo:     game.McNo,
		MId:      game.Player[i].UserId,
		TurnType: award_type,
		CName:    "chess_alg_jx",
	}
	if err := jx_nsq.NsqMatchTurnProducer(data); err != nil {
		logger.Error("玩家胡牌奖励--天地胡： ", err)
		return
	}

	//生成流水
	game.Player[i].Score += score * 3
	game.Player[i].MScore += score * 3
	game.Player[i].ScoreWater = append(game.Player[i].ScoreWater, Score{
		UserName: game.Player[i].UserName,
		Type:     score_type,
		Score:    score * 3,
	})
	for j := 0; j < 4; j++ {
		if err := game.GameNsqProducer(SEND_PLAYER_MONEY, j,
			[]int{i, score * 3}); err != nil {
			return
		}
	}

	//生成其他三个玩家中输钱的流水
	for j := 1; j < 4; j++ {
		//生成此玩家输钱的流水
		game.Player[(i+j)%4].Score -= score
		game.Player[(i+j)%4].MScore -= score
		game.Player[(i+j)%4].ScoreWater = append(game.Player[(i+j)%4].ScoreWater, Score{
			UserName: game.Player[i].UserName,
			Type:     score_type,
			Score:    -score,
		})
		//修改玩家桌面上的钱
		for k := 0; k < 4; k++ {
			if err := game.GameNsqProducer(SEND_PLAYER_MONEY, k,
				[]int{(i + j) % 4, 0 - score}); err != nil {
				return
			}
		}
		//添加流水数据库
		data := jx_nsq.Nsq_match_bill{
			McNo:      game.McNo,
			FMid:      game.Player[(i+j)%4].UserId,
			TMid:      game.Player[i].UserId,
			FNickname: game.Player[(i+j)%4].UserName,
			TNickname: game.Player[i].UserName,
			SType:     score_type,
			WinType:   0,
			Amount:    score,
			CName:     "chess_alg_jx",
		}
		if err := jx_nsq.NsqMatchBillProducer(data); err != nil {
			logger.Error("玩家天地胡，bill_to_api: ", err)
			return
		}
	}
	game.SyncRedis()
}

/*生成抄庄的流水*/
func (game *Room) CreateSearchBankerScore() {
	i := game.BankerId

	game.Player[i].Score -= 30
	game.Player[i].MScore -= 30
	game.Player[i].ScoreWater = append(game.Player[i].ScoreWater, Score{
		UserName: game.Player[i].UserName,
		Type:     SCORE_TYPE_SEARCH_BANKER,
		Score:    -30,
	})

	for j := 1; j < 4; j++ {
		game.Player[(i+j)%4].Score += 10
		game.Player[(i+j)%4].MScore += 10
		game.Player[(i+j)%4].ScoreWater = append(game.Player[(i+j)%4].ScoreWater, Score{
			UserName: game.Player[i].UserName,
			Type:     SCORE_TYPE_SEARCH_BANKER,
			Score:    10,
		})
		//修改玩家桌面上的钱
		for k := 0; k < 4; k++ {
			if err := game.GameNsqProducer(SEND_PLAYER_MONEY, k,
				[]int{(i + j) % 4, 10}); err != nil {
				return
			}
		}
		//添加流水数据库
		data := jx_nsq.Nsq_match_bill{
			McNo:      game.McNo,
			FMid:      game.Player[i].UserId,
			TMid:      game.Player[(i+j)%4].UserId,
			FNickname: game.Player[i].UserName,
			TNickname: game.Player[(i+j)%4].UserName,
			SType:     SCORE_TYPE_SEARCH_BANKER,
			WinType:   0,
			Amount:    10,
			CName:     "chess_alg_jx",
		}
		if err := jx_nsq.NsqMatchBillProducer(data); err != nil {
			logger.Error("抄庄，bill_to_api: ", err)
			return
		}
	}
	game.SyncRedis()
}
