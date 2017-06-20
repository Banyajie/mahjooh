package room

import (
	"chess_alg_jx/utils"
)

/*一副江西南昌麻将的所有牌*/
var MajPool = []int{
	1, 2, 3, 4, 5, 6, 7, 8, 9, //万
	1, 2, 3, 4, 5, 6, 7, 8, 9,
	1, 2, 3, 4, 5, 6, 7, 8, 9,
	1, 2, 3, 4, 5, 6, 7, 8, 9,
	11, 12, 13, 14, 15, 16, 17, 18, 19, //筒
	11, 12, 13, 14, 15, 16, 17, 18, 19,
	11, 12, 13, 14, 15, 16, 17, 18, 19,
	11, 12, 13, 14, 15, 16, 17, 18, 19,
	21, 22, 23, 24, 25, 26, 27, 28, 29, //条
	21, 22, 23, 24, 25, 26, 27, 28, 29,
	21, 22, 23, 24, 25, 26, 27, 28, 29,
	21, 22, 23, 24, 25, 26, 27, 28, 29,
	31, 32, 33, 34, 31, 32, 33, 34, //东南西北
	31, 32, 33, 34, 31, 32, 33, 34, //东南西北
	41, 42, 43, 41, 42, 43, //中 发 白
	41, 42, 43, 41, 42, 43, //中 发 白
}

//玩家手牌中类型
const (
	HANDCARDS_TYPE_EAT = iota
	HANDCARDS_TYPE_ALT
	HANDCARDS_TYPE_BRIGHT
	HANDCARDS_TYPE_DARK
	HANDCARDS_TYPE_HAND
)

/*所有的牌值*/
var MajBasePai = []int{
	1, 2, 3, 4, 5, 6, 7, 8, 9,
	11, 12, 13, 14, 15, 16, 17, 18, 19,
	21, 22, 23, 24, 25, 26, 27, 28, 29,
	31, 32, 33, 34,
	41, 42, 43,
}

//发牌器: 每次从一副麻将中返回一张牌, 没有牌时返回0
func DealOneCard(maj []int) int {
	//找到第一张牌
	for i := 0; i < len(MajPool); i++ {
		if maj[i] != 0 {
			temp := maj[i]
			maj[i] = 0
			return temp
		}
	}
	return 0
}

//由一张牌求出本张牌的下一张序牌
func NextOrder(pai int) int {
	//牌值必须属于麻将池中
	if b, _ := utils.Contain(pai, MajBasePai); b == false {
		return 0
	}

	if pai <= 29 { //如果时万筒条
		if (pai+1)%10 == 0 {
			return pai - 8
		} else {
			return pai + 1
		}
	} else if pai <= 34 && pai >= 31 { //风
		if pai == 34 {
			return 31
		} else {
			return pai + 1
		}
	} else if pai <= 43 && pai >= 41 {
		if pai == 43 {
			return 41
		} else {
			return pai + 1
		}
	}

	return 0
}

/*
	玩家手中牌墙表示
	[0][]    //万
	[1][]    //筒
	[2][]    //条
	[3][]    //风
	[4][]    //箭
*/
//玩家手中牌墙
type WallCards [][]int

//吃牌的数据结构
type EatPai []int

/*玩家游戏中的手牌结构*/
type Hand struct {
	Pai    WallCards `json:"pai"`    //手牌
	Alt    []int     `json:"alt"`    //碰牌
	Bright []int     `json:"bright"` //明杠
	Dark   []int     `json:"dark"`   //暗杠
	Eat    []EatPai  `json:"eat"`    //吃牌
}

/*胡牌的数据结构*/
type Win struct {
	Selfdrawn bool  `json:"selfdrawn"` //是否是自摸
	Award     []int `json:"award"`     //胡牌的奖励
	Share     int   `json:"share"`     //胡的那张牌
	Loser     int   `json:"loser"`     //点炮的玩家
}

/*听牌的数据结构*/
type Ting struct {
	PlayCard   int `json:"play_card"`
	ListenCard int `json:"listen_card"`
	Score      int `json:"score"`
	LeftNum    int `json:"left_num"`
}

/*
	将玩家手中二维数组的手牌转化为一维数组表示的方式
*/
func (hand Hand) ReverChange() []int {
	var reselt []int

	for i := 0; i < 5; i++ {
		if hand.Pai[i][0] != 0 {
			for j := 1; j <= 9; j++ {
				if hand.Pai[i][j] != 0 {
					for k := 0; k < hand.Pai[i][j]; k++ {
						reselt = append(reselt, 10*i+j)
					}
				}
			}
		}
	}

	return reselt
}

/*
	判断玩家手牌是否可以开杠
*/
func (hand Hand) IsWind(value int) bool {
	var ok bool
	var err error
	if ok, err = utils.Contain(value, hand.Alt); err != nil {
		return false
	}

	return ok
}

/*
	判断玩家手牌是否可以暗杠
*/
func (hand Hand) IsConcealedKong() int {
	for i := 0; i < 5; i++ {
		for j := 1; j <= 9; j++ {
			if hand.Pai[i][j] == 4 {
				return i*10 + j
			}
		}
	}
	return 0
}

//	从玩家手中取出默认的一张牌
func (hand Hand) DefaultPlayAHand() int {
	value := 0
	tmp := hand.ReverChange()
	if tmp == nil {
		return 0
	}
	value = tmp[len(tmp)-1]

	return value
}

/*判断玩家是否可以吃牌*/
func (hand Hand) IsMayEat(eat int) bool {
	remainder := eat % 10
	mod := eat / 10
	b := false

	if mod == 3 {
		num := 0
		for i := 1; i <= 4; i++ {
			if hand.Pai[mod][i] != 0 && i != remainder {
				num++
			}
		}
		if num >= 2 {
			return true
		}
	} else {
		if remainder == 1 {
			if hand.Pai[mod][remainder+1] != 0 && hand.Pai[mod][remainder+2] != 0 {
				b = true
			}
		} else if remainder == 9 {
			if hand.Pai[mod][remainder-1] != 0 && hand.Pai[mod][remainder-2] != 0 {
				b = true
			}
		} else if remainder == 2 {
			if hand.Pai[mod][remainder+1] != 0 && hand.Pai[mod][remainder+2] != 0 {
				b = true
			}
			if hand.Pai[mod][remainder+1] != 0 && hand.Pai[mod][remainder-1] != 0 {
				b = true
			}
		} else if remainder == 8 {
			if hand.Pai[mod][remainder-1] != 0 && hand.Pai[mod][remainder+1] != 0 {
				b = true
			}
			if hand.Pai[mod][remainder-1] != 0 && hand.Pai[mod][remainder-2] != 0 {
				b = true
			}
		} else {
			if hand.Pai[mod][remainder+1] != 0 && hand.Pai[mod][remainder+2] != 0 {
				b = true
			}
			if hand.Pai[mod][remainder-1] != 0 && hand.Pai[mod][remainder-2] != 0 {
				b = true
			}
			if hand.Pai[mod][remainder-1] != 0 && hand.Pai[mod][remainder+1] != 0 {
				b = true
			}
		}
	}

	return b
}
