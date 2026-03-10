package service

import (
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/model"
	"github.com/mpv945/openwrt-custom-tools-gin/src/com/ewancle/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		repo: repository.NewUserRepository(),
	}
}

func (s *UserService) GetUser(id int64) (*model.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) CreateUser(name, email string) (*model.User, error) {

	user := &model.User{
		Name:  name,
		Email: email,
	}

	err := s.repo.Create(user)

	return user, err
}
