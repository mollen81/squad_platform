package com.squad.event.service;

import com.squad.event.client.ClanServiceGrpcClient;
import com.squad.event.kafka.event.MatchFinishedEvent;
import com.squad.event.kafka.event.ServerReadyEvent;
import com.squad.event.kafka.event.ServerStopEvent;
import com.squad.event.kafka.producer.EventKafkaProducer;
import com.squad.event.kafka.event.ServerRentEvent;
import com.squad.event.model.domain.*;
import com.squad.event.model.dto.CreateEventRequest;
import com.squad.event.model.enums.EventMatchStatus;
import com.squad.event.model.enums.EventServerStatus;
import com.squad.event.model.enums.EventStatus;
import com.squad.event.repo.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;
import java.util.stream.IntStream;

@Slf4j
@RequiredArgsConstructor
@Service
public class EventLifecycleManagementService {

    private final EventRepository eventRepository;
    private final EventSideRepository eventSideRepository;
    private final EventParticipantRepository eventParticipantRepository;
    private final EventMatchRepository eventMatchRepository;
    private final EventMatchLayerVoteRepository eventMatchLayerVoteRepository;
    private final EventServerRepository eventServerRepository;
    private final PasswordGeneratorService passwordGeneratorService;

    private final ClanServiceGrpcClient clanServiceGrpcClient;

    private final EventKafkaProducer kafkaProducer;


    @Transactional
    public UUID createEvent(CreateEventRequest request) {
        log.info("Creating new event: {} by user {}", request.name(), request.creatorUserId());

        // business logic checks
        if(request.creatorUserId().equals(request.secondLeaderId())) {
            throw new IllegalArgumentException("One user cannot be leader for both team in one time");
        }

        if(request.timeStart().isBefore(Instant.now())
                || request.timeStart().isAfter(Instant.from(Instant.now().plus(1, ChronoUnit.YEARS)))) {
            throw new IllegalArgumentException("Start time is not in valid interval");
        }

        if(request.targetGameCount() < 1 || request.targetGameCount() > 5) {
            throw new IllegalArgumentException("Games target count is too low or too high");
        }

        // gRPC calls to clan-service
        UUID clanId1 = clanServiceGrpcClient.resolveUserClanId(request.creatorUserId());
        UUID clanId2 = clanServiceGrpcClient.resolveUserClanId(request.secondLeaderId());

        // event form
        Event event = Event.builder()
                .name(request.name())
                .creatorUserId(request.creatorUserId())
                .targetGameCount(request.targetGameCount())
                .timeStart(request.timeStart())
                .gameMode(request.gameMode())
                .status(EventStatus.DRAFT)
                .build();

        eventRepository.save(event);

        // event side form
        EventSide side1 = EventSide.builder()
                .eventId(event.getId())
                .name(request.creatorSideName())
                .leaderUserId(request.creatorUserId())
                .isReady(false)
                .build();

        EventSide side2 = EventSide.builder()
                .eventId(event.getId())
                .name(request.enemySideName())
                .leaderUserId(request.secondLeaderId())
                .isReady(false)
                .build();
        eventSideRepository.save(side1);
        eventSideRepository.save(side2);

        // leaders save in event_participant table
        EventParticipant leader1 = EventParticipant.builder()
                .eventId(event.getId())
                .sideId(side1.getId())
                .userId(side1.getLeaderUserId())
                .clanId(clanId1)
                .build();
        EventParticipant leader2 = EventParticipant.builder()
                .eventId(event.getId())
                .sideId(side2.getId())
                .userId(side2.getLeaderUserId())
                .clanId(clanId2)
                .build();

        eventParticipantRepository.saveAll(List.of(leader1, leader2));

        // matches save in event_match table
        List<EventMatch> matches = IntStream.rangeClosed(1, event.getTargetGameCount())
                .mapToObj(seq -> EventMatch.builder()
                        .eventId(event.getId())
                        .sequenceNumber(seq)
                        .status(EventMatchStatus.PENDING)
                        .build())
                .toList();
        eventMatchRepository.saveAll(matches);

        return event.getId();
    }


    @Transactional
    public void startEventPreparation(UUID eventId) {
        log.info("Starting preparation for event {}", eventId);

        Event event = eventRepository.findByIdForUpdate(eventId)
                .orElseThrow(() -> new IllegalArgumentException("Event " + eventId + " is not found"));

        if(event.getStatus() != EventStatus.REGISTRATION) {
            throw new IllegalStateException("Event " + eventId + " is not in REGISTRATION status");
        }

        event.setStatus(EventStatus.PREPARING);
        eventRepository.save(event);

        List<EventMatch> matches = eventMatchRepository.findAllByEventIdOrderBySequenceNumber(eventId);
        for(EventMatch match : matches) {
            String winnerMapName = eventMatchLayerVoteRepository.findWinnerLayerByMatchId(match.getId())
                    .orElse("Gorodok_RAAS_V1");

            match.setLayerName(winnerMapName);
            eventMatchRepository.save(match);
        }

        EventServer server = EventServer.builder()
                .eventId(eventId)
                .status(EventServerStatus.PENDING)
                .build();
        eventServerRepository.save(server);

        log.info("Preparation completed for event {}. Ready for server deployment", eventId);

        List<String> layerNames = matches.stream()
                .map(EventMatch::getLayerName)
                .toList();
        List<ServerRentEvent.AdminEntry> admins = List.of(new ServerRentEvent.AdminEntry(
                "admin",
                "76561198888277695"));
        List<String> whiteList = new ArrayList<>(eventParticipantRepository.findAllParticipantSteamIdsByEventId(eventId));
        whiteList.add("76561198888277695");

        List<UUID> sideIds = new ArrayList<>(eventSideRepository.findAllByEventId(eventId).stream()
                .map(EventSide::getId)
                .toList());

        ServerRentEvent serverRentEvent = new ServerRentEvent(
                eventId,
                event.getTimeStart(),
                new ServerRentEvent.ServerConfig(
                        event.getName(),
                        passwordGeneratorService.generatePassword(15),
                        layerNames,
                        100
                ),
                admins,
                whiteList,
                sideIds
        );
        kafkaProducer.sendServerRentEvent(serverRentEvent);
    }


    @Transactional
    public void handleServerReady(ServerReadyEvent readyEvent) {
        log.info("Received server.ready event for eventId: {}", readyEvent.eventId());

        EventServer server = eventServerRepository.findByEventId(readyEvent.eventId())
                .orElseThrow(() -> new IllegalArgumentException(
                        "EventServer for event " + readyEvent.eventId() + " is not found"));

        if(server.getStatus() == EventServerStatus.READY) {
            log.warn(
                    "Server for event {} is already in READY status. Ignoring duplicate message",
                    readyEvent.eventId()
            );
        }

        server.setIpAddress(readyEvent.ipAddress());
        server.setQueryPort(readyEvent.queryPort());
        server.setPassword(readyEvent.password());
        server.setStatus(EventServerStatus.READY);

        eventServerRepository.save(server);

        log.info("Server for event {} is now READY ({}:{})",
                readyEvent.eventId(), readyEvent.ipAddress(), readyEvent.queryPort());
    }


    @Transactional
    public void handleMatchFinished(MatchFinishedEvent finishedEvent) {
        UUID eventId = finishedEvent.eventId();
        log.info("Processing match finish for event {}, match #{}", eventId, finishedEvent.matchSequenceNumber());

        Event event = eventRepository.findByIdForUpdate(eventId)
                .orElseThrow(() -> new IllegalArgumentException("Event " + eventId + " is not found"));

        List<EventMatch> matches = eventMatchRepository.findAllByEventIdOrderBySequenceNumber(eventId);

        EventMatch currentMatch = matches.get(finishedEvent.matchSequenceNumber() - 1);

        if(currentMatch.getStatus() == EventMatchStatus.FINISHED) {
            log.warn("Match #{} for event {} is already FINISHED. Ignoring duplicate event.",
                    finishedEvent.matchSequenceNumber(), eventId);
            return;
        }

        if(event.getStatus() != EventStatus.LIVE) {
            throw new IllegalStateException("Cannot finish match: Event " + eventId + " is not in LIVE status");
        }

        Instant now = Instant.now();
        currentMatch.setStatus(EventMatchStatus.FINISHED);
        currentMatch.setWinnerSideId(finishedEvent.winnerSideId());
        currentMatch.setFinishedAt(now);

        int winsRequired = (event.getTargetGameCount() / 2) + 1;

        long winnerTotalWins = matches.stream()
                .filter(m -> m.getStatus() == EventMatchStatus.FINISHED)
                .filter(m -> finishedEvent.winnerSideId() != null && finishedEvent.winnerSideId().equals(m.getWinnerSideId()))
                .count();

        boolean isLastMatchInSchedule = currentMatch.getSequenceNumber() >= event.getTargetGameCount();
        boolean isSeriesDecided = (winnerTotalWins >= winsRequired) || isLastMatchInSchedule;

        if(isSeriesDecided) {
            log.info("Series decided for event {} (Winner side {} has {} wins). Finishing event",
                    eventId, finishedEvent.winnerSideId(), winnerTotalWins);

            event.setStatus(EventStatus.FINISHED);
            eventRepository.save(event);

            matches.stream()
                    .filter(m -> m.getStatus() == EventMatchStatus.PENDING)
                    .forEach(m -> m.setStatus(EventMatchStatus.CANCELED));

            eventServerRepository.findByEventId(eventId).ifPresent(server -> {
                server.setStatus(EventServerStatus.STOPPED);
                eventServerRepository.save(server);
            });

            kafkaProducer.sendServerStopEvent(new ServerStopEvent(eventId));
        }
        else {
            int nextSeq = currentMatch.getSequenceNumber() + 1;
            EventMatch nextMatch = matches.stream()
                    .filter(m -> m.getSequenceNumber() == nextSeq)
                    .findFirst()
                    .orElseThrow(() -> new IllegalStateException("Next match #" + nextSeq + " not found"));

            nextMatch.setStatus(EventMatchStatus.IN_PROGRESS);
            nextMatch.setStartedAt(now);

            log.info("Match #{} started for event {} on layer {}", nextSeq, eventId, nextMatch.getLayerName());
        }

        eventMatchRepository.saveAll(matches);
    }
}
