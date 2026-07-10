package com.squad.user.kafka;

import com.squad.user.event.UserRegisteredEvent;
import lombok.Builder;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

@Slf4j
@Builder
@Service
@RequiredArgsConstructor
public class UserEventProducer {
    private final KafkaTemplate<String, Object> kafkaTemplate;

    private static final String TOPIC_USER_REGISTERED = "user.registered";

    public void sendUserRegisteredEvent(UserRegisteredEvent event) {
        log.info("Sending an event to Kafka [{}] for the user: {}", TOPIC_USER_REGISTERED, event.getUserId());
        kafkaTemplate.send(TOPIC_USER_REGISTERED, event.getUserId().toString(), event);
    }
}
