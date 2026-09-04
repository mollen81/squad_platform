package com.squad.gateway.websocket;

import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import org.springframework.web.socket.CloseStatus;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;
import org.springframework.web.socket.handler.TextWebSocketHandler;

import java.io.IOException;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Slf4j
@Component
public class NotificationHandler extends TextWebSocketHandler {
    // key = userId, value = websocket session
    private final Map<String, WebSocketSession> sessions = new ConcurrentHashMap<>();

    @Override
    public void afterConnectionEstablished(WebSocketSession session) {
        String query = session.getUri().getQuery();
        String userId = extractUserId(query);

        if(userId != null) {
            sessions.put(userId, session);
            log.info("Websocket connection for user {} is opened. Active sessions: {}", userId, sessions.size());
        }
    }

    @Override
    public void afterConnectionClosed(WebSocketSession session, CloseStatus status) {
        String userId = extractUserId(session.getUri().getQuery());

        if(userId != null) {
            sessions.remove(userId);
            log.info("Websocket connection for user {} is closed. Active sessions: {}", userId, sessions.size());
        }
    }

    public void sendNotification(String userId, String messageJson) {
        WebSocketSession session = sessions.get(userId);
        if(session != null && session.isOpen()) {
            try {
                session.sendMessage(new TextMessage(messageJson));
            } catch (IOException e) {
                log.error("Error while message delivery to user {}", userId, e);
            }
        }
    }


    private String extractUserId(String query) {
        if(query != null && query.contains("userId=")) {
            return query.split("userId=")[1].split("&")[0];
        }
        return null;
    }
}