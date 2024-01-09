#! /bin/bash

set -o xtrace

authToken=$AUTH_TOKEN
initialBet=${INITIAL_BET:-3}

waitCycles=2

big=0
small=0
bet=$initialBet
skip=yes
betname=BIG
betID=100728853

while true
do
  gameFlow=$(curl 'https://lucknow.game/game/v1/game/gameHistoryGet' --compressed -X POST \
    -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0' \
    -H 'Accept: application/json, text/plain, */*' \
    -H 'Accept-Language: en-US,en;q=0.5' \
    -H 'Content-Type: application/json' \
    -H "Authorization: $authToken" \
    -H 'Origin: https://lucknow.game' \
    -H 'Connection: keep-alive' \
    -H 'Referer: https://lucknow.game/' \
    -H 'Sec-Fetch-Dest: empty' \
    -H 'Sec-Fetch-Mode: cors' \
    -H 'Sec-Fetch-Site: same-origin' \
    -H 'DNT: 1' \
    -H 'Sec-GPC: 1' \
    --data-raw '{"Total":0,"pageIndex":1,"pageSize":10,"gameID":24577,"total":1653}' | jq '.data.gameHistorys[0].gameFlowNo' 
  )

  ((gameFlow++))

  echo "gameFlowNo: $gameFlow"

  echo "Skip: $skip The value of bet is: $bet in $betname"

  goldBet=$(( bet * 100 ))

  if [[ "$skip" == "no" ]]; then
    betnamesmall=$(echo $betname | tr '[:upper:]' '[:lower:]')
    curl 'https://lucknow.game/game/v1/game/gameFlowBet' --compressed -X POST \
      -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0' \
      -H 'Accept: application/json, text/plain, */*' \
      -H 'Accept-Language: en-US,en;q=0.5' \
      -H 'Content-Type: application/json' \
      -H "Authorization: $authToken" \
      -H 'Origin: https://lucknow.game' \
      -H 'Connection: keep-alive' \
      -H 'Referer: https://lucknow.game/' \
      -H 'Sec-Fetch-Dest: empty' \
      -H 'Sec-Fetch-Mode: cors' \
      -H 'Sec-Fetch-Site: same-origin' \
      -H 'DNT: 1' \
      -H 'Sec-GPC: 1' \
      -H 'TE: trailers' \
      --data-raw '{"gameFlowBet":{"gameFlowNo":'$gameFlow',"gameID":24577,"betTimes":1,"deviceType":4,"deviceID":"dddddddddddddd","betZone":[{"gold":'$goldBet',"betID":'$betID',"betIdAssist":"{\"BetArea\":0,\"betName\":[\"\"],\"selectName\":\"'$betname'\",\"multiple\":\"'$bet'\",\"openPrice\":\"8658\"}","BetName":"'$betname'","rate":1,"longname":"'$betnamesmall'"}],"gameType":"1 minute","multiple":"'$bet'","time":53000,"endtime":0}}'
  fi

  sleep 60

  bigness=$(curl 'https://lucknow.game/game/v1/game/gameHistoryGet' --compressed -X POST \
    -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0' \
    -H 'Accept: application/json, text/plain, */*' \
    -H 'Accept-Language: en-US,en;q=0.5' \
    -H 'Content-Type: application/json' \
    -H "Authorization: $authToken" \
    -H 'Origin: https://lucknow.game' \
    -H 'Connection: keep-alive' \
    -H 'Referer: https://lucknow.game/' \
    -H 'Sec-Fetch-Dest: empty' \
    -H 'Sec-Fetch-Mode: cors' \
    -H 'Sec-Fetch-Site: same-origin' \
    -H 'DNT: 1' \
    -H 'Sec-GPC: 1' \
    --data-raw '{"Total":0,"pageIndex":1,"pageSize":10,"gameID":24577,"total":1653}' | jq -r '.data.gameHistorys[0].openResult'  | jq '.BigNess'   )

  if [ "$bigness" == "1" ]; then
      small=0
      ((big++ ))
  else
      big=0
      ((small++))
  fi

  if [[ $big -ge $waitCycles ]]; then
    betname=SMALL
    betID=100728854
    skip=no
    if [[ $big == $waitCycles ]]; then
      bet=$initialBet
    elif [[ $big > 4 ]]; then
      bet=$((bet * 2))
    else
      bet=$((bet * 3))
    fi
  elif [[ $small -ge $waitCycles ]]; then
    betname=BIG
    betID=100728853
    skip=no

    if [[ $small == $waitCycles ]]; then
      bet=$initialBet
    elif [[ $small > 4 ]]; then
      bet=$((bet * 2))
    else
      bet=$((bet * 3))
    fi
  else
    bet=$initialBet
    skip=yes
  fi

done
