package com.squad.controller;

import com.squad.grpc.user.ResolveSteamAuthResponse;
import com.squad.record.AuthResponse;
import com.squad.service.AuthGrpcClientService;
import lombok.RequiredArgsConstructor;
import org.springframework.graphql.data.method.annotation.Argument;
import org.springframework.graphql.data.method.annotation.MutationMapping;
import org.springframework.graphql.data.method.annotation.QueryMapping;
import org.springframework.stereotype.Controller;

@Controller
@RequiredArgsConstructor
public class AuthController {
    private final AuthGrpcClientService authService;

    @QueryMapping
    public String ping() {
        return "pong";
    }

    @MutationMapping(name = "loginWithSteam")
    public AuthResponse loginWithSteam(@Argument("openidParamsJson") String openidParamsJson) {
        ResolveSteamAuthResponse response = authService.loginWithSteam(openidParamsJson);

        return new AuthResponse(
                response.getUserId(),
                response.getSteamId(),
                response.getIsNewUser(),
                response.getToken()
        );
    }
}
