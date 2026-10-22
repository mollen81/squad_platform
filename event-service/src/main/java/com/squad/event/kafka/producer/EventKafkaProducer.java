package com.squad.event.kafka.producer;

import com.squad.event.kafka.event.EventEndedEvent;
import com.squad.event.kafka.event.ServerRentEvent;
import com.squad.event.kafka.event.ServerStopEvent;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class EventKafkaProducer {
    private final KafkaTemplate<String, Object> kafkaTemplate;

    private static final String TOPIC_SERVER_RENT = "server.rent";
    private static final String TOPIC_SERVER_STOP = "server.stop";
    private static final String TOPIC_EVENT_ENDED = "event.ended";


    public void sendServerRentEvent(ServerRentEvent event) {
        log.info("Sending an event to Kafka [{}] event: {}", TOPIC_SERVER_RENT, event.eventId());
        kafkaTemplate.send(TOPIC_SERVER_RENT, event.eventId().toString(), event);
    }

    public void sendServerStopEvent(ServerStopEvent serverStopEvent) {
        setKafkaMessageLog(TOPIC_SERVER_STOP, serverStopEvent.eventId());
        kafkaTemplate.send(TOPIC_SERVER_STOP, serverStopEvent.eventId().toString(), serverStopEvent);
    }

    public void sendEventEndedEvent(EventEndedEvent eventEndedEvent) {
        setKafkaMessageLog(TOPIC_EVENT_ENDED, eventEndedEvent.eventId());
        kafkaTemplate.send(TOPIC_EVENT_ENDED, eventEndedEvent.eventId().toString(), eventEndedEvent);
    }


    private static void setKafkaMessageLog(String topic, UUID eventId) {
        log.info("Sending an event to Kafka [{}] for event: {}", topic, eventId);
    }
}