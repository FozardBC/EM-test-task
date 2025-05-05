package types

type Pagination struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

func (p Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}
