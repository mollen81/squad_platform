package com.squad.event.model.domain;

import com.squad.event.model.enums.EventMatchMap;
import com.squad.event.model.enums.EventMatchStatus;
import jakarta.persistence.*;
import lombok.*;

import java.time.Instant;
import java.util.UUID;

@Entity
@Table(
        name = "event_match",
        uniqueConstraints = {
                @UniqueConstraint(name = "uq_event_match_event_sequence", columnNames = {"event_id", "sequence_number"})
        }
)
@AllArgsConstructor
@NoArgsConstructor
@Builder
@Getter
@Setter
public class EventMatch {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @Column(name = "event_id", nullable = false)
    private UUID eventId;

    @Column(name = "sequence_number", nullable = false)
    private int sequenceNumber; // 1, 2, 3 ...

    @Enumerated(EnumType.STRING)
    private EventMatchMap map;

    @Enumerated(EnumType.STRING)
    @Column(length = 32, nullable = false)
    private EventMatchStatus status;

    @Column(name = "winner_side_id")
    private UUID winnerSideId;

    @Column(name = "started_at")
    private Instant startedAt;

    @Column(name = "finished_at")
    private Instant finishedAt;
}
