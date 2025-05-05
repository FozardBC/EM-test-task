package filters

type Options struct {
	Name        *string `form:"name"`        // фильтр по имени (например, ?name=Иван)
	Surname     *string `form:"surname"`     // по фамилии
	Patronymic  *string `form:"patronymic"`  // по отчеству
	Age         *int    `form:"age"`         // точный возраст
	MinAge      *int    `form:"min_age"`     // возраст от
	MaxAge      *int    `form:"max_age"`     // возраст до
	Gender      *string `form:"gender"`      // "male"/"female"
	Nationality *string `form:"nationality"` // "ru", "us" и т.д.
}
