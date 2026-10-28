package com.squad.gateway.controller;

import com.squad.gateway.model.dto.request.CreateEventHttpRequest;
import com.squad.gateway.model.dto.request.JoinEventHttpRequest;
import com.squad.gateway.service.event.EventGrpcClientService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;
import java.util.UUID;

@RestController
@RequestMapping("api/v1/event")
@RequiredArgsConstructor
public class EventController {
    private final EventGrpcClientService eventGrpcClientService;

    @PostMapping("/createEvent")
    public ResponseEntity<Map<String, UUID>> createClan(@Valid @RequestBody CreateEventHttpRequest request,
                                                        @AuthenticationPrincipal UUID creatorUserID) {
        UUID eventId = eventGrpcClientService.createEvent(request, creatorUserID);
        return ResponseEntity.status(HttpStatus.CREATED).body(Map.of("eventId", eventId));
    }

    @PostMapping("/joinEvent")
    public ResponseEntity<Map<String, UUID>> joinEvent(@Valid @RequestBody JoinEventHttpRequest request,
                                                       @AuthenticationPrincipal UUID userId) {
        UUID participantId = eventGrpcClientService.joinEvent(request, userId);
        return ResponseEntity.status(HttpStatus.CREATED).body(Map.of("participantId", participantId));
    }
}
