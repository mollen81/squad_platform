package com.squad.deployment.kafka;

import com.squad.deployment.dto.ServerDeploymentFinishedEvent;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

@Component
@Slf4j
@RequiredArgsConstructor
public class ServerDeploymentFinishedProducer {
    private final KafkaTemplate<String, Object> kafkaTemplate;

    private static final String TOPIC_DEPLOYMENT_FINISHED = "server.deployment.finished";

    public void sendServerDeploymentFinishedEvent(ServerDeploymentFinishedEvent event) {
        kafkaTemplate.send(TOPIC_DEPLOYMENT_FINISHED, event);
    }
}
