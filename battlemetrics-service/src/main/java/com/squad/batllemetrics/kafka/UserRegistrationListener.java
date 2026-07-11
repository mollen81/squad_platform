package com.squad.batllemetrics.kafka;

import com.squad.batllemetrics.dto.UserRegisteredEvent;
import com.squad.batllemetrics.service.StatsFetchService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

@Service
@Slf4j
@RequiredArgsConstructor
public class UserRegistrationListener {
    private static final String TOPIC_USER_REGISTERED = "user.registered";
    private final StatsFetchService statsFetchService;

    public void onUserRegistered(UserRegisteredEvent registeredEvent) {
        log.info("A new registration event has been intercepted. SteamId: {}. Starting to fetch statistics...",
                registeredEvent.getSteamId());
        statsFetchService.fetchAndSendStatus(registeredEvent.getId(), registeredEvent.getSteamId());
    }
}
