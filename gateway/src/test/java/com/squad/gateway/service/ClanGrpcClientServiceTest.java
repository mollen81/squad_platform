package com.squad.gateway.service;

import com.squad.clan.grpc.ClanServiceGrpc;
import com.squad.clan.grpc.CreateClanRequest;
import com.squad.clan.grpc.CreateClanResponse;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.Mockito;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.mockito.ArgumentMatchers.any;

@ExtendWith(MockitoExtension.class)
public class ClanGrpcClientServiceTest {
    @Mock
    private ClanServiceGrpc.ClanServiceBlockingStub clanStub;

    @InjectMocks
    private ClanGrpcClientService clanGrpcClientService;

    @Test
    void createClan_ShouldCallStubAndReturnResponse() {
        CreateClanResponse grpcResponse = CreateClanResponse.newBuilder()
                .setClanId("1234")
                .setMessage("Clan successfully created")
                .build();

        Mockito.when(clanStub.createClan(any(CreateClanRequest.class))).thenReturn(grpcResponse);

        CreateClanResponse result = clanGrpcClientService.createClan(
                "TestClan",
                "TST",
                "user-123",
                "test clan",
                "test requirements",
                "test-url:");

        assertNotNull(result);
        assertEquals("1234", result.getClanId());
        assertEquals("Clan successfully created", result.getMessage());

        Mockito.verify(clanStub, Mockito.times(1)).createClan(any(CreateClanRequest.class));
    }
}
