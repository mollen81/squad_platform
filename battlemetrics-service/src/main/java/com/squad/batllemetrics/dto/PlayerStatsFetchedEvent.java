package com.squad.batllemetrics.dto;

import lombok.*;

import java.util.UUID;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor
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
