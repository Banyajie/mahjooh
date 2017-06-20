package room

/*玩家正在游戏中所需的数据结构*/
type Player struct {
	UserId     int64   `json:"user_id"`    //玩家ID
	UserName   string  `json:"user_name"`  //玩家姓名
	Head       string  `json:"head"`       //玩家头像
	Score      int     `json:"score"`      //玩家在房间中的钱
	MScore     int     `json:"m_score"`    //玩家一局游戏的钱
	State      int     `json:"state"`      //玩家状态
	Discard    bool    `json:"discard"`    //解散房间状态
	MayHu      bool    `json:"may_hu"`     //是否可以胡牌
	DrawTimes  int     `json:"draw_times"` //玩家摸牌次数
	Opt        []int   `json:"player_opt"` //玩家游戏中的动作
	Played     []int   `json:"played"`     //玩家本局打出的牌
	HandCard   Hand    //玩家的手牌
	WinData    Win     //玩家胡牌信息
	ScoreWater []Score //玩家本局游戏的流水
}

/*玩家游戏中状态*/
const (
	PLAYER_STATE_FREE              = iota //空闲
	PLAYER_STATE_LEAF                     //离开
	PLAYER_STATE_SHUFFLED                 //洗牌完成
	PLAYER_STATE_DICE_BANKERED            //定庄完成
	PLAYER_STATE_SHOW_UNIVERSE_END        //展示精牌完成
	PLAYER_STATE_READY                    //已准备
	PLAYER_STATE_PLAING                   //游戏中
	PLAYER_STATE_DOWN                     //掉线
)

//玩家在游戏过程中的状态，每次玩家有所动作
const (
	PLAYING_STATUS_PLAY       = iota //出牌
	PLAYING_STATUS_EAT               //吃牌
	PLAYING_STATUS_ALT               //碰牌
	PLAYING_STATUS_ADD_BAR           //加杠
	PLAYING_STATUS_BRIGHT_BAR        //明杠
	PLAYING_STATUS_DARK_BAR          //暗杠
)

/*初始化玩家手牌信息*/
func (player *Player) InitPlayer() {
	player.MScore = 0
	player.State = PLAYER_STATE_FREE
	player.DrawTimes = 0
	player.MayHu = true
	player.Opt = player.Opt[:0]
	player.Played = player.Played[:0]

	for i := 0; i < 5; i++ {
		for j := 0; j < 10; j++ {
			player.HandCard.Pai[i][j] = 0
		}
	}

	player.HandCard.Eat = player.HandCard.Eat[:0]
	player.HandCard.Alt = player.HandCard.Alt[:0]
	player.HandCard.Bright = player.HandCard.Bright[:0]
	player.HandCard.Dark = player.HandCard.Dark[:0]
	player.ScoreWater = player.ScoreWater[:0]
	player.WinData.Selfdrawn = false
	player.WinData.Share = -1
	player.WinData.Loser = -1
}
