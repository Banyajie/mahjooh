package room

import (
	"strconv"
	"time"

	"chess_alg_jx/jx/jx_nsq"
	"chess_alg_jx/logger"
	"chess_alg_jx/timer"
	"chess_alg_jx/utils"
)

//操作房间类型
const (
	ROOM_ACTION_TYPE_CREAT   = 1
	ROOM_ACTION_TYPE_JOIN    = 2
	ROOM_ACTION_TYPE_LEAVE   = 3
	ROOM_ACTION_TYPE_DISCARD = 4
)

//游戏房间的数据
type Room struct {
	RoomSet      RoomSetting
	RoomId       int64     `json:"room_id"`       //房间ID
	CreatTime    int64     `json:"creat_time"`    //开房时间
	Player       [4]Player `json:"player"`        //四个玩家
	MasterId     int64     `json:"master_id"`     //房主ID
	BankerId     int       `json:"banker_id"`     //庄家ID
	McNo         int64     `json:"mc_no"`         //局索引号
	GameNum      int       `json:"game_num"`      //第几局游戏
	StartTime    int64     `json:"start_time"`    //当前游戏开始时间
	GameStatus   int       `json:"game_status"`   //当前牌局的状态
	Mahj         []int     `json:"mahj"`          //当前牌局所有牌
	LeftNum      int       `json:"left_num"`      //当前牌局还剩多少牌
	CurrentIndex int       `json:"current_index"` //当前控牌玩家ID
	MoPaiIndex   int       `json:"mo_pai_index"`  //当前摸牌玩家ID
	ChuPaiIndex  int       `json:"chu_pai_index"` //当前出牌玩家ID
	SharePai     int       `json:"share_pai"`     //最后出的牌
	UpUniverse   int       `json:"up_uneverse"`   //当前牌局的上精
	DownUniverse int       `json:"down_uneverse"` //当前牌局的下精
	Shai         []int     `json:"shai"`          //色子
	MayMes       []Monitor
	CurrentMes   []Monitor
}

//游戏定制规则
type RoomSetting struct {
	DealSpeed    uint32 `json:"deal_speed"`    //出牌速度
	McCnt        int    `json:"mc_cnt"`        //游戏局数
	Option       bool   `json:"option"`        //点炮三家付还是一家付
	Snake        bool   `json:"snake"`         //霸王精 ×2/+10
	Universe     bool   `json:"universe"`      //打出精牌算分
	SmileBack    bool   `json:"smile_back"`    //回头一笑  游戏发牌后，将上局的两张上精牌做以此精分计算
	LandMines    bool   `json:"land_mines"`    //埋地雷/终局上下翻精
	DownUniverse bool   `json:"down_universe"` //无下精
	TurnRound    bool   `json:"turn_round"`    //开局上下翻
	SameSong     bool   `json:"same_song"`     //同一首歌  假设1、2万为精，则1、2条 1、2筒都为精牌
}

/*监控控牌阶段收到的消息*/
type Monitor struct {
	SeatId   int   `json:"seat_id"`
	MessType int   `json:"mess_type"`
	Data     []int `json:"data"`
}

type CreateRoom struct {
	RoomId int64
	UserId int64
}

/*创建一个新的房间，并将房主信息写入房间玩家信息中*/
func NewRoom(roomId int64, masterId int64, set RoomSetting) *Room {
	room := &Room{
		RoomSet:      set,
		RoomId:       roomId,
		CreatTime:    time.Now().Unix(),
		Player:       [4]Player{Player{UserId: masterId}, Player{UserId: 0}, Player{UserId: 0}, Player{UserId: 0}},
		MasterId:     masterId,
		BankerId:     0,
		McNo:         0,
		GameNum:      0,
		StartTime:    time.Now().Unix(),
		GameStatus:   GAME_STATUS_FREE,
		Mahj:         MajPool,
		LeftNum:      len(MajPool),
		CurrentIndex: 0,
		MoPaiIndex:   0,
		ChuPaiIndex:  0,
		SharePai:     0,
		UpUniverse:   0,
		DownUniverse: 0,
		Shai:         nil,
		MayMes:       nil,
		CurrentMes:   nil,
	}
	room.Player[0] = Player{
		UserId:    masterId,
		UserName:  "",
		Head:      "",
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
	//创建房间为房主添加一条redis数据，只要房间解散时删除数据
	if err := utils.RedisClient.SAdd("master", strconv.FormatInt(masterId, 10)); err != nil {
		logger.Error("SyncRedis SAdd", err)
		return nil
	}
	if err := utils.RedisClient.Set("return"+strconv.FormatInt(masterId, 10), strconv.FormatInt(room.RoomId, 10)); err != nil {
		logger.Error("SyncRedis SAdd", err)
		return nil
	}

	timer.SetTimer(strconv.FormatInt(room.RoomId, 10), 60*24*24, room.checkRoom, nil)
	room.SyncRedis()

	return room
}

//24小时清楚除房主外的玩家信息
func (room *Room) checkRoom(interface{}) {
	logger.Debug("时间超时---不是房主离开房间--")
	if err := room.GetRoomData(strconv.FormatInt(room.RoomId, 10)); err != nil {
		return
	}
	for i := 1; i < 4; i++ {
		if room.Player[i].UserId != 0 && room.Player[i].State != PLAYER_STATE_LEAF {
			for j := 0; j < 4; j++ {
				if room.Player[j].UserId != 0 && room.Player[j].State != PLAYER_STATE_LEAF {
					if err := room.GameNsqProducer(SEND_PLAYER_OUT_ROOM, j, i); err != nil {
						return
					}
				}
			}
			room.Player[i].UserId = 0
			room.Player[i].State = PLAYER_STATE_LEAF
		}
	}
	room.GameStatus = GAME_STATUS_FREE
	timer.DelTimer(strconv.FormatInt(room.RoomId, 10))
	timer.SetTimer(strconv.FormatInt(room.RoomId, 10), 60*24*24, room.checkRoom, nil)
	room.SyncRedis()
	return
}

/*
	初始化房间游戏结构，为下一局游戏开始做准备
	1：修改庄家ID，上一局中第一个胡牌的玩家为庄家，或者玩家一炮多响
	2：初始化麻将牌
	3：修改游戏控制数据相关结构
	4：初始化每个玩家游戏数据
*/
func (room *Room) InitRoomGame() {
	room.GameStatus = GAME_STATUS_FREE
	room.Mahj = MajPool
	room.LeftNum = len(MajPool)
	room.CurrentIndex = room.BankerId
	room.MoPaiIndex = room.BankerId
	room.ChuPaiIndex = room.BankerId
	room.MayMes = room.MayMes[:0]
	room.CurrentMes = room.CurrentMes[:0]
	room.SharePai = 0

	for i := 0; i < 4; i++ {
		room.Player[i].InitPlayer()
	}
	logger.Debug("初始化房间---为下一局游戏准备：", room)
	room.SyncRedis()
}

/*开始游戏*/
func (room *Room) StartGame() {
	logger.Debug("----开始游戏---：")
	room.StartTime = time.Now().Unix()
	room.McNo = time.Now().Unix()
	room.GameNum++
	//更新每个玩家的总分
	for i := 0; i < 4; i++ {
		room.CurrentIndex = i
		for j := 0; j < 4; j++ {
			if err := room.GameNsqProducer(SEND_TOTAL_SCORE, j, room.Player[i].Score); err != nil {
				return
			}
		}
	}
	data := jx_nsq.Nsq_match{
		RId:      room.RoomId,
		McNo:     room.McNo,
		EastMid:  room.Player[0].UserId,
		SouthMid: room.Player[1].UserId,
		WestMid:  room.Player[2].UserId,
		NorthMid: room.Player[3].UserId,
		CName:    "chess_alg_jx",
	}
	if err := jx_nsq.NsqMatchProducer(data); err != nil {
		logger.Error("NsqMatchCardProducer: ", err)
		return
	}
	room.Shuffle()
	room.SyncRedis()
}
