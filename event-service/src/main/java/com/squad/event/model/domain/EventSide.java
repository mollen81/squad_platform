package com.squad.event.model.domain;

import jakarta.persistence.*;
import lombok.*;

import java.time.Instant;
import java.util.UUID;

@Entity
@Table(name = "event_side")
@AllArgsConstructor
@NoArgsConstructor
@Builder
@Getter
@Setter
public class EventSide {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @Column(name = "event_id", nullable = false)
    private UUID eventId;

    private String name; // for example: (Team A / Clan name / ...)

    @Column(name = "leader_user_id", nullable = false)
    private UUID leaderUserId;

    @Column(name = "is_ready", nullable = false)
    private boolean isReady = false;

    @Column(name = "ready_at")
    private Instant readyAt;
}