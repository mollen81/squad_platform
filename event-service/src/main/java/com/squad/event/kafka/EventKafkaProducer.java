package com.squad.event.kafka;

import com.squad.event.kafka.event.ServerRentEvent;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

@Slf4j
@Service
@RequiredArgsConstructor
public class EventKafkaProducer {
    private final KafkaTemplate<String, Object> kafkaTemplate;

    private static final String TOPIC_SERVER_RENT = "server.rent";

    public void sendServerRentEvent(ServerRentEvent event) {
        log.info("Sending an event to Kafka [{}] for the Event: {}", TOPIC_SERVER_RENT, event.eventId());
        kafkaTemplate.send(TOPIC_SERVER_RENT, event.eventId().toString(), event);
    }

}