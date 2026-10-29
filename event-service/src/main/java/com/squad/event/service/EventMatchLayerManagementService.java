package com.squad.event.service;

import com.squad.event.model.domain.Event;
import com.squad.event.model.dto.MapVoteHttpRequest;
import com.squad.event.model.enums.EventStatus;
import com.squad.event.repo.EventMatchLayerVoteRepository;
import com.squad.event.repo.EventParticipantRepository;
import com.squad.event.repo.EventRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class EventMatchLayerManagementService {

    private final EventRepository eventRepository;
    private final EventParticipantRepository eventParticipantRepository;
    private final EventMatchLayerVoteRepository eventMatchLayerVoteRepository;

    @Transactional
    public void mapVote(MapVoteHttpRequest request) {
        Event event = eventRepository.findById(request.eventId())
                .orElseThrow(() -> new IllegalArgumentException("Event with id " + request.eventId() + " is not found"));

        if(event.getStatus() != EventStatus.REGISTRATION) {
            throw new IllegalStateException("Event not in REGISTRATION status");
        }

        if(!eventParticipantRepository.existsByEventIdAndUserId(event.getId(), request.userId())) {
            throw new IllegalStateException("User " + request.userId() + " is not joined event " + event.getId());
        }

        eventMatchLayerVoteRepository.upsertVote(
                UUID.randomUUID(),
                request.matchId(),
                request.userId(),
                request.layer());

        log.info("User {} voted for layer {} in match {}", request.userId(), request.layer(), request.matchId());
    }
}
