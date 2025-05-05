package types

type Meta struct {
	Total  int  `form:"total"`
	Limit  int  `form:"limit"`
	Offset int  `form:"offset"`
	Next   bool `form:"next"`
}
