package com.squad.stats.kafka;

import com.squad.stats.dto.EventEndedEvent;
import com.squad.stats.service.EventGrpcClient;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Service;

@Service
@Slf4j
@RequiredArgsConstructor
public class EventEndedListener {

    EventGrpcClient eventGrpcClient;

    @KafkaListener(topics = "event.ended")
    public void onEventEnded(EventEndedEvent event) {

    }
}
