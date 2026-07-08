import com.squad.GatewayApplication;
import com.squad.grpc.user.ResolveSteamAuthResponse;
import com.squad.service.AuthGrpcClientService;
import org.junit.jupiter.api.Test;
import org.springframework.boot.test.autoconfigure.graphql.GraphQlTest;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.graphql.test.tester.GraphQlTester;

import static org.mockito.Mockito.*;


@ComponentScan(basePackages = "/gateway")
@SpringBootTest(classes = GatewayApplication.class)
public class AuthControllerTest {

    private GraphQlTester graphQlTester;

    @MockBean
    private AuthGrpcClientService authGrpcClientService;

    // Steam: Mollen (https://steamcommunity.com/profiles/76561198888277695)
    private final String VALID_STEAM_ID = "76561198888277695";

    @Test
    void loginWithSteam_ReturnsAuthResponse() {
        ResolveSteamAuthResponse response = ResolveSteamAuthResponse.newBuilder()
                .setUserId("user-uuid-12345")
                .setSteamId(VALID_STEAM_ID)
                .setToken("jwt-example-token")
                .setIsNewUser(true)
                .build();

        when(authGrpcClientService.loginWithSteam(VALID_STEAM_ID)).thenReturn(response);

        String mutation = """
                mutation {
                    loginWithSteam(openidParamsJson: "{\\"openid.mode\\":\\id_res\\"})
                        userId
                        steamId
                        token
                        isNewUser
                }
                """;

        graphQlTester.document(mutation)
                .execute()
                .path("loginWithSteam.userId").entity(String.class).isEqualTo("user-uuid-12345")
                .path("loginWithSteam.steamId").entity(String.class).isEqualTo(VALID_STEAM_ID)
                .path("loginWithSteam.token").entity(String.class).isEqualTo("jwt-example-token")
                .path("loginWithSteam.isNewUser").entity(Boolean.class).isEqualTo(true);
    }
}
