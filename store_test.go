package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupTestStore(t *testing.T) (*Store, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "liuxia-test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}

	dbPath := filepath.Join(dir, "test.db")
	store, err := InitStore(dbPath)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("InitStore() 失败: %v", err)
	}

	return store, func() {
		store.Close()
		os.RemoveAll(dir)
	}
}

func TestInitStore(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	if store == nil {
		t.Fatal("InitStore() 返回 nil")
	}
	if store.db == nil {
		t.Error("InitStore() db 应已初始化")
	}
	if store.Cache == nil {
		t.Error("InitStore() Cache 应已初始化")
	}
}

func TestInitStoreInvalidPath(t *testing.T) {
	// SQLite 允许使用空字符串或 ":memory:" 作为内存数据库
	// 测试无效目录路径
	_, err := InitStore("/nonexistent/path/test.db")
	if err == nil {
		t.Error("InitStore() 在无效路径时应该返回错误")
	}
}

func TestUpsertRecord(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	quality := 0.85
	aod := 0.3

	record := SunsetRecord{
		City:      "北京",
		Date:      "2024-01-01",
		Time:      "18:30",
		EventType: "evening",
		Model:     "GFS",
		Quality:   &quality,
		AOD:       &aod,
	}

	err := store.UpsertRecord(record)
	if err != nil {
		t.Fatalf("UpsertRecord() 失败: %v", err)
	}

	// 验证记录已插入
	records, err := store.QueryRecords("北京", "evening", "2024-01-01", "2024-01-01")
	if err != nil {
		t.Fatalf("QueryRecords() 失败: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("QueryRecords() 返回 %d 条记录, want 1", len(records))
	}
	if records[0].City != "北京" {
		t.Errorf("records[0].City = %q, want %q", records[0].City, "北京")
	}
}

func TestUpsertRecordUpdate(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	quality1 := 0.85
	quality2 := 0.95

	record1 := SunsetRecord{
		City:      "北京",
		Date:      "2024-01-01",
		Time:      "18:30",
		EventType: "evening",
		Model:     "GFS",
		Quality:   &quality1,
	}

	record2 := SunsetRecord{
		City:      "北京",
		Date:      "2024-01-01",
		Time:      "18:45",
		EventType: "evening",
		Model:     "GFS",
		Quality:   &quality2,
	}

	store.UpsertRecord(record1)
	store.UpsertRecord(record2)

	records, _ := store.QueryRecords("北京", "evening", "2024-01-01", "2024-01-01")
	if len(records) != 1 {
		t.Fatalf("QueryRecords() 返回 %d 条记录, want 1", len(records))
	}
	if records[0].Time != "18:45" {
		t.Errorf("records[0].Time = %q, want %q (应该被更新)", records[0].Time, "18:45")
	}
}

func TestQueryRecords(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	quality := 0.85
	records := []SunsetRecord{
		{City: "北京", Date: "2024-01-01", Time: "18:30", EventType: "evening", Model: "GFS", Quality: &quality},
		{City: "上海", Date: "2024-01-01", Time: "18:35", EventType: "evening", Model: "EC", Quality: &quality},
		{City: "北京", Date: "2024-01-02", Time: "18:30", EventType: "evening", Model: "GFS", Quality: &quality},
	}

	for _, r := range records {
		store.UpsertRecord(r)
	}

	t.Run("按城市查询", func(t *testing.T) {
		result, err := store.QueryRecords("北京", "", "", "")
		if err != nil {
			t.Fatalf("QueryRecords() 失败: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("QueryRecords() 返回 %d 条记录, want 2", len(result))
		}
	})

	t.Run("按事件类型查询", func(t *testing.T) {
		result, err := store.QueryRecords("", "evening", "", "")
		if err != nil {
			t.Fatalf("QueryRecords() 失败: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("QueryRecords() 返回 %d 条记录, want 3", len(result))
		}
	})

	t.Run("按日期范围查询", func(t *testing.T) {
		result, err := store.QueryRecords("", "", "2024-01-02", "2024-01-02")
		if err != nil {
			t.Fatalf("QueryRecords() 失败: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("QueryRecords() 返回 %d 条记录, want 1", len(result))
		}
	})
}

func TestExportCSV(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	quality := 0.85
	store.UpsertRecord(SunsetRecord{
		City:      "北京",
		Date:      "2024-01-01",
		Time:      "18:30",
		EventType: "evening",
		Model:     "GFS",
		Quality:   &quality,
	})

	var buf bytes.Buffer
	err := store.ExportCSV(&buf, "", "", "", "")
	if err != nil {
		t.Fatalf("ExportCSV() 失败: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("ExportCSV() 输出为空")
	}
	// CSV 应该包含表头和数据行
	lines := splitLines(output)
	if len(lines) < 2 {
		t.Errorf("ExportCSV() 应该包含表头和至少一行数据, got %d 行", len(lines))
	}
}

func TestExportJSON(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	quality := 0.85
	store.UpsertRecord(SunsetRecord{
		City:      "北京",
		Date:      "2024-01-01",
		Time:      "18:30",
		EventType: "evening",
		Model:     "GFS",
		Quality:   &quality,
	})

	var buf bytes.Buffer
	err := store.ExportJSON(&buf, "", "", "", "")
	if err != nil {
		t.Fatalf("ExportJSON() 失败: %v", err)
	}

	var records []SunsetRecord
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("ExportJSON() 输出不是有效的 JSON: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("ExportJSON() 返回 %d 条记录, want 1", len(records))
	}
}

func TestGetCities(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	quality := 0.85
	store.UpsertRecord(SunsetRecord{City: "北京", Date: "2024-01-01", Time: "18:30", EventType: "evening", Model: "GFS", Quality: &quality})
	store.UpsertRecord(SunsetRecord{City: "上海", Date: "2024-01-01", Time: "18:35", EventType: "evening", Model: "EC", Quality: &quality})
	store.UpsertRecord(SunsetRecord{City: "北京", Date: "2024-01-02", Time: "18:30", EventType: "evening", Model: "GFS", Quality: &quality})

	cities, err := store.GetCities()
	if err != nil {
		t.Fatalf("GetCities() 失败: %v", err)
	}
	if len(cities) != 2 {
		t.Errorf("GetCities() 返回 %d 个城市, want 2", len(cities))
	}
}

func TestGetTotalRecords(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	quality := 0.85
	store.UpsertRecord(SunsetRecord{City: "北京", Date: "2024-01-01", Time: "18:30", EventType: "evening", Model: "GFS", Quality: &quality})
	store.UpsertRecord(SunsetRecord{City: "上海", Date: "2024-01-01", Time: "18:35", EventType: "evening", Model: "EC", Quality: &quality})

	count, err := store.GetTotalRecords()
	if err != nil {
		t.Fatalf("GetTotalRecords() 失败: %v", err)
	}
	if count != 2 {
		t.Errorf("GetTotalRecords() = %d, want 2", count)
	}
}

func TestDeleteOldRecords(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	quality := 0.85
	// 使用当前日期附近的日期来测试
	// 一条记录是很久以前的（会被删除），一条是最近的（不会被删除）
	store.UpsertRecord(SunsetRecord{City: "北京", Date: "2020-01-01", Time: "18:30", EventType: "evening", Model: "GFS", Quality: &quality})
	store.UpsertRecord(SunsetRecord{City: "北京", Date: "2099-12-31", Time: "18:30", EventType: "evening", Model: "EC", Quality: &quality})

	deleted, err := store.DeleteOldRecords(365)
	if err != nil {
		t.Fatalf("DeleteOldRecords() 失败: %v", err)
	}
	if deleted != 1 {
		t.Errorf("DeleteOldRecords() 删除了 %d 条记录, want 1", deleted)
	}

	count, _ := store.GetTotalRecords()
	if count != 1 {
		t.Errorf("DeleteOldRecords() 后剩余 %d 条记录, want 1", count)
	}
}

func TestDeleteAbnormalDates(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	quality := 0.85
	store.UpsertRecord(SunsetRecord{City: "北京", Date: "2019-01-01", Time: "18:30", EventType: "evening", Model: "GFS", Quality: &quality})
	store.UpsertRecord(SunsetRecord{City: "北京", Date: "2024-01-01", Time: "18:30", EventType: "evening", Model: "GFS", Quality: &quality})

	deleted, err := store.DeleteAbnormalDates()
	if err != nil {
		t.Fatalf("DeleteAbnormalDates() 失败: %v", err)
	}
	if deleted != 1 {
		t.Errorf("DeleteAbnormalDates() 删除了 %d 条记录, want 1", deleted)
	}

	count, _ := store.GetTotalRecords()
	if count != 1 {
		t.Errorf("DeleteAbnormalDates() 后剩余 %d 条记录, want 1", count)
	}
}

func TestGetStatistics(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	quality1 := 0.85
	quality2 := 0.95
	store.UpsertRecord(SunsetRecord{City: "北京", Date: "2024-01-01", Time: "18:30", EventType: "evening", Model: "GFS", Quality: &quality1})
	store.UpsertRecord(SunsetRecord{City: "北京", Date: "2024-01-02", Time: "18:30", EventType: "evening", Model: "EC", Quality: &quality2})

	stats, err := store.GetStatistics("", "", "", "")
	if err != nil {
		t.Fatalf("GetStatistics() 失败: %v", err)
	}
	if stats.TotalRecords != 2 {
		t.Errorf("GetStatistics().TotalRecords = %d, want 2", stats.TotalRecords)
	}
	if len(stats.Models) != 2 {
		t.Errorf("GetStatistics().Models 长度 = %d, want 2", len(stats.Models))
	}
}

func TestGetCityComparison(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	quality1 := 0.85
	quality2 := 0.95
	store.UpsertRecord(SunsetRecord{City: "北京", Date: "2024-01-01", Time: "18:30", EventType: "evening", Model: "GFS", Quality: &quality1})
	store.UpsertRecord(SunsetRecord{City: "上海", Date: "2024-01-01", Time: "18:35", EventType: "evening", Model: "EC", Quality: &quality2})

	result, err := store.GetCityComparison("", "", "")
	if err != nil {
		t.Fatalf("GetCityComparison() 失败: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("GetCityComparison() 返回 %d 个城市, want 2", len(result))
	}
}

func TestGetTodayTomorrowData(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	today := "2024-01-15"
	quality := 0.85
	store.UpsertRecord(SunsetRecord{City: "北京", Date: today, Time: "18:30", EventType: "evening", Model: "GFS", Quality: &quality})

	// 注意：GetTodayTomorrowData 使用当前日期，这里我们只验证基本功能
	result, err := store.GetTodayTomorrowData("北京")
	if err != nil {
		t.Fatalf("GetTodayTomorrowData() 失败: %v", err)
	}
	// 结果取决于当前日期是否匹配测试数据
	_ = result
}

func TestGetRankings(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	quality1 := 0.85
	quality2 := 0.95
	store.UpsertRecord(SunsetRecord{City: "北京", Date: "2024-01-01", Time: "18:30", EventType: "evening", Model: "GFS", Quality: &quality1})
	store.UpsertRecord(SunsetRecord{City: "上海", Date: "2024-01-01", Time: "18:35", EventType: "evening", Model: "EC", Quality: &quality2})

	rankings, err := store.GetRankings("", "", 10)
	if err != nil {
		t.Fatalf("GetRankings() 失败: %v", err)
	}
	if len(rankings.BestDates) != 2 {
		t.Errorf("GetRankings().BestDates 长度 = %d, want 2", len(rankings.BestDates))
	}
}

func TestClose(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	err := store.Close()
	if err != nil {
		t.Fatalf("Close() 失败: %v", err)
	}
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range []byte(s) {
		if line == '\n' {
			lines = append(lines, "")
		} else {
			if len(lines) == 0 {
				lines = append(lines, "")
			}
			lines[len(lines)-1] += string(line)
		}
	}
	return lines
}
