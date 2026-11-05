package com.squad.gateway.service.event;

import com.squad.event.grpc.*;
import com.squad.gateway.model.dto.response.TimelineEventHttpResponse;
import io.grpc.Context;
import io.grpc.stub.StreamObserver;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.http.MediaType;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.io.IOException;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;

@Service
@Slf4j
public class EventStreamBridgeService {

    // Внедряем асинхронный gRPC-стаб (для стриминга нужен именно Stub, не BlockingStub!)
    @GrpcClient("event-service")
    private EventServiceGrpc.EventServiceStub eventServiceStub;

    private static final long SSE_TIMEOUT = 60 * 60 * 1000L; // 1 час

    // Хранилище браузерных подписчиков: eventId -> список SseEmitter
    private final Map<UUID, List<SseEmitter>> emittersByMatch = new ConcurrentHashMap<>();
    private final Map<UUID, Context.CancellableContext> grpcStreamContexts = new ConcurrentHashMap<>();

    public SseEmitter subscribe(UUID eventId, UUID matchId) {
        SseEmitter emitter = new SseEmitter(SSE_TIMEOUT);

        // Получаем список подписчиков конкретного матча
        List<SseEmitter> emitters = emittersByMatch.computeIfAbsent(
                matchId, k -> new CopyOnWriteArrayList<>()
        );
        emitters.add(emitter);

        Runnable cleanup = () -> removeEmitter(matchId, emitter);
        emitter.onCompletion(cleanup);
        emitter.onTimeout(cleanup);
        emitter.onError(e -> cleanup.run());

        try {
            emitter.send(SseEmitter.event()
                    .name("INIT")
                    .data("Subscribed to live stream for match " + matchId));
        } catch (IOException e) {
            removeEmitter(matchId, emitter);
            return emitter;
        }

        // Запускаем gRPC стрим (если он еще не запущен для этого матча)
        ensureGrpcStreamStarted(eventId, matchId);

        log.info("Client subscribed to match {}. Viewers count: {}", matchId, emitters.size());
        return emitter;
    }

    private synchronized void ensureGrpcStreamStarted(UUID eventId, UUID matchId) {
        if (grpcStreamContexts.containsKey(matchId)) {
            return;
        }

        log.info("Opening single gRPC stream to event-service for match {}", matchId);

        StreamMatchTimelineRequest request = StreamMatchTimelineRequest.newBuilder()
                .setEventId(eventId.toString()) // Передаем для проверки иерархии
                .setMatchId(matchId.toString())
                .build();

        Context.CancellableContext cancellableContext = Context.current().withCancellation();
        grpcStreamContexts.put(matchId, cancellableContext);

        cancellableContext.run(() -> eventServiceStub.streamMatchTimeline(request, new StreamObserver<TimelineEvent>() {
            @Override
            public void onNext(TimelineEvent message) {
                // Преобразуем Protobuf в чистый Record для сериализации
                TimelineEventHttpResponse dto = new TimelineEventHttpResponse(
                        UUID.fromString(message.getId()),
                        UUID.fromString(message.getMatchId()),
                        message.getTimelineEventType(),
                        message.getGameTimeSeconds(),
                        message.getPayload(),
                        message.getCreatedAt()
                );

                broadcastToSse(matchId, dto);
            }

            @Override
            public void onError(Throwable t) {
                log.error("gRPC stream error from event-service for match {}: {}", matchId, t.getMessage());
                closeAllEmittersForMatch(matchId);
            }

            @Override
            public void onCompleted() {
                log.info("gRPC stream completed by event-service for match {}", matchId);
                closeAllEmittersForMatch(matchId);
            }
        }));
    }

    private void broadcastToSse(UUID matchId, TimelineEventHttpResponse eventData) {
        List<SseEmitter> emitters = emittersByMatch.get(matchId);
        if (emitters == null || emitters.isEmpty()) {
            return;
        }

        for (SseEmitter emitter : emitters) {
            try {
                emitter.send(SseEmitter.event()
                        .name("MATCH_TIMELINE_EVENT")
                        .data(eventData, MediaType.APPLICATION_JSON));
            } catch (Exception e) {
                removeEmitter(matchId, emitter);
            }
        }
    }

    private synchronized void removeEmitter(UUID matchId, SseEmitter emitter) {
        List<SseEmitter> emitters = emittersByMatch.get(matchId);
        if (emitters != null) {
            emitters.remove(emitter);

            if (emitters.isEmpty()) {
                emittersByMatch.remove(matchId);
                Context.CancellableContext context = grpcStreamContexts.remove(matchId);
                if (context != null) {
                    log.info("No more viewers for match {}. Cancelling gRPC stream.", matchId);
                    context.cancel(null);
                }
            }
        }
    }

    private synchronized void closeAllEmittersForMatch(UUID matchId) {
        grpcStreamContexts.remove(matchId);
        List<SseEmitter> emitters = emittersByMatch.remove(matchId);
        if (emitters != null) {
            emitters.forEach(SseEmitter::complete);
        }
    }

    // Фоновый Heartbeat каждые 15 секунд
    @Scheduled(fixedRate = 15000)
    public void sendHeartbeat() {
        if (emittersByMatch.isEmpty()) {
            return;
        }

        emittersByMatch.forEach((matchId, emitters) -> {
            for (SseEmitter emitter : emitters) {
                try {
                    emitter.send(SseEmitter.event().comment("ping"));
                } catch (Exception e) {
                    log.debug("Dead SSE connection detected during heartbeat for match {}. Removing.", matchId);
                    removeEmitter(matchId, emitter);
                }
            }
        });
    }
}