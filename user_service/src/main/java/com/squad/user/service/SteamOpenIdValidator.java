package com.squad.user.service;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpEntity;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;
import org.springframework.http.HttpHeaders;
import org.springframework.web.client.RestTemplate;

import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Service
@Slf4j
@RequiredArgsConstructor
public class SteamOpenIdValidator {
    private static final String STEAM_OPENID_URL = "https://steamcommunity.com/openid/login";
    private static final Pattern STEAM_ID_PATTERN =
            Pattern.compile("https://steamcommunity\\\\.com/openid/id/(\\\\d+)");

    private final RestTemplate restTemplate = new RestTemplate();

    public String validateAndExtractSteamId(Map<String, String> openIdParams) {
        MultiValueMap<String, String> requestBody = new LinkedMultiValueMap<>();
        openIdParams.forEach(requestBody::add);
        requestBody.set("openIdMode", "check_authentication");

        HttpHeaders httpHeaders = new HttpHeaders();
        httpHeaders.setContentType(MediaType.APPLICATION_FORM_URLENCODED);

        HttpEntity<MultiValueMap<String, String>> request =
                new HttpEntity<>(requestBody, httpHeaders);

        try {
            ResponseEntity<String> response = restTemplate.postForEntity(
                    STEAM_OPENID_URL,
                    request,
                    String.class);
            String responseBody = response.getBody();

            if(responseBody != null && responseBody.contains("is_valid:true")) {
                String claimedId = openIdParams.get("openid.claimed_id");
                Matcher matcher = STEAM_ID_PATTERN.matcher(claimedId);

                if(matcher.find()) {
                    String steamId64 = matcher.group(1);
                    log.info("Successful Steam authorization. Extracted SteamID: {}", steamId64);
                    return steamId64;
                }
                else {
                    log.error("Couldn't extract SteamID from claimed_id: {}", claimedId);
                    throw new RuntimeException("Incorrect format claimedId");
                }

            }
            else {
                log.warn("Steam rejected the signature verification. Answer: {}", responseBody);
                throw new RuntimeException("Invalid Steam OpenID Signature");
            }
        }
        catch (Exception e) {
            log.error("Error during Steam OpenID validation", e);
            throw new RuntimeException("Steam OpenID validation error: " + e.getMessage());
        }
    }
}
