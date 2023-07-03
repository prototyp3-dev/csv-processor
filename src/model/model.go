package model

import (
  "fmt"
  "encoding/json"
)

type User struct {
  OpenClaims map[string]struct{}  `json:"openClaims"`
  OpenDisputes map[string]struct{}`json:"openDisputes"`
  TotalDisputes uint32            `json:"totalDisputes"`
  WonDisputes uint32              `json:"wonDisputes"`
  TotalClaims uint32              `json:"totalClaims"`
  CorrectClaims uint32            `json:"correctClaims"`
}

type Claim struct {
  UserAddress string              `json:"userAddress"`
  DisputingUserAddress string     `json:"disputingUserAddress"`
  Value uint64                    `json:"value"`
  LastEdited uint64               `json:"lastEdited"`
  Status Status                   `json:"status"`
  DataChunks *DataChunks          `json:"dataChunks"`
}

type SimplifiedClaim struct {
  Id string                       `json:"id"`
  Status Status                   `json:"status"`
  Value uint64                    `json:"value"`
}

type DataChunks struct {
  ChunksData map[uint32]*Chunk
  TotalChunks uint32
}
func (dc DataChunks) MarshalJSON() ([]byte, error) {
  var size uint64
  var chunkIndexes []uint32
  for index, chunk := range dc.ChunksData {
    size += uint64(len(chunk.Data))
    chunkIndexes = append(chunkIndexes,index)
  }
  return json.Marshal(struct{
    TotalChunks uint32            `json:"totalChunks"`
    CurrentSize uint64            `json:"size"`
    Chunks []uint32               `json:"chunks"`
  }{TotalChunks:dc.TotalChunks,CurrentSize:size,Chunks:chunkIndexes})
}

type Chunk struct {
  Data []byte
}
func (c Chunk) String() string {
  return fmt.Sprintf("%db",len(c.Data))
}

type Status uint8

const (
  Undefined Status = iota
  Open
  Disputing
  Finalized
  Disputed
  Validated
  Contradicted
)

func (s Status) String() string {

	statuses := [...]string{"undefined", "open", "disputing", "finalized","disputed","validated","contradicted"}
	if len(statuses) < int(s) {
		return "unknown"
	}
	return statuses[s]
}

func (s Status) MarshalJSON() ([]byte, error) {
  return json.Marshal(s.String())
}