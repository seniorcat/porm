package porm

import (
	"log"
	"strconv"
)

type queryStruct struct {
	insertTable string
	find        []string
	insert      []string
	where       []string
	groupBy     []string
	limit       []string
	order       []string
	join        []interface{}
	from        interface{}
}

// Start init general struct with queryStruct
func Start() (p Porm) {
	p = Porm{
		q: query(),
	}
	return
}

func query() (q queryStruct) {
	return
}

func (q queryStruct) Select(fields ...string) queryStruct {
	q.find = fields
	return q
}

func (q queryStruct) AndSelect(fields ...string) queryStruct {
	q.find = append(q.find, fields...)
	return q
}

func (q queryStruct) From(table interface{}) queryStruct {
	q.from = table
	return q
}

func (q queryStruct) Join(joinTable ...interface{}) queryStruct {
	q.join = joinTable
	return q
}

func (q queryStruct) Where(fields ...string) queryStruct {
	q.where = fields
	return q
}

func (q queryStruct) AndWhere(fields ...string) queryStruct {
	q.where = append(q.where, fields...)
	return q
}

func (q queryStruct) GroupBy(groupBy ...string) queryStruct {
	q.groupBy = groupBy
	return q
}

func (q queryStruct) AndGroupBy(groupBy ...string) queryStruct {
	q.groupBy = append(q.groupBy, groupBy...)
	return q
}

func (q queryStruct) Order(fields ...string) queryStruct {
	q.order = fields
	return q
}

func (q queryStruct) Limit(lo ...string) queryStruct {
	q.limit = lo
	return q
}

func (q queryStruct) Insert(table string, fields ...string) queryStruct {
	q.insertTable = table
	q.insert = fields
	return q
}

func (q queryStruct) prepare() (qS string) {

	if len(q.insert) > 0 {
		qS = q.prepareInsert()
	} else {
		qS = build("SELECT", q.find)
		qS += q.prepareFrom()
		qS += q.prepareJoin()
		qS += build("WHERE", q.where)
		if len(q.groupBy) > 0 {
			qS += build(" GROUP BY ", q.groupBy)
		}
		qS += q.prepareOrder()
		qS += q.prepareLimit()
	}
	return
}

func (q queryStruct) prepareFrom() (qS string) {
	qS = " FROM "
	switch q.from.(type) {
	case string:
		qS += q.from.(string)
		break
	case queryStruct:
		qF, _ := q.from.(queryStruct)
		qS += "(" + qF.prepare() + ")"
		break
	}

	return
}

func (q queryStruct) prepareJoin() (qS string) {
	for _, a := range q.join {
		switch a.(type) {
		case string:
			qS += " " + a.(string) + " "
		case queryStruct:
			qS += "(" + a.(queryStruct).prepare() + ")"
		}
	}

	return
}

func (q queryStruct) prepareOrder() (qS string) {
	if len(q.order) > 0 {
		if len(q.order) == 1 && len(q.order[0]) > 0 {
			qS = " ORDER BY " + q.order[0] + " "
		} else if len(q.order) == 2 && len(q.order[0]) > 0 && len(q.order[1]) > 0 {
			qS = " ORDER BY " + q.order[0] + " " + q.order[1]
		}
	}

	return
}

func (q queryStruct) prepareLimit() (qS string) {
	if len(q.limit) > 0 {
		qS = " LIMIT "
		if len(q.limit) == 1 {
			if len(q.limit[0]) > 0 {
				qS += q.limit[0]
			} else {
				qS += "10"
			}
		} else if len(q.limit) == 2 {
			if len(q.limit[0]) > 0 && len(q.limit[1]) > 0 {
				iLimit, _ := strconv.Atoi(q.limit[0])
				iOffset, _ := strconv.Atoi(q.limit[1])
				offset := iLimit * iOffset
				qS += strconv.Itoa(offset) + "," + q.limit[1]
			} else if len(q.limit[0]) > 0 {
				qS += "0, " + q.limit[0]
			} else if len(q.limit[1]) > 0 {
				qS += "0, " + q.limit[1]
			} else {
				qS += "0, 10"
			}

		}
	}

	return
}

func (q queryStruct) prepareInsert() (qS string) {
	if len(q.insert) > 0 && len(q.insertTable) > 0 {
		qS = "INSERT INTO " + q.insertTable + " "
		li := len(q.insert)
		for i, f := range q.insert {
			qS += f
			if i < li-1 {
				qS += ", "
			}
		}
		qS += " VALUES ("
		for i, _ := range q.insert {
			qS += "$" + strconv.Itoa(i)
			if i < li-1 {
				qS += ", "
			}

		}
	}

	return
}

func (p Porm) All() []map[string]string {
	qS := p.q.prepare()
	return p.GetDb().Query(qS)
}

func (p Porm) One() map[string]string {
	qS := p.q.prepare()
	return p.GetDb().QueryRow(qS)
}

func (p Porm) Count() int64 {
	qS := p.q.prepare()
	return p.GetDb().Count(qS)
}

func (p Porm) BulkInsert(rows ...map[string]interface{}) {
	ch := p.GetDb()
	qS := p.q.prepareInsert()
	tx, _ := ch.Begin()
	stmt := ch.PrepareInsert(tx, qS)

	if stmt != nil {
		for _, row := range rows {
			for _, f := range p.q.insert {
				var fields []interface{}
				fields = append(fields, row[f])
				ch.Exec(stmt, fields)
			}
		}

		if err := tx.Commit(); err != nil {
			log.Println(err.Error())
		}
	}
}

func build(action string, params []string) (qS string) {
	separator := ", "
	if len(params) > 0 {
		if action == "WHERE" {
			separator = " AND "
		}

		qS += " " + action + " "
		if action == "USING" {
			qS += "("
		}
		if len(params) == 1 {
			qS += params[0]
		} else {
			for i, val := range params {
				qS += val
				if i < len(params)-1 {
					qS += separator + " "
				}
			}
		}
		if action == "USING" {
			qS += ")"
		}
	}

	return
}
