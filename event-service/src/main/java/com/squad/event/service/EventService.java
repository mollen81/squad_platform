package com.squad.event.service;

import com.squad.event.model.domain.*;
import com.squad.event.model.dto.CreateEventRequest;
import com.squad.event.model.dto.JoinEventRequest;
import com.squad.event.model.enums.EventMatchMap;
import com.squad.event.model.enums.EventMatchStatus;
import com.squad.event.model.enums.EventStatus;
import com.squad.event.repo.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.List;
import java.util.UUID;

@Service
@Slf4j
@RequiredArgsConstructor
public class EventService {

    private final EventRepository eventRepository;
    private final EventSideRepository eventSideRepository;
    private final EventParticipantRepository eventParticipantRepository;
    private final EventMatchRepository eventMatchRepository;
    private final EventMatchMapVoteRepository eventMatchMapVoteRepository;

    private static final int MAX_PLAYERS_PER_SIDE = 50;

    @Transactional
    public UUID createEvent(CreateEventRequest request) {
        log.info("Creating new event: {} by user {}", request.name(), request.creatorUserId());

        Event event = Event.builder()
                .name(request.name())
                .creatorUserId(request.creatorUserId())
                .targetGameCount(request.targetGameCount())
                .timeStart(request.timeStart())
                .status(EventStatus.DRAFT)
                .build();

        eventRepository.save(event);

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

        return event.getId();
    }

    @Transactional
    public UUID joinEvent(JoinEventRequest request) {
        Event event = eventRepository.findByIdForUpdate(request.eventId())
                .orElseThrow(() -> new IllegalArgumentException("Event " + request.eventId() + " is not found"));

        if(event.getStatus() != EventStatus.REGISTRATION) {
            throw new IllegalStateException("Event registration is closed. Current status: " + event.getStatus().name());
        }

        EventSide side = eventSideRepository.findById(request.sideId())
                .orElseThrow(() -> new IllegalArgumentException("Side " + request.sideId() + " is not found"));

        if(!side.getEventId().equals(event.getId())) {
            throw new IllegalArgumentException("Side " + side.getId() + " does not belong to event " + event.getId());
        }

        if(eventParticipantRepository.existsByEventIdAndUserId(event.getId(), request.userId())) {
            throw new IllegalStateException("User " + request.userId() + " is already registered to event " + event.getId());
        }

        int currentSidePlayers = eventParticipantRepository.countBySideId(side.getId());
        if(currentSidePlayers >= MAX_PLAYERS_PER_SIDE) {
            throw new IllegalStateException("Side " + side.getId() + " is full: current players count = " + currentSidePlayers);
        }

        EventParticipant participant = EventParticipant.builder()
                .eventId(request.eventId())
                .sideId(request.sideId())
                .userId(request.userId())
                .clanId(request.clanId())
                .build();

        eventParticipantRepository.save(participant);

        log.info("User {} is joined to event {} for side {}",
                request.userId(), request.eventId(), request.sideId());

        return participant.getId();
    }

    @Transactional
    public void setSideReady(UUID eventId, UUID userId) {
        log.info("User {} is setting side ready for event {}", userId, eventId);

        Event event = eventRepository.findByIdForUpdate(eventId)
                .orElseThrow(() -> new IllegalArgumentException("Event " + eventId + " is not found"));

        if(event.getStatus() != EventStatus.LIVE && event.getStatus() != EventStatus.PREPARING) {
            throw new IllegalStateException("Server is not ready yet or match already finished");
        }

        EventSide currentSide = eventSideRepository.findAllByEventIdAndLeaderUserId(eventId, userId)
                .orElseThrow(() -> new IllegalStateException("User " + userId + "is not the side leader"));

        if(currentSide.isReady()) {
            return;
        }

        currentSide.setReady(true);
        currentSide.setReadyAt(Instant.now());
        eventSideRepository.save(currentSide);

        List<EventSide> allSides = eventSideRepository.findAllByEventId(eventId);
        boolean allReady = allSides.stream().allMatch(EventSide::isReady);

        if(allReady) {
            log.info("All sides are ready. Starting the first match for event {}", eventId);

            event.setStatus(EventStatus.LIVE);
            eventRepository.save(event);

            EventMatch firstMatch = eventMatchRepository.findByEventIdAndSequenceNumber(eventId, 1)
                    .orElseThrow(() -> new IllegalStateException("Match 1 does not generated"));

            firstMatch.setStatus(EventMatchStatus.IN_PROGRESS);
            firstMatch.setStartedAt(Instant.now());
            eventMatchRepository.save(firstMatch);
        }
    }

    @Transactional
    public void mapVote(UUID matchId, UUID userId, EventMatchMap map) {
        Event event = eventRepository.findByMatchId(matchId)
                .orElseThrow(() -> new IllegalArgumentException("Event for match " + matchId + " is not found"));

        if(event.getStatus() != EventStatus.REGISTRATION) {
            throw new IllegalStateException("Event not in REGISTRATION status");
        }

        if(!eventParticipantRepository.existsByEventIdAndUserId(event.getId(), userId)) {
            throw new IllegalStateException("User " + userId + " is not joined event " + event.getId());
        }

       eventMatchMapVoteRepository.upsertVote(
               UUID.randomUUID(),
               matchId,
               userId,
               map.name()
       );

        log.info("User {} voted for map {} in match {}", userId, map.name(), matchId);
    }
}
