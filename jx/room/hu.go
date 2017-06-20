package room

import (
	"chess_alg_jx/logger"
	"fmt"
)

//胡牌的牌型
const (
	PING_HU      = 1
	LITTLE_SEVEN = 2
	BIG_SEVEEN   = 3
	DEGUO        = 4
	DEZHONGDE    = 5
)

/*分解手牌中所有的万/筒/条/风/箭      为刻/顺的组合*/
func Analyze(pai []int) bool {
	if pai[0] == 0 {
		return true
	}
	//寻找第一张牌
	index := 0
	for i := 1; i < 10; i++ {
		if pai[i] != 0 {
			index = i
			break
		}
	}

	result := false
	if pai[index] >= 3 { //刻牌
		//除去这三张刻牌
		pai[index] -= 3
		pai[0] -= 3
		result = Analyze(pai)
		//还原这三张牌
		pai[index] += 3
		pai[0] += 3

		return result
	}

	//顺牌
	if index < 8 && pai[index+1] > 0 && pai[index+2] > 0 {
		//除去这三张顺牌
		pai[index] -= 1
		pai[index+1] -= 1
		pai[index+2] -= 1
		pai[0] -= 3
		result = Analyze(pai)
		//还原这三张牌
		pai[index] += 1
		pai[index+1] += 1
		pai[index+2] += 1
		pai[0] += 3

		return result
	}

	return false
}

//对字牌特别判断
func AnalyzeZi(pai []int) bool {
	if pai[0] == 0 {
		return true
	}

	//寻找第一张牌
	index := 0
	for i := 1; i < 5; i++ {
		if pai[i] != 0 {
			index = i
			break
		}
	}
	result := false
	if pai[index] >= 3 { //刻牌
		//除去这三张刻牌
		pai[index] -= 3
		pai[0] -= 3
		result = AnalyzeZi(pai)
		//还原这三张牌
		pai[index] += 3
		pai[0] += 3

		return result
	}

	//任何不重复的三张字牌都可以组成顺牌
	if index < 3 {
		if pai[index+1] > 0 && pai[index+2] > 0 {
			//除去这三张顺牌
			pai[index] -= 1
			pai[index+1] -= 1
			pai[index+2] -= 1
			pai[0] -= 3
			result = AnalyzeZi(pai)
			//还原这三张牌
			pai[index] += 1
			pai[index+1] += 1
			pai[index+2] += 1
			pai[0] += 3
			return result
		}
		if pai[index+1] > 0 && pai[index+3] > 0 {
			//除去这三张顺牌
			pai[index] -= 1
			pai[index+1] -= 1
			pai[index+3] -= 1
			pai[0] -= 3
			result = AnalyzeZi(pai)
			//还原这三张牌
			pai[index] += 1
			pai[index+1] += 1
			pai[index+3] += 1
			pai[0] += 3
			return result
		}
		if pai[index+2] > 0 && pai[index+3] > 0 {
			//除去这三张顺牌
			pai[index] -= 1
			pai[index+2] -= 1
			pai[index+3] -= 1
			pai[0] -= 3
			result = AnalyzeZi(pai)
			//还原这三张牌
			pai[index] += 1
			pai[index+2] += 1
			pai[index+3] += 1
			pai[0] += 3
			return result
		}
	}

	return false
}

/*
	判断当前牌型是否是基本胡牌牌型 3333 2 牌型
	判断是否只有2张牌
*/
func (hand Hand) baseHuMode() bool {
	jangIndex := 0 //将牌的位置
	yushu := 0     //余数
	jangExisted := false

	//首先判断牌型是否满足3 3 3 2 的模型
	for i := 0; i < 5; i++ {
		yushu = hand.Pai[i][0] % 3
		if yushu == 1 {
			return false
		} else if yushu == 2 {
			if jangExisted {
				return false
			}
			jangIndex = i
			jangExisted = true
		}
	}

	//对不含将牌的万/筒/条/分别分析，看是否可以分解为刻/顺的组合
	for i := 0; i < 5; i++ {
		if i != jangIndex && hand.Pai[i][0] != 0 {
			if i == 3 {
				//对字牌特殊判断 分解成顺子/刻子
				if AnalyzeZi(hand.Pai[i]) == false {
					return false
				}
			} else {
				if Analyze(hand.Pai[i]) == false {
					return false
				}
			}
		}
	}

	//对含将牌的行进行判断
	success := false
	for i := 0; i < 10; i++ {
		if hand.Pai[jangIndex][i] >= 2 {
			//除去这两张牌
			hand.Pai[jangIndex][i] -= 2
			hand.Pai[jangIndex][0] -= 2
			if Analyze(hand.Pai[jangIndex]) {
				success = true
			}
			//还原这两张牌
			hand.Pai[jangIndex][i] += 2
			hand.Pai[jangIndex][0] += 2
			if success {
				break
			}
		}
	}

	return success
}

/*
	判断当前牌是否是小七对
		玩家手中所有的牌的个数必须是偶数 而且没有碰过和杠过
*/
func (hand Hand) littleSevenPairs(universe int) bool {
	//首先检查是否有碰牌或者杠牌
	if len(hand.Alt) != 0 || len(hand.Eat) != 0 || len(hand.Bright) != 0 || len(hand.Dark) != 0 {
		return false
	}
	for i := 0; i < 5; i++ {
		for j := 1; j < 10; j++ {
			if (hand.Pai[i][j] & 1) != 0 {
				//牌的个数是奇数
				return false
			}
		}
	}
	for i := 0; i < 5; i++ {
		for j := 1; j < 10; j++ {
			if hand.Pai[i][j] > 0 {
				if i*10+j == universe || i*10+j == NextOrder(universe) {
					hand.Pai[i][0]--
					hand.Pai[i][j]--
					for _, v := range MajBasePai {
						if v == universe || v == NextOrder(universe) {
							continue
						}
						hand.Pai[v/10][0] += 1
						hand.Pai[v/10][v%10] += 1
						if is := hand.littleSevenPairs(universe); is == true {
							logger.Debug("玩家可以胡牌---小七对")
							hand.Pai[v/10][0] -= 1
							hand.Pai[v/10][v%10] -= 1
							hand.Pai[i][j]++
							hand.Pai[i][0]++
							return true
						}
						hand.Pai[v/10][0] -= 1
						hand.Pai[v/10][v%10] -= 1
					}
					hand.Pai[i][0]++
					hand.Pai[i][j]++
				}
			}
		}
	}
	return true
}

/*
	判断当前牌型是否是大七对  33332
		只有一对将牌，并且其他三对都是三张一样(刻牌)的或者杠
*/
func (hand Hand) bigSevenPairs(universe int) bool {
	jang := false
	keNum := 0
	for i := 0; i < 5; i++ {
		if hand.Pai[i][0] != 0 {
			for j := 1; j < 10; j++ {
				if hand.Pai[i][j] != 0 {
					if hand.Pai[i][j] < 2 {
						return false
					} else if hand.Pai[i][j] == 2 {
						if jang == true {
							return false
						}
						jang = true
					} else if hand.Pai[i][j] >= 3 {
						keNum += 1
					}
				}
			}
		}
	}
	keNum += len(hand.Alt)
	keNum += len(hand.Bright)
	keNum += len(hand.Dark)
	fmt.Println(keNum, jang)
	if keNum == 4 && jang == true {
		return true
	}

	for i := 0; i < 5; i++ {
		for j := 1; j < 10; j++ {
			if hand.Pai[i][j] > 0 {
				if i*10+j == universe || i*10+j == NextOrder(universe) {
					hand.Pai[i][0]--
					hand.Pai[i][j]--
					for _, v := range MajBasePai {
						if v == universe || v == NextOrder(universe) {
							continue
						}
						hand.Pai[v/10][0] += 1
						hand.Pai[v/10][v%10] += 1
						if is := hand.bigSevenPairs(universe); is == true {
							logger.Debug("玩家可以胡牌---大七对")
							hand.Pai[v/10][0] -= 1
							hand.Pai[v/10][v%10] -= 1
							hand.Pai[i][j]++
							hand.Pai[i][0]++
							return true
						}
						hand.Pai[v/10][0] -= 1
						hand.Pai[v/10][v%10] -= 1
					}
					hand.Pai[i][0]++
					hand.Pai[i][j]++
				}
			}
		}
	}
	return false
}

//判断是否是十三栏, 十三栏必须门清，不能有吃  碰  杠
func (hand Hand) isThirteen() bool {
	if len(hand.Alt) != 0 || len(hand.Eat) != 0 || len(hand.Bright) != 0 || len(hand.Dark) != 0 {
		return false
	}
	//三元牌任意两两之间不靠牌和重复
	for i := 0; i < 3; i++ {
		if hand.Pai[i][0] != 0 {
			for j := 1; j <= 9; j++ {
				if hand.Pai[i][j] != 0 {
					//不重复
					if hand.Pai[i][j] > 1 {
						return false
					}
					//不靠牌
					if j == 1 {
						if hand.Pai[i][j+1] != 0 || hand.Pai[i][j+2] != 0 {
							return false
						}
					} else if j == 9 {
						if hand.Pai[i][j-1] != 0 || hand.Pai[i][j-2] != 0 {
							return false
						}
					} else if j == 2 {
						if hand.Pai[i][j-1] != 0 || hand.Pai[i][j+1] != 0 || hand.Pai[i][j+2] != 0 {
							return false
						}
					} else if j == 8 {
						if hand.Pai[i][j+1] != 0 || hand.Pai[i][j-1] != 0 || hand.Pai[i][j-2] != 0 {
							return false
						}
					} else {
						if hand.Pai[i][j+1] != 0 || hand.Pai[i][j+2] != 0 || hand.Pai[i][j-1] != 0 || hand.Pai[i][j-2] != 0 {
							return false
						}
					}
				}
			}
		}
	}
	//风牌之间不能重复
	for i := 3; i < 5; i++ {
		if hand.Pai[i][0] != 0 {
			for j := 1; j <= 9; j++ {
				if hand.Pai[i][j] != 0 {
					if hand.Pai[i][j] > 1 {
						return false
					}
				}
			}
		}
	}

	logger.Debug("玩家可以胡牌---十三烂")
	return true
}

//判断是否时七星十三栏
//在十三烂的基础上凑齐东南西北中发白
func (hand Hand) isSevenThirteen() bool {
	if hand.isThirteen() == false {
		return false
	}

	if hand.Pai[3][1] == 0 ||
		hand.Pai[3][2] == 0 ||
		hand.Pai[3][3] == 0 ||
		hand.Pai[3][4] == 0 ||
		hand.Pai[4][1] == 0 ||
		hand.Pai[4][2] == 0 ||
		hand.Pai[4][3] == 0 {
		return false
	}

	logger.Debug("玩家可以胡牌---七星十三烂")
	return true
}

//判断是否是精钓   除去精牌后其余牌构成定口牌面，就是可以分拆成刻/顺的牌面
//精钓：自摸的一种。当手中的牌通过摸、吃、碰、杠行成四副牌外加一张精的定口牌面
func (hand Hand) isUniverseDiao(universe int) bool {
	if hand.Pai[universe/10][universe%10]+hand.Pai[NextOrder(universe)/10][NextOrder(universe)%10] != 1 {
		return false
	}
	pai := 0
	if hand.Pai[universe/10][universe%10] == 1 {
		pai = universe
	} else {
		pai = NextOrder(universe)
	}
	//除去此张精牌后能行成定口
	hand.Pai[pai/10][0]--
	hand.Pai[pai/10][pai%10]--
	for i := 0; i < 3; i++ {
		if hand.Pai[i][0] > 0 {
			if i == 3 {
				if AnalyzeZi(hand.Pai[i]) == false {
					hand.Pai[pai/10][0]++
					hand.Pai[pai/10][pai%10]++
					return false
				}
			} else {
				if Analyze(hand.Pai[i]) == false {
					hand.Pai[pai/10][0]++
					hand.Pai[pai/10][pai%10]++
					return false
				}
			}
		}
	}
	hand.Pai[pai/10][0]++
	hand.Pai[pai/10][pai%10]++
	return true
}

//判断当前胡牌是否是德国 胡牌的时候没有精牌或者精牌没有发挥作用
func (hand Hand) IsDeGuo() bool {
	return hand.baseHuMode()
}

func (hand Hand) MayHu(universe int, share int) bool {
	//1:先判断万能牌在本意上面能不能胡牌
	if share != 0 {
		hand.Pai[share/10][0]++
		hand.Pai[share/10][share%10]++
	}
	if hu := hand.Hu(universe); hu == true {
		if share != 0 {
			hand.Pai[share/10][0]--
			hand.Pai[share/10][share%10]--
		}
		return true
	}
	if share != 0 {
		hand.Pai[share/10][0]--
		hand.Pai[share/10][share%10]--
	}

	//2:如果不能胡牌，将本张牌以此替换，并考虑多张万能牌
	for i := 0; i < 5; i++ {
		if hand.Pai[i][0] > 0 {
			for j := 1; j <= 9; j++ {
				if hand.Pai[i][j] > 0 {
					if i*10+j == universe || i*10+j == NextOrder(universe) {
						hand.Pai[i][0]--
						hand.Pai[i][j]--
						for _, v := range MajBasePai {
							if v == universe || v == NextOrder(universe) {
								continue
							}
							hand.Pai[v/10][0] += 1
							hand.Pai[v/10][v%10] += 1
							if is := hand.MayHu(universe, share); is == true {
								hand.Pai[v/10][0] -= 1
								hand.Pai[v/10][v%10] -= 1
								hand.Pai[i][j]++
								hand.Pai[i][0]++
								return true
							}
							hand.Pai[v/10][0] -= 1
							hand.Pai[v/10][v%10] -= 1
						}
						hand.Pai[i][0]++
						hand.Pai[i][j]++
					}
				}
			}
		}
	}
	return false
}

//判断一副牌是否可以胡牌
func (hand Hand) Hu(universe int) bool {
	//判断是否是小七对
	if hand.littleSevenPairs(universe) == true || //小七对
		hand.bigSevenPairs(universe) == true || //大七对
		hand.isThirteen() == true || //十三烂
		hand.isSevenThirteen() == true || //七星十三烂
		hand.baseHuMode() == true { //基本胡
		return true
	}

	return false
}

/*
//判断当前牌型形成顺／刻定口所需universe个数
func getNeedUniverseNum(pai []int, jiang bool) int {
	if pai[0] == 0 {
		return 0
	}
	mod := pai[0] % 3
	needNum := []int{0, 2, 1}
	if jiang {
		needNum = []int{2, 1, 0}
	}
	return needNum[mod]
}

//新的万能牌胡牌算法
func (hand Hand) UniverseHu(universe int) bool {
	//当前牌中精牌数量
	cur := hand.Pai[universe/10][universe%10] + hand.Pai[NextOrder(universe)/10][NextOrder(universe)%10]
	if cur == 0 {
		return hand.Hu()
	}

	//将精牌从手中除去，计算剩下牌如果胡牌所需精牌数量
	hand.Pai[universe/10][0] -= hand.Pai[universe/10][universe%10]
	//找出不同牌种形成定口需要的精牌数量
	for i := 0; i < 5; i++ {
		need := 0
		//如果将在本牌型中，则其他牌型必须是顺子/刻子
		//1：算出其他牌型可以组成顺子的情况下需要的精牌数量
		for j := 1; j < 5; j++ {
			need += getNeedUniverseNum(hand.Pai[(i+j)%5], false)
		}
		if need > cur {
			continue
		}
		//2：判断当前牌型形成将顺/将刻需要的精牌数量
		need += getNeedUniverseNum(hand.Pai[i], true)
		if need <= cur {
			//恢复手牌
			return true
		}
	}

	//恢复手牌
	return false
}*/
