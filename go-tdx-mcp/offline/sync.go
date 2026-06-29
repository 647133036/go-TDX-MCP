package offline

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func SyncDaily(vipdoc, market, code string, bars []DayBar) error {
	if strings.ToLower(market) != "sh" && strings.ToLower(market) != "sz" {
		market = strings.ToLower(market)
		if market == "0" {
			market = "sz"
		} else if market == "1" {
			market = "sh"
		}
	}
	market = strings.ToLower(market)
	dir := filepath.Join(vipdoc, market, "day")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	filePath := filepath.Join(dir, fmt.Sprintf("%s%s.day", market, code))
	existing, err := ReadDaily(filePath)
	if err != nil {
		// File doesn't exist, create new
		return WriteDaily(filePath, bars)
	}
	existingDates := make(map[string]bool)
	for _, b := range existing {
		existingDates[b.Date] = true
	}
	var newBars []DayBar
	for _, b := range bars {
		if !existingDates[b.Date] {
			newBars = append(newBars, b)
		}
	}
	if len(newBars) == 0 {
		return nil
	}
	all := append(existing, newBars...)
	sort.Slice(all, func(i, j int) bool { return all[i].Date < all[j].Date })
	return WriteDaily(filePath, all)
}

func WriteDaily(path string, bars []DayBar) error {
	sort.Slice(bars, func(i, j int) bool { return bars[i].Date < bars[j].Date })
	data := make([]byte, len(bars)*32)
	for i, b := range bars {
		off := i * 32
		date := parseDateUint32(b.Date)
		binary.LittleEndian.PutUint32(data[off:off+4], date)
		binary.LittleEndian.PutUint32(data[off+4:off+8], uint32(b.Open*100+0.5))
		binary.LittleEndian.PutUint32(data[off+8:off+12], uint32(b.High*100+0.5))
		binary.LittleEndian.PutUint32(data[off+12:off+16], uint32(b.Low*100+0.5))
		binary.LittleEndian.PutUint32(data[off+16:off+20], uint32(b.Close*100+0.5))
		binary.LittleEndian.PutUint32(data[off+20:off+24], uint32(b.Amount+0.5))
		binary.LittleEndian.PutUint32(data[off+24:off+28], uint32(b.Vol+0.5))
		binary.LittleEndian.PutUint32(data[off+28:off+32], 0)
	}
	return os.WriteFile(path, data, 0644)
}

func parseDateUint32(date string) uint32 {
	s := strings.ReplaceAll(date, "-", "")
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(v)
}
