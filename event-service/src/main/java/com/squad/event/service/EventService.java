package com.squad.event.service;

import com.squad.event.kafka.event.EventMatchTimelineEvent;
import com.squad.event.kafka.event.MatchFinishedEvent;
import com.squad.event.kafka.event.ServerReadyEvent;
import com.squad.event.model.domain.*;
import com.squad.event.model.dto.CreateEventRequest;
import com.squad.event.model.dto.JoinEventRequest;
import com.squad.event.repo.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.UUID;

@Service
@Slf4j
@RequiredArgsConstructor
public class EventService {

    private final EventLifecycleManagementService eventLifecycleManagementService;
    private final EventParticipantManagementService eventParticipantManagementService;
    private final EventSideManagementService eventSideManagementService;
    private final EventMatchLayerManagementService eventMatchLayerManagementService;
    private final EventMatchTimelineService eventMatchTimelineService;


    public UUID createEvent(CreateEventRequest request) {
        return eventLifecycleManagementService.createEvent(request);
    }


    public UUID joinEvent(JoinEventRequest request) {
        return eventParticipantManagementService.joinEvent(request);
    }


    public void setSideReady(UUID eventId, UUID userId) {
        eventSideManagementService.setSideReady(eventId, userId);
    }


    public void mapVote(UUID matchId, UUID userId, String layer) {
        eventMatchLayerManagementService.mapVote(matchId, userId, layer);
    }


    public void startEventPreparation(UUID eventId) {
        eventLifecycleManagementService.startEventPreparation(eventId);
    }


    public void handleMatchFinished(MatchFinishedEvent matchFinishedEvent) {
        eventLifecycleManagementService.handleMatchFinished(matchFinishedEvent);
    }


    public void handleServerReady(ServerReadyEvent event) {
        eventLifecycleManagementService.handleServerReady(event);
    }

    public void processMatchTimelineEvent(EventMatchTimelineEvent eventMatchTimelineEvent) {
        eventMatchTimelineService.processTimelineEvent(eventMatchTimelineEvent);
    }
}
