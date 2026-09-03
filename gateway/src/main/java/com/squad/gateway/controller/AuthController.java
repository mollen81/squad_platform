package com.squad.gateway.controller;

import com.squad.gateway.record.AuthService.AuthResponse;
import com.squad.gateway.record.AuthService.SteamLoginRequest;
import com.squad.grpc.user.ResolveSteamAuthResponse;
import com.squad.gateway.service.AuthGrpcClientService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/api/v1/auth")
@RequiredArgsConstructor
public class AuthController {
    private final AuthGrpcClientService authService;

    @GetMapping
    public String ping() {
        return "pong";
    }

    @PostMapping(name = "loginWithSteam")
    public ResponseEntity<?> loginWithSteam(@RequestBody SteamLoginRequest request) {
        try {
            ResolveSteamAuthResponse grpcResponse = authService.loginWithSteam(request.openIdParamsJson());

            AuthResponse response = new AuthResponse(
                    grpcResponse.getUserId(),
                    grpcResponse.getSteamId(),
                    grpcResponse.getToken(),
                    grpcResponse.getIsNewUser()
            );

            return ResponseEntity.ok(response);
        } catch (Exception e) {
            return ResponseEntity.status(401).body(Map.of("error", e.getMessage()));
        }
    }
}