package com.squad.event.grpc;

import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.advice.GrpcAdvice;
import net.devh.boot.grpc.server.advice.GrpcExceptionHandler;

@GrpcAdvice
@Slf4j
public class GrpcGlobalExceptionHandler {

    @GrpcExceptionHandler(IllegalArgumentException.class)
    public StatusRuntimeException handleIllegalArgumentException(IllegalArgumentException e) {
        log.warn("gRPC Invalid argument: {}", e.getMessage());
        return Status.INVALID_ARGUMENT
                .withDescription(e.getMessage())
                .asRuntimeException();
    }

    @GrpcExceptionHandler(IllegalArgumentException.class)
    public StatusRuntimeException handleIllegalStateException(IllegalStateException e) {
        log.warn("gRPC Failed precondition: {}", e.getMessage());
        return Status.FAILED_PRECONDITION
                .withDescription(e.getMessage())
                .asRuntimeException();
    }

    @GrpcExceptionHandler(Exception.class)
    public StatusRuntimeException handleGeneralException(Exception e) {
        log.warn("Unhandled gRPC server exception {}", e.getMessage(), e);
        return Status.INTERNAL
                .withDescription(e.getMessage())
                .asRuntimeException();
    }
}
