package storage

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	tarantool "github.com/tarantool/go-tarantool"
)

const (
	//keyFieldIndex   = 0
	valueFieldIndex = 1
)

type TarantoolError struct {
	Code      uint32
	Operation string
	Message   string
}

func (e *TarantoolError) Error() string {
	return fmt.Sprintf("Tarantool %s error (code=%d): %s", e.Operation, e.Code, e.Message)
}

type TarantoolStorage struct {
	conn *tarantool.Connection
}

func NewTarantoolStorage(host, user, pass string) (*TarantoolStorage, error) {
	opts := tarantool.Opts{
		User: user,
		Pass: pass,
	}
	conn, err := tarantool.Connect(host, opts)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к Tarantool: %w", err)
	}
	return &TarantoolStorage{
		conn: conn,
	}, nil
}

func (t *TarantoolStorage) Close() {
	if t.conn != nil {
		t.conn.Close()
	}
}

func (t *TarantoolStorage) checkTarantoolResp(resp *tarantool.Response, operation string) error {
	if resp.Code == 0 {
		return nil
	}
	return &TarantoolError{
		Code:      resp.Code,
		Operation: operation,
		Message:   fmt.Sprintf("%v", resp.Data),
	}
}

func (t *TarantoolStorage) exists(key string) (bool, error) {
	resp, err := t.conn.Select("kv", "primary", 0, 1, tarantool.IterEq, []interface{}{key})
	if err != nil {
		return false, err
	}
	if err := t.checkTarantoolResp(resp, "Select"); err != nil {
		return false, err
	}
	return len(resp.Data) > 0, nil
}

func (t *TarantoolStorage) Insert(key string, value map[string]interface{}) error {
	resp, err := t.conn.Insert("kv", []interface{}{key, value})
	if err != nil {
		return err
	}
	return t.checkTarantoolResp(resp, "Insert")
}

func (t *TarantoolStorage) Update(key string, value map[string]interface{}) error {
	found, err := t.exists(key)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("key '%s' not found", key)
	}
	ops := []interface{}{
		[]interface{}{"=", valueFieldIndex, value},
	}
	resp, err := t.conn.Update("kv", "primary", []interface{}{key}, ops)
	if err != nil {
		return err
	}
	return t.checkTarantoolResp(resp, "Update")
}

func (t *TarantoolStorage) Get(key string) (map[string]interface{}, error) {
	resp, err := t.conn.Select("kv", "primary", 0, 1, tarantool.IterEq, []interface{}{key})
	if err != nil {
		return nil, err
	}
	if err := t.checkTarantoolResp(resp, "Select"); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("key '%s' not found", key)
	}
	tuple, ok := resp.Data[0].([]interface{})
	if !ok || len(tuple) <= valueFieldIndex {
		return nil, fmt.Errorf("неверный формат tuple для ключа '%s'", key)
	}
	rawValue := tuple[valueFieldIndex]
	var result map[string]interface{}
	if err := mapstructure.Decode(rawValue, &result); err != nil {
		return nil, fmt.Errorf("ошибка декодирования значения: %w", err)
	}
	return result, nil
}

func (t *TarantoolStorage) Delete(key string) error {
	resp, err := t.conn.Delete("kv", "primary", []interface{}{key})
	if err != nil {
		return err
	}
	if err := t.checkTarantoolResp(resp, "Delete"); err != nil {
		return err
	}
	if len(resp.Data) == 0 {
		return fmt.Errorf("key '%s' not found", key)
	}
	return nil
}
