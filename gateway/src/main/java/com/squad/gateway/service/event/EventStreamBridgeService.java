package com.squad.gateway.service.event;

import com.squad.event.grpc.*;
import io.grpc.Context;
import io.grpc.stub.StreamObserver;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.http.MediaType;
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
    private final Map<UUID, List<SseEmitter>> emittersByEvent = new ConcurrentHashMap<>();

    // Хранилище контекстов gRPC-стримов для их отмены: eventId -> CancellableContext
    private final Map<UUID, Context.CancellableContext> grpcStreamContexts = new ConcurrentHashMap<>();

    public SseEmitter subscribe(UUID eventId) {
        SseEmitter emitter = new SseEmitter(SSE_TIMEOUT);

        List<SseEmitter> emitters = emittersByEvent.computeIfAbsent(
                eventId, k -> new CopyOnWriteArrayList<>()
        );
        emitters.add(emitter);

        Runnable cleanup = () -> removeEmitter(eventId, emitter);
        emitter.onCompletion(cleanup);
        emitter.onTimeout(cleanup);
        emitter.onError(e -> cleanup.run());

        // Шлем приветственное системное событие в браузер
        try {
            emitter.send(SseEmitter.event()
                    .name("INIT")
                    .data("Subscribed to live stream for event " + eventId));
        } catch (IOException e) {
            removeEmitter(eventId, emitter);
            return emitter;
        }

        // Если это первый зритель — открываем единственный gRPC-стрим к event-service
        ensureGrpcStreamStarted(eventId);

        log.info("Client subscribed to event {}. Viewers count: {}", eventId, emitters.size());
        return emitter;
    }

    private synchronized void ensureGrpcStreamStarted(UUID eventId) {
        if (grpcStreamContexts.containsKey(eventId)) {
            return; // Стрим уже запущен для предыдущих зрителей
        }

        log.info("Opening single gRPC stream to event-service for event {}", eventId);

        StreamMatchTimelineRequest request = StreamMatchTimelineRequest.newBuilder()
                .setEventId(eventId.toString())
                .build();

        // Оборачиваем вызов в CancellableContext, чтобы прервать стрим, когда зрителей не останется
        Context.CancellableContext cancellableContext = Context.current().withCancellation();
        grpcStreamContexts.put(eventId, cancellableContext);

        cancellableContext.run(() -> eventServiceStub.streamMatchTimeline(request, new StreamObserver<>() {
            @Override
            public void onNext(TimelineEvent message) {
                TimelineEvent dto = TimelineEvent.newBuilder()
                        .setId(message.getId())
                        .setMatchId(message.getMatchId())
                        .setTimelineEventType(message.getTimelineEventType())
                        .setGameTimeSeconds(message.getGameTimeSeconds())
                        .setPayload(message.getPayload())
                        .setCreatedAt(message.getCreatedAt())
                        .build();

                // Fan-Out: рассылаем полученное событие всем браузерам
                broadcastToSse(eventId, dto);
            }

            @Override
            public void onError(Throwable t) {
                log.error("gRPC stream error from event-service for event {}: {}", eventId, t.getMessage());
                closeAllEmittersForEvent(eventId);
            }

            @Override
            public void onCompleted() {
                log.info("gRPC stream completed by event-service for event {}", eventId);
                closeAllEmittersForEvent(eventId);
            }
        }));
    }

    private void broadcastToSse(UUID eventId, TimelineEvent eventData) {
        List<SseEmitter> emitters = emittersByEvent.get(eventId);
        if (emitters == null || emitters.isEmpty()) {
            return;
        }

        for (SseEmitter emitter : emitters) {
            try {
                emitter.send(SseEmitter.event()
                        .name("MATCH_TIMELINE_EVENT")
                        .data(eventData, MediaType.APPLICATION_JSON));
            } catch (Exception e) {
                removeEmitter(eventId, emitter);
            }
        }
    }

    private synchronized void removeEmitter(UUID eventId, SseEmitter emitter) {
        List<SseEmitter> emitters = emittersByEvent.get(eventId);
        if (emitters != null) {
            emitters.remove(emitter);

            // Если зрителей больше нет — прерываем gRPC стрим к event-service
            if (emitters.isEmpty()) {
                emittersByEvent.remove(eventId);
                Context.CancellableContext context = grpcStreamContexts.remove(eventId);
                if (context != null) {
                    log.info("No more viewers for event {}. Cancelling gRPC stream.", eventId);
                    context.cancel(null); // Посылает RST_STREAM в event-service
                }
            }
        }
    }

    private synchronized void closeAllEmittersForEvent(UUID eventId) {
        grpcStreamContexts.remove(eventId);
        List<SseEmitter> emitters = emittersByEvent.remove(eventId);
        if (emitters != null) {
            emitters.forEach(SseEmitter::complete);
        }
    }
}