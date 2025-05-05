package api

import (
	"log/slog"
	_ "test-task/docs"
	"test-task/internal/api/handlers/people/create"
	deleteHandler "test-task/internal/api/handlers/people/delete"
	"test-task/internal/api/handlers/people/list"
	"test-task/internal/api/handlers/people/update"
	"test-task/internal/services/enrich"
	"test-task/internal/storage"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	httpSwagger "github.com/swaggo/http-swagger"
)

type API struct {
	Router   *gin.Engine
	storage  storage.Storage
	log      *slog.Logger
	Enricher *enrich.Enricher
}

func New(log *slog.Logger, storage storage.Storage) *API {
	api := &API{
		Router:   gin.New(),
		storage:  storage,
		log:      log,
		Enricher: enrich.New(log),
	}

	api.Endpoints()

	return api
}

func (api *API) Endpoints() {

	v1 := api.Router.Group("api/v1/")

	v1.Use(requestid.New())
	v1.Use(gin.Logger())

	v1.GET("/people", list.New(api.log, api.storage)) //ADD SORTING
	v1.POST("/people", create.New(api.log, api.Enricher, api.storage))
	v1.PATCH("/people/:id", update.New(api.log, api.storage, api.storage))
	v1.DELETE("/people/:id", deleteHandler.New(api.log, api.storage))

	v1.GET("/swagger/*any", gin.WrapH(httpSwagger.Handler()))

}
