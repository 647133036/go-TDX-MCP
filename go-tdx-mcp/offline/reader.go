package offline

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DayBar represents a daily K-line record from .day file.
type DayBar struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Amount float64 `json:"amount"`
	Vol    float64 `json:"vol"`
}

// MinBar represents a minute K-line record from .lc1/.lc5 file.
type MinBar struct {
	Time   string  `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Amount float64 `json:"amount"`
	Vol    float64 `json:"vol"`
}

// GBBQRecord represents a stock equity change record (gu ben bian qian).
type GBBQRecord struct {
	Market    int     `json:"market"`
	Code      string  `json:"code"`
	Date      string  `json:"date"`
	EventType int     `json:"event_type"`
	Bonus     float64 `json:"bonus"`
	Placement float64 `json:"placement"`
	Transfer  float64 `json:"transfer"`
	Price     float64 `json:"price"`
}

// FinancialRecord represents one period of financial data.
type FinancialRecord struct {
	Code    string             `json:"code"`
	Date    string             `json:"date"`
	Items   map[string]float64 `json:"items"`
}

// BlockInfo represents a block (sector/industry/custom) with stock members.
type BlockInfo struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Members []string `json:"members"`
}

// ReadDaily reads .day file and returns K-line bars.
// Format: 32 bytes per record, uint32 fields: date, open*100, high*100, low*100, close*100, amount, vol, reserved
func ReadDaily(path string) ([]DayBar, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取日线文件失败: %w", err)
	}
	if len(data)%32 != 0 {
		return nil, fmt.Errorf("日线文件格式错误: 大小 %d 不整除32", len(data))
	}
	n := len(data) / 32
	bars := make([]DayBar, n)
	for i := 0; i < n; i++ {
		off := i * 32
		date := binary.LittleEndian.Uint32(data[off : off+4])
		open := float64(binary.LittleEndian.Uint32(data[off+4:off+8])) / 100.0
		high := float64(binary.LittleEndian.Uint32(data[off+8:off+12])) / 100.0
		low := float64(binary.LittleEndian.Uint32(data[off+12:off+16])) / 100.0
		close_ := float64(binary.LittleEndian.Uint32(data[off+16:off+20])) / 100.0
		amount := float64(binary.LittleEndian.Uint32(data[off+20:off+24]))
		vol := float64(binary.LittleEndian.Uint32(data[off+24:off+28]))
		bars[i] = DayBar{
			Date:   formatDate(date),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close_,
			Amount: amount,
			Vol:    vol,
		}
	}
	return bars, nil
}

// ReadMin reads .lc1 (1min) or .lc5 (5min) file and returns minute bars.
// Format: 32 bytes per record, first 4 bytes = YYYYMMDDHHMM, same field layout as .day
func ReadMin(path string) ([]MinBar, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取分钟线文件失败: %w", err)
	}
	if len(data)%32 != 0 {
		return nil, fmt.Errorf("分钟线文件格式错误: 大小 %d 不整除32", len(data))
	}
	n := len(data) / 32
	bars := make([]MinBar, n)
	for i := 0; i < n; i++ {
		off := i * 32
		datetime := binary.LittleEndian.Uint32(data[off : off+4])
		open := float64(binary.LittleEndian.Uint32(data[off+4:off+8])) / 100.0
		high := float64(binary.LittleEndian.Uint32(data[off+8:off+12])) / 100.0
		low := float64(binary.LittleEndian.Uint32(data[off+12:off+16])) / 100.0
		close_ := float64(binary.LittleEndian.Uint32(data[off+16:off+20])) / 100.0
		amount := float64(binary.LittleEndian.Uint32(data[off+20:off+24]))
		vol := float64(binary.LittleEndian.Uint32(data[off+24:off+28]))
		bars[i] = MinBar{
			Time:   formatDateTime(datetime),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close_,
			Amount: amount,
			Vol:    vol,
		}
	}
	return bars, nil
}

// ReadGBBQ reads gbbq.dat (gu ben bian qian) file.
// Format: header(2byte count + 14byte reserved) + records(1byte market + 9byte code + 4byte date + 1byte type + 4byte bonus + 4byte placement + 4byte transfer + 4byte price)
func ReadGBBQ(path string) ([]GBBQRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取股本变迁文件失败: %w", err)
	}
	if len(data) < 16 {
		return nil, fmt.Errorf("股本变迁文件太小: %d 字节", len(data))
	}
	count := int(binary.LittleEndian.Uint16(data[0:2]))
	recordSize := 1 + 9 + 4 + 1 + 4 + 4 + 4 + 4 // 31 bytes
	records := make([]GBBQRecord, 0, count)
	for i := 0; i < count; i++ {
		off := 16 + i*recordSize
		if off+recordSize > len(data) {
			break
		}
		market := int(data[off])
		code := trimNull(data[off+1 : off+10])
		date := binary.LittleEndian.Uint32(data[off+10 : off+14])
		eventType := int(data[off+14])
		bonus := float64(binary.LittleEndian.Uint32(data[off+15:off+19])) / 10.0
		placement := float64(binary.LittleEndian.Uint32(data[off+19:off+23])) / 10.0
		transfer := float64(binary.LittleEndian.Uint32(data[off+23:off+27])) / 10.0
		price := float64(binary.LittleEndian.Uint32(data[off+27:off+31])) / 100.0
		records = append(records, GBBQRecord{
			Market:    market,
			Code:      code,
			Date:      formatDate(date),
			EventType: eventType,
			Bonus:     bonus,
			Placement: placement,
			Transfer:  transfer,
			Price:     price,
		})
	}
	return records, nil
}

// ReadBlocks reads custom block files from blocknew directory.
// Each .dat file contains: header(2byte count) + records(1byte market + 9byte code + 4byte reserved)
func ReadBlocks(blockDir string) ([]BlockInfo, error) {
	entries, err := os.ReadDir(blockDir)
	if err != nil {
		return nil, fmt.Errorf("读取板块目录失败: %w", err)
	}
	var blocks []BlockInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".dat") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		path := filepath.Join(blockDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if len(data) < 2 {
			continue
		}
		count := int(binary.LittleEndian.Uint16(data[0:2]))
		members := make([]string, 0, count)
		recordSize := 1 + 9 + 4 // 14 bytes
		for i := 0; i < count; i++ {
			off := 2 + i*recordSize
			if off+recordSize > len(data) {
				break
			}
			code := trimNull(data[off+1 : off+10])
			market := int(data[off])
			var prefix string
			if market == 0 {
				prefix = "SZ"
			} else {
				prefix = "SH"
			}
			members = append(members, prefix+code)
		}
		sort.Strings(members)
		blocks = append(blocks, BlockInfo{Name: name, Type: "custom", Members: members})
	}
	return blocks, nil
}

// DetectHome detects the TDX installation directory from common paths.
func DetectHome() string {
	candidates := []string{
		`C:\new_jyplug`,
		`C:\new_tdx`,
		`D:\new_jyplug`,
		`D:\new_tdx`,
		`C:\zd_jyplug`,
		`C:\zd_tdx`,
		"./vipdoc",
		"../vipdoc",
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "vipdoc"))
	}
	for _, c := range candidates {
		info, err := os.Stat(c)
		if err == nil && info.IsDir() {
			// Check if candidate itself is the vipdoc directory
			if strings.HasSuffix(c, "vipdoc") || strings.HasSuffix(c, "vipdoc/") || strings.HasSuffix(c, `vipdoc\`) {
				return filepath.Dir(c)
			}
			// Check if candidate has a vipdoc subdirectory
			sub := filepath.Join(c, "vipdoc")
			if fi, err := os.Stat(sub); err == nil && fi.IsDir() {
				return c
			}
		}
	}
	return ""
}

// formatDate converts YYYYMMDD uint32 to "YYYY-MM-DD" string.
func formatDate(d uint32) string {
	year := d / 10000
	month := (d % 10000) / 100
	day := d % 100
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

// formatDateTime converts YYMMDDHHMM uint32 to "20YY-MM-DD HH:MM" string.
func formatDateTime(d uint32) string {
	min := d % 100
	hour := (d / 100) % 100
	day := (d / 10000) % 100
	month := (d / 1000000) % 100
	year := d / 100000000
	if year < 50 {
		year += 2000
	} else {
		year += 1900
	}
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d", year, month, day, hour, min)
}

func trimNull(b []byte) string {
	s := string(b)
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return s[:i]
		}
	}
	return s
}

// SupportedFileTypes lists all supported offline data file extensions.
var SupportedFileTypes = []string{
	".day",    // 日线
	".lc1",    // 1分钟线
	".lc5",    // 5分钟线
	"gbbq",    // 股本变迁
	".dat",    // 板块/财务
}

// CategoryNames maps TDX vipdoc category names.
var CategoryNames = map[string]string{
	"day":   "日线",
	"min":   "分钟线",
	"minline": "分钟线",
}
