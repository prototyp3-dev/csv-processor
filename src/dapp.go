package main

import (
  "encoding/json"
  "log"
  "os"
  "fmt"
  "strconv"
  "io/ioutil"
  "strings"
	"regexp"

  "dapp/model"
  "dapp/processor"

  "github.com/prototyp3-dev/go-rollups/rollups"
  "github.com/prototyp3-dev/go-rollups/handler"
)

var (
  infolog  = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)
  errlog   = log.New(os.Stderr, "[ error ] ", log.Lshortfile)
)

var claimTimeout uint64
var disputeTimeout uint64
var users map[string]*model.User
var claims map[string]*model.Claim


func GetUser(address string) *model.User {
  user := users[address]
  if user == nil {
    newUser := model.User{OpenClaims: make(map[string]struct{}), OpenDisputes: make(map[string]struct{})}
    users[address] = &newUser
    user = users[address]
  }
  return user
}

func GetClaimList(payloadMap map[string]interface{}) error {
  infolog.Println("Got claim list request")
  claimList := []*model.SimplifiedClaim{}
  for k, _ := range claims {
    claimList = append(claimList, &model.SimplifiedClaim{Id:k,Status:claims[k].Status,Value:claims[k].Value})
  }

  claimListJson, err := json.Marshal(claimList)
  if err != nil {
    return err
  }
  
  report := rollups.Report{rollups.Str2Hex(string(claimListJson))}
  res, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("GetClaimList: error making http request: %s", err)
  }
  infolog.Println("Received report status", strconv.Itoa(res.StatusCode))
  
  return nil
}

func ShowUser(payloadMap map[string]interface{}) error {
  infolog.Println("Got show user request")
  userAddress, ok := payloadMap["id"].(string)

  if !ok || userAddress == "" {
    message := "ShowUser: Not enough parameters, you must provide string 'id'"
    report := rollups.Report{rollups.Str2Hex(message)}
    _, err := rollups.SendReport(&report)
    if err != nil {
      return fmt.Errorf("ShowUser: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }

  userAddress = strings.ToLower(userAddress)
  infolog.Println("For user user",userAddress)

  if users[userAddress] == nil {
    message := "ShowUser: User doesn't exist"
    report := rollups.Report{rollups.Str2Hex(message)}
    _, err := rollups.SendReport(&report)
    if err != nil {
      return fmt.Errorf("ShowUser: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }

  user := users[userAddress]

  userJson, err := json.Marshal(user)
  if err != nil {
    return err
  }
  
  report := rollups.Report{rollups.Str2Hex(string(userJson))}
  res, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("ShowUser: error making http request: %s", err)
  }
  infolog.Println("Received report status", strconv.Itoa(res.StatusCode))
  
  return nil
}

func ShowClaim(payloadMap map[string]interface{}) error {
  infolog.Println("Got show claim request")
  claimId, ok := payloadMap["id"].(string)

  if !ok || claimId == "" {
    message := "ShowClaim: Not enough parameters, you must provide string 'id'"
    report := rollups.Report{rollups.Str2Hex(message)}
    _, err := rollups.SendReport(&report)
    if err != nil {
      return fmt.Errorf("ShowClaim: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }
  infolog.Println("For claim",claimId)

  if claims[claimId] == nil {
    message := "ShowClaim: Claim doesn't exist"
    report := rollups.Report{rollups.Str2Hex(message)}
    _, err := rollups.SendReport(&report)
    if err != nil {
      return fmt.Errorf("ShowClaim: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }
  
  claim := claims[claimId]
  
  claimJson, err := json.Marshal(claim)
  if err != nil {
    return err
  }
  
  report := rollups.Report{rollups.Str2Hex(string(claimJson))}
  res, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("ShowClaim: error making http request: %s", err)
  }
  infolog.Println("Received report status", strconv.Itoa(res.StatusCode))
  
  return nil
}

func GetWasm(payloadMap map[string]interface{}) error {
  infolog.Println("Got wasm request")
  files, err := ioutil.ReadDir(".")
  if err != nil {
      log.Fatal(err)
  }

	var validFile = regexp.MustCompile(`^.+\.wasm$`)

  for _, file := range files {
    if !file.IsDir() && validFile.MatchString(file.Name()) {
      infolog.Println("Found file",file.Name())

      fileBytes, err := ioutil.ReadFile(file.Name())
      if err != nil {
        return fmt.Errorf("GetWasm: error opening file %s: %s", file.Name(), err)
      }
      
      report := rollups.Report{rollups.Bin2Hex(fileBytes)}
      res, err := rollups.SendReport(&report)
      if err != nil {
        return fmt.Errorf("ShowClaim: error making http request: %s", err)
      }
      infolog.Println("Received report status", strconv.Itoa(res.StatusCode))  
    }
  }
  return nil
}

// Receive and store claim
func HandleClaim(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  infolog.Println("Got claim request")
  user := GetUser(metadata.MsgSender)

  claimId, ok1 := payloadMap["id"].(string)
  claimValueFloat, ok2 := payloadMap["value"].(float64) // value 100,000 == 100%

  if !ok1 || !ok2 || claimId == "" || claimValueFloat > 1000000 {
    message := "HandleClaim: Not enough parameters, you must provide string 'id' and uint 'value'"
    report := rollups.Report{rollups.Str2Hex(message)}
    _, err := rollups.SendReport(&report)
    if err != nil {
      return fmt.Errorf("HandleSet: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }
  claimValue := uint64(claimValueFloat)

  // Check if claim already exists
  if claims[claimId] != nil {
    return fmt.Errorf("HandleClaim: Claim already exists")
  }

  claim := model.Claim{Status: model.Open, Value: claimValue, LastEdited: metadata.Timestamp, UserAddress: metadata.MsgSender}
  claims[claimId] = &claim
  user.OpenClaims[claimId] = struct{}{}

  message := fmt.Sprint("Claim ",claimId," created: ", claim)
  
  report := rollups.Report{rollups.Str2Hex(message)}
  _, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleClaim: error making http request: %s", err)
  }

  infolog.Println(message)

  return nil
}

// Finalize a claim
func HandleFinalize(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  infolog.Println("Got finalize request")
  // note: it doesn't require user that claimed or disputed to finalize claim
  
  claimId, ok := payloadMap["id"].(string)
  if !ok || claimId == "" {
    message := "HandleFinalize: Not enough parameters, you must provide string 'claimId'"
    report := rollups.Report{rollups.Str2Hex(message)}
    _, err := rollups.SendReport(&report)
    if err != nil {
      return fmt.Errorf("HandleFinalize: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }

  // Check if claim exists
  if claims[claimId] == nil {
    return fmt.Errorf("HandleFinalize: Claim doesn't exist")
  }
  claim := claims[claimId]

  switch claim.Status {
  case model.Open:
    // Check if enought time passed
    if metadata.Timestamp < claim.LastEdited + claimTimeout {
      secondsToAccept := claim.LastEdited + claimTimeout - metadata.Timestamp
      return fmt.Errorf("HandleFinalize: Claim can't be finalized yet, %d more seconds to go",secondsToAccept)
    }
    
    // finalize claim
    claim.Status = model.Finalized // change status
    claim.LastEdited = metadata.Timestamp
    
    user := GetUser(claim.UserAddress)
    user.TotalClaims += 1 // add to user finalized claims
    user.CorrectClaims += 1 // add to user finalized correct claims

    delete(user.OpenClaims,claimId) // delete from users open claims 

  case model.Disputing:
    // finalizing disputing claims is always lost dispute

    // Check if enought time passed
    if metadata.Timestamp < claim.LastEdited + disputeTimeout {
      secondsToAccept := claim.LastEdited + disputeTimeout - metadata.Timestamp
      return fmt.Errorf("HandleFinalize: Claim can't be finalized yet, %d more seconds to go",secondsToAccept)
    }
    
    // finalize claim
    claim.Status = model.Disputed // change status
    claim.LastEdited = metadata.Timestamp
    
    user := GetUser(claim.UserAddress)
    user.TotalClaims += 1 // add to user finalized claims
    user.TotalDisputes += 1 // add to user disputes
    delete(user.OpenDisputes,claimId) // delete from users open claims 

    disputingUser := GetUser(claim.DisputingUserAddress)
    disputingUser.TotalDisputes += 1 // add to user disputes
    disputingUser.WonDisputes += 1 // add to user won disputes

  default:
    return fmt.Errorf("HandleFinalize: Can only finalize Open or Disputing claims")

  }

  message := fmt.Sprint("Claim ",claimId," finalized: ", claim)
  
  report := rollups.Report{rollups.Str2Hex(message)}
  _, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleFinalize: error making http request: %s", err)
  }

  infolog.Println(message)

  return nil
}

// Dispute a claim
func HandleDispute(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  infolog.Println("Got dispute request")

  // check claim id
  claimId, ok := payloadMap["id"].(string)
  if !ok || claimId == "" {
    message := "HandleDispute: Not enough parameters, you must provide string 'claimId'"
    report := rollups.Report{rollups.Str2Hex(message)}
    _, err := rollups.SendReport(&report)
    if err != nil {
      return fmt.Errorf("HandleDispute: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }

  // Check if claim exists
  if claims[claimId] == nil {
    return fmt.Errorf("HandleDispute: Claim doesn't exist")
  }
  claim := claims[claimId]

  if claim.Status != model.Open {
    return fmt.Errorf("HandleDispute: Can only dispute Open claims")
  }

  if claim.UserAddress == metadata.MsgSender {
    return fmt.Errorf("HandleDispute: Can not dispute own claims")
  }

  // dispute claim
  claim.Status = model.Disputing // change status
  claim.DisputingUserAddress = metadata.MsgSender
  claim.LastEdited = metadata.Timestamp

  user := GetUser(claim.UserAddress)
  user.OpenDisputes[claimId] = struct{}{} // add to users open disputes
  delete(user.OpenClaims,claimId) // delete from users open claims 

  message := fmt.Sprint("Claim ",claimId," disputed: ", claim)
  
  report := rollups.Report{rollups.Str2Hex(message)}
  _, err := rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleDispute: error making http request: %s", err)
  }

  infolog.Println(message)

  return nil
}

func HandleValidateChunk(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  // check claim id
  claimId, ok1 := payloadMap["id"].(string)
  claimData, ok2 := payloadMap["data"].(string)

  if !ok1 || !ok2 || claimId == "" || len(claimData) == 0 {
    message := "HandleValidateChunk: Not enough parameters, you must provide string 'claimId' and 'data' bytes"
    report := rollups.Report{rollups.Str2Hex(message)}
    _, err := rollups.SendReport(&report)
    if err != nil {
      return fmt.Errorf("HandleValidateChunk: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }

  // Check if claim exists
  if claims[claimId] == nil {
    return fmt.Errorf("HandleValidateChunk: Claim doesn't exist")
  }
  claim := claims[claimId]

  if claim.Status != model.Open && claim.Status != model.Disputing {
    return fmt.Errorf("HandleValidateChunk: Can only dispute Open and Disputing claims")
  }

  if claim.UserAddress != metadata.MsgSender {
    return fmt.Errorf("HandleValidateChunk: Can only validate own claims")
  }

  if claim.DataChunks == nil {
    claim.DataChunks = &model.DataChunks{ChunksData:make(map[uint32]*model.Chunk)}
  }

  err := processor.UpdateDataChunks(claim.DataChunks,claimData)
  if err != nil {
    return fmt.Errorf("HandleValidatePart: Error updating data chunks: %s",err)
  }

  if uint32(len(claim.DataChunks.ChunksData)) == claim.DataChunks.TotalChunks {
    composed,err := processor.ComposeDataFromChunks(claim.DataChunks)
    if err != nil {
      return fmt.Errorf("HandleValidatePart: Error composing data chunks: %s",err)
    }
    claim.DataChunks = nil

    return ValidateAndFinalizeClaim(claimId,string(composed),metadata.Timestamp)
  }
  return nil
}

// validate an open claim and finalize it (in dispute or not)
func HandleValidate(metadata *rollups.Metadata, payloadMap map[string]interface{}) error {
  infolog.Println("Got validate request")
  // notes: require user that claimed to validate claim
  //        can even validate claims not in dispute

  // check claim id
  claimId, ok1 := payloadMap["id"].(string)
  claimData, ok2 := payloadMap["data"].(string)

  if !ok1 || !ok2 || claimId == "" || claimData == "" {
    message := "HandleValidate: Not enough parameters, you must provide string 'claimId' and 'data'"
    report := rollups.Report{rollups.Str2Hex(message)}
    _, err := rollups.SendReport(&report)
    if err != nil {
      return fmt.Errorf("HandleValidate: error making http request: %s", err)
    }
    return fmt.Errorf(message)
  }

  // Check if claim exists
  if claims[claimId] == nil {
    return fmt.Errorf("HandleValidate: Claim doesn't exist")
  }
  claim := claims[claimId]

  if claim.Status != model.Open && claim.Status != model.Disputing {
    return fmt.Errorf("HandleValidate: Can only dispute Open and Disputing claims")
  }

  if claim.UserAddress != metadata.MsgSender {
    return fmt.Errorf("HandleValidate: Can only validate own claims")
  }

  return ValidateAndFinalizeClaim(claimId,claimData,metadata.Timestamp)
}

func ValidateAndFinalizeClaim(claimId string,claimData string, timestamp uint64) error {
  claim := claims[claimId]

  isClaimValid, err := ValidateClaim(claimId,claim.Value,claimData)
  if err != nil {
    report := rollups.Report{rollups.Str2Hex(fmt.Sprint("HandleValidate: Error during claim validation: %s",err))}
    _, err = rollups.SendReport(&report)
    if err != nil {
      return fmt.Errorf("HandleValidate: error making http request:", err)
    }
  }

  var message string

  if isClaimValid {
    // validate claim
    claim.Status = model.Validated // change status
    claim.LastEdited = timestamp
    
    user := GetUser(claim.UserAddress)
    user.TotalClaims += 1 // add to user finalized claims
    user.CorrectClaims += 1 // add to user finalized correct claims

    delete(user.OpenClaims,claimId) // delete from users open claims 

    if claim.Status == model.Disputing {
      disputingUser := GetUser(claim.DisputingUserAddress)
      disputingUser.TotalDisputes += 1 // add to user disputes

      delete(user.OpenDisputes,claimId) // delete from users open claims 
    }
    message = fmt.Sprint("Claim ",claimId," validated: ", claim)

  } else {
    // contradict claim
    claim.Status = model.Contradicted // change status
    claim.LastEdited = timestamp
    
    user := GetUser(claim.UserAddress)
    user.TotalClaims += 1 // add to user finalized claims

    delete(user.OpenClaims,claimId) // delete from users open claims 

    if claim.Status  == model.Disputing {
      disputingUser := GetUser(claim.DisputingUserAddress)
      disputingUser.TotalDisputes += 1 // add to user disputes
      disputingUser.WonDisputes += 1 // add to user won disputes

      delete(user.OpenDisputes,claimId) // delete from users open claims 
    }
  
    message = fmt.Sprint("Claim ",claimId," contradicted: ", claim)
  }
  
  report := rollups.Report{rollups.Str2Hex(message)}
  _, err = rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleValidate: error making http request: %s", err)
  }

  infolog.Println(message)

  return nil
}

func ValidateClaim(claimId string, claimValue uint64, claimData string) (bool,error) {

  // validate processing, any error processing or failed process contradicts
  cid, err := processor.GetDataCid(claimData)
  if err != nil {
    return false, err
  }

  infolog.Println("claimId",claimId,"and got the CID", cid)

  equalCid, err := processor.CompareCidWithString(cid,claimId)
  if err != nil || !equalCid {
    return false, err
  }

  permillionValue, err := processor.CsvBlankCellPermillionage(claimData,"na")
  if err != nil || permillionValue != claimValue {
    return false, err
  }

  return true, nil
}

func HandleDefault(payloadHex string) error {

  payload, err := rollups.Hex2Str(payloadHex)
  if err != nil {
    return fmt.Errorf("HandleDefault: hex error decoding payload: %s", err)
  }

  message := fmt.Sprint("HandleDefault: Unrecognized ",payload," input, you should send a valid json")
  report := rollups.Report{rollups.Str2Hex(message)}
  _, err = rollups.SendReport(&report)
  if err != nil {
    return fmt.Errorf("HandleDefault: error making http request: %s", err)
  }
  return fmt.Errorf(message)
}

func main() {
  users = make(map[string]*model.User)
  claims = make(map[string]*model.Claim)
  claimTimeout = 30 //86400
  disputeTimeout = 30 //43200

  jsonHandler := handler.NewJsonHandler("action")

  jsonHandler.HandleInspectRoute("showUser",ShowUser)
  jsonHandler.HandleInspectRoute("showClaim",ShowClaim)
  jsonHandler.HandleInspectRoute("getClaimList",GetClaimList)
  jsonHandler.HandleInspectRoute("wasm",GetWasm)

  jsonHandler.HandleAdvanceRoute("claim", HandleClaim)
  jsonHandler.HandleAdvanceRoute("dispute", HandleDispute)
  jsonHandler.HandleAdvanceRoute("finalize", HandleFinalize)
  jsonHandler.HandleAdvanceRoute("validate", HandleValidate)
  jsonHandler.HandleAdvanceRoute("validateChunk", HandleValidateChunk)
  
  handler.HandleDefault(HandleDefault)

  err := jsonHandler.Run()
  if err != nil {
    log.Panicln(err)
  }
}