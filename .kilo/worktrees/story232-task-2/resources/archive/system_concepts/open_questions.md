1. ~~In policy.sql in the namespace_family table there's a~~
   ~~validity time range of type tstzrange. My understanding~~
   ~~of the namespace family is that it's a spec for who owns~~
   ~~the forecast, things like "Retail", "Warehouse Labor Plan" and so on. While I believe that we want to specify how far~~
   ~~out a plan should go (and how far back actuals should be processed), I don't think that tstzrange should be used for~~
   ~~that since it can only represent absolute date/time ranges~~
   ~~not relative ones. I've commented it out of the definition~~
   ~~for the time being, but we should figure out what's the right thing here. Same question for namespace. - Doug 2/10/2026~~

**RESOLVED (2/19/2026):** The validity field has been re-enabled in `namespace_family`, `namespace`, and `category` tables. The decision is to use a simplified bitemporal approach where:

- **Lower bound**: Set to `current_timestamp` at creation time
- **Upper bound**: Set to `infinity` (represented as `NULL` in PostgreSQL's `tstzrange`)

This represents "valid from now until further notice" and allows for future implementation of full bitemporal versioning (close-and-insert pattern for updates) without requiring immediate complexity. The validity field is used for system-time tracking (when a record version is valid in the database), not for business-time planning horizons. See Story011 for implementation details.

2. If namespaces are meant to be things like "Retail Prod", "Retail Dev", "Personal Sandbox", then having to recreate the categories for each of those seems like the wrong thing. Maybe categories are associated with the namespace family and inherited by the namespace? Or perhaps there's a set of canonical categories associated with the namespace family and they're automagically or on-demand replicated into the namespace? Same questions for category. - Doug 2/10/2026
