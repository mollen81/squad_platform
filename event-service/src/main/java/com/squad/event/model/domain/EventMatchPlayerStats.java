package com.squad.event.model.domain;

import jakarta.persistence.*;
import lombok.*;

import java.util.UUID;

@Entity
@Table(
        name = "match_player_stats",
        uniqueConstraints = {
                @UniqueConstraint(
                        name = "uq_match_player_stats_match_participant",
                        columnNames = {"match_id", "participant_id"}
                )
        })
@AllArgsConstructor
@NoArgsConstructor
@Getter
@Setter
@Builder
public class EventMatchPlayerStats {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @Column(name = "event_match_id", nullable = false)
    private UUID eventMatchId;

    @Column(name = "participant_id")
    private UUID participantId;

    @Column(length = 64, name = "primary_role", nullable = false)
    private String primaryRole;

    @Column(nullable = false)
    private int kills = 0;

    @Column(nullable = false)
    private int deaths = 0;

    @Column(nullable = false)
    private int points = 0;

    @Column(nullable = false)
    private int revives = 0;

    @Column(name = "team_kills", nullable = false)
    private int teamKills = 0;

    @Column(name = "vehicles_destroyed", nullable = false)
    private int destroyedVehicles = 0;

    @Column(name = "vehicles_lost", nullable = false)
    private int vehiclesLost = 0;

    @Column(name = "flying_time_seconds", nullable = false)
    private int flyingTimeSeconds = 0;

    @Column(name = "supply_points", nullable = false)
    private int supplyPoints = 0;
}
