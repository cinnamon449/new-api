package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedFlowQuotaData(t *testing.T, quotaData QuotaData) {
	t.Helper()
	require.NoError(t, DB.Create(&quotaData).Error)
}

func TestQuotaDataCacheCountersDatabaseMatrix(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql", "postgres"} {
		t.Run(dialect, func(t *testing.T) {
			dsn := "local"
			if dialect == "mysql" {
				dsn = os.Getenv("TEST_MYSQL_DSN")
			} else if dialect == "postgres" {
				dsn = os.Getenv("TEST_POSTGRES_DSN")
			}
			if dsn == "" {
				t.Skip("test database DSN is not configured")
			}
			previousDB, previousType, previousPath := DB, common.MainDatabaseType(), common.SQLitePath
			previousCache := CacheQuotaData
			common.SQLitePath = filepath.Join(t.TempDir(), "quota.db")
			t.Setenv("QUOTA_COUNTER_TEST_DSN", dsn)
			db, dbType, err := chooseDB("QUOTA_COUNTER_TEST_DSN", false)
			require.NoError(t, err)
			sqlDB, err := db.DB()
			require.NoError(t, err)
			DB = db
			common.SetMainDatabaseType(dbType)
			initCol()
			CacheQuotaData = make(map[string]*QuotaData)
			t.Cleanup(func() {
				assert.NoError(t, db.Migrator().DropTable(&QuotaData{}))
				assert.NoError(t, sqlDB.Close())
				DB, common.SQLitePath, CacheQuotaData = previousDB, previousPath, previousCache
				common.SetMainDatabaseType(previousType)
				initCol()
			})
			require.NoError(t, db.AutoMigrate(&QuotaData{}))
			// Merge two requests in memory, then update the persisted hourly row.
			for i, quantity := range []struct{ read, write int64 }{{8, 3}, {4, 2}, {-3, -5}, {2, 1}} {
				LogQuotaData(QuotaDataLogParams{
					UserID: 1, Username: "sync-qa", ModelName: "sync-qa-model",
					CreatedAt: 3661, UseGroup: "default", TokenID: 1, ChannelID: 1,
					Quota: 5, TokenUsed: 10, CacheReadTokens: quantity.read, CacheCreationTokens: quantity.write,
				})
				if i == 1 || i == 3 {
					SaveQuotaDataCache()
				}
			}
			var stored []QuotaData
			require.NoError(t, db.Find(&stored).Error)
			require.Len(t, stored, 1)
			assert.EqualValues(t, 14, stored[0].CacheReadTokens)
			assert.EqualValues(t, 6, stored[0].CacheCreationTokens)
			assert.Equal(t, 4, stored[0].Count)
			assert.Equal(t, 20, stored[0].Quota)
			assert.Equal(t, 40, stored[0].TokenUsed)
			assert.EqualValues(t, 3600, stored[0].CreatedAt)
			queries := map[string]func() ([]*QuotaData, error){
				"username":   func() ([]*QuotaData, error) { return GetQuotaDataByUsername("sync-qa", 0, 7200) },
				"user_id":    func() ([]*QuotaData, error) { return GetQuotaDataByUserId(1, 0, 7200) },
				"by_user":    func() ([]*QuotaData, error) { return GetQuotaDataGroupByUser(0, 7200) },
				"all_models": func() ([]*QuotaData, error) { return GetAllQuotaDates(0, 7200, "") },
			}
			for name, query := range queries {
				t.Run(name, func(t *testing.T) {
					rows, err := query()
					require.NoError(t, err)
					require.Len(t, rows, 1)
					assert.EqualValues(t, 14, rows[0].CacheReadTokens)
					assert.EqualValues(t, 6, rows[0].CacheCreationTokens)
					assert.Equal(t, 4, rows[0].Count)
					assert.Equal(t, 20, rows[0].Quota)
					assert.Equal(t, 40, rows[0].TokenUsed)
				})
			}
		})
	}
}

func seedFlowLookupData(t *testing.T) {
	t.Helper()
	require.NoError(t, DB.Create(&Channel{Id: 1, Name: "east"}).Error)
	require.NoError(t, DB.Create(&Channel{Id: 2, Name: "west"}).Error)
	require.NoError(t, DB.Create(&Token{Id: 11, UserId: 1, Key: "sk-primary", Name: "primary"}).Error)
	require.NoError(t, DB.Create(&Token{Id: 22, UserId: 2, Key: "sk-backup", Name: "backup"}).Error)
	require.NoError(t, DB.Delete(&Token{Id: 11}).Error)
}

func TestGetFlowQuotaDataUsesQuotaDataRoleSpecificDimensions(t *testing.T) {
	truncateTables(t)
	seedFlowLookupData(t)

	seedFlowQuotaData(t, QuotaData{
		UserID:    1,
		Username:  "alice",
		NodeName:  "node-a",
		TokenID:   11,
		UseGroup:  "vip",
		ModelName: "gpt-a",
		ChannelID: 1,
		CreatedAt: 1000,
		Count:     2,
		Quota:     100,
		TokenUsed: 40,
	})
	seedFlowQuotaData(t, QuotaData{
		UserID:    1,
		Username:  "alice",
		NodeName:  "node-a",
		TokenID:   11,
		UseGroup:  "vip",
		ModelName: "gpt-a",
		ChannelID: 1,
		CreatedAt: 1100,
		Count:     1,
		Quota:     50,
		TokenUsed: 20,
	})
	seedFlowQuotaData(t, QuotaData{
		UserID:    1,
		Username:  "alice",
		NodeName:  "node-a",
		TokenID:   11,
		UseGroup:  "vip",
		ModelName: "gpt-a",
		ChannelID: 2,
		CreatedAt: 1200,
		Count:     1,
		Quota:     25,
		TokenUsed: 10,
	})
	seedFlowQuotaData(t, QuotaData{
		UserID:    2,
		Username:  "bob",
		NodeName:  "node-b",
		TokenID:   22,
		UseGroup:  "default",
		ModelName: "gpt-b",
		ChannelID: 1,
		CreatedAt: 1300,
		Count:     3,
		Quota:     70,
		TokenUsed: 30,
	})
	seedFlowQuotaData(t, QuotaData{
		UserID:    1,
		Username:  "alice",
		ModelName: "legacy",
		CreatedAt: 1400,
		Count:     99,
		Quota:     999,
		TokenUsed: 999,
	})

	rootRows, err := GetFlowQuotaData(900, 2000, "", 0, common.RoleRootUser)
	require.NoError(t, err)
	require.Len(t, rootRows, 3)
	// Token 11 was soft-deleted, so its name is intentionally left empty for the
	// frontend to render a localized "deleted (id)" label instead.
	require.Equal(t, FlowQuotaData{
		UserID:      1,
		Username:    "alice",
		NodeName:    "node-a",
		TokenID:     11,
		TokenName:   "",
		UseGroup:    "vip",
		ChannelID:   1,
		ChannelName: "east",
		ModelName:   "gpt-a",
		TokenUsed:   60,
		Count:       3,
		Quota:       150,
	}, *rootRows[0])
	// A token that still exists resolves to its current name.
	require.Equal(t, 22, rootRows[1].TokenID)
	require.Equal(t, "backup", rootRows[1].TokenName)

	adminRows, err := GetFlowQuotaData(900, 2000, "alice", 0, common.RoleAdminUser)
	require.NoError(t, err)
	require.Len(t, adminRows, 2)
	require.Equal(t, 0, adminRows[0].TokenID)
	require.Empty(t, adminRows[0].TokenName)
	require.Empty(t, adminRows[0].NodeName)
	require.Equal(t, "alice", adminRows[0].Username)
	require.Equal(t, "vip", adminRows[0].UseGroup)
	require.Equal(t, "east", adminRows[0].ChannelName)
	require.Equal(t, 150, adminRows[0].Quota)

	selfRows, err := GetFlowQuotaData(900, 2000, "", 1, common.RoleCommonUser)
	require.NoError(t, err)
	require.Len(t, selfRows, 1)
	require.Empty(t, selfRows[0].Username)
	require.Equal(t, 0, selfRows[0].ChannelID)
	require.Empty(t, selfRows[0].ChannelName)
	require.Empty(t, selfRows[0].TokenName)
	require.Equal(t, "vip", selfRows[0].UseGroup)
	require.Equal(t, 175, selfRows[0].Quota)
}

func TestLogQuotaDataSplitsRowsByUseGroupTokenChannelAndNode(t *testing.T) {
	truncateTables(t)
	CacheQuotaDataLock.Lock()
	CacheQuotaData = make(map[string]*QuotaData)
	CacheQuotaDataLock.Unlock()

	LogQuotaData(QuotaDataLogParams{
		UserID:    1,
		Username:  "alice",
		ModelName: "gpt-a",
		CreatedAt: 3661,
		UseGroup:  "vip",
		TokenID:   11,
		ChannelID: 1,
		NodeName:  "node-a",
		Quota:     100,
		TokenUsed: 40,
	})
	LogQuotaData(QuotaDataLogParams{
		UserID:    1,
		Username:  "alice",
		ModelName: "gpt-a",
		CreatedAt: 3700,
		UseGroup:  "vip",
		TokenID:   11,
		ChannelID: 1,
		NodeName:  "node-a",
		Quota:     50,
		TokenUsed: 20,
	})
	LogQuotaData(QuotaDataLogParams{
		UserID:    1,
		Username:  "alice",
		ModelName: "gpt-a",
		CreatedAt: 3700,
		UseGroup:  "default",
		TokenID:   11,
		ChannelID: 1,
		NodeName:  "node-a",
		Quota:     25,
		TokenUsed: 10,
	})

	SaveQuotaDataCache()

	var rows []QuotaData
	require.NoError(t, DB.Order("quota DESC").Find(&rows).Error)
	require.Len(t, rows, 2)
	require.Equal(t, int64(3600), rows[0].CreatedAt)
	require.Equal(t, "vip", rows[0].UseGroup)
	require.Equal(t, 11, rows[0].TokenID)
	require.Equal(t, 1, rows[0].ChannelID)
	require.Equal(t, "node-a", rows[0].NodeName)
	require.Equal(t, 2, rows[0].Count)
	require.Equal(t, 150, rows[0].Quota)
	require.Equal(t, 60, rows[0].TokenUsed)
	require.Equal(t, "default", rows[1].UseGroup)
	require.Equal(t, 25, rows[1].Quota)
}
