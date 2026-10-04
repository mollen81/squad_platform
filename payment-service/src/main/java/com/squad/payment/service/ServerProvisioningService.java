package com.squad.payment.service;

import com.squad.payment.client.AezaApiClient;
import com.squad.payment.client.dto.AezaOrderResponse;
import com.squad.payment.client.dto.AezaServiceDetailResponse;
import com.squad.payment.kafka.PaymentKafkaProducer;
import com.squad.payment.model.domain.Payment;
import com.squad.payment.model.domain.VdsHosting;
import com.squad.payment.model.dto.ServerRentEvent;
import com.squad.payment.model.enums.Hosting;
import com.squad.payment.model.enums.PaymentStatus;
import com.squad.payment.repository.PaymentRepository;
import com.squad.payment.repository.HostingRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
@Slf4j
@RequiredArgsConstructor
public class ServerProvisioningService {
    private final PaymentRepository paymentRepository;
    private final HostingRepository hostingRepository;
    private final AezaApiClient aezaApiClient;
    private final PaymentKafkaProducer paymentKafkaProducer;

    @Async
    @Transactional
    public void processServerRent(ServerRentEvent event) {
        log.info("Starting the rental process for the event {}", event.event_id());

        VdsHosting aezaHosting = hostingRepository.findByHostingName(Hosting.AEZA.toString())
                .orElseThrow(() -> new IllegalStateException("AEZA provider is not configured"));

        Payment payment = Payment.builder()
                .eventId(event.event_id())
                .vdsHosting(aezaHosting)
                .paymentStatus(PaymentStatus.PROCESSING)
                .build();
        payment = paymentRepository.save(payment);

        try {
            AezaOrderResponse order = aezaApiClient.orderVds(aezaHosting.getApiToken(), 152, 6);

            payment.setExternalServiceId(order.data().serviceId());
            paymentRepository.save(payment);

            AezaServiceDetailResponse serviceDetails = waitForServerResponse(aezaHosting.getApiToken(), order.data().serviceId());

            payment.setAmount(serviceDetails.data().price());
            payment.setPaymentStatus(PaymentStatus.SUCCESS);
            paymentRepository.save(payment);

            paymentKafkaProducer.publishVdsPurchased(
                    event.event_id(),
                    event.serverConfig().serverName(),
                    event.serverConfig().password(),
                    event.serverConfig(),
                    event.admins(),
                    event.whitelistSteamIds()
            );
        }
        catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    private AezaServiceDetailResponse waitForServerResponse(String token, Long serviceId) throws InterruptedException {
        int attempts = 0;
        while(attempts < 60) {
            AezaServiceDetailResponse serviceDetails = aezaApiClient.getServiceDetail(token, serviceId);

            if("active".equalsIgnoreCase(serviceDetails.data().status()) && serviceDetails.data().ipAddress() != null) {
                return serviceDetails;
            }
            Thread.sleep(5000);
            attempts++;
        }

        throw new RuntimeException("IP receiving timeout from AEZA");
    }


    public void processServerTeardown(String eventId) {
        Payment payment = paymentRepository.findByEventIdAndPaymentStatus(eventId, PaymentStatus.SUCCESS)
                .orElseThrow(() -> new IllegalArgumentException("Active payment is not found"));

        String token = payment.getVdsHosting().getApiToken();

        try {
            aezaApiClient.deleteVds(token, payment.getExternalServiceId());

            payment.setPaymentStatus(PaymentStatus.CLOSED);
            paymentRepository.save(payment);

            log.info("Server for event {} is successfully deleted", eventId);
        }
        catch (Exception e) {
            log.error("Error while deleting AEZA server (Balance leak!): {}", e.getMessage());
        }
    }
}
