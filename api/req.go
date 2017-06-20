package api

type roomReq struct {
	MId      int64  `json:"m_id"`     //创建人id
	McCnt    int    `json:"mc_cnt"`   //局数
	Option   int    `json:"option"`   //点炮几家付
	Snake    int    `json:"snake"`    //霸王
	Universe int    `json:"universe"` //打出的精是否算分
	Rule     int    `json:"rule"`     //玩法
	IsOver   int    `json:"is_over"`  //房间局数是否已用完 1 完 0 未完
	CName    string `json:"c_name"`   //创建人
	MName    string `json:"m_name"`   //修改人
	Remark   string `json:"remark"`   //备注
	Other    string `json:"other"`    //扩展
}

type CardCnt struct {
	MId      int64  `json:"m_id"`
	NickName string `json:"nick_name"`
	FullName string `json:"full_name"`
	CardCnt  int    `json:"card_cnt"`
	WinCnt   int    `json:"win_cnt"`
	FailCnt  int    `json:"fail_cnt"`
	DrawCnt  int    `json:"draw_cnt"`
}

type CardCntResp struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Data   CardCnt
}
