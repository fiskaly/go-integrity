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

default: build

build:
	docker build \
	-t go-integrity:latest \
	-t go-integrity:`git describe --tags --always --dirty` \
	.
