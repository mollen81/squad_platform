package com.squad.deployment.kafka;

import com.squad.deployment.dto.VpsPurchasedEvent;
import com.squad.deployment.service.ServerDeploymentService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;

@Component
@Slf4j
@RequiredArgsConstructor
public class VpsPurchasedListener {

    private final ServerDeploymentService deploymentService;

    @KafkaListener(topics = "vps.purchased", groupId = "deploy-service-group")
    public void onServerPurchase(VpsPurchasedEvent event) {
        deploymentService.deploySquadServer(event);
    }
}
