# Design decisions

## Contract

The workspace did not include the frontend source, dialogue transcript, or tests. The following is therefore the explicit implementation assumption, not a claim that it was confirmed from frontend code.

All endpoints use JSON and amounts are integer currency units (for example, yen). IDs are opaque strings.

| Method | Path | Request | Response |
|---|---|---|---|
| `POST` | `/api/groups` | `{ "name": string }` | `201` group |
| `GET` | `/api/groups` | none | `200` group array |
| `GET` | `/api/groups/:groupId` | none | `200` group |
| `POST` | `/api/groups/:groupId/members` | `{ "name": string }` | `201` member |
| `GET` | `/api/groups/:groupId/members` | none | `200` member array |
| `POST` | `/api/groups/:groupId/expenses` | `{ "description": string, "amount": integer, "paidBy": member ID, "participants": member ID[] }` | `201` expense |
| `GET` | `/api/groups/:groupId/expenses` | none | `200` expense array |
| `GET` | `/api/groups/:groupId/balances` | none | `200` `{ "memberId": string, "balance": integer }[]` |
| `GET` | `/api/groups/:groupId/settlements` | none | `200` `{ "from": string, "to": string, "amount": integer }[]` |
| `GET` | `/healthz` | none | `200` `{ "status": "ok" }` |

Errors have the form `{ "error": string }`.

## Settlement rules and assumptions

- An expense is paid by one group member and shared equally by its listed participants. The payer does not need to be a participant.
- A member's balance is `total paid - total owed`. Positive means receive money; negative means pay money.
- Each expense is split into integer units. The remainder is assigned one unit at a time to participants in request order. This conserves the exact original amount and is deterministic.
- Settlements greedily match debtors and creditors in member/request order. This produces a small transfer set for the resulting creditor/debtor lists, but does not promise a globally preferred ordering.
- Settlement output is calculated from current expenses; there is no completed-payment workflow.
- Amounts must be positive integers. Duplicate participants and unknown member IDs are rejected.

## Design and rejected options

The service uses Go's standard `net/http` package and an in-memory store. This keeps the container small and has no runtime dependency. The trade-off is that data disappears on restart and the service cannot be scaled across replicas.

SQLite was considered for persistence but rejected because the frontend contract and persistence requirements are unavailable. A database-backed repository can replace `Store` later without changing the HTTP shapes.

Floating-point amounts were rejected because currency arithmetic must conserve exact units. Decimal strings were left out because fractional-currency representation was not confirmed.

Recording completed transfers was rejected because the prompt establishes calculated settlement output, not a payment workflow or authentication model.

## Scope and unknowns

Authentication, authorization, pagination, deletion/editing semantics, currency code, persistence, and whether responses should use an envelope such as `{ "data": ... }` could not be confirmed from this workspace. They are intentionally outside this implementation. The service has no authentication, durable storage, or CORS policy.