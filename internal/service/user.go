package service

import (
	"archi/internal/domain"
	"archi/internal/repository"
	"archi/pkg/logger"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"os"
	"path/filepath"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码不对")
)

//go:generate mockgen -source=./user.go -package=mocks -destination=./mocks/user.mock.go UserService
type UserService interface {
	Signup(ctx context.Context, user domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	UpdateAvatarPath(ctx context.Context, uid int64, newPath string) error
	UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error
	FindById(ctx context.Context, uid int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
}

type DefaultUserService struct {
	l    logger.Logger
	repo repository.UserRepository
}

func NewUserService(log logger.Logger, repo repository.UserRepository) UserService {
	return &DefaultUserService{
		l:    log,
		repo: repo,
	}
}

func (svc *DefaultUserService) Signup(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost) // bcrypt.DefaultCost 表示加密的复杂度，默认值为 10。
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *DefaultUserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 检查密码对不对
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

// UpdateAvatarPath 只负责业务逻辑：更新数据库和删除旧文件
func (svc *DefaultUserService) UpdateAvatarPath(ctx context.Context, uid int64, newPath string) error {
	// 1. 获取旧头像路径
	oldUser, err := svc.repo.FindById(ctx, uid)
	if err != nil {
		return err // 获取用户信息失败
	}
	oldAvatarPath := oldUser.Avatar

	// 2. 更新数据库为新路径
	err = svc.repo.UpdateAvatar(ctx, uid, newPath)
	if err != nil {
		// 数据库更新失败，直接返回错误。
		// 新文件的清理工作应该由调用方（Handler）负责。
		return err
	}

	// 3. 数据库更新成功后，删除旧的头像文件
	if oldAvatarPath != "" && oldAvatarPath != "path/to/default/avatar.png" { // 避免删除默认头像
		// 使用绝对路径进行删除，更安全
		absOldPath, err := filepath.Abs(oldAvatarPath)
		if err != nil {
			// 转换路径失败，只记录日志，不影响主流程
			svc.l.Warn("转换旧头像为绝对路径失败", logger.Error(err), logger.String("old_avatar_path", oldAvatarPath))
		} else {
			if err := os.Remove(absOldPath); err != nil {
				// 删除旧文件失败是一个可以容忍的错误，记录日志即可
				svc.l.Warn("数据库更新成功,但删除旧头像失败", logger.Error(err), logger.String("old_avatar_path", absOldPath))
			}
		}
	}

	return nil
}

func (svc *DefaultUserService) UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error {
	// UpdateNicknameAndXXAnd
	return svc.repo.UpdateNonZeroFields(ctx, user)
}

func (svc *DefaultUserService) FindById(ctx context.Context, uid int64) (domain.User, error) {
	return svc.repo.FindById(ctx, uid)
}

func (svc *DefaultUserService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 先找一下，我们认为，大部分用户是已经存在的用户
	u, err := svc.repo.FindByPhone(ctx, phone)
	if !errors.Is(err, repository.ErrUserNotFound) {
		// 有两种情况
		// err == nil, u 是可用的
		// err != nil，系统错误，
		return u, err
	}
	// 用户没找到
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
	})
	// 有两种可能，一种是 err 恰好是唯一索引冲突（phone）
	// 一种是 err != nil，系统错误
	if err != nil && !errors.Is(err, repository.ErrDuplicatePhone) {
		return domain.User{}, err
	}
	// 要么 err ==nil，要么ErrDuplicateUser，也代表用户存在
	// 主从延迟，理论上来讲，强制走主库
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *DefaultUserService) FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error) {
	u, err := svc.repo.FindByWechat(ctx, wechatInfo.OpenID)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return u, err
	}
	// 这边就是意味着是一个新用户
	// JSON 格式的 wechatInfo
	//zap.L().Info("新用户", zap.Any("wechatInfo", wechatInfo))
	//svc.logger.Info("新用户", zap.Any("wechatInfo", wechatInfo))
	err = svc.repo.Create(ctx, domain.User{
		WechatInfo: wechatInfo,
	})
	if err != nil && !errors.Is(err, repository.ErrDuplicateWechat) {
		return domain.User{}, err
	}
	return svc.repo.FindByWechat(ctx, wechatInfo.OpenID)
}
