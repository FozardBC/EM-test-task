package list

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"test-task/internal/api/types"
	"test-task/internal/domain/filters"
	"test-task/internal/domain/models"
	"test-task/internal/lib/api/response"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

var ErrConvertParam = errors.New("can't convernt int query parameter")

const (
	defaultLimit = "10"
	defaultPage  = "1"
)

type Response struct {
	Resp response.Response `json:"response"`
	Data []*models.Person  `json:"data"`
	Meta *types.Meta       `json:"meta"`
}

type Pager interface {
	FilteredPages(ctx context.Context, offset int, limit int, options *filters.Options) ([]*models.Person, int, error)
}

// Listgodoc
// @Summary      List poeple
// @Description  get accounts by filters
// @Tags         poeple
// @Accept       json
// @Produce      json
// @Param        page			query  int		false  "num of page"					example(1)			default:"1"
// @Param        limit			query  int		false  "limit wrties on page"			example(3)			default:"10"
// @Param        name			query  string	false  "person filter by name"			example(oleg)
// @Param        surname		query  string	false  "person filter by surname"		example(invanov)
// @Param        patronymic		query  string	false  "person filter by patronymic"	example(petrovich)
// @Param        age    		query  int		false  "person filter by exact age" 	example(32)
// @Param        minage			query  int		false  "person filter by min age" 		example(10)
// @Param        maxage			query  int		false  "person filter by max age" 		example(35)
// @Param        gender 		query  string	false  "person filter by gender" 		example(male)
// @Param        nationality	query  string	false  "person filter by nationality" 	example(RU)
// @Success      200  {object}  Response
// @Failure      400  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /people [get]
func New(log *slog.Logger, Pager Pager) gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx := c.Request.Context()

		logHandler := log.With(
			slog.String("requestID", requestid.Get(c)),
		)

		filterOp, err := setFilterQueries(logHandler, c)
		if err != nil {
			if errors.Is(err, ErrConvertParam) {
				logHandler.Error(ErrConvertParam.Error(), "err", err)

				c.JSON(http.StatusBadRequest, response.Error("Invalid Parameters"))
			}
		}

		var pag types.Pagination

		pageQuery := c.DefaultQuery("page", defaultPage)
		if pageQuery == "0" {
			logHandler.Error("page can't be 0")

			c.JSON(http.StatusBadRequest, "page parameter can't be 0")
		}

		pag.Page, err = strconv.Atoi(pageQuery)
		if err != nil {
			logHandler.Error(ErrConvertParam.Error(), "param", "page", "query", pageQuery)

			c.JSON(http.StatusBadRequest, response.Error(fmt.Sprintf("Invalid parameter:%s", pageQuery)))

			return
		}

		limitQurey := c.DefaultQuery("limit", defaultLimit)

		pag.Limit, err = strconv.Atoi(limitQurey)
		if err != nil {
			logHandler.Error(ErrConvertParam.Error(), "param", "limit", "query", limitQurey)

			c.JSON(http.StatusBadRequest, response.Error(fmt.Sprintf("Invalid parameter:%s", limitQurey)))

			return
		}

		offset := pag.Offset()

		logHandler.Debug("Pagination query", "page", pag.Page, "limit", pag.Limit)

		users, count, err := Pager.FilteredPages(ctx, offset, pag.Limit, filterOp)
		if err != nil {
			logHandler.Error("can't get list of persons", "err", err.Error())

			c.JSON(http.StatusInternalServerError, response.Error("Internal Server Error"))
			return
		}

		meta := &types.Meta{
			Total:  count,
			Limit:  pag.Limit,
			Offset: pag.Offset(),
			Next:   (pag.Offset() + pag.Limit) < count,
		}

		c.JSON(http.StatusOK, Response{Resp: response.OK(), Data: users, Meta: meta})

	}
}

func setFilterQueries(log *slog.Logger, c *gin.Context) (*filters.Options, error) {

	var op filters.Options

	name := c.Query("name")
	if name != "" {
		op.Name = &name
	}

	surname := c.Query("surname")
	if surname != "" {
		op.Surname = &surname
	}

	patronymic := c.Query("patronymic")
	if patronymic != "" {
		op.Patronymic = &patronymic
	}

	gender := c.Query("gender")
	if gender != "" {
		op.Gender = &gender
	}

	nationality := c.Query("nationality")
	if nationality != "" {
		op.Nationality = &nationality
	}

	age := c.Query("age") //точный возраст
	if age != "" {
		ageInt, err := strconv.Atoi(age)
		if err != nil {
			log.Error(ErrConvertParam.Error(), "param", "age", "query", age)

			return nil, fmt.Errorf("%w:%s", ErrConvertParam, age)
		}

		op.Age = &ageInt

		return &op, nil
	}

	minAge := c.Query("minAge")
	if minAge != "" {
		minAgeInt, err := strconv.Atoi(age)
		if err != nil {
			log.Error(ErrConvertParam.Error(), "param", "minAge", "query", minAge)

			return nil, fmt.Errorf("%w:%s", ErrConvertParam, minAge)
		}

		op.MinAge = &minAgeInt
	}

	maxAge := c.Query("maxAge")
	if maxAge != "" {
		maxAgeInt, err := strconv.Atoi(maxAge)
		if err != nil {
			log.Error(ErrConvertParam.Error(), "param", "maxAge", "query", maxAge)

			return nil, fmt.Errorf("%w:%s", ErrConvertParam, minAge)
		}

		op.MaxAge = &maxAgeInt
	}

	return &op, nil
}
