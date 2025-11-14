# ADR-001: User authorization

## Status

Draft

## Context

We have many ready-made solutions that handle registration, login, refresh tokens, and MFA for us.
Examples include: boilerplate code, AWS Cognito, and Supabase. The choice isn’t straightforward
due to the nature of startups — uncertainty, limited budget, unknown costs, and the need for rapid iteration.

Knowing that the team may not be large, and that we want to stay focused on fast product iterations,
scalability, security, and keeping the team manageable in later stages:

We have different types of users, with different data models and different login frequencies:
  - Business units (Venues, Sponsors, Franchises)
  - Streamers
  - Players / Viewers / Participants
  - Partners (e.g., Instreamly)

We need to remember that we are looking for trade-offs — rapid implementation at the first stage
and the ability to support 10,000+ to <500,000 users in the first 6 months.

Ideally, such a system should:
  - Support multiple roles and access levels
  - Be ready to handle corporate clients (e.g., Costa Coffee), who may require:
      - High security standards (SOC2, ISO27001, possibly HIPAA)
      - GDPR/RODO-compliant data storage
      - Multi-tenancy (data separation)
      - Audit logs
      - SSO integration
  - Require minimal security maintenance from our team
  - Be ready for microservices and serverless architectures

## Decision

It's Supabase

## Consequences

### Positive

- ✅ Faster implementation for early stages
    - Supabase Auth works out-of-the-box.
    - Ideal for small teams focused on delivering product iterations, not maintaining auth infrastructure. Cogito requires much time for the configuration

- ✅ Fully managed Postgres + Auth 
    - No need to maintain or scale the database early on.
    - RLS (Row Level Security) gives strong data isolation without writing IAM policies. (I would say it more neutral, due to feature migration)
    - Easy to prototype multi-tenant access rules at the DB layer.

- ✅ Simple developer experience / Lower curve learning
   - Everything visible in one dashboard.
   - Onboarding new developers is much easier than navigating Cognito, IAM, API Gateway, ALB, Secrets Manager, etc.

- ✅ Lower initial cost and operational overhead

- ✅ Works well with external microservices
   - Even without Edge Functions, Supabase issues standard JWTs.
   - Our AWS Fargate microservices / lambdas can validate these JWTs using shared keys.
   - No vendor-specific coupling at the service layer.

### Negative

- ❌ No deep integration with AWS ecosystem
   - Cognito integrates natively with:
       - API Gateway authorizers
       - IAM roles
       - Private link / VPC endpoints
   - With Supabase, every microservice must handle its own JWT verification.
- ❌ Operational risk at scale
- ❌ No built-in multi-tenant identity isolation
    - Supabase supports multi-tenancy at the DB level (via RLS),
    - but not on the identity provider level like Cognito + IAM.
    - For enterprise clients, this may be a blocker.
- ❌ Migration complexity later
- ❌ Not ideal for enterprise identity needs
- ❌ Scaling limits compared to AWS
   - Supabase Postgres can become a bottleneck under very high load.
   - Cognito is designed for tens of millions of monthly active users with near-infinite horizontal scaling.

### Risks

1. Scaling limitations of Supabase Postgres
   P: Medium
   I: High
   Severity: High
   Description: At higher volumes (hundreds of thousands+ MAU), Postgres may become a bottleneck for auth events, writes, RLS checks, and real-time features. Scaling requires vertical upgrades and careful DB design.

2. Vendor downtime or degraded performance
   P: Medium
   I: High
   Severity: High
   Description: Outages of Supabase Auth or Postgres can take the entire platform down. 
   No control over incident response or SLA (unless enterprise plan).

3. Limited enterprise security and compliance

4. No immediate session invalidation (JWT-based model)
   P: High
   I: Medium
   Severity: High
   Description: Supabase JWTs are stateless; session revocation is not immediate. 
                Potential security risk for high-sensitivity apps.

5. Limited IAM/granular authorization
   P: High
   I: Medium
   Severity: High
   Description: Supabase handles identity, not deep authorization. 
                Complex RBAC/ABAC requires custom logic or RLS-based workarounds 
                that become harder as the system grows.

6. Reliance on single database for identity + operational data

7. Hard migration path to AWS Cognito later
   P: Medium
   I: Medium/High
   Severity: Medium/High

8. Cost unpredictability at scale
   P: Medium
   I: Medium
   Severity: Medium

## Alternatives Considered

### Alternative 1: Cogito

### Alternative 2: Pure Code / Boilerplate

## References

- [Why I Switched From AWS Cognito To Supabase The Week Before My Startup Launched](https://dev.to/sleeplessfox/why-i-switched-from-aws-cognito-to-supabase-the-week-before-my-startup-launched-269c)
- [Paid 360$ for AWS Cognito in December. Just switched to Supabase server side auth](https://www.reddit.com/r/Supabase/comments/1i27oow/paid_360_for_aws_cognito_in_december_just/)
- [Supabase Part 3: Multi Tenancy](https://arda.beyazoglu.com/supabase-multi-tenancy)
- [Multi-Tenancy Architecture - System Design](https://www.geeksforgeeks.org/system-design/multi-tenancy-architecture-system-design/)
- [supabase-vs-aws-pricing](https://www.bytebase.com/blog/supabase-vs-aws-pricing/)