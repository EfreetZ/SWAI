package db

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/ds"
)

var ErrInvalidCommand = errors.New("invalid command")

// ExecuteCommand 执行命令并返回字符串结果。
func (d *DB) ExecuteCommand(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "", ErrInvalidCommand
	}
	cmd := strings.ToUpper(args[0])

	switch cmd {
	case "PING":
		return "PONG", nil
	case "ECHO":
		if len(args) < 2 {
			return "", ErrInvalidCommand
		}
		return args[1], nil
	case "SET":
		if len(args) < 3 {
			return "", ErrInvalidCommand
		}
		ttl := time.Duration(0)
		if len(args) >= 5 && strings.EqualFold(args[3], "EX") {
			seconds, err := strconv.Atoi(args[4])
			if err != nil {
				return "", err
			}
			ttl = time.Duration(seconds) * time.Second
		}
		if err := d.SetString(ctx, args[1], args[2], ttl); err != nil {
			return "", err
		}
		return "OK", nil
	case "GET":
		if len(args) < 2 {
			return "", ErrInvalidCommand
		}
		value, ok, err := d.GetString(ctx, args[1])
		if err != nil {
			return "", err
		}
		if !ok {
			return "(nil)", nil
		}
		return value, nil
	case "DEL":
		if len(args) < 2 {
			return "", ErrInvalidCommand
		}
		deleted, err := d.Del(ctx, args[1])
		if err != nil {
			return "", err
		}
		if deleted {
			return "1", nil
		}
		return "0", nil
	case "EXISTS":
		if len(args) < 2 {
			return "", ErrInvalidCommand
		}
		exists, err := d.Exists(ctx, args[1])
		if err != nil {
			return "", err
		}
		if exists {
			return "1", nil
		}
		return "0", nil
	case "KEYS":
		pattern := "*"
		if len(args) >= 2 {
			pattern = args[1]
		}
		keys, err := d.Keys(ctx, pattern)
		if err != nil {
			return "", err
		}
		return strings.Join(keys, ","), nil
	case "EXPIRE":
		if len(args) < 3 {
			return "", ErrInvalidCommand
		}
		seconds, err := strconv.Atoi(args[2])
		if err != nil {
			return "", err
		}
		ok, err := d.Expire(ctx, args[1], time.Duration(seconds)*time.Second)
		if err != nil {
			return "", err
		}
		if ok {
			return "1", nil
		}
		return "0", nil
	case "TTL":
		if len(args) < 2 {
			return "", ErrInvalidCommand
		}
		ttl, err := d.TTL(ctx, args[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", ttl), nil
	case "LPUSH":
		if len(args) < 3 {
			return "", ErrInvalidCommand
		}
		length, err := d.LPush(ctx, args[1], args[2])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", length), nil
	case "RPUSH":
		if len(args) < 3 {
			return "", ErrInvalidCommand
		}
		length, err := d.RPush(ctx, args[1], args[2])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", length), nil
	case "LPOP":
		if len(args) < 2 {
			return "", ErrInvalidCommand
		}
		value, ok, err := d.LPop(ctx, args[1])
		if err != nil {
			return "", err
		}
		if !ok {
			return "(nil)", nil
		}
		return value, nil
	case "LRANGE":
		if len(args) < 4 {
			return "", ErrInvalidCommand
		}
		start, err := strconv.Atoi(args[2])
		if err != nil {
			return "", err
		}
		stop, err := strconv.Atoi(args[3])
		if err != nil {
			return "", err
		}
		items, err := d.LRange(ctx, args[1], start, stop)
		if err != nil {
			return "", err
		}
		return strings.Join(items, ","), nil
	case "SADD":
		if len(args) < 3 {
			return "", ErrInvalidCommand
		}
		added, err := d.SAdd(ctx, args[1], args[2])
		if err != nil {
			return "", err
		}
		if added {
			return "1", nil
		}
		return "0", nil
	case "SMEMBERS":
		if len(args) < 2 {
			return "", ErrInvalidCommand
		}
		members, err := d.SMembers(ctx, args[1])
		if err != nil {
			return "", err
		}
		return strings.Join(members, ","), nil
	case "ZADD":
		if len(args) < 4 {
			return "", ErrInvalidCommand
		}
		score, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return "", err
		}
		if err = d.ZAdd(ctx, args[1], args[3], score); err != nil {
			return "", err
		}
		return "1", nil
	case "ZRANGEBYSCORE":
		if len(args) < 4 {
			return "", ErrInvalidCommand
		}
		min, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return "", err
		}
		max, err := strconv.ParseFloat(args[3], 64)
		if err != nil {
			return "", err
		}
		items, err := d.ZRangeByScore(ctx, args[1], min, max)
		if err != nil {
			return "", err
		}
		res := make([]string, 0, len(items))
		for _, item := range items {
			res = append(res, item.Member)
		}
		return strings.Join(res, ","), nil
	case "HSET":
		if len(args) < 4 {
			return "", ErrInvalidCommand
		}
		if err := d.HSet(ctx, args[1], args[2], args[3]); err != nil {
			return "", err
		}
		return "1", nil
	case "HGET":
		if len(args) < 3 {
			return "", ErrInvalidCommand
		}
		value, ok, err := d.HGet(ctx, args[1], args[2])
		if err != nil {
			return "", err
		}
		if !ok {
			return "(nil)", nil
		}
		return value, nil
	case "HGETALL":
		if len(args) < 2 {
			return "", ErrInvalidCommand
		}
		values, err := d.HGetAll(ctx, args[1])
		if err != nil {
			return "", err
		}
		pairs := make([]string, 0, len(values))
		for field, value := range values {
			pairs = append(pairs, field+"="+value)
		}
		sort.Strings(pairs)
		return strings.Join(pairs, ","), nil
	default:
		return "", ErrInvalidCommand
	}
}

// ExportZItems 导出 zset 项（供持久化使用）。
func ExportZItems(items []ds.ZItem) []ds.ZItem {
	copyItems := make([]ds.ZItem, len(items))
	copy(copyItems, items)
	return copyItems
}
