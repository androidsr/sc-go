package model

const (
	OK       = 200
	FAIL     = 500
	OK_MSG   = "处理成功"
	FAIL_MSG = "处理失败"
)

type PageInfo struct {
	Current int64       `json:"current" keyword:"eq"`
	Size    int64       `json:"size"`
	Orders  []OrderItem `json:"orders"`
}

func (m *PageInfo) SetPage(current int64, size int64) *PageInfo {
	m.Current = current
	m.Size = size
	return m
}

func (m *PageInfo) AddOrder(column string, asc bool) *PageInfo {
	item := OrderItem{Column: column, Asc: asc}
	if m.Orders == nil || len(m.Orders) == 0 {
		m.Orders = make([]OrderItem, 0)
	}
	m.Orders = append(m.Orders, item)
	return m
}

type OrderItem struct {
	Column string `json:"column"`
	Asc    bool   `json:"asc"`
}

type SelectQueryDTO struct {
	Page     PageInfo               `json:"page" column:"-"`
	Value    string                 `json:"value"`
	Label    string                 `json:"label"`
	SupperId string                 `json:"supperId"`
	Selected []string               `json:"selected"`
	P1       string                 `json:"p1"`
	Of       string                 `json:"of"`
	Vars     map[string]interface{} `json:"vars"`
}

type SelectVO struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	SupperId string `json:"supperId"`
}

type PageResult struct {
	//当前页
	Current int64 `json:"current"`
	//分页大小
	Size int64 `json:"size"`
	//总条数
	Total int64 `json:"total"`
	//数据
	Rows interface{} `json:"rows"`
}

type HttpResult struct {
	Code int64       `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func NewFailDefault(msg string) HttpResult {
	return HttpResult{Code: FAIL, Msg: FAIL_MSG}
}

func NewFailDefaultCode(msg string) HttpResult {
	return HttpResult{Code: FAIL, Msg: msg}
}

func NewFail(code int64, msg string) HttpResult {
	return HttpResult{Code: code, Msg: msg}
}

func NewOK(data interface{}) HttpResult {
	return HttpResult{Code: OK, Msg: OK_MSG, Data: data}
}
