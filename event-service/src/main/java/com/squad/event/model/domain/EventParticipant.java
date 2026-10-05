package com.squad.event.model.domain;

import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;

import java.time.Instant;
import java.util.UUID;

@Entity
@Table(name = "event_participant",
        uniqueConstraints = {
                @UniqueConstraint(name = "uq_event_participant_event_user", columnNames = {"event_id", "user_id"})
        })
@AllArgsConstructor
@NoArgsConstructor
@Getter
@Setter
@Builder
public class EventParticipant {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @Column(name = "event_id", nullable = false)
    private UUID eventId;

    @Column(name = "side_id", nullable = false)
    private UUID sideId;

    @Column(name = "user_id", nullable = false)
    private UUID userId;

    @Column(name = "clan_id")
    private UUID clanId;

    @CreationTimestamp
    @Column(name = "joined_at", nullable = false)
    private Instant joinedAt;
}
