#!/bin/bash

CREATE TABLE loki_log.log_entry
(
    timestamp String,
    job String,
    namespace String,
    node_name String,
    pod String,
    app String,
    container String,
    filename String,
    stream String,
    time DateTime64(9, 'UTC'),
    apiName String,
    proxyResponseCode String,
    destination String,
    apiCreatorTenantDomain String,
    platform String,
    apiMethod String,
    apiVersion String,
    gatewayType String,
    apiCreator String,
    responseCacheHit String,
    backendLatency String,
    correlationId String,
    requestMediationLatency String,
    keyType String,
    apiId String,
    applicationName String,
    targetResponseCode String,
    requestTimestamp String,
    applicationOwner String,
    userAgent String,
    eventType String,
    apiResourceTemplate String,
    regionId String,
    responseLatency String,
    responseMediationLatency String,
    userIp String,
    apiContext String,
    applicationId String,
    apiType String
)

ENGINE = MergeTree
PARTITION BY toDate(time)
ORDER BY (time);