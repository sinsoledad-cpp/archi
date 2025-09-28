package repository

import (
	"archi/internal/domain"
	"archi/internal/repository/cache"
	"archi/internal/repository/dao"
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrDuplicatePhone  = dao.ErrDuplicatePhone
	ErrDuplicateEmail  = dao.ErrDuplicateEmail
	ErrDuplicateWechat = dao.ErrDuplicateWechat
	ErrUserNotFound    = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateAvatar(ctx context.Context, id int64, avatar string) error
	UpdateNonZeroFields(ctx context.Context, user domain.User) error
	FindByPhone(ctx context.Context, phone string) (domain.User, error)

	FindById(ctx context.Context, uID int64) (domain.User, error)
	FindByWechat(ctx context.Context, openID string) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewCachedUserRepository(userDAO dao.UserDAO, userCache cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   userDAO,
		cache: userCache,
	}
}

func (c *CachedUserRepository) Create(ctx context.Context, user domain.User) error {
	return c.dao.Insert(ctx, c.toEntity(user))
}
func (c *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := c.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return c.toDomain(u), nil
}

func (c *CachedUserRepository) UpdateAvatar(ctx context.Context, id int64, avatar string) error {
	// 更新数据库
	err := c.dao.UpdateAvatar(ctx, id, avatar)
	if err != nil {
		return err
	}
	// 操作缓存：这里选择直接删除缓存，让下一次查询重新加载
	return c.cache.Delete(ctx, id)
}

func (c *CachedUserRepository) UpdateNonZeroFields(ctx context.Context, user domain.User) error {
	// 更新 DB 之后，删除
	err := c.dao.UpdateById(ctx, c.toEntity(user))
	if err != nil {
		return err
	}
	// 延迟一秒
	time.AfterFunc(time.Second, func() {
		_ = c.cache.Delete(ctx, user.ID)
	})
	return c.cache.Delete(ctx, user.ID)
}
func (c *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := c.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return c.toDomain(u), nil
}

func (c *CachedUserRepository) FindById(ctx context.Context, uid int64) (domain.User, error) {
	du, err := c.cache.Get(ctx, uid)
	switch {
	case err == nil: // 只要 err 为 nil，就返回
		return du, nil
	case errors.Is(err, cache.ErrKeyNotExist):
		u, err := c.dao.FindById(ctx, uid)
		if err != nil {
			return domain.User{}, err
		}
		du = c.toDomain(u)
		//go func() {
		//	err = repo.cache.Set(ctx, du)
		//	if err != nil {
		//		log.Println(err)
		//	}
		//}()
		err = c.cache.Set(ctx, du)
		if err != nil {
			// 网络崩了，也可能是 redis 崩了
		}
		return du, nil
	default:
		// 接近降级的写法
		return domain.User{}, err
	}
}
func (c *CachedUserRepository) FindByWechat(ctx context.Context, openID string) (domain.User, error) {
	ue, err := c.dao.FindByWechat(ctx, openID)
	if err != nil {
		return domain.User{}, err
	}
	return c.toDomain(ue), nil
}

func (c *CachedUserRepository) toEntity(user domain.User) dao.User {
	return dao.User{
		ID: user.ID,
		Email: sql.NullString{
			String: user.Email,
			Valid:  user.Email != "",
		},
		Phone: sql.NullString{
			String: user.Phone,
			Valid:  user.Phone != "",
		},
		Password: user.Password,
		Birthday: sql.NullInt64{
			Int64: user.Birthday.UnixMilli(),
			Valid: !user.Birthday.IsZero(), // 表示这个值是有效的，不是 NULL
		},
		Avatar: user.Avatar,
		WechatUnionId: sql.NullString{
			String: user.WechatInfo.UnionID,
			Valid:  user.WechatInfo.UnionID != "",
		},
		WechatOpenId: sql.NullString{
			String: user.WechatInfo.OpenID,
			Valid:  user.WechatInfo.OpenID != "",
		},
		AboutMe:  user.AboutMe,
		Nickname: user.Nickname,
	}
}
func (c *CachedUserRepository) toDomain(u dao.User) domain.User {
	var birthday time.Time
	// 检查从数据库取出的 birthday 是否有效
	if u.Birthday.Valid {
		birthday = time.UnixMilli(u.Birthday.Int64)
	}
	return domain.User{
		ID:       u.ID,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		AboutMe:  u.AboutMe,
		Nickname: u.Nickname,
		Birthday: birthday,
		Avatar:   u.Avatar,
		Ctime:    time.UnixMilli(u.Ctime),
		WechatInfo: domain.WechatInfo{
			OpenID:  u.WechatOpenId.String,
			UnionID: u.WechatUnionId.String,
		},
	}
}
