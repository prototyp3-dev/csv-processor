package processor

import (
  "io"
  "fmt"
  "strings"
	"bytes"
  "encoding/csv"
	"encoding/binary"
	"encoding/hex"
	"compress/gzip"
  cid "github.com/ipfs/go-cid"
  // mc "github.com/multiformats/go-multicodec"
  // mh "github.com/multiformats/go-multihash"

  "dapp/model"
  // "github.com/prototyp3-dev/go-rollups"
)

func CsvBlankCellPermillionage(csvString string, nilFields ...string) (uint64,error) {

  reader := csv.NewReader(strings.NewReader(csvString))

  nilFieldsMap := make(map[string]bool)
  for _, nilField := range nilFields {
    nilFieldsMap[strings.ToLower(nilField)] = true
  }

  var totalCells uint64
  var emptyCells uint64

  firstRow := true
  for {
    record, err := reader.Read()
    if err == io.EOF {
      break
    }
    if err != nil {
      return 0,err
    }
    if firstRow {
      firstRow = false
      continue
    }

    for _, value := range record {
      totalCells += 1
      if value == "" || nilFieldsMap[strings.ToLower(value)] {
        emptyCells += 1
      }
    }
  }
  nDataCells := totalCells-emptyCells
  value := 1000000*nDataCells/totalCells

  return value,nil
}

func GetDataCid(data string) (cid.Cid,error) {
  pref := cid.Prefix{
    Version: 1,
    Codec: uint64(85), // uint64(mc.Raw),
    MhType: 0x12, // mh.SHA2_256,
    MhLength: -1, // default length
  }
  
  dataCid, err := pref.Sum([]byte(data))
  if err != nil {
    return cid.Cid{}, fmt.Errorf("GetDataCid: error getting CID: %s", err)
  }

  return dataCid,nil
}

func CompareCidWithString(dataCid cid.Cid, marshaledString string) (bool,error) {
  cidFromString, err := cid.Decode(marshaledString)
  if err != nil {
    return false, fmt.Errorf("CompareCidWithString: error getting CID: %s", err)
  }
  if cidFromString.Equals(dataCid) {
    return true, nil
  }
  return false, nil
}

func CompressData(data []byte) ([]byte,error) {
  var buf bytes.Buffer
  zw := gzip.NewWriter(&buf)

  _, err := zw.Write(data)
  if err != nil {
    return buf.Bytes(), fmt.Errorf("CompressData: error compressing data: %s", err)
  }
  if err := zw.Close(); err != nil {
    return buf.Bytes(), fmt.Errorf("CompressData: error closing writer: %s", err)
  }
  return buf.Bytes(), nil
}

func DecompressData(data []byte) ([]byte,error) {
  buf := bytes.NewBuffer(data)
  var bufOut bytes.Buffer

  zr, err := gzip.NewReader(buf)
  if err != nil {
    return bufOut.Bytes(), fmt.Errorf("DecompressData: error decompressing data: %s", err)
  }

  if _, err := io.Copy(&bufOut, zr); err != nil {
    return bufOut.Bytes(), fmt.Errorf("DecompressData: error copying data to out bufer: %s", err)
  }
  if err := zr.Close(); err != nil {
    return bufOut.Bytes(), fmt.Errorf("DecompressData: error closing reader: %s", err)
  }
  return bufOut.Bytes(),nil
}

func PrepareDataToSend(data []byte, maxSize uint64) ([]string,error) {
  preparedData := []string{}
	if len(data) < 1 {
		return preparedData,fmt.Errorf("PrepareData: Invalid empty data")
	}
  compressed,err := CompressData(data)
	if err != nil {
		return preparedData,fmt.Errorf("PrepareData: error compressing data: %s", err)
	}
  sizeData := uint64(len(compressed))
  totalChunks := uint32(sizeData/maxSize)
  if sizeData % maxSize == 0 {
    totalChunks -= 1
  }

  totalChunksBytes := make([]byte, 4)
  binary.BigEndian.PutUint32(totalChunksBytes, totalChunks)

  for chunkIndex := uint32(0); chunkIndex <= totalChunks; chunkIndex += 1 {
    chunksIndexBytes := make([]byte, 4)
    binary.BigEndian.PutUint32(chunksIndexBytes, chunkIndex)
    metadata := append(chunksIndexBytes, totalChunksBytes...)
    top := uint64(chunkIndex+1)*(maxSize)
    if top > sizeData {
      top = sizeData
    }
    allData := append(metadata,compressed[uint64(chunkIndex)*maxSize:top]...)

    hx := hex.EncodeToString(allData)
    allDataHex := "0x"+string(hx)
    preparedData = append(preparedData, allDataHex)
  }

  return preparedData,nil
}

func UpdateDataChunks(dataChunks *model.DataChunks, chunkHex string) error {
  chunk, err := hex.DecodeString(chunkHex[2:])
  if err != nil {
    return fmt.Errorf("UpdateDataChunks: Error converting hex to bytes %s",err)
  }
  chunkIndex := binary.BigEndian.Uint32(chunk[0:4])
  totalChunks := binary.BigEndian.Uint32(chunk[4:8]) + 1
  data := chunk[8:]

  if chunkIndex > totalChunks {
    return fmt.Errorf("UpdateDataChunks: Inconsistent chunk index, greater than total")
  }

  if dataChunks.TotalChunks == 0 {
    dataChunks.ChunksData = make(map[uint32]*model.Chunk)
    dataChunks.TotalChunks = totalChunks
  } else {
    if totalChunks != dataChunks.TotalChunks {
      return fmt.Errorf("UpdateDataChunks: Can't append chunk, Inconsistent number of chunks")
    }
  }
  dataChunks.ChunksData[chunkIndex] = &model.Chunk{Data:data}
  return nil
}

func ComposeDataFromChunks(dataChunks *model.DataChunks) ([]byte,error) {
  var data []byte
  if uint32(len(dataChunks.ChunksData)) != dataChunks.TotalChunks {
    return data,fmt.Errorf("ComposeDataFromChunks: Wrong number of chunks")
  }
  orderedChunks := make([][]byte,dataChunks.TotalChunks)
  for i, chunk := range dataChunks.ChunksData {
    orderedChunks[i] = chunk.Data
  }
  var compressed []byte
  for _, chunkData := range orderedChunks {
    compressed = append(compressed, chunkData...)
  }

  decompressed,err := DecompressData(compressed)
	if err != nil {
		return data,fmt.Errorf("ComposeDataFromChunks: Error decompressing data %s",err)
	}
  return decompressed,nil
}
