package com.squad.gateway.controller;

import com.squad.gateway.model.dto.request.CreateEventHttpRequest;
import com.squad.gateway.model.dto.request.JoinEventHttpRequest;
import com.squad.gateway.model.dto.request.MapVoteHttpRequest;
import com.squad.gateway.model.dto.request.UpdateSideReadyHttpRequest;
import com.squad.gateway.security.UserPrincipal;
import com.squad.gateway.service.event.EventGrpcClientService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.*;

import java.util.Map;
import java.util.UUID;

@RestController
@RequestMapping("api/v1/events")
@RequiredArgsConstructor
public class EventController {
    private final EventGrpcClientService eventGrpcClientService;

    // TODO getUpcomingEvents @GetMapping getEvent @GetMapping("/{eventId}")

    @PostMapping
    public ResponseEntity<Map<String, UUID>> createEvent(@Valid @RequestBody CreateEventHttpRequest request,
                                                        @AuthenticationPrincipal UserPrincipal userPrincipal) {
        UUID eventId = eventGrpcClientService.createEvent(request, userPrincipal.id());
        return ResponseEntity.status(HttpStatus.CREATED).body(Map.of("eventId", eventId));
    }

    @PostMapping("/{eventId}/join")
    public ResponseEntity<Map<String, UUID>> joinEvent(
            @PathVariable UUID eventId,
            @Valid @RequestBody JoinEventHttpRequest request,
            @AuthenticationPrincipal UserPrincipal userPrincipal
    ) {
        UUID participantId = eventGrpcClientService.joinEvent(request, userPrincipal.id());
        return ResponseEntity.status(HttpStatus.CREATED).body(Map.of("participantId", participantId));
    }

    @PatchMapping("/{eventId}/sides/{sideId}/ready")
    public ResponseEntity<Void> updateSideReady(
            @PathVariable UUID eventId,
            @PathVariable UUID sideId,
            @Valid @RequestBody UpdateSideReadyHttpRequest request,
            @AuthenticationPrincipal UserPrincipal userPrincipal
    ) {
        eventGrpcClientService.setSideReady(
                eventId, sideId, userPrincipal.id(), request.isReady());
        return ResponseEntity.ok().build();
    }

    @PutMapping("/{eventId}/matches/{matchId}/map-vote")
    public ResponseEntity<Void> mapVote(
            @PathVariable UUID eventId,
            @PathVariable UUID matchId,
            @Valid @RequestBody MapVoteHttpRequest request,
            @AuthenticationPrincipal UserPrincipal userPrincipal
    ) {
        eventGrpcClientService.mapVote(eventId, matchId, userPrincipal.id(), request.layer());
        return ResponseEntity.ok().build();
    }



}
