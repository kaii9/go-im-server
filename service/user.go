package service

import (
	"errors"
	"time"

	"go-im-server/common"
	"go-im-server/config"
	"go-im-server/db"
	"go-im-server/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Nickname string `json:"nickname"`
}

type UserLoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginResp struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

type UserUpdateReq struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func Register(req *UserRegisterReq) error {
	var exist model.User
	if err := db.DB.Where("username = ?", req.Username).First(&exist).Error; err == nil {
		return errors.New(common.GetErrMsg(common.ErrUsernameExists))
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := model.User{
		ID:        common.GenID(),
		Username:  req.Username,
		Password:  string(hashed),
		Nickname:  req.Nickname,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if user.Nickname == "" {
		user.Nickname = req.Username
	}

	return db.DB.Create(&user).Error
}

func Login(req *UserLoginReq) (*UserLoginResp, error) {
	var user model.User
	if err := db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(common.GetErrMsg(common.ErrUserNotFound))
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New(common.GetErrMsg(common.ErrPasswordWrong))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": user.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(config.AppConfig.JWT.Secret))
	if err != nil {
		return nil, err
	}

	return &UserLoginResp{Token: tokenStr, User: user}, nil
}

func GetUserByID(uid int64) (*model.User, error) {
	var user model.User
	if err := db.DB.First(&user, uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(common.GetErrMsg(common.ErrUserNotFound))
		}
		return nil, err
	}
	return &user, nil
}

func UpdateUser(uid int64, req *UserUpdateReq) error {
	updates := map[string]interface{}{}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if len(updates) == 0 {
		return errors.New(common.GetErrMsg(common.ErrInvalidParam))
	}
	return db.DB.Model(&model.User{}).Where("id = ?", uid).Updates(updates).Error
}

func SearchUsers(keyword string) ([]model.User, error) {
	var users []model.User
	if err := db.DB.Where("username LIKE ? OR nickname LIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Select("id, username, nickname, avatar").Limit(20).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
