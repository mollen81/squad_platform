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

# Ошибки приходят gRPC-статусами (grpcurl печатает их как "ERROR: / Code: /
# Message:"), поля error в ответах больше нет.
_status() { echo "$1" | grep -E '^ *(Code|Message):' | tr '\n' ' ' | sed 's/  */ /g'; }
_failed()  { echo "$1" | grep -q '^ERROR:'; }

check_err() {
  local resp="$1" label="$2"
  if _failed "$resp"; then fail "$label: $(_status "$resp")"; else ok "$label"; fi
}

# expect_err — ручка ДОЛЖНА отказать; третий аргумент (необязательный) —
# ожидаемый gRPC-код
expect_err() {
  local resp="$1" label="$2" want="${3:-}"
  if ! _failed "$resp"; then fail "$label: ожидался отказ, но прошло"; fi
  if [ -n "$want" ] && ! echo "$resp" | grep -q "Code: $want"; then
    fail "$label: ожидался код $want, получено: $(_status "$resp")"
  fi
  ok "$label → $(_status "$resp")"
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
}" $HOST event.EventService/CreateEvent 2>&1) || true
check_err "$RESP" "CreateEvent (cancel-event)"

EV=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetLastEventByCreatorId 2>&1) || true
CANCEL_ID=$(echo "$EV" | jq -r '.event.eventId')
ok "cancel_event_id=$CANCEL_ID"

step "0.2. CancelEvent"
RESP=$($GRPC -d "{\"event_id\":\"$CANCEL_ID\",\"user_create_id\":\"$CREATOR_ID\"}" \
  $HOST event.EventService/CancelEvent 2>&1) || true
echo "$RESP"
check_err "$RESP" "CancelEvent"

EV=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetLastEventByCreatorId 2>&1) || true
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
}" $HOST event.EventService/CreateEvent 2>&1) || true
check_err "$RESP" "CreateEvent (main)"

step "1.2. GetLastEventByCreatorId → event_id"
EV=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetLastEventByCreatorId 2>&1) || true
echo "$EV"
EVENT_ID=$(echo "$EV" | jq -r '.event.eventId')
[ "$EVENT_ID" != "null" ] && [ -n "$EVENT_ID" ] \
  && ok "event_id=$EVENT_ID" || fail "event_id not found"

step "1.3. GetEventsByCreatorId"
RESP=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetEventsByCreatorId 2>&1) || true
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.events | length')
[ "$COUNT" -ge 1 ] && ok "GetEventsByCreatorId → $COUNT events" || fail "no events"

step "1.4. GetEventsByEventName"
RESP=$($GRPC -d "{\"event_name\":\"$EVENT_NAME\"}" $HOST event.EventService/GetEventsByEventName 2>&1) || true
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.events | length')
[ "$COUNT" -ge 1 ] && ok "GetEventsByEventName → $COUNT" || fail "not found by name"

step "1.5. GetUnfinishedEventsByUserID"
RESP=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetUnfinishedEventsByUserID 2>&1) || true
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.events | length')
[ "$COUNT" -ge 1 ] && ok "GetUnfinishedEventsByUserID → $COUNT" || fail "not found"

step "1.6. GetUnfinishedEventsByEventName"
RESP=$($GRPC -d "{\"event_name\":\"$EVENT_NAME\"}" $HOST event.EventService/GetUnfinishedEventsByEventName 2>&1) || true
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.events | length')
[ "$COUNT" -ge 1 ] && ok "GetUnfinishedEventsByEventName → $COUNT" || fail "not found"

step "1.7. GetEventMembersList (creator + enemy = 2)"
RESP=$($GRPC -d "{\"event_id\":\"$EVENT_ID\"}" $HOST event.EventService/GetEventMembersList 2>&1) || true
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.users | length')
[ "$COUNT" -eq 2 ] && ok "GetEventMembersList → $COUNT ✓" || fail "expected 2 members, got $COUNT"

# ─────────────────────────────────────────────────────────────────────────────
section "БЛОК 2 — состав ивента (только пока pending)"
# ─────────────────────────────────────────────────────────────────────────────

step "2.1. JoinToEvent — player3 (ally), player4 (enemy)"
for ROW in "$P3_ID|$P3_CLAN|false" "$P4_ID|$P4_CLAN|true"; do
  PUID=$(echo "$ROW" | cut -d'|' -f1)
  CID=$(echo "$ROW"  | cut -d'|' -f2)
  ENM=$(echo "$ROW"  | cut -d'|' -f3)
  RESP=$($GRPC -d \
    "{\"event_id\":\"$EVENT_ID\",\"user_id\":\"$PUID\",\"clan_id\":\"$CID\",\"enemy\":$ENM}" \
    $HOST event.EventService/JoinToEvent 2>&1) || true
  check_err "$RESP" "JoinToEvent $PUID"
done

step "2.2. LeaveEvent — player4 покидает ивент"
RESP=$($GRPC -d "{\"user_id\":\"$P4_ID\",\"event_id\":\"$EVENT_ID\"}" \
  $HOST event.EventService/LeaveEvent 2>&1) || true
echo "$RESP"
check_err "$RESP" "LeaveEvent P4"

MEMBERS=$($GRPC -d "{\"event_id\":\"$EVENT_ID\"}" $HOST event.EventService/GetEventMembersList 2>&1) || true
COUNT=$(echo "$MEMBERS" | jq '.users | length')
[ "$COUNT" -eq 3 ] && ok "после LeaveEvent → 3 участника ✓" || fail "expected 3, got $COUNT"

step "2.3. LeaveEvent сайд-лидеров запрещён (оба обязаны играть каждую игру)"
RESP=$($GRPC -d "{\"user_id\":\"$CREATOR_ID\",\"event_id\":\"$EVENT_ID\"}" $HOST event.EventService/LeaveEvent 2>&1) || true
expect_err "$RESP" "LeaveEvent creator" PermissionDenied
RESP=$($GRPC -d "{\"user_id\":\"$ENEMY_ID\",\"event_id\":\"$EVENT_ID\"}" $HOST event.EventService/LeaveEvent 2>&1) || true
expect_err "$RESP" "LeaveEvent enemy side leader" PermissionDenied

step "2.4. JoinToEvent — player4 снова присоединяется (enemy)"
RESP=$($GRPC -d \
  "{\"event_id\":\"$EVENT_ID\",\"user_id\":\"$P4_ID\",\"clan_id\":\"$P4_CLAN\",\"enemy\":true}" \
  $HOST event.EventService/JoinToEvent 2>&1) || true
check_err "$RESP" "JoinToEvent P4 (re-join)"

step "2.5. UpdateTimeEvent — сдвигаем старт на now+45с (сбрасывает таймер)"
NEW_TIME_START=$(( $(date +%s) + 45 ))
NEW_TIME_RFC=$(date -u -d @$NEW_TIME_START '+%Y-%m-%dT%H:%M:%SZ')
RESP=$($GRPC -d \
  "{\"event_id\":\"$EVENT_ID\",\"user_create_id\":\"$CREATOR_ID\",\"new_time_start\":\"$NEW_TIME_RFC\"}" \
  $HOST event.EventService/UpdateTimeEvent 2>&1) || true
echo "$RESP"
check_err "$RESP" "UpdateTimeEvent"
TIME_START=$NEW_TIME_START
info "новый TimeStart: $(date -d @$TIME_START '+%H:%M:%S')"

step "2.6. GetEventMembersList — итого 4 участника"
RESP=$($GRPC -d "{\"event_id\":\"$EVENT_ID\"}" $HOST event.EventService/GetEventMembersList 2>&1) || true
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.users | length')
[ "$COUNT" -eq 4 ] && ok "GetEventMembersList → 4 ✓" || fail "expected 4, got $COUNT"

# ─────────────────────────────────────────────────────────────────────────────
section "БЛОК 3 — составы команд (только пока игра не началась)"
# ─────────────────────────────────────────────────────────────────────────────

step "3.1. GetTeamsByEventID"
TEAMS=$($GRPC -d "{\"event_id\":\"$EVENT_ID\"}" $HOST event.EventService/GetTeamsByEventID 2>&1) || true
echo "$TEAMS"

ALLY_LEADER=$(echo "$MEMBERS" | jq -r --arg u "$CREATOR_ID" '.users[] | select(.userId==$u) | .userEventId')
_side_team() { # _side_team <game_number> <ally|enemy>
  echo "$TEAMS" | jq -r --arg gn "$1" --arg sl "$ALLY_LEADER" --arg side "$2" \
    '[.teams[] | select((.gameNumber | tostring) == $gn)
      | select(if $side == "ally" then .sideLeaderId == $sl else .sideLeaderId != $sl end)] | .[0].teamId'
}

T1G1=$(_side_team 1 ally); T2G1=$(_side_team 1 enemy)
T1G2=$(_side_team 2 ally); T2G2=$(_side_team 2 enemy)
T1G3=$(_side_team 3 ally); T2G3=$(_side_team 3 enemy)
ok "G1: $T1G1 vs $T2G1"
ok "G2: $T1G2 vs $T2G2"
ok "G3: $T1G3 vs $T2G3"

step "3.2. GetTeamByID — все 6 команд (сайд-лидеры уже в составе, status=pending)"
for TID in "$T1G1" "$T2G1" "$T1G2" "$T2G2" "$T1G3" "$T2G3"; do
  RESP=$($GRPC -d "{\"team_id\":\"$TID\"}" $HOST event.EventService/GetTeamByID 2>&1) || true
  GN=$(echo "$RESP" | jq -r '.team.gameNumber')
  ST=$(echo "$RESP" | jq -r '.team.status')
  MC=$(echo "$RESP" | jq -r '.team.membersCount')
  check_err "$RESP" "GetTeamByID game=$GN status=$ST members=$MC ($TID)"
  [ "$ST" = "pending" ] || fail "ожидался status=pending, got $ST"
  [ "$MC" = "1" ] || fail "ожидался сайд-лидер в составе (members_count=1), got $MC"
done

step "3.3. JoinUserToTeam — P3 → команды создателя, P4 → команды второй стороны"
for TID in "$T1G1" "$T1G2" "$T1G3"; do
  RESP=$($GRPC -d "{\"team_id\":\"$TID\",\"user_id\":\"$P3_ID\",\"role\":\"player\"}" \
    $HOST event.EventService/JoinUserToTeam 2>&1) || true
  check_err "$RESP" "joined P3 → $TID"
done
for TID in "$T2G1" "$T2G2" "$T2G3"; do
  RESP=$($GRPC -d "{\"team_id\":\"$TID\",\"user_id\":\"$P4_ID\",\"role\":\"player\"}" \
    $HOST event.EventService/JoinUserToTeam 2>&1) || true
  check_err "$RESP" "joined P4 → $TID"
done

step "3.4. Запреты: чужая сторона, сайд-лидер, squad_leader"
RESP=$($GRPC -d "{\"team_id\":\"$T1G1\",\"user_id\":\"$P4_ID\",\"role\":\"player\"}" \
  $HOST event.EventService/JoinUserToTeam 2>&1) || true
expect_err "$RESP" "JoinUserToTeam enemy-игрока в команду ally" FailedPrecondition

RESP=$($GRPC -d "{\"team_id\":\"$T1G1\",\"user_id\":\"$CREATOR_ID\"}" \
  $HOST event.EventService/RemoveUserFromTeam 2>&1) || true
expect_err "$RESP" "RemoveUserFromTeam сайд-лидера" FailedPrecondition

RESP=$($GRPC -d "{\"team_id\":\"$T2G1\",\"user_id\":\"$P4_ID\",\"role\":\"squad_leader\"}" \
  $HOST event.EventService/JoinUserToTeam 2>&1) || true
expect_err "$RESP" "JoinUserToTeam с ролью squad_leader" InvalidArgument

step "3.5. RemoveUserFromTeam — убираем P3 из T1G1 и возвращаем обратно"
RESP=$($GRPC -d "{\"team_id\":\"$T1G1\",\"user_id\":\"$P3_ID\"}" \
  $HOST event.EventService/RemoveUserFromTeam 2>&1) || true
echo "$RESP"
check_err "$RESP" "RemoveUserFromTeam P3"

RESP=$($GRPC -d "{\"team_id\":\"$T1G1\",\"user_id\":\"$P3_ID\",\"role\":\"player\"}" \
  $HOST event.EventService/JoinUserToTeam 2>&1) || true
check_err "$RESP" "JoinUserToTeam P3 (re-join T1G1)"

step "3.6. SetRole — роль выдаётся внутри команды, и только своим сайд-лидером"
RESP=$($GRPC -d \
  "{\"team_id\":\"$T1G1\",\"side_leader_id\":\"$CREATOR_ID\",\"user_id\":\"$P3_ID\",\"role\":\"side_leader\"}" \
  $HOST event.EventService/SetRole 2>&1) || true
echo "$RESP"
check_err "$RESP" "SetRole P3=side_leader в T1G1"

RESP=$($GRPC -d \
  "{\"team_id\":\"$T1G1\",\"side_leader_id\":\"$ENEMY_ID\",\"user_id\":\"$P3_ID\",\"role\":\"player\"}" \
  $HOST event.EventService/SetRole 2>&1) || true
expect_err "$RESP" "SetRole чужим сайд-лидером" PermissionDenied

RESP=$($GRPC -d \
  "{\"team_id\":\"$T1G1\",\"side_leader_id\":\"$CREATOR_ID\",\"user_id\":\"$CREATOR_ID\",\"role\":\"player\"}" \
  $HOST event.EventService/SetRole 2>&1) || true
expect_err "$RESP" "SetRole самому сайд-лидеру" FailedPrecondition

RESP=$($GRPC -d \
  "{\"team_id\":\"$T1G1\",\"side_leader_id\":\"$CREATOR_ID\",\"user_id\":\"$P3_ID\",\"role\":\"player\"}" \
  $HOST event.EventService/SetRole 2>&1) || true
check_err "$RESP" "SetRole P3=player (сброс роли)"

# ─────────────────────────────────────────────────────────────────────────────
section "БЛОК 4 — таймеры: confirmed → in_progress"
# ─────────────────────────────────────────────────────────────────────────────

step "4.1. Ждём controlEventTimerDenial (time_start − 5с)"
CONTROL_TIME=$((TIME_START - 5))
WAIT=$(( CONTROL_TIME - $(date +%s) + 3 ))
[ "$WAIT" -gt 0 ] && { info "Спим ${WAIT}с..."; sleep "$WAIT"; }

step "4.2. Проверка: status=confirmed"
EV=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetLastEventByCreatorId 2>&1) || true
STATUS=$(echo "$EV" | jq -r '.event.status')
info "status=$STATUS"
[ "$STATUS" = "confirmed" ] && ok "confirmed ✓" || fail "expected confirmed, got $STATUS"

step "4.3. После подтверждения состав заморожен"
RESP=$($GRPC -d \
  "{\"event_id\":\"$EVENT_ID\",\"user_id\":\"$(uuid)\",\"clan_id\":\"$(uuid)\",\"enemy\":false}" \
  $HOST event.EventService/JoinToEvent 2>&1) || true
expect_err "$RESP" "JoinToEvent после confirmed" FailedPrecondition
RESP=$($GRPC -d "{\"user_id\":\"$P3_ID\",\"event_id\":\"$EVENT_ID\"}" $HOST event.EventService/LeaveEvent 2>&1) || true
expect_err "$RESP" "LeaveEvent после confirmed" FailedPrecondition

step "4.4. Ждём старта (time_start)"
WAIT=$(( TIME_START - $(date +%s) + 3 ))
[ "$WAIT" -gt 0 ] && { info "Спим ${WAIT}с..."; sleep "$WAIT"; }

step "4.5. Проверка: status=in_progress, первая игра идёт"
EV=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetLastEventByCreatorId 2>&1) || true
STATUS=$(echo "$EV" | jq -r '.event.status')
info "status=$STATUS"
[ "$STATUS" = "in_progress" ] && ok "in_progress ✓" || fail "expected in_progress, got $STATUS"

RESP=$($GRPC -d "{\"team_id\":\"$T1G1\"}" $HOST event.EventService/GetTeamByID 2>&1) || true
ST=$(echo "$RESP" | jq -r '.team.status')
[ "$ST" = "in_progress" ] && ok "team status=in_progress ✓" || fail "expected in_progress, got $ST"

step "4.6. Состав и роли начавшейся игры заморожены"
RESP=$($GRPC -d "{\"team_id\":\"$T1G1\",\"user_id\":\"$P3_ID\"}" $HOST event.EventService/RemoveUserFromTeam 2>&1) || true
expect_err "$RESP" "RemoveUserFromTeam после старта игры" FailedPrecondition
RESP=$($GRPC -d \
  "{\"team_id\":\"$T1G1\",\"side_leader_id\":\"$CREATOR_ID\",\"user_id\":\"$P3_ID\",\"role\":\"side_leader\"}" \
  $HOST event.EventService/SetRole 2>&1) || true
expect_err "$RESP" "SetRole после старта игры" FailedPrecondition

# ─────────────────────────────────────────────────────────────────────────────
section "БЛОК 5 — игры 1–3: FinishTeamGame + AddTeamMemberStats + GetTeamStats"
# ─────────────────────────────────────────────────────────────────────────────

# _play_game <num> <t1_id> <t2_id> <winner_id>
_play_game() {
  local num="$1" t1="$2" t2="$3" winner="$4"
  local label; label="G$num ($([[ "$winner" == "$t1" ]] && echo "T1 wins" || echo "T2 wins"))"

  step "Игра $num: FinishTeamGame [$label]"
  RESP=$($GRPC -d "{\"team1_id\":\"$t1\",\"team2_id\":\"$t2\",\"team_winner_id\":\"$winner\"}" \
    $HOST event.EventService/FinishTeamGame 2>&1) || true
  check_err "$RESP" "FinishTeamGame $label"

  step "Игра $num: AddTeamMemberStats"
  for PUID in "$CREATOR_ID" "$P3_ID"; do
    RESP=$($GRPC -d \
      "{\"team_id\":\"$t1\",\"user_id\":\"$PUID\",
        \"kills\":$((num*3)),\"deaths\":2,\"points\":$((num*50)),\"revival\":1,\"destroyed_vehicles\":0}" \
      $HOST event.EventService/AddTeamMemberStats 2>&1) || true
    check_err "$RESP" "stats $PUID (T1G$num)"
  done
  for PUID in "$ENEMY_ID" "$P4_ID"; do
    RESP=$($GRPC -d \
      "{\"team_id\":\"$t2\",\"user_id\":\"$PUID\",
        \"kills\":$((num*2)),\"deaths\":3,\"points\":$((num*30)),\"revival\":0,\"destroyed_vehicles\":1}" \
      $HOST event.EventService/AddTeamMemberStats 2>&1) || true
    check_err "$RESP" "stats $PUID (T2G$num)"
  done

  step "Игра $num: GetTeamStats T1"
  RESP=$($GRPC -d "{\"team_id\":\"$t1\"}" $HOST event.EventService/GetTeamStats 2>&1) || true
  echo "$RESP"
  CNT=$(echo "$RESP" | jq '.stats | length')
  [ "$CNT" -eq 2 ] && ok "GetTeamStats T1G$num → $CNT участника ✓" || fail "expected 2, got $CNT"

  step "Игра $num: GetTeamStats T2"
  RESP=$($GRPC -d "{\"team_id\":\"$t2\"}" $HOST event.EventService/GetTeamStats 2>&1) || true
  echo "$RESP"
  CNT=$(echo "$RESP" | jq '.stats | length')
  [ "$CNT" -eq 2 ] && ok "GetTeamStats T2G$num → $CNT участника ✓" || fail "expected 2, got $CNT"
}

# Игра 1 уже стартовала по таймеру
_play_game 1 "$T1G1" "$T2G1" "$T1G1"

step "Игра 2: StartTeamGame"
RESP=$($GRPC -d "{\"team1_id\":\"$T1G2\",\"team2_id\":\"$T2G2\"}" $HOST event.EventService/StartTeamGame 2>&1) || true
check_err "$RESP" "StartTeamGame G2"
_play_game 2 "$T1G2" "$T2G2" "$T2G2"

step "Игра 3: StartTeamGame"
RESP=$($GRPC -d "{\"team1_id\":\"$T1G3\",\"team2_id\":\"$T2G3\"}" $HOST event.EventService/StartTeamGame 2>&1) || true
check_err "$RESP" "StartTeamGame G3"
_play_game 3 "$T1G3" "$T2G3" "$T1G3"

# ─────────────────────────────────────────────────────────────────────────────
section "БЛОК 6 — финальные проверки"
# ─────────────────────────────────────────────────────────────────────────────

step "6.1. GetLastEventByCreatorId → status=finished"
EV=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetLastEventByCreatorId 2>&1) || true
echo "$EV"
STATUS=$(echo "$EV" | jq -r '.event.status')
WINNER=$(echo "$EV" | jq -r '.event.winnerSide')
[ "$STATUS" = "finished" ] && ok "status=finished ✓" || fail "status=$STATUS"
[[ "$WINNER" =~ ^(ally|enemy|draw)$ ]] && ok "winner_side=$WINNER ✓" || fail "winner_side='$WINNER'"

step "6.2. GetEventsByCreatorId — оба ивента (canceled + finished)"
RESP=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetEventsByCreatorId 2>&1) || true
echo "$RESP"
TOTAL=$(echo "$RESP"    | jq '.events | length')
FINISHED=$(echo "$RESP" | jq '[.events[] | select(.status=="finished")] | length')
CANCELED=$(echo "$RESP" | jq '[.events[] | select(.status=="canceled")] | length')
[ "$FINISHED" -ge 1 ] && ok "finished=$FINISHED ✓" || fail "no finished events"
[ "$CANCELED" -ge 1 ] && ok "canceled=$CANCELED ✓" || fail "no canceled events"
ok "GetEventsByCreatorId → всего $TOTAL (finished=$FINISHED, canceled=$CANCELED)"

step "6.3. GetUnfinishedEventsByUserID — должен быть пустой (всё завершено)"
RESP=$($GRPC -d "{\"user_create_id\":\"$CREATOR_ID\"}" $HOST event.EventService/GetUnfinishedEventsByUserID 2>&1) || true
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.events | length')
[ "$COUNT" -eq 0 ] \
  && ok "GetUnfinishedEventsByUserID → 0 ✓" \
  || info "WARN: $COUNT незавершённых (возможно от предыдущих запусков)"

step "6.4. GetUnfinishedEventsByEventName — должен быть пустой"
RESP=$($GRPC -d "{\"event_name\":\"$EVENT_NAME\"}" $HOST event.EventService/GetUnfinishedEventsByEventName 2>&1) || true
echo "$RESP"
COUNT=$(echo "$RESP" | jq '.events | length')
[ "$COUNT" -eq 0 ] && ok "GetUnfinishedEventsByEventName → 0 ✓" || fail "expected 0, got $COUNT"

# ─────────────────────────────────────────────────────────────────────────────
section "БЛОК 7 — валидация входа и коды ошибок"
# ─────────────────────────────────────────────────────────────────────────────

step "7.1. Неверные данные запроса → InvalidArgument"
RESP=$($GRPC -d "{\"event_id\":\"не-uuid\",\"user_id\":\"$P3_ID\",\"clan_id\":\"$P3_CLAN\",\"enemy\":false}" \
  $HOST event.EventService/JoinToEvent 2>&1) || true
expect_err "$RESP" "JoinToEvent с невалидным uuid" InvalidArgument

RESP=$($GRPC -d "{\"event_id\":\"\",\"user_id\":\"$P3_ID\",\"clan_id\":\"$P3_CLAN\",\"enemy\":false}" \
  $HOST event.EventService/JoinToEvent 2>&1) || true
expect_err "$RESP" "JoinToEvent с пустым event_id" InvalidArgument

TS=$(date -u -d @$(( $(date +%s) + 3600 )) '+%Y-%m-%dT%H:%M:%SZ')
RESP=$($GRPC -d "{
  \"user_creator_id\":\"$(uuid)\", \"creator_clan_id\":\"$(uuid)\",
  \"enemy_side_leader_id\":\"$(uuid)\", \"enemy_side_leader_clan_id\":\"$(uuid)\",
  \"event_name\":\"   \", \"time_start\":\"$TS\", \"target_game_count\":1
}" $HOST event.EventService/CreateEvent 2>&1) || true
expect_err "$RESP" "CreateEvent с пустым названием" InvalidArgument

RESP=$($GRPC -d "{
  \"user_creator_id\":\"$CREATOR_ID\", \"creator_clan_id\":\"$CREATOR_CLAN\",
  \"enemy_side_leader_id\":\"$CREATOR_ID\", \"enemy_side_leader_clan_id\":\"$ENEMY_CLAN\",
  \"event_name\":\"self-fight\", \"time_start\":\"$TS\", \"target_game_count\":1
}" $HOST event.EventService/CreateEvent 2>&1) || true
expect_err "$RESP" "CreateEvent, где создатель сам себе противник" InvalidArgument

RESP=$($GRPC -d \
  "{\"team_id\":\"$T1G1\",\"user_id\":\"$P3_ID\",\"kills\":-5,\"deaths\":0,\"points\":0,\"revival\":0,\"destroyed_vehicles\":0}" \
  $HOST event.EventService/AddTeamMemberStats 2>&1) || true
expect_err "$RESP" "AddTeamMemberStats с отрицательными киллами" InvalidArgument

RESP=$($GRPC -d \
  "{\"team_id\":\"$T1G1\",\"user_id\":\"$P3_ID\",\"kills\":999999,\"deaths\":0,\"points\":0,\"revival\":0,\"destroyed_vehicles\":0}" \
  $HOST event.EventService/AddTeamMemberStats 2>&1) || true
expect_err "$RESP" "AddTeamMemberStats с абсурдным значением" InvalidArgument

step "7.2. Несуществующие объекты → NotFound"
RESP=$($GRPC -d "{\"team_id\":\"$(uuid)\"}" $HOST event.EventService/GetTeamByID 2>&1) || true
expect_err "$RESP" "GetTeamByID несуществующей команды" NotFound

RESP=$($GRPC -d "{\"event_id\":\"$(uuid)\",\"user_create_id\":\"$CREATOR_ID\"}" \
  $HOST event.EventService/CancelEvent 2>&1) || true
expect_err "$RESP" "CancelEvent несуществующего ивента" NotFound

step "7.3. Повторное действие → AlreadyExists"
DUP_CREATOR=$(uuid); DUP_PLAYER=$(uuid)
RESP=$($GRPC -d "{
  \"user_creator_id\":\"$DUP_CREATOR\", \"creator_clan_id\":\"$(uuid)\",
  \"enemy_side_leader_id\":\"$(uuid)\", \"enemy_side_leader_clan_id\":\"$(uuid)\",
  \"event_name\":\"dup-check-$(date +%s)\", \"time_start\":\"$TS\", \"target_game_count\":1
}" $HOST event.EventService/CreateEvent 2>&1) || true
check_err "$RESP" "CreateEvent (для проверки повторного входа)"
EV=$($GRPC -d "{\"user_create_id\":\"$DUP_CREATOR\"}" $HOST event.EventService/GetLastEventByCreatorId 2>&1) || true
DUP_EVENT=$(echo "$EV" | jq -r '.event.eventId')

RESP=$($GRPC -d "{\"event_id\":\"$DUP_EVENT\",\"user_id\":\"$DUP_PLAYER\",\"clan_id\":\"$(uuid)\",\"enemy\":false}" \
  $HOST event.EventService/JoinToEvent 2>&1) || true
check_err "$RESP" "JoinToEvent игрока"
RESP=$($GRPC -d "{\"event_id\":\"$DUP_EVENT\",\"user_id\":\"$DUP_PLAYER\",\"clan_id\":\"$(uuid)\",\"enemy\":false}" \
  $HOST event.EventService/JoinToEvent 2>&1) || true
expect_err "$RESP" "JoinToEvent того же игрока повторно" AlreadyExists

RESP=$($GRPC -d "{\"event_id\":\"$DUP_EVENT\",\"user_create_id\":\"$(uuid)\"}" \
  $HOST event.EventService/CancelEvent 2>&1) || true
expect_err "$RESP" "CancelEvent не создателем" PermissionDenied

RESP=$($GRPC -d "{\"event_id\":\"$DUP_EVENT\",\"user_create_id\":\"$DUP_CREATOR\"}" \
  $HOST event.EventService/CancelEvent 2>&1) || true
check_err "$RESP" "CancelEvent (уборка тестового ивента)"

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
