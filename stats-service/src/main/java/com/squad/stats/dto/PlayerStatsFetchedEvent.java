package com.squad.stats.dto;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.UUID;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class PlayerStatsFetchedEvent {
    private UUID id;
    private String steamId;
    private int totalPlaytimeHours;
    private int kills;
    private int deaths;
    private int revives;
    private String favouriteRole;
    private long timestamp;
}
