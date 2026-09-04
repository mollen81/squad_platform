package com.squad.gateway.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.squad.clan.grpc.CreateClanResponse;
import com.squad.clan.grpc.GetClanResponse;
import com.squad.gateway.service.ClanGrpcClientService;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import static org.mockito.ArgumentMatchers.anyString;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@WebMvcTest(ClanController.class)
public class ClanControllerTest {
    @Autowired
    private MockMvc mockMvc;

    @Autowired
    private ObjectMapper objectMapper;

    @MockBean
    private ClanGrpcClientService clanGrpcClientService;

    @Test
    void createClan_ShouldReturnOk_WhenValidRequest() throws Exception {
        ClanController.CreateClanDto request = new ClanController.CreateClanDto(
                "TestClan",
                "TST",
                "user-123",
                "test clan",
                "test requirements",
                "test-url:"
        );

        CreateClanResponse mockResponse = CreateClanResponse.newBuilder()
                .setClanId("1234")
                .setMessage("Clan successfully created")
                .build();

        Mockito.when(clanGrpcClientService.createClan(
                anyString(), anyString(), anyString(), anyString(), anyString(), anyString()))
                .thenReturn(mockResponse);

        mockMvc.perform(post("/api/v1/clans")
                    .contentType(MediaType.APPLICATION_JSON)
                    .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.id").value("1234"))
                .andExpect(jsonPath("$.message").value("Clan successfully created"));
    }

    @Test
    void getClanWithMembers_ShouldReturnOk_WhenValidRequest() throws Exception {
        String clanId = "1234";

        GetClanResponse mockResponse = GetClanResponse.newBuilder()
                .setId("1234")
                .setName("TestClan")
                .build();
    }
}
