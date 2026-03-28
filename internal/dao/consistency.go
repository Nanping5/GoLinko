package dao

import (
	"context"
	"time"

	"GoLinko/pkg/zlog"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ConsistencyManager 数据一致性管理器
// 用于保障读写分离场景下的数据一致性
type ConsistencyManager struct {
	replicationLag time.Duration // 主从同步延迟容忍时间
}

var consistencyManager *ConsistencyManager

// InitConsistencyManager 初始化一致性管理器
func InitConsistencyManager(replicationLagSeconds int) {
	consistencyManager = &ConsistencyManager{
		replicationLag: time.Duration(replicationLagSeconds) * time.Second,
	}
}

// GetConsistencyManager 获取一致性管理器
func GetConsistencyManager() *ConsistencyManager {
	if consistencyManager == nil {
		consistencyManager = &ConsistencyManager{
			replicationLag: 3 * time.Second,
		}
	}
	return consistencyManager
}

//读后写一致性

// WriteOperation 写操作包装器
// 自动记录写入时间戳，用于后续读操作的一致性判断
func WriteOperation(ctx context.Context, tableName string, fn func(db *gorm.DB) error) error {
	err := fn(GetWriteDB().WithContext(ctx))
	if err != nil {
		return err
	}

	// 记录写入时间
	RecordWriteTime(tableName)
	return nil
}

// ReadOperation 读操作包装器
// 根据写入历史决定是否从主库读取
func ReadOperation(ctx context.Context, tableName string, fn func(db *gorm.DB) error) error {
	db := ConsistentRead(ctx, tableName)
	return fn(db)
}

//会话一致性

// SessionContext 会话上下文
// 在一个请求周期内保持读写一致性
type SessionContext struct {
	ctx          context.Context
	writeTables  map[string]bool // 本次会话写入过的表
	sessionStart time.Time       // 会话开始时间
}

// NewSessionContext 创建会话上下文
func NewSessionContext(ctx context.Context) *SessionContext {
	return &SessionContext{
		ctx:          ctx,
		writeTables:  make(map[string]bool),
		sessionStart: time.Now(),
	}
}

// Write 执行写操作并记录
func (s *SessionContext) Write(tableName string, fn func(db *gorm.DB) error) error {
	err := fn(GetWriteDB().WithContext(s.ctx))
	if err != nil {
		return err
	}
	s.writeTables[tableName] = true
	RecordWriteTime(tableName)
	return nil
}

// Read 执行读操作，如果该表在本次会话中写入过，从主库读取
func (s *SessionContext) Read(tableName string, fn func(db *gorm.DB) error) error {
	var db *gorm.DB

	// 如果本次会话写入过该表，从主库读取
	if s.writeTables[tableName] {
		db = GetWriteDB().WithContext(s.ctx)
		zlog.GetLogger().Debug("会话一致性：从主库读取", zap.String("table", tableName))
	} else {
		db = GetReadDB().WithContext(s.ctx)
	}

	return fn(db)
}

//因果一致性

// CausalContext 因果一致性上下文
// 通过向量时钟追踪因果关系
type CausalContext struct {
	vectorClock map[string]int64 // 向量时钟
}

// NewCausalContext 创建因果一致性上下文
func NewCausalContext() *CausalContext {
	return &CausalContext{
		vectorClock: make(map[string]int64),
	}
}

// BeforeWrite 写前操作：更新向量时钟
func (c *CausalContext) BeforeWrite(tableName string) {
	c.vectorClock[tableName] = time.Now().UnixNano()
}

// GetReadDBForCausal 根据因果一致性选择读取源
func (c *CausalContext) GetReadDBForCausal(tableName string) *gorm.DB {
	// 如果向量时钟中存在该表的写操作记录，从主库读
	if _, exists := c.vectorClock[tableName]; exists {
		return GetWriteDB()
	}
	return GetReadDB()
}

//监控与调试

// ConsistencyStats 一致性统计
type ConsistencyStats struct {
	ReadFromMasterCount  int64 // 从主库读取次数
	ReadFromReplicaCount int64 // 从从库读取次数
	WriteCount           int64 // 写操作次数
}

var stats ConsistencyStats

// GetConsistencyStats 获取一致性统计
func GetConsistencyStats() ConsistencyStats {
	return stats
}

// RecordReadFromMaster 记录从主库读取
func RecordReadFromMaster() {
	stats.ReadFromMasterCount++
}

// RecordReadFromReplica 记录从从库读取
func RecordReadFromReplica() {
	stats.ReadFromReplicaCount++
}

// RecordWrite 记录写操作
func RecordWrite() {
	stats.WriteCount++
}
