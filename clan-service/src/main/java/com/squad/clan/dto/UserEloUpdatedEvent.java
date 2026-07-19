package com.squad.clan.dto;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.UUID;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor
public class UserEloUpdatedEvent {
    private UUID userId;
    private int oldElo;
    private int newElo;
}
