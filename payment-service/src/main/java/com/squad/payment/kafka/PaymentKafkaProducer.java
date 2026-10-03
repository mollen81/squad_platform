package com.squad.payment.kafka;

import com.squad.payment.model.dto.ServerRentEvent;
import com.squad.payment.model.dto.VdsPurchasedEvent;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

import java.util.List;

@Component
@Slf4j
@RequiredArgsConstructor
public class PaymentKafkaProducer {

    private final KafkaTemplate<String, Object> kafkaTemplate;

    @Value("${spring.kafka.topics:vds-service}")
    private String vdsTopic;

    public void publishVdsPurchased(
            String eventId,
            String serverIp,
            String serverPassword,
            ServerRentEvent.ServerConfig serverConfig,
            List<ServerRentEvent.AdminEntry> admins,
            List<String> whitelistSteamIds) {

        VdsPurchasedEvent event = new VdsPurchasedEvent(
          "vps.purchased",
                eventId,
                serverIp,
                serverPassword,
                serverConfig,
                admins,
                whitelistSteamIds
        );

        log.info("vps.purchased publication for event: {}, on IP: {}", eventId, serverIp);

        kafkaTemplate.send(vdsTopic, eventId, event);
    }
}
