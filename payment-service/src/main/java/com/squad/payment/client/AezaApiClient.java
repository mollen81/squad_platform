package com.squad.payment.client;

import com.squad.payment.client.dto.AezaOrderRequest;
import com.squad.payment.client.dto.AezaOrderResponse;
import com.squad.payment.client.dto.AezaServiceDetailResponse;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;

import java.util.Map;

@Slf4j
@Component
public class AezaApiClient {

    private final RestClient restClient;

    public AezaApiClient(@Value("${aeza.api.url:https://my.aeza.net/api/v2/}") String baseUrl) {
        this.restClient = RestClient.builder()
                .baseUrl(baseUrl)
                .defaultHeader("Content-type", MediaType.APPLICATION_JSON_VALUE)
                .build();
    }

    public AezaOrderResponse orderVds(String apiToken, int tariffId, int osId) {
        AezaOrderRequest request = new AezaOrderRequest(
                tariffId,
                "hour",
                Map.of("os", osId)
        );

        return restClient.post()
                .uri("/services/orders")
                .header("Authorization", "Bearer " + apiToken)
                .body(request)
                .retrieve()
                .body(AezaOrderResponse.class);
    }

    public AezaServiceDetailResponse getServiceDetail(String apiToken, Long serviceId) {
        return restClient.get()
                .uri("/services/{id}", serviceId)
                .header("Authorization", "Bearer " + apiToken)
                .retrieve()
                .body(AezaServiceDetailResponse.class);
    }

    public void deleteVds(String apiToken, Long serviceId) {
        restClient.delete()
                .uri("/services/{id}", serviceId)
                .header("Authorization", "Bearer " + apiToken)
                .retrieve()
                .toBodilessEntity();
    }
}
