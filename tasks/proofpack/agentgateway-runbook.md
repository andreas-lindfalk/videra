# Videra + AgentGateway Runbook (Proof Pack)

Date: 2026-03-04

## Purpose

Define a practical, non-intrusive integration model where AgentGateway is the enterprise edge and Videra remains focused on MCP tool/resource execution.

## Architecture Responsibility Split

- AgentGateway owns:
  - authn/authz (JWT, RBAC, CEL policy)
  - edge observability/audit
  - federation of multiple MCP services
- Videra owns:
  - ingestion, indexing, hybrid retrieval
  - deterministic ranking behavior
  - tool/resource schema compatibility

## Deployment Modes

### Local / On-Prem Validation

1. Start Videra via Docker (`run-http` or compose).
2. Route AgentGateway MCP endpoint to Videra MCP endpoint.
3. Validate tool calls through gateway:
   - `index_video`
   - `search_video`
   - `list_videos`
   - `video://{id}/transcript`

### Cloud Path (Later)

- Keep search path latency-oriented in service runtime.
- Keep heavy indexing workloads in async job boundaries.
- Preserve MCP contract compatibility while changing infrastructure shape.

## Federation Notes

- Videra should remain one federated MCP service among others.
- Avoid embedding gateway concerns into Videra core packages.
- Prefer transport compatibility expected by gateway MCP integrations.

## Validation Checklist

- MCP handshake succeeds through AgentGateway endpoint.
- Tool/resource payloads unchanged vs direct Videra endpoint (except gateway-added headers/metadata).
- Auth/policy failures are enforced at gateway, not by Videra internal logic.
- Observability traces/metrics visible at gateway edge.

## References

- AgentGateway docs: https://agentgateway.dev/docs/
- AgentGateway MCP overview: https://agentgateway.dev/docs/mcp/
