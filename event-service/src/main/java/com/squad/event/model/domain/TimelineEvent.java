package com.squad.event.model.domain;

import com.squad.event.model.enums.TimelineEventType;
import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.JdbcTypeCode;
import org.hibernate.type.SqlTypes;

import java.time.Instant;
import java.util.Map;
import java.util.UUID;

@Entity
@Table(name = "timeline_event")
@Getter
@Setter
@AllArgsConstructor
@NoArgsConstructor
@Builder
public class TimelineEvent {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @Column(name = "match_id", nullable = false)
    private UUID matchId;

    @Enumerated(EnumType.STRING)
    @Column(name = "timeline_event_type", nullable = false)
    private TimelineEventType timelineEventType;

    @Column(name = "game_time_seconds", nullable = false)
    private int gameTimeSeconds;

    @JdbcTypeCode(SqlTypes.JSON)
    @Column(name = "payload", columnDefinition = "jsonb", nullable = false)
    private Map<String, Object> payload;

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    private Instant createdAt;
}
