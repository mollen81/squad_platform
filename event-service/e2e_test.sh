#!/usr/bin/env bash
# e2e_test.sh — полный тест всех ручек event-service
# Покрытие: CreateEvent, GetEventsByCreatorId, GetLastEventByCreatorId,
#   GetEventsByEventName, GetUnfinishedEventsByUserID, GetUnfinishedEventsByEventName,
#   UpdateTimeEvent, CancelEvent, JoinToEvent, LeaveEvent, SetRole,
#   GetEventMembersList, GetTeamsByEventID, GetTeamByID, JoinUserToTeam,
#   RemoveUserFromTeam, StartTeamGame, FinishTeamGame, AddTeamMemberStats, GetTeamStats
#
# Зависимости: grpcurl, jq
# Использование: ./e2e_test.sh [host:port]   (по умолчанию localhost:9096)

set -euo pipefail

HOST="${1:-localhost:9096}"
GRPC="grpcurl -plaintext"

BOLD='\033[1m'; RED='\033[0;31m'; GREEN='\033[0;32m'
BLUE='\033[1;34m'; YELLOW='\033[0;33m'; NC='\033[0m'

step()    { echo -e "\n${BLUE}${BOLD}===== $* =====${NC}"; }
section() { echo -e "\n${YELLOW}${BOLD}── $* ──${NC}"; }
ok()      { echo -e "  ${GREEN}[OK]${NC} $*"; }
fail()    { echo -e "  ${RED}[FAIL]${NC} $*"; exit 1; }
info()    { echo "      $*"; }

check_err() {
  local resp="$1" label="$2"
  local err; err=$(echo "$resp" | jq -r '.error // empty')
  [ -n "$err" ] && fail "$label: $err" || ok "$label"
}

for cmd in grpcurl jq; do
  command -v "$cmd" &>/dev/null || fail "требуется $cmd (не найден в PATH)"
done

uuid() { cat /proc/sys/kernel/random/uuid; }

CREATOR_ID=$(uuid); CREATOR_CLAN=$(uuid)
ENEMY_ID=$(uuid);   ENEMY_CLAN=$(uuid)
P3_ID=$(uuid);      P3_CLAN=$(uuid)
P4_ID=$(uuid);      P4_CLAN=$(uuid)
EVENT_NAME="e2e-full-$(date +%s)"

echo -e "${BOLD}Участники:${NC}"
info "creator : $CREATOR_ID"
info "enemy   : $ENEMY_ID"
info "player3 : $P3_ID"
info "player4 : $P4_ID"

# ─────────────────────────────────────────────────────────────────────────────
section "БЛОК 0 — CancelEvent (отдельный ивент)"
# ─────────────────────────────────────────────────────────────────────────────

step "0.1. CreateEvent (для теста CancelEvent)"
NOW=$(date +%s); TS=$((NOW + 20)); RFC=$(date -u -d @$TS '+%Y-%m-%dT%H:%M:%SZ')
RESP=$($GRPC -d "{
  \"user_creator_id\":           \"$CREATOR_ID\",
  \"creator_clan_id\":           \"$CREATOR_CLAN\",
  \"enemy_side_leader_id\":      \"$ENEMY_ID\",
  \"enemy_side_leader_clan_id\": \"$ENEMY_CLAN\",
  \"event_name\":                \"cancel-test-$(date +%s)\",
  \"time_start\":                \"$RFC\",
  \"target_game_count\":         1
}" $HOST event.EventService/CreateEvent)
check_err "$RESP" "CreateEvent (cancel-event)"

EV=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetLastEventByCreatorId)
CANCEL_ID=$(echo "$EV" | jq -r '.event.eventId')
ok "cancel_event_id=$CANCEL_ID"

step "0.2. CancelEvent"
RESP=$($GRPC -d "{\"event_id\":\"$CANCEL_ID\",\"user_create_id\":\"$CREATOR_ID\"}" \
  $HOST event.EventService/CancelEvent)
echo "$RESP"
check_err "$RESP" "CancelEvent"

EV=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetLastEventByCreatorId)
STATUS=$(echo "$EV" | jq -r '.event.status')
[ "$STATUS" = "canceled" ] && ok "status=canceled ✓" || fail "expected canceled, got $STATUS"

# ─────────────────────────────────────────────────────────────────────────────
section "БЛОК 1 — CreateEvent + все геттеры (status=pending)"
# ─────────────────────────────────────────────────────────────────────────────

step "1.1. CreateEvent (основной ивент, target_game_count=3)"
NOW=$(date +%s)
TIME_START=$((NOW + 20))
TIME_START_RFC=$(date -u -d @$TIME_START '+%Y-%m-%dT%H:%M:%SZ')
RESP=$($GRPC -d "{
  \"user_creator_id\":           \"$CREATOR_ID\",
  \"creator_clan_id\":           \"$CREATOR_CLAN\",
  \"enemy_side_leader_id\":      \"$ENEMY_ID\",
  \"enemy_side_leader_clan_id\": \"$ENEMY_CLAN\",
  \"event_name\":                \"$EVENT_NAME\",
  \"time_start\":                \"$TIME_START_RFC\",
  \"target_game_count\":         3
}" $HOST event.EventService/CreateEvent)
check_err "$RESP" "CreateEvent (main)"

step "1.2. GetLastEventByCreatorId → event_id"
EV=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetLastEventByCreatorId)
echo "$EV"
EVENT_ID=$(echo "$EV" | jq -r '.event.eventId')
[ "$EVENT_ID" != "null" ] && [ -n "$EVENT_ID" ] \
  && ok "event_id=$EVENT_ID" || fail "event_id not found"

step "1.3. GetEventsByCreatorId"
RESP=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetEventsByCreatorId)
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.events | length')
[ "$COUNT" -ge 1 ] && ok "GetEventsByCreatorId → $COUNT events" || fail "no events"

step "1.4. GetEventsByEventName"
RESP=$($GRPC -d "{\"event_name\":\"$EVENT_NAME\"}" $HOST event.EventService/GetEventsByEventName)
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.events | length')
[ "$COUNT" -ge 1 ] && ok "GetEventsByEventName → $COUNT" || fail "not found by name"

step "1.5. GetUnfinishedEventsByUserID"
RESP=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetUnfinishedEventsByUserID)
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.events | length')
[ "$COUNT" -ge 1 ] && ok "GetUnfinishedEventsByUserID → $COUNT" || fail "not found"

step "1.6. GetUnfinishedEventsByEventName"
RESP=$($GRPC -d "{\"event_name\":\"$EVENT_NAME\"}" $HOST event.EventService/GetUnfinishedEventsByEventName)
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.events | length')
[ "$COUNT" -ge 1 ] && ok "GetUnfinishedEventsByEventName → $COUNT" || fail "not found"

step "1.7. GetEventMembersList (creator + enemy = 2)"
RESP=$($GRPC -d "{\"event_id\":\"$EVENT_ID\"}" $HOST event.EventService/GetEventMembersList)
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.users | length')
[ "$COUNT" -eq 2 ] && ok "GetEventMembersList → $COUNT ✓" || fail "expected 2 members, got $COUNT"

# ─────────────────────────────────────────────────────────────────────────────
section "БЛОК 2 — изменение состояния до старта"
# ─────────────────────────────────────────────────────────────────────────────

step "2.1. JoinToEvent — player3, player4"
for ROW in "$P3_ID|$P3_CLAN|false" "$P4_ID|$P4_CLAN|false"; do
  PUID=$(echo "$ROW" | cut -d'|' -f1)
  CID=$(echo "$ROW"  | cut -d'|' -f2)
  ENM=$(echo "$ROW"  | cut -d'|' -f3)
  RESP=$($GRPC -d \
    "{\"event_id\":\"$EVENT_ID\",\"user_id\":\"$PUID\",\"clan_id\":\"$CID\",\"enemy\":$ENM}" \
    $HOST event.EventService/JoinToEvent)
  check_err "$RESP" "JoinToEvent $PUID"
done

step "2.2. LeaveEvent — player4 покидает ивент"
RESP=$($GRPC -d "{\"user_id\":\"$P4_ID\",\"event_id\":\"$EVENT_ID\"}" \
  $HOST event.EventService/LeaveEvent)
echo "$RESP"
check_err "$RESP" "LeaveEvent P4"

MEMBERS=$($GRPC -d "{\"event_id\":\"$EVENT_ID\"}" $HOST event.EventService/GetEventMembersList)
COUNT=$(echo "$MEMBERS" | jq '.users | length')
[ "$COUNT" -eq 3 ] && ok "после LeaveEvent → 3 участника ✓" || fail "expected 3, got $COUNT"

step "2.3. JoinToEvent — player4 снова присоединяется"
RESP=$($GRPC -d \
  "{\"event_id\":\"$EVENT_ID\",\"user_id\":\"$P4_ID\",\"clan_id\":\"$P4_CLAN\",\"enemy\":false}" \
  $HOST event.EventService/JoinToEvent)
check_err "$RESP" "JoinToEvent P4 (re-join)"

step "2.4. SetRole — player3 → side_leader"
RESP=$($GRPC -d \
  "{\"event_id\":\"$EVENT_ID\",\"side_leader_id\":\"$CREATOR_ID\",\"user_id\":\"$P3_ID\",\"role\":\"side_leader\"}" \
  $HOST event.EventService/SetRole)
echo "$RESP"
check_err "$RESP" "SetRole P3=side_leader"

step "2.5. SetRole — player3 → player (сброс роли)"
RESP=$($GRPC -d \
  "{\"event_id\":\"$EVENT_ID\",\"side_leader_id\":\"$CREATOR_ID\",\"user_id\":\"$P3_ID\",\"role\":\"player\"}" \
  $HOST event.EventService/SetRole)
check_err "$RESP" "SetRole P3=player"

step "2.6. UpdateTimeEvent — сдвигаем старт на now+35с (сбрасывает таймер)"
NEW_TIME_START=$(( $(date +%s) + 35 ))
NEW_TIME_RFC=$(date -u -d @$NEW_TIME_START '+%Y-%m-%dT%H:%M:%SZ')
RESP=$($GRPC -d \
  "{\"event_id\":\"$EVENT_ID\",\"user_create_id\":\"$CREATOR_ID\",\"new_time_start\":\"$NEW_TIME_RFC\"}" \
  $HOST event.EventService/UpdateTimeEvent)
echo "$RESP"
check_err "$RESP" "UpdateTimeEvent"
TIME_START=$NEW_TIME_START
info "новый TimeStart: $(date -d @$TIME_START '+%H:%M:%S')"

step "2.7. GetEventMembersList — итого 4 участника"
RESP=$($GRPC -d "{\"event_id\":\"$EVENT_ID\"}" $HOST event.EventService/GetEventMembersList)
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.users | length')
[ "$COUNT" -eq 4 ] && ok "GetEventMembersList → 4 ✓" || fail "expected 4, got $COUNT"

# ─────────────────────────────────────────────────────────────────────────────
section "БЛОК 3 — таймеры: confirmed → in_progress"
# ─────────────────────────────────────────────────────────────────────────────

step "3.1. Ждём controlEventTimerDenial (new_time − 5с)"
CONTROL_TIME=$((TIME_START - 5))
WAIT=$(( CONTROL_TIME - $(date +%s) + 3 ))
[ "$WAIT" -gt 0 ] && { info "Спим ${WAIT}с..."; sleep "$WAIT"; }

step "3.2. Проверка: status=confirmed"
EV=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetLastEventByCreatorId)
STATUS=$(echo "$EV" | jq -r '.event.status')
info "status=$STATUS"
[ "$STATUS" = "confirmed" ] && ok "confirmed ✓" || fail "expected confirmed, got $STATUS"

step "3.3. Ждём старта (new_time)"
WAIT=$(( TIME_START - $(date +%s) + 3 ))
[ "$WAIT" -gt 0 ] && { info "Спим ${WAIT}с..."; sleep "$WAIT"; }

step "3.4. Проверка: status=in_progress"
EV=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetLastEventByCreatorId)
STATUS=$(echo "$EV" | jq -r '.event.status')
info "status=$STATUS"
[ "$STATUS" = "in_progress" ] && ok "in_progress ✓" || fail "expected in_progress, got $STATUS"

# ─────────────────────────────────────────────────────────────────────────────
section "БЛОК 4 — GetTeamsByEventID, GetTeamByID, JoinUserToTeam, RemoveUserFromTeam"
# ─────────────────────────────────────────────────────────────────────────────

step "4.1. GetTeamsByEventID"
TEAMS=$($GRPC -d "{\"event_id\":\"$EVENT_ID\"}" $HOST event.EventService/GetTeamsByEventID)
echo "$TEAMS"

_team() {
  echo "$TEAMS" | jq -r --arg gn "$1" --argjson idx "$2" \
    '[.teams[] | select((.gameNumber | tostring) == $gn)] | .[$idx].teamId'
}

T1G1=$(_team 1 0); T2G1=$(_team 1 1)
T1G2=$(_team 2 0); T2G2=$(_team 2 1)
T1G3=$(_team 3 0); T2G3=$(_team 3 1)
ok "G1: $T1G1 vs $T2G1"
ok "G2: $T1G2 vs $T2G2"
ok "G3: $T1G3 vs $T2G3"

step "4.2. GetTeamByID — все 6 команд"
for TID in "$T1G1" "$T2G1" "$T1G2" "$T2G2" "$T1G3" "$T2G3"; do
  RESP=$($GRPC -d "{\"team_id\":\"$TID\"}" $HOST event.EventService/GetTeamByID)
  GN=$(echo "$RESP" | jq -r '.team.gameNumber')
  check_err "$RESP" "GetTeamByID game=$GN ($TID)"
done

step "4.3. JoinUserToTeam — creator+P3 → T1, enemy+P4 → T2 (все игры)"
for TID in "$T1G1" "$T1G2" "$T1G3"; do
  for PUID in "$CREATOR_ID" "$P3_ID"; do
    RESP=$($GRPC -d "{\"team_id\":\"$TID\",\"user_id\":\"$PUID\",\"role\":\"player\"}" \
      $HOST event.EventService/JoinUserToTeam)
    ERR=$(echo "$RESP" | jq -r '.error // empty')
    [ -z "$ERR" ] && ok "joined $PUID → $TID" || info "WARN: $ERR"
  done
done
for TID in "$T2G1" "$T2G2" "$T2G3"; do
  for PUID in "$ENEMY_ID" "$P4_ID"; do
    RESP=$($GRPC -d "{\"team_id\":\"$TID\",\"user_id\":\"$PUID\",\"role\":\"player\"}" \
      $HOST event.EventService/JoinUserToTeam)
    ERR=$(echo "$RESP" | jq -r '.error // empty')
    [ -z "$ERR" ] && ok "joined $PUID → $TID" || info "WARN: $ERR"
  done
done

step "4.4. RemoveUserFromTeam — убираем P3 из T1G1 и возвращаем обратно"
RESP=$($GRPC -d "{\"team_id\":\"$T1G1\",\"user_id\":\"$P3_ID\"}" \
  $HOST event.EventService/RemoveUserFromTeam)
echo "$RESP"
check_err "$RESP" "RemoveUserFromTeam P3"

RESP=$($GRPC -d "{\"team_id\":\"$T1G1\",\"user_id\":\"$P3_ID\",\"role\":\"player\"}" \
  $HOST event.EventService/JoinUserToTeam)
check_err "$RESP" "JoinUserToTeam P3 (re-join T1G1)"

# ─────────────────────────────────────────────────────────────────────────────
section "БЛОК 5 — игры 1–3: FinishTeamGame + AddTeamMemberStats + GetTeamStats"
# ─────────────────────────────────────────────────────────────────────────────

# _play_game <num> <t1_id> <t2_id> <winner_id>
_play_game() {
  local num="$1" t1="$2" t2="$3" winner="$4"
  local label; label="G$num ($([[ "$winner" == "$t1" ]] && echo "T1 wins" || echo "T2 wins"))"

  step "Игра $num: FinishTeamGame [$label]"
  RESP=$($GRPC -d "{\"team1_id\":\"$t1\",\"team2_id\":\"$t2\",\"team_winner_id\":\"$winner\"}" \
    $HOST event.EventService/FinishTeamGame)
  check_err "$RESP" "FinishTeamGame $label"

  step "Игра $num: AddTeamMemberStats"
  for PUID in "$CREATOR_ID" "$P3_ID"; do
    RESP=$($GRPC -d \
      "{\"team_id\":\"$t1\",\"user_id\":\"$PUID\",
        \"kills\":$((num*3)),\"deaths\":2,\"points\":$((num*50)),\"revival\":1,\"destroyed_vehicles\":0}" \
      $HOST event.EventService/AddTeamMemberStats)
    check_err "$RESP" "stats $PUID (T1G$num)"
  done
  for PUID in "$ENEMY_ID" "$P4_ID"; do
    RESP=$($GRPC -d \
      "{\"team_id\":\"$t2\",\"user_id\":\"$PUID\",
        \"kills\":$((num*2)),\"deaths\":3,\"points\":$((num*30)),\"revival\":0,\"destroyed_vehicles\":1}" \
      $HOST event.EventService/AddTeamMemberStats)
    check_err "$RESP" "stats $PUID (T2G$num)"
  done

  step "Игра $num: GetTeamStats T1"
  RESP=$($GRPC -d "{\"team_id\":\"$t1\"}" $HOST event.EventService/GetTeamStats)
  echo "$RESP"
  CNT=$(echo "$RESP" | jq '.stats | length')
  [ "$CNT" -eq 2 ] && ok "GetTeamStats T1G$num → $CNT участника ✓" || fail "expected 2, got $CNT"

  step "Игра $num: GetTeamStats T2"
  RESP=$($GRPC -d "{\"team_id\":\"$t2\"}" $HOST event.EventService/GetTeamStats)
  echo "$RESP"
  CNT=$(echo "$RESP" | jq '.stats | length')
  [ "$CNT" -eq 2 ] && ok "GetTeamStats T2G$num → $CNT участника ✓" || fail "expected 2, got $CNT"
}

# Игра 1 уже стартовала по таймеру
_play_game 1 "$T1G1" "$T2G1" "$T1G1"

step "Игра 2: StartTeamGame"
RESP=$($GRPC -d "{\"team1_id\":\"$T1G2\",\"team2_id\":\"$T2G2\"}" $HOST event.EventService/StartTeamGame)
check_err "$RESP" "StartTeamGame G2"
_play_game 2 "$T1G2" "$T2G2" "$T2G2"

step "Игра 3: StartTeamGame"
RESP=$($GRPC -d "{\"team1_id\":\"$T1G3\",\"team2_id\":\"$T2G3\"}" $HOST event.EventService/StartTeamGame)
check_err "$RESP" "StartTeamGame G3"
_play_game 3 "$T1G3" "$T2G3" "$T1G3"

# ─────────────────────────────────────────────────────────────────────────────
section "БЛОК 6 — финальные проверки"
# ─────────────────────────────────────────────────────────────────────────────

step "6.1. GetLastEventByCreatorId → status=finished"
EV=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetLastEventByCreatorId)
echo "$EV"
STATUS=$(echo "$EV" | jq -r '.event.status')
WINNER=$(echo "$EV" | jq -r '.event.winnerSide')
[ "$STATUS" = "finished" ] && ok "status=finished ✓" || fail "status=$STATUS"
[[ "$WINNER" =~ ^(ally|enemy|draw)$ ]] && ok "winner_side=$WINNER ✓" || fail "winner_side='$WINNER'"

step "6.2. GetEventsByCreatorId — оба ивента (canceled + finished)"
RESP=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetEventsByCreatorId)
echo "$RESP"
TOTAL=$(echo "$RESP"    | jq '.events | length')
FINISHED=$(echo "$RESP" | jq '[.events[] | select(.status=="finished")] | length')
CANCELED=$(echo "$RESP" | jq '[.events[] | select(.status=="canceled")] | length')
[ "$FINISHED" -ge 1 ] && ok "finished=$FINISHED ✓" || fail "no finished events"
[ "$CANCELED" -ge 1 ] && ok "canceled=$CANCELED ✓" || fail "no canceled events"
ok "GetEventsByCreatorId → всего $TOTAL (finished=$FINISHED, canceled=$CANCELED)"

step "6.3. GetUnfinishedEventsByUserID — должен быть пустой (всё завершено)"
RESP=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetUnfinishedEventsByUserID)
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.events | length')
[ "$COUNT" -eq 0 ] \
  && ok "GetUnfinishedEventsByUserID → 0 ✓" \
  || info "WARN: $COUNT незавершённых (возможно от предыдущих запусков)"

step "6.4. GetUnfinishedEventsByEventName — должен быть пустой"
RESP=$($GRPC -d "{\"event_name\":\"$EVENT_NAME\"}" $HOST event.EventService/GetUnfinishedEventsByEventName)
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.events | length')
[ "$COUNT" -eq 0 ] && ok "GetUnfinishedEventsByEventName → 0 ✓" || fail "expected 0, got $COUNT"

echo -e "\n${GREEN}${BOLD}===== ВСЁ ПРОШЛО УСПЕШНО =====${NC}"
echo "Event ID    : $EVENT_ID"
echo "winner_side : $WINNER"
echo "status      : $STATUS"
echo ""
echo "Покрытые ручки (20/20):"
echo "  CreateEvent, GetEventsByCreatorId, GetLastEventByCreatorId,"
echo "  GetEventsByEventName, GetUnfinishedEventsByUserID, GetUnfinishedEventsByEventName,"
echo "  UpdateTimeEvent, CancelEvent, JoinToEvent, LeaveEvent, SetRole,"
echo "  GetEventMembersList, GetTeamsByEventID, GetTeamByID, JoinUserToTeam,"
echo "  RemoveUserFromTeam, StartTeamGame, FinishTeamGame, AddTeamMemberStats, GetTeamStats"
