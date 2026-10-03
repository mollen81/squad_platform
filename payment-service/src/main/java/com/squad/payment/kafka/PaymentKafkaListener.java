package com.squad.payment.kafka;

import com.squad.payment.model.dto.ServerRentEvent;
import com.squad.payment.service.ServerProvisioningService;
import lombok.RequiredArgsConstructor;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;

@Component
@RequiredArgsConstructor
public class PaymentKafkaListener {

    private final ServerProvisioningService provisioningService;

    @KafkaListener(topics = "server.rent", groupId = "payment-group")
    public void onServerRent(ServerRentEvent event) {
        provisioningService.processServerRent(event);
    }

    @KafkaListener(topics = "server.teardown", groupId = "payment-group")
    public void onServerTearDown() {
        provisioningService.
    }
}
