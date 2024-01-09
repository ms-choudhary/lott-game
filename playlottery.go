package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	authToken := os.Getenv("AUTH_TOKEN")
	initialBetStr := os.Getenv("INITIAL_BET")
	initialBet, err := strconv.Atoi(initialBetStr)
	if err != nil {
		initialBet = 3
	}

	waitCycles := 3

	big := 0
	small := 0
	bet := initialBet
	skip := true
	betname := "BIG"
	betID := 100728853

	client := &http.Client{}

	for {
		req, _ := http.NewRequest("POST", "https://lucknow.game/game/v1/game/gameHistoryGet", bytes.NewBuffer([]byte(`{"Total":0,"pageIndex":1,"pageSize":10,"gameID":24577,"total":1653}`)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authToken)
		resp, _ := client.Do(req)
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		gameFlowNo := int(result["data"].(map[string]interface{})["gameHistorys"].([]interface{})[0].(map[string]interface{})["gameFlowNo"].(int64)) + 1

		fmt.Printf("gameFlowNo: %d\n", gameFlowNo)

		fmt.Printf("Skip: %v The value of bet is: %d in %s\n", skip, bet, betname)

		goldBet := bet * 100

		if !skip {
			betnamesmall := strings.ToLower(betname)
			payload := map[string]interface{}{
				"gameFlowBet": map[string]interface{}{
					"gameFlowNo": gameFlowNo,
					"gameID":     24577,
					"betTimes":   1,
					"deviceType": 4,
					"deviceID":   "dddddddddddddd",
					"betZone": []map[string]interface{}{
						{
							"gold":        goldBet,
							"betID":       betID,
							"betIdAssist": fmt.Sprintf("{\"BetArea\":0,\"betName\":[\"\"],\"selectName\":\"%s\",\"multiple\":\"%d\",\"openPrice\":\"8658\"}", betname, bet),
							"BetName":     betname,
							"rate":        1,
							"longname":    betnamesmall,
						},
					},
					"gameType": "1 minute",
					"multiple": fmt.Sprintf("%d", bet),
					"time":     53000,
					"endtime":  0,
				},
			}

			payloadBytes, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", "https://lucknow.game/game/v1/game/gameFlowBet", bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", authToken)
			resp, _ := client.Do(req)
			_ = resp.Body.Close()
		}

		time.Sleep(60 * time.Second)

		req, _ = http.NewRequest("POST", "https://lucknow.game/game/v1/game/gameHistoryGet", bytes.NewBuffer([]byte(`{"Total":0,"pageIndex":1,"pageSize":10,"gameID":24577,"total":1653}`)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authToken)
		resp, _ = client.Do(req)
		defer resp.Body.Close()

		body, _ = ioutil.ReadAll(resp.Body)

		json.Unmarshal(body, &result)
		bigness := int(result["data"].(map[string]interface{})["gameHistorys"].([]interface{})[0].(map[string]interface{})["openResult"].(map[string]interface{})["BigNess"].(float64))

		if bigness == 1 {
			small = 0
			big++
		} else {
			big = 0
			small++
		}

		if big >= waitCycles {
			betname = "SMALL"
			betID = 100728854
			skip = false
			if big == waitCycles {
				bet = initialBet
			} else if big > 4 {
				bet *= 2
			} else {
				bet *= 3
			}
		} else if small >= waitCycles {
			betname = "BIG"
			betID = 100728853
			skip = false
			if small == waitCycles {
				bet = initialBet
			} else if small > 4 {
				bet *= 2
			} else {
				bet *= 3
			}
		} else {
			bet = initialBet
			skip = true
		}
	}
}
