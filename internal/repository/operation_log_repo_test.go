package repository

import (
	"context"
	"testing"
	"time"

	"go-shop/internal/entity"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type OperationLogRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo OperationLogRepository
}

func (suite *OperationLogRepositoryTestSuite) SetupSuite() {
	// 使用内存SQLite数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// 自动迁移
	err = db.AutoMigrate(&entity.OperationLog{})
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = NewOperationLogRepository(db)
}

func (suite *OperationLogRepositoryTestSuite) TearDownTest() {
	// 清理测试数据
	suite.db.Exec("DELETE FROM operation_logs")
}

func (suite *OperationLogRepositoryTestSuite) TestCreate() {
	ctx := context.Background()

	// 测试创建操作日志
	userID := int64(1)
	userType := entity.UserTypeUser
	status := 200
	duration := int64(100)

	log := &entity.OperationLog{
		UserID:    &userID,
		UserType:  &userType,
		Operation: "login",
		Module:    "auth",
		Method:    "POST",
		Path:      "/api/v2/auth/login",
		IP:        "127.0.0.1",
		Status:    &status,
		Duration:  &duration,
	}

	err := suite.repo.Create(ctx, log)
	suite.NoError(err)
	suite.NotZero(log.ID)

	// 测试nil日志
	err = suite.repo.Create(ctx, nil)
	suite.Error(err)

	// 测试缺少必填字段
	invalidLog := &entity.OperationLog{
		UserID: &userID,
		Module: "auth",
	}
	err = suite.repo.Create(ctx, invalidLog)
	suite.Error(err)
}

func (suite *OperationLogRepositoryTestSuite) TestGetByID() {
	ctx := context.Background()

	// 创建测试日志
	userID := int64(1)
	userType := entity.UserTypeUser
	log := &entity.OperationLog{
		UserID:    &userID,
		UserType:  &userType,
		Operation: "login",
		Module:    "auth",
	}
	err := suite.repo.Create(ctx, log)
	suite.NoError(err)

	// 测试获取存在的日志
	foundLog, err := suite.repo.GetByID(ctx, log.ID)
	suite.NoError(err)
	suite.Equal(log.Operation, foundLog.Operation)
	suite.Equal(log.Module, foundLog.Module)

	// 测试获取不存在的日志
	_, err = suite.repo.GetByID(ctx, 99999)
	suite.Error(err)

	// 测试无效ID
	_, err = suite.repo.GetByID(ctx, 0)
	suite.Error(err)
}

func (suite *OperationLogRepositoryTestSuite) TestList() {
	ctx := context.Background()

	// 创建多个测试日志
	for i := 1; i <= 5; i++ {
		userID := int64(i)
		userType := entity.UserTypeUser
		log := &entity.OperationLog{
			UserID:    &userID,
			UserType:  &userType,
			Operation: "test_operation",
			Module:    "test_module",
		}
		err := suite.repo.Create(ctx, log)
		suite.NoError(err)
	}

	// 测试分页查询
	logs, total, err := suite.repo.List(ctx, 0, 3)
	suite.NoError(err)
	suite.Equal(int64(5), total)
	suite.Len(logs, 3)

	// 测试第二页
	logs, total, err = suite.repo.List(ctx, 3, 3)
	suite.NoError(err)
	suite.Equal(int64(5), total)
	suite.Len(logs, 2)
}

func (suite *OperationLogRepositoryTestSuite) TestListByUser() {
	ctx := context.Background()

	// 创建不同用户的日志
	userID1 := int64(1)
	userID2 := int64(2)
	userType := entity.UserTypeUser

	for i := 0; i < 3; i++ {
		log := &entity.OperationLog{
			UserID:    &userID1,
			UserType:  &userType,
			Operation: "test_operation",
			Module:    "test_module",
		}
		err := suite.repo.Create(ctx, log)
		suite.NoError(err)
	}

	for i := 0; i < 2; i++ {
		log := &entity.OperationLog{
			UserID:    &userID2,
			UserType:  &userType,
			Operation: "test_operation",
			Module:    "test_module",
		}
		err := suite.repo.Create(ctx, log)
		suite.NoError(err)
	}

	// 测试查询用户1的日志
	logs, total, err := suite.repo.ListByUser(ctx, userID1, 0, 10)
	suite.NoError(err)
	suite.Equal(int64(3), total)
	suite.Len(logs, 3)

	// 测试查询用户2的日志
	logs, total, err = suite.repo.ListByUser(ctx, userID2, 0, 10)
	suite.NoError(err)
	suite.Equal(int64(2), total)
	suite.Len(logs, 2)

	// 测试无效用户ID
	_, _, err = suite.repo.ListByUser(ctx, 0, 0, 10)
	suite.Error(err)
}

func (suite *OperationLogRepositoryTestSuite) TestListByModule() {
	ctx := context.Background()

	// 创建不同模块的日志
	userID := int64(1)
	userType := entity.UserTypeUser

	for i := 0; i < 3; i++ {
		log := &entity.OperationLog{
			UserID:    &userID,
			UserType:  &userType,
			Operation: "test_operation",
			Module:    "auth",
		}
		err := suite.repo.Create(ctx, log)
		suite.NoError(err)
	}

	for i := 0; i < 2; i++ {
		log := &entity.OperationLog{
			UserID:    &userID,
			UserType:  &userType,
			Operation: "test_operation",
			Module:    "order",
		}
		err := suite.repo.Create(ctx, log)
		suite.NoError(err)
	}

	// 测试查询auth模块的日志
	logs, total, err := suite.repo.ListByModule(ctx, "auth", 0, 10)
	suite.NoError(err)
	suite.Equal(int64(3), total)
	suite.Len(logs, 3)

	// 测试查询order模块的日志
	logs, total, err = suite.repo.ListByModule(ctx, "order", 0, 10)
	suite.NoError(err)
	suite.Equal(int64(2), total)
	suite.Len(logs, 2)

	// 测试空模块名
	_, _, err = suite.repo.ListByModule(ctx, "", 0, 10)
	suite.Error(err)
}

func (suite *OperationLogRepositoryTestSuite) TestListByOperation() {
	ctx := context.Background()

	// 创建不同操作类型的日志
	userID := int64(1)
	userType := entity.UserTypeUser

	for i := 0; i < 3; i++ {
		log := &entity.OperationLog{
			UserID:    &userID,
			UserType:  &userType,
			Operation: "login",
			Module:    "auth",
		}
		err := suite.repo.Create(ctx, log)
		suite.NoError(err)
	}

	for i := 0; i < 2; i++ {
		log := &entity.OperationLog{
			UserID:    &userID,
			UserType:  &userType,
			Operation: "create_order",
			Module:    "order",
		}
		err := suite.repo.Create(ctx, log)
		suite.NoError(err)
	}

	// 测试查询login操作的日志
	logs, total, err := suite.repo.ListByOperation(ctx, "login", 0, 10)
	suite.NoError(err)
	suite.Equal(int64(3), total)
	suite.Len(logs, 3)

	// 测试查询create_order操作的日志
	logs, total, err = suite.repo.ListByOperation(ctx, "create_order", 0, 10)
	suite.NoError(err)
	suite.Equal(int64(2), total)
	suite.Len(logs, 2)

	// 测试空操作名
	_, _, err = suite.repo.ListByOperation(ctx, "", 0, 10)
	suite.Error(err)
}

func (suite *OperationLogRepositoryTestSuite) TestListByTimeRange() {
	ctx := context.Background()

	// 创建不同时间的日志
	userID := int64(1)
	userType := entity.UserTypeUser
	now := time.Now()

	// 创建昨天的日志
	yesterday := now.AddDate(0, 0, -1)
	log1 := &entity.OperationLog{
		UserID:    &userID,
		UserType:  &userType,
		Operation: "test_operation",
		Module:    "test_module",
		CreatedAt: yesterday,
	}
	suite.db.Create(log1)

	// 创建今天的日志
	log2 := &entity.OperationLog{
		UserID:    &userID,
		UserType:  &userType,
		Operation: "test_operation",
		Module:    "test_module",
		CreatedAt: now,
	}
	suite.db.Create(log2)

	// 测试查询今天的日志
	startTime := now.Truncate(24 * time.Hour)
	endTime := now.Add(24 * time.Hour)
	logs, total, err := suite.repo.ListByTimeRange(ctx, startTime, endTime, 0, 10)
	suite.NoError(err)
	suite.Equal(int64(1), total)
	suite.Len(logs, 1)

	// 测试查询所有日志
	startTime = yesterday.Truncate(24 * time.Hour)
	logs, total, err = suite.repo.ListByTimeRange(ctx, startTime, endTime, 0, 10)
	suite.NoError(err)
	suite.Equal(int64(2), total)
	suite.Len(logs, 2)

	// 测试无效时间范围
	_, _, err = suite.repo.ListByTimeRange(ctx, endTime, startTime, 0, 10)
	suite.Error(err)
}

func (suite *OperationLogRepositoryTestSuite) TestListWithFilters() {
	ctx := context.Background()

	// 创建测试数据
	userID1 := int64(1)
	userID2 := int64(2)
	userType := entity.UserTypeUser
	adminType := entity.UserTypeAdmin

	// 用户1的auth模块日志
	log1 := &entity.OperationLog{
		UserID:    &userID1,
		UserType:  &userType,
		Operation: "login",
		Module:    "auth",
	}
	suite.repo.Create(ctx, log1)

	// 用户2的order模块日志
	log2 := &entity.OperationLog{
		UserID:    &userID2,
		UserType:  &userType,
		Operation: "create_order",
		Module:    "order",
	}
	suite.repo.Create(ctx, log2)

	// 管理员的product模块日志
	log3 := &entity.OperationLog{
		UserID:    &userID1,
		UserType:  &adminType,
		Operation: "create_product",
		Module:    "product",
	}
	suite.repo.Create(ctx, log3)

	// 测试按用户ID筛选
	logs, total, err := suite.repo.ListWithFilters(ctx, &userID1, nil, nil, nil, nil, nil, 0, 10)
	suite.NoError(err)
	suite.Equal(int64(2), total)
	suite.Len(logs, 2)

	// 测试按用户类型筛选
	logs, total, err = suite.repo.ListWithFilters(ctx, nil, &adminType, nil, nil, nil, nil, 0, 10)
	suite.NoError(err)
	suite.Equal(int64(1), total)
	suite.Len(logs, 1)

	// 测试按模块筛选
	module := "auth"
	logs, total, err = suite.repo.ListWithFilters(ctx, nil, nil, &module, nil, nil, nil, 0, 10)
	suite.NoError(err)
	suite.Equal(int64(1), total)
	suite.Len(logs, 1)

	// 测试组合筛选
	logs, total, err = suite.repo.ListWithFilters(ctx, &userID1, &userType, nil, nil, nil, nil, 0, 10)
	suite.NoError(err)
	suite.Equal(int64(1), total)
	suite.Len(logs, 1)
}

func (suite *OperationLogRepositoryTestSuite) TestDelete() {
	ctx := context.Background()

	// 创建测试日志
	userID := int64(1)
	userType := entity.UserTypeUser
	log := &entity.OperationLog{
		UserID:    &userID,
		UserType:  &userType,
		Operation: "test_operation",
		Module:    "test_module",
	}
	err := suite.repo.Create(ctx, log)
	suite.NoError(err)

	// 测试删除日志
	err = suite.repo.Delete(ctx, log.ID)
	suite.NoError(err)

	// 验证日志已删除
	_, err = suite.repo.GetByID(ctx, log.ID)
	suite.Error(err)

	// 测试删除不存在的日志
	err = suite.repo.Delete(ctx, 99999)
	suite.Error(err)

	// 测试无效ID
	err = suite.repo.Delete(ctx, 0)
	suite.Error(err)
}

func (suite *OperationLogRepositoryTestSuite) TestDeleteByUser() {
	ctx := context.Background()

	// 创建多个用户的日志
	userID1 := int64(1)
	userID2 := int64(2)
	userType := entity.UserTypeUser

	for i := 0; i < 3; i++ {
		log := &entity.OperationLog{
			UserID:    &userID1,
			UserType:  &userType,
			Operation: "test_operation",
			Module:    "test_module",
		}
		suite.repo.Create(ctx, log)
	}

	for i := 0; i < 2; i++ {
		log := &entity.OperationLog{
			UserID:    &userID2,
			UserType:  &userType,
			Operation: "test_operation",
			Module:    "test_module",
		}
		suite.repo.Create(ctx, log)
	}

	// 删除用户1的所有日志
	err := suite.repo.DeleteByUser(ctx, userID1)
	suite.NoError(err)

	// 验证用户1的日志已删除
	logs, total, err := suite.repo.ListByUser(ctx, userID1, 0, 10)
	suite.NoError(err)
	suite.Equal(int64(0), total)
	suite.Len(logs, 0)

	// 验证用户2的日志仍然存在
	logs, total, err = suite.repo.ListByUser(ctx, userID2, 0, 10)
	suite.NoError(err)
	suite.Equal(int64(2), total)
	suite.Len(logs, 2)

	// 测试无效用户ID
	err = suite.repo.DeleteByUser(ctx, 0)
	suite.Error(err)
}

func (suite *OperationLogRepositoryTestSuite) TestDeleteOldLogs() {
	ctx := context.Background()

	// 创建不同时间的日志
	userID := int64(1)
	userType := entity.UserTypeUser
	now := time.Now()

	// 创建100天前的日志
	oldLog := &entity.OperationLog{
		UserID:    &userID,
		UserType:  &userType,
		Operation: "old_operation",
		Module:    "test_module",
		CreatedAt: now.AddDate(0, 0, -100),
	}
	suite.db.Create(oldLog)

	// 创建今天的日志
	newLog := &entity.OperationLog{
		UserID:    &userID,
		UserType:  &userType,
		Operation: "new_operation",
		Module:    "test_module",
		CreatedAt: now,
	}
	suite.db.Create(newLog)

	// 删除90天前的日志
	err := suite.repo.DeleteOldLogs(ctx, 90)
	suite.NoError(err)

	// 验证旧日志已删除
	_, err = suite.repo.GetByID(ctx, oldLog.ID)
	suite.Error(err)

	// 验证新日志仍然存在
	_, err = suite.repo.GetByID(ctx, newLog.ID)
	suite.NoError(err)

	// 测试无效天数
	err = suite.repo.DeleteOldLogs(ctx, 0)
	suite.Error(err)
}

func (suite *OperationLogRepositoryTestSuite) TestBatchCreate() {
	ctx := context.Background()

	// 创建批量日志
	userID := int64(1)
	userType := entity.UserTypeUser
	logs := []*entity.OperationLog{
		{
			UserID:    &userID,
			UserType:  &userType,
			Operation: "operation1",
			Module:    "module1",
		},
		{
			UserID:    &userID,
			UserType:  &userType,
			Operation: "operation2",
			Module:    "module2",
		},
		{
			UserID:    &userID,
			UserType:  &userType,
			Operation: "operation3",
			Module:    "module3",
		},
	}

	err := suite.repo.BatchCreate(ctx, logs)
	suite.NoError(err)

	// 验证所有日志都已创建
	for _, log := range logs {
		suite.NotZero(log.ID)
	}

	// 测试空列表
	err = suite.repo.BatchCreate(ctx, []*entity.OperationLog{})
	suite.NoError(err)

	// 测试包含nil的列表
	invalidLogs := []*entity.OperationLog{nil}
	err = suite.repo.BatchCreate(ctx, invalidLogs)
	suite.Error(err)
}

func (suite *OperationLogRepositoryTestSuite) TestCountByModule() {
	ctx := context.Background()

	// 创建不同模块的日志
	userID := int64(1)
	userType := entity.UserTypeUser

	for i := 0; i < 3; i++ {
		log := &entity.OperationLog{
			UserID:    &userID,
			UserType:  &userType,
			Operation: "test_operation",
			Module:    "auth",
		}
		suite.repo.Create(ctx, log)
	}

	for i := 0; i < 2; i++ {
		log := &entity.OperationLog{
			UserID:    &userID,
			UserType:  &userType,
			Operation: "test_operation",
			Module:    "order",
		}
		suite.repo.Create(ctx, log)
	}

	// 测试统计auth模块
	count, err := suite.repo.CountByModule(ctx, "auth")
	suite.NoError(err)
	suite.Equal(int64(3), count)

	// 测试统计order模块
	count, err = suite.repo.CountByModule(ctx, "order")
	suite.NoError(err)
	suite.Equal(int64(2), count)

	// 测试空模块名
	_, err = suite.repo.CountByModule(ctx, "")
	suite.Error(err)
}

func (suite *OperationLogRepositoryTestSuite) TestCountByOperation() {
	ctx := context.Background()

	// 创建不同操作类型的日志
	userID := int64(1)
	userType := entity.UserTypeUser

	for i := 0; i < 3; i++ {
		log := &entity.OperationLog{
			UserID:    &userID,
			UserType:  &userType,
			Operation: "login",
			Module:    "auth",
		}
		suite.repo.Create(ctx, log)
	}

	for i := 0; i < 2; i++ {
		log := &entity.OperationLog{
			UserID:    &userID,
			UserType:  &userType,
			Operation: "create_order",
			Module:    "order",
		}
		suite.repo.Create(ctx, log)
	}

	// 测试统计login操作
	count, err := suite.repo.CountByOperation(ctx, "login")
	suite.NoError(err)
	suite.Equal(int64(3), count)

	// 测试统计create_order操作
	count, err = suite.repo.CountByOperation(ctx, "create_order")
	suite.NoError(err)
	suite.Equal(int64(2), count)

	// 测试空操作名
	_, err = suite.repo.CountByOperation(ctx, "")
	suite.Error(err)
}

func (suite *OperationLogRepositoryTestSuite) TestCountErrorLogs() {
	ctx := context.Background()

	// 创建带错误的日志
	userID := int64(1)
	userType := entity.UserTypeUser
	now := time.Now()

	for i := 0; i < 3; i++ {
		log := &entity.OperationLog{
			UserID:    &userID,
			UserType:  &userType,
			Operation: "test_operation",
			Module:    "test_module",
			Error:     "test error",
			CreatedAt: now,
		}
		suite.db.Create(log)
	}

	// 创建没有错误的日志
	for i := 0; i < 2; i++ {
		log := &entity.OperationLog{
			UserID:    &userID,
			UserType:  &userType,
			Operation: "test_operation",
			Module:    "test_module",
			CreatedAt: now,
		}
		suite.db.Create(log)
	}

	// 测试统计错误日志
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(1 * time.Hour)
	count, err := suite.repo.CountErrorLogs(ctx, startTime, endTime)
	suite.NoError(err)
	suite.Equal(int64(3), count)

	// 测试无效时间范围
	_, err = suite.repo.CountErrorLogs(ctx, endTime, startTime)
	suite.Error(err)
}

func TestOperationLogRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(OperationLogRepositoryTestSuite))
}
