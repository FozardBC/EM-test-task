package delete

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"test-task/internal/lib/api/response"
	"test-task/internal/storage"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

type PersonDeleter interface {
	Delete(ctx context.Context, id int64) error
}

// Delete godoc
//
// @Summary 	Delete
// @Description Delete the user by id
// @Tags 		people
// @ID 			create
// @Accept 		json
// @Produce 	json
// @Param		input path int true "Person ID"
// @Success 200 {object} response.Response"OK"
// @Failure 	204
// @Failure 	400 {object} response.Response "invalid id"
// @Failure 	500 {object} response.Response "Internal error"
// @Router 		/people [delete]
func New(log *slog.Logger, Deleter PersonDeleter) gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx := c.Request.Context()

		logHandler := log.With(
			slog.String("requestID", requestid.Get(c)),
		)

		paramId := c.Param("id")

		id, err := strconv.Atoi(paramId)
		if err != nil {
			logHandler.Error("can't param ID make int64")

			c.JSON(http.StatusBadRequest, response.Error("Invalid ID"))
		}

		err = Deleter.Delete(ctx, int64(id))
		if err != nil {
			if errors.Is(err, storage.ErrIDNotFound) {
				logHandler.Error("can't delete person", "err", err.Error())

				c.JSON(http.StatusNoContent, nil)
			}
			logHandler.Error("can't delete person", "err", err.Error())

			c.JSON(http.StatusInternalServerError, response.Error("Internal server error")) //TODO: УЗНАТЬ ЧТО ЛУЧШЕ ВОЗВРАЩАТЬ В ТЕКСТЕ ОШИБКИ. 500 или описание ошибки

			return
		}

		logHandler.Info("person deleted", "ID", id)

		c.JSON(http.StatusOK, response.OK())

	}
}
