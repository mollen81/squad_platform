package com.squad.battlemetrics.kafka;

import com.squad.battlemetrics.dto.UserRegisteredEvent;
import com.squad.battlemetrics.service.StatsFetchService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Service;

@Service
@Slf4j
@RequiredArgsConstructor
public class UserRegistrationListener {
    private static final String TOPIC_USER_REGISTERED = "user.registered";
    private final StatsFetchService statsFetchService;

    @KafkaListener(topics = TOPIC_USER_REGISTERED, groupId = "battlemetrics-workers")
    public void onUserRegistered(UserRegisteredEvent registeredEvent) {
        log.info("A new registration event has been intercepted. SteamId: {}. Starting to fetch statistics...",
                registeredEvent.getSteamId());
        statsFetchService.fetchAndSendStatus(registeredEvent.getUserId(), registeredEvent.getSteamId());
    }
}
