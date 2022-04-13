package redis

import (
	"encoding/binary"
	"os"
	"testing"
)

type TestStruct struct {
	I int
	S string
}

func (t TestStruct) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 8)
	binary.BigEndian.PutUint32(data, uint32(t.I))
	binary.BigEndian.PutUint32(data[4:], uint32(len(t.S)))
	data = append(data, t.S...)
	return data, err
}

func (t *TestStruct) UnmarshalBinary(data []byte) error {
	t.I = int(binary.BigEndian.Uint32(data))
	length := binary.BigEndian.Uint32(data[4:])
	t.S = string(data[8 : 8+length])
	return nil
}

func TestRedisCache(t *testing.T) {
	err := Cache("fuck", &TestStruct{
		I: 1,
		S: "fuck",
	}, 1)
	if err != nil {
		panic(err)
		return
	}

	result := &TestStruct{}
	err = GetCache("fuck", result)
	if err != nil {
		panic(err)
	}

	if result.S != "fuck" || result.I != 1 {
		t.Error("fuck is not fuck")
	}
}

func TestRedisStore(t *testing.T) {
	err := Store("fuck", &TestStruct{
		I: 1,
		S: "fuck",
	})
	if err != nil {
		panic(err)
		return
	}

	result := &TestStruct{}
	err = GetStore("fuck", result)
	if err != nil {
		panic(err)
	}

	if result.S != "fuck" || result.I != 1 {
		t.Error("fuck is not fuck")
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	os.Setenv("bool", "false")
	os.Setenv("int", "0")
	os.Setenv("float", "0.123")
	os.Setenv("string", "fuck")

	if getEnvOrDefault("bool", true) ||
		getEnvOrDefault("int", 1) != 0 ||
		getEnvOrDefault("float", 1.234) != 0.123 ||
		getEnvOrDefault("string", "virgin") != "fuck" {
		t.Error("getEnvOrDefault error")
	}
}
