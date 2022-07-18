package miner_response

import (
	"fmt"

	"bjungle-consenso/internal/logger"
	"bjungle-consenso/internal/models"
	"github.com/asaskevich/govalidator"
)

type PortsServerMinerResponse interface {
	CreateMinerResponse(id string, lotteryId string, participantsId string, hash string) (*MinerResponse, int, error)
	UpdateMinerResponse(id string, lotteryId string, participantsId string, hash string) (*MinerResponse, int, error)
	DeleteMinerResponse(id string) (int, error)
	GetMinerResponseByID(id string) (*MinerResponse, int, error)
	GetAllMinerResponse() ([]*MinerResponse, error)
}

type service struct {
	repository ServicesMinerResponseRepository
	user       *models.User
	txID       string
}

func NewMinerResponseService(repository ServicesMinerResponseRepository, user *models.User, TxID string) PortsServerMinerResponse {
	return &service{repository: repository, user: user, txID: TxID}
}

func (s *service) CreateMinerResponse(id string, lotteryId string, participantsId string, hash string) (*MinerResponse, int, error) {
	m := NewMinerResponse(id, lotteryId, participantsId, hash)
	if valid, err := m.valid(); !valid {
		logger.Error.Println(s.txID, " - don't meet validations:", err)
		return m, 15, err
	}

	if err := s.repository.create(m); err != nil {
		if err.Error() == "ecatch:108" {
			return m, 108, nil
		}
		logger.Error.Println(s.txID, " - couldn't create MinerResponse :", err)
		return m, 3, err
	}
	return m, 29, nil
}

func (s *service) UpdateMinerResponse(id string, lotteryId string, participantsId string, hash string) (*MinerResponse, int, error) {
	m := NewMinerResponse(id, lotteryId, participantsId, hash)
	if valid, err := m.valid(); !valid {
		logger.Error.Println(s.txID, " - don't meet validations:", err)
		return m, 15, err
	}
	if err := s.repository.update(m); err != nil {
		logger.Error.Println(s.txID, " - couldn't update MinerResponse :", err)
		return m, 18, err
	}
	return m, 29, nil
}

func (s *service) DeleteMinerResponse(id string) (int, error) {
	if !govalidator.IsUUID(id) {
		logger.Error.Println(s.txID, " - don't meet validations:", fmt.Errorf("id isn't uuid"))
		return 15, fmt.Errorf("id isn't uuid")
	}

	if err := s.repository.delete(id); err != nil {
		if err.Error() == "ecatch:108" {
			return 108, nil
		}
		logger.Error.Println(s.txID, " - couldn't update row:", err)
		return 20, err
	}
	return 28, nil
}

func (s *service) GetMinerResponseByID(id string) (*MinerResponse, int, error) {
	if !govalidator.IsUUID(id) {
		logger.Error.Println(s.txID, " - don't meet validations:", fmt.Errorf("id isn't uuid"))
		return nil, 15, fmt.Errorf("id isn't uuid")
	}
	m, err := s.repository.getByID(id)
	if err != nil {
		logger.Error.Println(s.txID, " - couldn`t getByID row:", err)
		return nil, 22, err
	}
	return m, 29, nil
}

func (s *service) GetAllMinerResponse() ([]*MinerResponse, error) {
	return s.repository.getAll()
}
