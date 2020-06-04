package router

import (
	"github.com/vincentLiuxiang/lu"
	"github.com/wosai/deepmock/router/api"
)

func BuildRouter() *lu.Lu {
	app := lu.New()

	app.Get("/api/v1/rule", api.HandleGetRule)
	app.Post("/api/v1/rule", api.HandleCreateRule)
	app.Put("/api/v1/rule", api.HandlePutRule)
	app.Patch("/api/v1/rule", api.HandlePatchRule)
	app.Delete("/api/v1/rule", api.HandleDeleteRule)

	app.Get("/api/v1/rules", api.HandleExportRules)
	app.Post("/api/v1/rules", api.HandleImportRules)

	//app.Use("/", deepmock.HandleMockedAPI)

	return app
}
