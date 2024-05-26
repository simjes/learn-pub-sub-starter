package main

import (
	"bytes"
	"encoding/gob"
	"time"
)

func encode(gl GameLog) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(gl)

	return buf.Bytes(), err
}

func decode(data []byte) (GameLog, error) {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)

	var gameLog GameLog
	err := decoder.Decode(&gameLog)

	return gameLog, err
}

// don't touch below this line

type GameLog struct {
	CurrentTime time.Time
	Message     string
	Username    string
}
