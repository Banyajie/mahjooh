package api

func roomInsert(req *roomReq) (int64, error) {
	res, err := db.Exec("insert into room(m_id, mc_cnt, `option`, snake, universe, rule, is_over, c_name, m_name, remark, other) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		req.MId, req.McCnt, req.Option, req.Snake, req.Universe, req.Rule, req.IsOver, req.CName, req.MName, req.Remark, req.Other)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}
