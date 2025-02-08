--
-- Copyright (C) 2024 IOTech Ltd
--
-- SPDX-License-Identifier: Apache-2.0

-- "<serviceKey>".store is used to save StoredObject as JSONB on failure
-- The table is used to store the data that failed to be processed by the application service
-- note that "<serviceKey>" is a placeholder to be replaced with the actual service key in the runtime
CREATE TABLE IF NOT EXISTS "<serviceKey>".store (
    id UUID PRIMARY KEY,
    created timestamp NOT NULL DEFAULT now(),
    content JSONB NOT NULL
);
