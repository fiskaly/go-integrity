#
#   Copyright (C) 2026 fiskaly GmbH <https://fiskaly.com>
#   All rights reserved.
#
#   Developed by: Philipp Paulweber et al.
#   <https://github.com/fiskaly/go-integrity/graphs/contributors>
#
#   This file is part of go-integrity.
#
#   go-integrity is free software: you can redistribute it and/or modify
#   it under the terms of the Apache License 2.0.
#
#   go-integrity is distributed in the hope that it will be useful,
#   but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY
#   or FITNESS FOR A PARTICULAR PURPOSE. See the Apache License 2.0 for more details.
#
#   You should have received a copy of the Apache License 2.0 along with go-integrity.
#   If not, see <https://www.apache.org/licenses/LICENSE-2.0>
#

FROM golang:1.25-alpine \
  AS build

ENV USER=appuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

RUN apk update \
 && apk add --no-cache ca-certificates git \
 && update-ca-certificates

WORKDIR /app

RUN --mount=type=bind,source=.,target=. \
    CGO_ENABLED=0 \
    go build -o /go/bin/go-integrity . \
 && /go/bin/go-integrity -v

FROM scratch \
  AS image

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build /go/bin/go-integrity /go-integrity

USER appuser:appuser
ENTRYPOINT ["/go-integrity"]
CMD ["-help"]
