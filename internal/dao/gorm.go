package dao

import (
	"GoLinko/internal/config"
	"GoLinko/internal/model"
	"GoLinko/pkg/zlog"
	"context"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var db *gorm.DB

func init() {
	if config.GetConfig() == nil || config.GetConfig().MysqlConfig.Host == "" {
		return
	}
	cfg := config.GetConfig().MysqlConfig

	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}

	// 先连接 MySQL
	dsnWithoutDB := cfg.User + ":" + cfg.Password + "@tcp(" + cfg.Host + ":" + strconv.Itoa(cfg.Port) + ")/?charset=utf8mb4&parseTime=True&loc=Local"
	tempDB, err := gorm.Open(mysql.Open(dsnWithoutDB), gormCfg)
	if err != nil {
		zlog.GetLogger().Fatal("连接 MySQL 失败", zap.Error(err))
	}

	// 创建数据库（如果不存在）
	createDBSQL := "CREATE DATABASE IF NOT EXISTS " + cfg.DbName + " CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
	if err := tempDB.Exec(createDBSQL).Error; err != nil {
		zlog.GetLogger().Fatal("创建数据库失败", zap.Error(err))
	}

	// 连接到指定数据库
	dsn := cfg.User + ":" + cfg.Password + "@tcp(" + cfg.Host + ":" + strconv.Itoa(cfg.Port) + ")/" + cfg.DbName + "?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s&readTimeout=10s&writeTimeout=10s"
	db, err = gorm.Open(mysql.Open(dsn), gormCfg)
	if err != nil {
		zlog.GetLogger().Fatal("连接数据库失败", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		zlog.GetLogger().Fatal("获取数据库连接失败", zap.Error(err))

	}
	// 数据库连接池配置
	// SetMaxIdleConns: 设置连接池中空闲连接的最大数量
	// 空闲连接过多会占用资源，过少会导致频繁创建新连接
	sqlDB.SetMaxIdleConns(20)

	// SetMaxOpenConns: 设置连接池中打开连接的最大数量
	// 根据实际业务量调整，避免过多连接导致数据库压力
	sqlDB.SetMaxOpenConns(200)

	// SetConnMaxLifetime: 设置连接的最大生命周期
	// MySQL 默认 wait_timeout 通常是 8 小时，设置为小于该值避免连接被服务器关闭
	// 这里设置为 2 小时，确保连接在超时前被关闭
	sqlDB.SetConnMaxLifetime(time.Hour * 2)

	// SetConnMaxIdleTime: 设置空闲连接的最大空闲时间
	// 空闲连接在一段时间后自动关闭，释放资源
	// 这里设置为 10 分钟，让空闲连接在 10 分钟后被回收
	sqlDB.SetConnMaxIdleTime(time.Minute * 10)

}
func DbAutoMigrate() { //自动迁移数据库表结构
	err := db.AutoMigrate(&model.UserInfo{}, &model.Session{}, &model.Message{}, &model.ContactApply{}, &model.GroupInfo{}, &model.UserContact{}, &model.UserSessionHide{})
	if err != nil {
		zlog.GetLogger().Fatal("自动迁移数据库表结构失败", zap.Error(err))

	}
}
func GetDB() *gorm.DB {
	return db
}

// NewDbClient 创建一个新的数据库连接，使用上下文管理连接生命周期
func NewDbClient(ctx context.Context) *gorm.DB {
	if ctx == nil {
		ctx = context.Background()
	}
	if db == nil {
		return nil
	}
	return db.WithContext(ctx)
}
