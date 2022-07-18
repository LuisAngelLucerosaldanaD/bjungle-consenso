package reward_table

import (
	"github.com/jmoiron/sqlx"

	"bjungle-consenso/internal/logger"
	"bjungle-consenso/internal/models"
)

const (
	Postgresql = "postgres"
)

type ServicesRewardTableRepository interface {
	create(m *RewardTable) error
	update(m *RewardTable) error
	delete(id string) error
	getByID(id string) (*RewardTable, error)
	getAll() ([]*RewardTable, error)
}

func FactoryStorage(db *sqlx.DB, user *models.User, txID string) ServicesRewardTableRepository {
	var s ServicesRewardTableRepository
	engine := db.DriverName()
	switch engine {
	case Postgresql:
		return newRewardTablePsqlRepository(db, user, txID)
	default:
		logger.Error.Println("el motor de base de datos no está implementado.", engine)
	}
	return s
}