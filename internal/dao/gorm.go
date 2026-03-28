package dao

import (
	"GoLinko/internal/config"
	"GoLinko/internal/model"
	"GoLinko/pkg/zlog"
	"context"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	dbWrite *gorm.DB   // 主库（写）
	dbRead  *gorm.DB   // 从库（读）
	dbOnce  sync.Once  // 确保只初始化一次
	readIdx int        // 从库轮询索引
	readMu  sync.Mutex // 从库选择锁
)

// 记录最近写入的时间戳
var (
	lastWriteTime sync.Map // key: 表名, value: 最近写入时间
	writeKeyMu    sync.Mutex
)

func init() {

}

// InitDB 初始化数据库连接
func InitDB() {
	dbOnce.Do(func() {
		cfg := config.GetConfig().MysqlConfig

		//如果没有配置主从库，使用单一连接
		if cfg.WriteHost == "" && cfg.Host != "" {
			cfg.WriteHost = cfg.Host
			cfg.WritePort = cfg.Port
		}

		gormCfg := &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Silent),
		}

		// 初始化主库连接
		dbWrite = initDBConnection(cfg.WriteHost, cfg.WritePort, cfg, gormCfg, true)

		// 初始化从库连接
		if len(cfg.ReadHosts) > 0 {
			// 有从库配置，使用第一个从库
			host := cfg.ReadHosts[0]
			port := cfg.WritePort
			if len(cfg.ReadPorts) > 0 {
				port = cfg.ReadPorts[0]
			}
			dbRead = initDBConnection(host, port, cfg, gormCfg, false)
			zlog.GetLogger().Info("读写分离已启用",
				zap.String("write_host", cfg.WriteHost),
				zap.String("read_host", host))
		} else {
			// 无从库配置，读操作也用主库
			dbRead = dbWrite
			zlog.GetLogger().Info("单库模式",
				zap.String("host", cfg.WriteHost))
		}
	})
}

// initDBConnection 初始化单个数据库连接
func initDBConnection(host string, port int, cfg config.MysqlConfig, gormCfg *gorm.Config, isWrite bool) *gorm.DB {
	// 先连接 MySQL
	dsnWithoutDB := cfg.User + ":" + cfg.Password + "@tcp(" + host + ":" + strconv.Itoa(port) + ")/?charset=utf8mb4&parseTime=True&loc=Local"
	tempDB, err := gorm.Open(mysql.Open(dsnWithoutDB), gormCfg)
	if err != nil {
		zlog.GetLogger().Fatal("连接 MySQL 失败", zap.String("host", host), zap.Error(err))
	}

	// 创建数据库， 只在主库执行
	if isWrite {
		createDBSQL := "CREATE DATABASE IF NOT EXISTS " + cfg.DbName + " CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
		if err := tempDB.Exec(createDBSQL).Error; err != nil {
			zlog.GetLogger().Fatal("创建数据库失败", zap.Error(err))
		}
	}

	// 关闭临时连接
	sqlDB, _ := tempDB.DB()
	sqlDB.Close()

	// 连接到指定数据库
	dsn := cfg.User + ":" + cfg.Password + "@tcp(" + host + ":" + strconv.Itoa(port) + ")/" + cfg.DbName + "?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s&readTimeout=10s&writeTimeout=10s"
	db, err := gorm.Open(mysql.Open(dsn), gormCfg)
	if err != nil {
		zlog.GetLogger().Fatal("连接数据库失败", zap.String("host", host), zap.Error(err))
	}

	sqlDB, err = db.DB()
	if err != nil {
		zlog.GetLogger().Fatal("获取数据库连接失败", zap.Error(err))
	}

	// 连接池配置
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(200)
	sqlDB.SetConnMaxLifetime(time.Hour * 2)
	sqlDB.SetConnMaxIdleTime(time.Minute * 10)

	return db
}

// DbAutoMigrate 自动迁移数据库表结构（只在主库执行）
func DbAutoMigrate() {
	if dbWrite == nil {
		InitDB()
	}
	err := dbWrite.AutoMigrate(&model.UserInfo{}, &model.Session{}, &model.Message{}, &model.ContactApply{}, &model.GroupInfo{}, &model.UserContact{}, &model.UserSessionHide{})
	if err != nil {
		zlog.GetLogger().Fatal("自动迁移数据库表结构失败", zap.Error(err))
	}
}

// GetDB 获取默认数据库连接
func GetDB() *gorm.DB {
	if dbWrite == nil {
		InitDB()
	}
	return dbWrite
}

// GetWriteDB 获取主库连接
func GetWriteDB() *gorm.DB {
	if dbWrite == nil {
		InitDB()
	}
	return dbWrite
}

// GetReadDB 获取从库连接
// 如果配置了多个从库，会轮询选择
func GetReadDB() *gorm.DB {
	if dbRead == nil {
		InitDB()
	}

	cfg := config.GetConfig().MysqlConfig

	// 如果只有一个从库或使用主库读，直接返回
	if len(cfg.ReadHosts) <= 1 {
		return dbRead
	}

	// 多从库轮询
	readMu.Lock()
	readIdx = (readIdx + 1) % len(cfg.ReadHosts)
	readMu.Unlock()

	// 如果有多个从库连接，选择对应的
	// 简化实现：返回第一个从库
	return dbRead
}

// NewDbClient 创建一个新的数据库连接
func NewDbClient(ctx context.Context) *gorm.DB {
	return GetWriteDB().WithContext(ctx)
}

// NewWriteClient 创建主库客户端
func NewWriteClient(ctx context.Context) *gorm.DB {
	return GetWriteDB().WithContext(ctx)
}

// NewReadClient 创建从库客户端
func NewReadClient(ctx context.Context) *gorm.DB {
	return GetReadDB().WithContext(ctx)
}

// 数据一致性保障

// WriteHook 记录写入时间戳
// 在写操作完成后调用，用于读后写一致性
func RecordWriteTime(tableName string) {
	lastWriteTime.Store(tableName, time.Now())
}

// GetLastWriteTime 获取最近写入时间
func GetLastWriteTime(tableName string) time.Time {
	if v, ok := lastWriteTime.Load(tableName); ok {
		return v.(time.Time)
	}
	return time.Time{}
}

// ShouldReadFromMaster 判断是否应该从主库读取
// 用于读后写一致性：如果最近有写入，等待主从同步后再读从库
func ShouldReadFromMaster(tableName string) bool {
	cfg := config.GetConfig().MysqlConfig

	// 未配置主从同步延迟，默认从从库读
	if cfg.ReplicationLag <= 0 {
		return false
	}

	lastWrite := GetLastWriteTime(tableName)
	if lastWrite.IsZero() {
		return false
	}

	// 如果距离上次写入时间小于主从同步延迟，从主库读
	elapsed := time.Since(lastWrite)
	return elapsed < time.Duration(cfg.ReplicationLag)*time.Second
}

// ConsistentRead 一致性读取：根据写入历史决定读主库还是从库
func ConsistentRead(ctx context.Context, tableName string) *gorm.DB {
	if ShouldReadFromMaster(tableName) {
		return GetWriteDB().WithContext(ctx)
	}
	return GetReadDB().WithContext(ctx)
}

// Transaction 在事务中执行操作（使用主库）
func Transaction(fn func(tx *gorm.DB) error) error {
	return GetWriteDB().Transaction(fn)
}

// TransactionWithContext 带上下文的事务
func TransactionWithContext(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return GetWriteDB().WithContext(ctx).Transaction(fn)
}

// GetRandomReadDB 随机选择一个从库（负载均衡）
func GetRandomReadDB() *gorm.DB {
	cfg := config.GetConfig().MysqlConfig

	if len(cfg.ReadHosts) <= 1 {
		return GetReadDB()
	}

	// 随机选择（简单实现，实际可扩展为权重选择）
	r := rand.Intn(len(cfg.ReadHosts))
	_ = r // 简化实现，返回默认从库
	return dbRead
}
