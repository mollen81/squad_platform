package com.squad.event.service;

import com.squad.event.client.ClanServiceGrpcClient;
import com.squad.event.grpc.ClanServiceGrpc;
import com.squad.event.grpc.GetClanIdFromUserIdRequest;
import com.squad.event.kafka.EventKafkaProducer;
import com.squad.event.kafka.event.ServerRentEvent;
import com.squad.event.model.domain.*;
import com.squad.event.model.dto.CreateEventRequest;
import com.squad.event.model.enums.EventMatchStatus;
import com.squad.event.model.enums.EventServerStatus;
import com.squad.event.model.enums.EventStatus;
import com.squad.event.repo.*;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.time.temporal.ChronoUnit;
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

    private final ClanServiceGrpcClient clanServiceGrpcClient;

    private final EventKafkaProducer kafkaProducer;

    private final PasswordGeneratorService passwordGeneratorService;

    private static final String TOPIC_SERVER_RENT = "server.rent";

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

        // TODO Kafka producer (server.rent)
        List<EventMatch> layerList = eventMatchRepository.findAllByEventIdOrderBySequenceNumber(eventId);
        List<ServerRentEvent.AdminEntry> admins = List.of(new ServerRentEvent.AdminEntry(
                "admin",
                "76561198888277695"));
        List<String> whiteList = eventParticipantRepository.findAllParticipantSteamIdsByEventId(eventId);
        whiteList.add("76561198888277695");

        ServerRentEvent serverRentEvent = new ServerRentEvent(
                TOPIC_SERVER_RENT,
                eventId,
                event.getTimeStart(),
                new ServerRentEvent.ServerConfig(
                        event.getName(),
                        passwordGeneratorService.generatePassword(15),
                        layerList,
                        100
                ),
                admins,
                whiteList
        );
        kafkaProducer.sendServerRentEvent(serverRentEvent);
    }
}
