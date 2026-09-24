package com.squad.deployment.dto;

import lombok.Builder;

@Builder
public record ServerDeploymentFinishedEvent (
    String serverName,
    String password
) {}
