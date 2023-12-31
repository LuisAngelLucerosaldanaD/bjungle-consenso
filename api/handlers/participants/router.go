package participants

import (
	"bjungle-consenso/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

func RouterParticipants(app *fiber.App, db *sqlx.DB, txID string) {
	h := handlerParticipant{DB: db, TxID: txID}
	api := app.Group("/api")
	v1 := api.Group("/v1")
	partRouter := v1.Group("/participants")
	partRouter.Post("/register", middleware.JWTProtected(), h.RegisterParticipant)
	partRouter.Get("/info/:wallet", middleware.JWTProtected(), h.GetInfoParticipant)
}
