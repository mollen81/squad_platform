package com.squad.stats.dto;

import lombok.*;

import java.util.UUID;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Getter
@Setter
@Builder
public class PlayerStatsFetchedEvent {
    private UUID id;
    private String steamId;
    private int totalPlaytimeHours;
    private int kills;
    private int vehiclesDestroyed;
    private int deaths;
    private int revives;
    private String favouriteRole;
    private long timestamp;
}
