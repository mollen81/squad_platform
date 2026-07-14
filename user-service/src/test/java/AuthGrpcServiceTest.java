import com.squad.gateway.grpc.user.ResolveSteamAuthRequest;
import com.squad.gateway.grpc.user.ResolveSteamAuthResponse;
import com.squad.user.domain.UserEntity;
import com.squad.user.event.UserRegisteredEvent;
import com.squad.user.kafka.UserEventProducer;
import com.squad.user.repository.UserRepository;
import com.squad.user.service.AuthGrpcService;
import com.squad.user.service.SteamOpenIdValidator;
import io.grpc.StatusRuntimeException;
import io.grpc.stub.StreamObserver;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Captor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Optional;
import java.util.UUID;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyMap;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
public class AuthGrpcServiceTest {

    @InjectMocks
    AuthGrpcService authGrpcService;

    @Mock
    SteamOpenIdValidator steamValidator;

    @Mock
    UserRepository userRepo;

    @Mock
    StreamObserver<ResolveSteamAuthResponse> responseObserver;

    @Mock
    UserEventProducer userEventProducer;

    @Captor
    private ArgumentCaptor<ResolveSteamAuthResponse> responseCaptor;

    ResolveSteamAuthRequest request;
    // Steam: Mollen (https://steamcommunity.com/profiles/76561198888277695)
    private final String VALID_STEAM_ID = "76561198888277695";

    @BeforeEach
    void setup() {
        request = ResolveSteamAuthRequest.newBuilder()
                .putOpenidParams("opendid.mode", "id_res")
                .build();
    }

    // tests
    @Test
    public void resolveSteamAuth_NewUser_CreatesUserAndSendsEvent() {
        when(steamValidator.validateAndExtractSteamId(anyMap())).thenReturn(VALID_STEAM_ID);
        when(userRepo.findBySteamId(VALID_STEAM_ID)).thenReturn(Optional.empty());

        UserEntity savedUser = UserEntity.builder()
                .id(UUID.randomUUID())
                .steamId(VALID_STEAM_ID)
                .build();
        when(userRepo.save(any(UserEntity.class))).thenReturn(savedUser);

        authGrpcService.resolveSteamAuth(request, responseObserver);

        verify(userRepo, times(1)).save(any(UserEntity.class));
        verify(userEventProducer, times(1)).sendUserRegisteredEvent(any(UserRegisteredEvent.class));

        verify(responseObserver).onNext(responseCaptor.capture());
        verify(responseObserver).onCompleted();

        ResolveSteamAuthResponse response = responseCaptor.getValue();
        assertEquals(savedUser.getId().toString(), response.getUserId());
        assertEquals(VALID_STEAM_ID, response.getSteamId());
        assertTrue(response.getIsNewUser());
        assertNotNull(response.getToken());
    }

    @Test
    public void resolveSteamAuth_ExistingUser_UpdatesLoginTime() {
        when(steamValidator.validateAndExtractSteamId(anyMap())).thenReturn(VALID_STEAM_ID);
        UserEntity existingUser = UserEntity.builder()
                .id(UUID.randomUUID())
                .steamId(VALID_STEAM_ID)
                .build();
        when(userRepo.findBySteamId(VALID_STEAM_ID)).thenReturn(Optional.of(existingUser));

        when(userRepo.save(any(UserEntity.class))).thenReturn(existingUser);

        authGrpcService.resolveSteamAuth(request, responseObserver);

        verify(userRepo, times(1)).save(existingUser);
        verify(userEventProducer, never()).sendUserRegisteredEvent(any());

        verify(responseObserver).onNext(responseCaptor.capture());
        ResolveSteamAuthResponse response = responseCaptor.getValue();
        assertFalse(response.getIsNewUser());
    }

    @Test
    public void resolveSteamAuth_InvalidSteamSignature_ReturnsError() {
        when(steamValidator.validateAndExtractSteamId(anyMap()))
                .thenThrow(new RuntimeException("Invalid Steam OpenID Signature"));

        authGrpcService.resolveSteamAuth(request, responseObserver);

        verify(responseObserver).onError(any(StatusRuntimeException.class));
        verify(userRepo, never()).save(any());
        verify(responseObserver, never()).onNext(any());
    }
}
