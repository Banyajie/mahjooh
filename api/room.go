package api

import (
	"github.com/gin-gonic/gin"
	"net/http"

	"chess_alg_jx/config"
	"chess_alg_jx/jx/jx_nsq"
	"chess_alg_jx/jx/room"
	"chess_alg_jx/utils"
	"github.com/levigross/grequests"
	"strconv"
)

func RouteRegister() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.POST("/alg_api/room", addRoom)
	router.GET("/alg_api/test", index)
	router.GET("/alg_api/room/return", returnRoom)
	router.GET("/alg_api/room/join", joinRoom)
	router.GET("/alg_api/room/rule", rulesOfRoom)
	router.GET("/alg_api/room/judge", isCreate)

	router.Run(config.Config.ApiAddr)
}

func index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "1", "msg": "success", "data": nil})
}

func addRoom(c *gin.Context) {
	var args roomReq
	if err := c.BindJSON(&args); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	resp, err := grequests.Get("http://www.duobaoqu.cn/api/profile/"+strconv.FormatInt(args.MId, 10), nil)
	if nil != err {
		c.JSON(http.StatusOK, gin.H{"status": "0", "msg": "req mem room card error", "data": nil})
		return
	}
	if nil == resp {
		c.JSON(http.StatusOK, gin.H{"status": "0", "msg": "resp is nil", "data": nil})
		return
	}
	res := &CardCntResp{}
	err = resp.JSON(res)
	if nil != err {
		c.JSON(http.StatusOK, gin.H{"status": "0", "msg": "resp Json is nil", "data": nil})
		return
	}
	if res.Status == -1 {
		c.JSON(http.StatusOK, gin.H{"status": "0", "msg": "resp status is -1", "data": nil})
		return
	}
	if res.Data.CardCnt < args.McCnt {
		c.JSON(http.StatusOK, gin.H{"status": "0", "msg": "room card is not enough", "data": nil})
		return
	}

	roomSet := room.RoomSetting{
		DealSpeed:    uint32(config.Config.DealSpead),
		McCnt:        1,
		Option:       false,
		Snake:        false,
		Universe:     false,
		SmileBack:    false,
		LandMines:    false,
		DownUniverse: false,
		TurnRound:    false,
		SameSong:     false,
	}
	roomSet.McCnt = args.McCnt
	if args.Option == 3 {
		roomSet.Option = true
	} else if args.Option == 4 {
		roomSet.Option = true
	} else {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if args.Snake == 5 {
		roomSet.Snake = true
	} else if args.Snake == 6 {
		roomSet.Snake = false
	} else {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if args.Universe == 7 {
		roomSet.Universe = true
	} else if args.Universe == 8 {
		roomSet.Universe = false
	} else {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	switch args.Rule {
	case 9:
		roomSet.SmileBack = true
	case 10:
		roomSet.LandMines = true
	case 11:
		roomSet.DownUniverse = true
	case 12:
		roomSet.TurnRound = true
	case 13:
		roomSet.SameSong = true
	default:
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	r_id, err := roomInsert(&args)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "0", "msg": err.Error(), "data": nil})
		return
	}

	room_log := jx_nsq.Nsq_room_log{
		MId:    args.MId,
		RlType: room.ROOM_ACTION_TYPE_CREAT,
		CName:  "chess_alg_jx",
	}
	if err := jx_nsq.NsqRoomLogProducer(room_log); err != nil {
		return
	}

	rcard := jx_nsq.Nsq_Rcard{
		MId:   args.MId,
		Rid:   r_id,
		Type:  jx_nsq.RCARD_TYPE_DEDUCT,
		Value: args.McCnt,
		CName: "chess_alg_jx",
	}
	if err := jx_nsq.NsqRcardProducer(rcard); err != nil {
		return
	}

	room.NewRoom(r_id, args.MId, roomSet)
	c.JSON(http.StatusOK, gin.H{"status": "1", "msg": "success", "r_id": r_id, "data": roomSet})
}

//返回房间
func returnRoom(c *gin.Context) {
	u_id := c.Query("u_id")
	if u_id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "msg": "req parameter is error", "data": nil})
		return
	}

	if _, err := utils.RedisClient.IsKeyExit("return" + u_id); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "-1", "msg": "room not exist", "data": nil})
		return
	}

	var r_id string
	var err error
	if r_id, err = utils.RedisClient.Get("return" + u_id); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "-1", "msg": "room not exist", "data": nil})
		return
	}

	rm := room.Room{}
	if err := rm.GetRoomData(r_id); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "-1", "msg": "room not exist", "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "1", "msg": "sucess", "r_id": rm.RoomId, "data": rm.RoomSet})
}

//房间规则
func rulesOfRoom(c *gin.Context) {
	r_id := c.Query("r_id")
	if r_id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "msg": "roomid is nil", "data": nil})
		return
	}
	rm := room.Room{}
	if err := rm.GetRoomData(r_id); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "-1", "msg": "room not exist", "data": nil})
		return
	}
	banker := ""
	for i := 0; i < 4; i++ {
		if rm.Player[i].UserId == rm.MasterId {
			banker = rm.Player[i].UserName
			break
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "1", "msg": "sucess", "banker": banker, "data": rm.RoomSet})
}

func isCreate(c *gin.Context) {
	u_id := c.Query("u_id")
	if u_id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "msg": "u_id is nil", "data": nil})
		return
	}
	master := false
	var err error
	if master, err = utils.RedisClient.SIsMember("master", u_id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "msg": "redis.sIsMember error", "data": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "1", "msg": "sucess", "data": master})
}

func joinRoom(c *gin.Context) {
	r_id := c.Query("r_id")
	if r_id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "msg": "roomid is nil", "data": nil})
		return
	}
	rm := room.Room{}
	if err := rm.GetRoomData(r_id); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "-1", "msg": "room not exist", "data": nil})
		return
	}
	num := 0
	for i := 0; i < 4; i++ {
		if rm.Player[i].UserId != 0 && rm.Player[i].State != room.PLAYER_STATE_LEAF {
			num++
		}
	}
	if num == 4 {
		c.JSON(http.StatusOK, gin.H{"status": "-2", "msg": "room have four player!", "data": nil})
		return
	}

	if rm.GameNum == rm.RoomSet.McCnt*8 {
		c.JSON(http.StatusOK, gin.H{"status": "-3", "msg": "the room have no match", "data": nil})
		return
	}
	banker := ""
	for i := 0; i < 4; i++ {
		if rm.Player[i].UserId == rm.MasterId {
			banker = rm.Player[i].UserName
			break
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "1", "msg": "sucess", "banker": banker, "data": rm.RoomSet})
}
