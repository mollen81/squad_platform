package com.squad.event.service;

import com.squad.event.client.ClanServiceGrpcClient;
import com.squad.event.grpc.GetClanIdFromUserIdRequest;
import com.squad.event.model.domain.Event;
import com.squad.event.model.domain.EventParticipant;
import com.squad.event.model.domain.EventSide;
import com.squad.event.model.dto.JoinEventRequest;
import com.squad.event.model.enums.EventStatus;
import com.squad.event.repo.EventParticipantRepository;
import com.squad.event.repo.EventRepository;
import com.squad.event.repo.EventSideRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class EventParticipantManagementService {

    private final EventRepository eventRepository;
    private final EventSideRepository eventSideRepository;
    private final EventParticipantRepository eventParticipantRepository;

    private final ClanServiceGrpcClient clanServiceGrpcClient;

    private static final int MAX_PLAYERS_PER_SIDE = 50;

    @Transactional
    public UUID joinEvent(JoinEventRequest request) {
        GetClanIdFromUserIdRequest getClanIdFromUserIdRequest = GetClanIdFromUserIdRequest.newBuilder()
                .setUserId(request.userId().toString())
                .build();
        UUID clanId = clanServiceGrpcClient.resolveUserClanId(request.userId());

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
                .clanId(clanId)
                .build();

        eventParticipantRepository.save(participant);

        log.info("User {} is joined to event {} for side {}",
                request.userId(), request.eventId(), request.sideId());

        return participant.getId();
    }
}
