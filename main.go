package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type userTokens []string

type Member struct {
	Name           string
	Token          string
	BetAmount      int
	InitialBet     int
	CurrentBalance int
}

var members []Member

type OpenResultData struct {
	Dice1     int    `json:"Dice1"`
	BigNess   int    `json:"BigNess"`
	Color1    int    `json:"Color1"`
	Color2    int    `json:"Color2"`
	HashValue string `json:"HashValue"`
	Block     int    `json:"Block"`
	Seed      string `json:"Seed"`
	Create    int    `json:"create"`
}

var BetID = map[string]int{
	"big":   100728853,
	"small": 100728854,
}

var (
	waitCycles  = flag.Int("waitcycles", 2, "Cycles to wait before betting")
	doubleAfter = flag.Int("doubleafter", 1, "Cycles to double amount after")
	startBet    = flag.Int("startbet", 0, "Initial bet")
)

func getCurrentGold(token string) int {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://lucknow.game/user/v1/user/getUserItem", nil)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	return int(result["data"].(map[string]interface{})["gold"].(float64))
}

func gameFlowNumber(token string) (int, string) {
	client := &http.Client{}

	req, _ := http.NewRequest("POST", "https://lucknow.game/game/v1/game/gameHistoryGet", bytes.NewBuffer([]byte(`{"Total":0,"pageIndex":1,"pageSize":10,"gameID":24577,"total":1653}`)))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	gameNo := int(result["data"].(map[string]interface{})["gameHistorys"].([]interface{})[0].(map[string]interface{})["gameFlowNo"].(float64))

	openResultDataStr := string(result["data"].(map[string]interface{})["gameHistorys"].([]interface{})[0].(map[string]interface{})["openResult"].(string))
	var openResultData OpenResultData
	json.Unmarshal([]byte(openResultDataStr), &openResultData)

	bigness := openResultData.BigNess

	if bigness == 1 {
		return gameNo, "big"
	}
	return gameNo, "small"
}

func placeBet(no int, betname string, amount int, token string) {
	client := &http.Client{}

	payload := map[string]interface{}{
		"gameFlowBet": map[string]interface{}{
			"gameFlowNo": no,
			"gameID":     24577,
			"betTimes":   1,
			"deviceType": 4,
			"deviceID":   "dddddddddddddd",
			"betZone": []map[string]interface{}{
				{
					"gold":        amount * 100,
					"betID":       BetID[betname],
					"betIdAssist": fmt.Sprintf("{\"BetArea\":0,\"betName\":[\"\"],\"selectName\":\"%s\",\"multiple\":\"%d\",\"openPrice\":\"8658\"}", betname, amount),
					"BetName":     betname,
					"rate":        1,
					"longname":    strings.ToLower(betname),
				},
			},
			"gameType": "1 minute",
			"multiple": fmt.Sprintf("%d", amount),
			"time":     53000,
			"endtime":  0,
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://lucknow.game/game/v1/game/gameFlowBet", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		fmt.Println("Error: Received response status code", resp.StatusCode)
		responseBody := make([]byte, 0)
		if resp.Body != nil {
			responseBody, _ = io.ReadAll(resp.Body)
		}
		fmt.Println("Response Body:", string(responseBody))
		return
	}
}

func setInitialBet(m int) {
	if *startBet > 0 {
		members[m].InitialBet = *startBet
		return
	}

	if members[m].CurrentBalance < 590 {
		members[m].InitialBet = 1
	} else if members[m].CurrentBalance >= 590 && members[m].CurrentBalance < 980 {
		members[m].InitialBet = 3
	} else if members[m].CurrentBalance >= 980 && members[m].CurrentBalance < 1372 {
		members[m].InitialBet = 5
	} else if members[m].CurrentBalance >= 1372 && members[m].CurrentBalance < 1959 {
		members[m].InitialBet = 7
	} else if members[m].CurrentBalance >= 1959 {
		members[m].InitialBet = 10
	}
}

func setBetAmountToInitial() {
	for i, _ := range members {
		members[i].BetAmount = members[i].InitialBet
	}
}

func multiplyBetAmount(x int) {
	for i, _ := range members {
		members[i].BetAmount *= x
	}
}

func (t *userTokens) String() string {
	return fmt.Sprintf("%v", *t)
}

func (t *userTokens) Set(val string) error {
	*t = append(*t, val)
	return nil
}

func main() {
	var tokens userTokens
	flag.Var(&tokens, "token", "AuthToken")

	flag.Parse()

	for _, t := range tokens {
		userTok := strings.Split(t, ":")
		members = append(members, Member{Name: userTok[0], Token: userTok[1]})
	}

	for i, _ := range members {
		members[i].CurrentBalance = getCurrentGold(members[i].Token) / 100

		setInitialBet(i)
		fmt.Printf("%s : curr bal: %d; initialbet: %d\n", members[i].Name, members[i].CurrentBalance, members[i].InitialBet)
	}

	baseToken := members[0].Token

	betName := "big"
	big := 0
	small := 0
	skip := true

	fmt.Println("Starting game...")
	for {

		lastGameNo, lastResult := gameFlowNumber(baseToken)
		currentGameNo := lastGameNo + 1

		fmt.Printf("%d : skip %t bet on %s ; counts: big: %d small: %d; last result: %s\n", currentGameNo, skip, betName, big, small, lastResult)

		for i, _ := range members {
			members[i].CurrentBalance = getCurrentGold(members[i].Token) / 100
		}

		if !skip {
			// place betamounts on betname bet
			for i, _ := range members {
				if members[i].CurrentBalance < members[i].BetAmount {
					setInitialBet(i)
					members[i].BetAmount = members[i].InitialBet
				}

				betAmount := members[i].BetAmount
				if members[i].BetAmount > 100 {
					betAmount = members[i].BetAmount + int(float64(members[i].BetAmount)*0.02)
				}

				fmt.Printf("\t %s: placing bet %d, current balance: %d, initial bet: %d\n", members[i].Name, betAmount, members[i].CurrentBalance-members[i].BetAmount, members[i].InitialBet)
				placeBet(currentGameNo, betName, betAmount, members[i].Token)
			}
		} else {
			for i, _ := range members {
				setInitialBet(i)
				fmt.Printf("\t%s : curr bal: %d; initialbet: %d\n", members[i].Name, members[i].CurrentBalance, members[i].InitialBet)
			}
		}

		// wait for result
		var result string
		var gameNo int
		for {
			gameNo, result = gameFlowNumber(baseToken)
			if currentGameNo == gameNo {
				break
			}

			time.Sleep(2 * time.Second)
		}

		if result == "big" {
			small = 0
			big++
		} else if result == "small" {
			big = 0
			small++
		}

		if big >= *waitCycles {
			betName = "small"
			skip = false

			if big == *waitCycles {
				setBetAmountToInitial()
			} else if big > *waitCycles+*doubleAfter {
				multiplyBetAmount(2)
			} else {
				multiplyBetAmount(3)
			}
		} else if small >= *waitCycles {
			betName = "big"
			skip = false

			if small == *waitCycles {
				setBetAmountToInitial()
			} else if small > *waitCycles+*doubleAfter {
				multiplyBetAmount(2)
			} else {
				multiplyBetAmount(3)
			}
		} else {
			setBetAmountToInitial()
			skip = true
		}
	}
}
