package com.squad.event.grpc;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.Timestamp;
import io.grpc.stub.ServerCallStreamObserver;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;

@Service
@Slf4j
@RequiredArgsConstructor
public class GrpcTimelineStreamManager {
    private final Map<UUID, List<StreamObserver<TimelineEvent>>> observersByEvent = new ConcurrentHashMap<>();
    private final ObjectMapper objectMapper;


    // TODO: streams management logic
    public void registerGatewayObserver(UUID eventId, StreamObserver<TimelineEvent> streamObserver) {
        List<StreamObserver<TimelineEvent>> eventObservers = observersByEvent.computeIfAbsent(
                eventId, k -> new CopyOnWriteArrayList<>());
        eventObservers.add(streamObserver);

        if(streamObserver instanceof ServerCallStreamObserver<TimelineEvent> serverObserver) {
            serverObserver.setOnCancelHandler(() -> {
                log.info("gRPC stream canceled by gateway for event {}", eventId);
                removeObserver(eventId, streamObserver);
            });
        }

        log.info("Registered gRPC timeline stream for event {}. Active streams: {}", eventId, eventObservers.size());
    }

    public void broadcastTimelineEvent(UUID eventId, com.squad.event.model.domain.TimelineEvent timelineEvent) {
        List<StreamObserver<TimelineEvent>> eventObservers = observersByEvent.get(eventId);
        if(eventObservers == null || eventObservers.isEmpty()) {
            return;
        }

        TimelineEvent grpcMessage = mapToProto(timelineEvent);

        for(StreamObserver<TimelineEvent> observer : eventObservers) {
            try {
                synchronized (observer) {
                    observer.onNext(grpcMessage);
                }
            }
            catch (Exception e) {
                log.warn("Failed to send timeline event to gRPC stream for event {}. Removing observer.", eventId);
                removeObserver(eventId, observer);
            }
        }
    }

    private void removeObserver(UUID eventId, StreamObserver<TimelineEvent> observer) {
        List<StreamObserver<TimelineEvent>> observers = observersByEvent.get(eventId);
        if(observers != null) {
            observers.remove(observer);
            if(observers.isEmpty()) {
                observersByEvent.remove(eventId);
            }
        }
    }

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
