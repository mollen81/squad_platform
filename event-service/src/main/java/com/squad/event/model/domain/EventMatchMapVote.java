package com.squad.event.model.domain;

import com.squad.event.model.enums.EventMatchMap;
import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.Instant;
import java.util.UUID;

@Entity
@Table(
        name = "event_match_map_vote",
        uniqueConstraints = {
                @UniqueConstraint(
                        name = "uq_event_match_map_vote_match_id_event_id",
                        columnNames = {"match_id", "user_id"}
                )
        }
)
@AllArgsConstructor
@NoArgsConstructor
@Getter
@Setter
@Builder
public class EventMatchMapVote {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @Column(name = "match_id")
    private UUID matchId;

    @Column(name = "user_id")
    private UUID userId;

    @Enumerated(EnumType.STRING)
    private EventMatchMap map;

    @CreationTimestamp
    @Column(name = "created_at", updatable = false)
    private Instant createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at")
    private Instant updatedAt;
}
