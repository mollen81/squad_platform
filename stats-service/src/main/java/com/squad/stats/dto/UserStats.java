package com.squad.stats.dto;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import lombok.*;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.LocalDateTime;
import java.util.UUID;

@Entity
@Table(name = "user_stats")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class UserStats {
    @Id
    @Column(name = "user_id")
    private UUID userId;

    @Column(name = "steam_id", unique = true, nullable = false, length = 20)
    private String steamId;

    @Column(name = "elo_rating")
    private Integer eloRating;

    @Column(name = "matches_played")
    private Integer matchesPlayed;

    @Column(name = "total_playtime_hours")
    private Integer totalPlaytimeHours;

    @Column(name = "kills")
    private Integer kills;

    @Column(name = "vehicles_destroyed")
    private int vehiclesDestroyed;

    @Column(name = "deaths")
    private Integer deaths;

    @Column(name = "revives")
    private Integer revives;

    @Column(name = "favourite_role", length = 50)
    private String favouriteRole;

    @UpdateTimestamp
    @Column(name = "last_updated_at")
    private LocalDateTime lastUpdatedAt;
}
