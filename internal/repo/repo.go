package repo

import (
	"context"
	"errors"

	"github.com/ego-component/egorm"
	"gorm.io/gorm"

	"shorturl/internal/model"
)

// 数据访问接口：定义短链的基础 CRUD 与批量写入
type Repository interface {
	Create(ctx context.Context, code, longURL string) (*model.ShortURL, error)
	FindByCode(ctx context.Context, code string) (*model.ShortURL, error)
	FindByLong(ctx context.Context, longURL string) (*model.ShortURL, error)
	DeleteByCode(ctx context.Context, code string) error
	BatchCreate(ctx context.Context, items map[string]string) ([]model.ShortURL, error)
}

type GormRepo struct {
	db *egorm.Component
}

// 使用 Egorm 创建仓库实现
func NewGormRepo(db *egorm.Component) *GormRepo {
	return &GormRepo{db: db}
}

// 自动迁移表结构
func (r *GormRepo) Migrate() error {
	return r.db.AutoMigrate(&model.ShortURL{})
}

// 创建一条短链记录
func (r *GormRepo) Create(ctx context.Context, code, longURL string) (*model.ShortURL, error) {
	item := &model.ShortURL{Code: code, LongURL: longURL}
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

// 按短码查询
func (r *GormRepo) FindByCode(ctx context.Context, code string) (*model.ShortURL, error) {
	var item model.ShortURL
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &item, err
}

// 按长链查询（用于幂等）
func (r *GormRepo) FindByLong(ctx context.Context, longURL string) (*model.ShortURL, error) {
	var item model.ShortURL
	err := r.db.WithContext(ctx).Where("long_url = ?", longURL).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &item, err
}

// 删除指定短码
func (r *GormRepo) DeleteByCode(ctx context.Context, code string) error {
	return r.db.WithContext(ctx).Where("code = ?", code).Delete(&model.ShortURL{}).Error
}

// 批量创建短链记录
func (r *GormRepo) BatchCreate(ctx context.Context, items map[string]string) ([]model.ShortURL, error) {
	list := make([]model.ShortURL, 0, len(items))
	for code, long := range items {
		list = append(list, model.ShortURL{Code: code, LongURL: long})
	}
	if len(list) == 0 {
		return nil, nil
	}
	if err := r.db.WithContext(ctx).Create(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
