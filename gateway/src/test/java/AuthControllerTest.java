import com.squad.controller.AuthController;
import com.squad.grpc.user.ResolveSteamAuthResponse;
import com.squad.service.AuthGrpcClientService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.graphql.GraphQlTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.graphql.test.tester.GraphQlTester;
import org.springframework.test.context.ContextConfiguration;

import static org.mockito.Mockito.*;


@GraphQlTest(AuthController.class)
@ContextConfiguration(classes = AuthController.class)
public class AuthControllerTest {

    @Autowired
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

        when(authGrpcClientService.loginWithSteam(any())).thenReturn(response);

        //language=GraphQl
        String mutation = """
                mutation {
                    loginWithSteam(openidParamsJson: "{\\"openid.mode\\":\\"id_res\\"}") {
                        userId
                        steamId
                        token
                        isNewUser
                    }
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
