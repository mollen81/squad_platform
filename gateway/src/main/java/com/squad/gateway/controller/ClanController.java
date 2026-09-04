package com.squad.gateway.controller;

import com.squad.gateway.record.ClanService.ProcessAcceptanceResponse;
import com.squad.gateway.service.ClanGrpcClientService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/v1")
@RequiredArgsConstructor
public class ClanController {
    private final ClanGrpcClientService clanGrpcClientService;

    record CreateClanDto(String leaderId, String name, String tag, String description, String requirements, String avatarUrl) {}
    record ApplyToClanDto(String userId, String clanId, String socialLink, String experienceText) {}
    record ResolveAppDto(String moderatorId, boolean isApproved) {}
    public record ClanMemberDto(String id, String userId, String role) {}
    public record ClanDto(String id, String name, String tag, String avatarUrl, String description, String requirements, boolean isRecruiting,
            String status, int totalElo, int minElo, String createdAt, List<ClanMemberDto> members) {}
    public record ClanActionResponseDto(String id, String message) {}
    public record ProcessAcceptanceDto(String acceptorId) {}
    public record ProcessAcceptanceResponse(ClanMemberDto newMember, String message) {}

    // clan creation: POST /api/v1/clans
    @PostMapping("/clans")
    public ResponseEntity<?> createClan(@RequestBody CreateClanDto request) {
        try {
            var grpcResponse = clanGrpcClientService.createClan(
                    request.name, request.tag, request.leaderId,
                    request.description, request.requirements, request.avatarUrl);
            return ResponseEntity.ok(new ClanActionResponseDto(
                    grpcResponse.getClanId(),
                    grpcResponse.getMessage()));
        }
        catch (Exception e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    // clan get with members list: GET /api/v1/clans/{clanId}
    @GetMapping("/clans/{clanId}")
    public ResponseEntity<?> getClanWithMembers(@PathVariable String clanId) {
        try {
            var grpcResponse = clanGrpcClientService.getClanWithMembers(clanId);

            List<ClanMemberDto> members = grpcResponse.getMembersList().stream()
                    .map(m -> new ClanMemberDto(m.getId(), m.getUserId(), m.getRole()))
                    .toList();

            ClanDto response = new ClanDto(
                    grpcResponse.getId(),
                    grpcResponse.getName(),
                    grpcResponse.getTag(),
                    grpcResponse.getAvatarUrl(),
                    grpcResponse.getDescription(),
                    grpcResponse.getRequirements(),
                    grpcResponse.getIsRecruiting(),
                    grpcResponse.getStatus(),
                    grpcResponse.getTotalElo(),
                    grpcResponse.getMinElo(),
                    grpcResponse.getCreatedAt(),
                    members
            );
            return ResponseEntity.ok(response);
        }
        catch (Exception e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    @PostMapping("/clans/{clanId}/applications")
    public ResponseEntity<?> applyToClan(@PathVariable String clanId, @RequestBody ApplyToClanDto request) {
        try {
            var grpcResponse = clanGrpcClientService.applyToClan(
                    request.userId, request.clanId, request.socialLink, request.experienceText);
            return ResponseEntity.ok(new ClanActionResponseDto(
                    grpcResponse.getApplicationId(),
                    grpcResponse.getMessage()
            ));
        } catch (Exception e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    @PatchMapping("/applications/{applicationId}/status")
    public ResponseEntity<?> resolveApplication(
            @PathVariable String applicationId,
            @RequestBody ProcessAcceptanceDto request) {
        try {
            var grpcResponse = clanGrpcClientService.processAcceptance(applicationId, request.acceptorId);
            ClanMemberDto memberDto = grpcResponse.hasNewMember()
                    ? new ClanMemberDto(
                            grpcResponse.getNewMember().getId(),
                            grpcResponse.getNewMember().getUserId(),
                            grpcResponse.getNewMember().getRole())
                    : null;

            return ResponseEntity.ok(new ProcessAcceptanceResponse(memberDto, grpcResponse.getMessage()));
        }
        catch (Exception e) {
            throw new RuntimeException(e);
        }
    }
}
