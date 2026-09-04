package com.squad.gateway.config;

import com.squad.gateway.websocket.NotificationHandler;
import lombok.RequiredArgsConstructor;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.socket.config.annotation.EnableWebSocket;

@Configuration
@EnableWebSocket
@RequiredArgsConstructor
public class WebSocketConfig {

    private final NotificationHandler notificationHandler;
}
