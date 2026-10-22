package com.squad.event.grpc;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.Timestamp;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;

@Service
@Slf4j
@RequiredArgsConstructor
public class GrpcTimelineStreamManager {
    private final Map<UUID, List<StreamObserver<TimelineEvent>>> observersByEvent = new ConcurrentHashMap<>();
    private final ObjectMapper objectMapper;


    // TODO: streams management logic

    public TimelineEvent mapToProto(com.squad.event.model.domain.TimelineEvent event) {
        String payloadJson = "{}";

        try {
            if(event.getPayload() != null) {
                payloadJson = objectMapper.writeValueAsString(event.getPayload());
            }
        }
        catch (JsonProcessingException e) {
            log.error("Failed to serialize timeline payload for entry {}", event.getId(), e);
        }

        return TimelineEvent.newBuilder()
                .setId(event.getId().toString())
                .setMatchId(event.getMatchId().toString())
                .setTimelineEventType(event.getTimelineEventType().toString())
                .setGameTimeSeconds(event.getGameTimeSeconds())
                .setPayload(event.getPayload().toString())
                .setCreatedAt(event.getCreatedAt().toEpochMilli())
                .build();
    }
}
